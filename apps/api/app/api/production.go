package api

import (
	"context"
	"crypto/subtle"
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
	"github.com/getsentry/sentry-go"
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
	// SignupAllowlist holds normalized emails admitted to sign up
	// even while NewSignupsOn is false (VOC-038-D00/D01). Built from
	// NEW_USER_SIGNUP_ALLOWLIST, a comma-separated list, the same
	// way OAuthReturnURLs is built from OAUTH_REDIRECT_ALLOWLIST.
	SignupAllowlist map[string]struct{}
	SessionDomain   string
	SessionSecure   bool
	SessionLifetime time.Duration
	APIProvider     string
	APIBaseURL      string
	APIKey          string
	APIAccountID    string
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

	SentryDSN           string
	SentryEnvironment   string
	SentryRelease       string
	MonitoringTestToken string
	SmokeTestMintToken  string

	// SyntheticSmokeTestEmail is the reserved identity of the
	// deploy-seeded synthetic smoke-test account (VOC-050-T00). It is
	// normalized at load time because every comparison against it
	// happens on already-normalized addresses.
	SyntheticSmokeTestEmail string
}

// AI-provider identifiers and per-provider connection defaults.
// providerGemini has no aifeedback-package constant yet (adding one
// would touch aifeedback.go, outside VOC-035-T01's declared file set),
// so the selector value is named here instead of repeated as a literal.
const (
	providerGemini     = "gemini"
	providerCloudflare = "cloudflare"

	// defaultOpenCodeBaseURL is the loopback `opencode serve` host an
	// operator-hosted OpenCode deployment runs on.
	defaultOpenCodeBaseURL = "http://127.0.0.1:4096"

	// defaultGeminiModel mirrors aifeedback's own internal Gemini model
	// default. It is repeated here (rather than read from aifeedback)
	// so ServiceConfig telemetry records the model actually used instead
	// of an empty string when AI_PROVIDER_MODEL is unset.
	defaultGeminiModel = "gemini-2.5-flash"

	// defaultCloudflareModel mirrors aifeedback's internal Cloudflare
	// model default so telemetry still records a concrete value when
	// AI_PROVIDER_MODEL is unset.
	defaultCloudflareModel = "@cf/meta/llama-3.3-70b-instruct-fp8-fast"

	// defaultSyntheticSmokeTestEmail matches the default
	// apps/api/scripts/seed-synthetic-smoke-user.sh seeds, so the
	// address the deploy creates and the address the API refuses on
	// every real sign-in path stay the same when neither side is
	// configured explicitly. The `.invalid` TLD is reserved by
	// RFC 2606 and can never receive mail.
	defaultSyntheticSmokeTestEmail = "smoke-test-bot@synthetic.vocanova.invalid"
)

// aiProviderBaseURL resolves AI_PROVIDER_BASE_URL for the selected
// provider. OpenCode has no usable endpoint without one, so it keeps
// its loopback default. Gemini and Cloudflare both have fixed provider
// hosts in their own config defaults, so an unset variable must stay
// empty here instead of inheriting OpenCode's loopback default.
// An explicitly-set value is still honored for any provider.
func aiProviderBaseURL(provider string) string {
	if provider == providerGemini || provider == providerCloudflare {
		return os.Getenv("AI_PROVIDER_BASE_URL")
	}
	return getenv("AI_PROVIDER_BASE_URL", defaultOpenCodeBaseURL)
}

