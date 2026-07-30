package aifeedback

import (
	"context"
	"strings"
	"testing"
	"time"
)

// recordingProvider is a FeedbackProvider used by the live_eval
// tests. It records the ProviderTasks it was called with and
// returns a preset ProviderFeedback for each call. It is the
// test-only counterpart of the production OpenCode provider
// (which would never be reached from CI). It is named
// "recording" rather than "fake" to make its purpose obvious
// in failure messages: it records the inputs, it does not
// generate fake output to match the dataset's expectations.
type recordingProvider struct {
	t             *testing.T
	calls         []ProviderTask
	feedback      ProviderFeedback
	errOnCall     error
	delay         time.Duration
	failOnCallN   int
	callsObserved int
}

func (r *recordingProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	if r.delay > 0 {
		select {
		case <-time.After(r.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	r.calls = append(r.calls, task)
	r.callsObserved++
	if r.failOnCallN > 0 && r.callsObserved == r.failOnCallN {
		return nil, r.errOnCall
	}
	fb := r.feedback
	return &fb, nil
}

func TestInstrumentedProviderWrapsProvider(t *testing.T) {
	inner := &recordingProvider{
		t:        t,
		feedback: ProviderFeedback{Status: LearningStatusCorrect, Explanation: "ok"},
	}
	ip := NewInstrumentedProvider(inner)
	if ip == nil {
		t.Fatal("NewInstrumentedProvider returned nil for non-nil inner")
	}
	task := ProviderTask{
		SystemPrompt:    "system",
		DeveloperPrompt: "developer",
		UserPayload: map[string]any{
			"learner_sentence": "I work every day.",
			"target_word":      "work",
			"part_of_speech":   "verb",
			"target_meaning":   "to do a job",
			"learner_level":    "a2",
			"accepted_forms":   []string{"work", "works", "worked"},
		},
	}
	fb, err := ip.GenerateFeedback(context.Background(), task)
	if err != nil {
		t.Fatalf("GenerateFeedback: %v", err)
	}
	if fb == nil || fb.Status != LearningStatusCorrect {
		t.Fatalf("unexpected feedback: %+v", fb)
	}
	if len(ip.PerCallLatency) != 1 {
		t.Fatalf("expected 1 recorded latency, got %d", len(ip.PerCallLatency))
	}
	if ip.PerCallLatency[0] < 0 {
		t.Fatalf("latency must be non-negative, got %s", ip.PerCallLatency[0])
	}
	// InputChars is a rough estimate; assert it is non-zero
	// and that the known string fields contributed.
	if ip.InputChars == 0 {
		t.Fatal("expected non-zero InputChars")
	}
	// OutputChars should include Status + Explanation.
	if ip.OutputChars < len(LearningStatusCorrect)+len("ok") {
		t.Fatalf("OutputChars too small: %d", ip.OutputChars)
	}
	// The inner provider must have seen exactly one call.
	if len(inner.calls) != 1 {
		t.Fatalf("inner provider saw %d calls, expected 1", len(inner.calls))
	}
}

func TestInstrumentedProviderNilInner(t *testing.T) {
	ip := NewInstrumentedProvider(nil)
	if ip != nil {
		t.Fatal("NewInstrumentedProvider should return nil for nil inner")
	}
}

func TestInstrumentedProviderRecordsErrorLatency(t *testing.T) {
	inner := &recordingProvider{
		t:           t,
		errOnCall:   context.DeadlineExceeded,
		failOnCallN: 1,
	}
	ip := NewInstrumentedProvider(inner)
	task := ProviderTask{SystemPrompt: "x"}
	fb, err := ip.GenerateFeedback(context.Background(), task)
	if err == nil {
		t.Fatal("expected error from failing provider, got nil")
	}
	if fb != nil {
		t.Fatalf("expected nil feedback on error, got %+v", fb)
	}
	if len(ip.PerCallLatency) != 1 {
		t.Fatalf("expected 1 latency entry even on error, got %d", len(ip.PerCallLatency))
	}
	// A failed call has zero output chars.
	if ip.OutputChars != 0 {
		t.Fatalf("OutputChars should be 0 on error, got %d", ip.OutputChars)
	}
}

func TestInstrumentedProviderRespectsContextCancellation(t *testing.T) {
	inner := &recordingProvider{
		t:     t,
		delay: 50 * time.Millisecond,
	}
	ip := NewInstrumentedProvider(inner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ip.GenerateFeedback(ctx, ProviderTask{SystemPrompt: "x"})
	if err == nil {
		t.Fatal("expected cancelled-context error")
	}
	if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "context") {
		t.Fatalf("expected context error, got %v", err)
	}
}

func TestSummarizeLatenciesEmpty(t *testing.T) {
	s := summarizeLatencies(nil)
	if s.min != 0 || s.max != 0 || s.mean != 0 || s.p50 != 0 || s.p95 != 0 {
		t.Fatalf("expected zero summary for empty input, got %+v", s)
	}
}

func TestSummarizeLatenciesSingleValue(t *testing.T) {
	s := summarizeLatencies([]time.Duration{7 * time.Millisecond})
	if s.min != 7*time.Millisecond || s.max != 7*time.Millisecond {
		t.Fatalf("single-value min/max wrong: min=%s max=%s", s.min, s.max)
	}
	if s.p50 != 7*time.Millisecond || s.p95 != 7*time.Millisecond {
		t.Fatalf("single-value percentiles wrong: p50=%s p95=%s", s.p50, s.p95)
	}
}

func TestSummarizeLatenciesMultipleValues(t *testing.T) {
	// 1..20 ms
	values := make([]time.Duration, 20)
	for i := range values {
		values[i] = time.Duration(i+1) * time.Millisecond
	}
	s := summarizeLatencies(values)
	if s.min != time.Millisecond {
		t.Fatalf("min: %s", s.min)
	}
	if s.max != 20*time.Millisecond {
		t.Fatalf("max: %s", s.max)
	}
	expectedMean := time.Duration(0)
	for i := 1; i <= 20; i++ {
		expectedMean += time.Duration(i) * time.Millisecond
	}
	expectedMean /= 20
	if s.mean != expectedMean {
		t.Fatalf("mean: got %s want %s", s.mean, expectedMean)
	}
	// Nearest-rank p50 on 20 sorted values: rank = (50*(20-1))/100 = 9 (0-indexed) -> 10 ms.
	if s.p50 != 10*time.Millisecond {
		t.Fatalf("p50: got %s want 10ms", s.p50)
	}
	// Nearest-rank p95 on 20 sorted values: rank = (95*(20-1))/100 = 18 (0-indexed) -> 19 ms.
	if s.p95 != 19*time.Millisecond {
		t.Fatalf("p95: got %s want 19ms", s.p95)
	}
}

func TestSummarizeLatenciesIsDeterministic(t *testing.T) {
	values := []time.Duration{
		5 * time.Millisecond,
		2 * time.Millisecond,
		8 * time.Millisecond,
		1 * time.Millisecond,
		3 * time.Millisecond,
	}
	s1 := summarizeLatencies(values)
	s2 := summarizeLatencies(values)
	if s1 != s2 {
		t.Fatalf("summarizeLatencies is not deterministic: %+v != %+v", s1, s2)
	}
}

func TestPercentileNearestRankEdgeCases(t *testing.T) {
	if percentileNearestRank(nil, 50) != 0 {
		t.Fatal("empty input should yield 0")
	}
	sorted := []time.Duration{1, 2, 3, 4, 5}
	// p=0 -> first element
	if got := percentileNearestRank(sorted, 0); got != 1 {
		t.Fatalf("p=0: got %d want 1", got)
	}
	// p=100 -> last element
	if got := percentileNearestRank(sorted, 100); got != 5 {
		t.Fatalf("p=100: got %d want 5", got)
	}
	// p out of range -> clamped
	if got := percentileNearestRank(sorted, -10); got != 1 {
		t.Fatalf("p=-10 should clamp to 0: got %d", got)
	}
	if got := percentileNearestRank(sorted, 200); got != 5 {
		t.Fatalf("p=200 should clamp to 100: got %d", got)
	}
}

func TestProviderNameForOpenCodeAndMock(t *testing.T) {
	mock := &MockProvider{}
	if got := providerName(mock); got != "mock" {
		t.Fatalf("providerName(MockProvider): got %q want %q", got, "mock")
	}
	oc := NewOpenCodeFeedbackProvider(OpenCodeConfig{BaseURL: "http://example", Model: "opencode-go/deepseek-v4-pro"})
	if got := providerName(oc); got != "opencode" {
		t.Fatalf("providerName(OpenCodeFeedbackProvider): got %q want %q", got, "opencode")
	}
	if got := providerModel(oc); got != "opencode-go/deepseek-v4-pro" {
		t.Fatalf("providerModel(OpenCodeFeedbackProvider): got %q", got)
	}
	if got := providerModel(mock); got != "" {
		t.Fatalf("providerModel(MockProvider): got %q want empty", got)
	}
}

func TestRunLiveEvaluationRecordsReportShapeAgainstMockProvider(t *testing.T) {
	// The mock provider is the deterministic CI provider; running
	// RunLiveEvaluation against it exercises the library's full
	// end-to-end path without depending on a real provider. The
	// mock provider does NOT pass the default DOC-09 §23 spec
	// (its 50% match rate is below the 0.90 overall-status
	// threshold), so this test deliberately does not assert on
	// pass/fail; it only asserts on the report's structural
	// fields. The "pass" path is covered by the cmd's
	// expectationMatchingProvider test, which is the
	// in-package equivalent of a real provider that returns the
	// right status for every case.
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{})
	if report.Provider != "mock" {
		t.Fatalf("provider: got %q want %q", report.Provider, "mock")
	}
	if report.DatasetVersion != DatasetVersion && report.DatasetVersion != GoldenSetVersion {
		// RunEvaluation's DatasetVersion is set from the
		// EvaluationResult it produces, which uses DatasetVersion
		// (initial-dataset-v1) by default. The GoldenSet() data
		// is the same as the InitialDataset() in this codebase
		// (golden set is a curated subset of initial). Both
		// versions are acceptable; the field must be non-empty.
		t.Fatalf("DatasetVersion empty: %q", report.DatasetVersion)
	}
	if report.Thresholds.Total == 0 {
		t.Fatal("Thresholds.Total is zero; eval ran no cases")
	}
	if report.SpecVersion != "doc09-v1" {
		t.Fatalf("SpecVersion: got %q", report.SpecVersion)
	}
	if !report.StartedAt.Before(report.FinishedAt) && !report.StartedAt.Equal(report.FinishedAt) {
		t.Fatalf("StartedAt must be <= FinishedAt: %s > %s", report.StartedAt, report.FinishedAt)
	}
	if report.Duration < 0 {
		t.Fatalf("Duration must be non-negative: %s", report.Duration)
	}
}

func TestRunLiveEvaluationHonorsCustomCostCeiling(t *testing.T) {
	// CostCeilingUSD set to 0.5 with CostUSD 0.7 should flag
	// the ceiling as exceeded (the run is too expensive).
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		CostCeilingUSD: 0.50,
		CostUSD:        0.70,
	})
	if !report.CostCeilingExceeded {
		t.Fatal("expected CostCeilingExceeded=true with CostUSD > CostCeilingUSD")
	}

	// CostCeilingUSD set to 0.5 with CostUSD 0.4 should NOT
	// flag the ceiling as exceeded.
	report2 := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		CostCeilingUSD: 0.50,
		CostUSD:        0.40,
	})
	if report2.CostCeilingExceeded {
		t.Fatal("expected CostCeilingExceeded=false with CostUSD < CostCeilingUSD")
	}

	// CostCeilingUSD set to -1 (no ceiling) should NEVER
	// flag the ceiling as exceeded, even with a high
	// CostUSD. The report's rendered text should also
	// reflect "ceiling not set".
	report3 := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		CostCeilingUSD: -1,
		CostUSD:        999.0,
	})
	if report3.CostCeilingExceeded {
		t.Fatal("expected CostCeilingExceeded=false with negative CostCeilingUSD")
	}
	rendered := FormatLiveEvaluationReport(report3)
	if !strings.Contains(rendered, "ceiling not set") {
		t.Fatalf("rendered report should say 'ceiling not set' for negative CostCeilingUSD; got:\n%s", rendered)
	}
}

