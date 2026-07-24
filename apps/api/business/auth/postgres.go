package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgreSQLRepository implements Repository against the T00 migration schema.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

func (r *PostgreSQLRepository) CreateUser(ctx context.Context, email string, verifiedAt *time.Time) (*User, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, status, email_verified_at, created_at, updated_at)
		 VALUES ($1, lower($2), 'active', $3, $4, $5)`,
		id, email, verifiedAt, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &User{
		ID:              id,
		Email:           email,
		Status:          "active",
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *PostgreSQLRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, status, email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanUser(row)
}

func (r *PostgreSQLRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, status, email_verified_at, last_login_at, created_at, updated_at
		 FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL`, email)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	var email sql.NullString
	var verifiedAt, lastLoginAt sql.NullTime
	err := row.Scan(&u.ID, &email, &u.Status, &verifiedAt, &lastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Email = email.String
	if verifiedAt.Valid {
		u.EmailVerifiedAt = &verifiedAt.Time
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	return &u, nil
}

func (r *PostgreSQLRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
		t, t, id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, createdAt, expiresAt time.Time) (*Session, error) {
	id := uuid.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, userID, tokenHash, createdAt, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &Session{ID: id, UserID: userID, CreatedAt: createdAt, ExpiresAt: expiresAt}, nil
}

func (r *PostgreSQLRepository) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, created_at, expires_at, revoked_at
		 FROM sessions WHERE token_hash = $1`, tokenHash)
	var s Session
	var revokedAt sql.NullTime
	err := row.Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}
	if revokedAt.Valid {
		s.RevokedAt = &revokedAt.Time
	}
	return &s, nil
}

func (r *PostgreSQLRepository) RevokeSession(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE id = $2`, revokedAt, id)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) CreateMagicLink(ctx context.Context, email string, tokenHash []byte, environment string, createdAt, expiresAt time.Time) (*MagicLink, error) {
	id := uuid.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO magic_links (id, email, token_hash, environment, created_at, expires_at)
		 VALUES ($1, lower($2), $3, $4, $5, $6)`,
		id, email, tokenHash, environment, createdAt, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create magic link: %w", err)
	}
	return &MagicLink{ID: id, Email: email, Environment: environment, CreatedAt: createdAt, ExpiresAt: expiresAt}, nil
}

func (r *PostgreSQLRepository) GetMagicLinkByTokenHash(ctx context.Context, tokenHash []byte) (*MagicLink, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, email, environment, created_at, expires_at, consumed_at, revoked_at
		 FROM magic_links WHERE token_hash = $1`, tokenHash)
	var m MagicLink
	var userID sql.NullString
	var consumedAt, revokedAt sql.NullTime
	err := row.Scan(&m.ID, &userID, &m.Email, &m.Environment, &m.CreatedAt, &m.ExpiresAt, &consumedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("magic link not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan magic link: %w", err)
	}
	if userID.Valid {
		uid, err := uuid.Parse(userID.String)
		if err == nil {
			m.UserID = &uid
		}
	}
	if consumedAt.Valid {
		m.ConsumedAt = &consumedAt.Time
	}
	if revokedAt.Valid {
		m.RevokedAt = &revokedAt.Time
	}
	return &m, nil
}

func (r *PostgreSQLRepository) ConsumeMagicLink(ctx context.Context, id uuid.UUID, consumedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE magic_links SET consumed_at = $1 WHERE id = $2`, consumedAt, id)
	if err != nil {
		return fmt.Errorf("consume magic link: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) RevokeMagicLink(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE magic_links SET revoked_at = $1 WHERE id = $2`, revokedAt, id)
	if err != nil {
		return fmt.Errorf("revoke magic link: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) AttachMagicLinkUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE magic_links SET user_id = $1 WHERE id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("attach magic link user: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at <= $1 OR (revoked_at IS NOT NULL AND revoked_at <= $1)`, before)
	if err != nil {
		return 0, fmt.Errorf("cleanup sessions: %w", err)
	}
	return res.RowsAffected()
}

func (r *PostgreSQLRepository) CleanupExpiredMagicLinks(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM magic_links WHERE expires_at <= $1 OR consumed_at IS NOT NULL OR (revoked_at IS NOT NULL AND revoked_at <= $1)`, before)
	if err != nil {
		return 0, fmt.Errorf("cleanup magic links: %w", err)
	}
	return res.RowsAffected()
}
