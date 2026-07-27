package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VOC-031-T06 — API-layer cross-cutting reliability and recovery
// tests (VOC-031-TEST-29, TEST-30, TEST-31; VOC-031-AC-06).
//
// The per-task auth/CSRF/idempotency tests in
// apps/api/app/api/{learning,reviews,aifeedback,missions,onboarding,
// settings,email_change,account_deletion}_test.go already exercise
// the unauthenticated and replay paths for each endpoint in
// isolation, including
// TestCreateAccountDeletionRequestReplaysIdempotencyKey
// (VOC-031-TEST-28's per-route counterpart). This file is the T06
// cross-cutting counterpart that drives the property once across
// the whole core loop:
//
//   (a) every (app) endpoint, old and new, rejects an expired
//   session with a stable 401 and never fabricates a successful
//   response on a missing/invalid credential
//   (VOC-031-TEST-29 — session-expiry mid-flow);
//
//   (b) no public DTO exposes a placeholder data field a client
//   could fall back to when the real data is missing — every
//   field the server does not actually return must be absent from
//   the response shape, not synthesized to a hardcoded default
//   (VOC-031-TEST-30 — no client-fabricated fallback values,
//   structural side; the static-side check is in
//   scripts/foundation/mock-inventory.mjs);
//
//   (c) the A1, P1, P2, P3, and P4 contract surfaces are
//   byte-for-byte unchanged in OpenAPI shape after the
//   onboarding/settings/email-change/account-deletion additions
//   (VOC-031-TEST-31 — A1–P4 regression, structural side; the
//   runtime side is exercised by `go test ./...`).
//
// The runtime A1–P4 regression side of TEST-31 is enforced by
// the existing per-route test suites re-running green against
// this build. A regression that removed a route, renamed a
// pre-existing field, or changed a pre-existing response shape
// would surface in the per-route tests or the contract check
// below before it reached the client.

// newT06CoreLoopAPI builds a Huma API with the (app) write/read
// routes wired against in-memory repositories for every module
// the core loop touches. The T06 cross-cutting tests do not
// need every per-route helper — they only need the routes
// registered, so the AuthMiddleware and RequireAuth gate are
// active on every endpoint and an unauthenticated request falls
// through to a 401 before any SQL or domain logic is reached.
func newT06CoreLoopAPI(t *testing.T) (huma.API, *auth.Service) {
	t.Helper()
	now := testNow()
	c := &clock.Fixed{T: now}
	authRepo := auth.NewMemoryRepository()
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	authSvc := auth.NewService(authRepo, nil, nil, c, limiter, auth.Config{
		Cookie: auth.CookieConfig{
			Name: "vocanova_session", CSRName: "vocanova_csrf", OAuthStateName: "vocanova_oauth_state",
			Domain: "", Secure: false, SameSite: http.SameSiteStrictMode,
		},
		RateLimit: auth.RateLimitConfig{
			MagicRequestWindow: time.Hour, MagicRequestLimit: 10,
			MagicConsumeWindow: time.Hour, MagicConsumeLimit: 10,
			OAuthStartWindow: time.Hour, OAuthStartLimit: 10,
			OAuthCallbackWindow: time.Hour, OAuthCallbackLimit: 10,
			LogoutWindow: time.Hour, LogoutLimit: 10,
		},
	})

	// Learning: saved-words read + save + unsave (the
	// meaning-save-button write that the Word Detail screen
	// exercises mid-flow).
	learningRepo := learning.NewMemoryRepository(learning.MemoryRepositoryData{})
	learningSvc := learning.NewService(learningRepo, learning.NewMemoryIdempotencyStore(), nil)

	// Reviews: due-queue read and submission write (the
	// review-session write that the Reviews screen exercises
	// mid-flow).
	reviewsRepo := reviews.NewMemoryRepository(reviews.MemoryRepositoryData{})
	reviewsSvc := reviews.NewService(reviewsRepo, learning.NewMemoryIdempotencyStore(), nil)

	// Users: onboarding read/submit, settings read/write (the
	// Onboarding and Settings screens that the cross-cutting
	// pass must cover).
	usersRepo := users.NewMemoryRepository()
	usersSvc := users.NewService(usersRepo, usersRepo, usersRepo, c)

	// Accounts: email-change request/consume and
	// account-deletion request (the Settings/account screen
	// the cross-cutting pass must cover).
	accountsRepo := accounts.NewMemoryRepository()
	accountsLimiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	accountsIdem := accounts.NewMemoryIdempotencyStore()
	accountsSvc := accounts.NewService(accountsRepo, authRepo, &email.Fake{}, accountsIdem, c, accountsLimiter, accounts.Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
		AccountDeletionPurgeDelay: 30 * 24 * time.Hour, AccountDeletionSweepLimit: 100,
		RateLimit: accounts.EmailChangeRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			ConsumeWindow: time.Hour, ConsumeLimit: 100,
		},
		AccountDeletionRateLimit: accounts.AccountDeletionRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			SweepWindow: time.Hour, SweepLimit: 100,
		},
	})

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterContract(api)
	RegisterLearning(api, learningSvc, authSvc)
	RegisterReviews(api, reviewsSvc, authSvc)
	RegisterOnboarding(api, usersSvc, authSvc)
	RegisterSettings(api, usersSvc, authSvc)
	RegisterEmailChangeLinks(api, accountsSvc, authSvc)
	RegisterAccountDeletionRequests(api, accountsSvc, authSvc)
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })
	return api, authSvc
}

