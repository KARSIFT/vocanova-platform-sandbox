package aifeedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type nilFeedbackSequenceProvider struct {
	nilFirst    bool
	validRepair bool
	tasks       []ProviderTask
}

func (p *nilFeedbackSequenceProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	p.tasks = append(p.tasks, task)
	if len(p.tasks) == 1 {
		if p.nilFirst {
			return nil, nil
		}
		return &ProviderFeedback{}, nil
	}
	if p.validRepair {
		return NewMockProvider().GenerateFeedback(ctx, task)
	}
	return nil, nil
}

func TestServiceNilFeedbackUsesBoundedRepair(t *testing.T) {
	for _, tc := range []struct {
		name        string
		nilFirst    bool
		validRepair bool
	}{
		{"nil output repairs to valid feedback", true, true},
		{"nil output and nil repair settle retryable failure", true, false},
		{"invalid output and nil repair settle retryable failure", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newServiceFixture(t)
			provider := &nilFeedbackSequenceProvider{nilFirst: tc.nilFirst, validRepair: tc.validRepair}
			mission := &countedMissionAccounting{}
			f.service.provider = provider
			f.service.mission = mission
			req := f.request("I work every day.")
			var result *SentenceFeedbackResult
			var err error
			require.NotPanics(t, func() {
				result, err = f.service.SubmitSentenceFeedback(t.Context(), req)
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, provider.tasks, 2, "one initial call and at most one constrained repair")
			require.NotContains(t, provider.tasks[0].UserPayload, "repair_attempt")
			require.Equal(t, true, provider.tasks[1].UserPayload["repair_attempt"])
			require.Equal(t, normalizeSentence(req.SentenceText), provider.tasks[1].UserPayload["learner_sentence"])
			if tc.nilFirst {
				prior, exists := provider.tasks[1].UserPayload["prior_output"]
				require.True(t, exists)
				require.Nil(t, prior, "missing output is represented honestly, without fabricated feedback")
				require.Equal(t, "feedback is nil", provider.tasks[1].UserPayload["validation_error"])
			}
			require.Len(t, f.repo.attempts, 1)
			require.Len(t, f.repo.sentences, 1)
			require.Equal(t, result.AttemptID, f.repo.attempts[0].ID)
			require.Equal(t, result.SentenceID, f.repo.sentences[0].ID)
			require.Equal(t, req.SentenceText, result.OriginalSentence)
			require.Equal(t, req.SentenceText, f.repo.sentences[0].SentenceText)
			if tc.validRepair {
				require.Empty(t, result.ErrorCode)
				require.Equal(t, ProcessingStatusCompleted, result.ProcessingStatus)
				require.Equal(t, LearningStatusCorrect, result.Status)
				require.Equal(t, AttemptStatusSucceeded, f.repo.attempts[0].Status)
				require.Equal(t, SentenceStatusFeedbackReady, f.repo.sentences[0].Status)
				require.NotEmpty(t, f.repo.attempts[0].FeedbackJSON)
				require.Equal(t, 1, mission.calls)
				replay, err := f.service.SubmitSentenceFeedback(t.Context(), req)
				require.NoError(t, err)
				require.Equal(t, result.AttemptID, replay.AttemptID)
				require.Equal(t, LearningStatusCorrect, replay.Status)
				require.Len(t, provider.tasks, 2, "replay must not generate again")
				require.Equal(t, 1, mission.calls, "replay must not account again")
				return
			}
			require.Equal(t, ErrorCodeTemporaryFailure, result.ErrorCode)
			require.True(t, result.CanRetry)
			require.False(t, result.MissionCompleted)
			require.Equal(t, ProcessingStatusFailed, result.ProcessingStatus)
			require.Equal(t, AttemptStatusFailed, f.repo.attempts[0].Status)
			require.Equal(t, ErrorCodeTemporaryFailure, f.repo.attempts[0].ErrorCode)
			require.Equal(t, SentenceStatusFeedbackFailed, f.repo.sentences[0].Status)
			require.Empty(t, f.repo.attempts[0].FeedbackJSON)
			require.Zero(t, mission.calls, "failed output must not receive mission credit")
		})
	}
}
