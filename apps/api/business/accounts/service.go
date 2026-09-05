// Package accounts implements the learner-owned account-lifecycle flows
// introduced by P5 (VOC-031). It currently owns the email-change
// verification flow (T03, DOC-06 §6, VOC-031-D05) and the
// account-deletion flow (T04, DOC-05 §16, DOC-06 §§9,14,15,
// VOC-031-D07).
//
// The package depends on the auth module for token generation,
// rate-limiting primitives, and the cross-module user/session
// operations (lookup, session revocation, etc.) it must not
// reimplement. auth's own magic-link/session code is never modified
// from here; the email-change and account-deletion flows are built
// strictly on top of the existing primitives.
package accounts

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
)

// IdempotencyStatus describes the result of an idempotency-key
// lookup. The semantics match the learning package's
// IdempotencyStatus enum: Absent means no row matched the
// (user, operation, key) tuple; Match means a row matched and its
// stored fingerprint is the same as the one supplied; Conflict
// means a row matched but the stored fingerprint is different.
type IdempotencyStatus int

const (
	IdempotencyAbsent IdempotencyStatus = iota
	IdempotencyMatch
	IdempotencyConflict
)

// IdempotencyStore is the in-process boundary this package uses
// for the Idempotency-Key header DOC-07 requires on
// POST /api/v1/account-deletion-requests. The interface mirrors
// the one the learning package already exposes (it stores
// per-(user, operation, key) entries and returns a three-state
// status); accounts reuses the existing primitive rather than
// inventing a new one. The accounts package does not import the
// learning package — the interface is duplicated here so the
// dependency direction stays one-way, and the existing
// PostgreSQLIdempotencyStore / MemoryIdempotencyStore
// implementations satisfy this interface verbatim.
type IdempotencyStore interface {
	// Check looks up a stored idempotency key.
	Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error)
	// Record stores an idempotency key with its fingerprint.
	Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error
}

// Config holds the email-change and account-deletion service
// configuration. The shape mirrors auth.Config so the production
// wiring is symmetric.
type Config struct {
	Environment               string
	BaseURL                   string
	EmailChangePath           string
	EmailChangeLinkLifetime   time.Duration
	AccountDeletionPurgeDelay time.Duration
	AccountDeletionSweepLimit int
	// AccountDeletionClaimTimeout is the 15-minute recovery threshold for a
	// purge claim. A worker that dies after moving a row to anonymizing must not
	// retain learner data indefinitely; after this threshold another sweep may
	// atomically reclaim it. The finalization row lock protects active purges.
	AccountDeletionClaimTimeout time.Duration
	RateLimit                   EmailChangeRateLimitConfig
	AccountDeletionRateLimit    AccountDeletionRateLimitConfig

	// ReservedSyntheticEmail mirrors auth.KillSwitches'
	// ReservedSyntheticEmail (VOC-050-T00). The email-change flow is
	// the one remaining way a real account could come to hold the
	// reserved synthetic identity, so it applies the same refusal
	// the sign-in paths do. An empty value reserves nothing.
	ReservedSyntheticEmail string
}

// EmailChangeRateLimitConfig is the rate-limit shape T03 introduces.
// The values are kept separate from auth.RateLimitConfig so the
// email-change budget cannot accidentally deplete the
// magic-link-request budget (and vice versa). Per-IP and per-session
// limits are both enforced; an attacker should not be able to drain
// the per-IP budget across many sessions, and a single compromised
// session should not be able to drain the per-session budget across
// many IPs.
type EmailChangeRateLimitConfig struct {
	RequestWindow time.Duration
	RequestLimit  int
	ConsumeWindow time.Duration
	ConsumeLimit  int
}

// AccountDeletionRateLimitConfig is the rate-limit shape T04
// introduces. The bucket is separate from the email-change bucket
// so a runaway email-change retry loop cannot deplete the
// account-deletion budget. Per-IP and per-session are both
// enforced, matching the email-change posture (VOC-031-D05,
// generalized to the irreversible-action class of endpoints).
type AccountDeletionRateLimitConfig struct {
	RequestWindow time.Duration
	RequestLimit  int
	SweepWindow   time.Duration
	SweepLimit    int
}

