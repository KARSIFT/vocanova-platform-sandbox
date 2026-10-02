package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/danielgtaylor/huma/v2"
)

type ListLessonsOutput struct {
	Body struct {
		Items []lessons.Summary `json:"items"`
	}
}
type LessonSessionOutput struct{ Body lessons.Session }
type StartLessonInput struct {
	LessonKey      string `path:"lessonKey" maxLength:"100"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" maxLength:"128"`
}
type GetLessonSessionInput struct {
	SessionID string `path:"sessionId" format:"uuid"`
}
type LessonActionInput struct {
	SessionID      string `path:"sessionId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" maxLength:"128"`
	Body           lessons.Action
}

// RegisterLessons exposes guided practice. No route writes SRS or mission credit.
func RegisterLessons(api huma.API, svc *lessons.Service, authSvc *auth.Service) {
	responses := map[string]*huma.Response{
		"400": {Description: "Invalid action"}, "401": {Description: "Authentication required"}, "403": {Description: "Invalid CSRF token"}, "404": {Description: "Lesson or session not found"}, "409": {Description: "Stale revision or idempotency conflict; refetch the session"}, "503": {Description: "Canonical lesson content unavailable"},
	}
	huma.Register(api, huma.Operation{OperationID: "ListLessons", Method: http.MethodGet, Path: "/api/v1/lessons", Summary: "List guided lessons and personal progress", Tags: []string{"Lessons"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(ctx context.Context, _ *struct{}) (*ListLessonsOutput, error) {
		items, err := svc.List(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapLessonError(err)
		}
		out := &ListLessonsOutput{}
		out.Body.Items = items
		return out, nil
	})
	huma.Register(api, huma.Operation{OperationID: "StartLesson", Method: http.MethodPost, Path: "/api/v1/lessons/{lessonKey}/sessions", Summary: "Start or resume a guided lesson", Tags: []string{"Lessons"}, Responses: responses, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *StartLessonInput) (*LessonSessionOutput, error) {
		session, err := svc.Start(ctx, RequesterUserID(ctx), in.LessonKey, in.IdempotencyKey)
		if err != nil {
			return nil, mapLessonError(err)
		}
		return &LessonSessionOutput{Body: *session}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetLessonSession", Method: http.MethodGet, Path: "/api/v1/lesson-sessions/{sessionId}", Summary: "Read an owned guided lesson session", Tags: []string{"Lessons"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(ctx context.Context, in *GetLessonSessionInput) (*LessonSessionOutput, error) {
		session, err := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.SessionID))
		if err != nil {
			return nil, mapLessonError(err)
		}
		return &LessonSessionOutput{Body: *session}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "SubmitLessonAction", Method: http.MethodPost, Path: "/api/v1/lesson-sessions/{sessionId}/actions", Summary: "Record a guided lesson action and return authoritative progress", Tags: []string{"Lessons"}, Responses: responses, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *LessonActionInput) (*LessonSessionOutput, error) {
		session, err := svc.Act(ctx, RequesterUserID(ctx), parseUUID(in.SessionID), in.Body, in.IdempotencyKey)
		if err != nil {
			return nil, mapLessonError(err)
		}
		return &LessonSessionOutput{Body: *session}, nil
	})
}
func mapLessonError(err error) error {
	switch {
	case errors.Is(err, lessons.ErrNotFound):
		return huma.Error404NotFound("Lesson or session not found")
	case errors.Is(err, lessons.ErrConflict):
		return huma.Error409Conflict("This lesson changed. Reload it to continue.")
	case errors.Is(err, lessons.ErrInvalid):
		return huma.Error400BadRequest("Invalid lesson action")
	case errors.Is(err, lessons.ErrContentUnavailable):
		return huma.Error503ServiceUnavailable("This lesson is unavailable right now")
	default:
		return huma.Error500InternalServerError("We could not update this lesson. Try again.")
	}
}
