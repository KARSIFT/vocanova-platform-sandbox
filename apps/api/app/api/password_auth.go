package api

import (
	"context"
	"errors"
	"mime"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/password"
	"github.com/danielgtaylor/huma/v2"
)

type passwordSignupInput struct {
	Body struct {
		Email       string `json:"email" format:"email"`
		Password    string `json:"password" writeOnly:"true" doc:"Password of 15 to 128 characters"`
		DisplayName string `json:"displayName,omitempty" maxLength:"80"`
	}
}
type passwordEmailInput struct {
	Body struct {
		Email string `json:"email" format:"email" maxLength:"254"`
	}
}
type passwordTokenInput struct {
	Body struct {
		Token string `json:"token" writeOnly:"true"`
	}
}
type passwordResetInput struct {
	Body struct {
		Token    string `json:"token" writeOnly:"true"`
		Password string `json:"password" writeOnly:"true" doc:"Password of 15 to 128 characters"`
	}
}
type passwordLoginInput struct {
	Body struct {
		Email    string `json:"email" format:"email"`
		Password string `json:"password" writeOnly:"true" doc:"Password of 15 to 128 characters"`
	}
}
type passwordLoginOutput struct{ Body CurrentUser }
type emptyOutput struct{}

func RegisterPasswordAuth(api huma.API, svc *password.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{OperationID: "RequestPasswordSignup", Method: http.MethodPost, Path: "/api/v1/auth/password/signups", Tags: []string{"Authentication"}, MaxBodyBytes: 8192, Middlewares: []func(huma.Context, func(huma.Context)){passwordJSONOnly}, DefaultStatus: http.StatusNoContent, Errors: []int{401, 415, 422, 429, 500, 503}}, func(ctx context.Context, in *passwordSignupInput) (*emptyOutput, error) {
		c := authHumaContext(ctx)
		if err := svc.RequestSignup(ctx, clientIPFromHuma(c), in.Body.Email, in.Body.Password, in.Body.DisplayName); err != nil {
			return nil, mapPasswordError(err, true)
		}
		return &emptyOutput{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "VerifyPasswordSignup", Method: http.MethodPost, Path: "/api/v1/auth/password/signups/verify", Tags: []string{"Authentication"}, MaxBodyBytes: 8192, Middlewares: []func(huma.Context, func(huma.Context)){passwordJSONOnly}, DefaultStatus: http.StatusNoContent, Errors: []int{401, 415, 422, 429, 500, 503}}, func(ctx context.Context, in *passwordTokenInput) (*emptyOutput, error) {
		if err := svc.VerifySignup(ctx, clientIPFromHuma(authHumaContext(ctx)), in.Body.Token); err != nil {
			return nil, mapPasswordError(err, false)
		}
		return &emptyOutput{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "RequestPasswordReset", Method: http.MethodPost, Path: "/api/v1/auth/password/reset-requests", Tags: []string{"Authentication"}, MaxBodyBytes: 8192, Middlewares: []func(huma.Context, func(huma.Context)){passwordJSONOnly}, DefaultStatus: http.StatusNoContent, Errors: []int{401, 415, 422, 429, 500, 503}}, func(ctx context.Context, in *passwordEmailInput) (*emptyOutput, error) {
		c := authHumaContext(ctx)
		if err := svc.RequestReset(ctx, clientIPFromHuma(c), in.Body.Email); err != nil {
			return nil, mapPasswordError(err, true)
		}
		return &emptyOutput{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "ResetPassword", Method: http.MethodPost, Path: "/api/v1/auth/password/resets", Tags: []string{"Authentication"}, MaxBodyBytes: 8192, Middlewares: []func(huma.Context, func(huma.Context)){passwordJSONOnly}, DefaultStatus: http.StatusNoContent, Errors: []int{401, 415, 422, 429, 500, 503}}, func(ctx context.Context, in *passwordResetInput) (*emptyOutput, error) {
		if err := svc.Reset(ctx, clientIPFromHuma(authHumaContext(ctx)), in.Body.Token, in.Body.Password); err != nil {
			return nil, mapPasswordError(err, false)
		}
		return &emptyOutput{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "LoginWithPassword", Method: http.MethodPost, Path: "/api/v1/auth/password/login", Tags: []string{"Authentication"}, MaxBodyBytes: 8192, Middlewares: []func(huma.Context, func(huma.Context)){passwordJSONOnly}, DefaultStatus: http.StatusOK, Errors: []int{401, 415, 422, 429, 500, 503}}, func(ctx context.Context, in *passwordLoginInput) (*passwordLoginOutput, error) {
		c := authHumaContext(ctx)
		_, csrf := authSvc.IssueCSRFCookie()
		if csrf == nil {
			return nil, huma.Error500InternalServerError("unable to create session")
		}
		u, s, t, err := svc.Login(ctx, clientIPFromHuma(c), in.Body.Email, in.Body.Password)
		if err != nil {
			return nil, mapPasswordError(err, false)
		}
		c.AppendHeader("Set-Cookie", authSvc.SessionCookie(t, s.ExpiresAt).String())
		c.AppendHeader("Set-Cookie", csrf.String())
		body := currentUserFromAuth(u)
		body.OnboardingStatus = "not_started"
		if status, lookupErr := onboardingStatusLookup(ctx, u.ID); lookupErr == nil && status != "" {
			body.OnboardingStatus = status
		}
		return &passwordLoginOutput{Body: body}, nil
	})
}
func mapPasswordError(err error, generic bool) huma.StatusError {
	if errors.Is(err, auth.ErrRateLimited) {
		return huma.Error429TooManyRequests("rate limited")
	}
	if errors.Is(err, password.ErrDisabled) || errors.Is(err, auth.ErrSignupsDisabled) {
		return huma.Error503ServiceUnavailable("password sign-in is disabled")
	}
	if errors.Is(err, password.ErrInvalidPassword) {
		return huma.Error422UnprocessableEntity("password must be 15 to 128 characters")
	}
	if generic {
		return huma.Error500InternalServerError("unable to process request")
	}
	if errors.Is(err, password.ErrInvalidToken) || errors.Is(err, password.ErrInvalidCredentials) {
		return huma.Error401Unauthorized("invalid credentials or expired link")
	}
	return huma.Error500InternalServerError("internal error")
}

// Requiring JSON blocks cross-site HTML forms from signing a browser into an
// attacker-controlled account. Cross-origin JSON requests require the existing
// allowlisted CORS preflight before a browser may send them.
func passwordJSONOnly(ctx huma.Context, next func(huma.Context)) {
	mediaType, _, err := mime.ParseMediaType(ctx.Header("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeHumaError(ctx, huma.Error415UnsupportedMediaType("application/json is required"))
		return
	}
	next(ctx)
}
