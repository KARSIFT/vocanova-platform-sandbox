package aifeedback

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// LiveEvaluationReport is private synthetic evaluation evidence. It separates
// adapter observations, measured failures and missing acceptance requirements.
// Fake-provider runs validate the harness, never a real model's quality.
type LiveEvaluationReport struct {
	// Configured retry allowance; actual transport attempts remain unmeasured.
	MaxRetries    int
	ThresholdSpec GoldenThresholdSpec
	// These are caller-supplied configuration metadata, not transport measurements.
	Timeout          time.Duration
	RequestInterval  time.Duration
	AcceptanceState  string
	AcceptanceGaps   []string
	CaseEvidence     []EvaluationCaseEvidence
	Scope            string
	GoldenSetVersion string
	PromptVersion    string
	SchemaVersion    string
	Commit           string
	// Logical-call elapsed time includes wrapper pacing and hidden adapter retries.
	LatencyDefinition string

	// Provider identifies the feedback provider the run
	// actually exercised (e.g. "opencode", "mock"). The
	// identity hook unwraps supported pacing wrappers without exposing config.
	Provider string
	// Model is the model identifier the provider was
	// invoked with (e.g. "opencode-go/deepseek-v4-pro").
	// For mock-provider runs this is the empty string.
	Model string
	// DatasetVersion is the dataset the run was executed
	// against (mirrors EvaluationResult.DatasetVersion
	// and GoldenThresholds.DatasetVersion).
	DatasetVersion string
	// SpecVersion identifies the threshold spec the
	// report's violations were checked against. "doc09-v1"
	// for DefaultGoldenThresholdSpec (the current
	// accepted DOC-09 §23 spec).
	SpecVersion string
	// Thresholds holds the per-threshold computed values
	// (see apps/api/business/aifeedback/threshold_gate.go
	// for the field list).
	Thresholds GoldenThresholds
	// Violations is the list of DOC-09 §23 thresholds
	// the run's computed values failed against Spec.
	// An empty slice means no measured violation was found; acceptance can
	// still be INCOMPLETE. Consult AcceptanceState and AcceptanceGaps.
	Violations []ThresholdViolation
	// Duration is the wall-clock time the entire run
	// took, measured around RunEvaluation. Recorded so
	// an operator can compare against the pre-agreed
	// latency budget without having to re-time the run.
	Duration time.Duration
	// PerCallLatency holds the per-call latencies
	// observed across all provider calls the run made
	// (in the order they happened). The summary
	// statistics are computed live by the library and
	// exposed as LatencyMin / LatencyMax / LatencyMean
	// / LatencyP50 / LatencyP95 below; the raw slice is
	// retained for reproducibility (a future
	// investigator can recompute P99 or a different
	// percentile from it).
	PerCallLatency []time.Duration
	// LatencyMin / LatencyMax / LatencyMean / LatencyP50
	// / LatencyP95 are the per-call latency summary
	// statistics. LatencyP50 and LatencyP95 are
	// computed with a deterministic nearest-rank method
	// so the same input produces the same numbers
	// across runs (the report is reproducible).
	LatencyMin  time.Duration
	LatencyMax  time.Duration
	LatencyMean time.Duration
	LatencyP50  time.Duration
	LatencyP95  time.Duration
	// ProviderCalls is the number of times the provider
	// was actually invoked. Mirrors Thresholds.ProviderCalled
	// for convenience; redundant by design so the report
	// is self-contained when the operator copies it into
	// staging-evidence.md.
	ProviderCalls int
	// EstimatedInputChars is the sum of the input
	// character counts across every provider call.
	// Combined with the provider's published
	// chars-per-token ratio (or the operator's
	// measurement of the same run), this lets the
	// operator estimate token usage without the
	// provider's response including it.
	EstimatedInputChars int
	// EstimatedOutputChars is the sum of the output
	// character counts across every provider call.
	// See EstimatedInputChars.
	EstimatedOutputChars int
	// CostUSD is the operator-supplied billed cost of
	// the run, recorded in US dollars. The library
	// itself does NOT populate this field (it has no
	// access to the provider's billing console). The
	// cmd/eval-live command exposes a --cost flag (or
	// EVAL_LIVE_COST_USD env var) the operator can use
	// to set the value from the provider's billing
	// dashboard; the value is then embedded in the
	// written report. A value of -1 means unknown. Zero is a valid
	// recorded zero cost. Missing/invalid billing metadata remains a gap.
	CostUSD float64
	// CostCeilingUSD is the pre-agreed cost ceiling the
	// run was operating under, per DOC-12 §9. The
	// cmd/eval-live command exposes a --cost-ceiling
	// flag (or EVAL_LIVE_COST_CEILING_USD env var) the
	// operator must set before the run; the value is
	// compared against CostUSD and surfaced as a
	// CostCeilingExceeded flag. A recorded breach makes acceptance FAIL.
	// This is a post-run comparison, never a spending guard or reservation.
	CostCeilingUSD float64
	// CostCeilingExceeded is true when CostUSD >
	// CostCeilingUSD. Recorded so a reviewer
	// scanning the report sees the ceiling status
	// without having to do the comparison.
	CostCeilingExceeded bool
	// StartedAt and FinishedAt are the wall-clock
	// start and end timestamps of the run, in UTC.
	// Recorded so a reviewer can correlate the report
	// with the provider's billing-dashboard entry for
	// the same window.
	StartedAt  time.Time
	FinishedAt time.Time
	// OperatorNotes is a free-text field the operator
	// can use to record anything not captured by the
	// structured fields (e.g. "rate-limit warning
	// observed at case 17; one retry succeeded").
	// Optional; empty when no notes were supplied.
	OperatorNotes string
}

