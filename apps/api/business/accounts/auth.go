package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// EmailChangeLink is the service-layer projection of an
// email_change_links row. It mirrors the magic-links projection in
// the auth module: only the hash is persisted, the raw token never
// touches the database. user_id is non-null (requesting a change
// requires an authenticated session, per VOC-031-D05) and the row
// carries a single-use consumed_at, a revoke-able revoked_at, and a
// 15-minute expires_at identical to magic_links'.
type EmailChangeLink struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	NewEmail    string
	Environment string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	RevokedAt   *time.Time
}

// Valid reports whether the link has not expired, been consumed, or
// been revoked.
func (l EmailChangeLink) Valid(now time.Time) bool {
	return now.Before(l.ExpiresAt) && l.ConsumedAt == nil && l.RevokedAt == nil
}

// EmailChangeResult is the post-confirm projection returned to the
// API layer. The requester-scoped user identity is taken from the
// session and is never read from the body; the only field the
// handler needs back is the new email address that is now in effect
// on users.email (so the frontend can refresh local state) and the
// old email address that the notification was dispatched to (so the
// frontend can show the learner what was sent).
type EmailChangeResult struct {
	UserID         uuid.UUID
	OldEmail       string
	NewEmail       string
	NotificationTo string
	ChangedAt      time.Time
}

// Public service errors. The API layer maps these to stable 4xx
// responses and a 5xx for the unhandled case, matching the auth
// module's existing error mapping discipline.
var (
	// ErrInvalidEmailChangeLink is returned for any rejection of a
	// consume attempt (unknown token, expired, tampered,
	// previously-consumed, or wrong environment). The reason is
	// intentionally not distinguished in the response so a token
	// cannot be probed (mirrors auth.ErrInvalidMagicLink).
	ErrInvalidEmailChangeLink = errors.New("invalid or expired email change link")
	// ErrEmailAlreadyRegistered is returned when a confirm
	// attempt loses a duplicate-email race: another confirm
	// already claimed the new_email, and the partial unique
	// index on users.email rejected our UPDATE. Stable, non-500
	// response, never a crash (VOC-031-R02).
	ErrEmailAlreadyRegistered = errors.New("new email is already registered to another account")
	// ErrUserNotFound is returned when the requester's user
	// identity is no longer resolvable (deleted or otherwise
	// missing). Mirrors auth.ErrUserDisabled posture.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailChangeRateLimited is returned when the IP or
	// session rate-limiter rejects the request. Mirrors
	// auth.ErrRateLimited.
	ErrEmailChangeRateLimited = errors.New("email change rate limited")
	// ErrEmailChangeInvalidEmail is returned for syntactically
	// invalid or empty new_email values before any token work
	// runs.
	ErrEmailChangeInvalidEmail = errors.New("invalid email address")
	// ErrAccountDeletionIdempotencyKeyRequired is returned when
	// the request is missing the Idempotency-Key header that
	// DOC-07 makes required for this endpoint. Stable 400.
	ErrAccountDeletionIdempotencyKeyRequired = errors.New("idempotency key required")
	// ErrAccountDeletionIdempotencyConflict is returned when
	// the same Idempotency-Key was previously used for a
	// different request fingerprint. Stable 409.
	ErrAccountDeletionIdempotencyConflict = errors.New("idempotency key conflict")
	// ErrAccountDeletionAlreadyInFlight is returned when a
	// user already has a non-completed account_deletion_requests
	// row, so a second deactivation would race against the
	// first. Stable 409. The replay with the same
	// Idempotency-Key is still served (returning the original
	// row) — this error is only emitted when a fresh key
	// collides with an in-flight deletion.
	ErrAccountDeletionAlreadyInFlight = errors.New("account deletion already in flight")
	// ErrAccountDeletionRateLimited is returned when the IP or
	// session rate-limiter rejects the request.
	ErrAccountDeletionRateLimited = errors.New("account deletion rate limited")
	// ErrAccountDeletionSweep is the parent error wrapping any
	// per-table failure during the anonymization sweep. The
	// sweep is designed to be resumable: a wrapped error
	// indicates a row that needs to be retried, and the sweep
	// leaves the row in 'anonymizing' so the next call can pick
	// it up. The cause is preserved in the chain.
	ErrAccountDeletionSweep = errors.New("account deletion sweep failed")
)

// AccountDeletionRequest is the per-user "the learner requested
// account deletion" record (DOC-05 §16, DOC-06 §14,
// VOC-031-D07). The lifecycle is three-valued and strictly
// ordered: 'deactivated' (the user is deactivated and the
// purge_after clock is running), 'anonymizing' (the sweep
// claimed the row and is performing the per-table disposition),
// 'completed' (every per-table disposition has run).
type AccountDeletionRequest struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Status         string
	RequestedAt    time.Time
	PurgeAfter     time.Time
	CompletedAt    *time.Time
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Valid reports whether the request can still be processed by
// the sweep (i.e. its purge_after has passed and it is still in
// 'deactivated' status).
func (r AccountDeletionRequest) EligibleForPurge(now time.Time) bool {
	return r.Status == "deactivated" && !now.Before(r.PurgeAfter)
}