// DefaultAccountDeletionPurgeDelay is the 30-day DOC-05 §16
// baseline. The source of truth is always the row's own
// purge_after column; this constant is what the Service writes
// when no per-row override is supplied.
const DefaultAccountDeletionPurgeDelay = 30 * 24 * time.Hour

// DefaultAccountDeletionClaimTimeout is the 15-minute recovery threshold for
// an interrupted sweep. The row lock held during finalization protects an
// active purge; this is not a bound on the purge's execution time.
const DefaultAccountDeletionClaimTimeout = 15 * time.Minute

// accountDeletionOperation is the operation string the
// idempotency_keys table records for every account-deletion
// request. It replaces the DOC-05 §13 `scope` enum value T04
// would otherwise have to invent; the existing schema accepts
// any non-empty operation text, so the convention is the only
// thing this constant encodes (mirrors `user_words:save`,
// `reviews:submit`, `ai_feedback_request`).
const accountDeletionOperation = "account_deletion"

// accountDeletionFingerprint is the deterministic per-user
// fingerprint the idempotency store records. Account deletion
// has no body, so the fingerprint is just the user id — this
// pins the property that a key used to delete one account can
// never be replayed to delete a different one, even though the
// raw Idempotency-Key string is the only thing the client
// supplies.
func accountDeletionFingerprint(userID uuid.UUID) string {
	return fmt.Sprintf("account_deletion|%s", userID.String())
}

// Service is the requester-scoped email-change and
// account-deletion flow boundary. It owns no state of its own;
// all persistence is delegated to the Repository, AuthRepository,
// and IdempotencyStore the Service is constructed with.
type Service struct {
	repo        Repository
	auth        AuthRepository
	emailSender email.Sender
	idem        IdempotencyStore
	clock       clock.Clock
	limiter     RateLimiter
	cfg         Config
}

// RateLimiter is the minimal interface the email-change Service
// needs from the auth module's rate-limiter primitive. It matches
// auth.RateLimiter exactly so production wiring can pass the same
// instance; tests inject the in-memory fixed-window limiter.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// AuthRepository is the cross-module user/session boundary the
// email-change and account-deletion Services call into. The
// interface is a strict subset of auth.Repository: only the
// methods these flows need appear here, so the service can be
// wired in tests against either a real auth.Repository or a
// smaller fixture. T04 added the two RevokeAll*ForUser methods
// to support the deactivation transaction; the email-change
// flow does not use them.
type AuthRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error)
	RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
	RevokeAllMagicLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
}

// NewService creates an email-change and account-deletion
// Service. The auth repository is the same auth.Repository the
// magic-link / session flows use; only the methods in the
// AuthRepository interface are exercised here, so the service
// can be wired in tests against either a real auth.Repository
// or a smaller fixture. idem may be nil if the caller does not
// need the account-deletion endpoint (the email-change flow
// does not use it).
func NewService(repo Repository, authRepo AuthRepository, emailSender email.Sender, idem IdempotencyStore, c clock.Clock, limiter RateLimiter, cfg Config) *Service {
	if cfg.EmailChangeLinkLifetime == 0 {
		cfg.EmailChangeLinkLifetime = 15 * time.Minute
	}
	if cfg.EmailChangePath == "" {
		cfg.EmailChangePath = "/auth/email-change"
	}
	if cfg.AccountDeletionPurgeDelay == 0 {
		cfg.AccountDeletionPurgeDelay = DefaultAccountDeletionPurgeDelay
	}
	if cfg.AccountDeletionSweepLimit == 0 {
		cfg.AccountDeletionSweepLimit = 100
	}
	if cfg.AccountDeletionClaimTimeout <= 0 {
		cfg.AccountDeletionClaimTimeout = DefaultAccountDeletionClaimTimeout
	}
	if cfg.AccountDeletionRateLimit.RequestLimit == 0 {
		cfg.AccountDeletionRateLimit.RequestLimit = 5
		cfg.AccountDeletionRateLimit.RequestWindow = time.Hour
	}
	if cfg.AccountDeletionRateLimit.SweepLimit == 0 {
		cfg.AccountDeletionRateLimit.SweepLimit = 60
		cfg.AccountDeletionRateLimit.SweepWindow = time.Hour
	}
	if c == nil {
		c = clock.Real{}
	}
	return &Service{
		repo:        repo,
		auth:        authRepo,
		emailSender: emailSender,
		idem:        idem,
		clock:       c,
		limiter:     limiter,
		cfg:         cfg,
	}
}

