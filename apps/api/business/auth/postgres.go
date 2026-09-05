package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "users_active_email_key" {
			return nil, fmt.Errorf("create user: %w", ErrUserEmailAlreadyExists)
		}
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

// RevokeAllSessionsForUser revokes every active session for a user in
// one statement. Used by the account-deletion flow (T04) and any
// other irreversible user-state change; revoking the requester's
// current session at email-change confirm time is intentionally not
// done (VOC-031-D05) — changing a login address is not equivalent to
// revoking a session.
func (r *PostgreSQLRepository) RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`,
		revokedAt, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("revoke all sessions rows: %w", err)
	}
	return n, nil
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

// RevokeAllMagicLinksForUser revokes every unconsumed magic_links
// row for userID in one statement. Used by the account-deletion
// path (VOC-031-T04) so no in-flight sign-in link can be consumed
// after the account is deactivated.
func (r *PostgreSQLRepository) RevokeAllMagicLinksForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE magic_links SET revoked_at = $1
		 WHERE user_id = $2 AND consumed_at IS NULL AND revoked_at IS NULL`,
		revokedAt, userID)
	if err != nil {
		return 0, fmt.Errorf("revoke all magic links: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("revoke all magic links rows: %w", err)
	}
	return n, nil
}

func (r *PostgreSQLRepository) AttachMagicLinkUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE magic_links SET user_id = $1 WHERE id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("attach magic link user: %w", err)
	}
	return nil
}

func (r *PostgreSQLRepository) CreateOAuthState(ctx context.Context, tokenHash []byte, environment, provider, appReturnURL string, createdAt, expiresAt time.Time) (*OAuthState, error) {
	id := uuid.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO oauth_states (id, token_hash, environment, provider, app_return_url, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, tokenHash, environment, provider, appReturnURL, createdAt, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create oauth state: %w", err)
	}
	return &OAuthState{ID: id, Environment: environment, Provider: provider, AppReturnURL: appReturnURL, CreatedAt: createdAt, ExpiresAt: expiresAt}, nil
}

func (r *PostgreSQLRepository) GetOAuthStateByTokenHash(ctx context.Context, tokenHash []byte) (*OAuthState, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, environment, provider, app_return_url, created_at, expires_at, consumed_at
		 FROM oauth_states WHERE token_hash = $1`, tokenHash)
	var o OAuthState
	var consumedAt sql.NullTime
	err := row.Scan(&o.ID, &o.Environment, &o.Provider, &o.AppReturnURL, &o.CreatedAt, &o.ExpiresAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("oauth state not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan oauth state: %w", err)
	}
	if consumedAt.Valid {
		o.ConsumedAt = &consumedAt.Time
	}
	return &o, nil
}

func (r *PostgreSQLRepository) ConsumeOAuthState(ctx context.Context, id uuid.UUID, consumedAt time.Time) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE oauth_states SET consumed_at = $1 WHERE id = $2 AND consumed_at IS NULL`, consumedAt, id)
	if err != nil {
		return false, fmt.Errorf("consume oauth state: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("consume oauth state rows: %w", err)
	}
	return n == 1, nil
}

func (r *PostgreSQLRepository) GetExternalIdentity(ctx context.Context, provider, providerSubject string) (*ExternalIdentity, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_subject, provider_email, provider_email_verified, created_at, updated_at
		 FROM external_identities
		 WHERE provider = $1 AND provider_subject = $2 AND deleted_at IS NULL`,
		provider, providerSubject)
	return scanExternalIdentity(row)
}

func (r *PostgreSQLRepository) CreateExternalIdentity(ctx context.Context, userID uuid.UUID, provider, providerSubject, providerEmail string, providerEmailVerified bool) (*ExternalIdentity, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO external_identities (id, user_id, provider, provider_subject, provider_email, provider_email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, userID, provider, providerSubject, providerEmail, providerEmailVerified, now, now)
	if err != nil {
		return nil, fmt.Errorf("create external identity: %w", err)
	}
	return &ExternalIdentity{
		ID:                    id,
		UserID:                userID,
		Provider:              provider,
		ProviderSubject:       providerSubject,
		ProviderEmail:         providerEmail,
		ProviderEmailVerified: providerEmailVerified,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

func scanExternalIdentity(row *sql.Row) (*ExternalIdentity, error) {
	var e ExternalIdentity
	err := row.Scan(&e.ID, &e.UserID, &e.Provider, &e.ProviderSubject, &e.ProviderEmail, &e.ProviderEmailVerified, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("external identity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan external identity: %w", err)
	}
	return &e, nil
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

func (r *PostgreSQLRepository) CleanupExpiredOAuthStates(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM oauth_states WHERE expires_at <= $1 OR consumed_at IS NOT NULL`, before)
	if err != nil {
		return 0, fmt.Errorf("cleanup oauth states: %w", err)
	}
	return res.RowsAffected()
}
