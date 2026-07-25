package api

import (
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// NewContractAPI returns the OpenAPI contract with all registered routes. It
// uses an in-memory auth service so the contract can be generated without a
// database or secrets; the handlers are not exercised during generation.
func NewContractAPI() huma.API {
	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	config.Info.Description = "Explicit Vocanova HTTP DTO contract. Internal persistence models are not exposed."
	contractAPI := humachi.New(chi.NewMux(), config)
	contractAPI.UseMiddleware(withHumaContext)

	// Register auth routes for OpenAPI generation using a placeholder service.
	svc := auth.NewService(
		auth.NewMemoryRepository(),
		&email.Fake{},
		auth.NewFakeOAuthProvider(&auth.OAuthIdentity{Subject: "openapi-sub", Email: "user@example.com", EmailVerified: true}),
		clock.Real{},
		auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 10),
		auth.Config{
			Environment:            "openapi",
			BaseURL:                "https://example.com",
			MagicLinkPath:          "/auth/magic",
			OAuthRedirectURI:       "https://example.com/auth/oauth/google/callback",
			OAuthRedirectAllowlist: []string{"https://example.com/auth/oauth/google/callback"},
			SessionLifetime:        30 * 24 * time.Hour,
			MagicLinkLifetime:      15 * time.Minute,
			OAuthStateLifetime:     10 * time.Minute,
			Cookie: auth.CookieConfig{
				Name:           "vocanova_session",
				CSRName:        "vocanova_csrf",
				OAuthStateName: "vocanova_oauth_state",
				Domain:         "",
				Secure:         true,
				SameSite:       http.SameSiteStrictMode,
			},
		},
	)
	contractAPI.UseMiddleware(AuthMiddleware(svc))
	RegisterContract(contractAPI)
	RegisterAuth(contractAPI, svc)

	// Register content routes for OpenAPI generation using empty in-memory repos.
	contentSvc := content.NewService(
		content.NewMemoryRepository(content.MemoryRepositoryData{}),
		content.NewMemorySavedStateReader(nil),
	)
	RegisterContent(contractAPI, contentSvc)

	// Register learning routes for OpenAPI generation using an empty in-memory repo.
	learningSvc := learning.NewService(
		learning.NewMemoryRepository(learning.MemoryRepositoryData{}),
		learning.NewMemoryIdempotencyStore(),
		clock.Real{},
	)
	RegisterLearning(contractAPI, learningSvc, svc)

	// Register review routes for OpenAPI generation using an empty in-memory repo.
	reviewsSvc := reviews.NewService(reviews.NewMemoryRepository(reviews.MemoryRepositoryData{}))
	RegisterReviews(contractAPI, reviewsSvc)
	return contractAPI
}
