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

const missionCompletedAtIntegrityMigration = "20260908024500_mission_completed_at_integrity.sql"

func TestMissionCompletedAtIntegrityMigrationCarriesDatabaseInvariant(t *testing.T) {
	body, err := os.ReadFile(missionCompletedAtIntegrityMigration)
	require.NoError(t, err)
	text := string(body)
	assert.Contains(t, text, "daily_mission_snapshots_completed_at_only_on_done")
	assert.Contains(t, text, "CHECK (status = 'completed' OR completed_at IS NULL)")
	assert.Contains(t, text, "NOT VALID")
}

func TestMissionCompletedAtIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_mission_completion_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE daily_mission_snapshots (
			id uuid PRIMARY KEY,
			status text NOT NULL CHECK (status IN ('open', 'completed', 'missed', 'protected')),
			completed_at timestamptz,
			CONSTRAINT existing_completed_at_required
			  CHECK (status <> 'completed' OR completed_at IS NOT NULL)
		)`)
	require.NoError(t, err)

	now := time.Now().UTC()
	legacyCompleteByStatusID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO daily_mission_snapshots (id, status, completed_at)
		 VALUES ($1, 'open', $2)`, legacyCompleteByStatusID, now)
	require.NoError(t, err, "the original one-way check permits this legacy contradiction")
	legacyClearTimestampID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO daily_mission_snapshots (id, status, completed_at)
		 VALUES ($1, 'missed', $2)`, legacyClearTimestampID, now)
	require.NoError(t, err, "the original one-way check permits this legacy contradiction")

	migration, err := os.ReadFile(missionCompletedAtIntegrityMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must preserve rollout with a legacy contradiction")

	for _, tc := range []struct {
		status      string
		completedAt any
	}{
		{status: "open"},
		{status: "missed"},
		{status: "protected"},
		{status: "completed", completedAt: now},
	} {
		t.Run("valid "+tc.status, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO daily_mission_snapshots (id, status, completed_at)
				 VALUES ($1, $2, $3)`, uuid.New(), tc.status, tc.completedAt)
			require.NoError(t, err)
		})
	}

	for _, status := range []string{"open", "missed", "protected"} {
		t.Run("timestamp with "+status, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO daily_mission_snapshots (id, status, completed_at)
				 VALUES ($1, $2, $3)`, uuid.New(), status, now)
			requireMissionCompletionCheckViolation(t, err)
		})
	}

	// The constraint must apply to mutations as well as inserts: application
	// code writes the completion pair during a status transition, while an
	// accidental later timestamp on any non-completed state must be rejected.
	for _, status := range []string{"open", "missed", "protected"} {
		t.Run("updating timestamp on "+status, func(t *testing.T) {
			id := uuid.New()
			_, err := db.ExecContext(ctx,
				`INSERT INTO daily_mission_snapshots (id, status, completed_at)
				 VALUES ($1, $2, NULL)`, id, status)
			require.NoError(t, err)

			_, err = db.ExecContext(ctx,
				`UPDATE daily_mission_snapshots SET completed_at = $2 WHERE id = $1`, id, now)
			requireMissionCompletionCheckViolation(t, err)
		})
	}
	t.Run("updating the completion pair", func(t *testing.T) {
		id := uuid.New()
		_, err := db.ExecContext(ctx,
			`INSERT INTO daily_mission_snapshots (id, status, completed_at)
			 VALUES ($1, 'open', NULL)`, id)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`UPDATE daily_mission_snapshots
			 SET status = 'completed', completed_at = $2
			 WHERE id = $1`, id, now)
		require.NoError(t, err)
	})
	t.Run("removing completed status without clearing timestamp", func(t *testing.T) {
		id := uuid.New()
		_, err := db.ExecContext(ctx,
			`INSERT INTO daily_mission_snapshots (id, status, completed_at)
			 VALUES ($1, 'completed', $2)`, id, now)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`UPDATE daily_mission_snapshots SET status = 'missed' WHERE id = $1`, id)
		requireMissionCompletionCheckViolation(t, err)
	})
	t.Run("clearing timestamp from completed", func(t *testing.T) {
		id := uuid.New()
		_, err := db.ExecContext(ctx,
			`INSERT INTO daily_mission_snapshots (id, status, completed_at)
			 VALUES ($1, 'completed', $2)`, id, now)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`UPDATE daily_mission_snapshots SET completed_at = NULL WHERE id = $1`, id)
		requireMissionCompletionCheckViolation(t, err)
	})
	t.Run("completed without timestamp", func(t *testing.T) {
		_, err := db.ExecContext(ctx,
			`INSERT INTO daily_mission_snapshots (id, status, completed_at)
			 VALUES ($1, 'completed', NULL)`, uuid.New())
		requireMissionCompletionCheckViolation(t, err)
	})
	t.Run("legacy contradictions can be repaired through either side of the pair", func(t *testing.T) {
		_, err := db.ExecContext(ctx,
			`UPDATE daily_mission_snapshots SET status = 'completed' WHERE id = $1`, legacyCompleteByStatusID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`UPDATE daily_mission_snapshots SET completed_at = NULL WHERE id = $1`, legacyClearTimestampID)
		require.NoError(t, err)
	})

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM daily_mission_snapshots WHERE id IN ($1, $2)`,
		legacyCompleteByStatusID, legacyClearTimestampID).Scan(&legacyRows))
	assert.Equal(t, 2, legacyRows)

	var validated bool
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT convalidated FROM pg_constraint
		WHERE conrelid = 'daily_mission_snapshots'::regclass
		  AND conname = 'daily_mission_snapshots_completed_at_only_on_done'`).Scan(&validated))
	assert.False(t, validated)
}

func requireMissionCompletionCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
