package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/stories"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type StoryLibraryOutput struct{ Body stories.StoryLibrary }
type StoryReadingInput struct {
	StoryKey string `path:"storyKey" minLength:"1" maxLength:"100"`
}
type StoryReadingOutput struct{ Body stories.StoryReading }
type StoryStartInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           stories.StoryStartRequest
}
type StorySessionInput struct {
	SessionID string `path:"sessionId" format:"uuid"`
}
type StoryActionInput struct {
	SessionID      string `path:"sessionId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           stories.StoryAction
}
type StorySessionOutput struct{ Body stories.StorySession }

func RegisterStories(api huma.API, svc *stories.Service, authSvc *auth.Service) {
	read := []func(huma.Context, func(huma.Context)){RequireAuth()}
	write := []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}
	huma.Register(api, huma.Operation{OperationID: "ListStories", Method: http.MethodGet, Path: "/api/v1/stories", Summary: "Original A2–B1 story library and your latest attempts", Tags: []string{"Stories"}, Middlewares: read}, func(ctx context.Context, _ *struct{}) (*StoryLibraryOutput, error) {
		out, err := svc.List(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapStoryError(err)
		}
		return &StoryLibraryOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetStory", Method: http.MethodGet, Path: "/api/v1/stories/{storyKey}", Summary: "Read the full original story and vocabulary without credit", Tags: []string{"Stories"}, Middlewares: read}, func(ctx context.Context, in *StoryReadingInput) (*StoryReadingOutput, error) {
		out, err := svc.Read(ctx, RequesterUserID(ctx), in.StoryKey)
		if err != nil {
			return nil, mapStoryError(err)
		}
		return &StoryReadingOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "StartStorySession", Method: http.MethodPost, Path: "/api/v1/story-sessions", Summary: "Start a repeatable story using a frozen original content snapshot", Tags: []string{"Stories"}, Middlewares: write}, func(ctx context.Context, in *StoryStartInput) (*StorySessionOutput, error) {
		out, err := svc.Start(ctx, RequesterUserID(ctx), in.Body, in.IdempotencyKey)
		if err != nil {
			return nil, mapStoryError(err)
		}
		return &StorySessionOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetStorySession", Method: http.MethodGet, Path: "/api/v1/story-sessions/{sessionId}", Summary: "Resume your saved story", Tags: []string{"Stories"}, Middlewares: read}, func(ctx context.Context, in *StorySessionInput) (*StorySessionOutput, error) {
		out, err := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.SessionID))
		if err != nil {
			return nil, mapStoryError(err)
		}
		return &StorySessionOutput{Body: *out}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "SubmitStoryAction", Method: http.MethodPost, Path: "/api/v1/story-sessions/{sessionId}/actions", Summary: "Grade a story choice or continue after a line or correct retry", Tags: []string{"Stories"}, Middlewares: write}, func(ctx context.Context, in *StoryActionInput) (*StorySessionOutput, error) {
		out, err := svc.Act(ctx, RequesterUserID(ctx), parseUUID(in.SessionID), in.Body, in.IdempotencyKey)
		if err != nil {
			return nil, mapStoryError(err)
		}
		return &StorySessionOutput{Body: *out}, nil
	})
}
func mapStoryError(err error) error {
	switch {
	case errors.Is(err, stories.ErrNotFound):
		return huma.Error404NotFound("Story not found")
	case errors.Is(err, stories.ErrInvalid):
		return huma.Error400BadRequest("Invalid story request")
	case errors.Is(err, stories.ErrConflict):
		return huma.Error409Conflict("This story changed. Load your saved progress to continue.")
	case errors.Is(err, stories.ErrContentUnavailable):
		return huma.Error503ServiceUnavailable("This story content is unavailable right now.")
	default:
		return huma.Error500InternalServerError("We could not load or update your story. Try again.")
	}
}
