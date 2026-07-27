package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// CreateAccountDeletionRequestInput is the request body for
// POST /api/v1/account-deletion-requests. The body is
// deliberately empty: the requester-scoped user identity is
// taken from the session, never from the body, and the
// DOC-07-required Idempotency-Key is supplied via header. The
// requester is the *current* user; no path or body parameter
// may override it.
type CreateAccountDeletionRequestInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key (DOC-07)"`
}

// CreateAccountDeletionRequestDTO is the public projection
// of the post-deactivation state. The user is already
// deactivated at this point; the requester is told only what
// they need to render a clear "your account has been
// scheduled for deletion" confirmation (the dates, the
// deactivation status) and the Idempotency-Key fingerprint
// the response can be cross-checked against on a retry.
type CreateAccountDeletionRequestDTO struct {
	Status         string `json:"status" doc:"Deactivation status; always 'deactivated' on a successful request"`
	UserID         string `json:"userId" format:"uuid" doc:"The deactivated user's id"`
	RequestedAt    string `json:"requestedAt" format:"date-time" doc:"When the deactivation took effect"`
	PurgeAfter     string `json:"purgeAfter" format:"date-time" doc:"When the anonymization sweep will run"`
	IdempotencyKey string `json:"idempotencyKey" doc:"The Idempotency-Key that produced this result"`
	Replayed       bool   `json:"replayed" doc:"True when this call was a no-op replay of a prior request with the same key"`
}

// CreateAccountDeletionRequestOutput wraps the DTO so Huma
// can render the response shape.
type CreateAccountDeletionRequestOutput struct {
	Body CreateAccountDeletionRequestDTO
}

// RegisterAccountDeletionRequests wires the single
// POST /api/v1/account-deletion-requests route on api. The
// route is requester-scoped: the user is always taken from
// the session, never from a path or body parameter. The
// route requires an authenticated session, double-submit
// CSRF, and an Idempotency-Key header (DOC-07). On success,
// the session is invalidated server-side: every active
// session for the account is revoked in the same
// transaction. The client should follow up with a logout
// request to clear the cookie; the API layer renders a
// 200/202 with a clear post-deletion body.
func RegisterAccountDeletionRequests(api huma.API, svc *accounts.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "CreateAccountDeletionRequest",
		Method:      http.MethodPost,
		Path:        "/api/v1/account-deletion-requests",
		Summary:     "Deactivate the requester's account and schedule anonymization",
		Tags:        []string{"Account"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "User not found"},
			"409": {Description: "Account deletion is already in flight for this user, or the Idempotency-Key fingerprint conflicts"},
			"422": {Description: "Missing or invalid Idempotency-Key"},
			"429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *CreateAccountDeletionRequestInput) (*CreateAccountDeletionRequestOutput, error) {
		uid := RequesterUserID(ctx)
		c := authHumaContext(ctx)
		res, err := svc.CreateAccountDeletionRequest(ctx, uid.String(), clientIPFromHuma(c), sessionTokenFromHuma(c, authSvc), input.IdempotencyKey)
		if err != nil {
			return nil, mapAccountDeletionError(err)
		}
		return &CreateAccountDeletionRequestOutput{Body: CreateAccountDeletionRequestDTO{
			Status:         res.Status,
			UserID:         res.UserID.String(),
			RequestedAt:    res.RequestedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			PurgeAfter:     res.PurgeAfter.UTC().Format("2006-01-02T15:04:05Z07:00"),
			IdempotencyKey: res.IdempotencyKey,
			Replayed:       res.Replayed,
		}}, nil
	})
}

// mapAccountDeletionError maps the accounts package's
// exported errors to stable Huma responses. The 401 path is
// reserved for "missing or invalid session"; the 400 path
// is the missing-Idempotency-Key validation error; the 404
// is the user-not-found; the 409 covers the
// already-in-flight discipline and the Idempotency-Key
// fingerprint conflict; the 429 is the rate-limit error;
// anything else is a 500.
func mapAccountDeletionError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, accounts.ErrAccountDeletionIdempotencyKeyRequired):
		return huma.Error400BadRequest("idempotency key required")
	case errors.Is(err, accounts.ErrAccountDeletionIdempotencyConflict):
		return huma.Error409Conflict("idempotency key conflict")
	case errors.Is(err, accounts.ErrAccountDeletionAlreadyInFlight):
		return huma.Error409Conflict("account deletion already in flight")
	case errors.Is(err, accounts.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	case errors.Is(err, accounts.ErrAccountDeletionRateLimited):
		return huma.Error429TooManyRequests("account deletion rate limited")
	}
	return huma.Error500InternalServerError("internal error")
}
