package reviews

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepositorySubmitReview(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now})

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	attemptID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-1", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(1, sqlmock.AnyArg(), now, "correct", "good", 1, 1, 1, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-1",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.NotEqual(t, attemptID, attempt.ID) // random uuid
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	assert.Equal(t, now.Add(time.Hour), attempt.NextReviewAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositorySubmitReviewUserWordNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, clock.Fixed{T: time.Now()})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}))
	mock.ExpectRollback()

	_, err = repo.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      time.Now().UTC(),
		ClientAttemptID: "ca",
	})
	assert.ErrorIs(t, err, ErrUserWordNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositorySubmitReviewMeaningMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, clock.Fixed{T: time.Now()})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	storedMeaningID := uuid.MustParse("00000000-0000-0000-0000-000000000099")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, storedMeaningID, 0, 0, 0, 0))
	mock.ExpectRollback()

	_, err = repo.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      time.Now().UTC(),
		ClientAttemptID: "ca",
	})
	assert.ErrorIs(t, err, ErrUserWordNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositorySubmitReviewIdempotencyConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now})

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	attemptID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}).
			AddRow(attemptID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-1", nil, now))
	mock.ExpectRollback()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		Source:          SourceReview,
		ClientAttemptID: "ca-1",
	}
	// Same client_attempt_id with same body is idempotent, so alter the rating.
	req.Rating = RatingEasy
	_, err = repo.SubmitReview(t.Context(), req)
	assert.ErrorIs(t, err, ErrIdempotencyConflict)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRatingToRewardKind verifies the rating-to-reward-kind conversion.
func TestRatingToRewardKind(t *testing.T) {
	tests := []struct {
		rating    string
		expected  gamification.RewardKind
	}{
		{RatingAgain, gamification.RewardKindReviewAgain},
		{RatingHard, gamification.RewardKindReviewHard},
		{RatingGood, gamification.RewardKindReviewGood},
		{RatingEasy, gamification.RewardKindReviewEasy},
		{"", gamification.RewardKindReviewGood},
	}
	for _, tc := range tests {
		t.Run(tc.rating, func(t *testing.T) {
			kind := ratingToRewardKind(tc.rating)
			assert.Equal(t, tc.expected, kind)
		})
	}
}

// TestRewardKindToAmount verifies the reward-kind-to-amount conversion.
func TestRewardKindToAmount(t *testing.T) {
	tests := []struct {
		kind     gamification.RewardKind
		expected int
	}{
		{gamification.RewardKindReviewAgain, gamification.RewardReviewAgain},
		{gamification.RewardKindReviewHard, gamification.RewardReviewHard},
		{gamification.RewardKindReviewGood, gamification.RewardReviewGood},
		{gamification.RewardKindReviewEasy, gamification.RewardReviewEasy},
		{"unknown", 0},
	}
	for _, tc := range tests {
		t.Run(string(tc.kind), func(t *testing.T) {
			amount := rewardKindToAmount(tc.kind)
			assert.Equal(t, tc.expected, amount)
		})
	}
}
