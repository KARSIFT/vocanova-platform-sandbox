package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// testEmailChangeAPI wires a Huma API with the auth, contract,
// and email-change routes registered against a fresh in-memory
// auth repository and a fresh in-memory accounts repository.
// An auth.MemoryRepository is wired in as the AuthRepository the
// accounts service needs to look users up by ID, and the
// service's pre-seeded user is also mirrored into the
// accounts.MemoryRepository so the duplicate-email discipline
// observes the right state.
func testEmailChangeAPI(t *testing.T) (huma.API, *auth.Service, *accounts.Service, *accounts.MemoryRepository, *auth.MemoryRepository, *email.Fake) {
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
	fake := &email.Fake{}
	accountsLimiter := auth.NewFixedWindowRateLimiter(c, time.Hour, 100)
	accountsIdem := accounts.NewMemoryIdempotencyStore()
	accountsSvc := accounts.NewService(accountsRepo, authRepo, fake, accountsIdem, c, accountsLimiter, accounts.Config{
		Environment: "test", BaseURL: "https://test.example.com",
		EmailChangePath: "/auth/email-change", EmailChangeLinkLifetime: 15 * time.Minute,
		RateLimit: accounts.EmailChangeRateLimitConfig{
			RequestWindow: time.Hour, RequestLimit: 100,
			ConsumeWindow: time.Hour, ConsumeLimit: 100,
		},
	})

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterContract(api)
	RegisterAuth(api, authSvc)
	RegisterEmailChangeLinks(api, accountsSvc, authSvc)
	t.Cleanup(func() { SetOnboardingStatusLookup(nil) })
	return api, authSvc, accountsSvc, accountsRepo, authRepo, fake
}

// ecRequesterRequest issues a request with the requester injected
// directly, matching the helpers in onboarding_test.go and
// settings_test.go.
func ecRequesterRequest(t *testing.T, method, path, body string, userID uuid.UUID) *http.Request {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	return req
}

// seedUserForEmailChange installs a user with a known email in
// both the auth and accounts repositories so the email-change
// flow has a consistent view of the user table. The accounts
// repository is optional: tests that exercise only the auth/CSRF
// gate (e.g. unauthenticated, missing-CSRF) do not need it.
func seedUserForEmailChange(t *testing.T, authRepo *auth.MemoryRepository, accountsRepo *accounts.MemoryRepository, emailAddr string) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	authRepo.UpsertUser(&auth.User{ID: uid, Email: emailAddr, Status: "active"})
	if accountsRepo != nil {
		accountsRepo.SetUser(uid, emailAddr)
	}
	return uid
}