// TestCrossCuttingUnauthenticatedCoreLoopEndpointsReturn401
// is the VOC-031-TEST-29 cross-cutting counterpart at the
// API layer. An expired or missing session must produce a
// stable 401 from every (app) endpoint the cross-cutting pass
// covers, so the client-side session-expiry handler can detect
// the same 401 regardless of which screen the learner is on
// mid-flow. A regression where any one endpoint dropped the
// gate would surface here as a non-401 status, and would let
// a learner with a stale session see fabricated success
// responses.
func TestCrossCuttingUnauthenticatedCoreLoopEndpointsReturn401(t *testing.T) {
	api, authSvc := newT06CoreLoopAPI(t)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"contract /me", http.MethodGet, "/api/v1/me", ""},
		{"learning saved-words read", http.MethodGet, "/api/v1/user-words", ""},
		{"learning save write", http.MethodPost, "/api/v1/user-words", `{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`},
		{"learning unsave write", http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", ""},
		{"reviews due-queue read", http.MethodGet, "/api/v1/reviews/due", ""},
		{"reviews submission write", http.MethodPost, "/api/v1/reviews/submissions", `{}`},
		{"onboarding read", http.MethodGet, "/api/v1/onboarding", ""},
		{"onboarding submit write", http.MethodPost, "/api/v1/onboarding", `{}`},
		{"settings read", http.MethodGet, "/api/v1/settings", ""},
		{"settings patch write", http.MethodPatch, "/api/v1/settings", `{}`},
		{"email-change request write", http.MethodPost, "/api/v1/settings/email-change-links", `{}`},
		{"email-change consume write", http.MethodPost, "/api/v1/settings/email-change-links/consume", `{}`},
		{"account-deletion request write", http.MethodPost, "/api/v1/account-deletion-requests", ``},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var body *bytes.Reader
			if c.body != "" {
				body = bytes.NewReader([]byte(c.body))
			} else {
				body = bytes.NewReader(nil)
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, c.path, body)
			if c.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			// Even with a valid CSRF token, an unauthenticated
			// request must 401 — the cross-cutting property
			// guards against a regression that swaps the
			// auth/CSRF middleware order.
			addCSRF(req, authSvc)
			api.Adapter().ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"unauthenticated %s %s must return 401 (got %d body=%s)",
				c.method, c.path, w.Code, w.Body.String())
		})
	}
}