// LiveEvaluationCostCeilingUSDEnv is the environment variable
// the cmd/eval-live command reads for the pre-agreed cost
// ceiling. Documented here so the staging-evidence.md
// "EV-22" procedure section can name the exact variable the
// operator must set, and so the cmd/eval-live main's
// --cost-ceiling flag's help text can reference the same name.
const LiveEvaluationCostCeilingUSDEnv = "EVAL_LIVE_COST_CEILING_USD"

// LiveEvaluationCostUSDEnv is the environment variable the
// cmd/eval-live command reads for the operator-supplied
// post-run billed cost. Documented here for the same reason
// as LiveEvaluationCostCeilingUSDEnv.
const LiveEvaluationCostUSDEnv = "EVAL_LIVE_COST_USD"

// LiveEvaluationOutputEnv is the environment variable the
// cmd/eval-live command reads for the optional output file
// path. When unset, the report is written to stdout only;
// when set, the report is also written to the named file in
// addition to stdout. Documented here so the
// staging-evidence.md "EV-22" procedure section can name the
// exact variable.
const LiveEvaluationOutputEnv = "EVAL_LIVE_OUTPUT"

// InstrumentedProvider wraps a FeedbackProvider so the
// per-call latency and the input/output character counts are
// captured for LiveEvaluationReport. It is a thin shim
// specifically scoped to T10 - it does not change the
// FeedbackProvider interface, RunEvaluation, or any other
// shared code - and is removed from the call path as soon as
// RunEvaluation returns.
//
// The instrumented provider is constructed with
// NewInstrumentedProvider(provider). The wrapper's
// PerCallLatency, InputChars, and OutputChars fields are
// populated as GenerateFeedback is called.
type InstrumentedProvider struct {
	inner FeedbackProvider
	// PerCallLatency is appended to on every call.
	PerCallLatency []time.Duration
	// InputChars is the sum of the character counts
	// from every task passed to GenerateFeedback. It is
	// a proxy for the prompt's input size; the real
	// provider may add framing tokens, but the
	// chars-per-token ratio is provider-specific and
	// out of scope for the library.
	InputChars int
	// OutputChars is the sum of the character counts
	// from every ProviderFeedback returned by
	// GenerateFeedback. Same proxy caveat as
	// InputChars.
	OutputChars int
}

// NewInstrumentedProvider returns nil for a nil provider. RunLiveEvaluation
// records that configuration as unavailable rather than claiming success.
func NewInstrumentedProvider(inner FeedbackProvider) *InstrumentedProvider {
	if inner == nil {
		return nil
	}
	return &InstrumentedProvider{inner: inner}
}

