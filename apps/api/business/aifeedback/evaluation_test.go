package aifeedback

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialDatasetHasAtLeastTwoHundredCases(t *testing.T) {
	cases := InitialDataset()
	assert.GreaterOrEqual(t, len(cases), 200, "initial dataset must contain at least 200 synthetic cases")
}

func TestGoldenSetHasApproximatelyFiftyCases(t *testing.T) {
	cases := GoldenSet()
	assert.GreaterOrEqual(t, len(cases), 40, "golden set should contain around 50 cases")
	assert.LessOrEqual(t, len(cases), 60, "golden set should contain around 50 cases")
}

func TestInitialDatasetCoversAllRequiredCategories(t *testing.T) {
	cases := InitialDataset()
	required := []string{
		EvaluationCategoryCorrectness,
		EvaluationCategoryGrammarError,
		EvaluationCategoryRegionalVariant,
		EvaluationCategoryAmbiguity,
		EvaluationCategoryPromptInjection,
		EvaluationCategorySensitiveAllowed,
		EvaluationCategoryUnsafeBlocked,
		EvaluationCategoryA2B1Level,
		EvaluationCategoryIncorrectTargetUse,
	}
	counts := make(map[string]int)
	for _, c := range cases {
		counts[c.Category]++
	}
	for _, category := range required {
		assert.Greater(t, counts[category], 0, "category %s must be represented in the dataset", category)
	}
}

func TestGoldenSetIsSubsetOfInitialDataset(t *testing.T) {
	initial := InitialDataset()
	golden := GoldenSet()
	initialIDs := make(map[string]struct{}, len(initial))
	for _, c := range initial {
		initialIDs[c.ID] = struct{}{}
	}
	for _, c := range golden {
		assert.Contains(t, initialIDs, c.ID, "golden case %s must be in the initial dataset", c.ID)
	}
}

func TestMockEvaluationRunsWithoutRealProvider(t *testing.T) {
	result := RunMockEvaluation(t.Context())
	assert.Equal(t, DatasetVersion, result.DatasetVersion)
	assert.Greater(t, result.Total, 0)
	assert.GreaterOrEqual(t, result.ProviderCalled, 0)
	assert.GreaterOrEqual(t, result.Validated, 0)
	assert.NotNil(t, result.ByCategory)
	assert.NotNil(t, result.ByStatus)
}

func TestGoldenEvaluationRunsWithoutRealProvider(t *testing.T) {
	result := RunGoldenEvaluation(t.Context())
	assert.Greater(t, result.Total, 0)
}

func TestEvaluationResultContainsOnlyAllowedStatuses(t *testing.T) {
	result := RunMockEvaluation(t.Context())
	allowed := []string{
		LearningStatusCorrect,
		LearningStatusNeedsImprovement,
		LearningStatusIncorrect,
		"validation_failed",
		"provider_error",
	}
	for _, c := range result.MismatchedCases {
		assert.NotEmpty(t, c.Case.ID)
		assert.Contains(t, allowed, c.GotStatus, "mismatch status must be an allowed status or failure category")
		assert.NotEmpty(t, c.Case.Category)
	}
}

func TestDatasetContainsInjectionAndSafetyCases(t *testing.T) {
	cases := InitialDataset()
	var injection, unsafeCount, sensitive int
	for _, c := range cases {
		switch c.Category {
		case EvaluationCategoryPromptInjection:
			injection++
		case EvaluationCategoryUnsafeBlocked:
			unsafeCount++
		case EvaluationCategorySensitiveAllowed:
			sensitive++
		}
	}
	assert.Greater(t, injection, 0)
	assert.Greater(t, unsafeCount, 0)
	assert.Greater(t, sensitive, 0)
}
