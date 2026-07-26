package gamification

import "fmt"

// PointIdempotencyKey is the deterministic per-source-event idempotency key
// written to confidence_point_ledger. The unique partial index
// (user_id, idempotency_key) where idempotency_key is not null is the second
// line of defense (after each module's own guard) that a retried or re-entered
// transaction cannot award the same reward twice.
type PointIdempotencyKey string

// GraceIdempotencyKey is the deterministic per-source-event idempotency key
// written to grace_day_ledger, with the same uniqueness guarantee.
type GraceIdempotencyKey string

// String lets the key types be passed straight to a database driver parameter.
func (k PointIdempotencyKey) String() string { return string(k) }
func (k GraceIdempotencyKey) String() string { return string(k) }

// All canonical per-source-event idempotency key derivations. They are
// deterministic from the triggering row's own ID (or, for daily-mission
// completion, the (user_id, local_date) pair) so a retried or re-entered
// transaction can never award the same reward twice.
const (
	IdemKeyUserWordAdded         = "user_word:%s:added"
	IdemKeyReviewAttemptRated    = "review_attempt:%s:rated"
	IdemKeyLearnerSentenceSubmit = "learner_sentence:%s:submitted"
	IdemKeyAIFeedbackAttemptGot  = "ai_feedback_attempt:%s:received"
	IdemKeyDailyMissionCompleted = "daily_mission:%s:%s:completed"
	IdemKeyStreakGraceDayEarned  = "streak:%s:%s:grace_day_earned"
	IdemKeyStreakGraceDayUsed    = "streak:%s:%s:grace_day_used"
)

// UserWordAddedKey returns the per-user-word idempotency key for the +2
// add-word reward. id is the user_words row UUID.
func UserWordAddedKey(id string) PointIdempotencyKey {
	return PointIdempotencyKey(fmt.Sprintf(IdemKeyUserWordAdded, id))
}

// ReviewAttemptRatedKey returns the per-review-attempt idempotency key for
// the rating-tiered review reward. id is the review_attempts row UUID.
func ReviewAttemptRatedKey(id string) PointIdempotencyKey {
	return PointIdempotencyKey(fmt.Sprintf(IdemKeyReviewAttemptRated, id))
}

// LearnerSentenceSubmittedKey returns the per-learner-sentence idempotency
// key for the +3 sentence-submitted reward. id is the learner_sentences row
// UUID.
func LearnerSentenceSubmittedKey(id string) PointIdempotencyKey {
	return PointIdempotencyKey(fmt.Sprintf(IdemKeyLearnerSentenceSubmit, id))
}

// AIFeedbackAttemptReceivedKey returns the per-attempt idempotency key for
// the +2 AI-feedback-received reward. id is the ai_feedback_attempts row
// UUID.
func AIFeedbackAttemptReceivedKey(id string) PointIdempotencyKey {
	return PointIdempotencyKey(fmt.Sprintf(IdemKeyAIFeedbackAttemptGot, id))
}

// DailyMissionCompletedKey returns the per-(user,local_date) idempotency
// key for the +10 daily-mission-complete reward. localDate is the
// date-only representation (YYYY-MM-DD).
func DailyMissionCompletedKey(userID, localDate string) PointIdempotencyKey {
	return PointIdempotencyKey(fmt.Sprintf(IdemKeyDailyMissionCompleted, userID, localDate))
}

// StreakGraceDayEarnedKey returns the per-(user,local_date) idempotency
// key for the +1 grace day earned by streak.
func StreakGraceDayEarnedKey(userID, localDate string) GraceIdempotencyKey {
	return GraceIdempotencyKey(fmt.Sprintf(IdemKeyStreakGraceDayEarned, userID, localDate))
}

// StreakGraceDayUsedKey returns the per-(user,local_date) idempotency key
// for the -1 grace day used to protect a missed day.
func StreakGraceDayUsedKey(userID, localDate string) GraceIdempotencyKey {
	return GraceIdempotencyKey(fmt.Sprintf(IdemKeyStreakGraceDayUsed, userID, localDate))
}