// AnonymizationCounters is the per-table count of rows the
// anonymization sweep mutated. Returned to the API layer for
// observability and to support the T04 acceptance evidence
// (a "documented per-table count" claim requires a real
// count, not an assertion).
type AnonymizationCounters struct {
	ExternalIdentities     int64
	UserWords              int64
	LearnerSentences       int64
	ReviewAttempts         int64
	AIFeedbackAttempts     int64
	AIQualityReviewReports int64
	ConfidencePointLedger  int64
	GraceDayLedger         int64
	FeatureAuditLogs       int64
	UserOnboardingProfiles int64
	UserSettings           int64
	DailyMissionSnapshots  int64
	DailyActivitySummaries int64
	StreakStates           int64
}

// SweepResult is the aggregate result of one sweep pass. The
// API/observability layer can render this directly; the per-row
// state is captured in the AccountDeletionRequest status
// transitions.
type SweepResult struct {
	Processed           int
	Anonymized          int
	Failed              int
	AnonymizationTotals AnonymizationCounters
}

// Repository is the persistence boundary for the email-change
// flow and the account-deletion flow. It owns the
// email_change_links and account_deletion_requests tables; the
// user, session, and cross-table operations are delegated to the
// AuthRepository and the cross-module Repository methods the
// Service is constructed with.
type Repository interface {
	// CreateEmailChangeLink inserts one row and returns the
	// projection. The token hash, expiry, and environment
	// values are supplied by the service layer.
	CreateEmailChangeLink(ctx context.Context, userID uuid.UUID, newEmail string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*EmailChangeLink, error)
	// GetEmailChangeLinkByTokenHash returns the projection or
	// an error; the service layer treats "not found" the same
	// as "expired" or "consumed" so the error message is not
	// exposed.
	GetEmailChangeLinkByTokenHash(ctx context.Context, tokenHash []byte) (*EmailChangeLink, error)
	// ConsumeEmailChangeLink marks the row consumed. Caller
	// must check Valid() first; this is the unconditional
	// write.
	ConsumeEmailChangeLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error
	// RevokeAllEmailChangeLinksForUser revokes every
	// unconsumed email_change_links row for the user. Called
	// from the account-deletion path (T04) so no stale link
	// can be consumed after the account is deactivated.
	RevokeAllEmailChangeLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error)
	// UpdateUserEmail changes users.email for the given user,
	// returning a stable sentinel when the partial unique
	// index on lower(email) WHERE deleted_at IS NULL rejects
	// the write. Implementations translate a SQLSTATE 23505
	// (unique_violation) on the email index into
	// ErrEmailAlreadyRegistered rather than a 500.
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string, now time.Time) error

	// CreateAccountDeletionRequest performs the entire
	// deactivation transaction in one call: marks the user
	// deleted, revokes every active session, every unconsumed
	// magic link, and every unconsumed email change link,
	// then inserts the account_deletion_requests row in
	// 'deactivated' state with purge_after = requested_at +
	// 30 days. The (user_id) UNIQUE constraint on
	// account_deletion_requests means a second deactivation
	// for the same user surfaces a SQLSTATE 23505, which the
	// repository translates to
	// ErrAccountDeletionAlreadyInFlight (stable 409, never a
	// 500). Returns the persisted row.
	CreateAccountDeletionRequest(ctx context.Context, userID uuid.UUID, idempotencyKey string, now time.Time, purgeDelay time.Duration) (*AccountDeletionRequest, error)
	// GetAccountDeletionRequestByUserID returns the current
	// row for the user, or an error if none exists. Used to
	// (a) re-validate a replayed Idempotency-Key and (b) feed
	// the sweep's claim-step.
	GetAccountDeletionRequestByUserID(ctx context.Context, userID uuid.UUID) (*AccountDeletionRequest, error)
	// ListDeactivatedRequestsDueForPurge returns up to limit due
	// 'deactivated' rows and stale 'anonymizing' claims. A stale claim is
	// eligible for recovery only when its updated_at is at or before
	// staleBefore; fresh claims remain owned by their current sweeper.
	ListDeactivatedRequestsDueForPurge(ctx context.Context, now, staleBefore time.Time, limit int) ([]AccountDeletionRequest, error)
	// ClaimAccountDeletionRequestForAnonymization atomically
	// transitions a due 'deactivated' row, or atomically reclaims
	// a stale 'anonymizing' row, to 'anonymizing'
	// and returns true when the transition succeeded (this
	// caller now owns the row). Returns false when another
	// sweeper already claimed it; the caller should skip the
	// row.
	ClaimAccountDeletionRequestForAnonymization(ctx context.Context, id uuid.UUID, now, staleBefore time.Time) (bool, error)
	// AnonymizeUserData runs the per-table deletion/redaction
	// disposition inside one transaction. Learner-linked records
	// are removed in foreign-key-safe order; the retained deleted
	// user row is stripped of identifiers (DOC-05 §16). Returns
	// per-table counts for observability.
	AnonymizeUserData(ctx context.Context, userID uuid.UUID) (AnonymizationCounters, error)
	// FinalizeAccountDeletionClaim holds the request-row lock while it verifies
	// claim ownership, purges learner data, and marks the request completed.
	// A stale worker returns completed=false without touching learner data.
	FinalizeAccountDeletionClaim(ctx context.Context, id, userID uuid.UUID, claimedAt, now time.Time) (counters AnonymizationCounters, completed bool, err error)
}
