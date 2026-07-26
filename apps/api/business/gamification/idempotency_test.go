package gamification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdempotencyKeyShapes(t *testing.T) {
	// Per-source-event keys must be deterministic and uniquely identify the
	// triggering row.
	assert.Equal(t, "user_word:abc:added", string(UserWordAddedKey("abc")))
	assert.Equal(t, "review_attempt:def:rated", string(ReviewAttemptRatedKey("def")))
	assert.Equal(t, "learner_sentence:ghi:submitted", string(LearnerSentenceSubmittedKey("ghi")))
	assert.Equal(t, "ai_feedback_attempt:jkl:received", string(AIFeedbackAttemptReceivedKey("jkl")))
	assert.Equal(t, "daily_mission:u1:2026-07-26:completed", string(DailyMissionCompletedKey("u1", "2026-07-26")))
	assert.Equal(t, "streak:u1:2026-07-26:grace_day_earned", string(StreakGraceDayEarnedKey("u1", "2026-07-26")))
	assert.Equal(t, "streak:u1:2026-07-25:grace_day_used", string(StreakGraceDayUsedKey("u1", "2026-07-25")))
}
