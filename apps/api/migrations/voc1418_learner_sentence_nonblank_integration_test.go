//go:build integration

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
)

func TestVOC1418LearnerSentenceNonblankMigrationOnPostgres(t *testing.T) {
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

	ensureVOC1418LearnerSentenceConstraints(t, ctx, db)

	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	validSentenceID := uuid.New()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM learner_sentences WHERE id = $1`, validSentenceID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at)
		VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now); err != nil {
		t.Fatal(err)
	}
	insertSentence := func(id uuid.UUID, sentenceText, normalizedText string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO learner_sentences (
			id, user_id, sentence_text, normalized_sentence_text, source, status,
			submitted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'free_practice', 'submitted', $5, $5, $5)`,
			id, userID, sentenceText, normalizedText, now)
		return err
	}

	if err := insertSentence(validSentenceID, "I work every day.", "i work every day."); err != nil {
		t.Fatalf("insert valid learner sentence: %v", err)
	}
	assertLearnerSentenceCheckViolation(t, insertSentence(uuid.New(), " \t\r\n ", "valid normalized sentence"))
	assertLearnerSentenceCheckViolation(t, insertSentence(uuid.New(), "valid learner sentence", "\n\t  "))

	_, err = db.ExecContext(ctx, `UPDATE learner_sentences
		SET sentence_text = $2, updated_at = $3 WHERE id = $1`, validSentenceID, " \t\r\n ", now)
	assertLearnerSentenceCheckViolation(t, err)
	_, err = db.ExecContext(ctx, `UPDATE learner_sentences
		SET normalized_sentence_text = $2, updated_at = $3 WHERE id = $1`, validSentenceID, "\n\t  ", now)
	assertLearnerSentenceCheckViolation(t, err)
}

func ensureVOC1418LearnerSentenceConstraints(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	const countSQL = `SELECT count(*) FROM pg_constraint
		WHERE conrelid = 'learner_sentences'::regclass
		  AND conname IN (
			'learner_sentences_sentence_text_nonblank',
			'learner_sentences_normalized_sentence_text_nonblank'
		  )`
	var count int
	if err := db.QueryRowContext(ctx, countSQL).Scan(&count); err != nil {
		t.Fatalf("count learner sentence nonblank constraints: %v", err)
	}
	if count == 0 {
		migration, err := os.ReadFile("20260908030000_voc1418_learner_sentence_nonblank.sql")
		if err != nil {
			t.Fatalf("read voc1418 learner sentence nonblank migration: %v", err)
		}
		if _, err := db.ExecContext(ctx, string(migration)); err != nil {
			t.Fatalf("apply voc1418 learner sentence nonblank migration: %v", err)
		}
		count = 2
	}
	if count != 2 {
		t.Fatalf("learner sentence nonblank constraint count = %d, want 2", count)
	}

	for _, constraint := range []struct {
		name       string
		definition string
	}{
		{"learner_sentences_sentence_text_nonblank", "sentence_text ~ '[^[:space:]]'"},
		{"learner_sentences_normalized_sentence_text_nonblank", "normalized_sentence_text ~ '[^[:space:]]'"},
	} {
		var validated bool
		var definition string
		err := db.QueryRowContext(ctx, `SELECT convalidated, pg_get_constraintdef(oid)
			FROM pg_constraint WHERE conname = $1`, constraint.name).Scan(&validated, &definition)
		if err != nil {
			t.Fatalf("read %s catalog entry: %v", constraint.name, err)
		}
		if validated {
			t.Errorf("%s was validated during the additive rollout; want NOT VALID", constraint.name)
		}
		if !strings.Contains(definition, constraint.definition) {
			t.Errorf("%s definition = %q, want %q", constraint.name, definition, constraint.definition)
		}
	}
}

func assertLearnerSentenceCheckViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("whitespace-only learner sentence write unexpectedly succeeded")
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23514" {
		t.Fatalf("write error = %v, want PostgreSQL check violation", err)
	}
}
