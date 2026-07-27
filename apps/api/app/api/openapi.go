package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	// VOC-031-T01: register the onboarding service so the
	// /api/v1/onboarding routes appear in the OpenAPI document. The
	// lookup the contract handler uses to enrich GET /api/v1/me with
	// the additive onboardingStatus field is installed at the same
	// time so the contract can run end-to-end during OpenAPI
	// generation without panicking on a missing implementation.
	// VOC-031-T02 wires the same MemoryRepository in as the
	// SettingsRepository so the new /api/v1/settings routes also
	// appear in the contract.
	usersRepo := users.NewMemoryRepository()
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, clock.Real{})
	RegisterOnboarding(contractAPI, usersSvc, svc)
	RegisterSettings(contractAPI, usersSvc, svc)
	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		profile, err := usersSvc.GetOnboarding(ctx, userID)
		if err != nil {
			// An unseen user has not yet submitted onboarding, so
			// the gate status is "not_started". This matches the
			// production semantics where the contract default
			// is "not_started" until CompleteOnboarding is called.
			return users.OnboardingStatusNotStarted, nil
		}
		return profile.Status, nil
	})

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
	reviewsSvc := reviews.NewService(
		reviews.NewMemoryRepository(reviews.MemoryRepositoryData{}),
		learning.NewMemoryIdempotencyStore(),
		clock.Real{},
	)
	RegisterReviews(contractAPI, reviewsSvc, svc)

	// Register AI feedback routes for OpenAPI generation using the mock provider.
	aifeedbackSvc := aifeedback.NewService(
		aifeedback.NewMemoryRepository(aifeedback.MemoryRepositoryData{}),
		aifeedback.NewMockProvider(),
		nil,
		nil,
		learning.NewMemoryIdempotencyStore(),
		nil,
		aifeedback.NewNoopTelemetryRecorder(),
		aifeedback.NewDefaultTaskBuilder(),
		aifeedback.NewDefaultOutputValidator(),
		clock.Real{},
		aifeedback.DefaultServiceConfig(),
	)
	RegisterAIFeedback(contractAPI, aifeedbackSvc, svc)

	// Register missions/progress read routes for OpenAPI generation. The
	// missions and gamification modules are *sql.DB-backed today; the
	// OpenAPI generator only needs the constructor to succeed (handlers
	// are not invoked during generation), so a nil *sql.DB is safe here.
	var openapiDB *sql.DB
	missionsRepo := missions.NewRepository(openapiDB)
	gamRepo := gamification.NewRepository(openapiDB)
	gamSvc := gamification.NewService(gamRepo)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)
	RegisterMissions(contractAPI, missionsSvc)
	return contractAPI
}
