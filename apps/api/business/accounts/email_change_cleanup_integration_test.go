//go:build integration

package accounts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepositoryCleanupExpiredEmailChangeLinksPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL cleanup test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))
	schema := "email_change_cleanup_" + uuid.NewString()[:8]
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+schema)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE users (id uuid PRIMARY KEY);
		CREATE TABLE email_change_links (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL REFERENCES users(id),
			new_email text NOT NULL,
			token_hash bytea NOT NULL UNIQUE,
			environment text NOT NULL,
			created_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL,
			consumed_at timestamptz,
			revoked_at timestamptz
		)`)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id) VALUES ($1)`, userID)
	require.NoError(t, err)

	insert := func(label string, expiresAt time.Time, consumedAt, revokedAt *time.Time) uuid.UUID {
		t.Helper()
		hash := sha256.Sum256([]byte(label + userID.String()))
		id := uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at, consumed_at, revoked_at) VALUES ($1, $2, $3, $4, 'test', $5, $6, $7, $8)`, id, userID, fmt.Sprintf("%s@example.test", label), hash[:], now.Add(-time.Minute), expiresAt, consumedAt, revokedAt)
		require.NoError(t, err)
		return id
	}
	expiredID := insert("expired", now, nil, nil)
	consumedID := insert("consumed", now.Add(time.Minute), &now, nil)
	futureRevocation := now.Add(time.Minute)
	revokedID := insert("revoked", now.Add(2*time.Minute), nil, &futureRevocation)
	activeID := insert("active", now.Add(time.Minute), nil, nil)

	deleted, err := NewPostgreSQLRepository(db).CleanupExpiredEmailChangeLinks(ctx, now, 2)
	require.NoError(t, err)
	require.Equal(t, int64(2), deleted, "one pass must honor the batch limit")
	deleted, err = NewPostgreSQLRepository(db).CleanupExpiredEmailChangeLinks(ctx, now, 2)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted, "a later pass removes the remaining inactive row")

	var remaining []uuid.UUID
	rows, err := db.QueryContext(ctx, `SELECT id FROM email_change_links WHERE user_id = $1`, userID)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		require.NoError(t, rows.Scan(&id))
		remaining = append(remaining, id)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []uuid.UUID{activeID}, remaining)
	require.NotContains(t, remaining, expiredID)
	require.NotContains(t, remaining, consumedID)
	require.NotContains(t, remaining, revokedID)
}
