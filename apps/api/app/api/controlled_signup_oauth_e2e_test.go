package api

import (
	"database/sql"
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	controlledSignupAllowlistedEmail = "allowlisted-callback-e2e@synthetic.vocanova.invalid"
	controlledSignupUnlistedEmail    = "unlisted-callback-e2e@synthetic.vocanova.invalid"
	reservedSyntheticSmokeTestEmail  = "smoke-test-bot@synthetic.vocanova.invalid"

	controlledSignupE2EBaseURL       = "https://test.local"
	controlledSignupE2EOAuthRedirect = "https://test.local/api/v1/auth/oauth/google/callback"
	controlledSignupE2EAppReturnURL  = "https://test.local/onboarding"
	controlledSignupE2EOAuthClientID = "controlled-signup-e2e-client-id"
	controlledSignupE2EOAuthSecret   = "controlled-signup-e2e-client-secret"
	controlledSignupE2ESyntheticCode = "synthetic-authorization-code"
	controlledSignupE2EGoogleSubject = "google-sub-controlled-signup-e2e"
)

type controlledSignupFakeGoogle struct {
	tokenCalls    int
	userInfoCalls int
	invalidCalls  int
	identity      auth.OAuthIdentity
}

func (f *controlledSignupFakeGoogle) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			f.tokenCalls++
			if r.Method != http.MethodPost || r.ParseForm() != nil ||
				r.Form.Get("code") != controlledSignupE2ESyntheticCode ||
				r.Form.Get("client_id") != controlledSignupE2EOAuthClientID ||
				r.Form.Get("client_secret") != controlledSignupE2EOAuthSecret ||
				r.Form.Get("redirect_uri") != controlledSignupE2EOAuthRedirect ||
				r.Form.Get("grant_type") != "authorization_code" {
				f.invalidCalls++
				http.Error(w, "invalid synthetic token request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"synthetic-access-token","token_type":"Bearer","expires_in":3600}`))
		case strings.HasSuffix(r.URL.Path, "/userinfo"):
			f.userInfoCalls++
			if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic-access-token" {
				f.invalidCalls++
				http.Error(w, "invalid synthetic userinfo request", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			body, _ := json.Marshal(map[string]any{
				"sub":            f.identity.Subject,
				"email":          f.identity.Email,
				"email_verified": f.identity.EmailVerified,
				"name":           f.identity.DisplayName,
				"picture":        f.identity.AvatarURL,
			})
			_, _ = w.Write(body)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

type controlledSignupOAuthHarness struct {
	api        huma.API
	db         *sql.DB
	svc        *auth.Service
	fakeGoogle *controlledSignupFakeGoogle
}

func newControlledSignupOAuthHarness(t *testing.T, identity auth.OAuthIdentity, allowlist map[string]struct{}) *controlledSignupOAuthHarness {
	t.Helper()

	fake := &controlledSignupFakeGoogle{identity: identity}
	googleSrv := httptest.NewServer(fake.handler())
	t.Cleanup(googleSrv.Close)

	provider, err := auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     controlledSignupE2EOAuthClientID,
		ClientSecret: controlledSignupE2EOAuthSecret,
		RedirectURI:  controlledSignupE2EOAuthRedirect,
	})
	require.NoError(t, err)
	provider.TokenURL = googleSrv.URL + "/token"
	provider.UserInfoURL = googleSrv.URL + "/userinfo"
	provider.AuthEndpoint = googleSrv.URL + "/authorize"
	provider.Client = googleSrv.Client()

	db := newControlledSignupDisposablePostgres(t)

	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	repo := auth.NewPostgreSQLRepository(db)
	emailFake := &email.Fake{}
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	cfg := auth.Config{
		Environment:            "test",
		BaseURL:                controlledSignupE2EBaseURL,
		MagicLinkPath:          "/auth/magic",
		OAuthRedirectURI:       controlledSignupE2EOAuthRedirect,
		OAuthRedirectAllowlist: []string{controlledSignupE2EAppReturnURL},
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
			MagicRequestLimit:   100,
			MagicConsumeWindow:  time.Hour,
			MagicConsumeLimit:   100,
			OAuthStartWindow:    time.Hour,
			OAuthStartLimit:     100,
			OAuthCallbackWindow: time.Hour,
			OAuthCallbackLimit:  100,
			LogoutWindow:        time.Hour,
			LogoutLimit:         100,
		},
	}
	svc := auth.NewService(repo, emailFake, provider, c, limiter, cfg)
	svc.SetKillSwitches(&auth.KillSwitches{
		MagicLinkEnabled:       true,
		OAuthEnabled:           true,
		NewSignupsEnabled:      false,
		SignupAllowlist:        allowlist,
		ReservedSyntheticEmail: reservedSyntheticSmokeTestEmail,
	})

	humaConfig := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), humaConfig)
	api.UseMiddleware(withHumaContext)
	RegisterAuth(api, svc)

	return &controlledSignupOAuthHarness{
		api:        api,
		db:         db,
		svc:        svc,
		fakeGoogle: fake,
	}
}

