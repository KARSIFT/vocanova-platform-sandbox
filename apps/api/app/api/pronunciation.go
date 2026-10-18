package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/pronunciation"
	"github.com/danielgtaylor/huma/v2"
)

type PronunciationInput struct {
	Body struct {
		Text            string `json:"text" minLength:"1" maxLength:"500"`
		MeaningID       string `json:"meaningId,omitempty" format:"uuid"`
		MeaningRevision string `json:"meaningRevision,omitempty" pattern:"^[0-9a-f]{64}$"`
	}
}
type PronunciationOutput struct{ Body pronunciation.Audio }

func RegisterPronunciation(api huma.API, svc *pronunciation.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{OperationID: "GetPronunciation", Method: http.MethodPost, Path: "/api/v1/pronunciation", Summary: "Play cached natural pronunciation for canonical public teaching content", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *PronunciationInput) (*PronunciationOutput, error) {
		out, err := svc.SpeakMeaningRevision(ctx, RequesterUserID(ctx), in.Body.Text, in.Body.MeaningID, in.Body.MeaningRevision)
		switch {
		case errors.Is(err, pronunciation.ErrInvalid):
			return nil, huma.Error400BadRequest("Choose a word or example from your lesson.")
		case errors.Is(err, pronunciation.ErrNotFound):
			return nil, huma.Error404NotFound("This pronunciation is not available.")
		case errors.Is(err, pronunciation.ErrRateLimited), errors.Is(err, pronunciation.ErrBusy):
			return nil, huma.Error429TooManyRequests("Natural pronunciation is busy. Try device pronunciation.")
		case err != nil:
			return nil, huma.Error503ServiceUnavailable("Natural pronunciation is unavailable. Try device pronunciation.")
		default:
			return &PronunciationOutput{Body: *out}, nil
		}
	})
}
