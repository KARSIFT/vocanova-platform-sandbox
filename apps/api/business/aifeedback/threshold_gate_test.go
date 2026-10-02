package aifeedback

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fake-provider CI verifies mechanics; its scores are not model-quality evidence.
func TestGoldenSetThresholdsAgainstMockProvider(t *testing.T) {
	computed, violations, err := RunGoldenGate(t.Context(), DefaultGoldenThresholdSpec())
	require.NoError(t, err)
	require.Equal(t, len(GoldenSet()), computed.EvidenceCount)
	state, gaps := EvaluationAcceptance(computed, violations)
	assert.NotEqual(t, "PASS", state)
	assert.NotEmpty(t, gaps)
	assert.False(t, computed.MeaningMeasured)
	assert.False(t, computed.SafetyQualityMeasured)
}

// TestGoldenGateEnforcesViolatedThreshold proves the gate mechanism actually
// enforces thresholds rather than only reporting them: a hand-built result
// fixture that deliberately misses several DOC-09 §23 bounds must produce a
// matching violation for each one.
func TestGoldenGateEnforcesViolatedThreshold(t *testing.T) {
	spec := DefaultGoldenThresholdSpec()

	computed := GoldenThresholds{
		Total:           10,
		Validated:       10,
		ProviderCalled:  10,
		MatchedStatus:   1,
		CasesWithExpect: 10,

		CorrectnessTotal:          4,
		CorrectnessMatched:        1,
		IncorrectTargetUseTotal:   2,
		IncorrectTargetUseMatched: 0,

		CorrectExpectedTotal:           4,
		CorrectExpectedGotCorrect:      1,
		CorrectExpectedGotNeedsImprove: 1,
		CorrectExpectedGotIncorrect:    2,

		SelfHarmMeasured:          true,
		CorrectionQualityMeasured: true,
		CorrectionReviewedTotal:   4, UnnecessaryCorrections: 3, WrongCorrections: 2,
		SafetyQualityMeasured: true,
		SelfHarmTotal:         2,
		SelfHarmIntercepted:   0,

		ShamingOrInjectionCases:      3,
		ShamingOrInjectionViolations: 3,
	}

	violations := CheckGoldenThresholds(spec, computed)
	require.NotEmpty(t, violations, "a deliberately violated fixture must produce at least one violation")

	byMetric := make(map[string]ThresholdViolation, len(violations))
	for _, v := range violations {
		byMetric[v.Metric] = v
	}

	for _, metric := range []string{
		"overall_status_accuracy",
		"clearly_correct_accuracy",
		"clearly_incorrect_target_use_accuracy",
		"unnecessary_correction_on_clearly_correct",
		"wrong_correction_on_correct",
		"shaming_or_injection_or_unsafe_feedback",
		"correct_self_harm_intervention",
	} {
		assert.Contains(t, byMetric, metric, "expected metric %q to be reported as violated", metric)
	}

	report := FormatThresholdReport(computed, violations)
	assert.Contains(t, report, "Result: FAIL")
	assert.Contains(t, report, "violated")
}

// Measured status bounds can be met while release acceptance remains incomplete.
func TestGoldenGateCleanStatusFixtureStillNeedsEvidence(t *testing.T) {
	spec := DefaultGoldenThresholdSpec()

	computed := GoldenThresholds{
		Total:           20,
		Validated:       20,
		ProviderCalled:  20,
		MatchedStatus:   19,
		CasesWithExpect: 20,

		CorrectnessTotal:          10,
		CorrectnessMatched:        10,
		IncorrectTargetUseTotal:   4,
		IncorrectTargetUseMatched: 4,

		CorrectExpectedTotal:           10,
		CorrectExpectedGotCorrect:      10,
		CorrectExpectedGotNeedsImprove: 0,
		CorrectExpectedGotIncorrect:    0,

		SelfHarmTotal:       3,
		SelfHarmIntercepted: 3,

		ShamingOrInjectionCases:      5,
		ShamingOrInjectionViolations: 0,
	}

	violations := CheckGoldenThresholds(spec, computed)
	assert.Empty(t, violations)

	report := FormatThresholdReport(computed, violations)
	assert.True(t, strings.HasPrefix(report, "DatasetVersion="))
	assert.Contains(t, report, "Result: INCOMPLETE")
}

// TestCheckGoldenThresholdsSkipsNotTrackedFields confirms NotTracked spec
// fields never contribute a violation even when the corresponding computed
// data is entirely absent (RepairMeasured/MeaningMeasured both false) -
// the gate must report the gap, never silently pass or fail on data it
// cannot measure.
func TestCheckGoldenThresholdsSkipsNotTrackedFields(t *testing.T) {
	spec := DefaultGoldenThresholdSpec()
	require.Equal(t, 0.995, spec.StructuredOutputValidAfterOneRepair)
	require.Equal(t, 0.95, spec.MeaningPreservation)
	spec.StructuredOutputValidAfterOneRepair = NotTracked
	spec.MeaningPreservation = NotTracked

	computed := GoldenThresholds{
		Total:           1,
		Validated:       1,
		ProviderCalled:  1,
		MatchedStatus:   0,
		CasesWithExpect: 0,
	}

	violations := CheckGoldenThresholds(spec, computed)
	for _, v := range violations {
		assert.NotEqual(t, "structured_output_valid_after_one_repair", v.Metric)
		assert.NotEqual(t, "meaning_preservation", v.Metric)
	}

	report := FormatThresholdReport(computed, violations)
	assert.Contains(t, report, "Repair: not tracked")
	assert.Contains(t, report, "Meaning: not tracked")
}
