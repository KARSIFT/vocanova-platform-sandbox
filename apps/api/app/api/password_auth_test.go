package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/password"
	"github.com/stretchr/testify/require"
)

func TestPasswordEndpointsDisabledAndJSONOnly(t *testing.T) {
	api, authSvc, _, mail, clk := testAuthAPI(t)
	svc := password.NewService(nil, mail, clk, auth.NewFixedWindowRateLimiter(clk, time.Hour, 100), "https://test.example.com", "test", time.Hour, func() bool { return false }, nil, nil)
	RegisterPasswordAuth(api, svc, authSvc)
	for _, tc := range []struct{ path, body string }{
		{"signups", `{"email":"learner@example.com","password":"a sufficiently long password"}`},
		{"signups/verify", `{"token":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}`},
		{"reset-requests", `{"email":"learner@example.com"}`},
		{"resets", `{"token":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","password":"a sufficiently long password"}`},
		{"login", `{"email":"learner@example.com","password":"a sufficiently long password"}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			for _, contentType := range []string{"application/json", "text/plain", "application/x-www-form-urlencoded", ""} {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/"+tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", contentType)
				w := httptest.NewRecorder()
				api.Adapter().ServeHTTP(w, req)
				expected := http.StatusUnsupportedMediaType
				if contentType == "application/json" {
					expected = http.StatusServiceUnavailable
				}
				require.Equal(t, expected, w.Code, w.Body.String())
				require.NotContains(t, w.Body.String(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
				require.NotContains(t, w.Body.String(), "a sufficiently long password")
				require.Empty(t, w.Result().Cookies())
				require.Contains(t, w.Header().Get("Cache-Control"), "no-store")
			}
		})
	}
}
