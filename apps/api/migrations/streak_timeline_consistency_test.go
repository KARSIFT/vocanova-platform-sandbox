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

const streakTimelineConsistencyMigration = "20260908280000_voc1440_streak_timeline_consistency.sql"

func TestStreakTimelineConsistencyMigrationCarriesDatabaseAndEntInvariants(t *testing.T) {
	body, err := os.ReadFile(streakTimelineConsistencyMigration)
	require.NoError(t, err)
	text := string(body)
	for _, invariant := range []string{
		"streak_states_positive_streak_has_date",
		"current_streak_count = 0 OR last_completed_local_date IS NOT NULL",
		"streak_states_completion_is_activity",
		"last_activity_local_date IS NOT NULL",
		"last_completed_local_date <= last_activity_local_date",
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}

	schemaBody, err := os.ReadFile("../ent/schema/streakstate.go")
	require.NoError(t, err)
	schemaText := string(schemaBody)
	assert.Contains(t, schemaText, `"positive_streak_has_date": "current_streak_count = 0 OR last_completed_local_date IS NOT NULL"`)
	assert.Contains(t, schemaText, `"completion_is_activity":   "last_completed_local_date IS NULL OR (last_activity_local_date IS NOT NULL AND last_completed_local_date <= last_activity_local_date)"`)
}

func TestStreakTimelineConsistencyAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_streak_timeline_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	for _, filename := range []string{
		"20260724210000_identity_foundation.sql",
		"20260725130001_voc030_p4_mission_tables.sql",
		"20260725130002_voc030_p4_gamification_tables.sql",
	} {
		migration, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		_, err = db.ExecContext(ctx, string(migration))
		require.NoError(t, err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	today := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)
	legacyMissingDateUser := insertStreakTimelineUser(t, ctx, db, now)
	legacyReversedUser := insertStreakTimelineUser(t, ctx, db, now)
	legacyMissingDateID := uuid.New()
	legacyReversedID := uuid.New()
	insertStreakTimelineState(t, ctx, db, legacyMissingDateID, legacyMissingDateUser, 2, 2, nil, nil, "active", now)
	insertStreakTimelineState(t, ctx, db, legacyReversedID, legacyReversedUser, 2, 2, &today, &yesterday, "active", now)

	migration, err := os.ReadFile(streakTimelineConsistencyMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve contradictory legacy states")

	zeroUser := insertStreakTimelineUser(t, ctx, db, now)
	positiveUser := insertStreakTimelineUser(t, ctx, db, now)
	brokenUser := insertStreakTimelineUser(t, ctx, db, now)
	protectedUser := insertStreakTimelineUser(t, ctx, db, now)
	zeroID := uuid.New()
	positiveID := uuid.New()
	insertStreakTimelineState(t, ctx, db, zeroID, zeroUser, 0, 0, nil, nil, "active", now)
	insertStreakTimelineState(t, ctx, db, positiveID, positiveUser, 3, 5, &today, &today, "active", now)
	insertStreakTimelineState(t, ctx, db, uuid.New(), brokenUser, 0, 7, &yesterday, &today, "broken", now)
	// A protected daily_mission_snapshot is still activity after the last
	// completed anchor. Streak state has no "protected" status of its own;
	// it remains active while the snapshot carries that historical status.
	insertStreakTimelineState(t, ctx, db, uuid.New(), protectedUser, 4, 4, &twoDaysAgo, &today, "active", now)
	_, err = db.ExecContext(ctx, `INSERT INTO daily_mission_snapshots
		(id, user_id, local_date, timezone, review_target, policy_version, status, grace_applied, created_at, updated_at)
		VALUES ($1, $2, $3, 'UTC', 5, 'test', 'protected', true, $4, $4)`, uuid.New(), protectedUser, yesterday, now)
	require.NoError(t, err)

	missingDateUser := insertStreakTimelineUser(t, ctx, db, now)
	_, err = db.ExecContext(ctx, `INSERT INTO streak_states
		(id, user_id, current_streak_count, longest_streak_count, timezone, status, created_at, updated_at)
		VALUES ($1, $2, 1, 1, 'UTC', 'active', $3, $3)`, uuid.New(), missingDateUser, now)
	requireStreakTimelineCheckViolation(t, err)

	reversedUser := insertStreakTimelineUser(t, ctx, db, now)
	_, err = db.ExecContext(ctx, `INSERT INTO streak_states
		(id, user_id, current_streak_count, longest_streak_count, last_completed_local_date, last_activity_local_date, timezone, status, created_at, updated_at)
		VALUES ($1, $2, 1, 1, $3, $4, 'UTC', 'active', $5, $5)`, uuid.New(), reversedUser, today, yesterday, now)
	requireStreakTimelineCheckViolation(t, err)

	completionWithoutActivityUser := insertStreakTimelineUser(t, ctx, db, now)
	_, err = db.ExecContext(ctx, `INSERT INTO streak_states
		(id, user_id, current_streak_count, longest_streak_count, last_completed_local_date, timezone, status, created_at, updated_at)
		VALUES ($1, $2, 0, 1, $3, 'UTC', 'broken', $4, $4)`, uuid.New(), completionWithoutActivityUser, yesterday, now)
	requireStreakTimelineCheckViolation(t, err)

	_, err = db.ExecContext(ctx, `UPDATE streak_states SET current_streak_count = 1, longest_streak_count = 1 WHERE id = $1`, zeroID)
	requireStreakTimelineCheckViolation(t, err)
	_, err = db.ExecContext(ctx, `UPDATE streak_states SET last_activity_local_date = $2 WHERE id = $1`, positiveID, yesterday)
	requireStreakTimelineCheckViolation(t, err)
	_, err = db.ExecContext(ctx, `UPDATE streak_states SET last_activity_local_date = NULL WHERE id = $1`, positiveID)
	requireStreakTimelineCheckViolation(t, err)

	for _, legacy := range []uuid.UUID{legacyMissingDateID, legacyReversedID} {
		_, err = db.ExecContext(ctx, `UPDATE streak_states SET updated_at = updated_at + interval '1 second' WHERE id = $1`, legacy)
		requireStreakTimelineCheckViolation(t, err)
	}

	for name, expected := range map[string]string{
		"streak_states_positive_streak_has_date": "current_streak_count = 0",
		"streak_states_completion_is_activity":   "last_completed_local_date <= last_activity_local_date",
	} {
		var validated bool
		var definition string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT convalidated, pg_get_constraintdef(oid)
			FROM pg_constraint
			WHERE conrelid = 'streak_states'::regclass AND conname = $1`, name).Scan(&validated, &definition))
		assert.False(t, validated, name)
		assert.Contains(t, definition, expected, name)
	}
}

func insertStreakTimelineUser(t *testing.T, ctx context.Context, db *sql.DB, now time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, status, onboarding_status, created_at, updated_at)
		VALUES ($1, 'active', 'completed', $2, $2)`, id, now)
	require.NoError(t, err)
	return id
}

func insertStreakTimelineState(t *testing.T, ctx context.Context, db *sql.DB, id, userID uuid.UUID, current, longest int, completed, activity *time.Time, status string, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO streak_states
		(id, user_id, current_streak_count, longest_streak_count, last_completed_local_date, last_activity_local_date, timezone, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'UTC', $7, $8, $8)`, id, userID, current, longest, completed, activity, status, now)
	require.NoError(t, err)
}

func requireStreakTimelineCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
