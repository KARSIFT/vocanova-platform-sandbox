package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLearningPreferencesHTTPContract(t *testing.T) {
	api, authSvc, repo := testOnboardingAPI(t)
	RegisterLearningPreferences(api, users.NewLearningPreferencesService(repo), authSvc)
	uid := uuid.New()
	require.NoError(t, repo.SetOnboardingStatus(t.Context(), uid, users.OnboardingStatusCompleted, time.Now()))
	request := func(method, body string, csrf bool, user uuid.UUID) *httptest.ResponseRecorder {
		req := onbRequesterRequest(t, method, "/api/v1/learning-preferences", body, user)
		if csrf {
			token, cookie := authSvc.IssueCSRFCookie()
			req.AddCookie(cookie)
			req.Header.Set("X-CSRF-Token", token)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	w := request(http.MethodGet, "", false, uid)
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `{"$schema":"https://example.com/schemas/LearningPreferencesDTO.json","learningGoal":null,"mainUseCase":null,"revision":0}`, w.Body.String())
	body := `{"learningGoal":"travel","mainUseCase":"travel","expectedRevision":0}`
	w = request(http.MethodPatch, body, false, uid)
	require.Equal(t, 403, w.Code)
	w = request(http.MethodPatch, body, true, uid)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.JSONEq(t, `{"$schema":"https://example.com/schemas/LearningPreferencesDTO.json","learningGoal":"travel","mainUseCase":"travel","revision":1}`, w.Body.String())
	w = request(http.MethodPatch, body, true, uid)
	require.Equal(t, 200, w.Code)
	w = request(http.MethodPatch, `{"learningGoal":"work","mainUseCase":"work","expectedRevision":0}`, true, uid)
	require.Equal(t, 409, w.Code)
	for _, invalid := range []string{`{"learningGoal":"work","mainUseCase":"work"}`, `{"learningGoal":"work","expectedRevision":1}`, `{"learningGoal":"other","mainUseCase":"work","expectedRevision":1}`, `{"learningGoal":null,"mainUseCase":"work","expectedRevision":1}`} {
		w = request(http.MethodPatch, invalid, true, uid)
		require.Equal(t, 422, w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "", false, uuid.New())
	require.Equal(t, 404, w.Code)
	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/learning-preferences", nil))
	require.Equal(t, 401, w.Code)
}

type preferencesReaderStub struct {
	value users.LearningPreferences
	err   error
}

func (r preferencesReaderStub) GetLearningPreferences(context.Context, uuid.UUID) (users.LearningPreferences, error) {
	return r.value, r.err
}

func TestDiscoverEffectiveLearningPreferenceAndFallback(t *testing.T) {
	r := users.NewMemoryRepository()
	uid := uuid.New()
	ctx := onbRequesterRequest(t, http.MethodGet, "/", "", uid).Context()
	_, _, err := r.CompleteOnboarding(ctx, uid, users.OnboardingAnswers{EnglishLevel: "a2", NativeLanguage: "fa", LearningGoal: "general", MainUseCase: "daily_life", DailyReviewTarget: 20}, time.Now())
	require.NoError(t, err)
	svc := users.NewService(r, r, r, nil)
	focus := "travel"
	require.Equal(t, "travel", requesterMainUseCase(ctx, svc, preferencesReaderStub{value: users.LearningPreferences{MainUseCase: &focus}}))
	require.Equal(t, "", requesterMainUseCase(ctx, svc, preferencesReaderStub{}), "authoritative unknown must not fabricate a preference")
	require.Equal(t, "daily_life", requesterMainUseCase(ctx, svc, preferencesReaderStub{err: errors.New("unavailable")}))
	require.Equal(t, "daily_life", requesterMainUseCase(ctx, svc))
}
