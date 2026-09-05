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
