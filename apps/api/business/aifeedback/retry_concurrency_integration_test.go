//go:build integration

package aifeedback

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLRepositoryConcurrentRetryReturnsExistingGeneration verifies the
// real partial unique index race. It uses a throwaway schema in the opt-in
// validation database, never an application schema.
func TestPostgreSQLRepositoryConcurrentRetryReturnsExistingGeneration(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.PingContext(ctx))

	schema := "aifeedback_retry_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() { _, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()

	// search_path is set on the connection string so both concurrent repository
	// transactions use only the disposable schema.
	testDB, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer testDB.Close()
	require.NoError(t, testDB.PingContext(ctx))
	_, err = testDB.ExecContext(ctx, `
		CREATE TABLE learner_sentences (id uuid PRIMARY KEY, status text NOT NULL, updated_at timestamptz NOT NULL);
		CREATE TABLE ai_feedback_attempts (
			id uuid PRIMARY KEY, learner_sentence_id uuid NOT NULL, status text NOT NULL,
			provider text NOT NULL, model text NOT NULL, prompt_version text NOT NULL, request_hash text NOT NULL,
			feedback_json jsonb, feedback_text text, error_code text, error_message text,
			started_at timestamptz, completed_at timestamptz, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL
		);
		CREATE TABLE ai_feedback_quality_review_reports (ai_feedback_attempt_id uuid NOT NULL);
		CREATE UNIQUE INDEX ai_feedback_attempts_request_hash_active_key
			ON ai_feedback_attempts (request_hash) WHERE status IN ('pending', 'succeeded');`)
	require.NoError(t, err)

	now := time.Now().UTC()
	sentenceID, failedID := uuid.New(), uuid.New()
	require.NoError(t, execRetrySQL(ctx, testDB, `INSERT INTO learner_sentences (id, status, updated_at) VALUES ($1, 'feedback_failed', $2)`, sentenceID, now))
	failed := &StoredFeedbackAttempt{ID: failedID, LearnerSentenceID: sentenceID, Status: AttemptStatusFailed, RequestHash: "concurrent-retry-hash"}
	require.NoError(t, execRetrySQL(ctx, testDB, `INSERT INTO ai_feedback_attempts (id, learner_sentence_id, status, provider, model, prompt_version, request_hash, error_code, created_at, updated_at) VALUES ($1, $2, 'failed', 'mock', 'mock', 'v1', $3, 'temporary_failure', $4, $4)`, failedID, sentenceID, failed.RequestHash, now))

	start := make(chan struct{})
	results := make(chan *RetryAttempt, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			retry, err := NewPostgreSQLRepository(testDB, nil).CreateRetryAttempt(ctx, failed, ProviderMock, "mock", now)
			results <- retry
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var created, extant int
	for retry := range results {
		if retry.Pending != nil {
			created++
		} else {
			require.NotNil(t, retry.Existing)
			require.Equal(t, AttemptStatusPending, retry.Existing.Status)
			extant++
		}
	}
	require.Equal(t, 1, created)
	require.Equal(t, 1, extant)
	var active int
	require.NoError(t, testDB.QueryRowContext(ctx, `SELECT count(*) FROM ai_feedback_attempts WHERE request_hash = $1 AND status IN ('pending', 'succeeded')`, failed.RequestHash).Scan(&active))
	require.Equal(t, 1, active)
}

func execRetrySQL(ctx context.Context, db *sql.DB, query string, args ...any) error {
	_, err := db.ExecContext(ctx, query, args...)
	return err
}
