package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
)

// fakeHealthchecker is a small Healthchecker stub that returns
// the configured error (or nil) for PingContext. It is
// sufficient for the /healthz and NewProductionAPI tests; the
// full PostgreSQL stack is exercised by T09 in staging, not
// here.
type fakeHealthchecker struct {
	err   error
	calls int
}

func (f *fakeHealthchecker) PingContext(ctx context.Context) error {
	f.calls++
	return f.err
}

func newProductionTestConfig() ProductionConfig {
	return ProductionConfig{
		Port:            "8080",
		DatabaseURL:     "postgres://example/db",
		Environment:     "staging",
		BaseURL:         "https://staging.vocanova.site",
		MagicLinkPath:   "/auth/magic",
		OAuthRedirect:   "https://api-staging.vocanova.site/auth/oauth/google/callback",
		OAuthReturnURLs: []string{"https://api-staging.vocanova.site/auth/oauth/google/callback"},
		SessionDomain:   "staging.vocanova.site",
		SessionSecure:   true,
		SessionLifetime: 30 * 24 * time.Hour,
		APIProvider:     "opencode",
		APIBaseURL:      "http://127.0.0.1:4096",
		APIKey:          "test-key",
		APIModel:        "opencode-go/deepseek-v4-pro",
		APITimeout:      8 * time.Second,
		AIEnabled:       true,
		MagicLinkOn:     true,
		OAuthOn:         true,
		NewSignupsOn:    true,
	}
}

func TestLoadProductionConfig_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")

	_, err := LoadProductionConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoadProductionConfig_RequiresBaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")

	_, err := LoadProductionConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BASE_URL")
}

func TestLoadProductionConfig_RequiresOAuthRedirect(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")

	_, err := LoadProductionConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAUTH_REDIRECT_URI")
}

func TestLoadProductionConfig_RequiresSessionDomain(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "")

	_, err := LoadProductionConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SESSION_COOKIE_DOMAIN")
}

func TestLoadProductionConfig_DefaultsAreSensible(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("PORT", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("AI_FEATURES_ENABLED", "")
	t.Setenv("EMAIL_MAGIC_LINK_ENABLED", "")
	t.Setenv("GOOGLE_OAUTH_ENABLED", "")
	t.Setenv("NEW_USER_SIGNUP_ENABLED", "")
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("AI_PROVIDER_BASE_URL", "")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port, "PORT must default to 8080 when unset")
	assert.Equal(t, "staging", cfg.Environment, "ENVIRONMENT must default to staging when unset")
	assert.True(t, cfg.AIEnabled, "AI_FEATURES_ENABLED must default to true when unset")
	assert.True(t, cfg.MagicLinkOn, "EMAIL_MAGIC_LINK_ENABLED must default to true when unset")
	assert.True(t, cfg.OAuthOn, "GOOGLE_OAUTH_ENABLED must default to true when unset")
	assert.True(t, cfg.NewSignupsOn, "NEW_USER_SIGNUP_ENABLED must default to true when unset")
	assert.Equal(t, "opencode", cfg.APIProvider, "AI_PROVIDER must default to opencode when unset")
	assert.Equal(t, "http://127.0.0.1:4096", cfg.APIBaseURL, "AI_PROVIDER_BASE_URL must default to local opencode serve when unset")
}

func TestLoadProductionConfig_HonorsExplicitDisables(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("AI_FEATURES_ENABLED", "false")
	t.Setenv("EMAIL_MAGIC_LINK_ENABLED", "false")
	t.Setenv("GOOGLE_OAUTH_ENABLED", "false")
	t.Setenv("NEW_USER_SIGNUP_ENABLED", "false")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.False(t, cfg.AIEnabled)
	assert.False(t, cfg.MagicLinkOn)
	assert.False(t, cfg.OAuthOn)
	assert.False(t, cfg.NewSignupsOn)
}

