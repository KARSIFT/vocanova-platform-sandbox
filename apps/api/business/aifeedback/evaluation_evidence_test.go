package aifeedback

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluationEvidenceMissingMeasurementsCannotPass(t *testing.T) {
	report := FormatThresholdReport(GoldenThresholds{}, nil)
	assert.NotContains(t, report, "Result: PASS")
	assert.Contains(t, report, "INCOMPLETE")
}

type evidenceProvider func(context.Context, ProviderTask) (*ProviderFeedback, error)

func (p evidenceProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	return p(ctx, task)
}

func evidenceFixture(id string) EvaluationCase {
	return EvaluationCase{ID: id, TargetWord: "work", TargetMeaning: "to do a job", WordType: "word", PartOfSpeech: "verb", LearnerLevel: "a2", Sentence: "I work every day.", Category: EvaluationCategoryCorrectness, ExpectedStatus: LearningStatusCorrect, ExpectedOutcome: EvaluationOutcomeFeedback, EditorialRationale: "Present simple describes a regular activity."}
}

func TestEvaluationEvidenceEveryInputMeaningFailuresAndSafeOutput(t *testing.T) {
	cases := []EvaluationCase{evidenceFixture("ok"), evidenceFixture("validation"), evidenceFixture("error"), evidenceFixture("nil"), evidenceFixture("invalid")}
	cases[1].Sentence = "work"
	cases[1].ExpectedOutcome = EvaluationOutcomeValidationFailed
	cases[1].ExpectedStatus = ""
	calls := 0
	provider := evidenceProvider(func(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
		calls++
		require.Equal(t, "to do a job", task.UserPayload["target_meaning"])
		switch calls {
		case 2:
			return nil, errors.New("https://private.example?token=DO-NOT-RETAIN")
		case 3:
			return nil, nil
		case 4:
			return &ProviderFeedback{Status: LearningStatusCorrect}, nil
		}
		fb, err := NewMockProvider().GenerateFeedback(ctx, task)
		fb.RawJSON["private_provider_envelope"] = "DO-NOT-RETAIN"
		return fb, err
	})
	result := RunEvaluation(t.Context(), provider, cases)
	require.Len(t, result.CaseEvidence, len(cases))
	require.Equal(t, 4, result.ProviderCalled, "count attempts including errors and nil output")
	for i, want := range []string{"feedback", "validation_failed", "provider_error", "invalid_output", "invalid_output"} {
		e := result.CaseEvidence[i]
		assert.Equal(t, i, e.InputIndex)
		assert.Equal(t, cases[i].ID, e.Case.ID)
		assert.Equal(t, want, e.Outcome)
		assert.Nil(t, e.TransportAttempts)
		assert.Nil(t, e.RepairAttempts)
		assert.Nil(t, e.PersistenceConfirmed)
		assert.Equal(t, "not_reviewed", e.HumanReviewStatus)
		assert.GreaterOrEqual(t, e.ElapsedMillis, int64(0))
	}
	assert.Equal(t, ValidationCodeTooShort, result.CaseEvidence[1].Validation.Code)
	require.NotNil(t, result.CaseEvidence[1].ExpectedOutcomeMatched)
	assert.True(t, *result.CaseEvidence[1].ExpectedOutcomeMatched)
	assert.Equal(t, "provider_error", result.CaseEvidence[2].ProviderErrorCode)
	assert.NotEmpty(t, result.CaseEvidence[0].Feedback["explanation"])
	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "DO-NOT-RETAIN")
	assert.NotContains(t, string(encoded), "private.example")
	computed := ComputeGoldenThresholds(result, cases)
	assert.Equal(t, 4, computed.CasesWithExpect)
	assert.Equal(t, 1, computed.MatchedStatus, "expected-valid failures stay in the denominator")
}

