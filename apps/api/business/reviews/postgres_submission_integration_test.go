package reviews

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLSubmitReviewAnchorsScheduleToServerClock exercises the actual
// submission transaction against PostgreSQL. The connection-local tables keep
// the fixture isolated from persistent application data.
func TestPostgreSQLSubmitReviewAnchorsScheduleToServerClock(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL submission test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1) // TEMP tables are scoped to this connection.

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		CREATE TEMP TABLE idempotency_keys (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL,
			key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL,
			UNIQUE(user_id, operation, key)
		);
		CREATE TEMP TABLE user_words (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, meaning_id uuid NOT NULL,
			review_step integer NOT NULL, total_review_count integer NOT NULL,
			correct_review_count integer NOT NULL, consecutive_correct_count integer NOT NULL,
			consecutive_incorrect_count integer NOT NULL, next_review_at timestamptz,
			last_reviewed_at timestamptz, last_result text, last_rating text,
			updated_at timestamptz NOT NULL, deleted_at timestamptz
		);
		CREATE TEMP TABLE review_attempts (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, user_word_id uuid NOT NULL,
			meaning_id uuid NOT NULL, attempt_type text NOT NULL, prompt_type text NOT NULL,
			result text NOT NULL, rating text, review_step_before integer NOT NULL,
			review_step_after integer NOT NULL, answered_at timestamptz NOT NULL,
			response_time_ms integer NOT NULL, selected_option_meaning_id uuid,
			typed_answer text, was_hint_used boolean NOT NULL, source text NOT NULL,
			client_attempt_id text, metadata jsonb, created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			UNIQUE (user_id, client_attempt_id)
		);`)
	require.NoError(t, err)

	serverNow := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: serverNow})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), clock.Fixed{T: serverNow})
	userID, meaningID := uuid.New(), uuid.New()

	for _, tc := range []struct {
		name       string
		answeredAt time.Time
	}{
		{name: "stale client history", answeredAt: serverNow.Add(-30 * 24 * time.Hour)},
		{name: "future client history", answeredAt: serverNow.Add(30 * 24 * time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			userWordID := uuid.New()
			_, err := db.ExecContext(ctx, `
				INSERT INTO user_words (
					id, user_id, meaning_id, review_step, total_review_count,
					correct_review_count, consecutive_correct_count, consecutive_incorrect_count,
					updated_at
				) VALUES ($1, $2, $3, 0, 0, 0, 0, 0, $4)`,
				userWordID, userID, meaningID, serverNow)
			require.NoError(t, err)

			attempt, err := svc.SubmitReview(ctx, SubmitReviewRequest{
				UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
				AttemptType: AttemptTypeReview, PromptType: PromptTypeMultipleChoice,
				Result: ResultCorrect, Rating: RatingGood, SelectedOptionMeaningID: &meaningID,
				AnsweredAt: tc.answeredAt, Source: SourceReview,
				ClientAttemptID: "server-clock-" + userWordID.String(),
				IdempotencyKey:  "server-clock-" + userWordID.String(),
			})
			require.NoError(t, err)
			require.Equal(t, tc.answeredAt, attempt.AnsweredAt)
			require.Equal(t, serverNow.Add(time.Hour), attempt.NextReviewAt)

			var storedAnsweredAt, nextReviewAt, lastReviewedAt time.Time
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT answered_at FROM review_attempts WHERE id = $1`, attempt.ID).Scan(&storedAnsweredAt))
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT next_review_at, last_reviewed_at FROM user_words WHERE id = $1`, userWordID).
				Scan(&nextReviewAt, &lastReviewedAt))
			require.True(t, storedAnsweredAt.Equal(tc.answeredAt), "attempt history remains client-recorded")
			require.True(t, nextReviewAt.Equal(serverNow.Add(time.Hour)), "schedule is server-owned")
			require.True(t, lastReviewedAt.Equal(serverNow), "last review time is server-owned")
		})
	}
}
