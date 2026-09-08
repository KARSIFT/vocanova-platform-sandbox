//go:build integration

package missions

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const remainingActivityCounterMigration = "20260908180000_daily_activity_remaining_counter_integrity.sql"

func TestDailyActivityRemainingCounterConstraintsAgainstRealPostgres(t *testing.T) {
	db := newDailyActivityCounterValidationDB(t)
	applyDailyActivityCounterMigrations(t, db, "")
	userID := insertTestUser(t, db)
	baseDate := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)

	for index, tc := range []struct {
		name   string
		column string
	}{
		{name: "words_discovered", column: "words_discovered"},
		{name: "words_added", column: "words_added"},
		{name: "sentences_submitted", column: "sentences_submitted"},
		{name: "ai_feedback_received", column: "ai_feedback_received"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(t.Context(), fmt.Sprintf(`INSERT INTO daily_activity_summaries (
				id, user_id, local_date, timezone, %s, created_at, updated_at
			) VALUES ($1, $2, $3, 'UTC', -1, NOW(), NOW())`, tc.column),
				uuid.New(), userID, baseDate.AddDate(0, 0, index))
			requireConstraintViolation(t, err)

			rowID := uuid.New()
			_, err = db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
				id, user_id, local_date, timezone, created_at, updated_at
			) VALUES ($1, $2, $3, 'UTC', NOW(), NOW())`,
				rowID, userID, baseDate.AddDate(0, 0, index+10))
			require.NoError(t, err, "the zero/default aggregate state must remain writable")

			_, err = db.ExecContext(t.Context(), fmt.Sprintf(
				`UPDATE daily_activity_summaries SET %s = -1 WHERE id = $1`, tc.column), rowID)
			requireConstraintViolation(t, err)
		})
	}

	var defaults struct {
		wordsDiscovered    int
		wordsAdded         int
		sentencesSubmitted int
		aiFeedbackReceived int
	}
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT
		words_discovered, words_added, sentences_submitted, ai_feedback_received
		FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`,
		userID, baseDate.AddDate(0, 0, 10)).Scan(
		&defaults.wordsDiscovered, &defaults.wordsAdded, &defaults.sentencesSubmitted, &defaults.aiFeedbackReceived,
	))
	assert.Equal(t, 0, defaults.wordsDiscovered, "omitted counters must keep their PostgreSQL defaults")
	assert.Equal(t, 0, defaults.wordsAdded, "omitted counters must keep their PostgreSQL defaults")
	assert.Equal(t, 0, defaults.sentencesSubmitted, "omitted counters must keep their PostgreSQL defaults")
	assert.Equal(t, 0, defaults.aiFeedbackReceived, "omitted counters must keep their PostgreSQL defaults")

	_, err := db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
		id, user_id, local_date, timezone, words_discovered, words_added,
		sentences_submitted, ai_feedback_received, created_at, updated_at
	) VALUES ($1, $2, $3, 'UTC', 1, 2, 3, 4, NOW(), NOW())`,
		uuid.New(), userID, baseDate.AddDate(0, 0, 20))
	require.NoError(t, err, "positive counter values must remain writable")
}

func TestDailyActivityRemainingCounterMigrationPreservesLegacyRows(t *testing.T) {
	db := newDailyActivityCounterValidationDB(t)
	applyDailyActivityCounterMigrations(t, db, remainingActivityCounterMigration)
	userID := insertTestUser(t, db)
	legacyDate := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)

	_, err := db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
		id, user_id, local_date, timezone, words_discovered, created_at, updated_at
	) VALUES ($1, $2, $3, 'UTC', -1, NOW(), NOW())`, uuid.New(), userID, legacyDate)
	require.NoError(t, err, "the pre-migration schema must accept the legacy fixture")

	applyDailyActivityCounterMigration(t, db, remainingActivityCounterMigration)

	var legacyValue int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT words_discovered
		FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`, userID, legacyDate).Scan(&legacyValue))
	assert.Equal(t, -1, legacyValue, "the forward migration must not rewrite a legacy aggregate")

	var validated bool
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT convalidated
		FROM pg_constraint
		WHERE conname = 'daily_activity_summaries_remaining_counters_nonnegative'
		  AND conrelid = 'daily_activity_summaries'::regclass`).Scan(&validated))
	assert.False(t, validated, "the constraint must stay NOT VALID until a deliberate legacy reconciliation")

	_, err = db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
		id, user_id, local_date, timezone, sentences_submitted, created_at, updated_at
	) VALUES ($1, $2, $3, 'UTC', -1, NOW(), NOW())`, uuid.New(), userID, legacyDate.AddDate(0, 0, 1))
	requireConstraintViolation(t, err)

	_, err = db.ExecContext(t.Context(), `UPDATE daily_activity_summaries
		SET updated_at = NOW() WHERE user_id = $1 AND local_date = $2`, userID, legacyDate)
	requireConstraintViolation(t, err)
}

func newDailyActivityCounterValidationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })

	schema := "daily_activity_counters_" + randomDailyActivityCounterSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop validation schema %s: %v", schema, err)
		}
	})
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func applyDailyActivityCounterMigrations(t *testing.T, db *sql.DB, stopBefore string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(migrationsDirRelativeToPackage, "*.sql"))
	require.NoError(t, err)
	sort.Strings(paths)
	for _, path := range paths {
		name := filepath.Base(path)
		if stopBefore != "" && name == stopBefore {
			return
		}
		applyDailyActivityCounterMigration(t, db, name)
	}
	if stopBefore != "" {
		t.Fatalf("migration %s not found", stopBefore)
	}
}

func applyDailyActivityCounterMigration(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(migrationsDirRelativeToPackage, name))
	require.NoError(t, err)
	_, err = db.Exec(string(contents))
	require.NoErrorf(t, err, "apply migration %s", name)
}

func randomDailyActivityCounterSuffix(t *testing.T, bytes int) string {
	t.Helper()
	buffer := make([]byte, bytes)
	_, err := rand.Read(buffer)
	require.NoError(t, err)
	return hex.EncodeToString(buffer)
}

func requireConstraintViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.ErrorAs(t, err, &pqErr)
	assert.Equal(t, "23514", string(pqErr.Code))
}
