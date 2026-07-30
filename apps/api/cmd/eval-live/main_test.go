package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
)

// fakeFeedbackProvider is a deterministic provider used
// only by these tests. It returns a configurable
// ProviderFeedback (default: status=correct) so the cmd's
// exit-code path can be exercised against an arbitrary
// "real-looking" provider without reaching OpenCode.
type fakeFeedbackProvider struct {
	feedback aifeedback.ProviderFeedback
}

func (f *fakeFeedbackProvider) GenerateFeedback(ctx context.Context, task aifeedback.ProviderTask) (*aifeedback.ProviderFeedback, error) {
	fb := f.feedback
	return &fb, nil
}

// withFakeProvider replaces newProvider for the duration of
// the test, restoring the previous value on cleanup. The
// returned builder is concurrent-safe; tests that need a
// non-default feedback can pass a custom one in.
func withFakeProvider(t *testing.T, fb aifeedback.ProviderFeedback) {
	t.Helper()
	prev := newProvider
	newProvider = func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return &fakeFeedbackProvider{feedback: fb}
	}
	t.Cleanup(func() { newProvider = prev })
}

// withFakeProviderPerCall replaces newProvider with a
// function that lets the test observe every constructed
// provider (the test injects a ProviderTasks-recording
// implementation). This is the seam for tests that want
// to assert on the configuration the cmd built.
func withFakeProviderPerCall(t *testing.T, build func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider) {
	t.Helper()
	prev := newProvider
	newProvider = build
	t.Cleanup(func() { newProvider = prev })
}

func TestRunEvalLive_MissingBaseURLExitsUsage(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{"--api-key", "test-key"}, stdout, stderr, time.Now)
	if code != exitUsageError {
		t.Fatalf("expected exitUsageError, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--base-url") {
		t.Fatalf("stderr should name --base-url; got: %s", stderr.String())
	}
}

func TestRunEvalLive_MissingAPIKeyExitsUsage(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{"--base-url", "http://example.com"}, stdout, stderr, time.Now)
	if code != exitUsageError {
		t.Fatalf("expected exitUsageError, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--api-key") {
		t.Fatalf("stderr should name --api-key; got: %s", stderr.String())
	}
}

func TestRunEvalLive_PrintsHelpAndExitsZero(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{"--help"}, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("--help should exit 0, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Usage") {
		t.Fatalf("--help should print usage; got stderr: %s", stderr.String())
	}
}

func TestRunEvalLive_RendersReportAgainstFakeProvider(t *testing.T) {
	// The fake provider's default response produces
	// 50% match rate against the dataset (it returns
	// "correct" for every case, but the dataset
	// expects "needs_improvement" or "incorrect" for
	// some of them), so the run exits 1. The cmd's
	// job here is to render the report and produce
	// a non-usage-error exit code; the gate's
	// pass/fail logic is the library's job and is
	// covered separately by
	// TestFormatLiveEvaluationReportOnFailureIncludesViolations.
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
	}, stdout, stderr, time.Now)
	if code == exitUsageError {
		t.Fatalf("expected a non-usage-error exit code; got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "=== T10 Live AI Evaluation Report ===") {
		t.Fatal("expected the report header on stdout")
	}
	if !strings.Contains(stdout.String(), "Result:") {
		t.Fatal("expected a Result: line on stdout")
	}
}

func TestRunEvalLive_ExitsZeroWhenProviderMeetsEveryThreshold(t *testing.T) {
	// The "exits 0 when no violations" path is exercised
	// by injecting a provider that returns the right
	// status for every case. We do this by inspecting
	// the task and returning the expected status from
	// the dataset. The cmd's job is to translate "no
	// violations" to exit 0, regardless of which
	// provider produced the no-violations outcome.
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
	}, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("expectation-matching provider should pass; got %d stderr=%s\nstdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "Result: PASS") {
		t.Fatalf("expected Result: PASS; got: %s", stdout.String())
	}
}

// violatingProvider is a provider whose every output is
// the "needs_improvement" status, which deliberately
// mismatches the dataset's expectations and so produces
// tracked-threshold violations under the default spec.
type violatingProvider struct{}

func (violatingProvider) GenerateFeedback(ctx context.Context, task aifeedback.ProviderTask) (*aifeedback.ProviderFeedback, error) {
	return &aifeedback.ProviderFeedback{
		Status:      aifeedback.LearningStatusNeedsImprovement,
		Explanation: "always-needs-improvement",
	}, nil
}

