package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	csrfTokenLength = 32
)

// generateToken returns a URL-safe base64 token and its SHA-256 hash.
func generateToken() (string, []byte, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}
	hash := sha256.Sum256(b)
	return base64.URLEncoding.EncodeToString(b), hash[:], nil
}

// NewTokenAndHash is the exported wrapper around generateToken used by
// downstream business modules (notably accounts/email_change.go for
// VOC-031-T03). It returns a 32-random-byte URL-safe base64 token and
// its SHA-256 hash. The raw token is the only form the caller should
// ever embed in a confirmation link; only the hash is persisted.
func NewTokenAndHash() (string, []byte, error) {
	return generateToken()
}

// TokenAndHash decodes a base64 token (the form the requester supplies
// at consume time) and returns its raw bytes plus the SHA-256 hash
// the persistence layer stored. Exported so downstream modules
// (email-change in VOC-031-T03) can use the same secret-handling
// discipline as the magic-link flow without duplicating the
// base64-URL-decode + SHA-256 logic. The hash returned here is
// byte-identical to what NewTokenAndHash would have produced for
// the same raw bytes, so a magic-link row and an email-change-link
// row written from the same raw token are indistinguishable to a
// future audit.
//
// Returns an error on an empty token, non-base64 input, or a
// decoded length that is not exactly 32 bytes.
func TokenAndHash(token string) ([]byte, []byte, error) {
	return tokenAndHash(token)
}

func hashTokenString(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// normalizeEmail lowercases and trims whitespace.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeEmail is the exported form of normalizeEmail, for callers
// outside this package (e.g. the production-wiring layer building a
// KillSwitches.SignupAllowlist from an env var) that need to
// normalize an email the same way this package does internally.
func NormalizeEmail(email string) string {
	return normalizeEmail(email)
}

// Legacy cookie names, kept for local, test and unknown environments so
// fixtures and local development keep their existing behavior.
const (
	LegacySessionCookieName = "vocanova_session"
	LegacyCSRFCookieName    = "vocanova_csrf"
)

// EnvironmentCookieNames returns the session and CSRF cookie names for a
// deployment environment. Staging and production both scope these cookies
// to the shared parent domain, so each deployed tier needs its own names
// to keep a sign-in on one tier from overwriting the other's session.
// Matching ignores case and surrounding whitespace; any other environment
// (local, test, unknown) keeps the legacy names. The host-only OAuth state
// cookie keeps its single name and is not covered here.
func EnvironmentCookieNames(environment string) (sessionName, csrfName string) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "staging":
		return "vocanova_staging_session", "vocanova_staging_csrf"
	case "production":
		return "vocanova_production_session", "vocanova_production_csrf"
	default:
		return LegacySessionCookieName, LegacyCSRFCookieName
	}
}

// CookieConfig holds session and CSRF cookie settings.
type CookieConfig struct {
	Name           string
	CSRName        string // double-submit cookie name
	OAuthStateName string // OAuth state cookie name
	Domain         string

	// OAuthStateDomain is the Domain attribute for the OAuth state cookie,
	// intentionally separate from Domain. The state cookie only needs to
	// round-trip between OAuthStart and the OAuth callback, both served
	// from the API's own host - unlike the session/CSRF cookies, it has
	// no reason to be scoped to a different (e.g. web-app) host. Leave
	// empty for host-only scoping (no explicit Domain attribute), which
	// is correct whenever the API host differs from Domain, including
	// sibling subdomains (e.g. api-X.example.com vs X.example.com) where
	// a cookie scoped to Domain would never be sent back to the API host
	// by a real browser. See the T03 rehearsal finding this field fixes.
	OAuthStateDomain string

	Secure   bool
	SameSite http.SameSite
}

// SessionCookie writes the session bearer cookie.
func SessionCookie(cfg CookieConfig, token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Domain:   cfg.Domain,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
}

// ClearSessionCookie returns a cookie that deletes the session cookie.
func ClearSessionCookie(cfg CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Domain:   cfg.Domain,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
}

// ReadSessionCookie reads the session bearer from the request.
func ReadSessionCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

// CSRFToken returns a new random double-submit CSRF token and the cookie that
// should be sent to the client. The client must echo the value in the
// X-CSRF-Token header.
func CSRFToken(cfg CookieConfig) (string, *http.Cookie) {
	b := make([]byte, csrfTokenLength)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", nil
	}
	value := base64.URLEncoding.EncodeToString(b)
	return value, &http.Cookie{
		Name:     cfg.CSRName,
		Value:    value,
		Domain:   cfg.Domain,
		Path:     "/",
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
}

// ClearCSRFCookie returns a cookie that deletes the double-submit CSRF cookie.
func ClearCSRFCookie(cfg CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.CSRName,
		Value:    "",
		Domain:   cfg.Domain,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
}

// ValidateCSRF compares the double-submit cookie value with the request header.
func ValidateCSRF(cookieValue, headerValue string) bool {
	if cookieValue == "" || headerValue == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieValue), []byte(headerValue)) == 1
}

// OAuthStateCookie writes the OAuth state cookie.
func OAuthStateCookie(cfg CookieConfig, token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.OAuthStateName,
		Value:    token,
		Domain:   cfg.OAuthStateDomain,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearOAuthStateCookie returns a cookie that deletes the OAuth state cookie.
func ClearOAuthStateCookie(cfg CookieConfig) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.OAuthStateName,
		Value:    "",
		Domain:   cfg.OAuthStateDomain,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClientIP reads the single address normalized by the trusted reverse proxy.
// Public ingress must overwrite X-Real-IP; untrusted forwarding chains are never
// used as rate-limit identities. Without a valid proxy value, use the TCP peer.
func ClientIP(remoteAddr, realIP string) string {
	if ip := net.ParseIP(strings.TrimSpace(realIP)); ip != nil {
		return ip.String()
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.Trim(remoteAddr, "[]")
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return "unknown"
}