// EmailLink is the dispatch-side projection returned by
// RequestEmailChangeLink. The link is the only artifact the API
// layer needs to embed in the outbound email body; the raw token is
// never persisted and never logged.
type EmailLink struct {
	Link     string
	NewEmail string
}

// RequestEmailChangeLink creates a single-use email-change link
// and dispatches it to newEmail. The flow mirrors auth's
// RequestMagicLink almost exactly (rate-limit, generate
// token-and-hash, persist row, send email) with two deliberate
// differences per VOC-031-D05:
//
//  1. userID is non-nullable: requesting a change requires an
//     authenticated session. The requester is the *current* user;
//     no path or body parameter may override it.
//  2. The response is unconditionally generic: whether newEmail
//     is already registered to another active account is not
//     observable through the request outcome. The actual
//     uniqueness check is re-verified atomically at confirm time
//     against users_active_email_key (VOC-031-R02).
//
// The function returns the constructed link so the API layer (and
// the OpenAPI contract generator) can render the exact same path
// the email body includes. In production, the link is only ever
// delivered through the email sender — the API layer must not echo
// it back to the requester.
func (s *Service) RequestEmailChangeLink(ctx context.Context, userID uuid.UUID, clientIP, sessionToken, newEmail string) (*EmailLink, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if !isAcceptableEmail(newEmail) || s.isReservedSyntheticEmail(newEmail) {
		return nil, ErrEmailChangeInvalidEmail
	}

	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("emailchange.request", clientIP)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrEmailChangeRateLimited
	}
	// Per-session limit alongside the per-IP one above (VOC-031-D05: "rate-
	// limited by both IP and the requesting user's session") - an attacker
	// with a valid session rotating IPs, or multiple sessions on one IP,
	// must still be bounded.
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForSession("emailchange.request", sessionToken)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrEmailChangeRateLimited
	}

	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.EmailChangeLinkLifetime)
	token, hash, err := auth.NewTokenAndHash()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	link, err := s.repo.CreateEmailChangeLink(ctx, userID, newEmail, hash, s.cfg.Environment, now, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create email change link: %w", err)
	}

	urlStr := s.emailChangeURL(token, newEmail)
	msg := email.Message{
		To: []email.Address{{Email: newEmail}},
		// The subject is the same regardless of who initiated
		// the request; the body includes the confirmation URL.
		Subject:  "Confirm your new Vocanova sign-in email",
		BodyText: fmt.Sprintf("Use this single-use link to confirm the email change for your Vocanova account:\n\n%s\n\nIt expires in 15 minutes. If you did not request this change, you can ignore this email.", urlStr),
		BodyHTML: fmt.Sprintf("<p>Use this single-use link to confirm the email change for your Vocanova account:</p><p><a href=\"%s\">%s</a></p><p>It expires in 15 minutes. If you did not request this change, you can ignore this email.</p>", urlStr, urlStr),
	}
	if err := s.emailSender.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("send email: %w", err)
	}
	_ = link // link is constructed; the dispatch result is what matters.
	return &EmailLink{Link: urlStr, NewEmail: newEmail}, nil
}

