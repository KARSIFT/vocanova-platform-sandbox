package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/achievements"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dictionary"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/password"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/stories"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordlists"
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
	authRepo := auth.NewMemoryRepository()
	svc := auth.NewService(
		authRepo,
		&email.Fake{},
		auth.NewFakeOAuthProvider(&auth.OAuthIdentity{Subject: "openapi-sub", Email: "user@example.com", EmailVerified: true}),
		clock.Real{},
		auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 10),
		auth.Config{
			Environment:            "openapi",
			BaseURL:                "https://example.com",
			MagicLinkPath:          "/magic-link",
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
	RegisterPasswordAuth(contractAPI, password.NewService(nil, &email.Fake{}, clock.Real{}, auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 10), "https://example.com", "openapi", 30*24*time.Hour, func() bool { return true }, svc.PasswordSignupAllowed, svc.PasswordIdentityAllowed), svc)

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
	learningPreferencesSvc := users.NewLearningPreferencesService(usersRepo)
	RegisterOnboarding(contractAPI, usersSvc, svc)
	RegisterSettings(contractAPI, usersSvc, svc)
	RegisterLearningPreferences(contractAPI, learningPreferencesSvc, svc)
	SetOnboardingStatusLookup(newOnboardingStatusLookup(usersSvc))

	// VOC-031-T03: register the email-change routes on the
	// contract. The accounts service is constructed against the
	// same auth.MemoryRepository the contract is already using
	// and a fresh in-memory accounts repository; the limiter
	// and email sender are the same fixed-window / Fake
	// instances so the contract generation never panics on a
	// missing dependency.
	accountsRepo := accounts.NewMemoryRepository()
	accountsLimiter := auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 10)
	accountsIdem := accounts.NewMemoryIdempotencyStore()
	accountsSvc := accounts.NewService(accountsRepo, authRepo, &email.Fake{}, accountsIdem, clock.Real{}, accountsLimiter, accounts.Config{
		Environment: "openapi", BaseURL: "https://example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
		RateLimit: accounts.EmailChangeRateLimitConfig{
			RequestWindow: time.Minute, RequestLimit: 10,
			ConsumeWindow: time.Minute, ConsumeLimit: 10,
		},
	})
	RegisterEmailChangeLinks(contractAPI, accountsSvc, svc)

	// VOC-031-T04: register the account-deletion route on the
	// contract. The accounts service is the same one the
	// T03 email-change route uses; the in-memory idempotency
	// store is the same instance. The route is requester-
	// scoped, requires an authenticated session, double-
	// submit CSRF, and an Idempotency-Key header (DOC-07).
	RegisterAccountDeletionRequests(contractAPI, accountsSvc, svc)
	RegisterPersonalDataExports(contractAPI, accountsSvc, svc)

	// Register content routes for OpenAPI generation using empty in-memory repos.
	contentSvc := content.NewService(
		content.NewMemoryRepository(content.MemoryRepositoryData{}),
		content.NewMemorySavedStateReader(nil),
	)
	RegisterContent(contractAPI, contentSvc, usersSvc, learningPreferencesSvc)
	RegisterDictionary(contractAPI, dictionary.New())

	// Register learning routes for OpenAPI generation using an empty in-memory repo.
	learningSvc := learning.NewService(
		learning.NewMemoryRepository(learning.MemoryRepositoryData{}),
		learning.NewMemoryIdempotencyStore(),
		clock.Real{},
	)
	RegisterLearning(contractAPI, learningSvc, svc)
	RegisterLessons(contractAPI, lessons.NewService(lessons.NewPostgreSQLRepository(nil), clock.Real{}), svc)
	RegisterLessonRecommendation(contractAPI, lessons.NewRecommendationService(lessons.NewPostgreSQLRepository(nil)))
	RegisterWordKnowledge(contractAPI, wordknowledge.NewService(wordknowledge.NewPostgreSQLRepository(nil), clock.Real{}), svc)
	RegisterPractice(contractAPI, practice.NewService(practice.NewPostgreSQLRepository(nil), clock.Real{}), svc)
	RegisterWordLists(contractAPI, wordlists.NewService(wordlists.NewPostgreSQLRepository(nil), clock.Real{}), svc)
	RegisterStories(contractAPI, stories.NewService(stories.NewPostgreSQLRepository(nil), clock.Real{}), svc)
	RegisterAchievements(contractAPI, achievements.NewService(achievements.NewPostgreSQLRepository(nil)))

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
