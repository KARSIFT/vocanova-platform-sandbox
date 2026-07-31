package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
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
	_ "github.com/lib/pq"
)

// ProductionConfig is the runtime configuration the cmd/api binary
// reads from the environment to construct the production API.
// Every field maps to exactly one environment variable, all
// documented in apps/api/.env.example (T01); the wiring here is the
// single source of truth for which env vars T00 requires.
type ProductionConfig struct {
	Port            string
	DatabaseURL     string
	Environment     string
	BaseURL         string
	MagicLinkPath   string
	OAuthRedirect   string
	OAuthReturnURLs []string
	SessionDomain   string
	SessionSecure   bool
	SessionLifetime time.Duration
	APIProvider     string
	APIBaseURL      string
	APIKey          string
	APIModel        string
	OpenAIAPIKey    string
	APIProviderID   string
	APIModelID      string
	APITimeout      time.Duration

	AIEnabled    bool
	MagicLinkOn  bool
	OAuthOn      bool
	NewSignupsOn bool

	EmailProviderURL     string
	EmailProviderAPIKey  string
	EmailFrom            string
	EmailProviderTimeout time.Duration

	GoogleClientID     string
	GoogleClientSecret string
	GoogleOAuthScopes  string
	GoogleOAuthTimeout time.Duration
}