type controlledSignupOAuthFlowResult struct {
	startStatus    int
	callbackStatus int
	callbackBody   string
	location       string
	sessionCookie  *http.Cookie
}

func (h *controlledSignupOAuthHarness) runOAuthFlow(t *testing.T) controlledSignupOAuthFlowResult {
	t.Helper()

	startRecorder := httptest.NewRecorder()
	startReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/oauth/google/start",
		strings.NewReader(`{"redirectUri":"`+controlledSignupE2EAppReturnURL+`"}`),
	)
	startReq.Header.Set("Content-Type", "application/json")
	h.api.Adapter().ServeHTTP(startRecorder, startReq)
	require.Equal(t, http.StatusOK, startRecorder.Code, "oauth start must succeed")

	var startBody struct {
		URL string `json:"url"`
	}
	require.NoError(t, json.Unmarshal(startRecorder.Body.Bytes(), &startBody))
	require.NotEmpty(t, startBody.URL)

	parsedAuthURL, err := url.Parse(startBody.URL)
	if err != nil {
		t.Fatal("oauth start returned an invalid authorization URL")
	}
	state := parsedAuthURL.Query().Get("state")
	require.NotEmpty(t, state)

	oauthStateCookie := findCookie(startRecorder.Result().Cookies(), "vocanova_oauth_state")
	require.NotNil(t, oauthStateCookie, "oauth state cookie must be issued")

	callbackRecorder := httptest.NewRecorder()
	callbackReq := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/oauth/google/callback?code="+url.QueryEscape(controlledSignupE2ESyntheticCode)+"&state="+url.QueryEscape(state),
		nil,
	)
	callbackReq.AddCookie(oauthStateCookie)
	h.api.Adapter().ServeHTTP(callbackRecorder, callbackReq)

	return controlledSignupOAuthFlowResult{
		startStatus:    startRecorder.Code,
		callbackStatus: callbackRecorder.Code,
		callbackBody:   callbackRecorder.Body.String(),
		location:       callbackRecorder.Header().Get("Location"),
		sessionCookie:  findCookie(callbackRecorder.Result().Cookies(), "vocanova_session"),
	}
}

func countUsersByEmail(t *testing.T, db *sql.DB, emailAddr string) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL`,
		emailAddr,
	).Scan(&count))
	return count
}

func countExternalIdentitiesByEmail(t *testing.T, db *sql.DB, emailAddr string) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM external_identities WHERE lower(provider_email) = lower($1)`,
		emailAddr,
	).Scan(&count))
	return count
}

