package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
)

// Config holds the auth service configuration.
type Config struct {
	Environment       string
	BaseURL           string
	MagicLinkPath     string
	SessionLifetime   time.Duration
	MagicLinkLifetime time.Duration
	Cookie            CookieConfig
	RateLimit         RateLimitConfig
}

// RateLimitConfig configures rate limits.
type RateLimitConfig struct {
	MagicRequestWindow time.Duration
	MagicRequestLimit  int
	MagicConsumeWindow time.Duration
	MagicConsumeLimit  int
	LogoutWindow       time.Duration
	LogoutLimit        int
}

// Service implements magic-link, session, and logout lifecycle.
type Service struct {
	repo        Repository
	emailSender email.Sender
	clock       clock.Clock
	limiter     RateLimiter
	cfg         Config
}

// NewService creates an auth service.
func NewService(repo Repository, emailSender email.Sender, c clock.Clock, limiter RateLimiter, cfg Config) *Service {
	if cfg.SessionLifetime == 0 {
		cfg.SessionLifetime = 30 * 24 * time.Hour
	}
	if cfg.MagicLinkLifetime == 0 {
		cfg.MagicLinkLifetime = 15 * time.Minute
	}
	if cfg.MagicLinkPath == "" {
		cfg.MagicLinkPath = "/auth/magic"
	}
	return &Service{repo: repo, emailSender: emailSender, clock: c, limiter: limiter, cfg: cfg}
}

// RequestMagicLink creates a single-use magic link and sends it to the email.
func (s *Service) RequestMagicLink(ctx context.Context, clientIP, emailAddr string) error {
	allowed, err := s.limiter.Allow(ctx, KeyForIP("magic.request", clientIP))
	if err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}

	emailAddr = normalizeEmail(emailAddr)
	if emailAddr == "" {
		// Return the same generic result to avoid enumeration.
		return nil
	}

	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.MagicLinkLifetime)
	token, hash, err := generateToken()
	if err != nil {
		return err
	}

	if _, err := s.repo.CreateMagicLink(ctx, emailAddr, hash, s.cfg.Environment, now, expiresAt); err != nil {
		return fmt.Errorf("create magic link: %w", err)
	}

	link := s.magicLinkURL(token, emailAddr)
	msg := email.Message{
		To:       []email.Address{{Email: emailAddr}},
		Subject:  "Sign in to Vocanova",
		BodyText: fmt.Sprintf("Use this single-use link to sign in:\n\n%s\n\nIt expires in 15 minutes.", link),
		BodyHTML: fmt.Sprintf("<p>Use this single-use link to sign in:</p><p><a href=\"%s\">%s</a></p><p>It expires in 15 minutes.</p>", link, link),
	}
	if err := s.emailSender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (s *Service) magicLinkURL(token, email string) string {
	u, _ := url.Parse(s.cfg.BaseURL)
	u.Path = s.cfg.MagicLinkPath
	q := u.Query()
	q.Set("token", token)
	q.Set("email", email)
	u.RawQuery = q.Encode()
	return u.String()
}

// ConsumeMagicLink verifies a magic link, creates/links the user, and issues a
// session. The returned token is the opaque bearer that callers set as a cookie.
func (s *Service) ConsumeMagicLink(ctx context.Context, clientIP, token, emailAddr string) (*User, *Session, string, error) {
	allowed, err := s.limiter.Allow(ctx, KeyForIP("magic.consume", clientIP))
	if err != nil {
		return nil, nil, "", fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return nil, nil, "", ErrRateLimited
	}

	emailAddr = normalizeEmail(emailAddr)
	token = strings.TrimSpace(token)
	if token == "" || emailAddr == "" {
		return nil, nil, "", ErrInvalidMagicLink
	}

	_, hash, err := tokenAndHash(token)
	if err != nil {
		return nil, nil, "", ErrInvalidMagicLink
	}

	link, err := s.repo.GetMagicLinkByTokenHash(ctx, hash)
	if err != nil {
		return nil, nil, "", ErrInvalidMagicLink
	}
	if link.Environment != s.cfg.Environment || link.Email != emailAddr {
		return nil, nil, "", ErrInvalidMagicLink
	}
	now := s.clock.Now()
	if !link.Valid(now) {
		return nil, nil, "", ErrInvalidMagicLink
	}

	user, err := s.repo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		// Create a verified user from the magic link.
		user, err = s.repo.CreateUser(ctx, emailAddr, &now)
		if err != nil {
			return nil, nil, "", fmt.Errorf("create user: %w", err)
		}
	}
	if !user.Active() {
		return nil, nil, "", ErrUserDisabled
	}

	if err := s.repo.AttachMagicLinkUser(ctx, link.ID, user.ID); err != nil {
		return nil, nil, "", fmt.Errorf("attach magic link: %w", err)
	}
	if err := s.repo.ConsumeMagicLink(ctx, link.ID, now); err != nil {
		return nil, nil, "", fmt.Errorf("consume magic link: %w", err)
	}

	if err := s.repo.UpdateUserLastLogin(ctx, user.ID, now); err != nil {
		return nil, nil, "", fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now

	session, rawToken, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, nil, "", err
	}
	return user, session, rawToken, nil
}

