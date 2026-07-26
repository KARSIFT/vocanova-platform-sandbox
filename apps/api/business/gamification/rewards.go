package gamification

import "errors"

// RewardKind enumerates the per-source-event reward values. Each maps to the
// reason/source_type pair written to confidence_point_ledger.
type RewardKind string

const (
	RewardKindAddWord           RewardKind = "add_word"
	RewardKindReviewAgain       RewardKind = "review_again"
	RewardKindReviewHard        RewardKind = "review_hard"
	RewardKindReviewGood        RewardKind = "review_good"
	RewardKindReviewEasy        RewardKind = "review_easy"
	RewardKindDailyMissionDone  RewardKind = "daily_mission_done"
	RewardKindSentenceSubmitted RewardKind = "sentence_submitted"
	RewardKindAIFeedbackGot     RewardKind = "ai_feedback_got"
)

// RewardOutcome is the pure, deterministic computation result of a single
// reward grant. The transaction layer is responsible for materializing it
// into a confidence_point_ledger insert.
type RewardOutcome struct {
	Amount     int
	Reason     string
	SourceType string
}

// ErrUnknownRewardKind is returned when an unrecognized reward kind is
// requested. This is a programming-error guard; the package never silently
// returns 0 points for an unknown kind.
var ErrUnknownRewardKind = errors.New("unknown reward kind")

// RewardFor returns the deterministic reward outcome for a single triggering
// event. It is pure (no IO) and matches DOC-06 §11 exactly.
func RewardFor(kind RewardKind) (RewardOutcome, error) {
	switch kind {
	case RewardKindAddWord:
		return RewardOutcome{Amount: RewardAddWord, Reason: ReasonWordAdded, SourceType: SourceUserWord}, nil
	case RewardKindReviewAgain:
		return RewardOutcome{Amount: RewardReviewAgain, Reason: ReasonReviewCorrect, SourceType: SourceReviewAttempt}, nil
	case RewardKindReviewHard:
		return RewardOutcome{Amount: RewardReviewHard, Reason: ReasonReviewCorrect, SourceType: SourceReviewAttempt}, nil
	case RewardKindReviewGood:
		return RewardOutcome{Amount: RewardReviewGood, Reason: ReasonReviewCorrect, SourceType: SourceReviewAttempt}, nil
	case RewardKindReviewEasy:
		return RewardOutcome{Amount: RewardReviewEasy, Reason: ReasonReviewCorrect, SourceType: SourceReviewAttempt}, nil
	case RewardKindDailyMissionDone:
		return RewardOutcome{Amount: RewardDailyMissionDone, Reason: ReasonDailyMissionCompleted, SourceType: SourceDailyMission}, nil
	case RewardKindSentenceSubmitted:
		return RewardOutcome{Amount: RewardSentenceSubmitted, Reason: ReasonSentenceSubmitted, SourceType: SourceLearnerSentence}, nil
	case RewardKindAIFeedbackGot:
		return RewardOutcome{Amount: RewardAIFeedbackGot, Reason: ReasonAIFeedbackReceived, SourceType: SourceAIFeedbackAttempt}, nil
	default:
		return RewardOutcome{}, ErrUnknownRewardKind
	}
}
