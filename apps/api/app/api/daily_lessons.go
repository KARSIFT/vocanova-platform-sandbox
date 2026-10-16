package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dailylesson"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type DailyLessonOutput struct{ Body dailylesson.Session }
type CurrentDailyLessonOutput struct{ Body dailylesson.Current }
type StartDailyLessonInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
}
type GetDailyLessonInput struct {
	SessionID string `path:"sessionId" format:"uuid"`
}
type DailyLessonActionInput struct {
	SessionID      string `path:"sessionId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           dailylesson.Action
}

func RegisterDailyLessons(api huma.API, svc *dailylesson.Service, authSvc *auth.Service) {
	private := func(ctx huma.Context, next func(huma.Context)) { ctx.SetHeader("Cache-Control", "no-store"); next(ctx) }
	responses := map[string]*huma.Response{"400": {Description: "Invalid action"}, "401": {Description: "Authentication required"}, "403": {Description: "Invalid CSRF token"}, "404": {Description: "Session not found"}, "409": {Description: "Stale revision, conflicting retry, or no words due"}, "503": {Description: "Meaning choices unavailable"}}
	huma.Register(api, huma.Operation{OperationID: "GetCurrentDailyLesson", Method: http.MethodGet, Path: "/api/v1/daily-lessons/current", Summary: "Read your current due-word lesson and remaining due count", Tags: []string{"Daily lessons"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth()}}, func(ctx context.Context, _ *struct{}) (*CurrentDailyLessonOutput, error) {
		out, e := svc.Current(ctx, RequesterUserID(ctx))
		if e != nil {
			return nil, mapDailyLessonError(e)
		}
		return &CurrentDailyLessonOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "StartDailyLesson", Method: http.MethodPost, Path: "/api/v1/daily-lessons", Summary: "Start or resume a lesson from your actual due saved meanings", Tags: []string{"Daily lessons"}, Responses: responses, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *StartDailyLessonInput) (*DailyLessonOutput, error) {
		out, e := svc.Start(ctx, RequesterUserID(ctx), in.IdempotencyKey)
		if e != nil {
			return nil, mapDailyLessonError(e)
		}
		return &DailyLessonOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetDailyLesson", Method: http.MethodGet, Path: "/api/v1/daily-lessons/{sessionId}", Summary: "Read your durable daily lesson", Tags: []string{"Daily lessons"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth()}}, func(ctx context.Context, in *GetDailyLessonInput) (*DailyLessonOutput, error) {
		out, e := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.SessionID))
		if e != nil {
			return nil, mapDailyLessonError(e)
		}
		return &DailyLessonOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "SubmitDailyLessonAction", Method: http.MethodPost, Path: "/api/v1/daily-lessons/{sessionId}/actions", Summary: "Check a meaning answer and atomically record confirmed review progress", Tags: []string{"Daily lessons"}, Responses: responses, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *DailyLessonActionInput) (*DailyLessonOutput, error) {
		out, e := svc.Act(ctx, RequesterUserID(ctx), parseUUID(in.SessionID), in.Body, in.IdempotencyKey)
		if e != nil {
			return nil, mapDailyLessonError(e)
		}
		return &DailyLessonOutput{Body: *out}, nil
	})
}
func mapDailyLessonError(e error) error {
	switch {
	case errors.Is(e, dailylesson.ErrNotFound):
		return huma.Error404NotFound("Daily lesson not found")
	case errors.Is(e, dailylesson.ErrInvalid):
		return huma.Error400BadRequest("Invalid daily lesson action")
	case errors.Is(e, dailylesson.ErrConflict):
		return huma.Error409Conflict("This lesson changed. Reload it to continue.")
	case errors.Is(e, dailylesson.ErrNoDue):
		return huma.Error409Conflict("No words are due. Discover a word to start learning.")
	case errors.Is(e, dailylesson.ErrContentUnavailable):
		return huma.Error503ServiceUnavailable("Meaning choices are unavailable. Try again.")
	default:
		return huma.Error500InternalServerError("We could not update your daily lesson. Retry safely.")
	}
}