func (s *Service) createSession(ctx context.Context, userID uuid.UUID) (*Session, string, error) {
	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.SessionLifetime)
	token, hash, err := generateToken()
	if err != nil {
		return nil, "", err
	}
	session, err := s.repo.CreateSession(ctx, userID, hash, now, expiresAt)
	if err != nil {
		return nil, "", err
	}
	return session, token, nil
}

// ValidateSession returns the active user for a session bearer token.
func (s *Service) ValidateSession(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, ErrAuthenticationRequired
	}
	_, hash, err := tokenAndHash(token)
	if err != nil {
		return nil, ErrAuthenticationRequired
	}
	session, err := s.repo.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrAuthenticationRequired
	}
	now := s.clock.Now()
	if !session.Valid(now) {
		return nil, ErrAuthenticationRequired
	}
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrAuthenticationRequired
	}
	if !user.Active() {
		return nil, ErrUserDisabled
	}
	return user, nil
}

// Logout revokes the session matching the bearer token.
func (s *Service) Logout(ctx context.Context, token string) error {
	allowed, err := s.limiter.Allow(ctx, KeyForSession("logout", token))
	if err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return ErrRateLimited
	}
	if token == "" {
		return nil
	}
	_, hash, err := tokenAndHash(token)
	if err != nil {
		return nil
	}
	session, err := s.repo.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil
	}
	now := s.clock.Now()
	if !session.Valid(now) {
		return nil
	}
	return s.repo.RevokeSession(ctx, session.ID, now)
}

// Cleanup removes expired and revoked sessions and magic links.
func (s *Service) Cleanup(ctx context.Context) error {
	now := s.clock.Now()
	if _, err := s.repo.CleanupExpiredSessions(ctx, now); err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}
	if _, err := s.repo.CleanupExpiredMagicLinks(ctx, now); err != nil {
		return fmt.Errorf("cleanup magic links: %w", err)
	}
	return nil
}

// SessionCookie writes a session cookie for the given token.
func (s *Service) SessionCookie(token string, expiresAt time.Time) *http.Cookie {
	return SessionCookie(s.cfg.Cookie, token, expiresAt)
}

// ClearSessionCookie writes a cookie that deletes the session.
func (s *Service) ClearSessionCookie() *http.Cookie {
	return ClearSessionCookie(s.cfg.Cookie)
}

// ReadSessionCookie reads the session bearer from the request.
func (s *Service) ReadSessionCookie(r *http.Request) string {
	return ReadSessionCookie(r, s.cfg.Cookie.Name)
}

// SessionCookieName returns the configured session cookie name.
func (s *Service) SessionCookieName() string { return s.cfg.Cookie.Name }

// CSRFCookieName returns the configured CSRF double-submit cookie name.
func (s *Service) CSRFCookieName() string { return s.cfg.Cookie.CSRName }

// IssueCSRFCookie returns a new double-submit CSRF token and its cookie.
func (s *Service) IssueCSRFCookie() (string, *http.Cookie) {
	return CSRFToken(s.cfg.Cookie)
}

// ValidateCSRF validates the double-submit CSRF token.
func (s *Service) ValidateCSRF(cookieValue, headerValue string) bool {
	return ValidateCSRF(cookieValue, headerValue)
}

// tokenAndHash decodes a base64 token and returns its SHA-256 hash.
func tokenAndHash(token string) ([]byte, []byte, error) {
	b, err := base64URLDecode(token)
	if err != nil {
		return nil, nil, err
	}
	if len(b) != 32 {
		return nil, nil, errors.New("invalid token length")
	}
	h := sha256.Sum256(b)
	return b, h[:], nil
}

// base64URLDecode decodes a URL-safe base64 string without padding validation.
func base64URLDecode(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("empty token")
	}
	return base64.URLEncoding.DecodeString(s)
}

// Public errors returned by the service.
var (
	ErrInvalidMagicLink       = errors.New("invalid or expired magic link")
	ErrAuthenticationRequired = errors.New("authentication required")
	ErrUserDisabled           = errors.New("user disabled")
	ErrRateLimited            = errors.New("rate limited")
)
