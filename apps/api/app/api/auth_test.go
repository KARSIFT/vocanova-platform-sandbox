package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAuthAPI(t *testing.T) (huma.API, *auth.Service, *auth.MemoryRepository, *email.Fake, *clock.Fixed) {
	t.Helper()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := auth.NewMemoryRepository()
	fake := &email.Fake{}
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	svc := auth.NewService(repo, fake, c, limiter, auth.Config{
		Environment:       "test",
		BaseURL:           "https://test.example.com",
		MagicLinkPath:     "/auth/magic",
		SessionLifetime:   30 * 24 * time.Hour,
		MagicLinkLifetime: 15 * time.Minute,
		Cookie: auth.CookieConfig{
			Name:     "vocanova_session",
			CSRName:  "vocanova_csrf",
			Domain:   "",
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		},
		RateLimit: auth.RateLimitConfig{
			MagicRequestWindow: time.Hour,
			MagicRequestLimit:  10,
			MagicConsumeWindow: time.Hour,
			MagicConsumeLimit:  10,
			LogoutWindow:       time.Hour,
			LogoutLimit:        10,
		},
	})

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	RegisterAuth(api, svc)
	return api, svc, repo, fake, c
}

func TestRequestMagicLinkEndpointReturns204(t *testing.T) {
	api, _, _, fake, _ := testAuthAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	msg, ok := fake.Last()
	require.True(t, ok)
	assert.Equal(t, "user@example.com", msg.To[0].Email)
}

func TestRequestMagicLinkEndpointInvalidEmailStill422(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"not-an-email"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	// Huma validates format:email, so it should return 422 for invalid format.
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestConsumeMagicLinkEndpointSetsCookiesAndReturnsUser(t *testing.T) {
	api, _, _, fake, _ := testAuthAPI(t)

	// Request a link.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	// Consume does not require CSRF (the secret token itself is the credential).
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links/consume", strings.NewReader(`{"token":"`+rawToken+`","email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user@example.com")
	setCookies := w.Result().Cookies()
	require.True(t, hasCookie(setCookies, "vocanova_session"), "session cookie missing")
	require.True(t, hasCookie(setCookies, "vocanova_csrf"), "csrf cookie missing")

	// Replaying the link now fails.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links/consume", strings.NewReader(`{"token":"`+rawToken+`","email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestConsumeMagicLinkEndpointRejectsWrongEmail(t *testing.T) {
	api, svc, _, fake, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	w = httptest.NewRecorder()
	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links/consume", strings.NewReader(`{"token":"`+rawToken+`","email":"other@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogoutEndpointClearsCookie(t *testing.T) {
	api, svc, _, fake, _ := testAuthAPI(t)

	// Request and consume a link to get a session.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	w = httptest.NewRecorder()
	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links/consume", strings.NewReader(`{"token":"`+rawToken+`","email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	sessionCookie := findCookie(w.Result().Cookies(), "vocanova_session")
	require.NotNil(t, sessionCookie)

	// Logout without CSRF is rejected.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	// Logout with CSRF clears the session cookie.
	w = httptest.NewRecorder()
	csrfToken, csrfCookie = svc.IssueCSRFCookie()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	clearCookie := findCookie(w.Result().Cookies(), "vocanova_session")
	require.NotNil(t, clearCookie)
	assert.True(t, clearCookie.Expires.Before(time.Now()) || clearCookie.MaxAge < 0)
}

func hasCookie(cookies []*http.Cookie, name string) bool {
	return findCookie(cookies, name) != nil
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func extractTokenFromURL(t *testing.T, body string) string {
	t.Helper()
	start := ""
	for _, prefix := range []string{"\n\n", "href=\"", ">"} {
		if i := strings.Index(body, prefix); i >= 0 {
			candidate := body[i+len(prefix):]
			if j := strings.Index(candidate, "\n"); j >= 0 {
				candidate = candidate[:j]
			}
			if j := strings.Index(candidate, "\""); j >= 0 {
				candidate = candidate[:j]
			}
			if strings.Contains(candidate, "token=") {
				start = candidate
				break
			}
		}
	}
	require.NotEmpty(t, start, "no token URL found in body: %s", body)
	u, err := url.Parse(start)
	require.NoError(t, err)
	return u.Query().Get("token")
}
