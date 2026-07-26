package gamification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewardForAllKinds(t *testing.T) {
	cases := []struct {
		kind       RewardKind
		amount     int
		reason     string
		sourceType string
	}{
		{RewardKindAddWord, 2, ReasonWordAdded, SourceUserWord},
		{RewardKindReviewAgain, 1, ReasonReviewCorrect, SourceReviewAttempt},
		{RewardKindReviewHard, 2, ReasonReviewCorrect, SourceReviewAttempt},
		{RewardKindReviewGood, 5, ReasonReviewCorrect, SourceReviewAttempt},
		{RewardKindReviewEasy, 6, ReasonReviewCorrect, SourceReviewAttempt},
		{RewardKindDailyMissionDone, 10, ReasonDailyMissionCompleted, SourceDailyMission},
		{RewardKindSentenceSubmitted, 3, ReasonSentenceSubmitted, SourceLearnerSentence},
		{RewardKindAIFeedbackGot, 2, ReasonAIFeedbackReceived, SourceAIFeedbackAttempt},
	}
	for _, tc := range cases {
		out, err := RewardFor(tc.kind)
		require.NoError(t, err, "kind %s", tc.kind)
		assert.Equal(t, tc.amount, out.Amount, "kind %s amount", tc.kind)
		assert.Equal(t, tc.reason, out.Reason, "kind %s reason", tc.kind)
		assert.Equal(t, tc.sourceType, out.SourceType, "kind %s source_type", tc.kind)
	}
}

func TestRewardForUnknownKind(t *testing.T) {
	_, err := RewardFor(RewardKind("not_a_real_kind"))
	require.ErrorIs(t, err, ErrUnknownRewardKind)
}

func TestRewardValuesMatchDoc06S11(t *testing.T) {
	// This is the DOC-06 §11 reward table the migration depends on.
	assert.Equal(t, 2, RewardAddWord)
	assert.Equal(t, 1, RewardReviewAgain)
	assert.Equal(t, 2, RewardReviewHard)
	assert.Equal(t, 5, RewardReviewGood)
	assert.Equal(t, 6, RewardReviewEasy)
	assert.Equal(t, 10, RewardDailyMissionDone)
	assert.Equal(t, 3, RewardSentenceSubmitted)
	assert.Equal(t, 2, RewardAIFeedbackGot)
}
