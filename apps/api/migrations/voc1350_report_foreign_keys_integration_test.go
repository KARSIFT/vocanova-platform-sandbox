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

func TestVOC1350ReportForeignKeysRestrictParentDeletion(t *testing.T) {
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

	migration, err := os.ReadFile("20260908010000_voc1350_restrict_ai_feedback_report_foreign_keys.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply voc1350 migration: %v", err)
	}

	for _, name := range []string{
		"ai_feedback_quality_review_reports_ai_feedback_attempt_id_fkey",
		"ai_feedback_quality_review_reports_user_id_fkey",
	} {
		var action string
		if err := db.QueryRowContext(ctx, `SELECT confdeltype FROM pg_constraint WHERE conname = $1`, name).Scan(&action); err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if action != "r" {
			t.Fatalf("%s delete action = %q, want RESTRICT", name, action)
		}
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	reportUserID, sentenceUserID := uuid.New(), uuid.New()
	sentenceID, attemptID, reportID := uuid.New(), uuid.New(), uuid.New()
	t.Cleanup(func() {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			return
		}
		defer tx.Rollback()
		_, _ = tx.ExecContext(context.Background(), `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
		_, _ = tx.ExecContext(context.Background(), `DELETE FROM ai_feedback_quality_review_reports WHERE id = $1`, reportID)
		_, _ = tx.ExecContext(context.Background(), `DELETE FROM ai_feedback_attempts WHERE id = $1`, attemptID)
		_, _ = tx.ExecContext(context.Background(), `DELETE FROM learner_sentences WHERE id = $1`, sentenceID)
		_, _ = tx.ExecContext(context.Background(), `DELETE FROM users WHERE id IN ($1, $2)`, reportUserID, sentenceUserID)
		_ = tx.Commit()
	})

	for _, userID := range []uuid.UUID{reportUserID, sentenceUserID} {
		if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO learner_sentences (id, user_id, sentence_text, normalized_sentence_text, source, status, submitted_at, created_at, updated_at) VALUES ($1, $2, 'A private sentence.', 'a private sentence.', 'free_practice', 'feedback_ready', $3, $3, $3)`, sentenceID, sentenceUserID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_attempts (id, learner_sentence_id, status, provider, model, prompt_version, request_hash, feedback_json, completed_at, created_at, updated_at) VALUES ($1, $2, 'succeeded', 'test', 'test', 'v1', $3, '{"status":"correct"}', $4, $4, $4)`, attemptID, sentenceID, attemptID.String(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO ai_feedback_quality_review_reports (id, ai_feedback_attempt_id, user_id, reason, state, created_at, updated_at) VALUES ($1, $2, $3, 'already_correct', 'open', $4, $4)`, reportID, attemptID, reportUserID, now); err != nil {
		t.Fatal(err)
	}

	// The immutable-attempt trigger rejects an ordinary delete before PostgreSQL
	// reaches this FK. Enable only the account-purge-local gate here so this
	// focused migration test can still verify the report-before-attempt order.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`); err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM ai_feedback_attempts WHERE id = $1`, attemptID)
	assertForeignKeyViolation(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, reportUserID)
	assertForeignKeyViolation(t, err)
}

func assertForeignKeyViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("parent deletion succeeded despite linked report")
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23503" {
		t.Fatalf("delete error = %v, want PostgreSQL foreign-key violation", err)
	}
}
