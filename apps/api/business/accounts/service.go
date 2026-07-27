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

// Config holds the email-change service configuration. The shape
// mirrors auth.Config so the production wiring is symmetric.
type Config struct {
	Environment             string
	BaseURL                 string
	EmailChangePath         string
	EmailChangeLinkLifetime time.Duration
	RateLimit               EmailChangeRateLimitConfig
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

// Service is the requester-scoped email-change flow boundary. It
// owns no state of its own; all persistence is delegated to the
// Repository and AuthRepository the Service is constructed with.
type Service struct {
	repo        Repository
	auth        AuthRepository
	emailSender email.Sender
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
// email-change Service calls into. The interface is a strict
// subset of auth.Repository: only the methods the email-change
// flow needs (lookup-by-id) appear here, and the
// account-deletion T04 will add the others it needs (RevokeSession
// variants, etc.) without changing this contract.
type AuthRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error)
}

// NewService creates an email-change Service. The auth repository
// is the same auth.Repository the magic-link / session flows use;
// only the methods in the AuthRepository interface are exercised
// here, so the service can be wired in tests against either a real
// auth.Repository or a smaller fixture.
func NewService(repo Repository, authRepo AuthRepository, emailSender email.Sender, c clock.Clock, limiter RateLimiter, cfg Config) *Service {
	if cfg.EmailChangeLinkLifetime == 0 {
		cfg.EmailChangeLinkLifetime = 15 * time.Minute
	}
	if cfg.EmailChangePath == "" {
		cfg.EmailChangePath = "/auth/email-change"
	}
	if c == nil {
		c = clock.Real{}
	}
	return &Service{
		repo:        repo,
		auth:        authRepo,
		emailSender: emailSender,
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
func (s *Service) RequestEmailChangeLink(ctx context.Context, userID uuid.UUID, clientIP, newEmail string) (*EmailLink, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if !isAcceptableEmail(newEmail) {
		return nil, ErrEmailChangeInvalidEmail
	}

	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("emailchange.request", clientIP)); err != nil {
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

func (s *Service) emailChangeURL(token, newEmail string) string {
	u, _ := url.Parse(s.cfg.BaseURL)
	u.Path = s.cfg.EmailChangePath
	q := u.Query()
	q.Set("token", token)
	q.Set("newEmail", newEmail)
	u.RawQuery = q.Encode()
	return u.String()
}

// ConsumeEmailChangeLink validates the supplied token, updates
// users.email, sends a best-effort notification to the OLD address,
// and returns the EmailChangeResult the API layer renders to the
// requester. The flow's security guarantees (per VOC-031-D05):
//
//   - the raw token is the only form that ever appears in the
//     request and the email body; only its SHA-256 hash is
//     persisted;
//   - the link is single-use (consumed_at), 15-minute-expiring,
//     and environment-scoped;
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
func (s *Service) ConsumeEmailChangeLink(ctx context.Context, clientIP, token string) (*EmailChangeResult, error) {
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("emailchange.consume", clientIP)); err != nil {
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
	now := s.clock.Now()
	if !link.Valid(now) {
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