// aiProviderModel resolves AI_PROVIDER_MODEL for the selected provider
// (VOC-035-D02): the model identifier namespaces differ per provider,
// so an unset variable defaults to that provider's own model rather
// than to OpenCode's for every provider.
func aiProviderModel(provider string) string {
	if provider == providerGemini {
		return getenv("AI_PROVIDER_MODEL", defaultGeminiModel)
	}
	if provider == providerCloudflare {
		return getenv("AI_PROVIDER_MODEL", defaultCloudflareModel)
	}
	return getenv("AI_PROVIDER_MODEL", aifeedback.DefaultOpenCodeModel)
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
	aiProvider := getenv("AI_PROVIDER", string(aifeedback.ProviderOpenCode))
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
		APIProvider:     aiProvider,
		APIBaseURL:      aiProviderBaseURL(aiProvider),
		APIKey:          os.Getenv("AI_PROVIDER_API_KEY"),
		APIAccountID:    os.Getenv("AI_PROVIDER_ACCOUNT_ID"),
		APIModel:        aiProviderModel(aiProvider),
		APITimeout:      getenvDuration("AI_PROVIDER_TIMEOUT", 8*time.Second),
		AIEnabled:       getenvBool("AI_FEATURES_ENABLED", true),
		MagicLinkOn:     getenvBool("EMAIL_MAGIC_LINK_ENABLED", true),
		OAuthOn:         getenvBool("GOOGLE_OAUTH_ENABLED", true),
		NewSignupsOn:    getenvBool("NEW_USER_SIGNUP_ENABLED", true),

		EmailProviderURL:     os.Getenv("EMAIL_PROVIDER_URL"),
		EmailProviderAPIKey:  os.Getenv("EMAIL_PROVIDER_API_KEY"),
		EmailFrom:            os.Getenv("EMAIL_FROM"),
		EmailProviderTimeout: getenvDuration("EMAIL_PROVIDER_TIMEOUT", 10*time.Second),

		GoogleClientID:      os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleClientSecret:  os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleOAuthScopes:   os.Getenv("GOOGLE_OAUTH_SCOPES"),
		GoogleOAuthTimeout:  getenvDuration("GOOGLE_OAUTH_TIMEOUT", 8*time.Second),
		SentryDSN:           os.Getenv("SENTRY_DSN"),
		SentryEnvironment:   getenv("SENTRY_ENVIRONMENT", getenv("ENVIRONMENT", "staging")),
		SentryRelease:       os.Getenv("SENTRY_RELEASE"),
		MonitoringTestToken: os.Getenv("MONITORING_TEST_TOKEN"),
		SmokeTestMintToken:  os.Getenv("SMOKE_TEST_SESSION_MINT_TOKEN"),
		SyntheticSmokeTestEmail: auth.NormalizeEmail(
			getenv("VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL", defaultSyntheticSmokeTestEmail),
		),
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
	if raw := os.Getenv("NEW_USER_SIGNUP_ALLOWLIST"); raw != "" {
		cfg.SignupAllowlist = make(map[string]struct{})
		for _, e := range strings.Split(raw, ",") {
			if e = auth.NormalizeEmail(e); e != "" {
				cfg.SignupAllowlist[e] = struct{}{}
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
//   - When AI_PROVIDER=cloudflare AND both AI_PROVIDER_API_KEY and
//     AI_PROVIDER_ACCOUNT_ID are set, return a real
//     CloudflareFeedbackProvider and a CompositeSafetyClassifier
//     wrapping a real CloudflareModerationProvider.
//   - When those conditions are false, return aifeedback.NewMockProvider()
//     for the feedback role and a CompositeSafetyClassifier wrapping
//     MockProvider for the moderation role. This preserves the prior
//     fallback behavior exactly - the mock already implements both
//     FeedbackProvider and ModerationProvider (see aifeedback.go).
//
// Each real branch builds one shared per-provider config struct
// (OpenCodeConfig or GeminiConfig) that both roles read, so no
// configuration can mix one provider's feedback adapter with another
// provider's moderation adapter. cfg.APIProvider's literal value is
// the sole selector, so exactly one branch is ever active
// (VOC-035-D00). The Gemini branch passes cfg.APIBaseURL through
// as-is: LoadProductionConfig leaves it empty unless
// AI_PROVIDER_BASE_URL is explicitly set when AI_PROVIDER=gemini, so
// GeminiConfig applies Google's own endpoint default, while an
// operator who deliberately sets the variable keeps their override
// (VOC-035-D02).
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
	if cfg.APIProvider == providerGemini && cfg.APIKey != "" {
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
	if cfg.APIProvider == providerCloudflare && cfg.APIKey != "" && cfg.APIAccountID != "" {
		cloudflareCfg := aifeedback.CloudflareConfig{
			APIToken:   cfg.APIKey,
			AccountID:  cfg.APIAccountID,
			Model:      cfg.APIModel,
			BaseURL:    cfg.APIBaseURL,
			Timeout:    cfg.APITimeout,
			MaxRetries: 1,
		}
		fmt.Fprintf(os.Stderr, "api: ai feedback=CloudflareFeedbackProvider ai moderation=CloudflareModerationProvider (provider=cloudflare)\n")
		return aifeedback.NewCloudflareFeedbackProvider(cloudflareCfg),
			aifeedback.NewCompositeSafetyClassifier(
				aifeedback.NewDefaultLocalAbuseChecker(),
				aifeedback.NewCloudflareModerationProvider(cloudflareCfg),
			)
	}
	fmt.Fprintf(os.Stderr, "api: ai feedback=MockProvider ai moderation=MockProvider (AI_PROVIDER not configured for a complete real-provider config; AI features fall back to in-memory mock)\n")
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
		Status                string           `json:"status" example:"ok" doc:"Either 'ok' or 'unhealthy'"`
		Database              string           `json:"database" example:"ok" doc:"Either 'ok' or 'unhealthy'"`
		Timestamp             string           `json:"timestamp" format:"date-time" doc:"Probe timestamp in RFC3339"`
		KillSwitches          KillSwitchStatus `json:"kill_switches" doc:"Current DOC-11 kill-switch state - not secret, and needed by VOC-038-T02's smoke-test suite to assert the deploy wrote the intended posture"`
		ControlledSignupReady bool             `json:"controlled_signup_ready" example:"false" doc:"True when OAuth is enabled, global signup is disabled, and at least one controlled first-time signup is permitted; never exposes allowlist emails or cohort size"`
	}
}

// KillSwitchStatus reports the current state of the DOC-11 §3 kill
// switches, exactly as LoadProductionConfig resolved them from the
// environment. Reported on /healthz (unauthenticated) rather than
// only internally: these flags are already effectively observable
// from the outside by probing behavior (e.g. attempting a magic-link
// request), so surfacing them directly lets VOC-038-T02's smoke-test
// suite assert the intended posture without any state-mutating probe.
type KillSwitchStatus struct {
	MagicLinkEnabled  bool `json:"magic_link_enabled" example:"false"`
	OAuthEnabled      bool `json:"oauth_enabled" example:"false"`
	NewSignupsEnabled bool `json:"new_signups_enabled" example:"false"`
	AIEnabled         bool `json:"ai_enabled" example:"true"`
}

type MonitoringSentryTestInput struct {
	Authorization string `header:"Authorization" required:"true"`
}

type MonitoringSentryTestOutput struct {
	Body struct {
		Status    string `json:"status" example:"accepted"`
		EventID   string `json:"event_id" example:"0123456789abcdef0123456789abcdef"`
		Timestamp string `json:"timestamp" format:"date-time"`
	}
}

type SyntheticSmokeTestSessionMintInput struct {
	Authorization string `header:"Authorization" required:"true"`
}

type SyntheticSmokeTestSessionMintOutput struct {
	Body struct {
		Status        string `json:"status" example:"issued"`
		SessionCookie string `json:"session_cookie" doc:"Opaque vocanova_session value for the synthetic smoke-test account"`
		CSRFToken     string `json:"csrf_token" doc:"Double-submit CSRF token paired with the vocanova_csrf cookie"`
		ExpiresAt     string `json:"expires_at" format:"date-time"`
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
			// Intentionally NOT cfg.SessionDomain: the OAuth state cookie
			// only round-trips between this API host's own start/callback
			// endpoints, so it must be host-only (empty Domain), not
			// scoped to the web app's hostname. A production/staging
			// sibling-subdomain topology (api-X vs X) means a cookie
			// scoped to SessionDomain is never sent back to this host by
			// a real browser - found live via VOC-037-T03's rehearsal.
			OAuthStateDomain: "",
			Secure:           cfg.SessionSecure,
			// Lax, not Strict - matches OAuthStateCookie's own hardcoded
			// SameSiteLaxMode in tokens.go, and for the identical reason:
			// this session cookie is minted as the LAST hop of a redirect
			// chain that began at Google (accounts.google.com -> this
			// API's /callback -> the web app's /home), so the whole chain
			// counts as cross-site-initiated for SameSite purposes even
			// though every vocanova.site hop is same-site with each other.
			// Strict cookies are never attached anywhere in a navigation
			// chain that started cross-site, so the browser's request to
			// /home carried no session cookie at all - found live via
			// VOC-038-T03's first real founder sign-in attempt: OAuth
			// completed successfully (302 with Set-Cookie, confirmed in
			// nginx access logs), but the very next request landed back on
			// /signin?returnTo=/home because middleware.ts's own /me check
			// saw no cookie and treated the visitor as unauthenticated.
			SameSite: http.SameSiteLaxMode,
		},
	}
	authSvc := auth.NewService(authRepo, mailer, oauthProvider, clk, limiter, authCfg)
	authSvc.SetKillSwitches(&auth.KillSwitches{
		MagicLinkEnabled:       cfg.MagicLinkOn,
		OAuthEnabled:           cfg.OAuthOn,
		NewSignupsEnabled:      cfg.NewSignupsOn,
		SignupAllowlist:        cfg.SignupAllowlist,
		ReservedSyntheticEmail: cfg.SyntheticSmokeTestEmail,
	})

	usersRepo := users.NewPostgreSQLRepository(db)
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, clk)

	contentRepo := content.NewPostgreSQLRepository(db)
	contentSvc := content.NewService(contentRepo, learning.NewPostgreSQLRepository(db))

	learningIdem := learning.NewPostgreSQLIdempotencyStore(db)
	learningRepo := learning.NewPostgreSQLRepository(db)
	learningSvc := learning.NewService(learningRepo, learningIdem, clk)

	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)

	reviewsRepo := newProductionReviewsRepository(db, clk, gamSvc, missionsSvc)
	reviewsSvc := reviews.NewService(reviewsRepo, learningIdem, clk)

	accountsRepo := accounts.NewPostgreSQLRepository(db)
	accountsIdem := accountsIdempotencyAdapter{store: learningIdem}
	accountsLimiter := auth.NewFixedWindowRateLimiter(clk, time.Minute, 10)
	accountsSvc := accounts.NewService(accountsRepo, authRepo, mailer, accountsIdem, clk, accountsLimiter, accounts.Config{
		Environment:               cfg.Environment,
		BaseURL:                   cfg.BaseURL,
		EmailChangePath:           "/auth/email-change",
		EmailChangeLinkLifetime:   15 * time.Minute,
		AccountDeletionPurgeDelay: accounts.DefaultAccountDeletionPurgeDelay,
		ReservedSyntheticEmail:    cfg.SyntheticSmokeTestEmail,
	})

	aiProvider, safetyClassifier := buildAIProviders(cfg)
	aiGate := aifeedback.GenerationGate(aifeedback.NewAlwaysEnabledGate())
	if !cfg.AIEnabled {
		aiGate = aifeedback.NewDisabledGate()
	}
	// VOC-028-D01 / issue #1177: wire the real missions.MissionUpdater (not
	// the P3 StubMissionUpdater) so a qualifying sentence-feedback call
	// actually awards Confidence Points and updates the daily mission /
	// streak state, mirroring the same fix already applied to the reviews
	// P4 wiring above (see newProductionReviewsRepository / VOC-065-T01).
	aifeedbackSvc := aifeedback.NewService(
		aifeedback.NewPostgreSQLRepository(db, clk),
		aiProvider,
		safetyClassifier,
		nil,
		learningIdem,
		missions.NewMissionUpdater(missionsSvc, gamSvc),
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
	corsOrigins := corsAllowedOrigins(cfg.OAuthReturnURLs)
	mux.Use(corsMiddleware(corsOrigins))
	fmt.Fprintf(os.Stderr, "api: cors allowed origins: %s\n", corsOriginsSummary(corsOrigins))
	api := humachi.New(mux, humaCfg)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterContract(api)
	RegisterAuth(api, authSvc)
	RegisterOnboarding(api, usersSvc, authSvc)
	RegisterSettings(api, usersSvc, authSvc)
	RegisterEmailChangeLinks(api, accountsSvc, authSvc)
	RegisterAccountDeletionRequests(api, accountsSvc, authSvc)
	RegisterContent(api, contentSvc, usersSvc)
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

	RegisterHealthz(api, db, KillSwitchStatus{
		MagicLinkEnabled:  cfg.MagicLinkOn,
		OAuthEnabled:      cfg.OAuthOn,
		NewSignupsEnabled: cfg.NewSignupsOn,
		AIEnabled:         cfg.AIEnabled,
	}, ControlledSignupReady(
		cfg.OAuthOn,
		cfg.NewSignupsOn,
		cfg.SignupAllowlist,
		cfg.SyntheticSmokeTestEmail,
	))
	RegisterMonitoringSentryTest(api, cfg.MonitoringTestToken, cfg.Environment)
	RegisterSyntheticSmokeTestSessionMint(api, authSvc, cfg.SmokeTestMintToken)

	return api, db, nil
}

// newProductionReviewsRepository is the sole construction path for the live
// reviews PostgreSQL repository. It wires P4 gamification and missions
// dependencies so SubmitReview increments daily_mission_snapshots.reviews_completed.
// NewProductionAPI is the only caller; production_test.go asserts this wiring.
func newProductionReviewsRepository(db *sql.DB, clk clock.Clock, gamSvc *gamification.Service, missionsSvc *missions.Service) *reviews.PostgreSQLRepository {
	return reviews.NewPostgreSQLRepository(db, clk,
		reviews.WithGamificationService(gamSvc),
		reviews.WithMissionsService(missionsSvc),
	)
}

// ControlledSignupReady reports whether controlled first-time signup
// is operational: OAuth enabled, global signup disabled, and at
// least one non-reserved allowlisted identity may sign up. It never
// inspects or exposes cohort size beyond this boolean.
func ControlledSignupReady(oauthEnabled, newSignupsEnabled bool, allowlist map[string]struct{}, reservedSyntheticEmail string) bool {
	if !oauthEnabled || newSignupsEnabled {
		return false
	}
	for email := range allowlist {
		if email == "" {
			continue
		}
		if reservedSyntheticEmail != "" && email == reservedSyntheticEmail {
			continue
		}
		return true
	}
	return false
}

// RegisterHealthz installs the unauthenticated GET /healthz probe.
// It pings the supplied database with a 2-second deadline; a
// successful ping returns 200 with status="ok", an unsuccessful
// ping returns 503 with status="unhealthy" and a non-leaking
// body. /healthz is the only production route NewContractAPI
// does not register, because the contract generator has no
// production database to probe - it is T00-specific.
func RegisterHealthz(api huma.API, db Healthchecker, switches KillSwitchStatus, controlledSignupReady bool) {
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
		out.Body.KillSwitches = switches
		out.Body.ControlledSignupReady = controlledSignupReady
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

// RegisterMonitoringSentryTest registers the deliberate-Sentry-test-event
// endpoint only in the production environment. It is gated on both a
// non-empty token AND environment == "production" so that staging or any
// other tier that happens to have MONITORING_TEST_TOKEN set does not also
// expose the route.
func RegisterMonitoringSentryTest(api huma.API, expectedToken, environment string) {
	if strings.TrimSpace(expectedToken) == "" || environment != "production" {
		return
	}
	huma.Register(api, huma.Operation{
		OperationID: "PostMonitoringSentryTest",
		Method:      http.MethodPost,
		Path:        "/ops/monitoring/sentry-test",
		Summary:     "Emit a deliberate Sentry test event",
		Tags:        []string{"Operations"},
	}, func(ctx context.Context, input *MonitoringSentryTestInput) (*MonitoringSentryTestOutput, error) {
		token := strings.TrimSpace(strings.TrimPrefix(input.Authorization, "Bearer "))
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			return nil, huma.Error401Unauthorized("monitoring test token missing or invalid")
		}

		eventID := sentry.CaptureMessage(
			fmt.Sprintf("VOC-037-T04 monitoring test event (env=%s)", environment),
		)
		sentry.Flush(2 * time.Second)

		// A 2xx with no event ID does not prove the event reached Sentry
		// (e.g. DSN unset, network failure swallowed by the SDK) - treat
		// that as a failure rather than a false "accepted".
		if eventID == nil {
			return nil, huma.Error502BadGateway("Sentry did not return an event ID - DSN may be unset or unreachable")
		}

		out := &MonitoringSentryTestOutput{}
		out.Body.Status = "accepted"
		out.Body.EventID = string(*eventID)
		out.Body.Timestamp = time.Now().UTC().Format(time.RFC3339)
		return out, nil
	})
}

// RegisterSyntheticSmokeTestSessionMint registers a token-gated endpoint that
// mints a real session for only the reserved synthetic smoke-test user.
// The route is fail-closed: when expectedToken is empty, no route is registered.
func RegisterSyntheticSmokeTestSessionMint(api huma.API, authSvc *auth.Service, expectedToken string) {
	if strings.TrimSpace(expectedToken) == "" {
		return
	}
	huma.Register(api, huma.Operation{
		OperationID: "PostSyntheticSmokeTestSessionMint",
		Method:      http.MethodPost,
		Path:        "/ops/synthetic-smoke-test/session",
		Summary:     "Mint a synthetic smoke-test session cookie",
		Tags:        []string{"Operations"},
	}, func(ctx context.Context, input *SyntheticSmokeTestSessionMintInput) (*SyntheticSmokeTestSessionMintOutput, error) {
		token := strings.TrimSpace(strings.TrimPrefix(input.Authorization, "Bearer "))
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			return nil, huma.Error401Unauthorized("smoke-test mint token missing or invalid")
		}

		session, sessionToken, err := authSvc.MintSyntheticSmokeTestSession(ctx)
		if err != nil {
			switch err {
			case auth.ErrSyntheticSessionMintDisabled:
				return nil, huma.Error503ServiceUnavailable("synthetic smoke-test session minting is disabled")
			case auth.ErrSyntheticUserNotSeeded:
				return nil, huma.Error503ServiceUnavailable("synthetic smoke-test user is not seeded")
			case auth.ErrUserDisabled:
				return nil, huma.Error503ServiceUnavailable("synthetic smoke-test user is disabled")
			default:
				return nil, huma.Error500InternalServerError("failed to mint synthetic smoke-test session")
			}
		}

		c := authHumaContext(ctx)
		c.AppendHeader("Set-Cookie", authSvc.SessionCookie(sessionToken, session.ExpiresAt).String())
		csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
		if csrfCookie == nil || csrfToken == "" {
			return nil, huma.Error500InternalServerError("failed to issue csrf cookie")
		}
		c.AppendHeader("Set-Cookie", csrfCookie.String())

		out := &SyntheticSmokeTestSessionMintOutput{}
		out.Body.Status = "issued"
		out.Body.SessionCookie = sessionToken
		out.Body.CSRFToken = csrfToken
		out.Body.ExpiresAt = session.ExpiresAt.UTC().Format(time.RFC3339)
		return out, nil
	})
}
