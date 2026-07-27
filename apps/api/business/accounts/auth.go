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
)

// Repository is the persistence boundary for the email-change flow.
// It owns only the email_change_links table; user and session
// operations are delegated to the AuthRepository the Service is
// constructed with (mirrors the seed/settings split in the users
// module: the local Repository never holds a *sql.Tx and is the
// only thing the service calls to mutate EmailChangeLink rows).
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
}
