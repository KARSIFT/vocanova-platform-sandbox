package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// CreatePersonalDataExportInput has no body: identity is always derived from
// the authenticated session, and the key makes an accidental double-submit
// safe. The response is a portable JSON document downloaded by Settings.
type CreatePersonalDataExportInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key (DOC-07)"`
}

type CreatePersonalDataExportOutput struct {
	Body PersonalDataExportDTO
}

// PersonalDataExportDTO is intentionally additive: each collection preserves
// the learner's own persisted record shape without exposing internal API
// secrets. The stable top-level keys make the downloaded document practical
// for a learner to inspect or move to another service.
type PersonalDataExportDTO struct {
	SchemaVersion           string         `json:"schemaVersion"`
	ExportedAt              string         `json:"exportedAt,omitempty" format:"date-time"`
	Profile                 map[string]any `json:"profile"`
	Settings                map[string]any `json:"settings"`
	OnboardingProfile       any            `json:"onboardingProfile" nullable:"true"`
	SavedWords              []any          `json:"savedWords"`
	ReviewHistory           []any          `json:"reviewHistory"`
	SentenceFeedbackHistory []any          `json:"sentenceFeedbackHistory"`
	DailyMissions           []any          `json:"dailyMissions"`
	DailyActivity           []any          `json:"dailyActivity"`
	ConfidencePointLedger   []any          `json:"confidencePointLedger"`
	GraceDayLedger          []any          `json:"graceDayLedger"`
	StreakState             any            `json:"streakState" nullable:"true"`
}

func RegisterPersonalDataExports(api huma.API, svc *accounts.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "CreatePersonalDataExport",
		Method:      http.MethodPost, Path: "/api/v1/personal-data-export",
		Summary:     "Download the requester's personal data as JSON",
		Description: "Synchronous MVP export. Includes learner-visible profile, settings, saved-word, review, practice, feedback, and progress history; excludes authentication secrets, hidden prompts, provider credentials, and internal abuse/report classifications.",
		Tags:        []string{"Account"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"}, "403": {Description: "Invalid CSRF token"},
			"404": {Description: "User not found"}, "409": {Description: "Idempotency-Key conflict"},
			"422": {Description: "Missing or invalid Idempotency-Key"}, "429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *CreatePersonalDataExportInput) (*CreatePersonalDataExportOutput, error) {
		c := authHumaContext(ctx)
		payload, err := svc.ExportPersonalData(ctx, RequesterUserID(ctx).String(), clientIPFromHuma(c), sessionTokenFromHuma(c, authSvc), input.IdempotencyKey)
		if err != nil {
			return nil, mapDataExportError(err)
		}
		var body PersonalDataExportDTO
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, huma.Error500InternalServerError("internal error")
		}
		return &CreatePersonalDataExportOutput{Body: body}, nil
	})
}

func mapDataExportError(err error) huma.StatusError {
	switch {
	case errors.Is(err, accounts.ErrDataExportIdempotencyKeyRequired):
		return huma.Error400BadRequest("idempotency key required")
	case errors.Is(err, accounts.ErrDataExportIdempotencyConflict):
		return huma.Error409Conflict("idempotency key conflict")
	case errors.Is(err, accounts.ErrDataExportRateLimited):
		return huma.Error429TooManyRequests("personal data export rate limited")
	case errors.Is(err, accounts.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
