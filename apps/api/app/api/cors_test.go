package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorsAllowedOrigins_DerivesFromReturnURLs(t *testing.T) {
	origins := corsAllowedOrigins([]string{
		"https://production.vocanova.site:8443/onboarding",
		"https://production.vocanova.site:8443/home",
		"https://staging.vocanova.site/onboarding",
	})

	assert.Equal(t, []string{
		"https://production.vocanova.site:8443",
		"https://staging.vocanova.site",
	}, origins, "duplicate origins from different paths must collapse to one entry, order preserved")
}

func TestCorsAllowedOrigins_SkipsMalformedEntries(t *testing.T) {
	origins := corsAllowedOrigins([]string{
		"not a url at all \x7f",
		"https://production.vocanova.site:8443/onboarding",
		"",
	})

	assert.Equal(t, []string{"https://production.vocanova.site:8443"}, origins,
		"a malformed entry must not prevent well-formed entries from being allowed")
}

func TestCorsMiddleware_AllowedOriginGetsHeaders(t *testing.T) {
	mw := corsMiddleware([]string{"https://production.vocanova.site:8443"})
	handlerCalled := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	req.Header.Set("Origin", "https://production.vocanova.site:8443")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "a normal (non-preflight) request must still reach the handler")
	assert.Equal(t, "https://production.vocanova.site:8443", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCorsMiddleware_DisallowedOriginGetsNoHeaders(t *testing.T) {
	mw := corsMiddleware([]string{"https://production.vocanova.site:8443"})
	handlerCalled := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "an untrusted origin still reaches the handler for a simple request - the browser enforces the missing header, not this middleware")
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"), "an untrusted origin must never be echoed back")
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCorsMiddleware_PreflightFromAllowedOrigin(t *testing.T) {
	mw := corsMiddleware([]string{"https://production.vocanova.site:8443"})
	handlerCalled := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/oauth/google/start", nil)
	req.Header.Set("Origin", "https://production.vocanova.site:8443")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.False(t, handlerCalled, "a preflight request must be answered by the middleware, never forwarded to the real handler")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://production.vocanova.site:8443", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Equal(t, "content-type", rec.Header().Get("Access-Control-Allow-Headers"))
}

func TestCorsMiddleware_PreflightFromDisallowedOriginGetsNoAccessControlHeaders(t *testing.T) {
	mw := corsMiddleware([]string{"https://production.vocanova.site:8443"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight must never reach the real handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/oauth/google/start", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code, "the OPTIONS request itself still gets a response, just without any Access-Control-* headers - the browser is what actually blocks the follow-up request")
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
}

func TestCorsMiddleware_OptionsWithoutPreflightHeaderIsNotTreatedAsPreflight(t *testing.T) {
	// A plain OPTIONS request (no Access-Control-Request-Method) is not a
	// CORS preflight - e.g. a client probing which methods a route
	// supports. It must reach the real handler like any other request,
	// not be swallowed as if it were a preflight.
	mw := corsMiddleware([]string{"https://production.vocanova.site:8443"})
	handlerCalled := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
}

func TestCorsAllowedOrigins_EmptyInputProducesEmptyAllowlist(t *testing.T) {
	origins := corsAllowedOrigins(nil)
	assert.Empty(t, origins)
	assert.Equal(t, "(none - all cross-origin requests will be rejected)", corsOriginsSummary(origins))
}
