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

const userWordResultRatingIntegrityMigration = "20260908190000_voc1406_user_word_result_rating_integrity.sql"

func TestUserWordResultRatingIntegrityMigrationCarriesDatabaseInvariant(t *testing.T) {
	body, err := os.ReadFile(userWordResultRatingIntegrityMigration)
	require.NoError(t, err)
	text := string(body)
	assert.Contains(t, text, "user_words_last_result_rating_consistent")
	assert.Contains(t, text, "last_result IS NULL AND last_rating IS NULL")
	assert.Contains(t, text, "last_result = 'skipped' AND last_rating IS NULL")
	assert.Contains(t, text, "last_result = 'incorrect'")
	assert.Contains(t, text, "last_result = 'correct'")
	assert.Contains(t, text, "last_rating IS NOT NULL")
	assert.Contains(t, text, ") IS TRUE")
	assert.Contains(t, text, "NOT VALID")

	schemaBody, err := os.ReadFile("../ent/schema/userword.go")
	require.NoError(t, err)
	for _, invariant := range []string{
		"last_result_rating_consistent",
		"last_result IS NULL AND last_rating IS NULL",
		"last_result = 'skipped' AND last_rating IS NULL",
		"last_result = 'incorrect' AND last_rating IS NOT NULL AND last_rating = 'again'",
		"last_result = 'correct' AND last_rating IS NOT NULL AND last_rating IN ('hard', 'good', 'easy')",
		") IS TRUE",
	} {
		assert.Contains(t, string(schemaBody), invariant, "Ent metadata must mirror migration invariant %q", invariant)
	}
}

func TestUserWordResultRatingIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_user_word_pair_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE user_words (
			id uuid PRIMARY KEY,
			last_result text CHECK (last_result IN ('correct', 'incorrect', 'skipped')),
			last_rating text CHECK (last_rating IN ('again', 'hard', 'good', 'easy')),
			marker integer NOT NULL DEFAULT 0
		)`)
	require.NoError(t, err)

	legacyID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_words (id, last_result, last_rating) VALUES ($1, 'incorrect', 'good')`,
		legacyID)
	require.NoError(t, err, "the original independent enum checks permit this contradiction")

	migration, err := os.ReadFile(userWordResultRatingIntegrityMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must preserve rollout with a legacy contradiction")

	valid := []struct {
		name   string
		result any
		rating any
	}{
		{name: "no review state"},
		{name: "skipped", result: "skipped"},
		{name: "incorrect again", result: "incorrect", rating: "again"},
		{name: "correct hard", result: "correct", rating: "hard"},
		{name: "correct good", result: "correct", rating: "good"},
		{name: "correct easy", result: "correct", rating: "easy"},
	}
	for _, tc := range valid {
		t.Run("valid "+tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO user_words (id, last_result, last_rating) VALUES ($1, $2, $3)`,
				uuid.New(), tc.result, tc.rating)
			require.NoError(t, err)
		})
	}

	invalid := []struct {
		name   string
		result any
		rating any
	}{
		{name: "no result again", rating: "again"},
		{name: "no result hard", rating: "hard"},
		{name: "no result good", rating: "good"},
		{name: "no result easy", rating: "easy"},
		{name: "skipped again", result: "skipped", rating: "again"},
		{name: "skipped hard", result: "skipped", rating: "hard"},
		{name: "skipped good", result: "skipped", rating: "good"},
		{name: "skipped easy", result: "skipped", rating: "easy"},
		{name: "incorrect without rating", result: "incorrect"},
		{name: "incorrect hard", result: "incorrect", rating: "hard"},
		{name: "incorrect good", result: "incorrect", rating: "good"},
		{name: "incorrect easy", result: "incorrect", rating: "easy"},
		{name: "correct without rating", result: "correct"},
		{name: "correct again", result: "correct", rating: "again"},
	}
	for _, tc := range invalid {
		t.Run("invalid "+tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO user_words (id, last_result, last_rating) VALUES ($1, $2, $3)`,
				uuid.New(), tc.result, tc.rating)
			requireUserWordResultRatingCheckViolation(t, err)
		})
	}

	validUpdateID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_words (id, last_result, last_rating) VALUES ($1, 'incorrect', 'again')`,
		validUpdateID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE user_words SET last_rating = 'hard' WHERE id = $1`, validUpdateID)
	requireUserWordResultRatingCheckViolation(t, err)

	_, err = db.ExecContext(ctx, `UPDATE user_words SET marker = 1 WHERE id = $1`, legacyID)
	requireUserWordResultRatingCheckViolation(t, err)

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM user_words WHERE id = $1`, legacyID).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows)

	var validated bool
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT convalidated FROM pg_constraint
		WHERE conrelid = 'user_words'::regclass
		  AND conname = 'user_words_last_result_rating_consistent'`).Scan(&validated))
	assert.False(t, validated)
}

func requireUserWordResultRatingCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
