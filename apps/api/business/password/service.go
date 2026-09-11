package password

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrInvalidToken = errors.New("invalid or expired password token")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrDisabled = errors.New("password sign-in disabled")

type Service struct {
	db              *sql.DB
	mail            email.Sender
	clock           clock.Clock
	limiter         auth.RateLimiter
	baseURL, env    string
	sessionLifetime time.Duration
	enabled         func() bool
	signupAllowed   func(string) bool
	identityAllowed func(string) bool
}

func NewService(db *sql.DB, mail email.Sender, c clock.Clock, limiter auth.RateLimiter, baseURL, env string, sessionLifetime time.Duration, enabled func() bool, signupAllowed func(string) bool, identityAllowed func(string) bool) *Service {
	return &Service{db: db, mail: mail, clock: c, limiter: limiter, baseURL: strings.TrimRight(baseURL, "/"), env: env, sessionLifetime: sessionLifetime, enabled: enabled, signupAllowed: signupAllowed, identityAllowed: identityAllowed}
}
func (s *Service) allow(ctx context.Context, action, ip, address string) error {
	if s.enabled != nil && !s.enabled() {
		return ErrDisabled
	}
	for _, k := range []string{auth.KeyForIP(action, ip), auth.KeyForEmail(action, address)} {
		ok, e := s.limiter.Allow(ctx, k)
		if e != nil {
			return fmt.Errorf("rate limit: %w", e)
		}
		if !ok {
			return auth.ErrRateLimited
		}
	}
	return nil
}

// allowToken rate-limits proof consumption before database lookup or Argon2 work.
// Token keys are one-way hashes, so a limiter never retains bearer material.
func (s *Service) allowToken(ctx context.Context, action, ip, token string) error {
	if s.enabled != nil && !s.enabled() {
		return ErrDisabled
	}
	for _, k := range []string{auth.KeyForIP(action, ip), auth.KeyForSession(action, token)} {
		ok, err := s.limiter.Allow(ctx, k)
		if err != nil {
			return fmt.Errorf("rate limit: %w", err)
		}
		if !ok {
			return auth.ErrRateLimited
		}
	}
	return nil
}

func tokenIsBounded(token string) bool {
	// NewTokenAndHash emits 32 bytes in padded URL-safe base64: exactly 44
	// characters. Reject anything else before it reaches a limiter key or decoder.
	return len(token) == 44
}

func (s *Service) RequestSignup(ctx context.Context, ip, emailAddr, pass, name string) error {
	emailAddr = auth.NormalizeEmail(emailAddr)
	if err := s.allow(ctx, "password.signup", ip, emailAddr); err != nil {
		return err
	}
	if s.signupAllowed != nil && !s.signupAllowed(emailAddr) {
		return nil
	}
	// Signup requests are deliberately a no-op for an existing account. In
	// particular, they must never turn a request into a credential attachment.
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1) AND deleted_at IS NULL)`, emailAddr).Scan(&exists); err != nil {
		return fmt.Errorf("lookup signup account: %w", err)
	}
	if exists {
		return nil
	}
	name = strings.TrimSpace(name)
	if len([]rune(name)) > 80 {
		return errors.New("display name is too long")
	}
	if err := Validate(pass); err != nil {
		return err
	}
	hash, err := Hash(pass)
	if err != nil {
		return err
	}
	token, th, err := auth.NewTokenAndHash()
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	_, err = s.db.ExecContext(ctx, `INSERT INTO password_registration_links(id,email,display_name,password_hash,token_hash,environment,created_at,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, uuid.New(), emailAddr, name, hash, th, s.env, now, now.Add(15*time.Minute))
	if err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return s.mail.Send(ctx, email.Message{
		To:      []email.Address{{Email: emailAddr}},
		Subject: "Verify your VocaNova account",
		BodyText: "Finish setting up your VocaNova account by opening this link within 15 minutes:\n\n" +
			s.baseURL + "/auth/password/verify?token=" + token + "\n\n" +
			"If you did not request this, you can ignore this email. Do not share this link.",
	})
}

// Remaining token consumption is intentionally transaction-only in PostgreSQL; API wiring follows this service.

