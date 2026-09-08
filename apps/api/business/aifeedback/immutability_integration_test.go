//go:build integration

package aifeedback_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// feedbackAttemptImmutabilityDB applies every committed forward migration in
// an isolated schema. It deliberately shares no public-schema rows with other
// local PostgreSQL validation work.
func feedbackAttemptImmutabilityDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	schema := "feedback_immutability_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema+" application_name="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoError(t, err, "migration %s", path)
	}
	return db
}

func TestAIFeedbackAttemptsAreImmutableExceptScopedAccountPurge(t *testing.T) {
	db := feedbackAttemptImmutabilityDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID, sentenceID, successID := uuid.New(), uuid.New(), uuid.New()
	insertFeedbackFixture(t, ctx, db, userID, sentenceID, successID, "pending", "success-request", now)
	repo := aifeedback.NewPostgreSQLRepository(db, nil)
	feedback := &aifeedback.ProviderFeedback{
		Status:      aifeedback.LearningStatusCorrect,
		Explanation: "Well done.",
		RawJSON:     map[string]any{"status": aifeedback.LearningStatusCorrect},
	}
	require.NoError(t, repo.CompleteFeedbackAttempt(ctx, aifeedback.PendingAttempt{SentenceID: sentenceID, AttemptID: successID}, feedback, "", "", now))

	// A report proves the account disposition retains its report-before-attempt
	// order even when feedback-attempt deletion is guarded.
	_, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_quality_review_reports
		(id, ai_feedback_attempt_id, user_id, reason, state, created_at, updated_at)
		VALUES ($1, $2, $3, 'already_correct', 'open', $4, $4)`, uuid.New(), successID, userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET feedback_text = 'tampered' WHERE id = $1`, successID)
	requireAttemptMutationViolation(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM ai_feedback_attempts WHERE id = $1`, successID)
	requireAttemptMutationViolation(t, err)

	// Retrying appends a new pending attempt; it must never rewrite the failed
	// generation that provides the request identity.
	failedID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts
		(id, learner_sentence_id, status, provider, model, prompt_version, request_hash, created_at, updated_at)
		VALUES ($1, $2, 'pending', 'test', 'test', 'v1', $3, $4, $4)`, failedID, sentenceID, "failed-request", now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET provider = 'tampered' WHERE id = $1`, failedID)
	requireAttemptMutationViolation(t, err)
	require.NoError(t, repo.CompleteFeedbackAttempt(ctx, aifeedback.PendingAttempt{SentenceID: sentenceID, AttemptID: failedID}, nil, "temporary_failure", "safe failure", now.Add(time.Second)))
	retry, err := repo.CreateRetryAttempt(ctx, &aifeedback.StoredFeedbackAttempt{ID: failedID, LearnerSentenceID: sentenceID, Status: aifeedback.AttemptStatusFailed, RequestHash: "failed-request"}, "test", "test", now.Add(time.Second))
	require.NoError(t, err)
	require.NotNil(t, retry.Pending)
	require.NotEqual(t, failedID, retry.Pending.AttemptID)
	var attempts int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM ai_feedback_attempts WHERE learner_sentence_id = $1`, sentenceID).Scan(&attempts))
	require.Equal(t, 3, attempts)

	// The production account-anonymization transaction enables the local gate,
	// removes reports first, and commits without exposing the setting later.
	counters, err := accounts.NewPostgreSQLRepository(db).AnonymizeUserData(ctx, userID)
	require.NoError(t, err)
	require.EqualValues(t, 1, counters.AIQualityReviewReports)
	require.EqualValues(t, 3, counters.AIFeedbackAttempts)

	secondUserID, secondSentenceID, secondAttemptID := uuid.New(), uuid.New(), uuid.New()
	insertFeedbackFixture(t, ctx, db, secondUserID, secondSentenceID, secondAttemptID, "pending", "rollback-request", now)
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `DELETE FROM ai_feedback_attempts WHERE id = $1`, secondAttemptID)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	_, err = db.ExecContext(ctx, `DELETE FROM ai_feedback_attempts WHERE id = $1`, secondAttemptID)
	requireAttemptMutationViolation(t, err)
}

func insertFeedbackFixture(t *testing.T, ctx context.Context, db *sql.DB, userID, sentenceID, attemptID uuid.UUID, status, requestHash string, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at)
		VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO learner_sentences
		(id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at)
		VALUES ($1, $2, 'I work carefully.', 'i work carefully.', 'free_practice', 'submitted', $3, $3, $3)`, sentenceID, userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts
		(id, learner_sentence_id, status, provider, model, prompt_version, request_hash, created_at, updated_at)
		VALUES ($1, $2, $3, 'test', 'test', 'v1', $4, $5, $5)`, attemptID, sentenceID, status, requestHash, now)
	require.NoError(t, err)
}

func requireAttemptMutationViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "want PostgreSQL error, got %v", err)
	require.Equal(t, pq.ErrorCode("55000"), pqErr.Code)
}
