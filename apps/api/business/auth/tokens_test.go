package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestOAuthStateCookie_DomainIsIndependentOfSessionDomain covers the T03
// live-rehearsal finding: the OAuth state cookie must not reuse
// CookieConfig.Domain (the session/CSRF cookie's domain, typically the web
// app's hostname). It only round-trips between the API host's own
// OAuthStart/OAuthCallback endpoints, so a differing (or empty, host-only)
// OAuthStateDomain must be respected independently of Domain.
func TestOAuthStateCookie_DomainIsIndependentOfSessionDomain(t *testing.T) {
	cfg := CookieConfig{
		OAuthStateName:   "vocanova_oauth_state",
		Domain:           "production.vocanova.site", // web app's hostname
		OAuthStateDomain: "",                         // host-only: must NOT inherit Domain
		Secure:           true,
		SameSite:         http.SameSiteStrictMode,
	}

	cookie := OAuthStateCookie(cfg, "test-token", time.Now().Add(10*time.Minute))
	assert.Equal(t, "", cookie.Domain, "OAuth state cookie must not be scoped to the session/web-app Domain")
	assert.Equal(t, "vocanova_oauth_state", cookie.Name)
	assert.Equal(t, "test-token", cookie.Value)

	clearCookie := ClearOAuthStateCookie(cfg)
	assert.Equal(t, "", clearCookie.Domain, "clearing the OAuth state cookie must use the same domain scoping as setting it")
}

// TestOAuthStateCookie_UsesConfiguredOAuthStateDomain covers the case where
// an explicit OAuthStateDomain is set (e.g. a future deployment where the
// API and web app genuinely share a domain) - the field must still be
// honored, not silently ignored in favor of Domain.
func TestOAuthStateCookie_UsesConfiguredOAuthStateDomain(t *testing.T) {
	cfg := CookieConfig{
		OAuthStateName:   "vocanova_oauth_state",
		Domain:           "production.vocanova.site",
		OAuthStateDomain: "api-production.vocanova.site",
	}

	cookie := OAuthStateCookie(cfg, "test-token", time.Now().Add(10*time.Minute))
	assert.Equal(t, "api-production.vocanova.site", cookie.Domain)
	assert.NotEqual(t, cfg.Domain, cookie.Domain)
}

// TestSessionCookie_UnaffectedByOAuthStateDomain confirms adding
// OAuthStateDomain did not change SessionCookie/CSRFToken's existing
// behavior - they must keep using Domain, unchanged.
func TestSessionCookie_UnaffectedByOAuthStateDomain(t *testing.T) {
	cfg := CookieConfig{
		Name:             "vocanova_session",
		Domain:           "production.vocanova.site",
		OAuthStateDomain: "", // deliberately different from Domain
	}

	cookie := SessionCookie(cfg, "session-token", time.Now().Add(30*24*time.Hour))
	assert.Equal(t, "production.vocanova.site", cookie.Domain)
}

// TestCSRFToken_UnaffectedByOAuthStateDomain covers the reviewed Low
// finding: CSRFToken also uses cfg.Domain and must be equally unaffected
// by the new OAuthStateDomain field, not just SessionCookie.
func TestCSRFToken_UnaffectedByOAuthStateDomain(t *testing.T) {
	cfg := CookieConfig{
		CSRName:          "vocanova_csrf",
		Domain:           "production.vocanova.site",
		OAuthStateDomain: "",
	}

	_, cookie := CSRFToken(cfg)
	assert.Equal(t, "production.vocanova.site", cookie.Domain)
}

func TestClearCSRFCookieMatchesCSRFCookieScope(t *testing.T) {
	cfg := CookieConfig{
		CSRName:  "vocanova_csrf",
		Domain:   "production.vocanova.site",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	clearCookie := ClearCSRFCookie(cfg)

	assert.Equal(t, cfg.CSRName, clearCookie.Name)
	assert.Equal(t, cfg.Domain, clearCookie.Domain)
	assert.Equal(t, "/", clearCookie.Path)
	assert.Equal(t, -1, clearCookie.MaxAge)
	assert.False(t, clearCookie.HttpOnly)
	assert.True(t, clearCookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, clearCookie.SameSite)
}