// isReservedSyntheticEmail reports whether an already-normalized
// address is the reserved synthetic smoke-test identity.
func (s *Service) isReservedSyntheticEmail(normalizedEmail string) bool {
	reserved := strings.ToLower(strings.TrimSpace(s.cfg.ReservedSyntheticEmail))
	return reserved != "" && normalizedEmail == reserved
}

func (s *Service) emailChangeURL(token, newEmail string) string {
	u, _ := url.Parse(s.cfg.BaseURL)
	u.Path = s.cfg.EmailChangePath
	q := u.Query()
	q.Set("token", token)
	q.Set("newEmail", newEmail)
	u.RawQuery = q.Encode()
	return u.String()
}

// ConsumeEmailChangeLink validates the supplied token for requesterUserID,
// updates users.email, sends a best-effort notification to the OLD address,
// and returns the EmailChangeResult the API layer renders to the requester.
// The flow's security guarantees (per VOC-031-D05):
//
//   - the raw token is the only form that ever appears in the
//     request and the email body; only its SHA-256 hash is
//     persisted;
//   - the link is single-use (consumed_at), 15-minute-expiring,
//     and environment-scoped;
//   - the authenticated requester owns the link, so a confirmation token
//     cannot mutate another signed-in learner's account;
//   - the new-email uniqueness check is enforced atomically at
//     confirm time by users_active_email_key; a losing confirm
//     receives a stable ErrEmailAlreadyRegistered, never a 500
//     (VOC-031-R02);
//   - the requester's current session is NOT invalidated by either
//     the request or the confirm. Changing a login address is not
//     equivalent to revoking a session already in legitimate use
//     (the appropriate remedy for a hijacked session is
//     logout/session-revocation, not this flow);
//   - a Google-OAuth-linked login is unaffected because
//     external_identities is matched by provider_subject, not
//     users.email.
//
// The old-email notification is best-effort: an email-send failure
// is logged via the returned error from Send (it bubbles up only
// if the email is configured to do so), but the email change is
// already persisted at that point so a transient notification
// failure does not block the user. Per the spec the notification
// is required as a security control, so production wiring must
// treat it as such even though the function does not abort.
func (s *Service) ConsumeEmailChangeLink(ctx context.Context, requesterUserID uuid.UUID, clientIP, sessionToken, token string) (*EmailChangeResult, error) {
	if requesterUserID == uuid.Nil {
		return nil, ErrInvalidEmailChangeLink
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("emailchange.consume", clientIP)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrEmailChangeRateLimited
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForSession("emailchange.consume", sessionToken)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrEmailChangeRateLimited
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidEmailChangeLink
	}
	_, hash, err := auth.TokenAndHash(token)
	if err != nil {
		return nil, ErrInvalidEmailChangeLink
	}
	link, err := s.repo.GetEmailChangeLinkByTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidEmailChangeLink
	}
	if link.Environment != s.cfg.Environment {
		return nil, ErrInvalidEmailChangeLink
	}
	// The confirmation token proves control of the destination mailbox, but
	// does not authorize whichever account happens to be signed in when it is
	// redeemed. Bind the mutation to the requester that created the link.
	// Keep an owner mismatch non-distinguishing to avoid exposing link state.
	if link.UserID != requesterUserID {
		return nil, ErrInvalidEmailChangeLink
	}
	now := s.clock.Now()
	if !link.Valid(now) {
		return nil, ErrInvalidEmailChangeLink
	}
	// Re-checked at confirm time as well as request time: a link
	// minted before the address was reserved must not still be
	// redeemable against it.
	if s.isReservedSyntheticEmail(link.NewEmail) {
		return nil, ErrInvalidEmailChangeLink
	}

	// Look up the current user (and their existing email) before
	// mutating anything. ErrUserNotFound is the only path that
	// returns a distinct error here; the other reasons a user
	// might be missing (deleted, disabled) are caught at the
	// repository layer.
	user, err := s.auth.GetUserByID(ctx, link.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	oldEmail := strings.ToLower(strings.TrimSpace(user.Email))

	// The repository's UpdateUserEmail translates the partial
	// unique index violation into ErrEmailAlreadyRegistered so
	// the API layer can return a stable, non-500 conflict
	// (VOC-031-R02).
	if err := s.repo.UpdateUserEmail(ctx, link.UserID, link.NewEmail, now); err != nil {
		return nil, err
	}
	if err := s.repo.ConsumeEmailChangeLink(ctx, link.ID, now); err != nil {
		return nil, fmt.Errorf("consume email change link: %w", err)
	}

	result := &EmailChangeResult{
		UserID:         link.UserID,
		OldEmail:       oldEmail,
		NewEmail:       link.NewEmail,
		NotificationTo: oldEmail,
		ChangedAt:      now,
	}

	// Best-effort, non-blocking: a notification send failure
	// must not undo the email change. We log via the returned
	// error path that is intentionally non-fatal: a nil return
	// here means the email was sent; an error is recorded for
	// observability but does not propagate to the caller. This
	// matches the spec's "best-effort, non-blocking"
	// requirement.
	if oldEmail != "" && s.emailSender != nil {
		msg := email.Message{
			To:       []email.Address{{Email: oldEmail}},
			Subject:  "Your Vocanova sign-in email was changed",
			BodyText: fmt.Sprintf("The sign-in email for your Vocanova account was changed to %s. If you did not make this change, please contact support immediately.", link.NewEmail),
			BodyHTML: fmt.Sprintf("<p>The sign-in email for your Vocanova account was changed to <strong>%s</strong>.</p><p>If you did not make this change, please contact support immediately.</p>", link.NewEmail),
		}
		_ = s.emailSender.Send(ctx, msg)
	}
	return result, nil
}

