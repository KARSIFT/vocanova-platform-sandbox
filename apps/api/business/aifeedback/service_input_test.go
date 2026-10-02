package aifeedback

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type inputCaptureProvider struct {
	tasks       []ProviderTask
	failFirst   bool
	repairFirst bool
}

func (p *inputCaptureProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	p.tasks = append(p.tasks, task)
	if len(p.tasks) == 1 {
		if p.failFirst {
			return nil, errors.New("synthetic provider failure")
		}
		if p.repairFirst {
			return nil, nil
		}
	}
	return NewMockProvider().GenerateFeedback(ctx, task)
}

func TestServiceProviderInputPreservesOriginalCaseAndWhitespaceBounds(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"  I work in London.  ", "I work in London."},
		{"i work in london.", "i work in london."},
		{"\tＩ  work\n in  London.\t", "I work in London."},
		{"I work with US clients.", "I work with US clients."},
	} {
		t.Run(tc.want, func(t *testing.T) {
			f := newServiceFixture(t)
			p := &inputCaptureProvider{repairFirst: true}
			f.service.provider = p
			req := f.request(tc.input)
			result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
			require.NoError(t, err)
			require.Equal(t, ProcessingStatusCompleted, result.ProcessingStatus)
			require.Len(t, p.tasks, 2)
			for _, task := range p.tasks {
				require.Equal(t, tc.want, task.UserPayload["learner_sentence"])
			}
			require.Equal(t, true, p.tasks[1].UserPayload["repair_attempt"])
			require.Equal(t, tc.input, f.repo.sentences[0].SentenceText, "stored original is not rewritten")
			require.Equal(t, normalizeSentence(tc.input), f.repo.sentences[0].NormalizedSentenceText)
		})
	}
}

func TestServiceFailedRetryUsesFirstStoredOriginalAndCaseEquivalentReplay(t *testing.T) {
	f := newServiceFixture(t)
	p := &inputCaptureProvider{failFirst: true}
	f.service.provider = p
	m := &countedMissionAccounting{}
	f.service.mission = m
	first := f.request("  I work in London.  ")
	failed, err := f.service.SubmitSentenceFeedback(t.Context(), first)
	require.NoError(t, err)
	require.Equal(t, ProcessingStatusFailed, failed.ProcessingStatus)
	replay := first
	replay.SentenceText = "i WORK in LONDON."
	_, err = f.service.SubmitSentenceFeedback(t.Context(), replay)
	require.NoError(t, err)
	require.Len(t, p.tasks, 1, "same-key replay does not regenerate")
	replay.IdempotencyKey = uuid.NewString()
	result, err := f.service.SubmitSentenceFeedback(t.Context(), replay)
	require.NoError(t, err)
	require.Equal(t, ProcessingStatusCompleted, result.ProcessingStatus)
	require.Len(t, p.tasks, 2)
	require.Equal(t, "I work in London.", p.tasks[0].UserPayload["learner_sentence"])
	require.Equal(t, "I work in London.", p.tasks[1].UserPayload["learner_sentence"], "fresh failed retry evaluates persisted first submission, not current case variant")
	require.Len(t, f.repo.sentences, 1)
	require.Len(t, f.repo.attempts, 2)
	require.Equal(t, first.SentenceText, f.repo.sentences[0].SentenceText)
	require.Equal(t, 1, m.calls)
	replay.IdempotencyKey = uuid.NewString()
	replay.SentenceText = "I WORK IN LONDON."
	got, err := f.service.SubmitSentenceFeedback(t.Context(), replay)
	require.NoError(t, err)
	require.Equal(t, result.AttemptID, got.AttemptID)
	require.Len(t, p.tasks, 2)
	require.Equal(t, 1, m.calls, "existing case-equivalence policy cannot duplicate credit")
}

type retryInputUnavailableRepository struct {
	Repository
	sentence *LearnerSentence
	err      error
}

func (r retryInputUnavailableRepository) GetLearnerSentence(context.Context, uuid.UUID, uuid.UUID) (*LearnerSentence, error) {
	return r.sentence, r.err
}

func TestServiceRetryCannotGenerateWhenStoredOriginalUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		sentence *LearnerSentence
		err      error
	}{
		{"read error", nil, errors.New("synthetic stored sentence unavailable")},
		{"absent row", nil, nil},
		{"mismatched original", &LearnerSentence{OriginalSentence: "I work in Paris."}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newServiceFixture(t)
			p := &inputCaptureProvider{failFirst: true}
			f.service.provider = p
			req := f.request("I work in London.")
			_, err := f.service.SubmitSentenceFeedback(t.Context(), req)
			require.NoError(t, err)
			f.service.repo = retryInputUnavailableRepository{Repository: f.repo, sentence: tc.sentence, err: tc.err}
			req.IdempotencyKey = uuid.NewString()
			got, err := f.service.SubmitSentenceFeedback(t.Context(), req)
			require.NoError(t, err)
			require.Equal(t, ErrorCodeTemporaryFailure, got.ErrorCode)
			require.Len(t, p.tasks, 1)
			require.Len(t, f.repo.attempts, 1, "do not allocate a retry with no authoritative original")
		})
	}
}

func TestProviderInputPreparationKeepsExistingFingerprintPolicy(t *testing.T) {
	userID, attemptID := uuid.New(), uuid.New()
	for _, input := range []string{"I work with US clients.", "i work with us clients.", " Ｉ\twork with US clients. "} {
		require.Equal(t, "i work with us clients.", normalizeSentence(input))
		require.Equal(t,
			RequestHash(userID, attemptID, "work", "i work with us clients.", PromptVersionSentenceFeedbackV1),
			RequestHash(userID, attemptID, "work", normalizeSentence(input), PromptVersionSentenceFeedbackV1))
	}
	// Capitalization mistakes must remain visible, not guessed or repaired locally.
	task := NewDefaultTaskBuilder().Build(&Target{}, "  i work with US clients. ")
	require.Equal(t, "i work with US clients.", task.UserPayload["learner_sentence"])
	require.Equal(t, "sentence-input-v2-case-preserving", ProviderInputPreparationVersion)
}
