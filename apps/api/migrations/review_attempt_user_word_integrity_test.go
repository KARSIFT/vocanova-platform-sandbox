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

const reviewAttemptUserWordIntegrityMigration = "20260908021500_review_attempt_user_word_integrity.sql"

func TestReviewAttemptUserWordIntegrityMigrationCarriesDatabaseInvariants(t *testing.T) {
	sqlBytes, err := os.ReadFile(reviewAttemptUserWordIntegrityMigration)
	require.NoError(t, err)
	text := string(sqlBytes)

	for _, invariant := range []string{
		"CREATE UNIQUE INDEX user_words_id_user_id_meaning_id_key",
		"ON user_words (id, user_id, meaning_id)",
		"ADD CONSTRAINT review_attempts_user_word_owner_meaning_fk",
		"FOREIGN KEY (user_word_id, user_id, meaning_id)",
		"REFERENCES user_words (id, user_id, meaning_id)",
		"ON DELETE RESTRICT",
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}
	assert.NotContains(t, text, "ON DELETE CASCADE")
}

func TestReviewAttemptUserWordIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_review_linkage_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
			user_id uuid NOT NULL,
			meaning_id uuid NOT NULL,
			status text NOT NULL DEFAULT 'new',
			deleted_at timestamptz
		);
		CREATE TABLE review_attempts (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			user_word_id uuid NOT NULL REFERENCES user_words(id) ON DELETE RESTRICT,
			meaning_id uuid NOT NULL
		)`)
	require.NoError(t, err)

	userA, userB := uuid.New(), uuid.New()
	meaningA, meaningB := uuid.New(), uuid.New()
	activeWordID, deletedWordID := uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_words (id, user_id, meaning_id) VALUES ($1, $2, $3)`,
		activeWordID, userA, meaningA,
	)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO user_words (id, user_id, meaning_id, status, deleted_at)
		 VALUES ($1, $2, $3, 'archived', $4)`,
		deletedWordID, userA, meaningA, time.Now().UTC(),
	)
	require.NoError(t, err)

	// A legacy inconsistent row must not block the forward migration.
	legacyID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, user_id, user_word_id, meaning_id)
		 VALUES ($1, $2, $3, $4)`,
		legacyID, userB, activeWordID, meaningB,
	)
	require.NoError(t, err)

	migration, err := os.ReadFile(reviewAttemptUserWordIntegrityMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must preserve rollout with legacy mismatches")

	// Valid history remains insertable even when the saved word is archived or
	// soft-deleted; the relationship protects identity, not active-state policy.
	validID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, user_id, user_word_id, meaning_id)
		 VALUES ($1, $2, $3, $4)`,
		validID, userA, deletedWordID, meaningA,
	)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, user_id, user_word_id, meaning_id)
		 VALUES ($1, $2, $3, $4)`,
		uuid.New(), userB, activeWordID, meaningA,
	)
	requireForeignKeyViolation(t, err, "cross-user history")

	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, user_id, user_word_id, meaning_id)
		 VALUES ($1, $2, $3, $4)`,
		uuid.New(), userA, activeWordID, meaningB,
	)
	requireForeignKeyViolation(t, err, "cross-meaning history")

	_, err = db.ExecContext(ctx,
		`UPDATE review_attempts SET user_id = $1 WHERE id = $2`, userB, validID)
	requireForeignKeyViolation(t, err, "cross-user update")

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM review_attempts WHERE id = $1`, legacyID).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows, "the NOT VALID rollout must preserve legacy rows")
}

func requireForeignKeyViolation(t *testing.T, err error, operation string) {
	t.Helper()
	require.Error(t, err, operation+" must be rejected")
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23503"), pqErr.Code)
}