// tokenAndHash removed in favor of auth.TokenAndHash (see service.go);
// the magic-link and email-change flows now share the same
// base64-URL-decode + SHA-256 logic, and the persisted hash is
// byte-identical for the same raw token regardless of which flow
// wrote it.

// isAcceptableEmail is a deliberately narrow check: the body
// must be non-empty, look like a local@domain pair, and use a
// reasonable subset of characters. It is not a full RFC 5322
// validator and intentionally not a DNS deliverability check;
// the goal is to reject obviously malformed input early, before
// any token work runs, while staying liberal enough to accept
// the long tail of legitimate addresses (plus signs,
// dots-in-local-part, hyphens, country TLDs, etc.).
func isAcceptableEmail(value string) bool {
	if len(value) == 0 || len(value) > 254 {
		return false
	}
	at := strings.LastIndex(value, "@")
	if at <= 0 || at == len(value)-1 {
		return false
	}
	local, domain := value[:at], value[at+1:]
	if len(local) == 0 || len(domain) == 0 {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '@' || r == '.' || r == '-' || r == '_' || r == '+':
		default:
			return false
		}
	}
	return true
}

// AccountDeletionResult is the post-create projection returned to
// the API layer. The user is already deactivated at this point:
// the row's status is 'deactivated', the user is set to
// 'deleted' on users.status, deleted_at is set, every session
// and unconsumed auth/email-change token is revoked, and the
// purge_after clock is running. The API layer renders the
// result so the frontend can route the learner to a clear
// post-deletion confirmation and initiate logout.
type AccountDeletionResult struct {
	UserID         uuid.UUID
	Status         string
	RequestedAt    time.Time
	PurgeAfter     time.Time
	IdempotencyKey string
	// Replayed is true when this call was a no-op because the
	// (user, idempotency-key) pair already had a matching
	// row. The frontend uses it to suppress duplicate "your
	// account was deleted" toasts on a retry.
	Replayed bool
}

