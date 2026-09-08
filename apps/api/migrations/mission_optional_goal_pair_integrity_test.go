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

const missionOptionalGoalPairMigration = "20260908023500_mission_optional_goal_pair_integrity.sql"

func TestMissionOptionalGoalPairMigrationCarriesDatabaseInvariants(t *testing.T) {
	body, err := os.ReadFile(missionOptionalGoalPairMigration)
	require.NoError(t, err)
	text := string(body)

	for _, invariant := range []string{
		"daily_mission_snapshots_new_word_goal_pair_consistent",
		"new_word_target IS NULL AND new_words_completed IS NULL",
		"new_word_target IS NOT NULL",
		"new_words_completed IS NOT NULL",
		"new_words_completed >= 0",
		"new_words_completed <= new_word_target",
		"daily_mission_snapshots_sentence_goal_pair_consistent",
		"sentence_practice_target IS NULL AND sentence_practices_completed IS NULL",
		"sentence_practice_target IS NOT NULL",
		"sentence_practices_completed IS NOT NULL",
		"sentence_practices_completed >= 0",
		"sentence_practices_completed <= sentence_practice_target",
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}
}

func TestMissionOptionalGoalPairIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_goal_pairs_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
			new_word_target integer,
			new_words_completed integer,
			sentence_practice_target integer,
			sentence_practices_completed integer,
			CONSTRAINT existing_new_target CHECK (new_word_target IS NULL OR new_word_target BETWEEN 1 AND 100),
			CONSTRAINT existing_new_completed CHECK (new_word_target IS NULL OR (new_words_completed >= 0 AND new_words_completed <= new_word_target)),
			CONSTRAINT existing_sentence_target CHECK (sentence_practice_target IS NULL OR sentence_practice_target BETWEEN 1 AND 100),
			CONSTRAINT existing_sentence_completed CHECK (sentence_practice_target IS NULL OR (sentence_practices_completed >= 0 AND sentence_practices_completed <= sentence_practice_target))
		)`)
	require.NoError(t, err)

	legacyID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO daily_mission_snapshots
		 (id, new_word_target, new_words_completed, sentence_practice_target, sentence_practices_completed)
		 VALUES ($1, NULL, 999, 4, NULL)`, legacyID)
	require.NoError(t, err, "the old checks demonstrate both NULL loopholes")

	migration, err := os.ReadFile(missionOptionalGoalPairMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must preserve rollout with legacy rows")

	for _, tc := range []struct {
		name              string
		newTarget         any
		newCompleted      any
		sentenceTarget    any
		sentenceCompleted any
	}{
		{name: "both optional goals absent"},
		{name: "zero progress", newTarget: 5, newCompleted: 0, sentenceTarget: 3, sentenceCompleted: 0},
		{name: "targets completed", newTarget: 5, newCompleted: 5, sentenceTarget: 3, sentenceCompleted: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO daily_mission_snapshots
				 (id, new_word_target, new_words_completed, sentence_practice_target, sentence_practices_completed)
				 VALUES ($1, $2, $3, $4, $5)`,
				uuid.New(), tc.newTarget, tc.newCompleted, tc.sentenceTarget, tc.sentenceCompleted,
			)
			require.NoError(t, err)
		})
	}

	for _, tc := range []struct {
		name              string
		newTarget         any
		newCompleted      any
		sentenceTarget    any
		sentenceCompleted any
	}{
		{name: "new progress without target", newCompleted: 0},
		{name: "new target without progress", newTarget: 5},
		{name: "new progress negative", newTarget: 5, newCompleted: -1},
		{name: "new progress above target", newTarget: 5, newCompleted: 6},
		{name: "sentence progress without target", sentenceCompleted: 0},
		{name: "sentence target without progress", sentenceTarget: 3},
		{name: "sentence progress negative", sentenceTarget: 3, sentenceCompleted: -1},
		{name: "sentence progress above target", sentenceTarget: 3, sentenceCompleted: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO daily_mission_snapshots
				 (id, new_word_target, new_words_completed, sentence_practice_target, sentence_practices_completed)
				 VALUES ($1, $2, $3, $4, $5)`,
				uuid.New(), tc.newTarget, tc.newCompleted, tc.sentenceTarget, tc.sentenceCompleted,
			)
			requireMissionGoalCheckViolation(t, err, tc.name)
		})
	}

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM daily_mission_snapshots WHERE id = $1`, legacyID).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows)

	var unvalidated int
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_constraint
		WHERE conrelid = 'daily_mission_snapshots'::regclass
		  AND conname IN (
		    'daily_mission_snapshots_new_word_goal_pair_consistent',
		    'daily_mission_snapshots_sentence_goal_pair_consistent'
		  )
		  AND NOT convalidated`).Scan(&unvalidated))
	assert.Equal(t, 2, unvalidated)
}

func requireMissionGoalCheckViolation(t *testing.T, err error, operation string) {
	t.Helper()
	require.Error(t, err, operation+" must be rejected")
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
