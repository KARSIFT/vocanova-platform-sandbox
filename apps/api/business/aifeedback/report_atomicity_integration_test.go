package aifeedback

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// isolatedReportPostgres gives every test a schema shared by its pooled
// connections. It deliberately does not use temporary tables: concurrent
// transactions must see the same indexes and trigger in a real PostgreSQL
// database while never touching the shared public schema.
func isolatedReportPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "aifeedback_report_" + uuid.NewString()[:8]
	_, err = admin.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
		_ = admin.Close()
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCreateQualityReviewReportAtomicPostgreSQL(t *testing.T) {
	db := isolatedReportPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, `
		CREATE TABLE idempotency_keys (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL,
			key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL,
			UNIQUE (user_id, operation, key)
		);
		CREATE TABLE ai_feedback_quality_review_reports (
			id uuid PRIMARY KEY, ai_feedback_attempt_id uuid NOT NULL UNIQUE,
			user_id uuid NOT NULL, reason text NOT NULL, state text NOT NULL,
			classification text, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL
		);
		CREATE FUNCTION hold_report_claim() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN PERFORM pg_sleep(0.15); RETURN NEW; END $$;
		CREATE TRIGGER hold_report_claim BEFORE INSERT ON idempotency_keys
		FOR EACH ROW EXECUTE FUNCTION hold_report_claim();`)
	require.NoError(t, err)

	repo := NewPostgreSQLRepository(db, nil)
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	user, otherUser := uuid.New(), uuid.New()
	firstAttempt, secondAttempt := uuid.New(), uuid.New()
	report := func(userID, attemptID uuid.UUID, reason string, at time.Time) QualityReviewReport {
		return QualityReviewReport{ID: uuid.New(), UserID: userID, AttemptID: attemptID, Reason: reason, State: QualityReviewStateOpen, CreatedAt: at}
	}

	// The Go barrier releases both calls together and the trigger keeps their
	// INSERT .. ON CONFLICT statements overlapping on distinct connections.
	start := make(chan struct{})
	ready := make(chan struct{}, 2)
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, attemptID := range []uuid.UUID{firstAttempt, secondAttempt} {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			ready <- struct{}{}
			<-start
			_, err := repo.CreateQualityReviewReport(ctx, report(user, id, ReportReasonAlreadyCorrect, now), "contended")
			results <- err
		}(attemptID)
	}
	<-ready
	<-ready
	close(start)
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrReportIdempotencyConflict):
			conflicts++
		default:
			t.Fatalf("concurrent report: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("success/conflict = %d/%d, want 1/1", successes, conflicts)
	}
	var reports int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM ai_feedback_quality_review_reports WHERE user_id = $1`, user).Scan(&reports))
	require.Equal(t, 1, reports)

	// Determine the winning attempt from the database and verify its matching
	// retry creates neither a report nor an additional claim-side effect.
	var winner uuid.UUID
	require.NoError(t, db.QueryRowContext(ctx, `SELECT ai_feedback_attempt_id FROM ai_feedback_quality_review_reports WHERE user_id = $1`, user).Scan(&winner))
	created, err := repo.CreateQualityReviewReport(ctx, report(user, winner, ReportReasonAlreadyCorrect, now), "contended")
	require.NoError(t, err)
	require.False(t, created)
	created, err = repo.CreateQualityReviewReport(ctx, report(user, winner, ReportReasonAlreadyCorrect, now), "winner-different-key")
	require.NoError(t, err)
	require.False(t, created, "the per-attempt unique constraint prevents duplicate reports across keys")
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM ai_feedback_quality_review_reports WHERE user_id = $1`, user).Scan(&reports))
	require.Equal(t, 1, reports)

	// A key is scoped to user and operation. A pre-existing other-operation
	// claim cannot conflict, and another user can reuse the same key.
	_, err = db.ExecContext(ctx, `INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at) VALUES ($1, $2, 'other_operation', 'isolated', 'other', $3)`, uuid.New(), user, now)
	require.NoError(t, err)
	created, err = repo.CreateQualityReviewReport(ctx, report(user, uuid.New(), ReportReasonAlreadyCorrect, now), "isolated")
	require.NoError(t, err)
	require.True(t, created)
	created, err = repo.CreateQualityReviewReport(ctx, report(otherUser, uuid.New(), ReportReasonAlreadyCorrect, now), "contended")
	require.NoError(t, err)
	require.True(t, created)

	// At exactly 24 hours the old claim may be replaced for a new attempt.
	created, err = repo.CreateQualityReviewReport(ctx, report(user, uuid.New(), ReportReasonAlreadyCorrect, now.Add(24*time.Hour)), "contended")
	require.NoError(t, err)
	require.True(t, created)

	// A per-attempt reason conflict occurs after the claim statement. Its
	// transaction must roll that fresh claim back, leaving the key usable.
	freshKey := "reason-conflict-must-rollback"
	_, err = repo.CreateQualityReviewReport(ctx, report(user, winner, ReportReasonExplanationUnclear, now), freshKey)
	require.ErrorIs(t, err, ErrReportIdempotencyConflict)
	created, err = repo.CreateQualityReviewReport(ctx, report(user, uuid.New(), ReportReasonExplanationUnclear, now), freshKey)
	require.NoError(t, err)
	require.True(t, created)

	// A real database report-write rejection must roll back its preceding key
	// claim too, so the same request can be retried after the failure clears.
	failingUser, failingAttempt := uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE ai_feedback_quality_review_reports ADD CONSTRAINT reject_report_user CHECK (user_id <> '%s'::uuid)`, failingUser))
	require.NoError(t, err)
	_, err = repo.CreateQualityReviewReport(ctx, report(failingUser, failingAttempt, ReportReasonAlreadyCorrect, now), "write-fails")
	require.Error(t, err)
	var claims int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM idempotency_keys WHERE user_id = $1 AND operation = 'report_sentence_feedback' AND key = 'write-fails'`, failingUser).Scan(&claims))
	require.Zero(t, claims)
	_, err = db.ExecContext(ctx, `ALTER TABLE ai_feedback_quality_review_reports DROP CONSTRAINT reject_report_user`)
	require.NoError(t, err)
	created, err = repo.CreateQualityReviewReport(ctx, report(failingUser, failingAttempt, ReportReasonAlreadyCorrect, now), "write-fails")
	require.NoError(t, err)
	require.True(t, created)
}
