package gamification

// Confidence point and grace-day reward values per DOC-06 §11 and DOC-00 §3.
// These live in backend config (this file) and are not a database constraint
// (DOC-05 §12). Centralizing them as exported constants makes the values
// stable for tests, transactions, and analytics tagging.
const (
	RewardAddWord           = 2
	RewardReviewAgain       = 1
	RewardReviewHard        = 2
	RewardReviewGood        = 5
	RewardReviewEasy        = 6
	RewardDailyMissionDone  = 10
	RewardSentenceSubmitted = 3
	RewardAIFeedbackGot     = 2
)

// Reason and source_type enum values for confidence_point_ledger. These match
// DOC-05 §12 with the D02 reconciliation: `word_added` reason and `user_word`
// source_type were added for the Add word reward. Other P1/P2/P3 rewards
// reuse `review_correct` and the corresponding source_type.
const (
	ReasonWordAdded             = "word_added"
	ReasonReviewCorrect         = "review_correct"
	ReasonDailyMissionCompleted = "daily_mission_completed"
	ReasonSentenceSubmitted     = "sentence_submitted"
	ReasonAIFeedbackReceived    = "ai_feedback_received"
	ReasonStreakBonus           = "streak_bonus"
	ReasonAdminAdjustment       = "admin_adjustment"

	SourceUserWord          = "user_word"
	SourceReviewAttempt     = "review_attempt"
	SourceDailyMission      = "daily_mission"
	SourceLearnerSentence   = "learner_sentence"
	SourceAIFeedbackAttempt = "ai_feedback_attempt"
	SourceStreak            = "streak"
	SourceAdmin             = "admin"
)

// grace_day_ledger reason and source_type values per DOC-05 §12.
const (
	GraceReasonEarnedByStreak   = "earned_by_streak"
	GraceReasonManualGrant      = "manual_grant"
	GraceReasonUsedForMissedDay = "used_for_missed_day"
	GraceReasonExpired          = "expired"
	GraceReasonAdminAdjustment  = "admin_adjustment"

	GraceSourceDailyMission = "daily_mission"
	GraceSourceStreak       = "streak"
	GraceSourceAdmin        = "admin"
)

// streak_states status values per DOC-05 §12.
const (
	StreakStatusActive = "active"
	StreakStatusAtRisk = "at_risk"
	StreakStatusBroken = "broken"
)

// daily_mission_snapshots status values per DOC-05 §10.
const (
	MissionStatusOpen      = "open"
	MissionStatusCompleted = "completed"
	MissionStatusMissed    = "missed"
	MissionStatusProtected = "protected"
)

// Policy version emitted on snapshots/activity rows created by this milestone.
const MissionPolicyVersion = "p4-mission-policy-v1"

// Grace day configuration per DOC-00 §3: one grace day earned per seven
// completed days, capped at a balance of 2.
const (
	GraceDayEarnEveryCompletedDays = 7
	GraceDayMaxBalance             = 2
)
