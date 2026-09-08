package reviews

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestReviewAttemptResultRatingConstraintPostgreSQL uses a connection-local
// table to exercise PostgreSQL's actual CHECK behavior without touching the
// shared application schema. The check exactly mirrors the forward migration.
func TestReviewAttemptResultRatingConstraintPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		CREATE TEMP TABLE review_attempts (
			id integer PRIMARY KEY,
			result text NOT NULL CHECK (result IN ('correct', 'incorrect', 'skipped')),
			rating text CHECK (rating IS NULL OR rating IN ('again', 'hard', 'good', 'easy'))
		);
		ALTER TABLE review_attempts
			ADD CONSTRAINT review_attempts_result_rating_valid
			CHECK (
				(result = 'skipped' AND rating IS NULL) OR
				(result = 'incorrect' AND rating IS NOT NULL AND rating = 'again') OR
				(result = 'correct' AND rating IS NOT NULL AND rating IN ('hard', 'good', 'easy'))
			) NOT VALID;`)
	require.NoError(t, err)

	insert := func(id int, result string, rating any) error {
		_, err := db.ExecContext(ctx, `INSERT INTO review_attempts (id, result, rating) VALUES ($1, $2, $3)`, id, result, rating)
		return err
	}

	for id, valid := range []struct {
		result string
		rating any
	}{
		{result: ResultSkipped, rating: nil},
		{result: ResultIncorrect, rating: RatingAgain},
		{result: ResultCorrect, rating: RatingHard},
		{result: ResultCorrect, rating: RatingGood},
		{result: ResultCorrect, rating: RatingEasy},
	} {
		require.NoError(t, insert(id+1, valid.result, valid.rating))
	}

	for id, invalid := range []struct {
		result string
		rating any
	}{
		{result: ResultSkipped, rating: RatingEasy},
		{result: ResultIncorrect, rating: nil},
		{result: ResultIncorrect, rating: RatingHard},
		{result: ResultCorrect, rating: nil},
		{result: ResultCorrect, rating: RatingAgain},
	} {
		require.Errorf(t, insert(100+id, invalid.result, invalid.rating), "invalid result/rating pair %q/%v must be rejected", invalid.result, invalid.rating)
	}
}