func TestRunLiveEvaluationHonorsOperatorNotes(t *testing.T) {
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		OperatorNotes: "rate-limit warning at case 17; one retry succeeded",
	})
	if report.OperatorNotes != "rate-limit warning at case 17; one retry succeeded" {
		t.Fatalf("OperatorNotes not stored: %q", report.OperatorNotes)
	}
	rendered := FormatLiveEvaluationReport(report)
	if !strings.Contains(rendered, "OperatorNotes: rate-limit warning at case 17; one retry succeeded") {
		t.Fatalf("rendered report should contain OperatorNotes; got:\n%s", rendered)
	}
}

func TestRunLiveEvaluationTracksLatency(t *testing.T) {
	// Use a recording provider with a measurable delay so
	// the per-call latency is non-zero and the summary
	// statistics are exercised.
	inner := &recordingProvider{
		t:        t,
		feedback: ProviderFeedback{Status: LearningStatusCorrect, Explanation: "ok"},
		delay:    1 * time.Millisecond,
	}
	report := RunLiveEvaluation(context.Background(), inner, LiveEvaluationOptions{})
	if report.LatencyMin == 0 {
		t.Fatal("LatencyMin should be non-zero with a 1ms-delay provider")
	}
	if report.LatencyMax < report.LatencyMin {
		t.Fatalf("LatencyMax (%s) must be >= LatencyMin (%s)", report.LatencyMax, report.LatencyMin)
	}
	if report.LatencyP50 == 0 {
		t.Fatal("LatencyP50 should be non-zero")
	}
	if report.LatencyP95 == 0 {
		t.Fatal("LatencyP95 should be non-zero")
	}
	if report.LatencyMean == 0 {
		t.Fatal("LatencyMean should be non-zero")
	}
	if len(report.PerCallLatency) == 0 {
		t.Fatal("PerCallLatency should be non-empty")
	}
}