// GenerateFeedback forwards to the wrapped provider and
// records the call's wall-clock duration and character
// counts. The forwarding is otherwise transparent - errors
// from the inner provider are returned as-is and the
// per-call metrics are recorded even on error (a failed call
// has a non-zero latency and zero output chars, both of
// which are useful to record for the report).
//
// The character count is a rough estimate of the prompt's
// input size: the known string fields on ProviderTask
// (SystemPrompt, DeveloperPrompt) and the known string
// fields in UserPayload (learner_sentence, target_word,
// part_of_speech, target_meaning, learner_level) plus the
// length of any accepted_forms slice, summed. The estimate
// is deliberately not byte-for-byte the JSON the provider
// sees: the real wire format may add framing and the chars-
// per-token ratio is provider-specific. It is good enough
// for the report's "estimated_input_chars" field, which is
// the operator's input to a token-usage estimate, not a
// precise measurement.
func (p *InstrumentedProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	if p.inner == nil {
		return nil, errInstrumentedProviderEmpty
	}
	start := time.Now()
	feedback, err := p.inner.GenerateFeedback(ctx, task)
	elapsed := time.Since(start)
	p.PerCallLatency = append(p.PerCallLatency, elapsed)
	p.InputChars += taskInputCharCount(task)
	if feedback != nil {
		p.OutputChars += feedbackOutputCharCount(feedback)
	}
	return feedback, err
}

// taskInputCharCount returns the rough character count of
// the prompt's input. See the GenerateFeedback comment for
// what is and is not counted.
func taskInputCharCount(task ProviderTask) int {
	n := utf8.RuneCountInString(task.SystemPrompt) + utf8.RuneCountInString(task.DeveloperPrompt)
	for _, key := range []string{
		"learner_sentence",
		"target_word",
		"part_of_speech",
		"target_meaning",
		"learner_level",
	} {
		if v, ok := task.UserPayload[key].(string); ok {
			n += utf8.RuneCountInString(v)
		}
	}
	if forms, ok := task.UserPayload["accepted_forms"].([]string); ok {
		for _, f := range forms {
			n += utf8.RuneCountInString(f)
		}
	}
	return n
}

// feedbackOutputCharCount returns the rough character count
// of the response. The CorrectedSentence and ImprovementTip
// are *string fields and may be nil; both are dereferenced
// safely.
func feedbackOutputCharCount(feedback *ProviderFeedback) int {
	n := utf8.RuneCountInString(feedback.Status) + utf8.RuneCountInString(feedback.Explanation) + utf8.RuneCountInString(feedback.Headline) + utf8.RuneCountInString(feedback.Naturalness)
	if feedback.CorrectedSentence != nil {
		n += utf8.RuneCountInString(*feedback.CorrectedSentence)
	}
	if feedback.ImprovementTip != nil {
		n += utf8.RuneCountInString(*feedback.ImprovementTip)
	}
	return n
}

// errInstrumentedProviderEmpty is returned by
// InstrumentedProvider.GenerateFeedback when the wrapper
// was constructed with a nil inner provider. It is a
// distinct sentinel so the test suite can assert on it
// specifically.
var errInstrumentedProviderEmpty = &instrumentedProviderError{message: "InstrumentedProvider: inner provider is nil"}

// instrumentedProviderError is the error type
// InstrumentedProvider returns for its own pre-conditions.
// It is unexported because callers should not branch on it -
// the only public test for it is the existence of the error
// itself, which is enough to surface the programmer mistake.
type instrumentedProviderError struct {
	message string
}

func (e *instrumentedProviderError) Error() string { return e.message }

// LiveEvaluationOptions controls the optional parameters of
// RunLiveEvaluation. The zero value is valid: Cases
// defaults to GoldenSet(), Spec defaults to
// DefaultGoldenThresholdSpec(), OperatorNotes defaults to
// the empty string. CostUSD / CostCeilingUSD are supplied
// by the operator after the run (via env or flag) and are
// applied to the returned report, not consumed during the
// run itself.
type LiveEvaluationOptions struct {
	MaxRetries int
	// Non-secret configured timeout and wrapper pacing; the harness does not enforce them.
	Timeout         time.Duration
	RequestInterval time.Duration
	Commit          string

	// Cases is the dataset the run exercises. Nil
	// means GoldenSet().
	Cases []EvaluationCase
	// Spec is the threshold spec the report is checked
	// against. The zero GoldenThresholdSpec means
	// DefaultGoldenThresholdSpec().
	Spec GoldenThresholdSpec
	// CostCeilingUSD is the pre-agreed cost ceiling.
	// A value of -1 means "no ceiling recorded" (the
	// ceiling-exceeded check is skipped). A zero
	// value means "ceiling is zero dollars", which
	// any non-zero CostUSD will exceed; this is
	// intentional - the operator can set the ceiling
	// to zero to make a "spent anything" run fail
	// loudly.
	CostCeilingUSD float64
	// CostUSD is the operator-supplied post-run billed
	// cost. -1 means "not recorded"; other negative values are invalid.
	CostUSD float64
	// OperatorNotes is the free-text notes the
	// operator attached to the run.
	OperatorNotes string
}