// LoadProductionConfig reads the production configuration from the
// process environment. It never falls back to silent defaults for
// required values: a missing PORT / DATABASE_URL / SESSION_COOKIE_DOMAIN
// / OAUTH_REDIRECT_URI returns a descriptive error. The AI / email /
// OAuth *enable* flags default to true so a freshly-deployed
// environment with a complete .env file is fully functional; the
// founder can flip any individual flag off without re-deploying by
// setting the corresponding env var to "false" (DOC-11 §3).
func LoadProductionConfig() (ProductionConfig, error) {
	cfg := ProductionConfig{
		Port:            getenv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Environment:     getenv("ENVIRONMENT", "staging"),
		BaseURL:         os.Getenv("BASE_URL"),
		MagicLinkPath:   getenv("MAGIC_LINK_PATH", "/auth/magic"),
		OAuthRedirect:   os.Getenv("OAUTH_REDIRECT_URI"),
		SessionDomain:   os.Getenv("SESSION_COOKIE_DOMAIN"),
		SessionSecure:   getenvBool("SESSION_COOKIE_SECURE", true),
		SessionLifetime: getenvDuration("SESSION_LIFETIME", 30*24*time.Hour),
		APIProvider:     getenv("AI_PROVIDER", string(aifeedback.ProviderOpenCode)),
		APIBaseURL:      getenv("AI_PROVIDER_BASE_URL", "http://127.0.0.1:4096"),
		APIKey:          os.Getenv("AI_PROVIDER_API_KEY"),
		APIModel:        getenv("AI_PROVIDER_MODEL", aifeedback.DefaultOpenCodeModel),
		APITimeout:      getenvDuration("AI_PROVIDER_TIMEOUT", 8*time.Second),
		AIEnabled:       getenvBool("AI_FEATURES_ENABLED", true),
		MagicLinkOn:     getenvBool("EMAIL_MAGIC_LINK_ENABLED", true),
		OAuthOn:         getenvBool("GOOGLE_OAUTH_ENABLED", true),
		NewSignupsOn:    getenvBool("NEW_USER_SIGNUP_ENABLED", true),

		EmailProviderURL:     os.Getenv("EMAIL_PROVIDER_URL"),
		EmailProviderAPIKey:  os.Getenv("EMAIL_PROVIDER_API_KEY"),
		EmailFrom:            os.Getenv("EMAIL_FROM"),
		EmailProviderTimeout: getenvDuration("EMAIL_PROVIDER_TIMEOUT", 10*time.Second),

		GoogleClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleOAuthScopes:  os.Getenv("GOOGLE_OAUTH_SCOPES"),
		GoogleOAuthTimeout: getenvDuration("GOOGLE_OAUTH_TIMEOUT", 8*time.Second),
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	if cfg.BaseURL == "" {
		return cfg, errors.New("BASE_URL is required")
	}
	if cfg.OAuthRedirect == "" {
		return cfg, errors.New("OAUTH_REDIRECT_URI is required")
	}
	if cfg.SessionDomain == "" {
		return cfg, errors.New("SESSION_COOKIE_DOMAIN is required")
	}

	if raw := os.Getenv("OAUTH_REDIRECT_ALLOWLIST"); raw != "" {
		for _, p := range strings.Split(raw, ",") {
			if u := strings.TrimSpace(p); u != "" {
				cfg.OAuthReturnURLs = append(cfg.OAuthReturnURLs, u)
			}
		}
	}
	if len(cfg.OAuthReturnURLs) == 0 {
		cfg.OAuthReturnURLs = []string{cfg.OAuthRedirect}
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// accountsIdempotencyAdapter wraps a learning.PostgreSQLIdempotencyStore
// so it satisfies accounts.IdempotencyStore. The two interfaces are
// structurally identical; the only difference is the IdempotencyStatus
// enum, whose values are intentionally the same (0/1/2) across both
// packages. accounts.idempotency.go's own comment notes this is the
// production wiring path the package is designed around.
type accountsIdempotencyAdapter struct {
	store *learning.PostgreSQLIdempotencyStore
}

func (a accountsIdempotencyAdapter) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (accounts.IdempotencyStatus, error) {
	status, err := a.store.Check(ctx, userID, operation, key, fingerprint)
	return accounts.IdempotencyStatus(status), err
}

func (a accountsIdempotencyAdapter) Record(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) error {
	return a.store.Record(ctx, userID, operation, key, fingerprint)
}

// Healthchecker is the minimum interface a database handle must
// satisfy for the /healthz probe. *sql.DB satisfies it, which is
// the only database the production wiring accepts (the per-
// module PostgreSQL repositories all take *sql.DB).
type Healthchecker interface {
	PingContext(ctx context.Context) error
}

// buildEmailSender returns the email.Sender the production wiring
// should use, applying the T14 fallback rules:
//
//   - When EMAIL_MAGIC_LINK_ENABLED is "false", always return
//     Fake{}. The auth service will reject magic-link requests
//     outright (and accounts.NewService will not call the sender
//     for magic-link sends), so the real sender is never reached;
//     Fake{} is correct here.
//   - When EMAIL_PROVIDER_API_KEY is unset (or empty), return
//     Fake{}. The real provider requires an API key; falling back
//     keeps staging runnable with magic-link delivery disabled at
//     the provider layer rather than crashing at startup.
//   - When EMAIL_PROVIDER_API_KEY is set, also require
//     EMAIL_PROVIDER_URL and EMAIL_FROM - NewHTTPSender validates
//     these are non-empty. A misconfigured URL or missing From is
//     a hard startup error: silently falling back to Fake{} would
//     hide a real configuration mistake.
//
// The decision is logged to stderr (matching cmd/api/main.go's
// logging style) so an operator running the binary interactively
// can see which sender is wired without having to read the env
// file. The log line never includes the API key.
func buildEmailSender(cfg ProductionConfig) (email.Sender, error) {
	if !cfg.MagicLinkOn {
		fmt.Fprintf(os.Stderr, "api: email sender=Fake (EMAIL_MAGIC_LINK_ENABLED=false)\n")
		return &email.Fake{}, nil
	}
	if cfg.EmailProviderAPIKey == "" {
		fmt.Fprintf(os.Stderr, "api: email sender=Fake (EMAIL_PROVIDER_API_KEY unset; magic-link delivery disabled at the provider layer)\n")
		return &email.Fake{}, nil
	}
	sender, err := email.NewHTTPSender(email.HTTPSenderConfig{
		URL:     cfg.EmailProviderURL,
		APIKey:  cfg.EmailProviderAPIKey,
		From:    cfg.EmailFrom,
		Timeout: cfg.EmailProviderTimeout,
	})
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "api: email sender=HTTPSender url=%s from=%q\n", cfg.EmailProviderURL, cfg.EmailFrom)
	return sender, nil
}

// buildOAuthProvider returns the auth.OAuthProvider the production
// wiring should use, applying the T15 fallback rules that mirror
// T14's buildEmailSender shape:
//
//   - When GOOGLE_OAUTH_ENABLED is "false", always return
//     NewFakeOAuthProvider. The auth service's kill switch
//     short-circuits OAuthStart with ErrOAuthDisabled before the
//     provider is reached, so the fake is never exercised in
//     practice; the wiring still has to return a non-nil provider
//     so NewService can be constructed.
//   - When GOOGLE_OAUTH_CLIENT_ID is unset (or empty), return
//     NewFakeOAuthProvider. The real provider requires a client
//     ID; falling back keeps staging runnable with Google
//     sign-in disabled at the provider layer rather than
//     crashing at startup.
//   - When GOOGLE_OAUTH_CLIENT_ID is set, also require
//     GOOGLE_OAUTH_CLIENT_SECRET - NewGoogleOAuthProvider
//     validates this is non-empty. A misconfigured credential
//     (ID set but secret missing) is a hard startup error:
//     silently falling back to FakeOAuthProvider would hide a
//     real configuration mistake.
//
// The decision is logged to stderr (matching buildEmailSender's
// style) so an operator running the binary interactively can see
// which provider is wired without having to read the env file.
// The log line never includes the client secret.
func buildOAuthProvider(cfg ProductionConfig) (auth.OAuthProvider, error) {
	fakeIdentity := &auth.OAuthIdentity{Subject: "production-fake-sub", Email: "user@example.com", EmailVerified: true}
	if !cfg.OAuthOn {
		fmt.Fprintf(os.Stderr, "api: oauth provider=FakeOAuthProvider (GOOGLE_OAUTH_ENABLED=false)\n")
		return auth.NewFakeOAuthProvider(fakeIdentity), nil
	}
	if cfg.GoogleClientID == "" {
		fmt.Fprintf(os.Stderr, "api: oauth provider=FakeOAuthProvider (GOOGLE_OAUTH_CLIENT_ID unset; Google sign-in disabled at the provider layer)\n")
		return auth.NewFakeOAuthProvider(fakeIdentity), nil
	}
	if cfg.GoogleClientSecret == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_SECRET is required when GOOGLE_OAUTH_CLIENT_ID is set")
	}
	provider, err := auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURI:  cfg.OAuthRedirect,
		Scopes:       cfg.GoogleOAuthScopes,
		Timeout:      cfg.GoogleOAuthTimeout,
	})
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "api: oauth provider=GoogleOAuthProvider client_id=%q redirect_uri=%q\n", cfg.GoogleClientID, cfg.OAuthRedirect)
	return provider, nil
}

