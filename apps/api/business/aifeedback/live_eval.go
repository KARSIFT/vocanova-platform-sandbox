package aifeedback

import (
	"context"
	"sort"
	"time"
)

// LiveEvaluationReport is the structured output of a single
// T10 live-provider evaluation pass. It is the data T10 records
// in `staging-evidence.md` after the one real run, and the
// shape the cmd/eval-live command writes to stdout (and
// optionally to a file) when the founder has provisioned the
// staging AI-provider credentials (VOC-032-DEP-03).
//
// The report is intentionally structured (not free-text) so
// every DOC-09 §23 threshold T10 must record has a named
// field, and so the same report shape is what the test suite
// asserts against the fake-provider path (proving the
// procedure is exercisable end-to-end without depending on a
// real provider call). The provider-supplied data the report
// cannot determine on its own (billed cost, exact token
// usage) is in fields the operator fills in from the
// provider's billing console / dashboard after the run;
// leaving them as named fields rather than guessing them
// keeps the recorded evidence honest (a "not recorded" or
// "0" cost is a fact, not a fabricated number).
//
// LiveEvaluationReport is the single source of truth for T10
// evidence: every value the staging-evidence.md "EV-22"
// section requires has a field here.
type LiveEvaluationReport struct {
	// Provider identifies the feedback provider the run
	// actually exercised (e.g. "opencode", "mock"). The
	// cmd/eval-live command always reports the value
	// supplied to NewOpenCodeFeedbackProvider; tests
	// against a fake provider report the fake's name.
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
	// An empty slice means every tracked threshold was
	// met. The cmd/eval-live command's exit code is
	// derived from the slice being non-empty.
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
	// written report. Recording a zero or negative
	// value is treated as "cost not recorded" and the
	// report is still valid; the field exists so a
	// recorded cost is a recorded fact, not an empty
	// cell.
	CostUSD float64
	// CostCeilingUSD is the pre-agreed cost ceiling the
	// run was operating under, per DOC-12 §9. The
	// cmd/eval-live command exposes a --cost-ceiling
	// flag (or EVAL_LIVE_COST_CEILING_USD env var) the
	// operator must set before the run; the value is
	// compared against CostUSD and surfaced as a
	// CostCeilingExceeded flag (not as a threshold
	// violation - the threshold gate is the
	// release-blocking check; the cost ceiling is the
	// pre-run agreement, recorded for the reviewer's
	// benefit, not enforced as a gate).
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

// NewInstrumentedProvider returns an InstrumentedProvider
// wrapping inner. inner must be non-nil; a nil inner is
// treated as a programmer error and returns nil so the
// caller's later GenerateFeedback call panics loudly rather
// than silently producing a "successful" report with zero
// provider calls.
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
	n := len(task.SystemPrompt) + len(task.DeveloperPrompt)
	for _, key := range []string{
		"learner_sentence",
		"target_word",
		"part_of_speech",
		"target_meaning",
		"learner_level",
	} {
		if v, ok := task.UserPayload[key].(string); ok {
			n += len(v)
		}
	}
	if forms, ok := task.UserPayload["accepted_forms"].([]string); ok {
		for _, f := range forms {
			n += len(f)
		}
	}
	return n
}

