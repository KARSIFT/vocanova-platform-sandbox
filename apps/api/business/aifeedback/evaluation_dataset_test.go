package aifeedback

import (
	"fmt"
	"strings"
	"testing"
)

func TestEditorialRegressionLabels(t *testing.T) {
	validRegressions := 0
	for _, c := range InitialDataset() {
		if c.Sentence == "I happy every day." && c.ExpectedStatus == LearningStatusCorrect {
			t.Error("missing copula is incorrectly labeled correct")
		}
		if c.Sentence == "She is more careful than me." && c.ExpectedStatus != LearningStatusCorrect {
			t.Error("valid comparative is incorrectly labeled a grammar error")
		}
		if c.Sentence == "She is more careful than me." || c.Sentence == "I read yesterday." {
			validRegressions++
		}
		if c.Sentence == "I read yesterday." && c.ExpectedStatus != LearningStatusCorrect {
			t.Error("valid irregular past tense is incorrectly labeled a grammar error")
		}
		if c.ExpectedStatus == LearningStatusNeedsImprovement && strings.Contains(c.Sentence, "She read books every day.") {
			t.Error("valid habitual past is incorrectly treated as an agreement error")
		}
	}
	if validRegressions != 2 {
		t.Errorf("expected both valid historical false-correction regressions, got %d", validRegressions)
	}
}

func TestEditorialDatasetPreservesLegacyIDsAndGoldenMembership(t *testing.T) {
	words := []string{"work", "eat", "read", "run", "play", "write", "study", "drive", "cook", "help", "travel", "learn", "organize", "happy", "big", "quick", "careful", "busy", "book", "city", "friend", "school", "water", "time", "give up", "look after", "take off", "on time"}
	categories := []string{"correctness", "correctness", "grammar_error", "grammar_error", "incorrect_target_use", "regional_variant", "ambiguity", "prompt_injection", "sensitive_but_allowed", "unsafe_blocked", "a2_b1_level"}
	byID := make(map[string]EvaluationCase)
	for _, c := range InitialDataset() {
		if _, exists := byID[c.ID]; exists {
			t.Fatalf("duplicate fixture ID %s", c.ID)
		}
		byID[c.ID] = c
	}
	if len(byID) != 336 {
		t.Fatalf("expected 308 legacy cases plus 28 level pairs, got %d", len(byID))
	}
	for _, word := range words {
		for index, category := range categories {
			id := fmt.Sprintf("voc028-eval-%s-%s-%d", word, category, index)
			c, found := byID[id]
			if !found {
				t.Errorf("legacy fixture lost: %s", id)
			}
			if (index == 0 || index == 2) && !c.IsGolden {
				t.Errorf("legacy golden fixture dropped: %s", id)
			}
		}
	}
}

func TestEditorialCasesHaveMeaningRationaleAndExplicitOutcomes(t *testing.T) {
	for _, c := range InitialDataset() {
		t.Run(c.ID, func(t *testing.T) {
			if strings.TrimSpace(c.TargetMeaning) == "" || strings.TrimSpace(c.EditorialRationale) == "" {
				t.Fatal("missing meaning or editorial rationale")
			}
			switch c.ExpectedOutcome {
			case EvaluationOutcomeFeedback:
				if c.ExpectedStatus == "" && c.ScoringExclusionReason == "" {
					t.Fatal("unresolved status needs a visible review reason")
				}
			case EvaluationOutcomeSafetyIntercept:
				if c.ExpectedStatus != "" {
					t.Fatal("safety intervention is not a grammatical classification")
				}
			case EvaluationOutcomeValidationFailed:
				if c.ExpectedStatus != "" {
					t.Fatal("input rejection is not a grammatical classification")
				}
			default:
				t.Fatalf("unknown expected outcome %q", c.ExpectedOutcome)
			}
		})
	}
}

