package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPersonalDataExportAPI(t *testing.T) (huma.API, *auth.Service, *accounts.MemoryRepository, *auth.MemoryRepository) {
	t.Helper()
	c := &clock.Fixed{T: testNow()}
	authRepo := auth.NewMemoryRepository()
	authSvc := auth.NewService(authRepo, nil, nil, c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), auth.Config{Environment: "test", Cookie: auth.CookieConfig{Name: "session", CSRName: "csrf", Secure: false}})
	repo := accounts.NewMemoryRepository()
	svc := accounts.NewService(repo, authRepo, nil, accounts.NewMemoryIdempotencyStore(), c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), accounts.Config{})
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("test", "1"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterPersonalDataExports(api, svc, authSvc)
	return api, authSvc, repo, authRepo
}

func exportRequest(t *testing.T, uid uuid.UUID, key string, authn bool) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/personal-data-export", nil)
	req.Header.Set("Idempotency-Key", key)
	if authn {
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uid}))
	}
	return req
}

func TestPersonalDataExportHandlerRequiresAuthenticationAndCSRF(t *testing.T) {
	api, authSvc, repo, authRepo := testPersonalDataExportAPI(t)
	uid := uuid.New()
	authRepo.UpsertUser(&auth.User{ID: uid, Email: "user@example.com", Status: "active"})
	repo.SetUser(uid, "user@example.com")

	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, exportRequest(t, uid, "key", false))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, exportRequest(t, uid, "key", true))
	assert.Equal(t, http.StatusForbidden, w.Code)

	csrf, cookie := authSvc.IssueCSRFCookie()
	req := exportRequest(t, uid, "key", true)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	var body PersonalDataExportDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "UTC", body.Settings.Timezone)
	assert.Equal(t, 20, body.Settings.DailyReviewTarget)
	assert.Equal(t, "vocanova_default", body.Settings.ReviewIntervalPreset)
	assert.True(t, body.Settings.NotificationsEnabled)
}
