package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// RequestEmailChangeLinkInput is the request body for
// POST /api/v1/settings/email-change-links. The newEmail field
// is the destination the requester wants to switch to; the
// current sign-in address is taken from the session and is never
// trusted from the body.
type RequestEmailChangeLinkInput struct {
	Body struct {
		NewEmail string `json:"newEmail" format:"email" doc:"The destination email the requester wants to switch to"`
	}
}

// RequestEmailChangeLinkOutput is intentionally empty: the
// request response is unconditionally generic, so the new-email
// registration status is not observable through the request
// outcome (VOC-031-D05 anti-enumeration posture, mirroring the
// magic-link request).
type RequestEmailChangeLinkOutput struct{}

// ConsumeEmailChangeLinkInput is the request body for
// POST /api/v1/settings/email-change-links/consume. The token
// is the only form the requester supplies; the API layer never
// sees the email itself at consume time.
type ConsumeEmailChangeLinkInput struct {
	Body struct {
		Token string `json:"token" doc:"The single-use token from the email-change confirmation link"`
	}
}

// ConsumeEmailChangeLinkOutput is the post-confirm response.
// The new email is the one that is now in effect on users.email;
// the frontend can refresh its local state from this body.
type ConsumeEmailChangeLinkOutput struct {
	Body ConsumeEmailChangeLinkDTO
}

// ConsumeEmailChangeLinkDTO is the public projection of an
// EmailChangeResult. The old email is included so the frontend
// can show the learner which address the security notification
// was dispatched to; the security-control semantics of the
// notification are owned by the backend, not the frontend.
type ConsumeEmailChangeLinkDTO struct {
	Email         string `json:"email" format:"email" doc:"The new email address that is now in effect on users.email"`
	PreviousEmail string `json:"previousEmail" format:"email" doc:"The sign-in email address that was in effect before the change"`
	ChangedAt     string `json:"changedAt" format:"date-time" doc:"When the email change took effect"`
}

// RegisterEmailChangeLinks wires the two T03 routes on api.
// The routes are requester-scoped: userID is always taken from
// the session, never from a path or query parameter. Both
// routes require an authenticated session; both require
// double-submit CSRF on the state-changing POSTs.
func RegisterEmailChangeLinks(api huma.API, svc *accounts.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "RequestEmailChangeLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/settings/email-change-links",
		Summary:     "Request a single-use email-change link",
		Tags:        []string{"Settings"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid new email"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *RequestEmailChangeLinkInput) (*RequestEmailChangeLinkOutput, error) {
		uid := RequesterUserID(ctx)
		c := authHumaContext(ctx)
		_, err := svc.RequestEmailChangeLink(ctx, uid, clientIPFromHuma(c), input.Body.NewEmail)
		if err != nil {
			return nil, mapEmailChangeError(err)
		}
		return &RequestEmailChangeLinkOutput{}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "ConsumeEmailChangeLink",
		Method:      http.MethodPost,
		Path:        "/api/v1/settings/email-change-links/consume",
		Summary:     "Consume a single-use email-change link",
		Tags:        []string{"Settings"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Invalid or expired email change link"},
			"403": {Description: "Invalid CSRF token"},
			"409": {Description: "New email is already registered to another account"},
			"429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *ConsumeEmailChangeLinkInput) (*ConsumeEmailChangeLinkOutput, error) {
		c := authHumaContext(ctx)
		res, err := svc.ConsumeEmailChangeLink(ctx, clientIPFromHuma(c), input.Body.Token)
		if err != nil {
			return nil, mapEmailChangeError(err)
		}
		return &ConsumeEmailChangeLinkOutput{Body: ConsumeEmailChangeLinkDTO{
			Email:         res.NewEmail,
			PreviousEmail: res.OldEmail,
			ChangedAt:     res.ChangedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}}, nil
	})
}

// mapEmailChangeError maps the accounts package's exported
// errors to stable Huma responses. The 401 path is the
// deliberately non-distinguishing invalid-link error; the 409
// is the duplicate-email-conflict error from the partial unique
// index; 429 is the rate-limit error; 400 is the
// input-validation error.
func mapEmailChangeError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, accounts.ErrInvalidEmailChangeLink):
		return huma.Error401Unauthorized("invalid or expired email change link")
	case errors.Is(err, accounts.ErrEmailAlreadyRegistered):
		return huma.Error409Conflict("new email is already registered to another account")
	case errors.Is(err, accounts.ErrUserNotFound):
		return huma.Error401Unauthorized("invalid or expired email change link")
	case errors.Is(err, accounts.ErrEmailChangeRateLimited):
		return huma.Error429TooManyRequests("email change rate limited")
	case errors.Is(err, accounts.ErrEmailChangeInvalidEmail):
		return huma.Error400BadRequest("invalid email address")
	}
	return huma.Error500InternalServerError("internal error")
}
