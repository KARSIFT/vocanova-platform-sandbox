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

const aiFeedbackOutcomeIntegrityMigration = "20260908140000_voc1398_ai_feedback_outcome_integrity.sql"

func TestAIFeedbackOutcomeIntegrityMigrationAndEntSchemaAgree(t *testing.T) {
	migration, err := os.ReadFile(aiFeedbackOutcomeIntegrityMigration)
	require.NoError(t, err)
	for _, invariant := range []string{
		"ai_feedback_attempts_feedback_json_required_on_success",
		"status <> 'succeeded' OR (feedback_json IS NOT NULL AND jsonb_typeof(feedback_json) = 'object')",
		"status = 'succeeded' OR feedback_json IS NULL",
		"status = 'succeeded' OR feedback_text IS NULL",
		"status = 'failed' OR error_code IS NULL",
		"status = 'failed' OR error_message IS NULL",
		"NOT VALID",
	} {
		require.Contains(t, string(migration), invariant)
	}

	schema, err := os.ReadFile(filepath.Join("..", "ent", "schema", "aifeedbackattempt.go"))
	require.NoError(t, err)
	for _, invariant := range []string{
		"feedback_json_required_on_success",
		"feedback_json_only_on_success",
		"feedback_text_only_on_success",
		"error_code_only_on_failure",
		"error_message_only_on_failure",
		"jsonb_typeof(feedback_json) = 'object'",
	} {
		require.Contains(t, string(schema), invariant)
	}
}

// TestAIFeedbackOutcomeIntegrityConstraintAgainstRealPostgres applies the
// committed migration to the production-shaped table. It covers JSON null,
// scalar, array, and object semantics; terminal-field exclusivity; updates;
// legacy rollout; and PostgreSQL's NOT VALID catalog state.
func TestAIFeedbackOutcomeIntegrityConstraintAgainstRealPostgres(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	schemaName := "ai_feedback_outcome_" + randomAIFeedbackSchemaSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schemaName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		require.NoError(t, err)
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schemaName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	applyMigrationsBeforeAIFeedbackOutcomeIntegrity(t, db)

	ctx := context.Background()
	now := time.Now().UTC()
	userID, wordID, meaningID, userWordID, sentenceID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, $2, 'active', 'completed', $3, $3)`, userID, fmt.Sprintf("feedback-integrity-%s@example.test", userID), now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, language_code, created_at, updated_at)
		VALUES ($1, 'feedback', 'feedback', 'en', $2, $2)`, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'a response', 1, $3, $3)`, meaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO user_words (id, user_id, meaning_id, source, added_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'manual', $4, $4, $4)`, userWordID, userID, meaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO learner_sentences (
		id, user_id, meaning_id, user_word_id, sentence_text, normalized_sentence_text,
		source, status, submitted_at, created_at, updated_at
	) VALUES ($1, $2, $3, $4, 'I need feedback.', 'i need feedback.', 'word_detail', 'submitted', $5, $5, $5)`, sentenceID, userID, meaningID, userWordID, now)
	require.NoError(t, err)

	insert := func(id uuid.UUID, status string, feedbackJSON, feedbackText, errorCode, errorMessage, completedAt any) error {
		_, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts (
			id, learner_sentence_id, status, provider, model, prompt_version, request_hash,
			feedback_json, feedback_text, error_code, error_message, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, 'test', 'test', 'v1', $4, $5, $6, $7, $8, $9, $10, $10)`,
			id, sentenceID, status, id.String(), feedbackJSON, feedbackText, errorCode, errorMessage, completedAt, now)
		return err
	}

	legacyID := uuid.New()
	require.NoError(t, insert(legacyID, "succeeded", nil, nil, nil, nil, now), "legacy invalid state must be possible before the forward migration")
	migration, err := os.ReadFile(aiFeedbackOutcomeIntegrityMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID must install beside legacy invalid rows")

	for _, name := range []string{
		"ai_feedback_attempts_feedback_json_required_on_success",
		"ai_feedback_attempts_feedback_json_only_on_success",
		"ai_feedback_attempts_feedback_text_only_on_success",
		"ai_feedback_attempts_error_code_only_on_failure",
		"ai_feedback_attempts_error_message_only_on_failure",
	} {
		var validated bool
		err = db.QueryRowContext(ctx, `SELECT convalidated FROM pg_constraint
			WHERE conrelid = 'ai_feedback_attempts'::regclass AND conname = $1`, name).Scan(&validated)
		require.NoError(t, err)
		require.False(t, validated)
	}
	var legacyStatus string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM ai_feedback_attempts WHERE id = $1`, legacyID).Scan(&legacyStatus))
	require.Equal(t, "succeeded", legacyStatus)
	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET updated_at = $2 WHERE id = $1`, legacyID, now.Add(time.Second))
	require.Error(t, err, "every update must enforce the newly installed checks")

	require.NoError(t, insert(uuid.New(), "pending", nil, nil, nil, nil, nil))
	require.NoError(t, insert(uuid.New(), "succeeded", `{"outcome":"correct"}`, "Good explanation.", nil, nil, now))
	require.NoError(t, insert(uuid.New(), "failed", nil, nil, "temporary_failure", "provider timeout", now))
	require.NoError(t, insert(uuid.New(), "cancelled", nil, nil, nil, nil, now))
	for _, malformed := range []struct {
		name                                                 string
		status                                               string
		feedbackJSON, feedbackText, code, message, completed any
	}{
		{name: "success without payload", status: "succeeded", completed: now},
		{name: "success JSON null", status: "succeeded", feedbackJSON: `null`, completed: now},
		{name: "success JSON scalar", status: "succeeded", feedbackJSON: `"correct"`, completed: now},
		{name: "success JSON array", status: "succeeded", feedbackJSON: `[]`, completed: now},
		{name: "success failure fields", status: "succeeded", feedbackJSON: `{"outcome":"correct"}`, code: "temporary_failure", completed: now},
		{name: "failed success JSON", status: "failed", feedbackJSON: `{"outcome":"correct"}`, code: "temporary_failure", completed: now},
		{name: "failed learner text", status: "failed", feedbackText: "leaked", code: "temporary_failure", completed: now},
		{name: "pending error", status: "pending", code: "temporary_failure"},
		{name: "cancelled text", status: "cancelled", feedbackText: "leaked", completed: now},
	} {
		t.Run(malformed.name, func(t *testing.T) {
			require.Error(t, insert(uuid.New(), malformed.status, malformed.feedbackJSON, malformed.feedbackText, malformed.code, malformed.message, malformed.completed))
		})
	}

	validID := uuid.New()
	require.NoError(t, insert(validID, "succeeded", `{"outcome":"correct"}`, "Good explanation.", nil, nil, now))
	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET status = 'failed', error_code = 'temporary_failure' WHERE id = $1`, validID)
	require.Error(t, err, "status changes cannot retain a success payload")
}

func randomAIFeedbackSchemaSuffix(t *testing.T, bytes int) string {
	t.Helper()
	buf := make([]byte, bytes)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}

func applyMigrationsBeforeAIFeedbackOutcomeIntegrity(t *testing.T, db *sql.DB) {
	t.Helper()
	paths, err := filepath.Glob("*.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	for _, path := range paths {
		if path == aiFeedbackOutcomeIntegrityMigration {
			break
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoErrorf(t, err, "apply migration %s", path)
	}
}
