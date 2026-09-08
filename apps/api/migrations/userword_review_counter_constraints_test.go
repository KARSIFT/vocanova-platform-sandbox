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

const userWordCounterConstraintsMigration = "20260908170000_user_word_review_counter_constraints.sql"

func TestUserWordReviewCounterConstraintsMigration(t *testing.T) {
	sql, err := os.ReadFile(userWordCounterConstraintsMigration)
	require.NoError(t, err)
	for _, invariant := range []string{
		"ALTER TABLE user_words",
		"user_words_review_counts_nonnegative",
		"consecutive_correct_count >= 0",
		"consecutive_incorrect_count >= 0",
		"total_review_count >= 0",
		"correct_review_count >= 0",
		"NOT VALID",
	} {
		require.Contains(t, string(sql), invariant)
	}
}

func TestUserWordEntSchemaCarriesReviewCounterInvariants(t *testing.T) {
	schema, err := os.ReadFile(filepath.Join("..", "ent", "schema", "userword.go"))
	require.NoError(t, err)
	for _, invariant := range []string{
		"review_counts_nonnegative",
		"correct_review_count_within_total",
		"consecutive_correct_count >= 0",
		"consecutive_incorrect_count >= 0",
		"total_review_count >= 0",
		"correct_review_count >= 0",
		"correct_review_count <= total_review_count",
	} {
		require.Contains(t, string(schema), invariant)
	}
}

// TestUserWordReviewCounterConstraintsAgainstRealPostgres applies every
// committed migration in an isolated schema, so the proof exercises the
// production migration instead of a hand-written lookalike table.
func TestUserWordReviewCounterConstraintsAgainstRealPostgres(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })

	schemaName := "user_word_counts_" + randomSchemaSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schemaName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		require.NoError(t, err)
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schemaName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	applyAllForwardMigrations(t, db)

	ctx := context.Background()
	userID, meaningID, wordID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, $2, 'active', 'completed', $3, $3)`, userID, fmt.Sprintf("counter-%s@example.test", userID), now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, language_code, created_at, updated_at)
		VALUES ($1, 'count', 'count', 'en', $2, $2)`, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'a total', 1, $3, $3)`, meaningID, wordID, now)
	require.NoError(t, err)

	insert := func(id uuid.UUID, consecutiveCorrect, consecutiveIncorrect, total, correct int) error {
		_, err := db.ExecContext(ctx, `INSERT INTO user_words (
			id, user_id, meaning_id, source, consecutive_correct_count,
			consecutive_incorrect_count, total_review_count, correct_review_count,
			added_at, created_at, updated_at
		) VALUES ($1, $2, $3, 'manual', $4, $5, $6, $7, $8, $8, $8)`,
			id, userID, meaningID, consecutiveCorrect, consecutiveIncorrect, total, correct, now)
		return err
	}

	userWordID := uuid.New()
	require.NoError(t, insert(userWordID, 2, 0, 3, 2), "valid production review counters must remain accepted")
	for _, invalid := range []struct {
		name   string
		column string
		value  int
	}{
		{name: "negative consecutive correct", column: "consecutive_correct_count", value: -1},
		{name: "negative consecutive incorrect", column: "consecutive_incorrect_count", value: -1},
		{name: "negative total", column: "total_review_count", value: -1},
		{name: "negative correct", column: "correct_review_count", value: -1},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, "UPDATE user_words SET "+invalid.column+" = $1 WHERE id = $2", invalid.value, userWordID)
			require.Error(t, err)
		})
	}
	_, err = db.ExecContext(ctx, `UPDATE user_words SET correct_review_count = 4 WHERE id = $1`, userWordID)
	require.Error(t, err, "the pre-existing documented upper-bound invariant must remain enforced")
	_, err = db.ExecContext(ctx, `UPDATE user_words SET total_review_count = 4, correct_review_count = 3 WHERE id = $1`, userWordID)
	require.NoError(t, err, "a valid review-counter update must remain accepted")
}

// TestUserWordReviewCounterConstraintRolloutAcceptsLegacyRows proves the
// forward migration can be deployed before a separately planned historical
// repair. PostgreSQL enforces a NOT VALID check for every later write, but it
// must not scan and reject an already-corrupt row while adding the constraint.
func TestUserWordReviewCounterConstraintRolloutAcceptsLegacyRows(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })

	schemaName := "user_word_counts_rollout_" + randomSchemaSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schemaName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		require.NoError(t, err)
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schemaName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	applyForwardMigrationsBefore(t, db, userWordCounterConstraintsMigration)

	ctx := context.Background()
	userID, meaningID, wordID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, $2, 'active', 'completed', $3, $3)`, userID, fmt.Sprintf("counter-rollout-%s@example.test", userID), now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, language_code, created_at, updated_at)
		VALUES ($1, 'count', 'count', 'en', $2, $2)`, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'a total', 1, $3, $3)`, meaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO user_words (
		id, user_id, meaning_id, source, consecutive_correct_count,
		consecutive_incorrect_count, total_review_count, correct_review_count,
		added_at, created_at, updated_at
	) VALUES ($1, $2, $3, 'manual', -1, -1, -1, -1, $4, $4, $4)`,
		uuid.New(), userID, meaningID, now)
	require.NoError(t, err, "the pre-existing upper-bound check permits this historical corruption")

	migration, err := os.ReadFile(userWordCounterConstraintsMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must allow the forward migration to install beside legacy corruption")
	var validated bool
	err = db.QueryRowContext(ctx, `SELECT convalidated FROM pg_constraint WHERE conname = 'user_words_review_counts_nonnegative'`).Scan(&validated)
	require.NoError(t, err)
	require.False(t, validated, "the migration must remain NOT VALID until an explicit historical repair and validation")
}

func randomSchemaSuffix(t *testing.T, bytes int) string {
	t.Helper()
	buf := make([]byte, bytes)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}

func applyAllForwardMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	paths, err := filepath.Glob("*.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	for _, path := range paths {
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoErrorf(t, err, "apply migration %s", path)
	}
}

func applyForwardMigrationsBefore(t *testing.T, db *sql.DB, migrationToExclude string) {
	t.Helper()
	paths, err := filepath.Glob("*.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	for _, path := range paths {
		if path == migrationToExclude {
			break
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoErrorf(t, err, "apply migration %s", path)
	}
}
