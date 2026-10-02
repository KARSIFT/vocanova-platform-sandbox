package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
)

type LearningPreferencesReader interface {
	GetLearningPreferences(context.Context, uuid.UUID) (users.LearningPreferences, error)
}

type LearningPreferencesDTO struct {
	LearningGoal *string `json:"learningGoal" nullable:"true" enum:"general,work,travel,study,conversation,exam"`
	MainUseCase  *string `json:"mainUseCase" nullable:"true" enum:"daily_life,work,travel,study,social"`
	Revision     int64   `json:"revision" minimum:"0"`
}
type LearningPreferencesOutput struct{ Body LearningPreferencesDTO }
type UpdateLearningPreferencesInput struct {
	Body struct {
		LearningGoal     string `json:"learningGoal" enum:"general,work,travel,study,conversation,exam"`
		MainUseCase      string `json:"mainUseCase" enum:"daily_life,work,travel,study,social"`
		ExpectedRevision int64  `json:"expectedRevision" minimum:"0"`
	}
}

func learningPreferencesDTO(p users.LearningPreferences) LearningPreferencesDTO {
	return LearningPreferencesDTO{LearningGoal: p.LearningGoal, MainUseCase: p.MainUseCase, Revision: p.Revision}
}

func RegisterLearningPreferences(api huma.API, svc *users.LearningPreferencesService, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{OperationID: "GetLearningPreferences", Method: http.MethodGet, Path: "/api/v1/learning-preferences", Summary: "Get current learning goal and focus", Tags: []string{"Learning preferences"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}, Errors: []int{401, 404, 409, 500}}, func(ctx context.Context, _ *struct{}) (*LearningPreferencesOutput, error) {
		p, err := svc.GetLearningPreferences(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapLearningPreferencesError(err)
		}
		return &LearningPreferencesOutput{Body: learningPreferencesDTO(p)}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "UpdateLearningPreferences", Method: http.MethodPatch, Path: "/api/v1/learning-preferences", Summary: "Change current learning goal and focus", Tags: []string{"Learning preferences"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}, Errors: []int{400, 401, 403, 404, 409, 422, 500}}, func(ctx context.Context, input *UpdateLearningPreferencesInput) (*LearningPreferencesOutput, error) {
		p, err := svc.UpdateLearningPreferences(ctx, RequesterUserID(ctx), users.LearningPreferencesUpdate{LearningGoal: input.Body.LearningGoal, MainUseCase: input.Body.MainUseCase, ExpectedRevision: input.Body.ExpectedRevision})
		if err != nil {
			return nil, mapLearningPreferencesError(err)
		}
		return &LearningPreferencesOutput{Body: learningPreferencesDTO(p)}, nil
	})
}

func mapLearningPreferencesError(err error) huma.StatusError {
	switch {
	case errors.Is(err, users.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	case errors.Is(err, users.ErrLearningPreferencesOnboarding):
		return huma.Error409Conflict("complete onboarding before changing learning direction")
	case errors.Is(err, users.ErrLearningPreferencesConflict):
		return huma.Error409Conflict("learning preferences changed; load your saved preferences")
	case errors.Is(err, users.ErrInvalidLearningPreferences):
		return huma.Error400BadRequest("invalid learning preferences")
	default:
		return huma.Error500InternalServerError("learning preferences unavailable")
	}
}
