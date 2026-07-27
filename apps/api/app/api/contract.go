// Package api owns the public HTTP contract. Persistence entities must never be
// used as request or response bodies here.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// CurrentUser is the minimal public identity projection. It intentionally omits
// database IDs, provider subjects, tokens, and session metadata.
type CurrentUser struct {
	Email           *string    `json:"email,omitempty" format:"email" doc:"Verified email when available"`
	DisplayName     *string    `json:"displayName,omitempty" maxLength:"120"`
	AvatarURL       *string    `json:"avatarUrl,omitempty" format:"uri"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty" format:"date-time"`
	// OnboardingStatus is the additive field VOC-031-T01 introduces on
	// top of the A1 current-user contract. The Next.js middleware uses
	// it to gate routes on whether the learner has completed onboarding
	// (DOC-03 §3). Every other field above is unchanged, so consumers
	// that never read onboardingStatus are unaffected. The field is
	// always present in the JSON response; the default for a learner
	// who has not yet submitted onboarding is "not_started".
	OnboardingStatus string `json:"onboardingStatus" enum:"not_started,in_progress,completed" doc:"Onboarding status; not_started when the learner has not yet submitted onboarding"`
}

type CurrentUserOutput struct {
	Body CurrentUser
}

// Error is the stable application error projection.
type Error struct {
	Code    string `json:"code" example:"authentication_required"`
	Message string `json:"message" example:"Authentication is required."`
}

// OnboardingStatusLookup is the contract the API layer uses to enrich
// GET /api/v1/me with the additive onboardingStatus field. The users
// module implements it in production; tests use a closure over an
// in-memory fixture; NewContractAPI installs a stub that returns
// "not_started" so OpenAPI generation never panics on a missing
// implementation.
//
// The lookup must be safe to call concurrently and must return
// quickly — it is on the hot path of every authenticated page load
// (via the Next.js middleware).
type OnboardingStatusLookup func(ctx context.Context, userID uuid.UUID) (string, error)

// onboardingStatusLookup is the package-level service-locator hook.
// The default returns "not_started" so OpenAPI generation and
// RegisterContract work without a users service. NewContractAPI and
// tests replace it as needed.
var onboardingStatusLookup OnboardingStatusLookup = defaultOnboardingStatusLookup

// defaultOnboardingStatusLookup is the conservative default used by
// OpenAPI generation and by tests that don't care about the
// onboardingStatus value.
func defaultOnboardingStatusLookup(ctx context.Context, userID uuid.UUID) (string, error) {
	return "not_started", nil
}

// SetOnboardingStatusLookup installs the OnboardingStatusLookup the
// contract handler uses to enrich GET /api/v1/me. The production
// wiring is in NewContractAPI; tests may install a fixture-specific
// implementation directly.
func SetOnboardingStatusLookup(lookup OnboardingStatusLookup) {
	if lookup == nil {
		onboardingStatusLookup = defaultOnboardingStatusLookup
		return
	}
	onboardingStatusLookup = lookup
}

// RegisterContract establishes the current-user contract. Authentication
// is enforced by the AuthMiddleware and RequireAuth operation middleware.
// The handler enriches the A1 identity projection with the additive
// onboardingStatus field T01 introduces (DOC-03 §3, VOC-031-AC-02);
// every previously-existing field is unchanged, so consumers that
// never read onboardingStatus are unaffected.
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
		body := currentUserFromAuth(u)
		status, err := onboardingStatusLookup(ctx, u.ID)
		if err != nil || status == "" {
			status = "not_started"
		}
		body.OnboardingStatus = status
		return &CurrentUserOutput{Body: body}, nil
	})
}