func TestEvaluationEvidenceCanceledInputsAndDuplicateIDsRemainVisible(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cases := []EvaluationCase{evidenceFixture("same"), evidenceFixture("same")}
	provider := evidenceProvider(func(context.Context, ProviderTask) (*ProviderFeedback, error) {
		t.Fatal("canceled run must not call provider")
		return nil, nil
	})
	result := RunEvaluation(ctx, provider, cases)
	require.Len(t, result.CaseEvidence, 2)
	for _, e := range result.CaseEvidence {
		assert.Equal(t, "not_run", e.Outcome)
		assert.Equal(t, "canceled", e.ProviderErrorCode)
		assert.Zero(t, e.ProviderAttempts)
	}
	computed := ComputeGoldenThresholds(result, cases)
	assert.Equal(t, 2, computed.EvidenceCount)
	assert.Contains(t, computed.EvidenceGaps, "case identifiers are empty or duplicated")
}

func TestEvaluationEvidenceStatusMatchCannotCertifySafetyOrCorrection(t *testing.T) {
	c := evidenceFixture("injection")
	c.Category = EvaluationCategoryPromptInjection
	result := RunEvaluation(t.Context(), NewMockProvider(), []EvaluationCase{c})
	computed := ComputeGoldenThresholds(result, []EvaluationCase{c})
	assert.Equal(t, 1, computed.MatchedStatus)
	assert.False(t, computed.SafetyQualityMeasured)
	assert.False(t, computed.CorrectionQualityMeasured)
	state, gaps := EvaluationAcceptance(computed, CheckGoldenThresholds(DefaultGoldenThresholdSpec(), computed))
	assert.Equal(t, "INCOMPLETE", state)
	assert.NotEmpty(t, gaps)
	assert.False(t, result.CaseEvidence[0].SafetyChecked)
}

func TestEvaluationEvidenceOnlySpecificSafetyObservationCounts(t *testing.T) {
	cases := []EvaluationCase{evidenceFixture("one"), evidenceFixture("two")}
	for i := range cases {
		cases[i].Tags = []string{"self_harm"}
		cases[i].ExpectedOutcome = EvaluationOutcomeSafetyIntercept
		cases[i].ExpectedStatus = ""
	}
	result := EvaluationResult{CaseEvidence: []EvaluationCaseEvidence{
		{InputIndex: 0, Case: cases[0], SafetyChecked: true, Outcome: EvaluationOutcomeSafetyIntercept},
		{InputIndex: 1, Case: cases[1], SafetyChecked: true, Outcome: EvaluationOutcomeFeedback},
	}}
	computed := ComputeGoldenThresholds(result, cases)
	assert.Equal(t, 1, computed.SelfHarmIntercepted)
	assert.Equal(t, 2, computed.SelfHarmTotal)
	assert.True(t, computed.SelfHarmMeasured)
	assert.NotEmpty(t, CheckGoldenThresholds(DefaultGoldenThresholdSpec(), computed))
}

func TestEvaluationEvidenceExclusionIsVisibleAndUnscored(t *testing.T) {
	c := evidenceFixture("unresolved")
	c.ScoringExclusionReason = "Editorial judgment needs review."
	c.ExpectedStatus = LearningStatusIncorrect
	result := RunEvaluation(t.Context(), NewMockProvider(), []EvaluationCase{c})
	computed := ComputeGoldenThresholds(result, []EvaluationCase{c})
	assert.Equal(t, 1, computed.ExcludedCases)
	assert.Zero(t, computed.CasesWithExpect)
	assert.Nil(t, result.CaseEvidence[0].StatusMatched)
	assert.Equal(t, c.ScoringExclusionReason, result.CaseEvidence[0].Case.ScoringExclusionReason)
}

func TestEvaluationEvidenceReportMetadataAndEmptySubset(t *testing.T) {
	r := RunLiveEvaluation(t.Context(), NewMockProvider(), LiveEvaluationOptions{Cases: []EvaluationCase{}, Commit: "synthetic-test-revision"})
	assert.Equal(t, "INCOMPLETE", r.AcceptanceState)
	assert.Equal(t, EvaluationScopeAdapterOnly, r.Scope)
	assert.Equal(t, "synthetic-test-revision", r.Commit)
	assert.Equal(t, GoldenSetVersion, r.GoldenSetVersion)
	assert.Equal(t, PromptVersionSentenceFeedbackV1, r.PromptVersion)
	assert.Equal(t, SchemaVersionFeedbackV1, r.SchemaVersion)
	assert.Empty(t, r.CaseEvidence)
	assert.NotContains(t, FormatLiveEvaluationReport(r), "Result: PASS")
}

