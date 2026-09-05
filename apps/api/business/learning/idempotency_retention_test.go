package learning

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryIdempotencyStoreExpiresKeysAtTwentyFourHours(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	store := NewMemoryIdempotencyStore()
	store.now = func() time.Time { return now }
	userID := uuid.New()

	require.NoError(t, store.Record(t.Context(), userID, "save-user-word", "key", "first"))
	status, err := store.Check(t.Context(), userID, "save-user-word", "key", "first")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyMatch, status)
	status, err = store.Check(t.Context(), userID, "save-user-word", "key", "different")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyConflict, status, "active keys retain conflict behavior")

	now = now.Add(idempotencyRetention)
	status, err = store.Check(t.Context(), userID, "save-user-word", "key", "different")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyAbsent, status, "the exact 24-hour boundary is expired")

	require.NoError(t, store.Record(t.Context(), userID, "save-user-word", "key", "second"))
	status, err = store.Check(t.Context(), userID, "save-user-word", "key", "second")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyMatch, status)
}

func TestMemoryIdempotencyStoreScopesKeysByUserAndOperation(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	userA, userB := uuid.New(), uuid.New()
	require.NoError(t, store.Record(t.Context(), userA, "save-user-word", "shared", "first"))

	for _, tc := range []struct {
		name      string
		userID    uuid.UUID
		operation string
	}{
		{name: "other user", userID: userB, operation: "save-user-word"},
		{name: "other operation", userID: userA, operation: "submit-review"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, err := store.Check(t.Context(), tc.userID, tc.operation, "shared", "different")
			require.NoError(t, err)
			assert.Equal(t, IdempotencyAbsent, status)
		})
	}
}

func TestPostgreSQLIdempotencyStoreOnlyMatchesActiveKeysAndReplacesExpiredOnRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	store := NewPostgreSQLIdempotencyStore(db)
	store.now = func() time.Time { return now }
	userID := uuid.New()
	cutoff := now.Add(-idempotencyRetention)

	mock.ExpectQuery("SELECT fingerprint FROM idempotency_keys").
		WithArgs(userID, "save-user-word", "key", cutoff).
		WillReturnRows(sqlmock.NewRows([]string{"fingerprint"}).AddRow("first"))
	status, err := store.Check(t.Context(), userID, "save-user-word", "key", "first")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyMatch, status)

	mock.ExpectQuery("SELECT fingerprint FROM idempotency_keys").
		WithArgs(userID, "save-user-word", "expired", cutoff).
		WillReturnError(sql.ErrNoRows)
	status, err = store.Check(t.Context(), userID, "save-user-word", "expired", "different")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyAbsent, status)

	mock.ExpectExec("INSERT INTO idempotency_keys").
		WithArgs(sqlmock.AnyArg(), userID, "save-user-word", "expired", "second", now, cutoff).
		WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, store.Record(t.Context(), userID, "save-user-word", "expired", "second"))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLIdempotencyStoreRetentionPostgreSQL validates the conflict
// replacement against a disposable database. It uses a connection-local table
// and never reads or writes application tables.
func TestPostgreSQLIdempotencyStoreRetentionPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL retention test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		CREATE TEMP TABLE idempotency_keys (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL,
			key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL,
			UNIQUE (user_id, operation, key)
		)`)
	require.NoError(t, err)

	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	store := NewPostgreSQLIdempotencyStore(db)
	store.now = func() time.Time { return now }
	userA, userB := uuid.New(), uuid.New()
	require.NoError(t, store.Record(ctx, userA, "save-user-word", "shared", "old"))

	status, err := store.Check(ctx, userA, "save-user-word", "shared", "old")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyMatch, status)
	status, err = store.Check(ctx, userA, "save-user-word", "shared", "different")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyConflict, status, "active keys retain conflict behavior")
	for _, tc := range []struct {
		userID    uuid.UUID
		operation string
	}{
		{userID: userB, operation: "save-user-word"},
		{userID: userA, operation: "submit-review"},
	} {
		status, err = store.Check(ctx, tc.userID, tc.operation, "shared", "old")
		require.NoError(t, err)
		assert.Equal(t, IdempotencyAbsent, status)
	}

	now = now.Add(idempotencyRetention)
	status, err = store.Check(ctx, userA, "save-user-word", "shared", "old")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyAbsent, status, "the exact retention boundary is expired")
	require.NoError(t, store.Record(ctx, userA, "save-user-word", "shared", "new"))

	status, err = store.Check(ctx, userA, "save-user-word", "shared", "new")
	require.NoError(t, err)
	assert.Equal(t, IdempotencyMatch, status)
	var fingerprint string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT fingerprint FROM idempotency_keys WHERE user_id = $1`, userA).Scan(&fingerprint))
	assert.Equal(t, "new", fingerprint)
}