// RunLiveEvaluation executes adapter diagnostics and returns private case
// evidence. AcceptanceState is FAIL for observed violations, INCOMPLETE for
// missing evidence, and PASS only when no acceptance requirements are missing.
// This harness does not exercise the full production service.
func RunLiveEvaluation(ctx context.Context, provider FeedbackProvider, opts LiveEvaluationOptions) LiveEvaluationReport {
	startedAt := time.Now()
	if opts.Cases == nil {
		opts.Cases = GoldenSet()
	}
	spec := opts.Spec
	zeroSpec := GoldenThresholdSpec{}
	if spec == zeroSpec {
		spec = DefaultGoldenThresholdSpec()
	}
	var metadataGaps []string
	specVersion := "doc09-v1"
	if spec != DefaultGoldenThresholdSpec() {
		specVersion = "custom-diagnostic"
		metadataGaps = append(metadataGaps, "custom or disabled threshold bounds do not establish DOC-09 acceptance")
	}
	for _, bound := range []*float64{
		&spec.StructuredOutputValidFirstResponse, &spec.StructuredOutputValidAfterOneRepair,
		&spec.OverallStatusAccuracy, &spec.ClearlyCorrectAccuracy, &spec.ClearlyIncorrectTargetUseAccuracy,
		&spec.UnnecessaryCorrectionOnClearlyCorrect, &spec.MeaningPreservation, &spec.CorrectSelfHarmIntervention,
	} {
		if math.IsNaN(*bound) || math.IsInf(*bound, 0) || *bound < NotTracked || (*bound < 0 && *bound != NotTracked) || *bound > 1 {
			*bound = NotTracked
			metadataGaps = append(metadataGaps, "invalid threshold bound is unscored")
		}
	}
	ip := NewInstrumentedProvider(provider)
	var evaluatedProvider FeedbackProvider
	if ip == nil {
		// Keep the nil provider visible to RunEvaluation. The empty metrics
		// holder must never be invoked as though an adapter were available.
		ip = &InstrumentedProvider{}
	} else {
		evaluatedProvider = ip
	}
	result := RunEvaluation(ctx, evaluatedProvider, opts.Cases)
	finishedAt := time.Now()
	computed := ComputeGoldenThresholds(result, opts.Cases)
	violations := CheckGoldenThresholds(spec, computed)
	latencyStats := summarizeLatencies(ip.PerCallLatency)
	acceptanceState, gaps := EvaluationAcceptance(computed, violations)
	gaps = append(gaps, metadataGaps...)
	if opts.MaxRetries < 0 {
		gaps = append(gaps, "configured retry allowance is invalid")
	}
	if math.IsNaN(opts.CostUSD) || math.IsInf(opts.CostUSD, 0) || (opts.CostUSD < 0 && opts.CostUSD != -1) {
		gaps = append(gaps, "invalid billed-cost metadata")
		opts.CostUSD = -1
	}
	if math.IsNaN(opts.CostCeilingUSD) || math.IsInf(opts.CostCeilingUSD, 0) || (opts.CostCeilingUSD < 0 && opts.CostCeilingUSD != -1) {
		gaps = append(gaps, "invalid cost-ceiling metadata")
		opts.CostCeilingUSD = -1
	}
	if opts.CostUSD == -1 {
		gaps = append(gaps, "billed cost is unrecorded")
	}
	if opts.CostCeilingUSD == -1 {
		gaps = append(gaps, "agreed cost ceiling is unrecorded")
	}
	if opts.Timeout <= 0 {
		gaps = append(gaps, "configured provider timeout is unrecorded or invalid")
	}
	if opts.RequestInterval < 0 {
		gaps = append(gaps, "configured pacing interval is invalid")
	}
	if opts.Commit == "" {
		opts.Commit = "unknown"
	}
	if opts.Commit == "unknown" {
		gaps = append(gaps, "source commit is unrecorded")
	}
	if providerName(provider) == "unknown" {
		gaps = append(gaps, "provider identity is unknown")
	}
	if providerName(provider) == "mock" {
		gaps = append(gaps, "fake-provider results exercise the harness and do not establish model quality")
	}
	if len(gaps) > 0 && acceptanceState == "PASS" {
		acceptanceState = "INCOMPLETE"
	}
	if opts.CostCeilingUSD >= 0 && opts.CostUSD > opts.CostCeilingUSD {
		acceptanceState = "FAIL"
	}
	report := LiveEvaluationReport{
		MaxRetries: opts.MaxRetries, ThresholdSpec: spec,
		Timeout: opts.Timeout, RequestInterval: opts.RequestInterval,
		AcceptanceState: acceptanceState, AcceptanceGaps: gaps,
		CaseEvidence: result.CaseEvidence, Scope: result.Scope,
		GoldenSetVersion: GoldenSetVersion,
		PromptVersion:    PromptVersionSentenceFeedbackV1, SchemaVersion: SchemaVersionFeedbackV1,
		Commit:            opts.Commit,
		LatencyDefinition: "logical adapter-call elapsed time, including wrapper pacing and adapter retries; nearest-rank percentiles",

		Provider:             providerName(ip.inner),
		Model:                providerModel(ip.inner),
		DatasetVersion:       result.DatasetVersion,
		SpecVersion:          specVersion,
		Thresholds:           computed,
		Violations:           violations,
		Duration:             finishedAt.Sub(startedAt),
		PerCallLatency:       ip.PerCallLatency,
		LatencyMin:           latencyStats.min,
		LatencyMax:           latencyStats.max,
		LatencyMean:          latencyStats.mean,
		LatencyP50:           latencyStats.p50,
		LatencyP95:           latencyStats.p95,
		ProviderCalls:        len(ip.PerCallLatency),
		EstimatedInputChars:  ip.InputChars,
		EstimatedOutputChars: ip.OutputChars,
		CostUSD:              opts.CostUSD,
		CostCeilingUSD:       opts.CostCeilingUSD,
		CostCeilingExceeded:  opts.CostUSD > opts.CostCeilingUSD && opts.CostCeilingUSD >= 0,
		StartedAt:            startedAt.UTC(),
		FinishedAt:           finishedAt.UTC(),
		OperatorNotes:        opts.OperatorNotes,
	}
	return report
}

