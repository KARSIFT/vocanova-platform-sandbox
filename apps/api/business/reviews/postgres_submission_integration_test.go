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
		CREATE TEMP TABLE users(id uuid PRIMARY KEY,status text DEFAULT 'active',deleted_at timestamptz);
 CREATE TEMP TABLE word_meanings(id uuid PRIMARY KEY,word_id uuid NOT NULL);
 CREATE TEMP TABLE user_word_knowledge(user_id uuid,meaning_id uuid,self_reported_known bool);
 CREATE TEMP TABLE daily_review_successes(user_id uuid,meaning_id uuid,local_date date);
 CREATE TEMP TABLE user_settings(user_id uuid,timezone text);
 CREATE TEMP TABLE idempotency_keys (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, operation text NOT NULL,
			key text NOT NULL, fingerprint text NOT NULL, created_at timestamptz NOT NULL,
			UNIQUE(user_id, operation, key)
		);
		CREATE TEMP TABLE user_words (
			id uuid PRIMARY KEY, user_id uuid NOT NULL, meaning_id uuid NOT NULL, status text DEFAULT 'new',
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

	for _, tc := range []struct {
		name       string
		answeredAt time.Time
	}{
		{name: "stale client history", answeredAt: serverNow.Add(-30 * 24 * time.Hour)},
		{name: "future client history", answeredAt: serverNow.Add(30 * 24 * time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Each client-clock history is an independent learner and word. A
			// prior subcase's server-created daily audit must remain authoritative.
			userID, meaningID, userWordID := uuid.New(), uuid.New(), uuid.New()
			_, err := db.ExecContext(ctx, "INSERT INTO word_meanings(id,word_id)VALUES($1,$2)", meaningID, uuid.New())
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, "INSERT INTO users(id)VALUES($1)", userID)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, `
				INSERT INTO user_words (
					id, user_id, meaning_id, review_step, total_review_count,
					correct_review_count, consecutive_correct_count, consecutive_incorrect_count,
					updated_at
				) VALUES ($1, $2, $3, 0, 0, 0, 0, 0, $4)`,
				userWordID, userID, meaningID, serverNow)
			require.NoError(t, err)

			req := SubmitReviewRequest{
				UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
				AttemptType: AttemptTypeReview, PromptType: PromptTypeMultipleChoice,
				Result: ResultCorrect, Rating: RatingGood, SelectedOptionMeaningID: &meaningID,
				AnsweredAt: tc.answeredAt, Source: SourceReview,
				ClientAttemptID: "server-clock-" + userWordID.String(),
				IdempotencyKey:  "server-clock-" + userWordID.String(),
			}
			attempt, err := svc.SubmitReview(ctx, req)
			require.NoError(t, err)
			require.Equal(t, tc.answeredAt, attempt.AnsweredAt)
			require.Equal(t, serverNow.Add(24*time.Hour), attempt.NextReviewAt)

			var storedAnsweredAt, nextReviewAt, lastReviewedAt time.Time
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT answered_at FROM review_attempts WHERE id = $1`, attempt.ID).Scan(&storedAnsweredAt))
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT next_review_at, last_reviewed_at FROM user_words WHERE id = $1`, userWordID).
				Scan(&nextReviewAt, &lastReviewedAt))
			require.True(t, storedAnsweredAt.Equal(tc.answeredAt), "attempt history remains client-recorded")
			require.True(t, nextReviewAt.Equal(serverNow.Add(24*time.Hour)), "schedule is server-owned")
			require.True(t, lastReviewedAt.Equal(serverNow), "last review time is server-owned")
			// The active-user SELECT lock must accept a real row, and immutable
			// replay still works after the successful answer moves it into future.
			replayed, err := repo.SubmitReview(ctx, req)
			require.NoError(t, err)
			require.Equal(t, attempt.ID, replayed.ID)
			var total int
			require.NoError(t, db.QueryRowContext(ctx, `SELECT total_review_count FROM user_words WHERE id=$1`, userWordID).Scan(&total))
			require.Equal(t, 1, total)
		})
	}
	// A non-active, removed, or missing learner must fail at the user boundary
	// before any attempt or idempotency claim can be written.
	for _, boundary := range []struct {
		name             string
		status           string
		deleted, missing bool
	}{
		{name: "inactive", status: "suspended"},
		{name: "soft-deleted", status: "active", deleted: true},
		{name: "missing", missing: true},
	} {
		t.Run(boundary.name, func(t *testing.T) {
			owner := uuid.New()
			if !boundary.missing {
				_, err := db.ExecContext(ctx, `INSERT INTO users(id,status,deleted_at)VALUES($1,$2,CASE WHEN $3 THEN $4::timestamptz ELSE NULL END)`, owner, boundary.status, boundary.deleted, serverNow)
				require.NoError(t, err)
			}
			req := SubmitReviewRequest{UserID: owner, UserWordID: uuid.New(), MeaningID: uuid.New(), AttemptType: AttemptTypeReview, PromptType: PromptTypeSelfCheck, Result: ResultCorrect, Rating: RatingGood, AnsweredAt: serverNow, Source: SourceReview, ClientAttemptID: "boundary-" + boundary.name, IdempotencyKey: "boundary-" + boundary.name}
			_, err := repo.SubmitReview(ctx, req)
			require.ErrorIs(t, err, ErrUserWordNotFound)
			for _, table := range []string{"review_attempts", "idempotency_keys"} {
				var count int
				require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE user_id=$1`, owner).Scan(&count))
				require.Zero(t, count)
			}
		})
	}
}
