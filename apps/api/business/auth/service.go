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
)

// Config holds the auth service configuration.
type Config struct {
	Environment            string
	BaseURL                string
	MagicLinkPath          string
	OAuthRedirectURI       string
	OAuthRedirectAllowlist []string
	SessionLifetime        time.Duration
	MagicLinkLifetime      time.Duration
	OAuthStateLifetime     time.Duration
	Cookie                 CookieConfig
	RateLimit              RateLimitConfig
}

// RateLimitConfig configures rate limits.
type RateLimitConfig struct {
	MagicRequestWindow  time.Duration
	MagicRequestLimit   int
	MagicConsumeWindow  time.Duration
	MagicConsumeLimit   int
	OAuthStartWindow    time.Duration
	OAuthStartLimit     int
	OAuthCallbackWindow time.Duration
	OAuthCallbackLimit  int
	LogoutWindow        time.Duration
	LogoutLimit         int
}

// Service implements magic-link, OAuth, session, and logout lifecycle.
type Service struct {
	repo         Repository
	emailSender  email.Sender
	oauth        OAuthProvider
	clock        clock.Clock
	limiter      RateLimiter
	cfg          Config
	killSwitches *KillSwitches
}

// NewService creates an auth service.
func NewService(repo Repository, emailSender email.Sender, oauth OAuthProvider, c clock.Clock, limiter RateLimiter, cfg Config) *Service {
	if cfg.SessionLifetime == 0 {
		cfg.SessionLifetime = 30 * 24 * time.Hour
	}
	if cfg.MagicLinkLifetime == 0 {
		cfg.MagicLinkLifetime = 15 * time.Minute
	}
	if cfg.OAuthStateLifetime == 0 {
		cfg.OAuthStateLifetime = 10 * time.Minute
	}
	if cfg.MagicLinkPath == "" {
		cfg.MagicLinkPath = "/auth/magic"
	}
	return &Service{repo: repo, emailSender: emailSender, oauth: oauth, clock: c, limiter: limiter, cfg: cfg}
}

// OAuthConfigured reports whether an OAuth provider is wired.
func (s *Service) OAuthConfigured() bool { return s.oauth != nil }

// Clock returns the service clock for HTTP-layer cookie expiry calculations.
func (s *Service) Clock() clock.Clock { return s.clock }

// OAuthStateLifetime returns the configured OAuth state lifetime.
func (s *Service) OAuthStateLifetime() time.Duration { return s.cfg.OAuthStateLifetime }

// RequestMagicLink creates a single-use magic link and sends it to the email.
func (s *Service) RequestMagicLink(ctx context.Context, clientIP, emailAddr string) error {
	if s.killSwitches != nil && !s.killSwitches.MagicLinkEnabled {
		return ErrMagicLinkDisabled
	}
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
	if s.killSwitches != nil && !s.killSwitches.MagicLinkEnabled {
		return nil, nil, "", ErrMagicLinkDisabled
	}
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
		if s.killSwitches != nil && !s.killSwitches.NewSignupsEnabled {
			return nil, nil, "", ErrSignupsDisabled
		}
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

	user, session, token, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, nil, "", err
	}
	return user, session, token, nil
}

// OAuthStart begins a Google OAuth flow. It returns the provider authorization
// URL and the raw state token to set as a cookie. The appReturnURL is the
// application destination the user should be sent to after successful callback.
func (s *Service) OAuthStart(ctx context.Context, clientIP, appReturnURL string) (string, string, error) {
	if s.killSwitches != nil && !s.killSwitches.OAuthEnabled {
		return "", "", ErrOAuthDisabled
	}
	if !s.OAuthConfigured() {
		return "", "", ErrOAuthNotConfigured
	}
	if !s.allowedAppReturnURL(appReturnURL) {
		return "", "", ErrOAuthProviderFailed
	}

	allowed, err := s.limiter.Allow(ctx, KeyForIP("oauth.start", clientIP))
	if err != nil {
		return "", "", fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return "", "", ErrRateLimited
	}

	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.OAuthStateLifetime)
	token, hash, err := generateToken()
	if err != nil {
		return "", "", err
	}
	if _, err := s.repo.CreateOAuthState(ctx, hash, s.cfg.Environment, "google", appReturnURL, now, expiresAt); err != nil {
		return "", "", fmt.Errorf("create oauth state: %w", err)
	}

	authURL := s.oauth.AuthURL(token, s.cfg.OAuthRedirectURI)
	return authURL, token, nil
}

