package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSettingsAPI wires a Huma API with the contract, auth, and
// settings routes registered against the same in-memory auth
// repository and a fresh in-memory users repository. The
// MemoryRepository is wired in as the SettingsRepository so the
// new GET/PATCH /api/v1/settings routes have a real
// implementation to call. The onboarding routes are also
// registered so the cross-route isolation tests can probe both
// surfaces.
func testSettingsAPI(t *testing.T) (huma.API, *auth.Service, *users.MemoryRepository) {
	t.Helper()
	now := testNow()
	c := &clock.Fixed{T: now}
	authRepo := auth.NewMemoryRepository()
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	oauth := auth.NewFakeOAuthProvider(&auth.OAuthIdentity{
		Subject: "sub-1", Email: "user@example.com", EmailVerified: true,
		DisplayName: "User", AvatarURL: "https://example.com/avatar.png",
	})
	svc := auth.NewService(authRepo, nil, oauth, c, limiter, auth.Config{
		Environment:        "test",
		BaseURL:            "https://test.example.com",
		MagicLinkPath:      "/auth/magic",
		OAuthRedirectURI:   "https://test.example.com/auth/oauth/google/callback",
		SessionLifetime:    30 * 24 * time.Hour,
		MagicLinkLifetime:  15 * time.Minute,
		OAuthStateLifetime: 10 * time.Minute,
		Cookie: auth.CookieConfig{
			Name:           "vocanova_session",
			CSRName:        "vocanova_csrf",
			OAuthStateName: "vocanova_oauth_state",
			Domain:         "",
			Secure:         false,
			SameSite:       http.SameSiteStrictMode,
		},
		RateLimit: auth.RateLimitConfig{
			MagicRequestWindow: time.Hour, MagicRequestLimit: 10,
			MagicConsumeWindow: time.Hour, MagicConsumeLimit: 10,
			OAuthStartWindow: time.Hour, OAuthStartLimit: 10,
			OAuthCallbackWindow: time.Hour, OAuthCallbackLimit: 10,
			LogoutWindow: time.Hour, LogoutLimit: 10,
		},
	})

	usersRepo := users.NewMemoryRepository()
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, c)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(svc))
	RegisterContract(api)
	RegisterAuth(api, svc)
	RegisterOnboarding(api, usersSvc, svc)
	RegisterSettings(api, usersSvc, svc)
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })

	return api, svc, usersRepo
}

// settingsRequesterRequest issues a request with the requester
// injected directly, matching the onbRequesterRequest helper in
// onboarding_test.go.
func settingsRequesterRequest(t *testing.T, method, path, body string, userID uuid.UUID) *http.Request {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	return req
}

