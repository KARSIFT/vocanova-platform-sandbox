package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type failingOnboardingRepository struct{ *users.MemoryRepository }

func TestMapOnboardingErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"validation", users.ErrInvalidOnboarding, 400},
		{"missing profile", users.ErrOnboardingNotFound, 404},
		{"missing user", users.ErrUserNotFound, 404},
		{"conflict", users.ErrOnboardingConflict, 409},
		{"internal", errors.New("private database credential"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, err := range []error{tc.err, fmt.Errorf("private context: %w", tc.err)} {
				mapped := mapOnboardingError(err)
				assert.Equal(t, tc.status, mapped.GetStatus())
				assert.NotContains(t, mapped.Error(), "private")
			}
		})
	}
	assert.Nil(t, mapOnboardingError(nil))
}

func (r failingOnboardingRepository) GetOnboarding(context.Context, uuid.UUID) (*users.OnboardingProfile, error) {
	return nil, errors.New("private database address and credential")
}

func (r failingOnboardingRepository) CompleteOnboarding(context.Context, uuid.UUID, users.OnboardingAnswers, time.Time) (*users.OnboardingProfile, users.StoredUserSettings, error) {
	return nil, users.StoredUserSettings{}, errors.New("private database address and credential")
}

func TestOnboardingRepositoryErrorsArePrivate(t *testing.T) {
	_, authSvc, _ := testOnboardingAPI(t)
	repo := failingOnboardingRepository{users.NewMemoryRepository()}
	svc := users.NewService(repo, nil, nil, nil)
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Onboarding failures", "1"))
	api.UseMiddleware(withHumaContext)
	RegisterOnboarding(api, svc, authSvc)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			req := onbRequesterRequest(t, method, "/api/v1/onboarding", `{"englishLevel":"a2","nativeLanguage":"en","learningGoal":"general","mainUseCase":"daily_life","dailyReviewTarget":20}`, uuid.New())
			token, cookie := authSvc.IssueCSRFCookie()
			req.AddCookie(cookie)
			req.Header.Set("X-CSRF-Token", token)
			rec := httptest.NewRecorder()
			api.Adapter().ServeHTTP(rec, req)
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotContains(t, rec.Body.String(), "private database")
			assert.NotContains(t, rec.Body.String(), "credential")
		})
	}
}