func TestEditorialValidationGapsRemainVisibleWithoutChangingLanguageLabels(t *testing.T) {
	gaps := map[string]string{"travel": "I travelled to the city.", "learn": "I learnt English last year.", "organize": "I organised my notes."}
	seen := 0
	for _, c := range InitialDataset() {
		target := &Target{NormalizedWord: c.TargetWord, WordType: c.WordType, PartOfSpeech: c.PartOfSpeech, AcceptedForms: BuildAcceptedForms(c.TargetWord, c.WordType, c.PartOfSpeech)}
		validation := ValidateSentence(c.Sentence, target)
		if c.Category == EvaluationCategoryRegionalVariant && gaps[c.TargetWord] != "" {
			seen++
			if c.Sentence != gaps[c.TargetWord] || c.ExpectedStatus != LearningStatusCorrect || c.ExpectedOutcome != EvaluationOutcomeFeedback || c.ScoringExclusionReason == "" || !c.IsGolden {
				t.Errorf("regional regression was hidden or relabeled: %+v", c)
			}
			if validation.Valid || validation.Code != ValidationCodeMissingTarget {
				t.Errorf("validator behavior changed for %s; reassess and remove the stale scoring exclusion", c.ID)
			}
		} else if !validation.Valid {
			t.Errorf("unexpected validator failure for %s: %s", c.ID, validation.Code)
		}
	}
	if seen != 3 {
		t.Fatalf("expected all three retained regional gaps, got %d", seen)
	}
}

func TestEditorialGoldenCoverageAndPairedLevels(t *testing.T) {
	counts := make(map[string]int)
	pairs := make(map[string][]EvaluationCase)
	for _, c := range GoldenSet() {
		counts[c.Category]++
		if c.Category == EvaluationCategoryA2B1Level {
			pairs[c.TargetWord] = append(pairs[c.TargetWord], c)
		}
	}
	for _, category := range []string{EvaluationCategoryCorrectness, EvaluationCategoryGrammarError, EvaluationCategoryIncorrectTargetUse, EvaluationCategoryRegionalVariant, EvaluationCategoryAmbiguity, EvaluationCategoryPromptInjection, EvaluationCategorySensitiveAllowed, EvaluationCategoryUnsafeBlocked, EvaluationCategoryA2B1Level} {
		if counts[category] == 0 {
			t.Errorf("golden coverage missing %s", category)
		}
	}
	if len(GoldenSet()) != 91 {
		t.Errorf("expected explicit 91-case coverage, got %d", len(GoldenSet()))
	}
	if len(pairs) == 0 {
		t.Fatal("no paired golden sentences")
	}
	for word, pair := range pairs {
		if len(pair) != 2 {
			t.Fatalf("incomplete pair for %s", word)
		}
		a, b := pair[0], pair[1]
		if a.Sentence != b.Sentence || a.TargetMeaning != b.TargetMeaning || a.ExpectedStatus != b.ExpectedStatus || a.LearnerLevel == b.LearnerLevel || a.ID == b.ID {
			t.Errorf("invalid level-only pair for %s", word)
		}
	}
}

func TestEditorialSafetyAndInjectionHaveDifferentExpectations(t *testing.T) {
	checker := NewDefaultLocalAbuseChecker()
	for _, c := range InitialDataset() {
		outcome := checker.Check(t.Context(), ModerationInput{SentenceText: c.Sentence})
		if c.Category == EvaluationCategoryUnsafeBlocked {
			if c.ExpectedOutcome != EvaluationOutcomeSafetyIntercept || outcome == nil || outcome.Outcome != SafetySelfHarmIntervention {
				t.Errorf("unsafe fixture does not exercise intended local intervention: %s", c.ID)
			}
		} else if outcome != nil {
			t.Errorf("ordinary or injection fixture unexpectedly classified as abuse: %s", c.ID)
		}
		if c.Category == EvaluationCategoryPromptInjection && (c.ExpectedOutcome != EvaluationOutcomeFeedback || c.ExpectedStatus != LearningStatusNeedsImprovement) {
			t.Errorf("injection fixture should ignore instruction and grade the actual error: %s", c.ID)
		}
	}
}