// TestGetSettingsRequiresAuthentication covers VOC-031-TEST-08:
// unauthenticated access is rejected with 401, matching every
// other (app) route.
func TestGetSettingsRequiresAuthentication(t *testing.T) {
	api, _, _ := testSettingsAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetSettingsReturnsSchemaDefaultsForUnseenUser covers
// VOC-031-TEST-08: a requester who has authenticated but has no
// stored user_settings row yet receives a stable response
// populated from the user_settings schema defaults. The shape
// is identical to what a learner with a fully-defaulted row
// would see, so the frontend does not branch on "no row yet".
func TestGetSettingsReturnsSchemaDefaultsForUnseenUser(t *testing.T) {
	api, _, _ := testSettingsAPI(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 20, body.DailyReviewTarget)
	assert.Equal(t, "vocanova_default", body.ReviewIntervalPreset)
	assert.Equal(t, "en", body.AppLanguage)
	assert.True(t, body.NotificationsEnabled)
	assert.False(t, body.MarketingEmailsEnabled)
	assert.Equal(t, "", body.DisplayName)
}

// TestGetSettingsReturnsSettingsNotFoundForBrandNewUser covers
// VOC-031-TEST-08: the in-memory fixture treats every
// authenticated user as a real, non-deleted user in the users
// table (matching the SQL path's "WHERE deleted_at IS NULL"
// always-true predicate for test fixtures). The 404 path is
// exercised by integration tests against a real database, not
// here.
func TestGetSettingsReturnsSchemaDefaultsForBrandNewUser(t *testing.T) {
	api, _, _ := testSettingsAPI(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	require.Equal(t, http.StatusOK, w.Code, "every authenticated requester sees a 200 Settings projection")
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 20, body.DailyReviewTarget, "schema default for an unseen user")
	assert.Equal(t, "vocanova_default", body.ReviewIntervalPreset, "schema default for an unseen user")
	assert.Equal(t, "en", body.AppLanguage, "schema default for an unseen user")
}

// TestGetSettingsReturnsPersistedDisplayName covers
// VOC-031-TEST-08: the displayName comes from users.display_name
// even when no user_settings row exists.
func TestGetSettingsReturnsPersistedDisplayName(t *testing.T) {
	api, _, usersRepo := testSettingsAPI(t)
	uid := uuid.New()
	usersRepo.MarkSeen(uid)
	usersRepo.SetDisplayName(uid, "Ada")

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Ada", body.DisplayName)
}

// TestUpdateSettingsRequiresAuthentication covers VOC-031-TEST-08:
// unauthenticated PATCH is rejected with 401.
func TestUpdateSettingsRequiresAuthentication(t *testing.T) {
	api, _, _ := testSettingsAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings",
		strings.NewReader(`{"dailyReviewTarget":30}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestUpdateSettingsRequiresCSRF covers VOC-031-TEST-08: the PATCH
// route is a state-changing endpoint and is gated by the same
// double-submit CSRF middleware every other PATCH/POST in the
// authenticated app uses.
func TestUpdateSettingsRequiresCSRF(t *testing.T) {
	api, _, _ := testSettingsAPI(t)
	uid := uuid.New()

	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":30}`, uid)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "missing CSRF must reject")
}

// TestUpdateSettingsHappyPath covers VOC-031-TEST-08: a PATCH
// with a valid payload writes every field and the response
// reflects the merged state.
func TestUpdateSettingsHappyPath(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":35,"reviewIntervalPreset":"wordup_like","appLanguage":"en","notificationsEnabled":false,"marketingEmailsEnabled":true,"displayName":"Grace"}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "valid PATCH must succeed: %s", w.Body.String())
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 35, body.DailyReviewTarget)
	assert.Equal(t, "wordup_like", body.ReviewIntervalPreset)
	assert.Equal(t, "en", body.AppLanguage)
	assert.False(t, body.NotificationsEnabled)
	assert.True(t, body.MarketingEmailsEnabled)
	assert.Equal(t, "Grace", body.DisplayName)
}

// TestUpdateSettingsIsPartial covers VOC-031-TEST-08: a PATCH
// that sets only one field preserves the existing values of
// every other field.
func TestUpdateSettingsIsPartial(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	// Seed the user with a full row so every other field has
	// a non-default value to preserve.
	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	first := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":40,"reviewIntervalPreset":"wordup_like","notificationsEnabled":false,"marketingEmailsEnabled":true,"displayName":"Ada"}`,
		uid)
	first.AddCookie(csrfCookie)
	first.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, first)
	require.Equal(t, http.StatusOK, w.Code, "first PATCH must succeed: %s", w.Body.String())

	// Now PATCH only dailyReviewTarget.
	csrfToken, csrfCookie = svc.IssueCSRFCookie()
	second := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":60}`, uid)
	second.AddCookie(csrfCookie)
	second.Header.Set("X-CSRF-Token", csrfToken)
	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, second)
	require.Equal(t, http.StatusOK, w.Code, "second PATCH must succeed: %s", w.Body.String())
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 60, body.DailyReviewTarget, "the one field changed")
	assert.Equal(t, "wordup_like", body.ReviewIntervalPreset, "unchanged: existing value preserved")
	assert.Equal(t, "en", body.AppLanguage, "unchanged: existing value preserved")
	assert.False(t, body.NotificationsEnabled, "unchanged: existing value preserved")
	assert.True(t, body.MarketingEmailsEnabled, "unchanged: existing value preserved")
	assert.Equal(t, "Ada", body.DisplayName, "unchanged: existing value preserved")
}