func TestLoadProductionConfig_ParseAllowlist(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_ALLOWLIST", " https://a.example.com , https://b.example.com , ")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{
		"https://a.example.com",
		"https://b.example.com",
	}, cfg.OAuthReturnURLs, "allowlist must be parsed and whitespace-trimmed")
}

func TestLoadProductionConfig_FallsBackToRedirectWhenAllowlistEmpty(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_ALLOWLIST", "")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{cfg.OAuthRedirect}, cfg.OAuthReturnURLs, "empty allowlist must fall back to OAUTH_REDIRECT_URI")
}

// TestNewProductionAPI_RequiresDatabaseReachability covers the
// startup-time safety property the task requires: when the
// caller asks NewProductionAPI to open a connection against an
// unreachable URL, the function must refuse to wire a server.
// /healthz is a runtime probe, but a database that is already
// known-down at startup must not produce a runnable server.
func TestNewProductionAPI_RequiresDatabaseReachability(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://invalid:invalid@127.0.0.1:1/db?connect_timeout=1")
	_, _, err := NewProductionAPI(newProductionTestConfig(), nil)
	require.Error(t, err, "NewProductionAPI must refuse to build a server when the database is unreachable")
	assert.Contains(t, err.Error(), "ping database")
}

// TestNewProductionAPI_BuildsWithReachableDatabase covers the
// happy path: when the supplied *sql.DB is reachable (sqlmock),
// NewProductionAPI returns a fully-wired huma.API that exposes
// the /healthz probe.
func TestNewProductionAPI_BuildsWithReachableDatabase(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer sqlDB.Close()
	mock.ExpectPing()

	api, returnedDB, err := NewProductionAPI(newProductionTestConfig(), sqlDB)
	require.NoError(t, err)
	require.NotNil(t, api)
	assert.Same(t, sqlDB, returnedDB, "NewProductionAPI must return the same *sql.DB the caller supplied so it can be closed on shutdown")

	// /healthz should be registered.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "/healthz must return 200 when database ping succeeds")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestHealthzHandler_ReportsUnhealthyWithoutLeakingError covers
// the runtime probe's security property: a failing Ping must
// return 503 with a stable body, and must never surface the
// underlying driver error message (which can contain the host
// name, port, or other infrastructure detail the founder's
// network should not advertise).
func TestHealthzHandler_ReportsUnhealthyWithoutLeakingError(t *testing.T) {
	db := &fakeHealthchecker{err: errors.New("dial tcp 10.0.0.1:5432: i/o timeout (secret host)")}
	api := newHealthzOnlyAPI(t, db)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code, "/healthz must return 503 when the database ping fails")

	body := w.Body.String()
	assert.NotContains(t, strings.ToLower(body), "secret host", "/healthz must not leak the underlying error message")
	assert.NotContains(t, body, "10.0.0.1", "/healthz must not leak the underlying error message")
}

