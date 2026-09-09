package migrations_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const reviewAttemptResultRatingMigration = "20260909142062_review_attempt_result_rating_integrity.sql"

func TestReviewAttemptResultRatingMigrationAndEntSchemaAgree(t *testing.T) {
	migration, err := os.ReadFile(reviewAttemptResultRatingMigration)
	require.NoError(t, err)
	for _, invariant := range []string{
		"review_attempts_result_rating_valid",
		"result = 'skipped' AND rating IS NULL",
		"result = 'incorrect' AND rating IS NOT NULL AND rating = 'again'",
		"result = 'correct' AND rating IS NOT NULL AND rating IN ('hard', 'good', 'easy')",
		"NOT VALID",
	} {
		require.Contains(t, string(migration), invariant)
	}

	schema, err := os.ReadFile(filepath.Join("..", "ent", "schema", "reviewattempt.go"))
	require.NoError(t, err)
	for _, invariant := range []string{
		"result_rating_valid",
		"result = 'skipped' AND rating IS NULL",
		"result = 'incorrect' AND rating IS NOT NULL AND rating = 'again'",
		"result = 'correct' AND rating IS NOT NULL AND rating IN ('hard', 'good', 'easy')",
	} {
		require.Contains(t, string(schema), invariant)
	}
}

// TestReviewAttemptResultRatingConstraintAgainstRealPostgres applies the
// committed migration to the production-shaped table. It covers valid and
// invalid inserts, invalid updates, the NOT VALID catalog state, and a legacy
// invalid row that remains readable but cannot be updated after rollout.
func TestReviewAttemptResultRatingConstraintAgainstRealPostgres(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })

	schemaName := "review_result_rating_" + randomReviewSchemaSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schemaName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		require.NoError(t, err)
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schemaName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	applyMigrationsBeforeReviewResultRatingConstraint(t, db)

	ctx := context.Background()
	now := time.Now().UTC()
	userID, wordID, meaningID, userWordID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, $2, 'active', 'completed', $3, $3)`, userID, fmt.Sprintf("review-integrity-%s@example.test", userID), now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, language_code, created_at, updated_at)
		VALUES ($1, 'review', 'review', 'en', $2, $2)`, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'an assessment', 1, $3, $3)`, meaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO user_words (id, user_id, meaning_id, source, added_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'manual', $4, $4, $4)`, userWordID, userID, meaningID, now)
	require.NoError(t, err)

	insert := func(id uuid.UUID, result string, rating any) error {
		_, err := db.ExecContext(ctx, `INSERT INTO review_attempts (
			id, user_id, user_word_id, meaning_id, attempt_type, prompt_type,
			result, rating, review_step_before, review_step_after, answered_at,
			response_time_ms, was_hint_used, source, client_attempt_id, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'review', 'self_check', $5, $6, 0, 0, $7, 0, false, 'api', $8, '{}'::jsonb, $7, $7)`,
			id, userID, userWordID, meaningID, result, rating, now, id.String())
		return err
	}

	legacyID := uuid.New()
	require.NoError(t, insert(legacyID, "incorrect", "good"), "legacy corruption must be possible before the forward constraint")

	migration, err := os.ReadFile(reviewAttemptResultRatingMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must install without rejecting a legacy invalid row")

	var validated bool
	var definition string
	err = db.QueryRowContext(ctx, `SELECT convalidated, pg_get_constraintdef(oid)
		FROM pg_constraint
		WHERE conrelid = 'review_attempts'::regclass
		  AND conname = 'review_attempts_result_rating_valid'`).Scan(&validated, &definition)
	require.NoError(t, err)
	require.False(t, validated)
	require.Contains(t, definition, "result = 'skipped'")
	require.Contains(t, definition, "rating IS NULL")

	var legacyResult, legacyRating string
	err = db.QueryRowContext(ctx, `SELECT result, rating FROM review_attempts WHERE id = $1`, legacyID).Scan(&legacyResult, &legacyRating)
	require.NoError(t, err, "legacy rows remain readable after a NOT VALID rollout")
	require.Equal(t, "incorrect", legacyResult)
	require.Equal(t, "good", legacyRating)
	_, err = db.ExecContext(ctx, `UPDATE review_attempts SET response_time_ms = response_time_ms + 1 WHERE id = $1`, legacyID)
	require.Error(t, err, "every update must enforce the newly installed constraint")

	for _, valid := range []struct {
		result string
		rating any
	}{
		{result: "skipped", rating: nil},
		{result: "incorrect", rating: "again"},
		{result: "correct", rating: "hard"},
		{result: "correct", rating: "good"},
		{result: "correct", rating: "easy"},
	} {
		require.NoError(t, insert(uuid.New(), valid.result, valid.rating))
	}
	for _, invalid := range []struct {
		result string
		rating any
	}{
		{result: "skipped", rating: "easy"},
		{result: "incorrect", rating: nil},
		{result: "incorrect", rating: "hard"},
		{result: "correct", rating: nil},
		{result: "correct", rating: "again"},
	} {
		require.Error(t, insert(uuid.New(), invalid.result, invalid.rating), "invalid pair %q/%v must be rejected", invalid.result, invalid.rating)
	}

	validID := uuid.New()
	require.NoError(t, insert(validID, "correct", "good"))
	_, err = db.ExecContext(ctx, `UPDATE review_attempts SET result = 'incorrect' WHERE id = $1`, validID)
	require.Error(t, err, "an invalid update must be rejected")
	_, err = db.ExecContext(ctx, `UPDATE review_attempts SET rating = 'easy' WHERE id = $1`, validID)
	require.NoError(t, err, "a valid update must remain accepted")
}

func randomReviewSchemaSuffix(t *testing.T, bytes int) string {
	t.Helper()
	buf := make([]byte, bytes)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}

func applyMigrationsBeforeReviewResultRatingConstraint(t *testing.T, db *sql.DB) {
	t.Helper()
	paths, err := filepath.Glob("*.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	for _, path := range paths {
		if path == reviewAttemptResultRatingMigration {
			break
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoErrorf(t, err, "apply migration %s", path)
	}
}
