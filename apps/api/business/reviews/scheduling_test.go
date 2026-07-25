package reviews

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func zeroCounters() ReviewCounters {
	return ReviewCounters{}
}

func TestApplyReviewAgainStepsBack(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 5, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultIncorrect,
		Rating: RatingAgain,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 4, res.ReviewStep)
	assert.Equal(t, now.Add(StepInterval(4)), res.NextReviewAt)
	assert.Equal(t, 1, res.Counters.TotalReviewCount)
	assert.Equal(t, 0, res.Counters.CorrectReviewCount)
	assert.Equal(t, 0, res.Counters.ConsecutiveCorrectCount)
	assert.Equal(t, 1, res.Counters.ConsecutiveIncorrectCount)
	assert.Equal(t, ResultIncorrect, res.LastResult)
	assert.Equal(t, RatingAgain, res.LastRating)
	assert.True(t, res.LastReviewedAt.Equal(now))
}

func TestApplyReviewAgainFloorsAtZero(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 0, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultIncorrect,
		Rating: RatingAgain,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 0, res.ReviewStep)
	assert.Equal(t, now.Add(StepInterval(0)), res.NextReviewAt)
}

func TestApplyReviewHardKeepsStep(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 3, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: RatingHard,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 3, res.ReviewStep)
	assert.Equal(t, 1, res.Counters.TotalReviewCount)
	assert.Equal(t, 1, res.Counters.CorrectReviewCount)
	assert.Equal(t, 1, res.Counters.ConsecutiveCorrectCount)
	assert.Equal(t, 0, res.Counters.ConsecutiveIncorrectCount)
}

func TestApplyReviewGoodStepsForward(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 5, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: RatingGood,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 6, res.ReviewStep)
	assert.Equal(t, now.Add(StepInterval(6)), res.NextReviewAt)
}

func TestApplyReviewEasyCapsAtSeven(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 7, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: RatingEasy,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 7, res.ReviewStep)
	assert.Equal(t, now.Add(StepInterval(7)), res.NextReviewAt)
}

func TestApplyReviewTwoConsecutiveIncorrectResetsStep(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	state := ReviewState{ReviewStep: 4, Counters: ReviewCounters{
		TotalReviewCount:          3,
		CorrectReviewCount:        2,
		ConsecutiveCorrectCount:   1,
		ConsecutiveIncorrectCount: 0,
	}}

	first, err := ApplyReview(state, ApplyReviewRequest{
		Result: ResultIncorrect,
		Rating: RatingAgain,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 3, first.ReviewStep)
	assert.Equal(t, 1, first.Counters.ConsecutiveIncorrectCount)

	second, err := ApplyReview(ReviewState{ReviewStep: first.ReviewStep, Counters: first.Counters}, ApplyReviewRequest{
		Result: ResultIncorrect,
		Rating: RatingAgain,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 0, second.ReviewStep)
	assert.Equal(t, 2, second.Counters.ConsecutiveIncorrectCount)
	assert.Equal(t, 5, second.Counters.TotalReviewCount)
	assert.Equal(t, 2, second.Counters.CorrectReviewCount)
}

func TestApplyReviewCorrectResetsConsecutiveIncorrect(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	state := ReviewState{ReviewStep: 2, Counters: ReviewCounters{
		ConsecutiveIncorrectCount: 1,
	}}
	res, err := ApplyReview(state, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: RatingGood,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 3, res.ReviewStep)
	assert.Equal(t, 0, res.Counters.ConsecutiveIncorrectCount)
	assert.Equal(t, 1, res.Counters.ConsecutiveCorrectCount)
}

func TestApplyReviewCorrectWithoutRatingDefaultsToGood(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	res, err := ApplyReview(ReviewState{ReviewStep: 2, Counters: zeroCounters()}, ApplyReviewRequest{
		Result: ResultCorrect,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 3, res.ReviewStep)
	assert.Equal(t, RatingGood, res.LastRating)
	assert.Equal(t, 1, res.Counters.CorrectReviewCount)
}

func TestApplyReviewSkippedKeepsStepAndResetsStreaks(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	state := ReviewState{ReviewStep: 4, Counters: ReviewCounters{
		ConsecutiveCorrectCount: 2,
	}}
	res, err := ApplyReview(state, ApplyReviewRequest{
		Result: ResultSkipped,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 4, res.ReviewStep)
	assert.Equal(t, 0, res.Counters.ConsecutiveCorrectCount)
	assert.Equal(t, 0, res.Counters.ConsecutiveIncorrectCount)
	assert.Equal(t, 1, res.Counters.TotalReviewCount)
	assert.Equal(t, 0, res.Counters.CorrectReviewCount)
}

func TestApplyReviewCorrectCountNeverExceedsTotal(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	state := ReviewState{ReviewStep: 1, Counters: ReviewCounters{
		TotalReviewCount:   10,
		CorrectReviewCount: 5,
	}}
	res, err := ApplyReview(state, ApplyReviewRequest{
		Result: ResultIncorrect,
		Rating: RatingAgain,
	}, now)
	require.NoError(t, err)
	assert.Equal(t, 11, res.Counters.TotalReviewCount)
	assert.Equal(t, 5, res.Counters.CorrectReviewCount)
	assert.True(t, res.Counters.CorrectReviewCount <= res.Counters.TotalReviewCount)
}

func TestApplyReviewInvalidStep(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	_, err := ApplyReview(ReviewState{ReviewStep: 8}, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: RatingGood,
	}, now)
	assert.ErrorIs(t, err, ErrInvalidReviewStep)
}

func TestApplyReviewInvalidResult(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	_, err := ApplyReview(ReviewState{ReviewStep: 0}, ApplyReviewRequest{
		Result: "unknown",
	}, now)
	assert.ErrorIs(t, err, ErrInvalidResult)
}

func TestApplyReviewInvalidRating(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	_, err := ApplyReview(ReviewState{ReviewStep: 0}, ApplyReviewRequest{
		Result: ResultCorrect,
		Rating: "great",
	}, now)
	assert.ErrorIs(t, err, ErrInvalidRating)
}

func TestStepIntervalForEveryStep(t *testing.T) {
	// Sanity check that every step has a positive interval.
	for i := 0; i <= 7; i++ {
		assert.Positive(t, StepInterval(i))
	}
	assert.Equal(t, StepInterval(0), 10*time.Minute)
	assert.Equal(t, StepInterval(1), 1*time.Hour)
}