// OAuthCallback validates the OAuth state, verifies the code with the provider,
// and creates/links the internal identity and session. It returns the
// application return URL that was supplied at the start of the flow.
func (s *Service) OAuthCallback(ctx context.Context, clientIP, code, state, cookieState string) (*User, *Session, string, string, error) {
	if s.killSwitches != nil && !s.killSwitches.OAuthEnabled {
		return nil, nil, "", "", ErrOAuthDisabled
	}
	if !s.OAuthConfigured() {
		return nil, nil, "", "", ErrOAuthNotConfigured
	}
	allowed, err := s.limiter.Allow(ctx, KeyForIP("oauth.callback", clientIP))
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("rate limit: %w", err)
	}
	if !allowed {
		return nil, nil, "", "", ErrRateLimited
	}

	state = strings.TrimSpace(state)
	cookieState = strings.TrimSpace(cookieState)
	if state == "" || cookieState == "" || subtleCompare(state, cookieState) != 1 {
		return nil, nil, "", "", ErrInvalidOAuthState
	}

	_, hash, err := tokenAndHash(state)
	if err != nil {
		return nil, nil, "", "", ErrInvalidOAuthState
	}
	ostate, err := s.repo.GetOAuthStateByTokenHash(ctx, hash)
	if err != nil {
		return nil, nil, "", "", ErrInvalidOAuthState
	}
	now := s.clock.Now()
	if ostate.Environment != s.cfg.Environment || ostate.Provider != "google" || !ostate.Valid(now) {
		return nil, nil, "", "", ErrInvalidOAuthState
	}
	if err := s.repo.ConsumeOAuthState(ctx, ostate.ID, now); err != nil {
		return nil, nil, "", "", fmt.Errorf("consume oauth state: %w", err)
	}

	identity, err := s.oauth.Verify(ctx, code, state, s.cfg.OAuthRedirectURI)
	if err != nil {
		return nil, nil, "", "", ErrOAuthProviderFailed
	}
	if identity.Subject == "" || identity.Email == "" || !identity.EmailVerified {
		return nil, nil, "", "", ErrOAuthProviderFailed
	}

	emailAddr := normalizeEmail(identity.Email)
	if emailAddr == "" {
		return nil, nil, "", "", ErrOAuthProviderFailed
	}

	// Try to find an existing provider identity first.
	user, err := s.resolveOAuthIdentity(ctx, identity, emailAddr, now)
	if err != nil {
		return nil, nil, "", "", err
	}
	if !user.Active() {
		return nil, nil, "", "", ErrUserDisabled
	}

	_, session, token, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, nil, "", "", err
	}
	return user, session, token, ostate.AppReturnURL, nil
}

func (s *Service) resolveOAuthIdentity(ctx context.Context, identity *OAuthIdentity, emailAddr string, now time.Time) (*User, error) {
	// Existing provider identity takes precedence; this prevents email takeover.
	if ext, err := s.repo.GetExternalIdentity(ctx, "google", identity.Subject); err == nil {
		user, err := s.repo.GetUserByID(ctx, ext.UserID)
		if err != nil {
			return nil, ErrOAuthProviderFailed
		}
		return user, nil
	}

	// No existing provider identity. Link to an existing user by verified email
	// if one exists, otherwise create a new verified user.
	user, err := s.repo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		if s.killSwitches != nil && !s.killSwitches.NewSignupsEnabled {
			return nil, ErrSignupsDisabled
		}
		user, err = s.repo.CreateUser(ctx, emailAddr, &now)
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	}
	if _, err := s.repo.CreateExternalIdentity(ctx, user.ID, "google", identity.Subject, emailAddr, true); err != nil {
		return nil, fmt.Errorf("create external identity: %w", err)
	}
	return user, nil
}

func (s *Service) allowedAppReturnURL(uri string) bool {
	if uri == "" {
		return false
	}
	for _, allowed := range s.cfg.OAuthRedirectAllowlist {
		if uri == allowed {
			return true
		}
	}
	return false
}

func (s *Service) issueSession(ctx context.Context, user *User) (*User, *Session, string, error) {
	now := s.clock.Now()
	expiresAt := now.Add(s.cfg.SessionLifetime)
	token, hash, err := generateToken()
	if err != nil {
		return nil, nil, "", err
	}
	if err := s.repo.UpdateUserLastLogin(ctx, user.ID, now); err != nil {
		return nil, nil, "", fmt.Errorf("update last login: %w", err)
	}
	user.LastLoginAt = &now
	session, err := s.repo.CreateSession(ctx, user.ID, hash, now, expiresAt)
	if err != nil {
		return nil, nil, "", err
	}
	return user, session, token, nil
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

// Cleanup removes expired and revoked sessions, magic links, and oauth states.
func (s *Service) Cleanup(ctx context.Context) error {
	now := s.clock.Now()
	if _, err := s.repo.CleanupExpiredSessions(ctx, now); err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}
	if _, err := s.repo.CleanupExpiredMagicLinks(ctx, now); err != nil {
		return fmt.Errorf("cleanup magic links: %w", err)
	}
	if _, err := s.repo.CleanupExpiredOAuthStates(ctx, now); err != nil {
		return fmt.Errorf("cleanup oauth states: %w", err)
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

// OAuthStateCookieName returns the configured OAuth state cookie name.
func (s *Service) OAuthStateCookieName() string { return s.cfg.Cookie.OAuthStateName }

// IssueCSRFCookie returns a new double-submit CSRF token and its cookie.
func (s *Service) IssueCSRFCookie() (string, *http.Cookie) {
	return CSRFToken(s.cfg.Cookie)
}

// OAuthStateCookie returns the short-lived OAuth state cookie.
func (s *Service) OAuthStateCookie(token string, expiresAt time.Time) *http.Cookie {
	return OAuthStateCookie(s.cfg.Cookie, token, expiresAt)
}

// ClearOAuthStateCookie returns a cookie that deletes the OAuth state cookie.
func (s *Service) ClearOAuthStateCookie() *http.Cookie {
	return ClearOAuthStateCookie(s.cfg.Cookie)
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

// subtleCompare wraps subtle.ConstantTimeCompare using a 1/0 result.
func subtleCompare(a, b string) int {
	if len(a) != len(b) {
		return 0
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	if v == 0 {
		return 1
	}
	return 0
}