// TestHealthzHandler_ReportsOkWhenPingSucceeds covers the happy
// runtime path: a successful Ping returns 200 with a structured
// body the orchestrator (Docker HEALTHCHECK, k8s liveness probe)
// can parse.
func TestHealthzHandler_ReportsOkWhenPingSucceeds(t *testing.T) {
	db := &fakeHealthchecker{err: nil}
	api := newHealthzOnlyAPI(t, db)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "/healthz must return 200 when the database ping succeeds")

	var body struct {
		Status    string `json:"status"`
		Database  string `json:"database"`
		Timestamp string `json:"timestamp"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "ok", body.Database)
	assert.NotEmpty(t, body.Timestamp)
}

// newHealthzOnlyAPI builds a real huma.API that only has the
// /healthz route registered against db, so the runtime
// assertions above can probe the handler in isolation without
// paying for the full production wiring.
func newHealthzOnlyAPI(t *testing.T, db Healthchecker) huma.API {
	t.Helper()
	cfg := huma.DefaultConfig("Vocanova API", "0.1.0")
	mux := chi.NewMux()
	api := humachi.New(mux, cfg)
	RegisterHealthz(api, db)
	return api
}

// os import guard - kept to keep the import honest across
// future refactors; the test exercises env vars via t.Setenv
// (which uses os.Setenv under the hood) and t.Setenv cleans up
// after itself.
var _ = os.Getenv

// TestBuildEmailSender_FallsBackToFakeWhenKillSwitchOff covers
// the first T14 fallback rule: when EMAIL_MAGIC_LINK_ENABLED is
// "false" the production wiring always uses Fake{}, even if a
// real API key is also configured. The kill switch wins.
func TestBuildEmailSender_FallsBackToFakeWhenKillSwitchOff(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.MagicLinkOn = false
	cfg.EmailProviderURL = "https://api.example.com/emails"
	cfg.EmailProviderAPIKey = "should-be-ignored"
	cfg.EmailFrom = "Vocanova <[email protected]>"

	s, err := buildEmailSender(cfg)
	require.NoError(t, err)
	_, ok := s.(*email.Fake)
	assert.True(t, ok, "kill switch off must force Fake{} even when a credential is also set")
}

// TestBuildEmailSender_FallsBackToFakeWhenNoAPIKey covers the
// second T14 fallback rule: when the kill switch is on but no API
// key is set, the wiring falls back to Fake{} so staging can
// still run with magic-link delivery off at the provider layer
// rather than crashing at startup.
func TestBuildEmailSender_FallsBackToFakeWhenNoAPIKey(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.MagicLinkOn = true
	cfg.EmailProviderAPIKey = ""

	s, err := buildEmailSender(cfg)
	require.NoError(t, err)
	_, ok := s.(*email.Fake)
	assert.True(t, ok, "missing API key must fall back to Fake{} rather than fail")
}

// TestBuildEmailSender_BuildsHTTPSenderWhenFullyConfigured covers
// the happy path: with the kill switch on, a non-empty API key,
// URL, and From, the wiring constructs a real HTTPSender.
func TestBuildEmailSender_BuildsHTTPSenderWhenFullyConfigured(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.MagicLinkOn = true
	cfg.EmailProviderURL = "https://api.example.com/emails"
	cfg.EmailProviderAPIKey = "test-key"
	cfg.EmailFrom = "Vocanova <[email protected]>"
	cfg.EmailProviderTimeout = 5 * time.Second

	s, err := buildEmailSender(cfg)
	require.NoError(t, err)
	hs, ok := s.(*email.HTTPSender)
	require.True(t, ok, "fully-configured wiring must produce a real HTTPSender")
	assert.Equal(t, "https://api.example.com/emails", hs.URL)
	assert.Equal(t, "test-key", hs.APIKey)
	assert.Equal(t, "Vocanova <[email protected]>", hs.From)
	assert.Equal(t, 5*time.Second, hs.Client.Timeout)
}

// TestBuildEmailSender_HardErrorsOnMisconfiguredHTTPSender covers
// the "API key set but URL or From missing" path: a half-
// configured real sender is a hard startup error, not a silent
// Fake{} fallback. Silently falling back here would hide a real
// configuration mistake.
func TestBuildEmailSender_HardErrorsOnMisconfiguredHTTPSender(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.MagicLinkOn = true
	cfg.EmailProviderAPIKey = "test-key"
	cfg.EmailProviderURL = ""
	cfg.EmailFrom = "Vocanova <[email protected]>"

	_, err := buildEmailSender(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL is required")
}

// TestBuildEmailSender_HardErrorsOnMissingFrom covers the
// second half-configured path: API key and URL set but no From
// address. Same hard-error posture as the missing-URL case.
func TestBuildEmailSender_HardErrorsOnMissingFrom(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.MagicLinkOn = true
	cfg.EmailProviderAPIKey = "test-key"
	cfg.EmailProviderURL = "https://api.example.com/emails"
	cfg.EmailFrom = ""

	_, err := buildEmailSender(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "From is required")
}

// TestLoadProductionConfig_ReadsEmailProviderEnvVars covers the
// T01/T14 contract: every env var the production wiring reads
// for the email provider is exposed on ProductionConfig with the
// correct value. This is the half of T01's
// VOC-032-TEST-05 .env.example-completeness check that focuses
// specifically on the T14 additions.
func TestLoadProductionConfig_ReadsEmailProviderEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("EMAIL_PROVIDER_URL", "https://api.example.com/emails")
	t.Setenv("EMAIL_PROVIDER_API_KEY", "k")
	t.Setenv("EMAIL_FROM", "Vocanova <[email protected]>")
	t.Setenv("EMAIL_PROVIDER_TIMEOUT", "7s")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/emails", cfg.EmailProviderURL)
	assert.Equal(t, "k", cfg.EmailProviderAPIKey)
	assert.Equal(t, "Vocanova <[email protected]>", cfg.EmailFrom)
	assert.Equal(t, 7*time.Second, cfg.EmailProviderTimeout)
}

// TestLoadProductionConfig_EmailProviderTimeoutDefaults covers
// the timeout default: when EMAIL_PROVIDER_TIMEOUT is unset, the
// production wiring uses a positive default rather than the
// net/http zero value (which disables timeouts entirely).
func TestLoadProductionConfig_EmailProviderTimeoutDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("EMAIL_PROVIDER_TIMEOUT", "")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Greater(t, cfg.EmailProviderTimeout, time.Duration(0))
}

// TestBuildOAuthProvider_FallsBackToFakeWhenKillSwitchOff covers
// the first T15 fallback rule: when GOOGLE_OAUTH_ENABLED is
// "false" the production wiring always uses FakeOAuthProvider,
// even if a real client ID is also configured. The kill switch
// wins.
func TestBuildOAuthProvider_FallsBackToFakeWhenKillSwitchOff(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = false
	cfg.GoogleClientID = "should-be-ignored"
	cfg.GoogleClientSecret = "should-be-ignored"

	p, err := buildOAuthProvider(cfg)
	require.NoError(t, err)
	_, ok := p.(*auth.FakeOAuthProvider)
	assert.True(t, ok, "kill switch off must force FakeOAuthProvider even when credentials are also set")
}

// TestBuildOAuthProvider_FallsBackToFakeWhenNoClientID covers
// the second T15 fallback rule: when the kill switch is on but
// no client ID is set, the wiring falls back to FakeOAuthProvider
// so staging can still run with Google sign-in off at the
// provider layer rather than crashing at startup.
func TestBuildOAuthProvider_FallsBackToFakeWhenNoClientID(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = true
	cfg.GoogleClientID = ""

	p, err := buildOAuthProvider(cfg)
	require.NoError(t, err)
	_, ok := p.(*auth.FakeOAuthProvider)
	assert.True(t, ok, "missing client ID must fall back to FakeOAuthProvider rather than fail")
}

// TestBuildOAuthProvider_BuildsGoogleProviderWhenFullyConfigured
// covers the happy path: with the kill switch on, a non-empty
// client ID, client secret, and redirect URI, the wiring
// constructs a real GoogleOAuthProvider.
func TestBuildOAuthProvider_BuildsGoogleProviderWhenFullyConfigured(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = true
	cfg.GoogleClientID = "test-client-id"
	cfg.GoogleClientSecret = "test-client-secret"
	cfg.GoogleOAuthTimeout = 5 * time.Second
	cfg.GoogleOAuthScopes = "openid email profile https://www.googleapis.com/auth/drive.readonly"

	p, err := buildOAuthProvider(cfg)
	require.NoError(t, err)
	gp, ok := p.(*auth.GoogleOAuthProvider)
	require.True(t, ok, "fully-configured wiring must produce a real GoogleOAuthProvider")
	assert.Equal(t, "test-client-id", gp.ClientID)
	assert.Equal(t, "test-client-secret", gp.ClientSecret)
	assert.Equal(t, "https://api-staging.vocanova.site/auth/oauth/google/callback", gp.RedirectURI)
	assert.Equal(t, "openid email profile https://www.googleapis.com/auth/drive.readonly", gp.Scopes, "fully-configured wiring must forward custom scopes verbatim")
	assert.Equal(t, 5*time.Second, gp.Client.Timeout, "fully-configured wiring must honor the custom timeout")
}

// TestBuildOAuthProvider_DefaultsScopes covers the scopes
// default: when GOOGLE_OAUTH_SCOPES is unset, the production
// wiring uses the auth package's DefaultGoogleOAuthScopes
// ("openid email profile") rather than crashing or producing
// an empty scope set.
func TestBuildOAuthProvider_DefaultsScopes(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = true
	cfg.GoogleClientID = "test-client-id"
	cfg.GoogleClientSecret = "test-client-secret"
	cfg.GoogleOAuthScopes = ""

	p, err := buildOAuthProvider(cfg)
	require.NoError(t, err)
	gp := p.(*auth.GoogleOAuthProvider)
	assert.Equal(t, auth.DefaultGoogleOAuthScopes, gp.Scopes, "missing GOOGLE_OAUTH_SCOPES must fall back to the default scope set")
}

// TestBuildOAuthProvider_DefaultsTimeout covers the timeout
// default: when GOOGLE_OAUTH_TIMEOUT is unset, the production
// wiring uses a positive default rather than the net/http zero
// value (which disables timeouts entirely).
func TestBuildOAuthProvider_DefaultsTimeout(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = true
	cfg.GoogleClientID = "test-client-id"
	cfg.GoogleClientSecret = "test-client-secret"
	cfg.GoogleOAuthTimeout = 0

	p, err := buildOAuthProvider(cfg)
	require.NoError(t, err)
	gp := p.(*auth.GoogleOAuthProvider)
	assert.Greater(t, gp.Client.Timeout, time.Duration(0))
}

// TestBuildOAuthProvider_HardErrorsOnMissingSecret covers the
// "client ID set but client secret missing" path: a half-
// configured real provider is a hard startup error, not a
// silent FakeOAuthProvider fallback. Silently falling back
// here would hide a real configuration mistake.
func TestBuildOAuthProvider_HardErrorsOnMissingSecret(t *testing.T) {
	cfg := newProductionTestConfig()
	cfg.OAuthOn = true
	cfg.GoogleClientID = "test-client-id"
	cfg.GoogleClientSecret = ""

	_, err := buildOAuthProvider(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GOOGLE_OAUTH_CLIENT_SECRET")
}

// TestLoadProductionConfig_ReadsGoogleOAuthEnvVars covers the
// T15/T01 contract: every env var the production wiring reads
// for the Google OAuth provider is exposed on ProductionConfig
// with the correct value. This is the half of T01's
// VOC-032-TEST-05 .env.example-completeness check that focuses
// specifically on the T15 additions.
func TestLoadProductionConfig_ReadsGoogleOAuthEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "real-google-client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "real-google-client-secret")
	t.Setenv("GOOGLE_OAUTH_SCOPES", "openid email profile")
	t.Setenv("GOOGLE_OAUTH_TIMEOUT", "7s")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Equal(t, "real-google-client-id", cfg.GoogleClientID)
	assert.Equal(t, "real-google-client-secret", cfg.GoogleClientSecret)
	assert.Equal(t, "openid email profile", cfg.GoogleOAuthScopes)
	assert.Equal(t, 7*time.Second, cfg.GoogleOAuthTimeout)
}

// TestLoadProductionConfig_GoogleOAuthTimeoutDefaults covers
// the timeout default: when GOOGLE_OAUTH_TIMEOUT is unset, the
// production wiring uses a positive default rather than the
// net/http zero value (which disables timeouts entirely).
func TestLoadProductionConfig_GoogleOAuthTimeoutDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("GOOGLE_OAUTH_TIMEOUT", "")

	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	assert.Greater(t, cfg.GoogleOAuthTimeout, time.Duration(0))
}
