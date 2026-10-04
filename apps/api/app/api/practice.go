package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type StartPracticeInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           practice.StartRequest
}
type PracticeSessionInput struct {
	SessionID string `path:"sessionId" format:"uuid"`
}
type PracticeActionInput struct {
	SessionID      string `path:"sessionId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           practice.Action
}
type PracticeSessionOutput struct{ Body practice.Session }
type ListPracticeOutput struct{ Body practice.List }

func RegisterPractice(api huma.API, svc *practice.Service, authSvc *auth.Service) {
	read := []func(huma.Context, func(huma.Context)){RequireAuth()}
	write := []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}
	huma.Register(api, huma.Operation{OperationID: "ListPracticeSessions", Method: http.MethodGet, Path: "/api/v1/practice-sessions", Summary: "Recent practice sessions and total supported unresolved mistakes", Tags: []string{"Practice"}, Middlewares: read}, func(ctx context.Context, _ *struct{}) (*ListPracticeOutput, error) {
		out, err := svc.List(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapPracticeError(err)
		}
		return &ListPracticeOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "StartPracticeSession", Method: http.MethodPost, Path: "/api/v1/practice-sessions", Summary: "Start a new repeatable practice session without review credit", Tags: []string{"Practice"}, Middlewares: write}, func(ctx context.Context, in *StartPracticeInput) (*PracticeSessionOutput, error) {
		out, err := svc.Start(ctx, RequesterUserID(ctx), in.Body, in.IdempotencyKey)
		if err != nil {
			return nil, mapPracticeError(err)
		}
		return &PracticeSessionOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetPracticeSession", Method: http.MethodGet, Path: "/api/v1/practice-sessions/{sessionId}", Summary: "Resume your practice session", Tags: []string{"Practice"}, Middlewares: read}, func(ctx context.Context, in *PracticeSessionInput) (*PracticeSessionOutput, error) {
		out, err := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.SessionID))
		if err != nil {
			return nil, mapPracticeError(err)
		}
		return &PracticeSessionOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "SubmitPracticeAction", Method: http.MethodPost, Path: "/api/v1/practice-sessions/{sessionId}/actions", Summary: "Grade an answer or advance your practice session", Tags: []string{"Practice"}, Middlewares: write}, func(ctx context.Context, in *PracticeActionInput) (*PracticeSessionOutput, error) {
		out, err := svc.Act(ctx, RequesterUserID(ctx), parseUUID(in.SessionID), in.Body, in.IdempotencyKey)
		if err != nil {
			return nil, mapPracticeError(err)
		}
		return &PracticeSessionOutput{Body: *out}, nil
	})
}
func mapPracticeError(err error) error {
	switch {
	case errors.Is(err, practice.ErrNotFound):
		return huma.Error404NotFound("Practice session not found")
	case errors.Is(err, practice.ErrInvalid):
		return huma.Error400BadRequest("Invalid practice request")
	case errors.Is(err, practice.ErrConflict):
		return huma.Error409Conflict("This practice changed. Reload it to continue.")
	case errors.Is(err, practice.ErrListEmpty):
		return huma.Error409Conflict("This list has no supported meanings to practise. Add words from the starter course first.")
	case errors.Is(err, practice.ErrNoMistakes):
		return huma.Error409Conflict("There are no supported mistakes to practise right now.")
	case errors.Is(err, practice.ErrContentUnavailable):
		return huma.Error503ServiceUnavailable("This practice content is unavailable right now.")
	default:
		return huma.Error500InternalServerError("We could not update your practice. Try again.")
	}
}
