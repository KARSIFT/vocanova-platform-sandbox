package api

import (
	"context"
	"encoding/json"
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
	oauth := auth.NewFakeOAuthProvider(&auth.OAuthIdentity{Subject: "sub-123", Email: "user@example.com", EmailVerified: true, DisplayName: "User", AvatarURL: "https://example.com/avatar.png"})
	svc := auth.NewService(repo, fake, oauth, c, limiter, auth.Config{
		Environment:            "test",
		BaseURL:                "https://test.example.com",
		MagicLinkPath:          "/auth/magic",
		OAuthRedirectURI:       "https://test.example.com/auth/oauth/google/callback",
		OAuthRedirectAllowlist: []string{"https://test.example.com/app"},
		SessionLifetime:        30 * 24 * time.Hour,
		MagicLinkLifetime:      15 * time.Minute,
		OAuthStateLifetime:     10 * time.Minute,
		Cookie: auth.CookieConfig{
			Name:           "vocanova_session",
			CSRName:        "vocanova_csrf",
			OAuthStateName: "vocanova_oauth_state",
			Domain:         "",
			Secure:         false,
			SameSite:       http.SameSiteStrictMode,
		},
		RateLimit: auth.RateLimitConfig{
			MagicRequestWindow:  time.Hour,
			MagicRequestLimit:   10,
			MagicConsumeWindow:  time.Hour,
			MagicConsumeLimit:   10,
			OAuthStartWindow:    time.Hour,
			OAuthStartLimit:     10,
			OAuthCallbackWindow: time.Hour,
			OAuthCallbackLimit:  10,
			LogoutWindow:        time.Hour,
			LogoutLimit:         10,
		},
	})

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(svc))
	RegisterContract(api)
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

