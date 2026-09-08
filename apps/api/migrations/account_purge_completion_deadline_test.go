package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const accountPurgeCompletionDeadlineMigration = "20260908210000_voc1421_account_purge_completion_deadline.sql"

func TestAccountPurgeCompletionDeadlineMigrationCarriesDatabaseInvariant(t *testing.T) {
	body, err := os.ReadFile(accountPurgeCompletionDeadlineMigration)
	require.NoError(t, err)
	text := string(body)
	assert.Contains(t, text, "account_deletion_requests_completed_after_purge_due")
	assert.Contains(t, text, "CHECK (status <> 'completed' OR completed_at >= purge_after)")
	assert.Contains(t, text, "NOT VALID")

	schemaBody, err := os.ReadFile("../ent/schema/accountdeletionrequest.go")
	require.NoError(t, err)
	assert.Contains(t, string(schemaBody), `"completed_after_purge_due": "status <> 'completed' OR completed_at >= purge_after"`)
}

func TestAccountPurgeCompletionDeadlineAgainstPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL migration test unavailable")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	schema := "vocanova_purge_deadline_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE account_deletion_requests (
			id uuid PRIMARY KEY,
			status text NOT NULL CHECK (status IN ('deactivated', 'anonymizing', 'completed')),
			requested_at timestamptz NOT NULL,
			purge_after timestamptz NOT NULL,
			completed_at timestamptz,
			marker integer NOT NULL DEFAULT 0,
			CHECK (purge_after > requested_at),
			CHECK (purge_after <= requested_at + interval '365 days'),
			CHECK (status <> 'completed' OR completed_at IS NOT NULL),
			CHECK (status = 'completed' OR completed_at IS NULL)
		)`)
	require.NoError(t, err)

	requestedAt := time.Now().UTC().Truncate(time.Microsecond)
	purgeAfter := requestedAt.Add(30 * 24 * time.Hour)
	legacyID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO account_deletion_requests (id, status, requested_at, purge_after, completed_at)
		VALUES ($1, 'completed', $2, $3, $4)`,
		legacyID, requestedAt, purgeAfter, requestedAt.Add(24*time.Hour))
	require.NoError(t, err, "the original lifecycle checks permit completion before purge_after")

	migration, err := os.ReadFile(accountPurgeCompletionDeadlineMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must preserve legacy early-completion rows")

	validRows := []struct {
		name        string
		status      string
		completedAt any
	}{
		{name: "deactivated", status: "deactivated"},
		{name: "anonymizing", status: "anonymizing"},
		{name: "completed at deadline", status: "completed", completedAt: purgeAfter},
		{name: "completed after deadline", status: "completed", completedAt: purgeAfter.Add(time.Second)},
	}
	validID := uuid.Nil
	for _, tc := range validRows {
		t.Run("valid "+tc.name, func(t *testing.T) {
			id := uuid.New()
			_, insertErr := db.ExecContext(ctx, `
				INSERT INTO account_deletion_requests (id, status, requested_at, purge_after, completed_at)
				VALUES ($1, $2, $3, $4, $5)`,
				id, tc.status, requestedAt, purgeAfter, tc.completedAt)
			require.NoError(t, insertErr)
			if tc.name == "completed after deadline" {
				validID = id
			}
		})
	}
	require.NotEqual(t, uuid.Nil, validID)

	_, err = db.ExecContext(ctx, `
		INSERT INTO account_deletion_requests (id, status, requested_at, purge_after, completed_at)
		VALUES ($1, 'completed', $2, $3, $4)`,
		uuid.New(), requestedAt, purgeAfter, purgeAfter.Add(-time.Microsecond))
	requireAccountPurgeDeadlineViolation(t, err)

	// The new deadline check complements rather than replaces the lifecycle
	// checks that bind completed_at to the completed status.
	_, err = db.ExecContext(ctx, `
		INSERT INTO account_deletion_requests (id, status, requested_at, purge_after, completed_at)
		VALUES ($1, 'completed', $2, $3, NULL)`,
		uuid.New(), requestedAt, purgeAfter)
	requireAccountPurgeDeadlineViolation(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO account_deletion_requests (id, status, requested_at, purge_after, completed_at)
		VALUES ($1, 'deactivated', $2, $3, $4)`,
		uuid.New(), requestedAt, purgeAfter, purgeAfter)
	requireAccountPurgeDeadlineViolation(t, err)

	_, err = db.ExecContext(ctx, `
		UPDATE account_deletion_requests SET completed_at = $1 WHERE id = $2`,
		purgeAfter.Add(-time.Second), validID)
	requireAccountPurgeDeadlineViolation(t, err)
	_, err = db.ExecContext(ctx, `
		UPDATE account_deletion_requests SET marker = marker + 1 WHERE id = $1`, legacyID)
	requireAccountPurgeDeadlineViolation(t, err)

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM account_deletion_requests WHERE id = $1`, legacyID).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows)

	var validated bool
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT convalidated
		FROM pg_constraint
		WHERE conrelid = 'account_deletion_requests'::regclass
		  AND conname = 'account_deletion_requests_completed_after_purge_due'`).Scan(&validated))
	assert.False(t, validated)
}

func requireAccountPurgeDeadlineViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
