package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/danielgtaylor/huma/v2"
)

// SettingsDTO is the public projection of a learner's Settings.
// It is the response shape for both GET /api/v1/settings and the
// successful PATCH /api/v1/settings. Every field is always
// present in the JSON so the frontend can branch on presence
// rather than nullability.
type SettingsDTO struct {
	DailyReviewTarget      int    `json:"dailyReviewTarget" minimum:"5" maximum:"100" doc:"Daily review target (5-100)"`
	ReviewIntervalPreset   string `json:"reviewIntervalPreset" enum:"vocanova_default,wordup_like,custom" doc:"Review interval preset"`
	AppLanguage            string `json:"appLanguage" enum:"en" doc:"Persisted app language preference; only 'en' is accepted at launch"`
	NotificationsEnabled   bool   `json:"notificationsEnabled" doc:"Whether the learner opts into in-app notifications"`
	MarketingEmailsEnabled bool   `json:"marketingEmailsEnabled" doc:"Whether the learner opts into marketing emails"`
	DisplayName            string `json:"displayName" doc:"Learner-controlled display name (0-80 chars)"`
}

// GetSettingsOutput is the response body for GET /api/v1/settings.
type GetSettingsOutput struct {
	Body SettingsDTO
}

// UpdateSettingsInput is the request body for PATCH
// /api/v1/settings. Every field is a pointer so the service layer
// can distinguish "field omitted" (preserves the stored value)
// from "field set to the zero value" (writes the zero value).
// additionalProperties=false is the JSON-Schema knob that
// rejects unknown fields at the Huma boundary (DOC-07 §3).
type UpdateSettingsInput struct {
	Body struct {
		DailyReviewTarget      *int    `json:"dailyReviewTarget,omitempty" minimum:"5" maximum:"100" doc:"Daily review target (5-100)"`
		ReviewIntervalPreset   *string `json:"reviewIntervalPreset,omitempty" enum:"vocanova_default,wordup_like,custom" doc:"Review interval preset"`
		AppLanguage            *string `json:"appLanguage,omitempty" enum:"en" doc:"App language; only 'en' is accepted at launch"`
		NotificationsEnabled   *bool   `json:"notificationsEnabled,omitempty" doc:"Notifications opt-in"`
		MarketingEmailsEnabled *bool   `json:"marketingEmailsEnabled,omitempty" doc:"Marketing emails opt-in"`
		DisplayName            *string `json:"displayName,omitempty" maxLength:"80" doc:"Display name (0-80 chars)"`
	}
}

// UpdateSettingsOutput is the response body for PATCH
// /api/v1/settings. It echoes the post-merge Settings projection
// (the same shape as GET) so the frontend can refresh its
// local state from a single round-trip.
type UpdateSettingsOutput struct {
	Body SettingsDTO
}

// RegisterSettings wires the GET/PATCH /api/v1/settings routes
// on api. The routes are requester-scoped: the userID is always
// taken from the session and there is no path or query parameter
// to address another learner's settings. authSvc is the auth
// service used to enforce double-submit CSRF on the PATCH route;
// the GET route requires only a session.
func RegisterSettings(api huma.API, svc *users.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "GetSettings",
		Method:      http.MethodGet,
		Path:        "/api/v1/settings",
		Summary:     "Get the requester's settings",
		Tags:        []string{"Settings"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"404": {Description: "User not found"},
		},
	}, func(ctx context.Context, input *struct{}) (*GetSettingsOutput, error) {
		uid := RequesterUserID(ctx)
		settings, err := svc.GetSettings(ctx, uid)
		if err != nil {
			return nil, mapSettingsError(err)
		}
		return &GetSettingsOutput{Body: settingsToDTO(settings)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "UpdateSettings",
		Method:      http.MethodPatch,
		Path:        "/api/v1/settings",
		Summary:     "Update the requester's settings",
		Tags:        []string{"Settings"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid settings update"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "User not found"},
		},
	}, func(ctx context.Context, input *UpdateSettingsInput) (*UpdateSettingsOutput, error) {
		uid := RequesterUserID(ctx)
		update := users.SettingsUpdate{
			DailyReviewTarget:      input.Body.DailyReviewTarget,
			ReviewIntervalPreset:   input.Body.ReviewIntervalPreset,
			AppLanguage:            input.Body.AppLanguage,
			NotificationsEnabled:   input.Body.NotificationsEnabled,
			MarketingEmailsEnabled: input.Body.MarketingEmailsEnabled,
			DisplayName:            input.Body.DisplayName,
		}
		settings, err := svc.UpdateSettings(ctx, uid, update)
		if err != nil {
			return nil, mapSettingsError(err)
		}
		return &UpdateSettingsOutput{Body: settingsToDTO(settings)}, nil
	})
}

func settingsToDTO(s users.Settings) SettingsDTO {
	return SettingsDTO{
		DailyReviewTarget:      s.DailyReviewTarget,
		ReviewIntervalPreset:   s.ReviewIntervalPreset,
		AppLanguage:            s.AppLanguage,
		NotificationsEnabled:   s.NotificationsEnabled,
		MarketingEmailsEnabled: s.MarketingEmailsEnabled,
		DisplayName:            s.DisplayName,
	}
}

func mapSettingsError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, users.ErrSettingsNotFound):
		return huma.Error404NotFound("user not found")
	case errors.Is(err, users.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	}
	// Validation errors (dailyReviewTarget range, enum, appLanguage,
	// displayName length) bubble up from the service's Validate()
	// call. Map them to 400 so the frontend can present a clear
	// invalid-submission signal without a 500.
	msg := err.Error()
	if msg != "" {
		return huma.Error400BadRequest(msg)
	}
	return huma.Error500InternalServerError("internal error")
}