func TestRequestMagicLinkEndpointDisabledReturns503NotCrash(t *testing.T) {
	api, svc, _, _, _ := testAuthAPI(t)
	svc.SetKillSwitches(&auth.KillSwitches{MagicLinkEnabled: false, OAuthEnabled: true, NewSignupsEnabled: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestOAuthStartEndpointDisabledReturns503NotCrash(t *testing.T) {
	api, svc, _, _, _ := testAuthAPI(t)
	svc.SetKillSwitches(&auth.KillSwitches{MagicLinkEnabled: true, OAuthEnabled: false, NewSignupsEnabled: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"redirectUri":"https://test.example.com/app"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
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

	// Logout with CSRF clears the session and CSRF cookies.
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
	clearCSRFCookie := findCookie(w.Result().Cookies(), "vocanova_csrf")
	require.NotNil(t, clearCSRFCookie)
	assert.True(t, clearCSRFCookie.Expires.Before(time.Now()) || clearCSRFCookie.MaxAge < 0)
}

func TestOAuthStartEndpointReturnsURLAndSetsCookie(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"redirectUri":"https://test.example.com/app"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://fake-oauth.example.com/auth")
	setCookies := w.Result().Cookies()
	require.True(t, hasCookie(setCookies, "vocanova_oauth_state"), "oauth state cookie missing")
}

func TestOAuthStartEndpointRejectsUnknownRedirectURI(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"redirectUri":"https://evil.example.com/callback"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOAuthCallbackEndpointSetsSessionAndRedirects(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	// Start the flow.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"redirectUri":"https://test.example.com/app"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var startResp struct {
		URL string `json:"url"`
	}
	require.NoError(t, jsonDecode(w.Body.String(), &startResp))
	u, err := url.Parse(startResp.URL)
	require.NoError(t, err)
	state := u.Query().Get("state")
	require.NotEmpty(t, state)

	oauthStateCookie := findCookie(w.Result().Cookies(), "vocanova_oauth_state")
	require.NotNil(t, oauthStateCookie)

	// Callback with the returned state and cookie.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/callback?code=auth-code&state="+state, nil)
	req.AddCookie(oauthStateCookie)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://test.example.com/app", w.Header().Get("Location"))
	setCookies := w.Result().Cookies()
	require.True(t, hasCookie(setCookies, "vocanova_session"), "session cookie missing")

	// Oauth state cookie should be cleared.
	clearCookie := findCookie(setCookies, "vocanova_oauth_state")
	require.NotNil(t, clearCookie)
	assert.True(t, clearCookie.Expires.Before(time.Now()) || clearCookie.MaxAge < 0)

	// Replaying the callback fails.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/callback?code=auth-code&state="+state, nil)
	req.AddCookie(oauthStateCookie)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOAuthCallbackEndpointRejectsMissingStateCookie(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/google/start", strings.NewReader(`{"redirectUri":"https://test.example.com/app"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var startResp struct {
		URL string `json:"url"`
	}
	require.NoError(t, jsonDecode(w.Body.String(), &startResp))
	u, err := url.Parse(startResp.URL)
	require.NoError(t, err)
	state := u.Query().Get("state")

	// Callback without the state cookie should fail.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/google/callback?code=auth-code&state="+state, nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
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

func jsonDecode(body string, v any) error {
	return json.Unmarshal([]byte(body), v)
}

// consumeMagicLinkForEmail requests and consumes a magic link for the given
// email, returning the issued session cookie.
func consumeMagicLinkForEmail(t *testing.T, api huma.API, fake *email.Fake, emailAddr string) *http.Cookie {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links", strings.NewReader(`{"email":"`+emailAddr+`"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	msg, _ := fake.Last()
	rawToken := extractTokenFromURL(t, msg.BodyText)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-links/consume", strings.NewReader(`{"token":"`+rawToken+`","email":"`+emailAddr+`"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	return findCookie(w.Result().Cookies(), "vocanova_session")
}

func TestGetCurrentUserRequiresAuthentication(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authentication required")
}

func TestGetCurrentUserReturnsAuthenticatedRequester(t *testing.T) {
	api, _, _, fake, _ := testAuthAPI(t)

	sessionCookie := consumeMagicLinkForEmail(t, api, fake, "user@example.com")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body CurrentUser
	require.NoError(t, jsonDecode(w.Body.String(), &body))
	require.NotNil(t, body.Email)
	assert.Equal(t, "user@example.com", *body.Email)
}

func TestGetCurrentUserRejectsInvalidCookie(t *testing.T) {
	api, _, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "vocanova_session", Value: "not-a-valid-token"})
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUserRejectsExpiredSession(t *testing.T) {
	api, _, _, fake, c := testAuthAPI(t)

	sessionCookie := consumeMagicLinkForEmail(t, api, fake, "user@example.com")
	c.Advance(31 * 24 * time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUserRejectsRevokedSession(t *testing.T) {
	api, svc, _, fake, _ := testAuthAPI(t)

	sessionCookie := consumeMagicLinkForEmail(t, api, fake, "user@example.com")
	csrfToken, csrfCookie := svc.IssueCSRFCookie()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUserRejectsDisabledUser(t *testing.T) {
	api, _, repo, fake, _ := testAuthAPI(t)

	sessionCookie := consumeMagicLinkForEmail(t, api, fake, "user@example.com")
	u, err := repo.GetUserByEmail(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.NoError(t, repo.SetUserStatus(u.ID, "disabled"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUserCrossUserIsolation(t *testing.T) {
	api, _, _, fake, _ := testAuthAPI(t)

	sessionA := consumeMagicLinkForEmail(t, api, fake, "a@example.com")
	sessionB := consumeMagicLinkForEmail(t, api, fake, "b@example.com")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionA)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var bodyA CurrentUser
	require.NoError(t, jsonDecode(w.Body.String(), &bodyA))
	require.NotNil(t, bodyA.Email)
	assert.Equal(t, "a@example.com", *bodyA.Email)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(sessionB)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var bodyB CurrentUser
	require.NoError(t, jsonDecode(w.Body.String(), &bodyB))
	require.NotNil(t, bodyB.Email)
	assert.Equal(t, "b@example.com", *bodyB.Email)
}

func TestLogoutRequiresAuthentication(t *testing.T) {
	api, svc, _, _, _ := testAuthAPI(t)

	w := httptest.NewRecorder()
	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogoutRequiresCSRF(t *testing.T) {
	api, _, _, fake, _ := testAuthAPI(t)

	sessionCookie := consumeMagicLinkForEmail(t, api, fake, "user@example.com")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