// VerifySignup atomically consumes the proof token and creates the user and
// credential. Existing accounts are never modified by this path.
func (s *Service) VerifySignup(ctx context.Context, ip, token string) error {
	if !tokenIsBounded(token) {
		return ErrInvalidToken
	}
	if err := s.allowToken(ctx, "password.signup.verify", ip, token); err != nil {
		return err
	}
	_, th, err := auth.TokenAndHash(token)
	if err != nil {
		return ErrInvalidToken
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var emailAddr, name, hash string
	var expires time.Time
	var consumed, revoked sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT email,COALESCE(display_name,''),password_hash,expires_at,consumed_at,revoked_at FROM password_registration_links WHERE token_hash=$1 AND environment=$2 FOR UPDATE`, th, s.env).Scan(&emailAddr, &name, &hash, &expires, &consumed, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidToken
	}
	if err != nil {
		return fmt.Errorf("lookup registration link: %w", err)
	}
	if !s.clock.Now().Before(expires) || consumed.Valid || revoked.Valid || (s.signupAllowed != nil && !s.signupAllowed(emailAddr)) {
		return ErrInvalidToken
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1) AND deleted_at IS NULL)`, emailAddr).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrInvalidToken
	}
	now := s.clock.Now().UTC()
	id := uuid.New()
	if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,email,display_name,status,email_verified_at,created_at,updated_at) VALUES($1,$2,NULLIF($3,''),'active',$4,$4,$4)`, id, emailAddr, name, now); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			// A concurrent proof consumed the email first. Treat it exactly as
			// any other unusable proof; this path never attaches a credential.
			return ErrInvalidToken
		}
		return fmt.Errorf("create verified password user: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO password_credentials(user_id,password_hash,created_at,updated_at) VALUES($1,$2,$3,$3)`, id, hash, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE password_registration_links SET consumed_at=$1 WHERE token_hash=$2`, now, th); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) RequestReset(ctx context.Context, ip, emailAddr string) error {
	emailAddr = auth.NormalizeEmail(emailAddr)
	if err := s.allow(ctx, "password.reset", ip, emailAddr); err != nil {
		return err
	}
	if s.identityAllowed != nil && !s.identityAllowed(emailAddr) {
		return nil
	}
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT u.id FROM users u WHERE lower(u.email)=lower($1) AND u.status='active' AND u.deleted_at IS NULL AND u.email_verified_at IS NOT NULL`, emailAddr).Scan(&id)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	token, th, err := auth.NewTokenAndHash()
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	if _, err = s.db.ExecContext(ctx, `INSERT INTO password_reset_links(id,user_id,email_at_issue,token_hash,environment,created_at,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, uuid.New(), id, emailAddr, th, s.env, now, now.Add(15*time.Minute)); err != nil {
		return err
	}
	return s.mail.Send(ctx, email.Message{
		To:      []email.Address{{Email: emailAddr}},
		Subject: "Reset or add your VocaNova password",
		BodyText: "Open this link within 15 minutes to reset or add a VocaNova password:\n\n" +
			s.baseURL + "/auth/password/reset?token=" + token + "\n\n" +
			"If you did not request this, you can ignore this email. Do not share this link.",
	})
}

// Reset atomically replaces the credential, consumes sibling reset links and
// revokes every session and magic link. It never creates a session.
func (s *Service) Reset(ctx context.Context, ip, token, pass string) error {
	if !tokenIsBounded(token) {
		return ErrInvalidToken
	}
	if err := s.allowToken(ctx, "password.reset.consume", ip, token); err != nil {
		return err
	}
	_, th, err := auth.TokenAndHash(token)
	if err != nil {
		return ErrInvalidToken
	}
	if err := Validate(pass); err != nil {
		return err
	}
	hash, err := Hash(pass)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID uuid.UUID
	// Read the owner only to establish a canonical lock order. Every reset
	// operation locks users first, then its link; concurrent reset links for one
	// user therefore serialize instead of deadlocking while revoking siblings.
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM password_reset_links WHERE token_hash=$1 AND environment=$2`, th, s.env).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidToken
	}
	if err != nil {
		return fmt.Errorf("lookup reset link owner: %w", err)
	}
	var current string
	var verified sql.NullTime
	var status string
	err = tx.QueryRowContext(ctx, `SELECT email,email_verified_at,status FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, userID).Scan(&current, &verified, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidToken
	}
	if err != nil {
		return fmt.Errorf("lock reset user: %w", err)
	}
	if status != "active" || !verified.Valid || (s.identityAllowed != nil && !s.identityAllowed(current)) {
		return ErrInvalidToken
	}
	var issued string
	var expires time.Time
	var consumed, revoked sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT email_at_issue,expires_at,consumed_at,revoked_at FROM password_reset_links WHERE user_id=$1 AND token_hash=$2 AND environment=$3 FOR UPDATE`, userID, th, s.env).Scan(&issued, &expires, &consumed, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidToken
	}
	if err != nil {
		return fmt.Errorf("lock reset link: %w", err)
	}
	if issued != current || !s.clock.Now().Before(expires) || consumed.Valid || revoked.Valid {
		return ErrInvalidToken
	}
	now := s.clock.Now().UTC()
	if _, err = tx.ExecContext(ctx, `INSERT INTO password_credentials(user_id,password_hash,created_at,updated_at) VALUES($3,$1,$2,$2) ON CONFLICT(user_id) DO UPDATE SET password_hash=EXCLUDED.password_hash, updated_at=EXCLUDED.updated_at`, hash, now, userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE password_reset_links SET consumed_at=$1 WHERE token_hash=$2 AND consumed_at IS NULL AND revoked_at IS NULL`, now, th); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE password_reset_links SET revoked_at=$1 WHERE user_id=$2 AND token_hash<>$3 AND consumed_at IS NULL AND revoked_at IS NULL`, now, userID, th); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE magic_links SET revoked_at=$1 WHERE user_id=$2 AND consumed_at IS NULL AND revoked_at IS NULL`, now, userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=$1 WHERE user_id=$2 AND revoked_at IS NULL`, now, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// Login verifies and locks the current credential, then creates the session in
