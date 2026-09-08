package reviews

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestReviewAttemptsRejectMutationExceptScopedAccountPurgePostgreSQL(t *testing.T) {
	db := reviewKeyDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)

	userID, attemptID := seedImmutableReviewAttempt(t, ctx, db, now)
	for _, statement := range []string{
		`UPDATE review_attempts SET source = 'manual_practice' WHERE id = $1`,
		`DELETE FROM review_attempts WHERE id = $1`,
	} {
		_, err := db.ExecContext(ctx, statement, attemptID)
		requireImmutableReviewViolation(t, err)
	}
	// The scoped gate is deletion-only. Account disposal must not be able to
	// rewrite retained review history before it removes it.
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `UPDATE review_attempts SET source = 'manual_practice' WHERE id = $1`, attemptID)
	requireImmutableReviewViolation(t, err)
	require.NoError(t, tx.Rollback())

	counters, err := accounts.NewPostgreSQLRepository(db).AnonymizeUserData(ctx, userID)
	require.NoError(t, err)
	require.EqualValues(t, 1, counters.ReviewAttempts)

	secondUserID, secondAttemptID := seedImmutableReviewAttempt(t, ctx, db, now)
	_, err = db.ExecContext(ctx, `DELETE FROM review_attempts WHERE id = $1`, secondAttemptID)
	requireImmutableReviewViolation(t, err)

	var remaining int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM review_attempts WHERE user_id = $1`, secondUserID).Scan(&remaining))
	require.Equal(t, 1, remaining)
}

func TestReviewAttemptPurgeGateRollsBackPostgreSQL(t *testing.T) {
	db := reviewKeyDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID, attemptID := seedImmutableReviewAttempt(t, ctx, db, now)

	_, err := db.ExecContext(ctx, `
		CREATE FUNCTION vocanova_test_reject_review_attempt_purge()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			RAISE EXCEPTION 'injected review-attempt purge failure';
		END;
		$$;
		CREATE TRIGGER vocanova_test_reject_review_attempt_purge
			AFTER DELETE ON review_attempts
			FOR EACH ROW EXECUTE FUNCTION vocanova_test_reject_review_attempt_purge();
	`)
	require.NoError(t, err)

	_, err = accounts.NewPostgreSQLRepository(db).AnonymizeUserData(ctx, userID)
	require.ErrorContains(t, err, "delete review_attempts")

	var remaining int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM review_attempts WHERE id = $1`, attemptID).Scan(&remaining))
	require.Equal(t, 1, remaining)
	_, err = db.ExecContext(ctx, `DELETE FROM review_attempts WHERE id = $1`, attemptID)
	requireImmutableReviewViolation(t, err)
}

func seedImmutableReviewAttempt(t *testing.T, ctx context.Context, db *sql.DB, now time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()
	userID, wordID, meaningID, userWordID, attemptID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, email, created_at, updated_at)
		VALUES ($1, $2, $3, $3)`, userID, userID.String()+"@example.test", now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO canonical_words (id, text, normalized_text, status, created_at, updated_at)
		VALUES ($1, $2, $2, 'active', $3, $3)`, wordID, wordID.String(), now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at)
		VALUES ($1, $2, 'noun', 'A fixture.', 1, 'active', $3, $3)`, meaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO user_words (id, user_id, meaning_id, source, added_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'manual', $4, $4, $4)`, userWordID, userID, meaningID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO review_attempts (
		id, user_id, user_word_id, meaning_id, attempt_type, prompt_type, result, rating,
		review_step_before, review_step_after, answered_at, source, created_at, updated_at
	) VALUES ($1, $2, $3, $4, 'review', 'self_check', 'correct', 'good', 0, 1, $5, 'daily_review', $5, $5)`,
		attemptID, userID, userWordID, meaningID, now)
	require.NoError(t, err)
	return userID, attemptID
}

func requireImmutableReviewViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr), "want PostgreSQL error, got %v", err)
	require.Equal(t, pq.ErrorCode("55000"), pgErr.Code)
}
