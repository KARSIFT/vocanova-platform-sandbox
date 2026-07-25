// Package reviews implements the spaced-repetition review domain and persistence
// boundaries. It is intentionally independent of Huma/chi and of direct Ent writes
// inside the scheduling core; the transaction layer applies its output.
package reviews

import (
	"errors"
	"fmt"
	"time"
)

// Result values for a review attempt.
const (
	ResultCorrect   = "correct"
	ResultIncorrect = "incorrect"
	ResultSkipped   = "skipped"
)

// Rating values from the learner (SRS buttons).
const (
	RatingAgain = "again"
	RatingHard  = "hard"
	RatingGood  = "good"
	RatingEasy  = "easy"
)

// Scheduling errors.
var (
	ErrInvalidReviewStep = errors.New("review step must be between 0 and 7")
	ErrInvalidResult     = errors.New("result must be correct, incorrect, or skipped")
	ErrInvalidRating     = errors.New("rating must be empty, again, hard, good, or easy")
)

// ReviewCounters tracks the learning-state counters stored on user_words.
type ReviewCounters struct {
	TotalReviewCount          int
	CorrectReviewCount        int
	ConsecutiveCorrectCount   int
	ConsecutiveIncorrectCount int
}

// ReviewState captures the user_words scheduling fields before a review.
type ReviewState struct {
	ReviewStep int
	Counters   ReviewCounters
}

// ApplyReviewRequest captures the outcome of one review attempt.
type ApplyReviewRequest struct {
	Result string
	Rating string // empty means no rating was supplied
}

// ApplyReviewResult captures the updated user_words scheduling fields.
type ApplyReviewResult struct {
	ReviewStep     int
	NextReviewAt   time.Time
	Counters       ReviewCounters
	LastReviewedAt time.Time
	LastResult     string
	LastRating     string
}

// StepInterval returns the backend-owned interval for a review step.
// The mapping is deterministic and owned by the backend; callers must not invent
// their own intervals.
func StepInterval(step int) time.Duration {
	switch step {
	case 0:
		return 10 * time.Minute
	case 1:
		return 1 * time.Hour
	case 2:
		return 24 * time.Hour
	case 3:
		return 3 * 24 * time.Hour
	case 4:
		return 7 * 24 * time.Hour
	case 5:
		return 14 * 24 * time.Hour
	case 6:
		return 30 * 24 * time.Hour
	case 7:
		return 60 * 24 * time.Hour
	default:
		return 10 * time.Minute
	}
}

// ApplyReview computes the new user_words scheduling state from the prior state
// and a single review outcome. It is pure (no IO, no Ent writes) and deterministic
// so the transaction layer can call it exactly once per submission.
//
// Rules per DOC-06 §10:
//   - Again / incorrect → step back, floor 0.
//   - Two consecutive Again / incorrect attempts → reset to step 0.
//   - Hard → same step.
//   - Good / Easy → step forward, cap 7.
//   - Counters are updated and next_review_at is derived from the new step.
func ApplyReview(prior ReviewState, req ApplyReviewRequest, now time.Time) (ApplyReviewResult, error) {
	if prior.ReviewStep < 0 || prior.ReviewStep > 7 {
		return ApplyReviewResult{}, fmt.Errorf("%w: got %d", ErrInvalidReviewStep, prior.ReviewStep)
	}
	if req.Result != ResultCorrect && req.Result != ResultIncorrect && req.Result != ResultSkipped {
		return ApplyReviewResult{}, fmt.Errorf("%w: %q", ErrInvalidResult, req.Result)
	}
	if req.Rating != "" && req.Rating != RatingAgain && req.Rating != RatingHard && req.Rating != RatingGood && req.Rating != RatingEasy {
		return ApplyReviewResult{}, fmt.Errorf("%w: %q", ErrInvalidRating, req.Rating)
	}

	counters := prior.Counters
	counters.TotalReviewCount++

	newStep := prior.ReviewStep
	lastResult := req.Result
	lastRating := req.Rating

	// Treat a rating of Again as incorrect for scheduling, even if a caller
	// somehow passed result=correct. This is a defensive guard; the API layer
	// rejects objective-incorrect answers that are not Again.
	isIncorrect := req.Result == ResultIncorrect || req.Rating == RatingAgain

	switch {
	case req.Result == ResultSkipped:
		// Skipped attempts do not move the schedule; the word stays where it was
		// and the learner's consecutive streaks are reset.
		newStep = prior.ReviewStep
		counters.ConsecutiveCorrectCount = 0
		counters.ConsecutiveIncorrectCount = 0
	case isIncorrect:
		counters.ConsecutiveCorrectCount = 0
		counters.ConsecutiveIncorrectCount++
		if counters.ConsecutiveIncorrectCount >= 2 {
			newStep = 0
		} else {
			newStep = max(0, prior.ReviewStep-1)
		}
	default:
		// Correct: hard keeps the step, good/easy advances, empty rating defaults
		// to good.
		counters.CorrectReviewCount++
		counters.ConsecutiveIncorrectCount = 0
		counters.ConsecutiveCorrectCount++
		rating := req.Rating
		if rating == "" {
			rating = RatingGood
			lastRating = RatingGood
		}
		switch rating {
		case RatingHard:
			newStep = prior.ReviewStep
		case RatingGood, RatingEasy:
			newStep = min(7, prior.ReviewStep+1)
		}
	}

	return ApplyReviewResult{
		ReviewStep:     newStep,
		NextReviewAt:   now.Add(StepInterval(newStep)),
		Counters:       counters,
		LastReviewedAt: now,
		LastResult:     lastResult,
		LastRating:     lastRating,
	}, nil
}
