//go:build integration

package auth

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestEnsureEmailIdentityConcurrentFirstUsePostgreSQL proves that competing
// first-use magic-link completions do not expose the unique index's raw error.
// It uses random fixture IDs and removes only those fixtures, so it is safe for
// the disposable shared validation database.
func TestEnsureEmailIdentityConcurrentFirstUsePostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))
	userID := uuid.New()
	emailAddr := userID.String() + "@example.test"
	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, emailAddr, now)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM external_identities WHERE user_id = $1`, userID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	svc := NewService(NewPostgreSQLRepository(db), nil, nil, nil, nil, Config{})
	start := make(chan struct{})
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- svc.ensureEmailIdentity(ctx, userID, emailAddr)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM external_identities WHERE provider = 'email' AND provider_subject = $1 AND deleted_at IS NULL`, emailAddr).Scan(&count))
	require.Equal(t, 1, count)
	var owner uuid.UUID
	require.NoError(t, db.QueryRowContext(ctx, `SELECT user_id FROM external_identities WHERE provider = 'email' AND provider_subject = $1 AND deleted_at IS NULL`, emailAddr).Scan(&owner))
	require.Equal(t, userID, owner)
}

func TestOAuthStateSessionPreferencePostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))
	repo := NewPostgreSQLRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	for _, persistent := range []bool{false, true} {
		_, hash, err := NewTokenAndHash()
		require.NoError(t, err)
		created, err := repo.CreateOAuthState(ctx, hash, "test", "google", "https://example.test/home", now, now.Add(10*time.Minute), persistent)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = db.ExecContext(context.Background(), `DELETE FROM oauth_states WHERE id = $1`, created.ID)
		})
		stored, err := repo.GetOAuthStateByTokenHash(ctx, hash)
		require.NoError(t, err)
		require.Equal(t, persistent, stored.SessionPersistent)
		require.Equal(t, created.ID, stored.ID)
		claimed, err := repo.ConsumeOAuthState(ctx, stored.ID, now)
		require.NoError(t, err)
		require.True(t, claimed)
		claimed, err = repo.ConsumeOAuthState(ctx, stored.ID, now)
		require.NoError(t, err)
		require.False(t, claimed)
	}
	// Older deployed binaries omit the new column while a rollout is in flight.
	// Their pending states must retain the previous persistent-session behavior.
	_, hash, err := NewTokenAndHash()
	require.NoError(t, err)
	id := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO oauth_states (id, token_hash, environment, provider, app_return_url, created_at, expires_at) VALUES ($1, $2, 'test', 'google', 'https://example.test/home', $3, $4)`, id, hash, now, now.Add(10*time.Minute))
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM oauth_states WHERE id = $1`, id) })
	stored, err := repo.GetOAuthStateByTokenHash(ctx, hash)
	require.NoError(t, err)
	require.True(t, stored.SessionPersistent)
}
