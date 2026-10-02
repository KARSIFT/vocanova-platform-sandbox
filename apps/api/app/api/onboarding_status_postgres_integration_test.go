//go:build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Uses the real migrated schema and users projection with the shared production
// adapter. The existing HTTP helper injects an authenticated requester; this
// regression tests the onboarding gate, not session minting or browser redirects.
func TestCurrentUserPostgresPreservesStatusWithoutOnboardingAnswers(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(ctx))
	repo := users.NewPostgreSQLRepository(db)
	api, _, _ := testOnboardingAPI(t)
	SetOnboardingStatusLookup(newOnboardingStatusLookup(repo))

	for _, status := range []string{"completed", "in_progress", "not_started"} {
		t.Run(status, func(t *testing.T) {
			id := uuid.New()
			now := time.Now().UTC()
			_, err := db.ExecContext(ctx, `INSERT INTO users (id,email,status,onboarding_status,created_at,updated_at) VALUES ($1,$2,'active',$3,$4,$4)`, id, id.String()+"@onboarding.vocanova.invalid", status, now)
			require.NoError(t, err)
			t.Cleanup(func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cleanupCancel()
				_, err := db.ExecContext(cleanupCtx, "DELETE FROM users WHERE id=$1", id)
				require.NoError(t, err)
			})
			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM user_onboarding_profiles WHERE user_id=$1", id).Scan(&count))
			require.Zero(t, count, "grandfathered state must have no fabricated answers")
			profile, err := repo.GetOnboarding(ctx, id)
			require.ErrorIs(t, err, users.ErrOnboardingNotFound)
			require.NotNil(t, profile)
			require.Equal(t, status, profile.Status)
			require.Nil(t, profile.CompletedAt)

			response := httptest.NewRecorder()
			api.Adapter().ServeHTTP(response, onbRequesterRequest(t, http.MethodGet, "/api/v1/me", "", id))
			require.Equal(t, http.StatusOK, response.Code)
			var body CurrentUser
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			require.Equal(t, id.String(), body.ID)
			require.Equal(t, status, body.OnboardingStatus)
		})
	}
}
