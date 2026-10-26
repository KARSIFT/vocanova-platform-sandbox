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

func TestOAuthSessionCookieKeepsSecurityAndScopeForBothChoices(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, Config{Cookie: CookieConfig{
		Name: "vocanova_session", Domain: "production.vocanova.site",
		Secure: true, SameSite: http.SameSiteStrictMode,
	}})
	expires := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	for _, persistent := range []bool{false, true} {
		cookie := svc.OAuthSessionCookie("synthetic-test-value", &Session{ExpiresAt: expires, CookiePersistent: persistent})
		assert.Equal(t, "vocanova_session", cookie.Name)
		assert.Equal(t, "production.vocanova.site", cookie.Domain)
		assert.Equal(t, "/", cookie.Path)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		assert.Equal(t, !persistent, cookie.Expires.IsZero())
	}
}

// TestEnvironmentCookieNames pins the per-tier cookie names. Staging and
// production share the parent cookie domain, so their names must differ
// from each other and from the legacy names local/test environments use.
func TestEnvironmentCookieNames(t *testing.T) {
	cases := []struct {
		environment string
		session     string
		csrf        string
	}{
		{"staging", "vocanova_staging_session", "vocanova_staging_csrf"},
		{"production", "vocanova_production_session", "vocanova_production_csrf"},
		{" Production\n", "vocanova_production_session", "vocanova_production_csrf"},
		{"local", LegacySessionCookieName, LegacyCSRFCookieName},
		{"test", LegacySessionCookieName, LegacyCSRFCookieName},
		{"", LegacySessionCookieName, LegacyCSRFCookieName},
		{"prod", LegacySessionCookieName, LegacyCSRFCookieName},
		{"staging-preview", LegacySessionCookieName, LegacyCSRFCookieName},
	}
	for _, tc := range cases {
		session, csrf := EnvironmentCookieNames(tc.environment)
		assert.Equal(t, tc.session, session, "session cookie name for %q", tc.environment)
		assert.Equal(t, tc.csrf, csrf, "csrf cookie name for %q", tc.environment)
	}
}

func TestEnvironmentCookieNamesNeverCollideAcrossTiers(t *testing.T) {
	stagingSession, stagingCSRF := EnvironmentCookieNames("staging")
	productionSession, productionCSRF := EnvironmentCookieNames("production")
	seen := map[string]bool{}
	for _, name := range []string{
		stagingSession, stagingCSRF, productionSession, productionCSRF,
		LegacySessionCookieName, LegacyCSRFCookieName, "vocanova_oauth_state",
	} {
		assert.False(t, seen[name], "cookie name %q is reused", name)
		seen[name] = true
	}
}

// TestCookieHelpersUseEnvironmentCookieNames confirms the cookie writers
// emit whatever names the config carries, so the production wiring's tier
// names reach Set-Cookie while the OAuth state cookie stays host-only.
func TestCookieHelpersUseEnvironmentCookieNames(t *testing.T) {
	session, csrf := EnvironmentCookieNames("staging")
	cfg := CookieConfig{
		Name: session, CSRName: csrf, OAuthStateName: "vocanova_oauth_state",
		Domain: ".vocanova.site",
	}
	expires := time.Now().Add(time.Hour)

	assert.Equal(t, "vocanova_staging_session", SessionCookie(cfg, "v", expires).Name)
	assert.Equal(t, "vocanova_staging_session", ClearSessionCookie(cfg).Name)
	_, csrfCookie := CSRFToken(cfg)
	assert.Equal(t, "vocanova_staging_csrf", csrfCookie.Name)
	assert.Equal(t, "vocanova_staging_csrf", ClearCSRFCookie(cfg).Name)
	stateCookie := OAuthStateCookie(cfg, "v", expires)
	assert.Equal(t, "vocanova_oauth_state", stateCookie.Name)
	assert.Equal(t, "", stateCookie.Domain)
}