func TestEvaluationEvidenceNearestRankDoesNotUseFloorInterpolation(t *testing.T) {
	assert.Equal(t, 2*time.Millisecond, percentileNearestRank([]time.Duration{time.Millisecond, 2 * time.Millisecond}, 95))
}

type evidenceIdentityWrapper struct{ FeedbackProvider }

func (p evidenceIdentityWrapper) EvaluationIdentity() (string, string) {
	return ProviderEvaluationIdentity(p.FeedbackProvider)
}

func TestEvaluationEvidenceProviderIdentityDoesNotExposeConfig(t *testing.T) {
	for _, tc := range []struct {
		name     string
		provider FeedbackProvider
	}{
		{"cloudflare", NewCloudflareFeedbackProvider(CloudflareConfig{Model: "fixture-model", APIToken: "do-not-report", AccountID: "private-account", BaseURL: "https://private.example"})},
		{"gemini", NewGeminiFeedbackProvider(GeminiConfig{Model: "fixture-model", APIKey: "do-not-report", BaseURL: "https://private.example"})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name, model := ProviderEvaluationIdentity(evidenceIdentityWrapper{tc.provider})
			assert.Equal(t, tc.name, name)
			assert.Equal(t, "fixture-model", model)
			// No case means this identity-only test cannot make an HTTP request.
			r := RunLiveEvaluation(t.Context(), evidenceIdentityWrapper{tc.provider}, LiveEvaluationOptions{Cases: []EvaluationCase{}})
			encoded, err := json.Marshal(r)
			require.NoError(t, err)
			for _, forbidden := range []string{"do-not-report", "private-account", "private.example"} {
				assert.NotContains(t, string(encoded), forbidden)
			}
		})
	}
}

func TestEvaluationEvidenceWrongCorrectionTextRequiresHumanJudgment(t *testing.T) {
	c := evidenceFixture("correction")
	c.ExpectedStatus = LearningStatusNeedsImprovement
	provider := evidenceProvider(func(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
		feedback, err := NewMockProvider().GenerateFeedback(ctx, task)
		feedback.Status = LearningStatusNeedsImprovement
		feedback.CorrectedSentence = stringPtr("I work is every yesterday.")
		feedback.ImprovementTip = stringPtr("Use the suggested correction.")
		return feedback, err
	})
	result := RunEvaluation(t.Context(), provider, []EvaluationCase{c})
	require.True(t, *result.CaseEvidence[0].StatusMatched)
	assert.Equal(t, "I work is every yesterday.", result.CaseEvidence[0].Feedback["corrected_sentence"])
	computed := ComputeGoldenThresholds(result, []EvaluationCase{c})
	assert.False(t, computed.CorrectionQualityMeasured)
	assert.False(t, computed.MeaningMeasured)
	state, gaps := EvaluationAcceptance(computed, CheckGoldenThresholds(DefaultGoldenThresholdSpec(), computed))
	assert.Equal(t, "INCOMPLETE", state)
	assert.Contains(t, gaps, "correction quality and unnecessary corrections require human review")
}

func TestEvaluationEvidenceSingleCriticalWrongAcceptanceCannotHideInRatio(t *testing.T) {
	var cases []EvaluationCase
	var evidence []EvaluationCaseEvidence
	for i := range 28 {
		c := evidenceFixture(itoa(i))
		c.Category = EvaluationCategoryIncorrectTargetUse
		c.ExpectedStatus = LearningStatusIncorrect
		status := LearningStatusIncorrect
		if i == 0 {
			status = LearningStatusCorrect
		}
		cases = append(cases, c)
		evidence = append(evidence, EvaluationCaseEvidence{InputIndex: i, Case: c, Outcome: "feedback", Feedback: map[string]any{"status": status}})
	}
	computed := ComputeGoldenThresholds(EvaluationResult{CaseEvidence: evidence}, cases)
	assert.Equal(t, 27, computed.IncorrectTargetUseMatched)
	violations := CheckGoldenThresholds(DefaultGoldenThresholdSpec(), computed)
	require.Len(t, violations, 1, "27/28 clears ratios but one critical false acceptance blocks")
	assert.Equal(t, "critical_incorrect_target_use_marked_correct", violations[0].Metric)
	state, _ := EvaluationAcceptance(computed, violations)
	assert.Equal(t, "FAIL", state)
}

