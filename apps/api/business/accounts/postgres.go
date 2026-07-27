package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgreSQLRepository implements Repository against the T03
// email_change_links migration. It owns no state of its own; the
// user-lookup and session-revocation operations the email-change
// flow needs are delegated to the AuthRepository the service is
// constructed with, mirroring the seed/settings split in the
// users module.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a Repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// CreateEmailChangeLink inserts one row. The token_hash is the
// SHA-256 of the raw token (auth.TokenAndHash); only the hash is
// persisted. user_id is required (NOT NULL) and new_email is
// non-empty (mirrors the magic_links discipline with the
// D05-specific differences).
func (r *PostgreSQLRepository) CreateEmailChangeLink(ctx context.Context, userID uuid.UUID, newEmail string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*EmailChangeLink, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if newEmail == "" {
		return nil, errors.New("new email required")
	}
	id := uuid.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at)
		 VALUES ($1, $2, lower($3), $4, $5, $6, $7)`,
		id, userID, newEmail, tokenHash, environment, createdAt, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create email change link: %w", err)
	}
	return &EmailChangeLink{
		ID:          id,
		UserID:      userID,
		NewEmail:    strings.ToLower(newEmail),
		Environment: environment,
		CreatedAt:   createdAt.UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}, nil
}

// GetEmailChangeLinkByTokenHash returns the projection or a
// not-found error.
func (r *PostgreSQLRepository) GetEmailChangeLinkByTokenHash(ctx context.Context, tokenHash []byte) (*EmailChangeLink, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, new_email, environment, created_at, expires_at, consumed_at, revoked_at
		 FROM email_change_links WHERE token_hash = $1`, tokenHash)
	var l EmailChangeLink
	var consumedAt, revokedAt sql.NullTime
	err := row.Scan(&l.ID, &l.UserID, &l.NewEmail, &l.Environment, &l.CreatedAt, &l.ExpiresAt, &consumedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("email change link not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan email change link: %w", err)
	}
	if consumedAt.Valid {
		t := consumedAt.Time
		l.ConsumedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		l.RevokedAt = &t
	}
	return &l, nil
}

// ConsumeEmailChangeLink marks the row consumed. The service is
// expected to have already verified Valid() so the row exists
// and is unconsumed.
func (r *PostgreSQLRepository) ConsumeEmailChangeLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE email_change_links SET consumed_at = $1 WHERE id = $2`, consumedAt, id)
	if err != nil {
		return fmt.Errorf("consume email change link: %w", err)
	}
	return nil
}

// RevokeAllEmailChangeLinksForUser revokes every unconsumed link
// for the user in one statement. Used by the account-deletion
// path (T04).
func (r *PostgreSQLRepository) RevokeAllEmailChangeLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE email_change_links SET revoked_at = $1
		 WHERE user_id = $2 AND consumed_at IS NULL AND revoked_at IS NULL`,
		revokedAt, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all email change links: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("revoke all email change links rows: %w", err)
	}
	return n, nil
}

// UpdateUserEmail sets users.email for the user, enforcing the
// partial unique index users_active_email_key. A unique_violation
// (SQLSTATE 23505) on the email index is translated to
// ErrEmailAlreadyRegistered so the API layer returns a stable
// 409-style conflict response, never an unhandled 500
// (VOC-031-R02). The error is detected via lib/pq's pq.Error
// type, matching the convention already used by the auth and
// users modules.
func (r *PostgreSQLRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = lower($2), updated_at = $3
		 WHERE id = $1 AND deleted_at IS NULL`,
		userID, newEmail, now)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		// Unique violation: lower(email) WHERE deleted_at IS NULL
		// already has this email. Translate to the stable
		// public error; the rest of the error chain is preserved
		// in the wrapped value for logs.
		return fmt.Errorf("%w: %s", ErrEmailAlreadyRegistered, pqErr.Message)
	}
	return fmt.Errorf("update user email: %w", err)
}
