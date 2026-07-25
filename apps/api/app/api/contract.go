// Package api owns the public HTTP contract. Persistence entities must never be
// used as request or response bodies here.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// CurrentUser is the minimal public identity projection. It intentionally omits
// database IDs, provider subjects, tokens, and session metadata.
type CurrentUser struct {
	Email           *string    `json:"email,omitempty" format:"email" doc:"Verified email when available"`
	DisplayName     *string    `json:"displayName,omitempty" maxLength:"120"`
	AvatarURL       *string    `json:"avatarUrl,omitempty" format:"uri"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty" format:"date-time"`
}

type CurrentUserOutput struct {
	Body CurrentUser
}

// Error is the stable application error projection.
type Error struct {
	Code    string `json:"code" example:"authentication_required"`
	Message string `json:"message" example:"Authentication is required."`
}

// RegisterContract establishes the A1 current-user contract. Authentication is
// enforced by the AuthMiddleware and RequireAuth operation middleware.
func RegisterContract(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "GetCurrentUser",
		Method:      http.MethodGet,
		Path:        "/api/v1/me",
		Summary:     "Get the authenticated user",
		Tags:        []string{"Authentication"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *struct{}) (*CurrentUserOutput, error) {
		u := Requester(ctx)
		if u == nil {
			return nil, huma.Error401Unauthorized("authentication required")
		}
		return &CurrentUserOutput{Body: currentUserFromAuth(u)}, nil
	})
}