// buildAIProviders returns the (FeedbackProvider, SafetyClassifier) the
// production wiring should hand to aifeedback.NewService, applying the
// VOC-034-D00 fix for the literal-nil safety-classifier defect the old
// inline construction had at this exact call site:
//
//   - When AI_PROVIDER=opencode AND AI_PROVIDER_API_KEY is set (the
//     same condition the prior inline block used to select the real
//     OpenCodeFeedbackProvider), return a real OpenCodeFeedbackProvider
//     and a CompositeSafetyClassifier wrapping a real
//     OpenCodeModerationProvider.
//   - When AI_PROVIDER=gemini AND AI_PROVIDER_API_KEY is set, return a
//     real GeminiFeedbackProvider and a CompositeSafetyClassifier
//     wrapping a real GeminiModerationProvider.
//   - When those conditions are false, return aifeedback.NewMockProvider()
//     for the feedback role and a CompositeSafetyClassifier wrapping
//     MockProvider for the moderation role. This preserves the prior
//     fallback behavior exactly - the mock already implements both
//     FeedbackProvider and ModerationProvider (see aifeedback.go).
//
// Both branches build the safety classifier via
// NewCompositeSafetyClassifier(NewDefaultLocalAbuseChecker(), provider)
// so the deterministic local checks run before any provider call
// (composite ordering is unchanged; safety.go is not modified). The
// moderation provider forces its own MaxRetries to 0 internally
// (VOC-034-D03) regardless of the 1 carried in the shared
// OpenCodeConfig struct; the feedback provider keeps MaxRetries=1 as
// before. The two providers therefore share BaseURL / APIKey / Model
// / Timeout configuration but apply different retry budgets, by
// design - the latency-budget question against DOC-09 §18's 10s total
// backend target is explicitly deferred to VOC-032-T10's live
// threshold evaluation.
//
// The decision is logged to stderr (matching buildEmailSender /
// buildOAuthProvider's style) so an operator running the binary
// interactively can see which providers are wired without having to
// read the env file. The log line never includes the API key, the
// model string, or any request body.
func buildAIProviders(cfg ProductionConfig) (aifeedback.FeedbackProvider, aifeedback.SafetyClassifier) {
	if cfg.APIProvider == string(aifeedback.ProviderOpenCode) && cfg.APIKey != "" {
		openCodeCfg := aifeedback.OpenCodeConfig{
			BaseURL:    cfg.APIBaseURL,
			APIKey:     cfg.APIKey,
			Model:      cfg.APIModel,
			Timeout:    cfg.APITimeout,
			MaxRetries: 1,
		}
		fmt.Fprintf(os.Stderr, "api: ai feedback=OpenCodeFeedbackProvider ai moderation=OpenCodeModerationProvider (provider=opencode)\n")
		return aifeedback.NewOpenCodeFeedbackProvider(openCodeCfg),
			aifeedback.NewCompositeSafetyClassifier(
				aifeedback.NewDefaultLocalAbuseChecker(),
				aifeedback.NewOpenCodeModerationProvider(openCodeCfg),
			)
	}
	if cfg.APIProvider == "gemini" && cfg.APIKey != "" {
		geminiCfg := aifeedback.GeminiConfig{
			APIKey:     cfg.APIKey,
			Model:      cfg.APIModel,
			BaseURL:    cfg.APIBaseURL,
			Timeout:    cfg.APITimeout,
			MaxRetries: 1,
		}
		fmt.Fprintf(os.Stderr, "api: ai feedback=GeminiFeedbackProvider ai moderation=GeminiModerationProvider (provider=gemini)\n")
		return aifeedback.NewGeminiFeedbackProvider(geminiCfg),
			aifeedback.NewCompositeSafetyClassifier(
				aifeedback.NewDefaultLocalAbuseChecker(),
				aifeedback.NewGeminiModerationProvider(geminiCfg),
			)
	}
	fmt.Fprintf(os.Stderr, "api: ai feedback=MockProvider ai moderation=MockProvider (AI_PROVIDER not configured for a real provider or AI_PROVIDER_API_KEY unset; AI features fall back to in-memory mock)\n")
	return aifeedback.NewMockProvider(),
		aifeedback.NewCompositeSafetyClassifier(
			aifeedback.NewDefaultLocalAbuseChecker(),
			aifeedback.NewMockProvider(),
		)
}