func TestRunLiveEvaluationOperatorCostDefaultIsNotExceeded(t *testing.T) {
	// Default Options (zero value): CostUSD = 0, CostCeilingUSD = -1.
	// The run should NOT report CostCeilingExceeded and should
	// NOT mark the report as "cost not recorded" because the
	// zero value of CostCeilingUSD is -1 (the documented
	// "no ceiling" sentinel).
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{})
	if report.CostCeilingExceeded {
		t.Fatal("default (zero-value) options should not mark CostCeilingExceeded")
	}
}

func TestFormatLiveEvaluationReportContainsAllFields(t *testing.T) {
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		CostCeilingUSD: 1.00,
		CostUSD:        0.50,
		OperatorNotes:  "test notes",
	})
	rendered := FormatLiveEvaluationReport(report)
	expectedSubstrings := []string{
		"Provider:",
		"Model:",
		"Dataset:",
		"Spec:",
		"StartedAt:",
		"FinishedAt:",
		"Duration:",
		"ProviderCalls:",
		"EstimatedInputChars:",
		"EstimatedOutputChars:",
		"CostUSD:",
		"CostCeilingUSD:",
		"CostCeilingExceeded:",
		"LatencyMin:",
		"LatencyMax:",
		"LatencyMean:",
		"LatencyP50:",
		"LatencyP95:",
		"Per-threshold computed values",
		"Result:",
	}
	for _, s := range expectedSubstrings {
		if !strings.Contains(rendered, s) {
			t.Errorf("rendered report missing %q\n--- full report ---\n%s", s, rendered)
		}
	}
	if !strings.Contains(rendered, "OperatorNotes: test notes") {
		t.Errorf("rendered report should contain OperatorNotes line; got:\n%s", rendered)
	}
}