// TestRequestEmailChangeLinkRequiresAuthentication covers the
// 401 path: the route requires an authenticated requester.
func TestRequestEmailChangeLinkRequiresAuthentication(t *testing.T) {
	api, _, _, _, _, _ := testEmailChangeAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/email-change-links",
		strings.NewReader(`{"newEmail":"new@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRequestEmailChangeLinkRequiresCSRF covers the 403 path:
// the route requires double-submit CSRF.
func TestRequestEmailChangeLinkRequiresCSRF(t *testing.T) {
	api, _, _, _, authRepo, _ := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, nil, "old@example.com")
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links",
		`{"newEmail":"new@example.com"}`, uid)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestRequestEmailChangeLinkValidatesEmail covers the 400
// path: the new email must be a syntactically acceptable
// address.
func TestRequestEmailChangeLinkValidatesEmail(t *testing.T) {
	api, authSvc, _, _, authRepo, _ := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, nil, "old@example.com")

	for _, body := range []string{
		`{"newEmail":""}`,
		`{"newEmail":"  "}`,
		`{"newEmail":"no-at-sign"}`,
		`{"newEmail":"no-domain@"}`,
	} {
		csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
		w := httptest.NewRecorder()
		req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links", body, uid)
		req.AddCookie(csrfCookie)
		req.Header.Set("X-CSRF-Token", csrfToken)
		api.Adapter().ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusOK, w.Code, "body=%s", body)
		assert.Less(t, w.Code, 500, "body=%s", body)
	}
}

// TestRequestEmailChangeLinkHappyPathGenericResponse covers
// VOC-031-TEST-12: the request returns a generic 200 with no
// payload regardless of whether the new email is already
// registered. The email is dispatched to the new address.
func TestRequestEmailChangeLinkHappyPathGenericResponse(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo, fake := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, accountsRepo, "old@example.com")
	other := seedUserForEmailChange(t, authRepo, accountsRepo, "taken@example.com")
	_ = other

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links",
		`{"newEmail":"taken@example.com"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code, "request is unconditionally generic")

	require.Len(t, fake.Sent, 1, "email dispatched to the NEW address, not the existing owner")
	assert.Equal(t, "taken@example.com", fake.Sent[0].To[0].Email)
	assert.Contains(t, fake.Sent[0].BodyText, "/auth/email-change")
}

// TestConsumeEmailChangeLinkRequiresAuthentication covers the
// 401 path: the consume route requires an authenticated
// requester.
func TestConsumeEmailChangeLinkRequiresAuthentication(t *testing.T) {
	api, _, _, _, _, _ := testEmailChangeAPI(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/email-change-links/consume",
		strings.NewReader(`{"token":"some-token"}`))
	req.Header.Set("Content-Type", "application/json")
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestConsumeEmailChangeLinkRequiresCSRF covers the 403 path.
func TestConsumeEmailChangeLinkRequiresCSRF(t *testing.T) {
	api, _, _, _, authRepo, _ := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, nil, "old@example.com")
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links/consume",
		`{"token":"some-token"}`, uid)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestConsumeEmailChangeLinkRejectsInvalidToken covers the
// 401 path on an unknown/expired/tampered token: a stable
// error, never a 500, never a distinguishing detail.
func TestConsumeEmailChangeLinkRejectsInvalidToken(t *testing.T) {
	api, authSvc, _, _, authRepo, _ := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, nil, "old@example.com")

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links/consume",
		`{"token":"not-a-real-token"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "invalid token must produce 401, never 500")
}

// TestConsumeEmailChangeLinkHappyPath covers VOC-031-TEST-16:
// a successful confirm returns 200 with the new email and the
// previous email, and the user's email in the in-memory
// repository is updated. A second message is dispatched to the
// OLD address as the security notification.
func TestConsumeEmailChangeLinkHappyPath(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo, fake := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, accountsRepo, "old@example.com")

	// Request a change first.
	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links",
		`{"newEmail":"new@example.com"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.Len(t, fake.Sent, 1)
	token := extractTokenFromEmailBody(t, fake.Sent[0].BodyText)

	// Confirm.
	csrfToken, csrfCookie = authSvc.IssueCSRFCookie()
	w = httptest.NewRecorder()
	req = ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links/consume",
		`{"token":"`+token+`"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body ConsumeEmailChangeLinkDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "new@example.com", body.Email)
	assert.Equal(t, "old@example.com", body.PreviousEmail)
	assert.NotEmpty(t, body.ChangedAt)

	// The user record reflects the change.
	assert.Equal(t, "new@example.com", accountsRepo.UserEmail(uid))

	// Two messages: request (to NEW) + notification (to OLD).
	require.Len(t, fake.Sent, 2)
	notification := fake.Sent[1]
	assert.Equal(t, "old@example.com", notification.To[0].Email)
	assert.Equal(t, "Your Vocanova sign-in email was changed", notification.Subject)
}

// TestConsumeEmailChangeLinkDuplicateEmailConflict covers
// VOC-031-TEST-15: a confirm that loses the duplicate-email
// race receives a stable 409, never a 500.
//
// The race is constructed by seeding the in-memory user table so
// the partial unique index already has the new email taken by
// another user before the consume attempt. The same
// UpdateUserEmail implementation the SQL repository applies is
// in the in-memory store, so this is an end-to-end exercise of
// the duplicate-email discipline.
func TestConsumeEmailChangeLinkDuplicateEmailConflict(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo, fake := testEmailChangeAPI(t)

	uidA := seedUserForEmailChange(t, authRepo, accountsRepo, "a-old@example.com")
	uidB := seedUserForEmailChange(t, authRepo, accountsRepo, "b-old@example.com")
	other := uuid.New()
	authRepo.UpsertUser(&auth.User{ID: other, Email: "shared@example.com", Status: "active"})
	accountsRepo.SetUser(other, "shared@example.com")

	// uidB requests a change to "shared@example.com".
	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links",
		`{"newEmail":"shared@example.com"}`, uidB)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	tokenB := extractTokenFromEmailBody(t, fake.Sent[0].BodyText)

	// uidB's confirm must produce a stable 409: the partial
	// unique index already has shared@example.com owned by
	// `other`, and uidB's UPDATE loses.
	csrfToken, csrfCookie = authSvc.IssueCSRFCookie()
	w = httptest.NewRecorder()
	req = ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links/consume",
		`{"token":"`+tokenB+`"}`, uidB)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	// uidA's identity is unaffected.
	assert.Equal(t, "a-old@example.com", accountsRepo.UserEmail(uidA))
	_ = uidA
}

// TestConsumeEmailChangeLinkGoogleOAuthUnaffected covers
// VOC-031-TEST-17: the email-change flow never touches the
// external_identities table, so a Google-OAuth-linked account
// (matched by provider_subject) keeps its sign-in path. The
// test pins this structurally: the accounts service has no
// external_identities mutation in its request/consume path,
// and the user's email field on the in-memory identity stays
// unrelated to external_identities.
func TestConsumeEmailChangeLinkGoogleOAuthUnaffected(t *testing.T) {
	api, authSvc, _, accountsRepo, authRepo, fake := testEmailChangeAPI(t)
	uid := seedUserForEmailChange(t, authRepo, accountsRepo, "old@example.com")

	// Pre-existing Google identity: matched by provider_subject,
	// not by email.
	const providerSubject = "google-sub-abc-123"
	_, err := authRepo.CreateExternalIdentity(context.Background(), uid, "google", providerSubject, "old@example.com", true)
	require.NoError(t, err)

	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	w := httptest.NewRecorder()
	req := ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links",
		`{"newEmail":"new@example.com"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	token := extractTokenFromEmailBody(t, fake.Sent[0].BodyText)

	csrfToken, csrfCookie = authSvc.IssueCSRFCookie()
	w = httptest.NewRecorder()
	req = ecRequesterRequest(t, http.MethodPost, "/api/v1/settings/email-change-links/consume",
		`{"token":"`+token+`"}`, uid)
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// The Google identity is still resolvable by provider_subject
	// and still points at the same user_id, even though
	// users.email is now "new@example.com".
	ident, err := authRepo.GetExternalIdentity(context.Background(), "google", providerSubject)
	require.NoError(t, err)
	assert.Equal(t, uid, ident.UserID, "Google identity is matched by provider_subject, not by email — login continues to work")
	assert.Equal(t, "new@example.com", accountsRepo.UserEmail(uid))
}

// extractTokenFromEmailBody pulls the token query parameter from
// the confirmation URL inside an email body. The token is the
// only artifact the frontend needs to POST at consume time.
func extractTokenFromEmailBody(t *testing.T, body string) string {
	t.Helper()
	u, err := url.Parse(extractLinkFromBody(t, body))
	require.NoError(t, err)
	token := u.Query().Get("token")
	require.NotEmpty(t, token, "no token query in link: %s", body)
	return token
}

func extractLinkFromBody(t *testing.T, body string) string {
	t.Helper()
	// The body is plain text: "Use this single-use link to confirm ...\n\n<URL>\n\nIt expires...".
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://") {
			return line
		}
	}
	t.Fatalf("no https:// link in body: %s", body)
	return ""
}
