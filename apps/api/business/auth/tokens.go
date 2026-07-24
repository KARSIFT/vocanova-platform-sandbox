package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
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

func hashTokenString(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// normalizeEmail lowercases and trims whitespace.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// CookieConfig holds session and CSRF cookie settings.
type CookieConfig struct {
	Name     string
	CSRName  string // double-submit cookie name
	Domain   string
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

// ValidateCSRF compares the double-submit cookie value with the request header.
func ValidateCSRF(cookieValue, headerValue string) bool {
	if cookieValue == "" || headerValue == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieValue), []byte(headerValue)) == 1
}

// clientIP returns a best-effort client IP for rate-limiting from a Huma
// context or a raw HTTP request.
func clientIP(remoteAddr string, headers http.Header) string {
	if xff := headers.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := headers.Get("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, _ := strings.Cut(remoteAddr, ":")
	return host
}
