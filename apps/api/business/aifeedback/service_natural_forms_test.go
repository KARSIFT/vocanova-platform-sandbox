package aifeedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// This fake verifies orchestration and target data only, never model quality.
type naturalFormRecordingProvider struct {
	tasks     []ProviderTask
	incorrect bool
}

func (p *naturalFormRecordingProvider) GenerateFeedback(_ context.Context, task ProviderTask) (*ProviderFeedback, error) {
	p.tasks = append(p.tasks, task)
	feedback := &ProviderFeedback{Status: LearningStatusCorrect, TargetWordUsedCorrectly: true, GrammarAcceptable: true, MeaningClear: true, Naturalness: NaturalnessNatural, Headline: "Good sentence", Explanation: "The sentence uses the selected meaning."}
	if p.incorrect {
		correction := "We caught up over coffee yesterday."
		tip := "Use catch up for exchanging news with someone."
		feedback.Status = LearningStatusIncorrect
		feedback.TargetWordUsedCorrectly = false
		feedback.MeaningClear = false
		feedback.Naturalness = NaturalnessUnnatural
		feedback.Headline = "Check the meaning"
		feedback.Explanation = "The intended meaning is to exchange news with someone."
		feedback.CorrectedSentence = &correction
		feedback.ImprovementTip = &tip
	}
	feedback.RawJSON = feedback.StructuredJSON()
	return feedback, nil
}

func TestServiceNaturalFormsLoadAuthoritativeTargetAndReachProvider(t *testing.T) {
	for _, tc := range []struct {
		word, wordType, pos, meaning, form, sentence string
		incorrect                                    bool
	}{
		{"travel", "word", "verb", "to go from one place to another", "travelled", "I travelled to the city.", false},
		{"learn", "word", "verb", "to gain knowledge or a skill", "learnt", "I learnt English last year.", false},
		{"organize", "word", "verb", "to arrange things in order", "organised", "I organised my notes.", false},
		{"cancel", "word", "verb", "to decide a planned event will not happen", "cancelled", "I cancelled our dinner booking when two friends became ill.", false},
		{"catch up", "phrase", "verb", "to exchange recent news with someone", "caught up", "We caught up over coffee yesterday.", false},
		{"meet up", "phrasal_verb", "verb", "to meet someone by arrangement", "met up", "We met up with Ana after class and went for coffee.", false},
		{"keep in touch", "idiom", "verb", "to continue communicating with someone", "kept in touch", "We kept in touch by sending each other a message every week.", false},
		{"sounds good", "phrase", "phrase", "used to agree with a suggestion", "sounded good", "That sounded good.", false},
		{"catch up", "phrase", "verb", "to exchange recent news with someone", "caught up", "The paint caught up with my coffee.", true},
	} {
		t.Run(tc.sentence, func(t *testing.T) {
			f := newServiceFixture(t)
			f.repo.words[0].Text, f.repo.words[0].NormalizedText, f.repo.words[0].WordType = tc.word, tc.word, tc.wordType
			f.repo.meanings[0].PartOfSpeech, f.repo.meanings[0].ShortDefinition = tc.pos, tc.meaning
			provider := &naturalFormRecordingProvider{incorrect: tc.incorrect}
			f.service.provider = provider
			result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request(tc.sentence))
			require.NoError(t, err)
			require.Equal(t, ProcessingStatusCompleted, result.ProcessingStatus)
			require.Len(t, provider.tasks, 1)
			payload := provider.tasks[0].UserPayload
			require.Equal(t, tc.word, payload["target_word"])
			require.Equal(t, tc.meaning, payload["target_meaning"])
			require.Contains(t, payload["accepted_forms"], tc.form)
			require.Equal(t, tc.sentence, result.OriginalSentence)
			if tc.incorrect {
				require.Equal(t, LearningStatusIncorrect, result.Status, "lexical presence must not override provider's meaning judgment")
				require.False(t, result.TargetWordUsedCorrectly)
				require.NotNil(t, result.CorrectedSentence)
			} else {
				require.Equal(t, LearningStatusCorrect, result.Status)
			}
		})
	}
}
