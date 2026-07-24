package api

import (
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
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
	RegisterContract(contractAPI)

	// Register auth routes for OpenAPI generation using a placeholder service.
	svc := auth.NewService(
		auth.NewMemoryRepository(),
		&email.Fake{},
		clock.Real{},
		auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 10),
		auth.Config{
			Environment:       "openapi",
			BaseURL:           "https://example.com",
			MagicLinkPath:     "/auth/magic",
			SessionLifetime:   30 * 24 * time.Hour,
			MagicLinkLifetime: 15 * time.Minute,
			Cookie: auth.CookieConfig{
				Name:     "vocanova_session",
				CSRName:  "vocanova_csrf",
				Domain:   "",
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			},
		},
	)
	RegisterAuth(contractAPI, svc)
	return contractAPI
}
