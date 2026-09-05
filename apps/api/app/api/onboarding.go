package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/danielgtaylor/huma/v2"
)

// OnboardingProfileDTO is the public projection of a
// user_onboarding_profiles row. The five onboarding answers are
// always returned when the profile exists; the Status field is
// always returned.
type OnboardingProfileDTO struct {
	Status            string  `json:"status" doc:"Current onboarding status (not_started|in_progress|completed)"`
	EnglishLevel      *string `json:"englishLevel,omitempty" doc:"Self-reported English level (a1|a2|b1|b2|unknown)"`
	NativeLanguage    *string `json:"nativeLanguage,omitempty" doc:"Self-reported native language"`
	LearningGoal      *string `json:"learningGoal,omitempty" doc:"Self-reported learning goal"`
	MainUseCase       *string `json:"mainUseCase,omitempty" doc:"Self-reported main use case"`
	DailyReviewTarget *int    `json:"dailyReviewTarget,omitempty" doc:"Self-reported daily review target (5-100)"`
	CompletedAt       *string `json:"completedAt,omitempty" format:"date-time" doc:"Onboarding completion timestamp"`
}

// GetOnboardingOutput is the response body for
// GET /api/v1/onboarding. The 200 response always carries a body —
// a learner who has not yet submitted onboarding receives
// status='not_started' with no answers and no completedAt.
type GetOnboardingOutput struct {
	Body OnboardingProfileDTO
}

// CompleteOnboardingInput is the request body for
// POST /api/v1/onboarding. The five answers are required; the
// requester-scoped user identity is taken from the session.
type CompleteOnboardingInput struct {
	Body struct {
		EnglishLevel      string `json:"englishLevel" enum:"a1,a2,b1,b2,unknown" doc:"Self-reported English level"`
		NativeLanguage    string `json:"nativeLanguage" minLength:"1" doc:"Self-reported native language"`
		LearningGoal      string `json:"learningGoal" enum:"general,work,travel,study,conversation,exam" doc:"Self-reported learning goal"`
		MainUseCase       string `json:"mainUseCase" enum:"daily_life,work,travel,study,social" doc:"Self-reported main use case"`
		DailyReviewTarget int    `json:"dailyReviewTarget" minimum:"5" maximum:"100" doc:"Self-reported daily review target"`
	}
}

// CompleteOnboardingOutput echoes the persisted profile so the
// frontend can immediately route the learner away from /onboarding.
type CompleteOnboardingOutput struct {
	Body OnboardingProfileDTO
}

// RegisterOnboarding wires the GET/POST /api/v1/onboarding routes
// on api, requiring an authenticated requester for both. authSvc is
// the auth service used to enforce double-submit CSRF on the
// state-changing POST. The route is requester-scoped; the userID
// is always taken from the session.
func RegisterOnboarding(api huma.API, svc *users.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "GetOnboarding",
		Method:      http.MethodGet,
		Path:        "/api/v1/onboarding",
		Summary:     "Get the requester's onboarding profile",
		Tags:        []string{"Onboarding"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"500": {Description: "Internal server error"},
		},
	}, func(ctx context.Context, input *struct{}) (*GetOnboardingOutput, error) {
		uid := RequesterUserID(ctx)
		profile, err := svc.GetOnboarding(ctx, uid)
		if err != nil && !errors.Is(err, users.ErrOnboardingNotFound) {
			return nil, mapOnboardingError(err)
		}
		return &GetOnboardingOutput{Body: onboardingProfileToDTO(profile)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "CompleteOnboarding",
		Method:      http.MethodPost,
		Path:        "/api/v1/onboarding",
		Summary:     "Complete onboarding for the requester",
		Tags:        []string{"Onboarding"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid onboarding submission"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"409": {Description: "Onboarding profile already exists with different answers"},
			"500": {Description: "Internal server error"},
		},
	}, func(ctx context.Context, input *CompleteOnboardingInput) (*CompleteOnboardingOutput, error) {
		uid := RequesterUserID(ctx)
		answers := users.OnboardingAnswers{
			EnglishLevel:      input.Body.EnglishLevel,
			NativeLanguage:    input.Body.NativeLanguage,
			LearningGoal:      input.Body.LearningGoal,
			MainUseCase:       input.Body.MainUseCase,
			DailyReviewTarget: input.Body.DailyReviewTarget,
		}
		profile, _, err := svc.CompleteOnboarding(ctx, uid, answers)
		if err != nil {
			return nil, mapOnboardingError(err)
		}
		return &CompleteOnboardingOutput{Body: onboardingProfileToDTO(profile)}, nil
	})
}

func onboardingProfileToDTO(p *users.OnboardingProfile) OnboardingProfileDTO {
	if p == nil {
		return OnboardingProfileDTO{Status: users.OnboardingStatusNotStarted}
	}
	dto := OnboardingProfileDTO{Status: p.Status}
	if p.Status == users.OnboardingStatusCompleted {
		level := p.EnglishLevel
		native := p.NativeLanguage
		goal := p.LearningGoal
		use := p.MainUseCase
		target := p.DailyReviewTarget
		dto.EnglishLevel = &level
		dto.NativeLanguage = &native
		dto.LearningGoal = &goal
		dto.MainUseCase = &use
		dto.DailyReviewTarget = &target
		if p.CompletedAt != nil {
			s := p.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			dto.CompletedAt = &s
		}
	}
	return dto
}

func mapOnboardingError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, users.ErrOnboardingNotFound):
		return huma.Error404NotFound("onboarding profile not found")
	case errors.Is(err, users.ErrOnboardingConflict):
		return huma.Error409Conflict("onboarding profile already exists with different answers")
	case errors.Is(err, users.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	case errors.Is(err, users.ErrInvalidOnboarding):
		return huma.Error400BadRequest("invalid onboarding answers")
	}
	// Unexpected repository errors may include private infrastructure details.
	return huma.Error500InternalServerError("internal error")
}