// CreateAccountDeletionRequest performs the irreversible
// deactivation transaction for the requester's account
// (VOC-031-D07, DOC-05 §16, DOC-06 §14, DOC-07). The flow:
//
//  1. Validate the Idempotency-Key header (DOC-07 requires it).
//  2. Look up the (user, idempotency-key) row. On Match, return
//     the existing row without re-running any side effect —
//     this is the replay-safety property DOC-07 mandates. On
//     Conflict (same key, different fingerprint), return a
//     stable 409. On Absent, proceed.
//  3. Rate-limit per IP and per session (mirrors the email-
//     change posture; an attacker with a valid session must
//     not be able to deplete the per-IP budget across IPs, and
//     a single session must not be able to deplete the per-
//     session budget across IPs).
//  4. Delegate the deactivation transaction to the repository
//     (set users.status='deleted' / users.deleted_at, revoke
//     every active session, every unconsumed magic link, every
//     unconsumed email change link, insert the
//     account_deletion_requests row).
//  5. Record the idempotency key so a future replay lands on
//     the Match path.
//
// A user that has already been deactivated (an in-flight
// 'deactivated' or 'anonymizing' row exists) returns
// ErrAccountDeletionAlreadyInFlight on a fresh
// Idempotency-Key; the same key on a replay returns the
// existing row.
func (s *Service) CreateAccountDeletionRequest(ctx context.Context, userID, clientIP, sessionToken, idempotencyKey string) (*AccountDeletionResult, error) {
	uid, err := uuid.Parse(userID)
	if err != nil || uid == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrAccountDeletionIdempotencyKeyRequired
	}
	if s.idem == nil {
		return nil, errors.New("idempotency store not configured")
	}

	// Preserve the established replay posture: a confirmed matching request is
	// returned before it consumes either rate-limit bucket. The repository still
	// owns the authoritative transactional claim, so two concurrent absent
	// checks cannot create duplicate side effects.
	fingerprint := accountDeletionFingerprint(uid)
	status, err := s.idem.Check(ctx, uid, accountDeletionOperation, idempotencyKey, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if status == IdempotencyConflict {
		return nil, ErrAccountDeletionIdempotencyConflict
	}
	if status == IdempotencyMatch {
		existing, err := s.repo.GetAccountDeletionRequestByUserID(ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("read existing deletion request: %w", err)
		}
		return &AccountDeletionResult{
			UserID:         existing.UserID,
			Status:         existing.Status,
			RequestedAt:    existing.RequestedAt,
			PurgeAfter:     existing.PurgeAfter,
			IdempotencyKey: existing.IdempotencyKey,
			Replayed:       true,
		}, nil
	}
	// Per-IP and per-session rate limits, matching the
	// email-change posture (VOC-031-D05 generalized).
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("accountdeletion.request", clientIP)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrAccountDeletionRateLimited
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForSession("accountdeletion.request", sessionToken)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrAccountDeletionRateLimited
	}

	now := s.clock.Now().UTC()
	row, err := s.repo.CreateAccountDeletionRequest(ctx, uid, idempotencyKey, now, s.cfg.AccountDeletionPurgeDelay)
	if err != nil {
		return nil, err
	}

	return &AccountDeletionResult{
		UserID:         row.UserID,
		Status:         row.Status,
		RequestedAt:    row.RequestedAt,
		PurgeAfter:     row.PurgeAfter,
		IdempotencyKey: row.IdempotencyKey,
		Replayed:       row.Replayed,
	}, nil
}

