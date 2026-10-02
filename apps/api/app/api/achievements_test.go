package api

import (
	"context"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/achievements"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

type achievementAPIRepo struct {
	owner uuid.UUID
	calls int
}

func (r *achievementAPIRepo) Metrics(_ context.Context, u uuid.UUID) (map[string]achievements.Metric, error) {
	r.calls++
	if u != r.owner {
		return map[string]achievements.Metric{}, nil
	}
	return map[string]achievements.Metric{"writing": {Count: 1, Thresholds: map[int]time.Time{1: time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)}}}, nil
}
func TestAchievementsAPIAuthAndRequesterIsolation(t *testing.T) {
	repo := &achievementAPIRepo{owner: uuid.New()}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("achievements", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterAchievements(api, achievements.NewService(repo))
	for _, u := range []uuid.UUID{uuid.Nil, repo.owner, uuid.New()} {
		req := httptest.NewRequest("GET", "/api/v1/achievements", nil)
		if u != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: u}))
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		if u == uuid.Nil {
			require.Equal(t, 401, w.Code)
			require.Zero(t, repo.calls)
		} else {
			require.Equal(t, 200, w.Code, w.Body.String())
			if u == repo.owner {
				require.Contains(t, w.Body.String(), `"earnedAt":"2026-10-02T12:00:00Z"`)
			} else {
				require.NotContains(t, w.Body.String(), `"earnedAt"`)
			}
		}
	}
}
