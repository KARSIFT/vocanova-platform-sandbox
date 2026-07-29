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
	_, err := NewProductionAPI(newProductionTestConfig(), nil)
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

	api, err := NewProductionAPI(newProductionTestConfig(), sqlDB)
	require.NoError(t, err)
	require.NotNil(t, api)

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