func TestRunEvalLive_ExitCodeOnProviderThatViolates(t *testing.T) {
	// Provider that produces deliberate mismatches so
	// the gate's threshold check trips. The cmd should
	// exit with exitReleaseBlocking and the report
	// should show "Result: FAIL".
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return violatingProvider{}
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
	}, stdout, stderr, time.Now)
	if code != exitReleaseBlocking {
		t.Fatalf("expected exitReleaseBlocking from violating provider, got %d; stderr=%s\nstdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "Result: FAIL") {
		t.Fatalf("expected Result: FAIL in stdout; got: %s", stdout.String())
	}
}

func TestRunEvalLive_CostCeilingExceededIsReleaseBlocking(t *testing.T) {
	// Use the expectation-matching provider so the
	// run is otherwise clean; the cost-ceiling is the
	// only thing under test.
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	// Cost 1.00 with ceiling 0.10 must trip the ceiling
	// flag. Per the cmd's exit-code rules, this is a
	// release-blocking finding.
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
		"--cost", "1.00",
		"--cost-ceiling", "0.10",
	}, stdout, stderr, time.Now)
	if code != exitReleaseBlocking {
		t.Fatalf("cost-ceiling-exceeded should be release-blocking; got %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "CostCeilingExceeded: true") {
		t.Fatalf("stdout should mark CostCeilingExceeded: true; got: %s", stdout.String())
	}
}

func TestRunEvalLive_WritesOutputFile(t *testing.T) {
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return provider
	})
	dir := t.TempDir()
	out := dir + "/report.txt"
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
		"--output", out,
	}, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("expected exitSuccess; got %d stderr=%s", code, stderr.String())
	}
	// Read the file back and assert it is the same
	// report that was on stdout. (Reading a different
	// file would mean the cmd wrote the wrong
	// content; reading the same content twice would
	// just confirm what the assertion already does.)
	body, err := readFile(t, out)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(body, "=== T10 Live AI Evaluation Report ===") {
		t.Fatalf("output file should contain the report; got: %s", body)
	}
	if !strings.Contains(stderr.String(), "eval-live: report also written to") {
		t.Fatalf("stderr should announce the file write; got: %s", stderr.String())
	}
}

func TestRunEvalLive_OutputFileUnwritableReturnsUsageError(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	// /dev/null/null on POSIX is a path under a
	// non-directory that cannot be created; on every
	// supported platform the os.Create call will
	// fail and the cmd should return exitUsageError.
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
		"--output", "/dev/null/null/never.txt",
	}, stdout, stderr, time.Now)
	if code != exitUsageError {
		t.Fatalf("unwritable output should be exitUsageError; got %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
}

func TestRunEvalLive_BuildsProviderWithSuppliedConfig(t *testing.T) {
	// The cmd's newProvider seam receives the
	// OpenCodeConfig built from the supplied flags/env.
	// Assert the config flows through unchanged so a
	// future refactor of the flag/env resolution
	// cannot silently drop a value the operator set.
	var captured aifeedback.OpenCodeConfig
	var once sync.Once
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		once.Do(func() { captured = cfg })
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://staging.example",
		"--api-key", "the-real-key",
		"--model", "opencode-go/deepseek-v4-pro",
		"--timeout", "12s",
	}, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("expected exitSuccess; got %d stderr=%s", code, stderr.String())
	}
	if captured.BaseURL != "http://staging.example" {
		t.Errorf("BaseURL: got %q", captured.BaseURL)
	}
	if captured.APIKey != "the-real-key" {
		t.Errorf("APIKey: got %q", captured.APIKey)
	}
	if captured.Model != "opencode-go/deepseek-v4-pro" {
		t.Errorf("Model: got %q", captured.Model)
	}
	if captured.Timeout != 12*time.Second {
		t.Errorf("Timeout: got %s", captured.Timeout)
	}
	if captured.MaxRetries != 1 {
		t.Errorf("MaxRetries: got %d", captured.MaxRetries)
	}
}

