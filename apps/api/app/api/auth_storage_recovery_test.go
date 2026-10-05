package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type authValidationReadRepository struct {
	*auth.MemoryRepository
	sessionErr error
	userErr    error
}

func (r *authValidationReadRepository) GetSessionByTokenHash(ctx context.Context, hash []byte) (*auth.Session, error) {
	if r.sessionErr != nil {
		return nil, r.sessionErr
	}
	return r.MemoryRepository.GetSessionByTokenHash(ctx, hash)
}

func (r *authValidationReadRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	if r.userErr != nil {
		return nil, r.userErr
	}
	return r.MemoryRepository.GetUserByID(ctx, id)
}

func authStorageFixture(t *testing.T) (huma.API, *authValidationReadRepository, *auth.Service, *auth.User, *auth.Session, string, *clock.Fixed) {
	t.Helper()
	c := &clock.Fixed{T: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)}
	repo := &authValidationReadRepository{MemoryRepository: auth.NewMemoryRepository()}
	user, err := repo.CreateUser(t.Context(), "storage-test@example.com", nil)
	require.NoError(t, err)
	// Local synthetic credential only. No account or provider is involved.
	raw := bytes.Repeat([]byte{0x61}, 32)
	hash := sha256.Sum256(raw)
	token := base64.URLEncoding.EncodeToString(raw)
	session, err := repo.CreateSession(t.Context(), user.ID, hash[:], c.Now(), c.Now().Add(time.Hour))
	require.NoError(t, err)
	svc := auth.NewService(repo, &email.Fake{}, nil, c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), auth.Config{
		BaseURL: "https://test.example.com", Cookie: auth.CookieConfig{Name: "vocanova_session", CSRName: "vocanova_csrf"},
	})
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("auth-storage-test", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(svc))
	RegisterContract(api)
	RegisterAuth(api, svc)
	return api, repo, svc, user, session, token, c
}

func sendAuthStorageRequest(api huma.API, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "vocanova_session", Value: token})
	}
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	return w
}

func TestAuthMiddlewareStorageFailureIsRecoverableWithoutNewSession(t *testing.T) {
	for _, stage := range []string{"session", "user"} {
		t.Run(stage, func(t *testing.T) {
			api, repo, _, user, _, token, _ := authStorageFixture(t)
			healthy := sendAuthStorageRequest(api, http.MethodGet, "/api/v1/me", "", token)
			require.Equal(t, http.StatusOK, healthy.Code)
			require.Contains(t, healthy.Body.String(), user.ID.String())
			failure := errors.New("private synthetic database host and query detail")
			if stage == "session" {
				repo.sessionErr = failure
			} else {
				repo.userErr = failure
			}
			unavailable := sendAuthStorageRequest(api, http.MethodGet, "/api/v1/me", "", token)
			require.Equal(t, http.StatusServiceUnavailable, unavailable.Code)
			require.Contains(t, unavailable.Body.String(), "temporarily unavailable")
			require.NotContains(t, unavailable.Body.String(), failure.Error())
			require.NotContains(t, unavailable.Body.String(), token)
			require.Empty(t, unavailable.Header().Values("Set-Cookie"))
			require.Empty(t, unavailable.Header().Get("Location"))
			require.Equal(t, "no-store", unavailable.Header().Get("Cache-Control"))
			repo.sessionErr, repo.userErr = nil, nil
			recovered := sendAuthStorageRequest(api, http.MethodGet, "/api/v1/me", "", token)
			require.Equal(t, http.StatusOK, recovered.Code)
			require.Contains(t, recovered.Body.String(), user.ID.String())
		})
	}
}

func TestAuthMiddlewareMissingInvalidExpiredRevokedOrDisabledStillReturns401(t *testing.T) {
	for _, scenario := range []string{"missing-cookie", "malformed", "unknown-session", "missing-session", "missing-user", "expired", "revoked", "disabled"} {
		t.Run(scenario, func(t *testing.T) {
			api, repo, _, user, session, token, c := authStorageFixture(t)
			switch scenario {
			case "missing-cookie":
				token = ""
			case "malformed":
				token = "invalid-token"
			case "unknown-session":
				token = base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{0x62}, 32))
			case "missing-session":
				repo.sessionErr = fmt.Errorf("storage lookup: %w", auth.ErrSessionNotFound)
			case "missing-user":
				repo.userErr = fmt.Errorf("storage lookup: %w", auth.ErrUserNotFound)
			case "expired":
				c.Advance(2 * time.Hour)
			case "revoked":
				require.NoError(t, repo.RevokeSession(t.Context(), session.ID, c.Now()))
			case "disabled":
				require.NoError(t, repo.SetUserStatus(user.ID, "disabled"))
			}
			response := sendAuthStorageRequest(api, http.MethodGet, "/api/v1/me", "", token)
			require.Equal(t, http.StatusUnauthorized, response.Code)
			require.NotContains(t, response.Body.String(), user.Email)
			require.Empty(t, response.Header().Values("Set-Cookie"))
		})
	}
}

func TestAuthMiddlewareDoesNotBlockPublicSignInWithOldCookie(t *testing.T) {
	for _, scenario := range []string{"malformed-cookie", "storage-failure"} {
		t.Run(scenario, func(t *testing.T) {
			api, repo, _, _, _, token, _ := authStorageFixture(t)
			if scenario == "malformed-cookie" {
				token = "invalid-cookie"
			} else {
				repo.sessionErr = errors.New("synthetic session storage failure")
			}
			response := sendAuthStorageRequest(api, http.MethodPost, "/api/v1/auth/magic-links", `{"email":"new-sign-in@example.com"}`, token)
			require.Equal(t, http.StatusNoContent, response.Code)
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		})
	}
}

func TestRequireAuthReturns503BeforeCallingProtectedHandler(t *testing.T) {
	api, repo, _, _, _, token, _ := authStorageFixture(t)
	repo.sessionErr = errors.New("synthetic storage failure")
	called := false
	huma.Register(api, huma.Operation{OperationID: "TestProtectedStorageFailure", Method: http.MethodGet, Path: "/protected-storage-test", Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(context.Context, *struct{}) (*struct{ Body string }, error) {
		called = true
		return &struct{ Body string }{Body: "private data"}, nil
	})
	response := sendAuthStorageRequest(api, http.MethodGet, "/protected-storage-test", "", token)
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.False(t, called)
	require.NotContains(t, response.Body.String(), "private data")
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}