// feedbackOutputCharCount returns the rough character count
// of the response. The CorrectedSentence and ImprovementTip
// are *string fields and may be nil; both are dereferenced
// safely.
func feedbackOutputCharCount(feedback *ProviderFeedback) int {
	n := len(feedback.Status) + len(feedback.Explanation)
	if feedback.CorrectedSentence != nil {
		n += len(*feedback.CorrectedSentence)
	}
	if feedback.ImprovementTip != nil {
		n += len(*feedback.ImprovementTip)
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
	// Cases is the dataset the run exercises. Nil
	// means GoldenSet().
	Cases []EvaluationCase
	// Spec is the threshold spec the report is checked
	// against. The zero GoldenThresholdSpec means
	// DefaultGoldenThresholdSpec().
	Spec GoldenThresholdSpec
	// CostCeilingUSD is the pre-agreed cost ceiling.
	// A negative value means "no ceiling" (the
	// ceiling-exceeded check is skipped). A zero
	// value means "ceiling is zero dollars", which
	// any non-zero CostUSD will exceed; this is
	// intentional - the operator can set the ceiling
	// to zero to make a "spent anything" run fail
	// loudly.
	CostCeilingUSD float64
	// CostUSD is the operator-supplied post-run billed
	// cost. Negative means "not recorded".
	CostUSD float64
	// OperatorNotes is the free-text notes the
	// operator attached to the run.
	OperatorNotes string
}

// RunLiveEvaluation runs the dataset against the supplied
// provider, instruments the run for latency and character
// counts, computes the per-threshold values, and returns
// the structured report. It is the T10-deliverable
// library function: same threshold-computation logic
// T08's mock gate uses, but exercised against an
// arbitrary (real or fake) provider.
//
// The function does NOT make any policy decision: it
// returns the report and the violation list, leaving the
// caller's exit-code choice to itself. The cmd/eval-live
// command's main() is the caller that translates
// "violations non-empty" into "exit 1".
//
// The function is safe to call against a fake provider in
// tests: every code path it exercises is independent of
// network or external services.
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
	ip := NewInstrumentedProvider(provider)
	result := RunEvaluation(ctx, ip, opts.Cases)
	finishedAt := time.Now()
	computed := ComputeGoldenThresholds(result, opts.Cases)
	violations := CheckGoldenThresholds(spec, computed)
	latencyStats := summarizeLatencies(ip.PerCallLatency)
	report := LiveEvaluationReport{
		Provider:             providerName(ip.inner),
		Model:                providerModel(ip.inner),
		DatasetVersion:       result.DatasetVersion,
		SpecVersion:          "doc09-v1",
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

// providerName is a small reflection-free hook that returns
// the provider's identifying name for the report. It is
// implemented as a type switch so the library does not have
// to know about every concrete FeedbackProvider type via
// registration; if a future provider is added, the
// production-wiring site's own type switch is the place
// to add the matching case (or to introduce a Provider
// interface method that returns the name).
func providerName(p FeedbackProvider) string {
	switch v := p.(type) {
	case *OpenCodeFeedbackProvider:
		return "opencode"
	case *MockProvider:
		return "mock"
	default:
		_ = v
		return "unknown"
	}
}

// providerModel is the parallel hook for the model
// identifier. Only the OpenCode provider currently
// exposes one; everything else reports the empty string.
func providerModel(p FeedbackProvider) string {
	if oc, ok := p.(*OpenCodeFeedbackProvider); ok {
		return oc.config.Model
	}
	return ""
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

// summarizeLatencies computes the per-call latency
// summary statistics. The percentile method is the
// nearest-rank method (NIST handbook §7.2.1.1) with the
// standard floor((n-1)*p/100)+1 rank, applied to a
// pre-sorted slice. The function is deterministic: the
// same input slice produces the same output values.
//
// A zero-length input produces the zero latencySummary.
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

// percentileNearestRank returns the p-th percentile of the
// (assumed sorted ascending) input using the nearest-rank
// method. p is clamped to [0, 100]. A zero-length input
// returns the zero duration.
//
// The nearest-rank formula is:
//
//	rank = ceil(p / 100 * n)   (1-indexed)
//
// or equivalently in 0-indexed form:
//
//	rank = floor(p / 100 * (n-1))   (0-indexed)
//
// Go's integer division makes the 0-indexed form
// straightforward and is what this implementation uses.
// Both forms produce the same result for the same input;
// the 0-indexed form is what NIST's example uses.
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
	rank := (p * (n - 1)) / 100
	return sorted[rank]
}

// FormatLiveEvaluationReport renders a LiveEvaluationReport
// as a human-readable, multi-line plain-text block. The
// format is the same the cmd/eval-live command writes to
// stdout and to the optional output file, so the operator
// can copy-paste the rendered block straight into
// staging-evidence.md's "EV-22" section. The function is
// deterministic and produces stable output for a given
// input; tests assert on substrings, not on byte-for-byte
// equality, so a future field addition does not break
// existing assertions.
//
// Every field on LiveEvaluationReport appears in the
// rendered output - a report that omits a field is
// unusable as evidence, so the omission is not allowed.
func FormatLiveEvaluationReport(r LiveEvaluationReport) string {
	var s string
	s = s + "=== T10 Live AI Evaluation Report ===\n"
	s = s + "Provider: " + r.Provider + "\n"
	s = s + "Model: " + r.Model + "\n"
	s = s + "Dataset: " + r.DatasetVersion + "\n"
	s = s + "Spec: " + r.SpecVersion + "\n"
	s = s + "StartedAt: " + r.StartedAt.Format(time.RFC3339) + "\n"
	s = s + "FinishedAt: " + r.FinishedAt.Format(time.RFC3339) + "\n"
	s = s + "Duration: " + r.Duration.String() + "\n"
	s = s + "ProviderCalls: " + itoa(r.ProviderCalls) + "\n"
	s = s + "EstimatedInputChars: " + itoa(r.EstimatedInputChars) + "\n"
	s = s + "EstimatedOutputChars: " + itoa(r.EstimatedOutputChars) + "\n"
	s = s + "CostUSD: " + ftoa(r.CostUSD) + "\n"
	s = s + "CostCeilingUSD: " + ftoa(r.CostCeilingUSD) + "\n"
	if r.CostCeilingUSD < 0 {
		s = s + "CostCeilingExceeded: (ceiling not set; not enforced)\n"
	} else {
		if r.CostCeilingExceeded {
			s = s + "CostCeilingExceeded: true\n"
		} else {
			s = s + "CostCeilingExceeded: false\n"
		}
	}
	s = s + "LatencyMin: " + r.LatencyMin.String() + "\n"
	s = s + "LatencyMax: " + r.LatencyMax.String() + "\n"
	s = s + "LatencyMean: " + r.LatencyMean.String() + "\n"
	s = s + "LatencyP50: " + r.LatencyP50.String() + "\n"
	s = s + "LatencyP95: " + r.LatencyP95.String() + "\n"
	if r.OperatorNotes != "" {
		s = s + "OperatorNotes: " + r.OperatorNotes + "\n"
	}
	s = s + "--- Per-threshold computed values ---\n"
	s = s + FormatThresholdReport(r.Thresholds, r.Violations)
	if len(r.Violations) == 0 {
		s = s + "=== Result: PASS (every tracked DOC-09 §23 threshold met) ===\n"
	} else {
		s = s + "=== Result: FAIL (" + itoa(len(r.Violations)) + " tracked threshold(s) violated) ===\n"
	}
	return s
}

// itoa is a small helper for rendering integer values
// without dragging in fmt just for this. The cmd/eval-live
// command imports fmt directly; the library keeps its
// surface lean to keep the test assertions stable.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// ftoa renders a float64 with a fixed two-decimal precision
// for the report. Negative values render with a leading
// minus (so a "-1" cost from the operator is rendered as
// "-1.00" and distinguishable from a "0.00" not-recorded).
// A NaN renders as "NaN" so an operator who accidentally
// pipes a non-finite value does not produce a silent zero.
func ftoa(f float64) string {
	if f != f {
		return "NaN"
	}
	neg := f < 0
	if neg {
		f = -f
	}
	cents := int64(f*100 + 0.5)
	whole := cents / 100
	frac := cents % 100
	out := itoa64(whole)
	out += "."
	if frac < 10 {
		out += "0"
	}
	out += itoa64(frac)
	if neg {
		out = "-" + out
	}
	return out
}

// itoa64 is the int64 counterpart of itoa, used by ftoa.
// It is a separate function because Go's strconv.Itoa
// rejects int64, and the report's cost field is a float64
// whose integer part is built from cents / 100.
func itoa64(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
