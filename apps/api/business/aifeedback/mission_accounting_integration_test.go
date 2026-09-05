//go:build integration

package aifeedback_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Use all committed migrations and a private schema shared by several pooled
// connections. No production or shared public-schema rows are modified.
func missionAccountingDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	schema := "mission_atomic_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	t.Cleanup(func() { _ = db.Close() })
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoError(t, err, "migration %s", path)
	}
	return db
}

func TestProductionFeedbackMissionAccountingAtomicPostgreSQL(t *testing.T) {
	db := missionAccountingDB(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	userID, sentenceID, attemptID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO learner_sentences (id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at) VALUES ($1, $2, 'I work every day.', 'i work every day.', 'free_practice', 'submitted', $3, $3, $3)`, sentenceID, userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts (id, learner_sentence_id, status, provider, model, prompt_version, request_hash, created_at, updated_at) VALUES ($1, $2, 'pending', 'test', 'test', 'v1', $3, $4, $4)`, attemptID, sentenceID, attemptID.String(), now)
	require.NoError(t, err)

	gam := gamification.NewService(gamification.NewRepository(db))
	updater := missions.NewMissionUpdater(missions.NewService(missions.NewRepository(db), gam), gam)
	repo := aifeedback.NewPostgreSQLRepository(db, nil)
	pending := aifeedback.PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}
	feedback := &aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect, Explanation: "Good work.", RawJSON: map[string]any{"status": aifeedback.LearningStatusCorrect}}
	var accountingCalls atomic.Int32
	complete := func(ctx context.Context, tx *sql.Tx) (bool, error) {
		accountingCalls.Add(1)
		return updater.UpdateInTransaction(ctx, tx, userID, sentenceID)
	}

	// Fail a real activity write after the actual production updater has
	// awarded both ledger entries and incremented the sentence activity.
	_, err = db.ExecContext(ctx, `ALTER TABLE daily_activity_summaries ADD CONSTRAINT reject_feedback_activity CHECK (ai_feedback_received = 0)`)
	require.NoError(t, err)
	_, err = repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, complete)
	require.Error(t, err)
	var attemptStatus, sentenceStatus string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM ai_feedback_attempts WHERE id=$1`, attemptID).Scan(&attemptStatus))
	require.Equal(t, aifeedback.AttemptStatusPending, attemptStatus)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM learner_sentences WHERE id=$1`, sentenceID).Scan(&sentenceStatus))
	require.Equal(t, "submitted", sentenceStatus)
	for _, table := range []string{"confidence_point_ledger", "daily_activity_summaries", "daily_mission_snapshots", "streak_states"} {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", userID).Scan(&count))
		require.Zero(t, count, "rollback %s", table)
	}
	_, err = db.ExecContext(ctx, `ALTER TABLE daily_activity_summaries DROP CONSTRAINT reject_feedback_activity`)
	require.NoError(t, err)
	// Cancellation after the real updater has written all its rows must still
	// prevent the outer feedback transaction from publishing any of them.
	cancelledCtx, cancelCompletion := context.WithCancel(ctx)
	_, err = repo.CompleteSuccessfulFeedbackAttempt(cancelledCtx, pending, feedback, now, func(ctx context.Context, tx *sql.Tx) (bool, error) {
		completed, err := complete(ctx, tx)
		cancelCompletion()
		return completed, err
	})
	cancelCompletion()
	require.Error(t, err)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM ai_feedback_attempts WHERE id=$1`, attemptID).Scan(&attemptStatus))
	require.Equal(t, aifeedback.AttemptStatusPending, attemptStatus)
	for _, table := range []string{"confidence_point_ledger", "daily_activity_summaries", "daily_mission_snapshots", "streak_states"} {
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", userID).Scan(&count))
		require.Zero(t, count, "cancelled transaction %s", table)
	}
	accountingCalls.Store(0)

	// Contending completions execute on distinct pooled connections. Only
	// the pending-transition winner may run the non-idempotent activity writes.
	_, err = db.ExecContext(ctx, `CREATE FUNCTION hold_feedback_completion() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.1); RETURN NEW; END $$; CREATE TRIGGER hold_feedback_completion BEFORE UPDATE ON ai_feedback_attempts FOR EACH ROW EXECUTE FUNCTION hold_feedback_completion()`)
	require.NoError(t, err)
	start := make(chan struct{})
	ready := make(chan struct{}, 2)
	results := make(chan error, 2)
	for range 2 {
		go func() {
			ready <- struct{}{}
			<-start
			_, err := repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, complete)
			results <- err
		}()
	}
	<-ready
	<-ready
	close(start)
	require.NoError(t, <-results)
	require.NoError(t, <-results)
	require.EqualValues(t, 1, accountingCalls.Load())
	require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM ai_feedback_attempts WHERE id=$1`, attemptID).Scan(&attemptStatus))
	require.Equal(t, aifeedback.AttemptStatusSucceeded, attemptStatus)
	var awards, points, submitted, received, earned int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*), sum(amount) FROM confidence_point_ledger WHERE user_id=$1`, userID).Scan(&awards, &points))
	require.Equal(t, 2, awards)
	require.Equal(t, gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot, points)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT sentences_submitted, ai_feedback_received, confidence_points_earned FROM daily_activity_summaries WHERE user_id=$1`, userID).Scan(&submitted, &received, &earned))
	require.Equal(t, 1, submitted)
	require.Equal(t, 1, received)
	require.Equal(t, points, earned)

	_, err = repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, complete)
	require.NoError(t, err)
	require.EqualValues(t, 1, accountingCalls.Load(), "settled replay must not repeat production accounting")
}