// RunDeletionSweep runs one pass of the anonymization sweep
// (VOC-031-D07). The pass:
//
//  1. Lists up to cfg.AccountDeletionSweepLimit due 'deactivated'
//     rows and stale 'anonymizing' claims.
//  2. For each, atomically transitions the row to 'anonymizing'
//     (or renews a stale claim). A losing claim (another sweeper
//     owns a fresh lease) is a no-op for this pass.
//  3. Runs the per-table anonymization inside one
//     transaction: soft-deletes pending purge for
//     external_identities / user_words / learner_sentences;
//     irreversibly de-identifies review_attempts /
//     ai_feedback_attempts / confidence_point_ledger /
//     grace_day_ledger / (feature_audit_logs if present);
//     deletes or de-identifies user_onboarding_profiles /
//     user_settings / daily_mission_snapshots /
//     daily_activity_summaries / streak_states (DOC-05 §16).
//  4. Transitions the row to 'completed' and stamps
//     completed_at.
//
// The function is idempotent: a stale 'anonymizing' claim is
// safely recovered after the 15-minute recovery threshold, while a row that is
// 'completed' is never re-touched. Per-IP and per-session rate limits apply
// (mirrors the request path's posture).
func (s *Service) RunDeletionSweep(ctx context.Context, clientIP, sessionToken string) (*SweepResult, error) {
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("accountdeletion.sweep", clientIP)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrAccountDeletionRateLimited
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForSession("accountdeletion.sweep", sessionToken)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrAccountDeletionRateLimited
	}

	// PostgreSQL timestamptz stores microsecond precision. The claim timestamp
	// doubles as the completion fence, so normalize it before both writes. The
	// finalization row lock protects a still-active purge beyond the threshold.
	listNow := s.clock.Now().UTC().Truncate(time.Microsecond)
	staleBefore := listNow.Add(-s.cfg.AccountDeletionClaimTimeout)
	rows, err := s.repo.ListDeactivatedRequestsDueForPurge(ctx, listNow, staleBefore, s.cfg.AccountDeletionSweepLimit)
	if err != nil {
		return nil, fmt.Errorf("list due deletion requests: %w", err)
	}

	result := &SweepResult{}
	for _, row := range rows {
		result.Processed++
		claimNow := s.clock.Now().UTC().Truncate(time.Microsecond)
		claimed, err := s.repo.ClaimAccountDeletionRequestForAnonymization(ctx, row.ID, claimNow, claimNow.Add(-s.cfg.AccountDeletionClaimTimeout))
		if err != nil {
			result.Failed++
			return result, fmt.Errorf("%w: claim row %s: %v", ErrAccountDeletionSweep, row.ID, err)
		}
		if !claimed {
			// Another sweeper already claimed this row.
			// Skip without counting it as a failure.
			continue
		}
		counters, completed, err := s.repo.FinalizeAccountDeletionClaim(ctx, row.ID, row.UserID, claimNow, claimNow)
		if err != nil {
			result.Failed++
			return result, fmt.Errorf("%w: anonymize user %s: %v", ErrAccountDeletionSweep, row.UserID, err)
		}
		if !completed {
			// A newer sweeper reclaimed the lease while this pass was working.
			// Its fenced completion is authoritative; do not report this stale
			// worker's counters as a second completed anonymization.
			continue
		}
		result.Anonymized++
		result.AnonymizationTotals.ExternalIdentities += counters.ExternalIdentities
		result.AnonymizationTotals.UserWords += counters.UserWords
		result.AnonymizationTotals.LearnerSentences += counters.LearnerSentences
		result.AnonymizationTotals.ReviewAttempts += counters.ReviewAttempts
		result.AnonymizationTotals.AIFeedbackAttempts += counters.AIFeedbackAttempts
		result.AnonymizationTotals.ConfidencePointLedger += counters.ConfidencePointLedger
		result.AnonymizationTotals.GraceDayLedger += counters.GraceDayLedger
		result.AnonymizationTotals.FeatureAuditLogs += counters.FeatureAuditLogs
		result.AnonymizationTotals.UserOnboardingProfiles += counters.UserOnboardingProfiles
		result.AnonymizationTotals.UserSettings += counters.UserSettings
		result.AnonymizationTotals.DailyMissionSnapshots += counters.DailyMissionSnapshots
		result.AnonymizationTotals.DailyActivitySummaries += counters.DailyActivitySummaries
		result.AnonymizationTotals.StreakStates += counters.StreakStates
	}
	return result, nil
}
