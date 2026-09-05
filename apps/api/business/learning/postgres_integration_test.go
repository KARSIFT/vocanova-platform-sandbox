package learning

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

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