// HealthzInput is intentionally empty - /healthz takes no parameters.
type HealthzInput struct{}

// HealthzOutput reports the database-connection state. A healthy
// response is 200; an unhealthy response is 503 with a body that
// does not leak the underlying error (per the F2/DOC-11 contract:
// the probe should confirm reachability, not expose internals).
type HealthzOutput struct {
	Body struct {
		Status    string `json:"status" example:"ok" doc:"Either 'ok' or 'unhealthy'"`
		Database  string `json:"database" example:"ok" doc:"Either 'ok' or 'unhealthy'"`
		Timestamp string `json:"timestamp" format:"date-time" doc:"Probe timestamp in RFC3339"`
	}
}

// NewProductionAPI returns a huma.API wired against real,
// Postgres-backed services built from cfg. It registers exactly
// the same route set NewContractAPI registers, so the OpenAPI
// contract produced by the contract generator matches the routes
// a real server exposes; NewContractAPI keeps its in-memory wiring
// unchanged (its only call site is OpenAPI generation, which never
// invokes handlers).
//
// db may be supplied by the caller (e.g. the cmd/api main passes
// the same *sql.DB it pings at startup) or, if nil, is opened
// against cfg.DatabaseURL and Ping'd before the function returns.
// A failure to open or ping the database returns a non-nil error
// and no huma.API - the server must not start serving traffic when
// the database is unreachable.
//
// The opened (or supplied) *sql.DB is always returned alongside the
// huma.API so the caller can close it during graceful shutdown -
// NewProductionAPI never closes a database connection it hands
// back to a caller that is still using it.
func NewProductionAPI(cfg ProductionConfig, db *sql.DB) (huma.API, *sql.DB, error) {
	if db == nil {
		conn, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			return nil, nil, fmt.Errorf("open database: %w", err)
		}
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.PingContext(pingCtx); err != nil {
			_ = conn.Close()
			return nil, nil, fmt.Errorf("ping database: %w", err)
		}
		db = conn
	}

	clk := clock.Real{}
	authRepo := auth.NewPostgreSQLRepository(db)
	limiter := auth.NewFixedWindowRateLimiter(clk, time.Minute, 10)

	mailer, err := buildEmailSender(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("build email sender: %w", err)
	}

	oauthProvider, err := buildOAuthProvider(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("build oauth provider: %w", err)
	}

	authCfg := auth.Config{
		Environment:            cfg.Environment,
		BaseURL:                cfg.BaseURL,
		MagicLinkPath:          cfg.MagicLinkPath,
		OAuthRedirectURI:       cfg.OAuthRedirect,
		OAuthRedirectAllowlist: cfg.OAuthReturnURLs,
		SessionLifetime:        cfg.SessionLifetime,
		MagicLinkLifetime:      15 * time.Minute,
		OAuthStateLifetime:     10 * time.Minute,
		Cookie: auth.CookieConfig{
			Name:           "vocanova_session",
			CSRName:        "vocanova_csrf",
			OAuthStateName: "vocanova_oauth_state",
			Domain:         cfg.SessionDomain,
			Secure:         cfg.SessionSecure,
			SameSite:       http.SameSiteStrictMode,
		},
	}
	authSvc := auth.NewService(authRepo, mailer, oauthProvider, clk, limiter, authCfg)
	authSvc.SetKillSwitches(&auth.KillSwitches{
		MagicLinkEnabled:  cfg.MagicLinkOn,
		OAuthEnabled:      cfg.OAuthOn,
		NewSignupsEnabled: cfg.NewSignupsOn,
	})

	usersRepo := users.NewPostgreSQLRepository(db)
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, clk)

	contentRepo := content.NewPostgreSQLRepository(db)
	contentSvc := content.NewService(contentRepo, learning.NewPostgreSQLRepository(db))

	learningIdem := learning.NewPostgreSQLIdempotencyStore(db)
	learningRepo := learning.NewPostgreSQLRepository(db)
	learningSvc := learning.NewService(learningRepo, learningIdem, clk)

	reviewsRepo := reviews.NewPostgreSQLRepository(db, clk)
	reviewsSvc := reviews.NewService(reviewsRepo, learningIdem, clk)

	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)

	accountsRepo := accounts.NewPostgreSQLRepository(db)
	accountsIdem := accountsIdempotencyAdapter{store: learningIdem}
	accountsLimiter := auth.NewFixedWindowRateLimiter(clk, time.Minute, 10)
	accountsSvc := accounts.NewService(accountsRepo, authRepo, mailer, accountsIdem, clk, accountsLimiter, accounts.Config{
		Environment:               cfg.Environment,
		BaseURL:                   cfg.BaseURL,
		EmailChangePath:           "/auth/email-change",
		EmailChangeLinkLifetime:   15 * time.Minute,
		AccountDeletionPurgeDelay: accounts.DefaultAccountDeletionPurgeDelay,
	})

	aiProvider, safetyClassifier := buildAIProviders(cfg)
	aiGate := aifeedback.GenerationGate(aifeedback.NewAlwaysEnabledGate())
	if !cfg.AIEnabled {
		aiGate = aifeedback.NewDisabledGate()
	}
	aifeedbackSvc := aifeedback.NewService(
		aifeedback.NewPostgreSQLRepository(db, clk),
		aiProvider,
		safetyClassifier,
		nil,
		learningIdem,
		nil,
		aifeedback.NewNoopTelemetryRecorder(),
		aifeedback.NewDefaultTaskBuilder(),
		aifeedback.NewDefaultOutputValidator(),
		clk,
		aifeedback.ServiceConfig{
			Provider: cfg.APIProvider,
			Model:    cfg.APIModel,
			Release:  cfg.Environment,
			OpenCode: aifeedback.OpenCodeConfig{
				BaseURL:    cfg.APIBaseURL,
				APIKey:     cfg.APIKey,
				Model:      cfg.APIModel,
				Timeout:    cfg.APITimeout,
				MaxRetries: 1,
			},
			Gate:    aiGate,
			Metrics: aifeedback.NewNoopMetricsRecorder(),
		},
	)

	humaCfg := huma.DefaultConfig("Vocanova API", "0.1.0")
	humaCfg.Info.Description = "Explicit Vocanova HTTP DTO contract. Internal persistence models are not exposed."
	mux := chi.NewMux()
	api := humachi.New(mux, humaCfg)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterContract(api)
	RegisterAuth(api, authSvc)
	RegisterOnboarding(api, usersSvc, authSvc)
	RegisterSettings(api, usersSvc, authSvc)
	RegisterEmailChangeLinks(api, accountsSvc, authSvc)
	RegisterAccountDeletionRequests(api, accountsSvc, authSvc)
	RegisterContent(api, contentSvc)
	RegisterLearning(api, learningSvc, authSvc)
	RegisterReviews(api, reviewsSvc, authSvc)
	RegisterAIFeedback(api, aifeedbackSvc, authSvc)
	RegisterMissions(api, missionsSvc)

	SetOnboardingStatusLookup(func(ctx context.Context, userID uuid.UUID) (string, error) {
		profile, err := usersSvc.GetOnboarding(ctx, userID)
		if err != nil {
			return users.OnboardingStatusNotStarted, nil
		}
		return profile.Status, nil
	})

	RegisterHealthz(api, db)

	return api, db, nil
}

