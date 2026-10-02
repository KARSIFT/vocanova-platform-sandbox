package aifeedback

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestStarterCurriculumPhrasalVerbsUseScopedContiguousForms(t *testing.T) {
	for _, tc := range []struct {
		word                       string
		forms, sentences, rejected []string
	}{
		{"try on", []string{"tries on", "trying on", "tried on"}, []string{"She tries on the jacket.", "She is trying on the jacket.", "She tried on the shoes before buying them."}, []string{"She tried only the shoes.", "She tried the shoes on yesterday.", "We studied the retry online."}},
		{"call back", []string{"calls back", "calling back", "called back"}, []string{"He calls back after work.", "He is calling back after lunch.", "I called back when the meeting ended."}, []string{"She called the neighbour back.", "They recalled background details.", "We called backstage yesterday."}},
	} {
		t.Run(tc.word, func(t *testing.T) {
			forms := BuildAcceptedForms(tc.word, "phrasal_verb", "verb")
			for _, form := range tc.forms {
				require.Contains(t, forms, form)
			}
			target := &Target{WordText: tc.word, NormalizedWord: tc.word, WordType: "phrasal_verb", PartOfSpeech: "verb", AcceptedForms: forms}
			for _, sentence := range tc.sentences {
				require.True(t, ValidateSentence(sentence, target).Valid, sentence)
			}
			for _, sentence := range tc.rejected {
				require.Equal(t, ValidationCodeMissingTarget, ValidateSentence(sentence, target).Code, sentence)
			}
			for _, tuple := range [][2]string{{"phrase", "verb"}, {"idiom", "verb"}, {"phrasal_verb", "noun"}} {
				for _, form := range tc.forms {
					require.NotContains(t, BuildAcceptedForms(tc.word, tuple[0], tuple[1]), form)
				}
			}
		})
	}
}

func TestStarterCurriculumSelectedObjectPronouns(t *testing.T) {
	for _, tc := range []struct {
		word                                string
		heads, objects, sentences, rejected []string
	}{
		{"try on", []string{"try", "tries", "trying", "tried"}, []string{"it", "them"}, []string{"Please try it on today.", "She tries them on before buying.", "He is trying it on now.", "I tried them on yesterday."}, []string{"I tried her on the phone.", "Please try it right on today.", "She retried it on Tuesday.", "I tried it online yesterday.", "I tried the shoes on yesterday."}},
		{"call back", []string{"call", "calls", "calling", "called"}, []string{"me", "you", "him", "her", "us", "them"}, []string{"Please call me back tomorrow.", "I will call you back tomorrow.", "We called him back yesterday.", "Sam calls her back after lunch.", "She is calling us back tonight.", "I called them back after work."}, []string{"We recalled her background yesterday.", "I called her backstage yesterday.", "Please call me right back today.", "I called my friend back yesterday.", "They called it back yesterday."}},
	} {
		t.Run(tc.word, func(t *testing.T) {
			forms := BuildAcceptedForms(tc.word, "phrasal_verb", "verb")
			for _, head := range tc.heads {
				for _, object := range tc.objects {
					require.Contains(t, forms, head+" "+object+" "+map[string]string{"try on": "on", "call back": "back"}[tc.word])
				}
			}
			target := &Target{WordText: tc.word, NormalizedWord: tc.word, WordType: "phrasal_verb", PartOfSpeech: "verb", AcceptedForms: forms}
			for _, sentence := range tc.sentences {
				require.True(t, ValidateSentence(sentence, target).Valid, sentence)
			}
			for _, sentence := range tc.rejected {
				require.Equal(t, ValidationCodeMissingTarget, ValidateSentence(sentence, target).Code, sentence)
			}
			for _, tuple := range [][2]string{{"phrase", "verb"}, {"idiom", "verb"}, {"phrasal_verb", "noun"}} {
				for _, form := range forms {
					if len(strings.Fields(form)) == 3 {
						require.NotContains(t, BuildAcceptedForms(tc.word, tuple[0], tuple[1]), form)
					}
				}
			}
		})
	}
}

func TestStarterObjectFormsDoNotCrossPunctuationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		word, valid string
		invalid     []string
	}{
		{"try on", `She said, "Try it on!"`, []string{"Please try it. On Monday, we leave.", "Please try, it on today.", "Please try/it/on today.", "Please try-it-on today."}},
		{"call back", `She asked, "Call me back!"`, []string{"I called her. Back at home, I rested.", "Please call, her back tomorrow.", "Please call/her/back tomorrow.", "Please call-her-back tomorrow."}},
	} {
		target := &Target{NormalizedWord: tc.word, WordType: "phrasal_verb", PartOfSpeech: "verb", AcceptedForms: BuildAcceptedForms(tc.word, "phrasal_verb", "verb")}
		require.True(t, ValidateSentence(tc.valid, target).Valid)
		for _, sentence := range tc.invalid {
			require.Equal(t, ValidationCodeMissingTarget, ValidateSentence(sentence, target).Code, sentence)
		}
	}
}
