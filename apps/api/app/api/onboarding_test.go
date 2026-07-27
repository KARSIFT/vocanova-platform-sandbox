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

func testNow() time.Time {
	return time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
}

// testOnboardingAPI wires a Huma API with the contract, auth, and
// onboarding routes registered against the same in-memory auth
// repository and a fresh in-memory users repository. The
// OnboardingStatusLookup the contract handler uses is installed
// against the same in-memory users service, so GET /api/v1/me sees
// the additive onboardingStatus field.
func testOnboardingAPI(t *testing.T) (huma.API, *auth.Service, *users.MemoryRepository) {
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
	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		profile, err := usersSvc.GetOnboarding(ctx, userID)
		if err != nil {
			return users.OnboardingStatusNotStarted, nil
		}
		return profile.Status, nil
	})
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })

	return api, svc, usersRepo
}

// onbRequesterRequest issues a request with the requester injected
// directly, matching the pattern in content_test.go. It is the
// primary helper for the onboarding tests; CSRF tests use the
// same helper plus an extra CSRF cookie/header pair.
func onbRequesterRequest(t *testing.T, method, path, body string, userID uuid.UUID) *http.Request {
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

func TestGetOnboardingReturnsNotStartedForUnseenUser(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/onboarding", "", uid))

	require.Equal(t, http.StatusOK, w.Code)
	var body OnboardingProfileDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, users.OnboardingStatusNotStarted, body.Status)
	assert.Nil(t, body.EnglishLevel, "no profile: no answers")
}

func TestGetOnboardingRequiresAuthentication(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding", nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompleteOnboardingHappyPath(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "complete-onboarding must succeed for a valid submission: %s", w.Body.String())
	var body OnboardingProfileDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, users.OnboardingStatusCompleted, body.Status)
	require.NotNil(t, body.EnglishLevel)
	assert.Equal(t, "b1", *body.EnglishLevel)
	require.NotNil(t, body.DailyReviewTarget)
	assert.Equal(t, 25, *body.DailyReviewTarget)
}

func TestGetOnboardingReturnsCompletedAfterSubmit(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	post := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	post.AddCookie(csrfCookie)
	post.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, post)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/onboarding", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body OnboardingProfileDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, users.OnboardingStatusCompleted, body.Status)
	require.NotNil(t, body.EnglishLevel)
	assert.Equal(t, "b1", *body.EnglishLevel)
	require.NotNil(t, body.NativeLanguage)
	assert.Equal(t, "es", *body.NativeLanguage)
}

func TestCompleteOnboardingSeedsUserSettingsWhenNoRow(t *testing.T) {
	api, svc, usersRepo := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"a2","nativeLanguage":"es","learningGoal":"work","mainUseCase":"work","dailyReviewTarget":30}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	stored, err := usersRepo.GetStoredUserSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.True(t, stored.Stored, "D04: no existing user_settings row → row is created")
	assert.Equal(t, 30, stored.DailyReviewTarget, "D04: no existing row → seed with onboarding answer")
}

func TestCompleteOnboardingPreservesCustomizedUserSettings(t *testing.T) {
	api, svc, usersRepo := testOnboardingAPI(t)
	uid := uuid.New()
	// Pre-seed a customized user_settings row.
	usersRepo.UpsertStoredUserSettings(uid, users.MemoryUserSettings{
		Timezone:          "Europe/Madrid",
		DailyReviewTarget: 50,
	})

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":10}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	stored, err := usersRepo.GetStoredUserSettings(context.Background(), uid)
	require.NoError(t, err)
	assert.True(t, stored.Stored)
	assert.Equal(t, 50, stored.DailyReviewTarget, "D04: customized existing → never overwrite")
}

func TestCompleteOnboardingRejectsInvalidEnglishLevel(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"c1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.Less(t, w.Code, 500, "expected client error, got %d", w.Code)
}

func TestCompleteOnboardingRejectsDailyReviewOutOfRange(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":200}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.Less(t, w.Code, 500)
}

func TestCompleteOnboardingRejectsEmptyNativeLanguage(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.Less(t, w.Code, 500)
}

func TestCompleteOnboardingRequiresAuthentication(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding", nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompleteOnboardingRequiresCSRF(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()

	// Submit without a CSRF cookie/header. RequireAuth has
	// already authenticated the requester (we inject one
	// directly); the CSRF middleware is the next gate.
	req := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "missing CSRF must reject")
}