// EvaluationProviderIdentity exposes only non-secret provider/model identifiers.
// Wrappers should forward this optional contract; never return URLs or config.
type EvaluationProviderIdentity interface {
	EvaluationIdentity() (provider string, model string)
}

func ProviderEvaluationIdentity(p FeedbackProvider) (string, string) {
	if identity, ok := p.(EvaluationProviderIdentity); ok {
		return identity.EvaluationIdentity()
	}
	switch v := p.(type) {
	case *OpenCodeFeedbackProvider:
		if v != nil {
			return "opencode", v.config.Model
		}
	case *CloudflareFeedbackProvider:
		if v != nil && v.cloudflareTransport != nil {
			return "cloudflare", v.config.Model
		}
	case *GeminiFeedbackProvider:
		if v != nil && v.geminiTransport != nil {
			return "gemini", v.config.Model
		}
	case *MockProvider:
		return "mock", ""
	case *InstrumentedProvider:
		if v != nil {
			return ProviderEvaluationIdentity(v.inner)
		}
	}
	return "unknown", ""
}
func providerName(p FeedbackProvider) string { name, _ := ProviderEvaluationIdentity(p); return name }
func providerModel(p FeedbackProvider) string {
	_, model := ProviderEvaluationIdentity(p)
	return model
}

// latencySummary holds the per-call latency summary
// statistics computed by summarizeLatencies. All fields
// are zero when the input slice is empty (the run made
// zero provider calls - every case was intercepted by the
// safety layer or failed validation).
type latencySummary struct {
	min  time.Duration
	max  time.Duration
	mean time.Duration
	p50  time.Duration
	p95  time.Duration
}

// summarizeLatencies uses nearest-rank percentiles: ceil(p*n/100), one-based.
// A zero-length input produces a zero summary.
func summarizeLatencies(latencies []time.Duration) latencySummary {
	if len(latencies) == 0 {
		return latencySummary{}
	}
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	min := sorted[0]
	max := sorted[len(sorted)-1]
	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}
	mean := sum / time.Duration(len(sorted))
	return latencySummary{
		min:  min,
		max:  max,
		mean: mean,
		p50:  percentileNearestRank(sorted, 50),
		p95:  percentileNearestRank(sorted, 95),
	}
}

