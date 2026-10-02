package aifeedback

import (
	"strings"
	"testing"
)

func wrongSenseContractFeedback() *ProviderFeedback {
	tip := "Try a sentence about a place where children learn."
	return &ProviderFeedback{
		Status:                  LearningStatusIncorrect,
		TargetWordUsedCorrectly: false,
		GrammarAcceptable:       true,
		MeaningClear:            true,
		Naturalness:             NaturalnessNatural,
		Headline:                "Try the selected meaning",
		Explanation:             "Your sentence uses school for a group of fish, but this exercise asks for a place where children learn.",
		ImprovementTip:          &tip,
		RawJSON: map[string]any{
			"status":                     LearningStatusIncorrect,
			"target_word_used_correctly": false,
			"grammar_acceptable":         true,
			"meaning_clear":              true,
			"naturalness":                NaturalnessNatural,
			"corrected_sentence":         nil,
		},
	}
}

func TestFeedbackContractIncorrectAllowsNoFaithfulCorrection(t *testing.T) {
	feedback := wrongSenseContractFeedback()
	target := &Target{
		NormalizedWord:  "school",
		ShortDefinition: "a place where children go to learn",
		PartOfSpeech:    "noun",
	}
	if err := NewDefaultOutputValidator().Validate(feedback, target); err != nil {
		t.Fatalf("valid wrong-sense feedback must not require an invented correction: %v", err)
	}
	if feedback.CorrectedSentence != nil || feedback.Status != LearningStatusIncorrect || feedback.TargetWordUsedCorrectly {
		t.Fatal("validation changed the correction or wrong-sense classification")
	}
}

func TestFeedbackContractRejectsBlankNonNullCorrection(t *testing.T) {
	for _, status := range []string{LearningStatusNeedsImprovement, LearningStatusIncorrect} {
		for _, correction := range []string{"", " \t\n "} {
			t.Run(status+"/"+strings.ReplaceAll(correction, "\n", "newline"), func(t *testing.T) {
				feedback := wrongSenseContractFeedback()
				feedback.Status = status
				feedback.CorrectedSentence = &correction
				feedback.RawJSON["status"] = status
				feedback.RawJSON["corrected_sentence"] = correction
				if err := NewDefaultOutputValidator().Validate(feedback, nil); err == nil {
					t.Fatal("a present correction must contain useful text, or be null")
				}
			})
		}
	}
}

func TestFeedbackContractIncorrectStillRequiresSpecificFeedback(t *testing.T) {
	for _, missing := range []string{"explanation", "tip"} {
		t.Run(missing, func(t *testing.T) {
			feedback := wrongSenseContractFeedback()
			if missing == "explanation" {
				feedback.Explanation = " "
			} else {
				feedback.ImprovementTip = nil
			}
			if err := NewDefaultOutputValidator().Validate(feedback, nil); err == nil {
				t.Fatal("omitting a correction must not remove the explanation/tip requirement")
			}
		})
	}
}

func TestFeedbackContractNeedsImprovementAllowsFalseTargetDiagnostic(t *testing.T) {
	feedback := wrongSenseContractFeedback()
	feedback.Status = LearningStatusNeedsImprovement
	feedback.GrammarAcceptable = false
	feedback.Naturalness = NaturalnessUnderstandable
	feedback.Headline = "A small grammar change"
	feedback.Explanation = "Use works after she in the present tense."
	tip := "Use the third-person singular verb form."
	feedback.ImprovementTip = &tip
	feedback.RawJSON["status"] = feedback.Status
	feedback.RawJSON["grammar_acceptable"] = false
	feedback.RawJSON["naturalness"] = feedback.Naturalness
	if err := NewDefaultOutputValidator().Validate(feedback, nil); err != nil {
		t.Fatalf("existing needs-improvement diagnostic combinations must remain supported: %v", err)
	}
}