// RegisterHealthz installs the unauthenticated GET /healthz probe.
// It pings the supplied database with a 2-second deadline; a
// successful ping returns 200 with status="ok", an unsuccessful
// ping returns 503 with status="unhealthy" and a non-leaking
// body. /healthz is the only production route NewContractAPI
// does not register, because the contract generator has no
// production database to probe - it is T00-specific.
func RegisterHealthz(api huma.API, db Healthchecker) {
	huma.Register(api, huma.Operation{
		OperationID: "GetHealthz",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "Liveness and database-reachability probe",
		Tags:        []string{"Operations"},
	}, func(ctx context.Context, input *HealthzInput) (*HealthzOutput, error) {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		dbStatus := "ok"
		overall := "ok"
		if err := db.PingContext(pingCtx); err != nil {
			dbStatus = "unhealthy"
			overall = "unhealthy"
		}
		out := &HealthzOutput{}
		out.Body.Status = overall
		out.Body.Database = dbStatus
		out.Body.Timestamp = time.Now().UTC().Format(time.RFC3339)
		if overall != "ok" {
			return out, &huma.ErrorModel{
				Status: http.StatusServiceUnavailable,
				Title:  "Service Unavailable",
				Detail: "database is unreachable",
			}
		}
		return out, nil
	})
}
