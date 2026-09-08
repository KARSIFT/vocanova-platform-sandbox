//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestVOC1423FeedbackReportOwnerMigrationOnPostgres(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := t.Context()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	migration, err := os.ReadFile("20260908210000_voc1423_feedback_report_owner.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply voc1423 migration: %v", err)
	}

	var triggerEnabled string
	if err := db.QueryRowContext(ctx, `SELECT tgenabled FROM pg_trigger WHERE tgname = 'ai_feedback_quality_review_reports_owner_matches_attempt'`).Scan(&triggerEnabled); err != nil {
		t.Fatalf("read report-owner trigger: %v", err)
	}
	if triggerEnabled != "O" {
		t.Fatalf("report-owner trigger enabled state = %q, want O", triggerEnabled)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	ownerID, otherUserID := uuid.New(), uuid.New()
	sentenceID, otherSentenceID := uuid.New(), uuid.New()
	ownerAttemptID, otherAttemptID := uuid.New(), uuid.New()
	validReportID, rejectedReportID := uuid.New(), uuid.New()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM ai_feedback_quality_review_reports WHERE id IN ($1, $2)`, validReportID, rejectedReportID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM ai_feedback_attempts WHERE id IN ($1, $2)`, ownerAttemptID, otherAttemptID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM learner_sentences WHERE id IN ($1, $2)`, sentenceID, otherSentenceID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id IN ($1, $2)`, ownerID, otherUserID)
	})

	for _, userID := range []uuid.UUID{ownerID, otherUserID} {
		if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO learner_sentences (id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at) VALUES ($1, $2, 'A private sentence.', 'a private sentence.', 'free_practice', 'feedback_ready', $3, $3, $3)`, sentenceID, ownerID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO learner_sentences (id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at) VALUES ($1, $2, 'Another private sentence.', 'another private sentence.', 'free_practice', 'feedback_ready', $3, $3, $3)`, otherSentenceID, otherUserID, now); err != nil {
		t.Fatal(err)
	}
	for _, attemptID := range []uuid.UUID{ownerAttemptID, otherAttemptID} {
		if _, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts (id, learner_sentence_id, status, provider, model, prompt_version, request_hash, feedback_json, completed_at, created_at, updated_at) VALUES ($1, $2, 'succeeded', 'test', 'test', 'v1', $3, '{"status":"correct"}', $4, $4, $4)`, attemptID, sentenceID, attemptID.String(), now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_quality_review_reports (id, ai_feedback_attempt_id, user_id, reason, state, created_at, updated_at) VALUES ($1, $2, $3, 'already_correct', 'open', $4, $4)`, validReportID, ownerAttemptID, ownerID, now); err != nil {
		t.Fatalf("insert same-owner report: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO ai_feedback_quality_review_reports (id, ai_feedback_attempt_id, user_id, reason, state, created_at, updated_at) VALUES ($1, $2, $3, 'already_correct', 'open', $4, $4)`, rejectedReportID, otherAttemptID, otherUserID, now)
	assertReportOwnerViolation(t, err)

	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_quality_review_reports SET user_id = $1 WHERE id = $2`, otherUserID, validReportID)
	assertReportOwnerViolation(t, err)

	_, err = db.ExecContext(ctx, `UPDATE ai_feedback_attempts SET learner_sentence_id = $1 WHERE id = $2`, otherSentenceID, ownerAttemptID)
	assertReportOwnerViolation(t, err)

	_, err = db.ExecContext(ctx, `UPDATE learner_sentences SET user_id = $1 WHERE id = $2`, otherUserID, sentenceID)
	assertReportOwnerViolation(t, err)
}

func assertReportOwnerViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("cross-user report write succeeded")
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23503" {
		t.Fatalf("report-owner error = %v, want PostgreSQL foreign-key-style violation", err)
	}
}