func TestEvaluationEvidenceCostMetadataIsFiniteAndExplicit(t *testing.T) {
	for _, cost := range []float64{-1, -2, -0.5, math.NaN(), math.Inf(1)} {
		r := RunLiveEvaluation(t.Context(), NewMockProvider(), LiveEvaluationOptions{Cases: []EvaluationCase{}, CostUSD: cost, CostCeilingUSD: math.Inf(1)})
		assert.Equal(t, float64(-1), r.CostUSD)
		assert.Equal(t, float64(-1), r.CostCeilingUSD)
		assert.Contains(t, r.AcceptanceGaps, "billed cost is unrecorded")
		_, err := json.Marshal(r)
		require.NoError(t, err, "invalid metadata cannot make the private JSON report unwritable")
	}
	r := RunLiveEvaluation(t.Context(), NewMockProvider(), LiveEvaluationOptions{Cases: []EvaluationCase{}, CostUSD: 0, CostCeilingUSD: 1, Timeout: 8 * time.Second, RequestInterval: time.Second, MaxRetries: 1})
	assert.NotContains(t, r.AcceptanceGaps, "billed cost is unrecorded")
	assert.Equal(t, 8*time.Second, r.Timeout)
	assert.Equal(t, time.Second, r.RequestInterval)
	assert.Equal(t, 1, r.MaxRetries)
	r = RunLiveEvaluation(t.Context(), NewMockProvider(), LiveEvaluationOptions{Cases: []EvaluationCase{}, CostUSD: 2, CostCeilingUSD: 1})
	assert.Equal(t, "FAIL", r.AcceptanceState)
	assert.True(t, r.CostCeilingExceeded)
}

func TestEvaluationEvidenceDisabledOrInvalidBoundsCannotEstablishAcceptance(t *testing.T) {
	for _, bound := range []float64{NotTracked, math.NaN(), math.Inf(1)} {
		spec := DefaultGoldenThresholdSpec()
		spec.OverallStatusAccuracy = bound
		r := RunLiveEvaluation(t.Context(), NewMockProvider(), LiveEvaluationOptions{Cases: []EvaluationCase{evidenceFixture("one")}, Spec: spec})
		assert.Equal(t, "INCOMPLETE", r.AcceptanceState)
		assert.Equal(t, "custom-diagnostic", r.SpecVersion)
		assert.Contains(t, r.AcceptanceGaps, "custom or disabled threshold bounds do not establish DOC-09 acceptance")
		assert.Equal(t, NotTracked, r.ThresholdSpec.OverallStatusAccuracy)
		_, err := json.Marshal(r)
		require.NoError(t, err)
	}
}

func TestEvaluationEvidenceAggregateInterceptDoesNotCreditOtherCases(t *testing.T) {
	cases := []EvaluationCase{
		{ID: "one", Category: EvaluationCategoryUnsafeBlocked, Tags: []string{"self_harm"}},
		{ID: "two", Category: EvaluationCategoryUnsafeBlocked, Tags: []string{"self_harm"}},
	}
	got := ComputeGoldenThresholds(EvaluationResult{Total: 2, ByStatus: map[string]int{"safety_intercepted": 1}}, cases)
	assert.Zero(t, got.SelfHarmIntercepted, "aggregate count cannot identify which case was intercepted")
}

func TestEvaluationEvidenceIncorrectStatusDoesNotProveWrongCorrection(t *testing.T) {
	got := CheckGoldenThresholds(DefaultGoldenThresholdSpec(), GoldenThresholds{CorrectExpectedTotal: 1, CorrectExpectedGotIncorrect: 1})
	for _, violation := range got {
		assert.NotEqual(t, "wrong_correction_on_correct", violation.Metric)
		assert.NotEqual(t, "unnecessary_correction_on_clearly_correct", violation.Metric)
	}
}
