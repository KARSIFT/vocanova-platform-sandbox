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

const reviewAttemptProvenanceEnumsMigration = "20260908040000_review_attempt_provenance_enums.sql"

func TestReviewAttemptProvenanceEnumsMigrationCarriesDatabaseInvariants(t *testing.T) {
	sqlBytes, err := os.ReadFile(reviewAttemptProvenanceEnumsMigration)
	require.NoError(t, err)
	text := string(sqlBytes)

	for _, invariant := range []string{
		"ADD CONSTRAINT review_attempts_attempt_type_documented",
		"attempt_type IN ('review', 'practice', 'placement', 'mission')",
		"ADD CONSTRAINT review_attempts_source_documented",
		"source IN ('daily_review', 'word_detail', 'journey_practice', 'manual_practice')",
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}
}

func TestReviewAttemptProvenanceEnumsAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_review_provenance_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE review_attempts (
			id uuid PRIMARY KEY,
			attempt_type text NOT NULL CHECK (attempt_type <> ''),
			source text NOT NULL CHECK (source <> ''),
			metadata text
		)`)
	require.NoError(t, err)

	// Legacy values were permitted by the original independent non-empty
	// checks. They must not block this forward migration.
	legacyID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, attempt_type, source) VALUES ($1, 'legacy_import', 'legacy_source')`,
		legacyID,
	)
	require.NoError(t, err)
	migration, err := os.ReadFile(reviewAttemptProvenanceEnumsMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy provenance")

	for _, constraint := range []string{
		"review_attempts_attempt_type_documented",
		"review_attempts_source_documented",
	} {
		var validated bool
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT convalidated FROM pg_constraint
			WHERE conname = $1 AND conrelid = 'review_attempts'::regclass`, constraint,
		).Scan(&validated))
		assert.False(t, validated, "%s must remain NOT VALID for rollout", constraint)
	}

	for _, attemptType := range []string{"review", "practice", "placement", "mission"} {
		for _, source := range []string{"daily_review", "word_detail", "journey_practice", "manual_practice"} {
			_, err = db.ExecContext(ctx,
				`INSERT INTO review_attempts (id, attempt_type, source) VALUES ($1, $2, $3)`,
				uuid.New(), attemptType, source,
			)
			require.NoErrorf(t, err, "documented provenance %q/%q must be accepted", attemptType, source)
		}
	}

	for _, tc := range []struct {
		name, attemptType, source string
	}{
		{"invalid attempt type", "legacy_import", "daily_review"},
		{"invalid source", "review", "review_session"},
		{"empty attempt type", "", "daily_review"},
		{"empty source", "review", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx,
				`INSERT INTO review_attempts (id, attempt_type, source) VALUES ($1, $2, $3)`,
				uuid.New(), tc.attemptType, tc.source,
			)
			requireCheckViolation(t, err, tc.name)
		})
	}

	// The pre-existing NOT NULL columns remain the NULL boundary; PostgreSQL
	// CHECK alone would otherwise treat a NULL expression as unknown/accepted.
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, attempt_type, source) VALUES ($1, NULL, 'daily_review')`, uuid.New())
	requireNotNullViolation(t, err, "NULL attempt type")
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, attempt_type, source) VALUES ($1, 'review', NULL)`, uuid.New())
	requireNotNullViolation(t, err, "NULL source")

	validID := uuid.New()
	_, err = db.ExecContext(ctx,
		`INSERT INTO review_attempts (id, attempt_type, source, metadata) VALUES ($1, 'review', 'daily_review', 'valid')`, validID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE review_attempts SET attempt_type = 'unknown' WHERE id = $1`, validID)
	requireCheckViolation(t, err, "invalid attempt-type update")
	_, err = db.ExecContext(ctx, `UPDATE review_attempts SET source = 'review_session' WHERE id = $1`, validID)
	requireCheckViolation(t, err, "invalid source update")

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM review_attempts WHERE id = $1`, legacyID).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows, "the NOT VALID rollout must preserve legacy rows")
}

func requireCheckViolation(t *testing.T, err error, operation string) {
	t.Helper()
	require.Error(t, err, operation+" must be rejected")
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}

func requireNotNullViolation(t *testing.T, err error, operation string) {
	t.Helper()
	require.Error(t, err, operation+" must be rejected")
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23502"), pqErr.Code)
}