func countSessionsForEmail(t *testing.T, db *sql.DB, emailAddr string) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE lower(u.email) = lower($1) AND s.revoked_at IS NULL`,
		emailAddr,
	).Scan(&count))
	return count
}

func userIDForEmail(t *testing.T, db *sql.DB, emailAddr string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, db.QueryRow(
		`SELECT id FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL`,
		emailAddr,
	).Scan(&id))
	return id
}

// TestControlledSignupOAuth_AllowlistedCallbackSucceeds exercises the full
// OAuth start/callback HTTP path with a real GoogleOAuthProvider against a
// local fake Google server and disposable Postgres. A never-before-seen
// allowlisted synthetic identity must reach the configured onboarding return
// URL with persisted user, external identity, and session rows.
func TestControlledSignupOAuth_AllowlistedCallbackSucceeds(t *testing.T) {
	identity := auth.OAuthIdentity{
		Subject:       controlledSignupE2EGoogleSubject + "-allowlisted",
		Email:         controlledSignupAllowlistedEmail,
		EmailVerified: true,
		DisplayName:   "Allowlisted Callback E2E",
		AvatarURL:     "https://test.local/avatar.png",
	}
	h := newControlledSignupOAuthHarness(t, identity, map[string]struct{}{
		controlledSignupAllowlistedEmail: {},
	})

	require.Equal(t, 0, countUsersByEmail(t, h.db, controlledSignupAllowlistedEmail),
		"precondition: allowlisted identity must not exist before callback")

	result := h.runOAuthFlow(t)

	assert.Equal(t, http.StatusFound, result.callbackStatus)
	assert.Equal(t, controlledSignupE2EAppReturnURL, result.location)
	require.NotNil(t, result.sessionCookie, "session cookie must be issued on success")
	assert.NotEmpty(t, result.sessionCookie.Value)

	assert.Equal(t, 1, countUsersByEmail(t, h.db, controlledSignupAllowlistedEmail))
	assert.Equal(t, 1, countExternalIdentitiesByEmail(t, h.db, controlledSignupAllowlistedEmail))
	assert.Equal(t, 1, countSessionsForEmail(t, h.db, controlledSignupAllowlistedEmail))

	userID := userIDForEmail(t, h.db, controlledSignupAllowlistedEmail)
	var providerSubject string
	require.NoError(t, h.db.QueryRow(
		`SELECT provider_subject FROM external_identities WHERE user_id = $1 AND provider = 'google'`,
		userID,
	).Scan(&providerSubject))
	assert.Equal(t, identity.Subject, providerSubject)

	assert.GreaterOrEqual(t, h.fakeGoogle.tokenCalls, 1, "GoogleOAuthProvider must call the fake token endpoint")
	assert.GreaterOrEqual(t, h.fakeGoogle.userInfoCalls, 1, "GoogleOAuthProvider must call the fake userinfo endpoint")
	assert.Zero(t, h.fakeGoogle.invalidCalls, "GoogleOAuthProvider must use the expected token and userinfo request shapes")

	t.Log("controlled-signup OAuth allowlisted callback succeeded with redirect to onboarding and persisted auth rows")
}

// TestControlledSignupOAuth_UnlistedCallbackDenied exercises the same HTTP
// stack with a never-before-seen unlisted synthetic identity. Global signup
// remains disabled and the callback must return HTTP 503 with the stable
// new-signups-disabled response without creating a user row.
func TestControlledSignupOAuth_UnlistedCallbackDenied(t *testing.T) {
	identity := auth.OAuthIdentity{
		Subject:       controlledSignupE2EGoogleSubject + "-unlisted",
		Email:         controlledSignupUnlistedEmail,
		EmailVerified: true,
		DisplayName:   "Unlisted Callback E2E",
	}
	h := newControlledSignupOAuthHarness(t, identity, map[string]struct{}{
		controlledSignupAllowlistedEmail: {},
	})

	require.Equal(t, 0, countUsersByEmail(t, h.db, controlledSignupUnlistedEmail),
		"precondition: unlisted identity must not exist before callback")

	result := h.runOAuthFlow(t)

	assert.Equal(t, http.StatusServiceUnavailable, result.callbackStatus)
	assert.Contains(t, strings.ToLower(result.callbackBody), "new sign-ups are disabled")
	assert.Empty(t, result.location)
	assert.Nil(t, result.sessionCookie, "no session cookie may be issued on signup denial")

	assert.Equal(t, 0, countUsersByEmail(t, h.db, controlledSignupUnlistedEmail))
	assert.Equal(t, 0, countExternalIdentitiesByEmail(t, h.db, controlledSignupUnlistedEmail))
	assert.Equal(t, 0, countSessionsForEmail(t, h.db, controlledSignupUnlistedEmail))

	assert.GreaterOrEqual(t, h.fakeGoogle.tokenCalls, 1)
	assert.GreaterOrEqual(t, h.fakeGoogle.userInfoCalls, 1)
	assert.Zero(t, h.fakeGoogle.invalidCalls, "GoogleOAuthProvider must use the expected token and userinfo request shapes")

	t.Log("controlled-signup OAuth unlisted callback denied with HTTP 503 and no persisted user")
}