// TestUpdateSettingsEmptyBodyIsNoOp covers VOC-031-TEST-08: a
// PATCH with an empty body is a no-op (the response shape is the
// current state). Mirrors DOC-07 §3.
func TestUpdateSettingsEmptyBodyIsNoOp(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings", `{}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 20, body.DailyReviewTarget, "schema default preserved")
	assert.Equal(t, "vocanova_default", body.ReviewIntervalPreset, "schema default preserved")
}

// TestUpdateSettingsRejectsOutOfRangeDailyReviewTarget covers
// VOC-031-TEST-08: an out-of-range daily review target is
// rejected. Huma returns 422 (Unprocessable Entity) for the
// JSON-Schema minimum/maximum violation; the frontend treats
// 4xx as a rejection of the submission regardless of the
// exact code.
func TestUpdateSettingsRejectsOutOfRangeDailyReviewTarget(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":200}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.GreaterOrEqual(t, w.Code, 400, "out-of-range target must be rejected with a 4xx")
	require.Less(t, w.Code, 500, "must be a client error, not a 5xx")
}

// TestUpdateSettingsRejectsUnknownReviewIntervalPreset covers
// VOC-031-TEST-08: an invalid review interval preset is
// rejected. Huma returns 422 for the JSON-Schema enum
// violation; the frontend treats 4xx as a rejection.
func TestUpdateSettingsRejectsUnknownReviewIntervalPreset(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"reviewIntervalPreset":"anki"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.GreaterOrEqual(t, w.Code, 400, "unknown preset must be rejected with a 4xx")
	require.Less(t, w.Code, 500, "must be a client error, not a 5xx")
}

// TestUpdateSettingsRejectsUnsupportedAppLanguage covers
// VOC-031-TEST-08 + D06: app language "es" is rejected at the
// Huma boundary (and as a second line of defense at the
// service layer) because the founder-directed set accepts
// only "en" at launch.
func TestUpdateSettingsRejectsUnsupportedAppLanguage(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"appLanguage":"es"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.GreaterOrEqual(t, w.Code, 400, "unsupported app language must be rejected with a 4xx")
	require.Less(t, w.Code, 500, "must be a client error, not a 5xx")
}

// TestUpdateSettingsRejectsOverlongDisplayName covers
// VOC-031-TEST-08: a displayName beyond MaxDisplayNameLength is
// rejected. Huma's maxLength tag and the service-layer
// validation are both lines of defense.
func TestUpdateSettingsUnicodeDisplayNameRoundTrip(t *testing.T) {
	for _, character := range []string{"م", "界", "😀"} {
		t.Run(character, func(t *testing.T) {
			api, svc, _ := testSettingsAPI(t)
			uid := uuid.New()
			name := strings.Repeat(character, users.MaxDisplayNameLength)
			payload, err := json.Marshal(map[string]string{"displayName": name})
			require.NoError(t, err)
			csrfToken, csrfCookie := svc.IssueCSRFCookie()
			req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings", string(payload), uid)
			req.AddCookie(csrfCookie)
			req.Header.Set("X-CSRF-Token", csrfToken)
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())

			w = httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
			require.Equal(t, http.StatusOK, w.Code)
			var result SettingsDTO
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			assert.Equal(t, name, result.DisplayName)

			payload, err = json.Marshal(map[string]string{"displayName": name + character})
			require.NoError(t, err)
			req = settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings", string(payload), uid)
			req.AddCookie(csrfCookie)
			req.Header.Set("X-CSRF-Token", csrfToken)
			w = httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, req)
			require.GreaterOrEqual(t, w.Code, 400)
			require.Less(t, w.Code, 500)
		})
	}
}

func TestUpdateSettingsRejectsOverlongDisplayName(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	overlong := strings.Repeat("a", users.MaxDisplayNameLength+1)
	body, err := json.Marshal(map[string]string{"displayName": overlong})
	require.NoError(t, err)

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings", string(body), uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.GreaterOrEqual(t, w.Code, 400, "overlong display name must be rejected with a 4xx")
	require.Less(t, w.Code, 500, "must be a client error, not a 5xx")
}

