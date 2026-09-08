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

// TestVOC1438CanonicalWordFrequencyRankConstraint proves that the staged
// constraint leaves legacy rows untouched while checking all subsequent
// writes, including updates to pre-existing invalid rows.
func TestVOC1438CanonicalWordFrequencyRankConstraint(t *testing.T) {
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
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve connection for isolated schema: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// This isolated schema makes the test independent of whether the shared
	// database has already applied the full migration directory.
	schemaName := "voc1438_frequency_rank_" + uuid.NewString()[0:12]
	if _, err := conn.ExecContext(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schemaName+" CASCADE")
	})
	if _, err := conn.ExecContext(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatalf("set isolated schema search path: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE canonical_words (
			id uuid PRIMARY KEY,
			frequency_rank integer,
			updated_at timestamptz NOT NULL
		)`); err != nil {
		t.Fatalf("create canonical_words fixture: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	legacyID := uuid.New()
	if _, err := conn.ExecContext(ctx, `INSERT INTO canonical_words (id, frequency_rank, updated_at) VALUES ($1, -1, $2)`, legacyID, now); err != nil {
		t.Fatalf("insert pre-migration legacy row: %v", err)
	}
	migration, err := os.ReadFile("20260908270000_voc1438_canonical_word_frequency_rank.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply frequency-rank migration: %v", err)
	}

	var constraintType string
	var validated bool
	if err := conn.QueryRowContext(ctx, `
		SELECT contype, convalidated
		FROM pg_constraint
		WHERE conname = 'canonical_words_frequency_rank_positive'
		  AND conrelid = 'canonical_words'::regclass`).Scan(&constraintType, &validated); err != nil {
		t.Fatalf("read frequency-rank constraint catalog state: %v", err)
	}
	if constraintType != "c" {
		t.Fatalf("frequency-rank constraint type = %q, want check constraint c", constraintType)
	}
	if validated {
		t.Fatal("frequency-rank constraint must remain NOT VALID during staged rollout")
	}
	var legacyRank int
	if err := conn.QueryRowContext(ctx, `SELECT frequency_rank FROM canonical_words WHERE id = $1`, legacyID).Scan(&legacyRank); err != nil {
		t.Fatalf("read retained legacy rank: %v", err)
	}
	if legacyRank != -1 {
		t.Fatalf("legacy rank after staged migration = %d, want -1", legacyRank)
	}

	for _, rank := range []*int{nil, intPointer(1)} {
		if _, err := conn.ExecContext(ctx, `INSERT INTO canonical_words (id, frequency_rank, updated_at) VALUES ($1, $2, $3)`, uuid.New(), rank, now); err != nil {
			t.Fatalf("insert rank %v: %v", rank, err)
		}
	}

	assertFrequencyRankCheckViolation(t, execFrequencyRankWrite(ctx, conn, `
		INSERT INTO canonical_words (id, frequency_rank, updated_at) VALUES ($1, 0, $2)`, uuid.New(), now))
	assertFrequencyRankCheckViolation(t, execFrequencyRankWrite(ctx, conn, `
		INSERT INTO canonical_words (id, frequency_rank, updated_at) VALUES ($1, -1, $2)`, uuid.New(), now))

	validID := uuid.New()
	if _, err := conn.ExecContext(ctx, `INSERT INTO canonical_words (id, frequency_rank, updated_at) VALUES ($1, 1, $2)`, validID, now); err != nil {
		t.Fatalf("insert valid rank for update checks: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `UPDATE canonical_words SET frequency_rank = NULL WHERE id = $1`, validID); err != nil {
		t.Fatalf("update rank to NULL: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `UPDATE canonical_words SET frequency_rank = 1 WHERE id = $1`, validID); err != nil {
		t.Fatalf("update rank to boundary 1: %v", err)
	}
	assertFrequencyRankCheckViolation(t, execFrequencyRankWrite(ctx, conn, `
		UPDATE canonical_words SET frequency_rank = 0 WHERE id = $1`, validID))
	assertFrequencyRankCheckViolation(t, execFrequencyRankWrite(ctx, conn, `
		UPDATE canonical_words SET frequency_rank = -1 WHERE id = $1`, validID))
	assertFrequencyRankCheckViolation(t, execFrequencyRankWrite(ctx, conn, `
		UPDATE canonical_words SET updated_at = $2 WHERE id = $1`, legacyID, now.Add(time.Second)))
}

func intPointer(value int) *int { return &value }

func execFrequencyRankWrite(ctx context.Context, conn *sql.Conn, statement string, args ...any) error {
	_, err := conn.ExecContext(ctx, statement, args...)
	return err
}

func assertFrequencyRankCheckViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("write unexpectedly succeeded despite frequency-rank constraint")
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23514" {
		t.Fatalf("write error = %v, want PostgreSQL check violation", err)
	}
	if pqErr.Constraint != "canonical_words_frequency_rank_positive" {
		t.Fatalf("check violation constraint = %q, want canonical_words_frequency_rank_positive", pqErr.Constraint)
	}
}