func TestGetCurrentUserIncludesAdditiveOnboardingStatus(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body CurrentUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	// Additive contract: onboardingStatus is always present; the
	// other A1 fields are unchanged.
	assert.Equal(t, "not_started", body.OnboardingStatus)
}

func TestGetCurrentUserOnboardingStatusFlipsAfterComplete(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	post := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	post.AddCookie(csrfCookie)
	post.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, post)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body CurrentUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "completed", body.OnboardingStatus)
}

func TestGetCurrentUserFallbackWhenLookupErrors(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()
	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		return "", lookupErrSentinel{}
	})
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", uid))
	require.Equal(t, http.StatusOK, w.Code, "contract must remain available even if the lookup errors")
	var body CurrentUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "not_started", body.OnboardingStatus, "errors fall back to not_started")
}

type lookupErrSentinel struct{}

func (lookupErrSentinel) Error() string { return "intentional lookup failure" }

func TestGetCurrentUserFallbackWhenLookupReturnsEmpty(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()
	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		return "", nil
	})
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body CurrentUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "not_started", body.OnboardingStatus, "empty result falls back to not_started")
}

func TestSetOnboardingStatusLookupNilRestoresDefault(t *testing.T) {
	api, _, _ := testOnboardingAPI(t)
	uid := uuid.New()
	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		return "in_progress", nil
	})
	SetOnboardingStatusLookup(nil)

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", uid))
	require.Equal(t, http.StatusOK, w.Code)
	var body CurrentUser
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "not_started", body.OnboardingStatus, "restoring nil must reset the default")
}

func TestOnboardingProfileDTOCompletionHasCompletedAt(t *testing.T) {
	api, svc, _ := testOnboardingAPI(t)
	uid := uuid.New()

	csrfToken, csrfCookie := svc.IssueCSRFCookie()
	post := onbRequesterRequest(t, http.MethodPost, "/api/v1/onboarding",
		`{"englishLevel":"b1","nativeLanguage":"es","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":25}`,
		uid)
	post.AddCookie(csrfCookie)
	post.Header.Set("X-CSRF-Token", csrfToken)
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, post)
	require.Equal(t, http.StatusOK, w.Code)

	var body OnboardingProfileDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotNil(t, body.CompletedAt, "completed response must include completedAt")
	assert.NotEmpty(t, *body.CompletedAt)
}

func TestOnboardingProfileDTONotStartedHidesAnswers(t *testing.T) {
	dto := onboardingProfileToDTO(&users.OnboardingProfile{Status: users.OnboardingStatusNotStarted})
	assert.Equal(t, users.OnboardingStatusNotStarted, dto.Status)
	assert.Nil(t, dto.EnglishLevel)
	assert.Nil(t, dto.NativeLanguage)
	assert.Nil(t, dto.LearningGoal)
	assert.Nil(t, dto.MainUseCase)
	assert.Nil(t, dto.DailyReviewTarget)
	assert.Nil(t, dto.CompletedAt)
}

func TestOnboardingProfileDTONilProfileSynthesizesNotStarted(t *testing.T) {
	dto := onboardingProfileToDTO(nil)
	assert.Equal(t, users.OnboardingStatusNotStarted, dto.Status)
}

func TestOnboardingProfileDTOCompletedPopulatesAllFields(t *testing.T) {
	completed := testNow()
	dto := onboardingProfileToDTO(&users.OnboardingProfile{
		Status:            users.OnboardingStatusCompleted,
		EnglishLevel:      "b2",
		NativeLanguage:    "fr",
		LearningGoal:      "exam",
		MainUseCase:       "study",
		DailyReviewTarget: 35,
		CompletedAt:       &completed,
	})
	assert.Equal(t, users.OnboardingStatusCompleted, dto.Status)
	require.NotNil(t, dto.EnglishLevel)
	assert.Equal(t, "b2", *dto.EnglishLevel)
	require.NotNil(t, dto.NativeLanguage)
	assert.Equal(t, "fr", *dto.NativeLanguage)
	require.NotNil(t, dto.LearningGoal)
	assert.Equal(t, "exam", *dto.LearningGoal)
	require.NotNil(t, dto.MainUseCase)
	assert.Equal(t, "study", *dto.MainUseCase)
	require.NotNil(t, dto.DailyReviewTarget)
	assert.Equal(t, 35, *dto.DailyReviewTarget)
	require.NotNil(t, dto.CompletedAt)
	assert.NotEmpty(t, *dto.CompletedAt)
}