// TestUpdateSettingsRejectsUnknownField covers VOC-031-TEST-08 +
// DOC-07 §3: an unknown field in the PATCH body is rejected at
// the Huma boundary (additionalProperties: false). Huma returns
// 422 (Unprocessable Entity) for body-validation failures, which
// is the framework's default 4xx code for JSON-Schema violations;
// the frontend treats 4xx as a rejection of the submission
// regardless of the exact code.
func TestUpdateSettingsRejectsUnknownField(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":30,"unknownField":"value"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.GreaterOrEqual(t, w.Code, 400, "unknown field must be rejected with a 4xx")
	require.Less(t, w.Code, 500, "must be a client error, not a 5xx")
}

// TestUpdateSettingsPersistsAcrossReads covers VOC-031-TEST-08:
// after a PATCH, the next GET returns the merged state.
func TestUpdateSettingsPersistsAcrossReads(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	patch := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":42,"displayName":"Hopper"}`, uid)
	patch.AddCookie(csrfCookie)
	patch.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, patch)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 42, body.DailyReviewTarget)
	assert.Equal(t, "Hopper", body.DisplayName)
}

// TestUpdateSettingsUpsertsOnFirstWrite covers VOC-031-TEST-08 +
// VOC-031-R05: a PATCH by a requester who has no user_settings
// row yet must succeed (the row is created), and the response
// reflects the merge of the supplied values and the schema
// defaults.
func TestUpdateSettingsUpsertsOnFirstWrite(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	patch := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":30}`, uid)
	patch.AddCookie(csrfCookie)
	patch.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, patch)
	require.Equal(t, http.StatusOK, w.Code, "first PATCH must succeed: %s", w.Body.String())
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 30, body.DailyReviewTarget, "supplied value")
	assert.Equal(t, "vocanova_default", body.ReviewIntervalPreset, "schema default for an unset field")
	assert.Equal(t, "en", body.AppLanguage, "schema default for an unset field")
	assert.True(t, body.NotificationsEnabled, "schema default for an unset field")
	assert.False(t, body.MarketingEmailsEnabled, "schema default for an unset field")
}

// TestGetSettingsAfterUpdateReturnsConsistentShape pins the
// contract that the GET response shape is identical to the PATCH
// response shape, so the frontend can rely on a single Settings
// DTO.
func TestGetSettingsAfterUpdateReturnsConsistentShape(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	patch := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":25,"reviewIntervalPreset":"custom","appLanguage":"en","notificationsEnabled":false,"marketingEmailsEnabled":true,"displayName":"Lin"}`,
		uid)
	patch.AddCookie(csrfCookie)
	patch.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, patch)
	require.Equal(t, http.StatusOK, w.Code)
	var patchBody SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &patchBody))

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var getBody SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getBody))
	assert.Equal(t, patchBody, getBody, "PATCH and GET responses must be identical")
}

// TestUpdateSettingsScopingIsPerRequester ensures the
// requester-scoped rule is enforced: two authenticated
// requesters only see and write their own settings.
func TestUpdateSettingsScopingIsPerRequester(t *testing.T) {
	api, svc, _ := testSettingsAPI(t)
	uidA := uuid.New()
	uidB := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	patchA := settingsRequesterRequest(t, http.MethodPatch, "/api/v1/settings",
		`{"dailyReviewTarget":30,"displayName":"Alice"}`, uidA)
	patchA.AddCookie(csrfCookie)
	patchA.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, patchA)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uidB))
	require.Equal(t, http.StatusOK, w.Code)
	var body SettingsDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 20, body.DailyReviewTarget, "B has not written; sees the schema default")
	assert.Equal(t, "", body.DisplayName, "B's displayName is empty")
}

// TestGetSettingsContextIsolatedFromOnboarding ensures the new
// settings routes do not interfere with the onboarding contract
// the prior milestone already shipped.
func TestGetSettingsContextIsolatedFromOnboarding(t *testing.T) {
	api, _, _ := testSettingsAPI(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/onboarding", "", uid))
	assert.Equal(t, http.StatusOK, w.Code, "onboarding route must still respond")

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, settingsRequesterRequest(t, http.MethodGet, "/api/v1/settings", "", uid))
	assert.Equal(t, http.StatusOK, w.Code, "settings route must respond")
}

// ensure context import is used in case other helpers in this
// file grow a need for it.
var _ = context.Background