// percentileNearestRank uses ceil(p*n/100), clamped to the observed range.
func percentileNearestRank(sorted []time.Duration, p int) time.Duration {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	rank := (p*n + 99) / 100
	if rank < 1 {
		rank = 1
	}
	return sorted[rank-1]
}

// FormatLiveEvaluationReport includes the complete synthetic case observations.
// Keep this output private: provider-generated content is untrusted evidence for
// human review, not material to copy wholesale into CI or public release notes.
func FormatLiveEvaluationReport(r LiveEvaluationReport) string {
	var s strings.Builder
	s.WriteString("=== T10 Live AI Evaluation Report ===\n")
	s.WriteString("Provider: " + r.Provider + "\n")
	s.WriteString("Model: " + r.Model + "\n")
	s.WriteString("Dataset: " + r.DatasetVersion + "\n")
	s.WriteString("Spec: " + r.SpecVersion + "\n")
	s.WriteString("Scope: " + r.Scope + "\nGoldenSet: " + r.GoldenSetVersion + "\nPrompt: " + r.PromptVersion + "\nSchema: " + r.SchemaVersion + "\nCommit: " + r.Commit + "\n")
	s.WriteString("LatencyDefinition: " + r.LatencyDefinition + "\n")
	s.WriteString("ConfiguredTimeout: " + r.Timeout.String() + "\nRequestInterval: " + r.RequestInterval.String() + "\nConfiguredMaxRetries: " + itoa(r.MaxRetries) + "\n")
	specJSON, _ := json.Marshal(r.ThresholdSpec)
	s.WriteString("ThresholdSpec: " + string(specJSON) + "\n")
	for _, gap := range r.AcceptanceGaps {
		s.WriteString("AcceptanceGap: " + gap + "\n")
	}
	s.WriteString("StartedAt: " + r.StartedAt.Format(time.RFC3339) + "\n")
	s.WriteString("FinishedAt: " + r.FinishedAt.Format(time.RFC3339) + "\n")
	s.WriteString("Duration: " + r.Duration.String() + "\n")
	s.WriteString("ProviderCalls: " + itoa(r.ProviderCalls) + "\n")
	s.WriteString("EstimatedInputChars: " + itoa(r.EstimatedInputChars) + "\n")
	s.WriteString("EstimatedOutputChars: " + itoa(r.EstimatedOutputChars) + "\n")
	s.WriteString("CostUSD: " + ftoa(r.CostUSD) + "\n")
	s.WriteString("CostCeilingUSD: " + ftoa(r.CostCeilingUSD) + "\n")
	if r.CostCeilingUSD < 0 {
		s.WriteString("CostCeilingExceeded: (ceiling not set; not enforced)\n")
	} else {
		if r.CostCeilingExceeded {
			s.WriteString("CostCeilingExceeded: true\n")
		} else {
			s.WriteString("CostCeilingExceeded: false\n")
		}
	}
	s.WriteString("LatencyMin: " + r.LatencyMin.String() + "\n")
	s.WriteString("LatencyMax: " + r.LatencyMax.String() + "\n")
	s.WriteString("LatencyMean: " + r.LatencyMean.String() + "\n")
	s.WriteString("LatencyP50: " + r.LatencyP50.String() + "\n")
	s.WriteString("LatencyP95: " + r.LatencyP95.String() + "\n")
	if r.OperatorNotes != "" {
		s.WriteString("OperatorNotes: " + r.OperatorNotes + "\n")
	}
	s.WriteString("--- Per-threshold computed values ---\n")
	s.WriteString(FormatThresholdReport(r.Thresholds, r.Violations))
	// Case evidence is synthetic-only private report data, not a standard log.
	evidenceJSON, err := json.MarshalIndent(r.CaseEvidence, "", "  ")
	if err == nil {
		s.WriteString("--- Per-case evidence (human review pending) ---\n" + string(evidenceJSON) + "\n")
	}
	state := r.AcceptanceState
	if state == "" {
		state = "INCOMPLETE"
	}
	s.WriteString("=== Result: " + state + " ===\n")
	return s.String()
}

func itoa(i int) string { return strconv.Itoa(i) }

// ftoa uses standard fixed-point formatting without an integer-cents conversion.
// Finite large costs remain representable; -1 stays the unknown-cost sentinel.
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }
