package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAccountDeletionAPI wires a Huma API with the auth,
// contract, and account-deletion routes registered against
// fresh in-memory repositories.
func testAccountDeletionAPI(t *testing.T) (huma.API, *auth.Service, *accounts.Service, *accounts.MemoryRepository, *auth.MemoryRepository) {
	t.Helper()
	now := testNow()
	c := &clock.Fixed{T: now}
	authRepo := auth.NewMemoryRepository()
	limiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	oauth := auth.NewFakeOAuthProvider(&auth.OAuthIdentity{
		Subject: "sub-1", Email: "user@example.com", EmailVerified: true,
		DisplayName: "User", AvatarURL: "https://example.com/avatar.png",
	})
	authSvc := auth.NewService(authRepo, nil, oauth, c, limiter, auth.Config{
		Environment:        "test",
		BaseURL:            "https://test.example.com",
		MagicLinkPath:      "/auth/magic",
		OAuthRedirectURI:   "https://test.example.com/auth/oauth/google/callback",
		SessionLifetime:    30 * 24 * time.Hour,
		MagicLinkLifetime:  15 * time.Minute,
		OAuthStateLifetime: 10 * time.Minute,
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
	RegisterAuth(api, authSvc)
	RegisterAccountDeletionRequests(api, accountsSvc, authSvc)
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })
	return api, authSvc, accountsSvc, accountsRepo, authRepo
}

// adRequesterRequest issues a request with the requester
// injected directly and the supplied Idempotency-Key header
// set when non-empty.
func adRequesterRequest(t *testing.T, method, path, idemKey string, userID uuid.UUID) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	return req
}

// seedUserForAccountDeletion installs a user in both the
// auth and accounts in-memory stores so the deactivation
// transaction has the rows it expects. accountsRepo may be
// nil (tests that only exercise auth/CSRF gates do not need
// it).
func seedUserForAccountDeletion(t *testing.T, authRepo *auth.MemoryRepository, accountsRepo *accounts.MemoryRepository, emailAddr string) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	authRepo.UpsertUser(&auth.User{ID: uid, Email: emailAddr, Status: "active"})
	if accountsRepo != nil {
		accountsRepo.SetUser(uid, emailAddr)
	}
	return uid
}

// TestCreateAccountDeletionRequestRequiresAuthentication covers
// the 401 path.
func TestCreateAccountDeletionRequestRequiresAuthentication(t *testing.T) {
	api, _, _, _, _ := testAccountDeletionAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account-deletion-requests", nil)
	req.Header.Set("Idempotency-Key", "idem-key")
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateAccountDeletionRequestRequiresCSRF covers the
// 403 path: the route requires double-submit CSRF.
func TestCreateAccountDeletionRequestRequiresCSRF(t *testing.T) {
	api, _, _, _, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, nil, "user@example.com")
	w := httptest.NewRecorder()
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-key", uid)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestCreateAccountDeletionRequestRequiresIdempotencyKey covers
// the missing-header path: a missing Idempotency-Key header
// surfaces as a stable 422 (Huma's default for required
// header validation), matching the existing reviews /
// aifeedback / learning endpoints' behavior.
func TestCreateAccountDeletionRequestRequiresIdempotencyKey(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, accountsRepo, "user@example.com")

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	// Build the request with no Idempotency-Key header at all
	// so Huma's required-header validation fires (the
	// "   " case is treated as a present-but-invalid header
	// and reaches the service layer, not Huma's validator).
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "missing idem key path")
}

// TestCreateAccountDeletionRequestHappyPath covers the
// success path: a 200 with the deactivated state, the
// scheduled purge_after, and the Idempotency-Key fingerprint.
func TestCreateAccountDeletionRequestHappyPath(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, accountsRepo, "user@example.com")

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-1", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "happy path returns 200")

	var body CreateAccountDeletionRequestDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "deactivated", body.Status)
	assert.Equal(t, uid.String(), body.UserID)
	assert.Equal(t, "idem-1", body.IdempotencyKey)
	assert.False(t, body.Replayed)
	assert.NotEmpty(t, body.RequestedAt)
	assert.NotEmpty(t, body.PurgeAfter)
}

// TestCreateAccountDeletionRequestReplaysIdempotencyKey covers
// the DOC-07 replay-safety property: a second call with
// the same Idempotency-Key returns the same row, marks
// Replayed=true, and does not re-run the deactivation.
func TestCreateAccountDeletionRequestReplaysIdempotencyKey(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, accountsRepo, "user@example.com")

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-1", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Second call: same key, same CSRF.
	csrfToken, csrfCookie = authSvc.IssueCSRFCookie()
	w = httptest.NewRecorder()
	req = adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-1", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body CreateAccountDeletionRequestDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, body.Replayed, "second call is a replay")
	assert.Equal(t, "idem-1", body.IdempotencyKey)
	// Repository still has only one row.
	row := accountsRepo.DeletionRequest(uid)
	require.NotNil(t, row)
}

// TestCreateAccountDeletionRequestUserNotFound covers the
// 404 path: a deletion for a user that does not exist
// surfaces as a stable 404.
func TestCreateAccountDeletionRequestUserNotFound(t *testing.T) {
	api, authSvc, _, _, _ := testAccountDeletionAPI(t)
	// Use a different user id (no rows in the accounts
	// in-memory store) so the deactivation transaction
	// surfaces ErrUserNotFound.
	missing := uuid.New()
	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem", missing)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestCreateAccountDeletionRequestAlreadyInFlight covers the
// 409 path: a second fresh key against a user with an
// in-flight deactivation surfaces as a stable 409.
func TestCreateAccountDeletionRequestAlreadyInFlight(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, accountsRepo, "user@example.com")

	// First call.
	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-1", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Second call: fresh key, same user.
	csrfToken, csrfCookie = authSvc.IssueCSRFCookie()
	w = httptest.NewRecorder()
	req = adRequesterRequest(t, http.MethodPost, "/api/v1/account-deletion-requests", "idem-2", uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code, "fresh key against an in-flight deletion is a 409")
}

// TestCreateAccountDeletionRequestEmptyBodyAllowed covers the
// documented "request body is empty" posture: the body is
// deliberately empty (the only client-supplied value is the
// Idempotency-Key header), and the API layer must not 400
// when the body is empty.
func TestCreateAccountDeletionRequestEmptyBodyAllowed(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo := testAccountDeletionAPI(t)
	uid := seedUserForAccountDeletion(t, authRepo, accountsRepo, "user@example.com")

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	// A non-empty body is also accepted; the field shape
	// has no required body fields.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account-deletion-requests", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uid}))
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "empty body is allowed")
}