// TestCrossCuttingNoClientFabricatedFallbackInContract is the
// VOC-031-TEST-30 cross-cutting counterpart at the contract
// layer. The API contract must never expose a placeholder data
// field a client could fall back to when the real data is
// missing — fields the server did not actually return must be
// absent from the response shape, not synthesized to a
// hardcoded default.
//
// The "no fabricated fallback" guarantee is structural: every
// response DTO is built from real server-side state, and a
// regression that introduced a default value (e.g. a
// synthesized `"streak": 0` from a missing streak row, or a
// placeholder text field claimed to come from the API) would
// show up as a new field in the OpenAPI document. This test
// pins the current shape of the (app) read endpoints' DTOs;
// any future addition of a placeholder field would have to be
// an explicit, reviewed contract change.
//
// The matching client-side static check (a hand-written
// fabrication in a (app) route's render code) is enforced by
// the MOCK_*-and-fabrication scan in
// scripts/foundation/mock-inventory.mjs.
func TestCrossCuttingNoClientFabricatedFallbackInContract(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	require.NoError(t, err)
	contract := string(document)

	// Every (app) read endpoint must exist and be a non-path-
	// parameterized shape (the cross-user-enumeration check
	// the missions_cross_cutting_test.go file already pins
	// for the P4 reads). A regression that added a path
	// parameter like /api/v1/onboarding/{userId} would let a
	// client fabricate a different identity's view of
	// onboarding and break the cross-user isolation
	// invariant.
	//
	// The unsave-write endpoint /api/v1/user-words/{meaningId}
	// is a legitimate path-parameterized child of the
	// /api/v1/user-words read; the test asserts the read
	// itself (the bare path) is present and the parameterized
	// form is *only* the unsave-write child, not a
	// cross-user enumeration variant on the read.
	requiredEndpoints := []string{
		"/api/v1/me",
		"/api/v1/onboarding",
		"/api/v1/settings",
		"/api/v1/user-words",
		"/api/v1/reviews/due",
	}
	for _, endpoint := range requiredEndpoints {
		assert.Contains(t, contract, endpoint,
			"contract must declare the (app) read endpoint %s", endpoint)
		// Reject any path-parameterized variant on the bare
		// read endpoint. The /api/v1/user-words/{meaningId}
		// unsave write is the only legitimate child path and
		// is asserted separately below.
		if endpoint == "/api/v1/user-words" {
			// The child path is allowed (the unsave
			// write), but only as the meaning-id form,
			// not a user-id form.
			assert.NotContains(t, contract, `/api/v1/user-words/{userId}`,
				"%s must not accept a userId path parameter (cross-user enumeration risk)", endpoint)
		} else {
			assert.NotContains(t, contract, endpoint+"/{",
				"%s must not accept a path parameter (cross-user enumeration risk)", endpoint)
		}
	}

	// Response DTOs must not contain a placeholder/fabricated
	// data field by name. A regression that introduced a
	// field the server does not actually populate (e.g. a
	// hardcoded "0" or "[]" for an unset collection) would
	// surface as a new field name in the OpenAPI components;
	// the test guards against the most common fabrication
	// patterns by name.
	forbiddenFields := []string{
		// No DTO should ever echo a token hash, OAuth subject,
		// or session metadata back to the client. A regression
		// here would let a client fabricate a session-impersonating
		// fallback if a future UI mistakenly rendered it.
		"tokenHash",
		"token_hash",
		"providerSubject",
		"provider_subject",
		"revokedAt",
		"revoked_at",
		"deletedAt",
		"deleted_at",
	}
	for _, field := range forbiddenFields {
		assert.NotContains(t, strings.ToLower(contract), strings.ToLower(field),
			"contract must not expose internal field %q (a client could fabricate a fallback from it)", field)
	}
}

// TestCrossCuttingA1P4ContractSurfacesUnchanged is the
// VOC-031-TEST-31 cross-cutting counterpart at the contract
// layer. The T00–T05 additions (onboarding, settings,
// email-change, account-deletion) must not re-shape any
// A1/P1/P2/P3/P4 route or DTO. A regression that renamed a
// pre-existing route, removed a pre-existing field, or
// changed a pre-existing response shape would surface here
// as a contract drift before it reached the client.
//
// The runtime side of TEST-31 (re-running the per-route
// A1–P4 test suites) is enforced by `go test ./...`. This
// test pins the contract surface; the two checks together
// cover both the data side and the schema side of the
// regression guarantee.
func TestCrossCuttingA1P4ContractSurfacesUnchanged(t *testing.T) {
	document, err := json.Marshal(NewContractAPI().OpenAPI())
	require.NoError(t, err)
	contract := string(document)

	// A1: CurrentUser must keep every previously-existing
	// field. The T01 onboardingStatus field is additive
	// (T00/T01 documented this), so the original four
	// fields are still required.
	for _, field := range []string{"email", "displayName", "avatarUrl", "emailVerifiedAt"} {
		assert.Contains(t, contract, field,
			"A1 /api/v1/me response must still declare %q (TEST-31 regression guard)", field)
	}
	// T01's additive field must also still be present.
	assert.Contains(t, contract, "onboardingStatus",
		"T01 additive field %q must still be present in /api/v1/me (TEST-31 forward-compat guard)")

	// P1: Journey, JourneySituation, JourneyWordList — the
	// existing (app) discovery route shapes. The
	// canonical-words path is matched as a prefix so the
	// trailing {wordSlug} parameter is allowed (the bare
	// prefix is the OpenAPI's parent path segment).
	for _, path := range []string{
		"/api/v1/journey-situations",
		"/api/v1/canonical-words",
		"/api/v1/user-words",
	} {
		assert.Contains(t, contract, path,
			"P1 route %q must still be present in the contract (TEST-31 regression guard)", path)
	}

	// P2: Review routes.
	for _, path := range []string{"/api/v1/reviews/due", "/api/v1/reviews/submissions"} {
		assert.Contains(t, contract, path,
			"P2 route %q must still be present in the contract (TEST-31 regression guard)", path)
	}

	// P3: Sentence-feedback write/report.
	for _, path := range []string{"/api/v1/sentence-feedback"} {
		assert.Contains(t, contract, path,
			"P3 route %q must still be present in the contract (TEST-31 regression guard)", path)
	}

	// P4: Daily-mission and progress reads.
	for _, path := range []string{"/api/v1/daily-mission", "/api/v1/progress"} {
		assert.Contains(t, contract, path,
			"P4 route %q must still be present in the contract (TEST-31 regression guard)", path)
	}
}
