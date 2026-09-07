//go:build integration

package aifeedback_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreatePendingAttemptRejectsRemovalBetweenPreflightAndWrite proves the
// critical interleaving against every committed migration. LoadTarget is a
// deliberately separate preflight transaction; the removal commits before the
// pending-write transaction starts, exactly the window that previously let a
// stale Review Completion submission persist a new generation.
func TestCreatePendingAttemptRejectsRemovalBetweenPreflightAndWrite(t *testing.T) {
	db := missionAccountingDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	userID, wordID, meaningID, userWordID, reviewID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, seedReviewFeedbackTarget(ctx, db, userID, wordID, meaningID, userWordID, reviewID, now))

	repo := aifeedback.NewPostgreSQLRepository(db, nil)
	req := aifeedback.SubmitSentenceFeedbackRequest{
		UserID: userID, Source: aifeedback.SourceReview, AttemptID: reviewID,
		SentenceText: "I work every day.", IdempotencyKey: "removed-target-race",
	}
	target, err := repo.LoadTarget(ctx, aifeedback.LoadTargetRequest{UserID: userID, Source: req.Source, AttemptID: req.AttemptID})
	require.NoError(t, err)

	// This commits after the preflight but before the pending-write claim.
	_, err = db.ExecContext(ctx, `UPDATE user_words SET deleted_at = $1, updated_at = $1 WHERE id = $2`, now, userWordID)
	require.NoError(t, err)
	_, err = repo.CreatePendingAttempt(ctx, req, target, "i work every day.", "removed-target-race-hash", aifeedback.ProviderMock, "mock", now)
	assert.ErrorIs(t, err, aifeedback.ErrTargetNotFound)

	var sentences, attempts int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM learner_sentences`).Scan(&sentences))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM ai_feedback_attempts`).Scan(&attempts))
	assert.Zero(t, sentences)
	assert.Zero(t, attempts)
}

func seedReviewFeedbackTarget(ctx context.Context, db *sql.DB, userID, wordID, meaningID, userWordID, reviewID uuid.UUID, now time.Time) error {
	// Kept as individual statements so failures identify the schema contract
	// that changed rather than masking it behind one large fixture string.
	queries := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, []any{userID, userID.String() + "@example.test", now}},
		{`INSERT INTO canonical_words (id, text, normalized_text, word_type, language_code, status, difficulty_level, created_at, updated_at) VALUES ($1, 'work', 'work', 'word', 'en', 'active', 'a2', $2, $2)`, []any{wordID, now}},
		{`INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, meaning_order, status, created_at, updated_at) VALUES ($1, $2, 'verb', 'do a job', 1, 'active', $3, $3)`, []any{meaningID, wordID, now}},
		{`INSERT INTO user_words (id, user_id, meaning_id, status, source, added_at, created_at, updated_at) VALUES ($1, $2, $3, 'learning', 'manual', $4, $4, $4)`, []any{userWordID, userID, meaningID, now}},
		{`INSERT INTO review_attempts (id, user_id, user_word_id, meaning_id, attempt_type, prompt_type, result, rating, review_step_before, review_step_after, answered_at, source, created_at, updated_at) VALUES ($1, $2, $3, $4, 'review', 'multiple_choice', 'correct', 'good', 0, 1, $5, 'review', $5, $5)`, []any{reviewID, userID, userWordID, meaningID, now}},
	}
	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q.query, q.args...); err != nil {
			return err
		}
	}
	return nil
}
