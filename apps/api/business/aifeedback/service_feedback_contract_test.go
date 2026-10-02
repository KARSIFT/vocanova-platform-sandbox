package aifeedback

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type selectedSenseFeedbackProvider struct {
	calls    int
	tasks    []ProviderTask
	feedback *ProviderFeedback
}

func (p *selectedSenseFeedbackProvider) GenerateFeedback(_ context.Context, task ProviderTask) (*ProviderFeedback, error) {
	p.calls++
	p.tasks = append(p.tasks, task)
	return p.feedback, nil
}

func TestServiceWrongSelectedSenseWithoutCorrectionFinalizesOnce(t *testing.T) {
	f := newServiceFixture(t)
	tip := "Write a sentence about a person doing a job or task."
	provider := &selectedSenseFeedbackProvider{feedback: &ProviderFeedback{
		Status:                  LearningStatusIncorrect,
		TargetWordUsedCorrectly: false,
		GrammarAcceptable:       true,
		MeaningClear:            true,
		Naturalness:             NaturalnessNatural,
		CorrectedSentence:       nil,
		Headline:                "Try the selected meaning",
		Explanation:             "Your sentence is grammatical, but work means function here, rather than do a job or task.",
		ImprovementTip:          &tip,
	}}
	provider.feedback.RawJSON = provider.feedback.StructuredJSON()
	provider.feedback.RawJSON["corrected_sentence"] = nil
	f.service.provider = provider
	mission := &countedMissionAccounting{}
	f.service.mission = mission
	finalizations := 0
	f.service.repo = &finalizerAwareRepository{Repository: f.repo, finalize: func(ctx context.Context, pending PendingAttempt, feedback *ProviderFeedback, now time.Time, completion SuccessfulFeedbackCompletion) (bool, error) {
		finalizations++
		require.Nil(t, feedback.CorrectedSentence)
		return f.repo.CompleteSuccessfulFeedbackAttempt(ctx, pending, feedback, now, completion)
	}}
	req := f.request("The old clock does not work.")
	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Empty(t, result.ErrorCode)
	require.Equal(t, ProcessingStatusCompleted, result.ProcessingStatus)
	require.Equal(t, LearningStatusIncorrect, result.Status)
	require.False(t, result.TargetWordUsedCorrectly)
	require.True(t, result.GrammarAcceptable)
	require.True(t, result.MeaningClear)
	require.Equal(t, NaturalnessNatural, result.Naturalness)
	require.Nil(t, result.CorrectedSentence, "do not fabricate a different meaning as a correction")
	require.Equal(t, req.SentenceText, result.OriginalSentence)
	require.Equal(t, provider.feedback.Explanation, result.Explanation)
	require.Equal(t, &tip, result.ImprovementTip)
	require.False(t, result.CanRetry)
	require.Equal(t, 1, provider.calls, "contract-valid feedback must not require repair")
	require.Equal(t, "to do a job or task", provider.tasks[0].UserPayload["target_meaning"])
	require.Equal(t, 1, finalizations, "use the normal successful-attempt finalizer")
	require.Equal(t, 1, mission.calls)
	require.Len(t, f.repo.attempts, 1)
	require.Len(t, f.repo.sentences, 1)
	require.Equal(t, AttemptStatusSucceeded, f.repo.attempts[0].Status)
	require.Equal(t, SentenceStatusFeedbackReady, f.repo.sentences[0].Status)
	require.Equal(t, LearningStatusIncorrect, f.repo.attempts[0].FeedbackJSON["status"])
	require.Equal(t, tip, f.repo.attempts[0].FeedbackJSON["improvement_tip"])
	require.NotContains(t, f.repo.attempts[0].FeedbackJSON, "corrected_sentence", "nil correction remains absent in stored feedback")

	replay, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, result.AttemptID, replay.AttemptID)
	require.Equal(t, result.SentenceID, replay.SentenceID)
	require.Equal(t, ProcessingStatusCompleted, replay.ProcessingStatus)
	require.Equal(t, LearningStatusIncorrect, replay.Status)
	require.Nil(t, replay.CorrectedSentence)
	require.Equal(t, result.Explanation, replay.Explanation)
	require.Equal(t, result.ImprovementTip, replay.ImprovementTip)
	require.Equal(t, 1, provider.calls)
	require.Equal(t, 1, finalizations)
	require.Equal(t, 1, mission.calls, "idempotent replay must not duplicate mission accounting")
	require.Len(t, f.repo.attempts, 1)
}
