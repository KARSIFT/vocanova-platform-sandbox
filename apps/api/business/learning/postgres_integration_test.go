package learning

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestSaveUserWordRejectsInactiveCanonicalWordPostgreSQL verifies the save
// eligibility predicate against a real PostgreSQL engine. It uses connection-local
// temporary tables and never touches persistent data.
func TestSaveUserWordRejectsInactiveCanonicalWordPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `
		CREATE TEMP TABLE canonical_words (id uuid PRIMARY KEY, status text NOT NULL);
		CREATE TEMP TABLE word_meanings (
			id uuid PRIMARY KEY,
			word_id uuid NOT NULL REFERENCES canonical_words(id),
			status text NOT NULL
		);
		CREATE TEMP TABLE user_words (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			meaning_id uuid NOT NULL,
			status text NOT NULL,
			source text NOT NULL,
			review_step integer NOT NULL,
			next_review_at timestamptz,
			last_reviewed_at timestamptz,
			last_result text,
			last_rating text,
			consecutive_correct_count integer NOT NULL,
			consecutive_incorrect_count integer NOT NULL,
			total_review_count integer NOT NULL,
			correct_review_count integer NOT NULL,
			added_at timestamptz NOT NULL,
			mastered_at timestamptz,
			ignored_at timestamptz,
			deleted_at timestamptz,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL
		);`); err != nil {
		t.Fatal(err)
	}

	wordID, meaningID := uuid.New(), uuid.New()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO canonical_words (id, status) VALUES ($1, 'archived')`, wordID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO word_meanings (id, word_id, status) VALUES ($1, $2, 'active')`, meaningID, wordID,
	); err != nil {
		t.Fatal(err)
	}

	_, err = NewPostgreSQLRepository(db).SaveUserWord(ctx, SaveUserWordRequest{
		UserID:    uuid.New(),
		MeaningID: meaningID,
		Source:    "journey",
	}, time.Now().UTC())
	if !errors.Is(err, ErrMeaningNotFound) {
		t.Fatalf("SaveUserWord() error = %v, want ErrMeaningNotFound", err)
	}

	var savedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_words`).Scan(&savedCount); err != nil {
		t.Fatal(err)
	}
	if savedCount != 0 {
		t.Fatalf("saved rows = %d, want 0", savedCount)
	}
}

// TestSaveUserWordAtomicIdempotencyPostgreSQL exercises the production
// transaction path against PostgreSQL. Temporary tables and one connection keep
// it isolated from application data while still using the real unique-index and
// rollback semantics.
func TestSaveUserWordAtomicIdempotencyPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		CREATE TEMP TABLE canonical_words (id uuid PRIMARY KEY, text text NOT NULL DEFAULT 'word', normalized_text text NOT NULL DEFAULT 'word', status text NOT NULL);
		CREATE TEMP TABLE word_meanings (id uuid PRIMARY KEY, word_id uuid NOT NULL REFERENCES canonical_words(id), part_of_speech text NOT NULL DEFAULT 'noun', short_definition text NOT NULL DEFAULT 'definition', status text NOT NULL);
		CREATE TEMP TABLE user_words (id uuid PRIMARY KEY, user_id uuid NOT NULL, meaning_id uuid NOT NULL, status text NOT NULL, source text NOT NULL, review_step integer NOT NULL, next_review_at timestamptz, last_reviewed_at timestamptz, last_result text, last_rating text, consecutive_correct_count integer NOT NULL, consecutive_incorrect_count integer NOT NULL, total_review_count integer NOT NULL, correct_review_count integer NOT NULL, added_at timestamptz NOT NULL, mastered_at timestamptz, ignored_at timestamptz, deleted_at timestamptz, created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL);
		CREATE TEMP TABLE idempotency_keys (id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL, key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL, UNIQUE(user_id, operation, key));`)
	if err != nil {
		t.Fatal(err)
	}
	wordID, firstMeaning, secondMeaning, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, status) VALUES ($1, 'active'); INSERT INTO word_meanings (id, word_id, status) VALUES ($2, $1, 'active'), ($3, $1, 'active')`, wordID, firstMeaning, secondMeaning)
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewPostgreSQLRepository(db), nil, &clock.Fixed{T: time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, meaningID := range []uuid.UUID{firstMeaning, secondMeaning} {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			_, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: id, Source: "journey", IdempotencyKey: "one"})
			errs <- err
		}(meaningID)
	}
	wg.Wait()
	close(errs)
	var success, conflict int
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, ErrIdempotencyConflict) {
			conflict++
		} else {
			t.Fatalf("save: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success/conflict = %d/%d, want 1/1", success, conflict)
	}
	var words int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM user_words`).Scan(&words); err != nil || words != 1 {
		t.Fatalf("words = %d, err = %v; want 1", words, err)
	}

	// A failed save rolls its just-created claim back, so the same key can retry.
	badMeaning := uuid.New()
	_, err = svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: badMeaning, Source: "journey", IdempotencyKey: "retry"})
	if !errors.Is(err, ErrMeaningNotFound) {
		t.Fatalf("failed save = %v, want ErrMeaningNotFound", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, status) VALUES ($1, $2, 'active')`, badMeaning, wordID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: badMeaning, Source: "journey", IdempotencyKey: "retry"}); err != nil {
		t.Fatalf("retry after rollback: %v", err)
	}
}