func TestFormatLiveEvaluationReportOnFailureIncludesViolations(t *testing.T) {
	// Use a spec that is guaranteed to fail (every threshold
	// set to a value the mock provider cannot meet).
	impossible := GoldenThresholdSpec{
		StructuredOutputValidFirstResponse:    0.9999,
		StructuredOutputValidAfterOneRepair:   0.9999,
		OverallStatusAccuracy:                 0.9999,
		ClearlyCorrectAccuracy:                0.9999,
		ClearlyIncorrectTargetUseAccuracy:     0.9999,
		UnnecessaryCorrectionOnClearlyCorrect: 0.0,
		WrongCorrectionOnCorrect:              0,
		MeaningPreservation:                   0.9999,
		ShamingOrInjectionOrUnsafeFeedback:    0,
		CorrectSelfHarmIntervention:           1.0,
	}
	report := RunLiveEvaluation(context.Background(), NewMockProvider(), LiveEvaluationOptions{
		Spec: impossible,
	})
	if len(report.Violations) == 0 {
		t.Fatal("expected violations with impossible spec; got none")
	}
	rendered := FormatLiveEvaluationReport(report)
	if !strings.Contains(rendered, "Result: FAIL") {
		t.Fatalf("rendered report should show FAIL; got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "tracked threshold(s) violated") {
		t.Fatalf("rendered report should mention violation count; got:\n%s", rendered)
	}
}

func TestFtoaHandlesNegativesAndZero(t *testing.T) {
	if got := ftoa(0); got != "0.00" {
		t.Fatalf("ftoa(0): got %q want %q", got, "0.00")
	}
	if got := ftoa(1.5); got != "1.50" {
		t.Fatalf("ftoa(1.5): got %q want %q", got, "1.50")
	}
	if got := ftoa(-2.75); got != "-2.75" {
		t.Fatalf("ftoa(-2.75): got %q want %q", got, "-2.75")
	}
}

func TestLiveEvaluationEnvironmentConstantsMatchDocString(t *testing.T) {
	// A future change to the env-var names is a breaking
	// change to the documented operator procedure in
	// staging-evidence.md. Guard the names here so a
	// rename shows up as a test failure on the rename
	// commit rather than as a silent documentation drift.
	if LiveEvaluationCostCeilingUSDEnv != "EVAL_LIVE_COST_CEILING_USD" {
		t.Fatalf("LiveEvaluationCostCeilingUSDEnv changed: got %q", LiveEvaluationCostCeilingUSDEnv)
	}
	if LiveEvaluationCostUSDEnv != "EVAL_LIVE_COST_USD" {
		t.Fatalf("LiveEvaluationCostUSDEnv changed: got %q", LiveEvaluationCostUSDEnv)
	}
	if LiveEvaluationOutputEnv != "EVAL_LIVE_OUTPUT" {
		t.Fatalf("LiveEvaluationOutputEnv changed: got %q", LiveEvaluationOutputEnv)
	}
}