// the same transaction. Reset and deletion lock the user row too, so an old
// password cannot win a race and mint a session after either mutation.
func (s *Service) Login(ctx context.Context, ip, emailAddr, pass string) (*auth.User, *auth.Session, string, error) {
	emailAddr = auth.NormalizeEmail(emailAddr)
	if err := s.allow(ctx, "password.login", ip, emailAddr); err != nil {
		return nil, nil, "", err
	}
	if err := Validate(pass); err != nil {
		return nil, nil, "", ErrInvalidCredentials
	}
	if s.identityAllowed != nil && !s.identityAllowed(emailAddr) {
		return nil, nil, "", ErrInvalidCredentials
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, "", err
	}
	defer tx.Rollback()
	var u auth.User
	var verified sql.NullTime
	// User is always locked first. Reset and account deletion lock this row
	// before touching credentials, which gives all credential/session mutations
	// one lock order and prevents planner-dependent lock inversions.
	err = tx.QueryRowContext(ctx, `SELECT id,email,COALESCE(display_name,''),COALESCE(avatar_url,''),status,email_verified_at,created_at,updated_at FROM users WHERE lower(email)=lower($1) AND deleted_at IS NULL FOR UPDATE`, emailAddr).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Status, &verified, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		burnPasswordCheck(pass)
		return nil, nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("lock password user: %w", err)
	}
	if !u.Active() || !verified.Valid {
		burnPasswordCheck(pass)
		return nil, nil, "", ErrInvalidCredentials
	}
	var hash string
	err = tx.QueryRowContext(ctx, `SELECT password_hash FROM password_credentials WHERE user_id=$1 FOR UPDATE`, u.ID).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		burnPasswordCheck(pass)
		return nil, nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("lock password credential: %w", err)
	}
	u.EmailVerifiedAt = &verified.Time
	u.HasPassword = true
	ok, err := Verify(hash, pass)
	if err != nil {
		if errors.Is(err, ErrMalformedHash) {
			// Deliberately static: diagnostics must not include a password,
			// credential hash, identity, or any user-controlled value.
			fmt.Fprintln(os.Stderr, "password: malformed stored credential hash")
		}
		return nil, nil, "", ErrInvalidCredentials
	}
	if !ok {
		return nil, nil, "", ErrInvalidCredentials
	}
	now := s.clock.Now().UTC()
	raw, th, err := auth.NewTokenAndHash()
	if err != nil {
		return nil, nil, "", err
	}
	expires := now.Add(s.sessionLifetime)
	sid := uuid.New()
	if _, err = tx.ExecContext(ctx, `UPDATE users SET last_login_at=$1,updated_at=$1 WHERE id=$2 AND status='active' AND deleted_at IS NULL`, now, u.ID); err != nil {
		return nil, nil, "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,created_at,expires_at) VALUES($1,$2,$3,$4,$5)`, sid, u.ID, th, now, expires); err != nil {
		return nil, nil, "", err
	}
	if err = tx.Commit(); err != nil {
		return nil, nil, "", err
	}
	u.LastLoginAt = &now
	return &u, &auth.Session{ID: sid, UserID: u.ID, CreatedAt: now, ExpiresAt: expires}, raw, nil
}