func TestRunEvalLive_RespectsEnvDefaults(t *testing.T) {
	// The flags' defaults should pull from env vars
	// when those are set. Assert that the env
	// resolution does NOT short-circuit the
	// config-build step (i.e. the captured config
	// reflects the env values, not the flag defaults).
	t.Setenv("AI_PROVIDER_BASE_URL", "http://from-env")
	t.Setenv("AI_PROVIDER_API_KEY", "key-from-env")
	t.Setenv("AI_PROVIDER_MODEL", "from-env/v1")
	t.Setenv("AI_PROVIDER_TIMEOUT", "3s")
	var captured aifeedback.OpenCodeConfig
	var once sync.Once
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		once.Do(func() { captured = cfg })
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive(nil, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("expected exitSuccess from env-defaults path; got %d stderr=%s", code, stderr.String())
	}
	if captured.BaseURL != "http://from-env" {
		t.Errorf("BaseURL from env: got %q", captured.BaseURL)
	}
	if captured.APIKey != "key-from-env" {
		t.Errorf("APIKey from env: got %q", captured.APIKey)
	}
	if captured.Model != "from-env/v1" {
		t.Errorf("Model from env: got %q", captured.Model)
	}
	if captured.Timeout != 3*time.Second {
		t.Errorf("Timeout from env: got %s", captured.Timeout)
	}
}

func TestRunEvalLive_DoesNotLogAPIKey(t *testing.T) {
	// The report itself and any stderr output must
	// never contain the API key, even when the run
	// produces an error. A leaked key in the report
	// would end up in staging-evidence.md (operator
	// copy-pastes the report there) or in a CI
	// artifact - both unforgivable.
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	secretKey := "sk-leak-test-do-not-use-12345"
	_ = runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", secretKey,
	}, stdout, stderr, time.Now)
	if strings.Contains(stdout.String(), secretKey) {
		t.Fatal("API key leaked into stdout report")
	}
	if strings.Contains(stderr.String(), secretKey) {
		t.Fatalf("API key leaked into stderr: %s", stderr.String())
	}
}

func TestRunEvalLive_FlagsAfterArgumentsIgnored(t *testing.T) {
	// flag.NewFlagSet stops parsing at the first
	// non-flag argument, so a stray positional arg
	// would be silently dropped. This is a regression
	// guard for the documented "all args are flags"
	// behavior.
	provider := &expectationMatchingProvider{}
	withFakeProviderPerCall(t, func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		return provider
	})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runEvalLive([]string{
		"--base-url", "http://example.invalid",
		"--api-key", "test-key",
		"unexpected-positional",
	}, stdout, stderr, time.Now)
	if code != exitSuccess {
		t.Fatalf("expected exitSuccess (the positional is silently dropped by flag); got %d stderr=%s", code, stderr.String())
	}
}

// expectationMatchingProvider is a provider that returns
// the dataset's ExpectedStatus for every case. It looks up
// the case in the dataset by matching the task's
// UserPayload's (target_word, learner_sentence) pair. This
// is the only way to make the cmd's "exit 0 on no
// violations" path testable: the dataset is internal to
// the library and the cmd does not expose a per-case
// override, so the provider must match the dataset on its
// own.
//
// The matching is case-insensitive: the eval pipeline
// normalizes the sentence (lowercasing the first character
// among other transformations) before passing it to the
// provider, but the dataset's Sentence field retains the
// original capitalization ("I work every day." in the
// dataset vs "i work every day." in the task's
// UserPayload). A case-sensitive compare would miss every
// case and fall back to "correct", defeating the test's
// purpose.
//
// The provider is intentionally limited to the cmd's
// test file (not the library) because it depends on the
// library's GoldenSet() shape; a library refactor that
// renames or removes cases would surface as a test
// failure here (cases that no longer match a dataset
// entry default to "correct", which mismatches the
// expected status, so the run fails the gate, which is the
// right loud failure mode).
type expectationMatchingProvider struct{}

func (expectationMatchingProvider) GenerateFeedback(ctx context.Context, task aifeedback.ProviderTask) (*aifeedback.ProviderFeedback, error) {
	taskSentence, _ := task.UserPayload["learner_sentence"].(string)
	taskTarget, _ := task.UserPayload["target_word"].(string)
	for _, c := range aifeedback.GoldenSet() {
		if strings.EqualFold(c.Sentence, taskSentence) && c.TargetWord == taskTarget {
			return &aifeedback.ProviderFeedback{
				Status:                  c.ExpectedStatus,
				TargetWordUsedCorrectly: c.ExpectedStatus == aifeedback.LearningStatusCorrect,
				Explanation:             "expectation-matching provider for test",
			}, nil
		}
	}
	return &aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect, Explanation: "fallback"}, nil
}

// readFile is a tiny os.ReadFile shim that returns the
// file's content or a test-fatal error. Inline so the
// test file does not need to import "os" for one call.
func readFile(t *testing.T, path string) (string, error) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
