package main

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// This opt-in test uses one pinned connection and only temporary tables. It
// exercises the canonical schema's uniqueness/FK constraints, not mocked SQL.
func TestCanonicalSeedPostgreSQLRerunPreservesLearningReference(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()

	// Reuse the original six content-table definitions and indexes, including
	// their actual check, unique and foreign-key constraints. No migration runs
	// against persistent tables: every CREATE TABLE is made temporary first.
	migration, err := os.ReadFile("../../migrations/20260725100000_voc026_p1_content_tables.sql")
	require.NoError(t, err)
	ddl, _, found := strings.Cut(string(migration), "CREATE TABLE user_words (")
	require.True(t, found, "content schema boundary must remain explicit")
	require.Equal(t, 6, strings.Count(ddl, "CREATE TABLE "))
	ddl = strings.ReplaceAll(ddl, "CREATE TABLE ", "CREATE TEMP TABLE ")
	_, err = conn.ExecContext(ctx, ddl)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `CREATE TEMP TABLE seed_saved_learning (
  id uuid PRIMARY KEY,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  review_step integer NOT NULL,
  total_review_count integer NOT NULL,
  next_review_at timestamptz NOT NULL
 )`)
	require.NoError(t, err)
	seed, err := loadSeed()
	require.NoError(t, err)
	expectedCounts := map[string]int{
		"journey_situations": 17, "canonical_words": 89, "word_meanings": 92,
		"word_examples": 148, "usage_notes": 200, "journey_words": 92,
	}
	const savedID = "edf9c1dc-0c0d-4bf1-a6e5-22c2c1fd1bf7"
	const retainedMeaningID = "3d64c3c9-ede0-5ffd-b1ef-278f6b70e486" // catch up
	nextReview := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	for pass := 1; pass <= 2; pass++ {
		tx, err := conn.BeginTx(ctx, nil)
		require.NoError(t, err)
		if err := applySeed(tx, seed); err != nil {
			_ = tx.Rollback()
			t.Fatalf("seed pass %d: %v", pass, err)
		}
		require.NoError(t, tx.Commit())
		for table, want := range expectedCounts {
			var count int
			// Table names are fixed test constants; pg_temp prevents persistent reads.
			require.NoError(t, conn.QueryRowContext(ctx, "SELECT count(*) FROM pg_temp."+table).Scan(&count))
			require.Equal(t, want, count, "%s after pass %d", table, pass)
		}
		if pass == 1 {
			_, err = conn.ExecContext(ctx, `INSERT INTO seed_saved_learning (id,meaning_id,review_step,total_review_count,next_review_at) VALUES ($1,$2,4,7,$3)`, savedID, retainedMeaningID, nextReview)
			require.NoError(t, err)
		}
	}
	var meaningID, wordText string
	var step, reviews int
	var next time.Time
	err = conn.QueryRowContext(ctx, `SELECT s.meaning_id,s.review_step,s.total_review_count,s.next_review_at,cw.text
  FROM pg_temp.seed_saved_learning s
  JOIN pg_temp.word_meanings m ON m.id=s.meaning_id
  JOIN pg_temp.canonical_words cw ON cw.id=m.word_id
  WHERE s.id=$1`, savedID).Scan(&meaningID, &step, &reviews, &next, &wordText)
	require.NoError(t, err)
	require.Equal(t, retainedMeaningID, meaningID)
	require.Equal(t, "catch up", wordText)
	require.Equal(t, 4, step)
	require.Equal(t, 7, reviews)
	require.True(t, next.Equal(nextReview), "saved review schedule changed")

	// Match the API's core-first ordering, which precedes display_order.
	rows, err := conn.QueryContext(ctx, `SELECT cw.text FROM pg_temp.journey_words jw
  JOIN pg_temp.word_meanings m ON m.id=jw.meaning_id
  JOIN pg_temp.canonical_words cw ON cw.id=m.word_id
  WHERE jw.journey_situation_id=$1 AND m.status='active' AND cw.status='active'
  ORDER BY jw.is_core DESC, jw.display_order ASC NULLS LAST, jw.relevance_score DESC, m.id ASC`, dailySituationID)
	require.NoError(t, err)
	defer rows.Close()
	var order []string
	for rows.Next() {
		var text string
		require.NoError(t, rows.Scan(&text))
		order = append(order, text)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"greeting", "small talk", "casual", "weekend plans", "available", "invite", "join", "suggest", "sounds good", "arrange", "meet up", "confirm", "on time", "reschedule", "cancel", "catch up", "keep in touch", "farewell"}, order)
}
