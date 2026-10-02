package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
)

func TestRunEvalLive_HelpAndParseErrorsDoNotRevealEnvironment(t *testing.T) {
	secrets := map[string]string{
		"AI_PROVIDER_API_KEY":      "synthetic-private-key",
		"AI_PROVIDER_BASE_URL":     "https://synthetic-user:synthetic-password@example.invalid",
		"AI_PROVIDER_ACCOUNT_ID":   "synthetic-private-account",
		"EVAL_LIVE_OPERATOR_NOTES": "synthetic-private-note",
	}
	for key, value := range secrets {
		t.Setenv(key, value)
	}
	for _, args := range [][]string{{"--help"}, {"--unknown-flag"}} {
		var stdout, stderr bytes.Buffer
		runEvalLive(args, &stdout, &stderr, time.Now)
		for key, value := range secrets {
			if strings.Contains(stdout.String()+stderr.String(), value) {
				t.Errorf("%s leaked for %v", key, args)
			}
		}
	}
}

func TestRunEvalLive_RejectsInvalidSettingsBeforeProviderConstruction(t *testing.T) {
	constructed := 0
	withFakeProviderPerCall(t, func(aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		constructed++
		return &fakeFeedbackProvider{}
	})
	for _, extra := range [][]string{
		{"--provider", "typo"}, {"--timeout", "0s"}, {"--timeout", "-1s"},
		{"--request-interval", "-1s"}, {"--cost", "NaN"}, {"--cost", "+Inf"},
		{"--cost", "-2"}, {"--cost-ceiling", "NaN"}, {"--cost-ceiling", "-2"},
		{"unexpected-positional"},
	} {
		var stdout, stderr bytes.Buffer
		args := append([]string{"--base-url", "http://example.invalid", "--api-key", "test-key"}, extra...)
		if code := runEvalLive(args, &stdout, &stderr, time.Now); code != exitUsageError {
			t.Errorf("%v: got code %d", extra, code)
		}
	}
	if constructed != 0 {
		t.Fatalf("constructed %d providers for invalid settings", constructed)
	}
}

func TestRunEvalLive_ExistingEvidenceIsNeverTruncated(t *testing.T) {
	path := t.TempDir() + "/evidence.txt"
	const original = "previous reviewed evidence\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	constructed := false
	withFakeProviderPerCall(t, func(aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
		constructed = true
		return &fakeFeedbackProvider{}
	})
	var stdout, stderr bytes.Buffer
	code := runEvalLive([]string{"--base-url", "http://example.invalid", "--api-key", "test-key", "--output", path}, &stdout, &stderr, time.Now)
	if code != exitUsageError {
		t.Errorf("existing evidence: got code %d", code)
	}
	if constructed {
		t.Error("provider constructed before output destination rejected")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != original {
		t.Fatalf("existing evidence changed: %q, %v", got, err)
	}
}

func TestRunEvalLive_NewEvidenceIsPrivate(t *testing.T) {
	path := t.TempDir() + "/evidence.txt"
	withFakeProvider(t, aifeedback.ProviderFeedback{Status: aifeedback.LearningStatusCorrect})
	var stdout, stderr bytes.Buffer
	runEvalLive([]string{"--base-url", "http://example.invalid", "--api-key", "test-key", "--output", path}, &stdout, &stderr, time.Now)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0077 != 0 {
		t.Fatalf("private report permissions: %o", info.Mode().Perm())
	}
}

func TestRunEvalLive_EnvironmentValidationAndExplicitOverrides(t *testing.T) {
	for _, binding := range [][2]string{{"AI_PROVIDER_TIMEOUT", "timeout"}, {"EVAL_LIVE_REQUEST_INTERVAL", "request-interval"}, {"EVAL_LIVE_COST_USD", "cost"}, {"EVAL_LIVE_COST_CEILING_USD", "cost-ceiling"}} {
		t.Run(binding[0], func(t *testing.T) {
			t.Setenv(binding[0], "synthetic-invalid-value")
			constructed := false
			withFakeProviderPerCall(t, func(aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
				constructed = true
				return &fakeFeedbackProvider{}
			})
			withIncompleteEvaluation(t)
			args := []string{"--base-url", "http://example.invalid", "--api-key", "test-key"}
			var stdout, stderr bytes.Buffer
			if code := runEvalLive(args, &stdout, &stderr, time.Now); code != exitUsageError || constructed {
				t.Fatalf("invalid env reached provider: code=%d constructed=%v", code, constructed)
			}
			if strings.Contains(stderr.String(), "synthetic-invalid-value") {
				t.Fatal("invalid environment value leaked")
			}
			override := "0"
			if binding[1] == "timeout" || binding[1] == "request-interval" {
				override = "1ms"
			}
			if code := runEvalLive(append(args, "--"+binding[1], override), &stdout, &stderr, time.Now); code != exitIncomplete || !constructed {
				t.Fatalf("explicit override ignored: code=%d", code)
			}
		})
	}
}

func TestRunEvalLive_JSONPreservesEvidenceAndCommit(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{})
	withFakeRunLiveEvaluation(t, func(_ context.Context, _ aifeedback.FeedbackProvider, opts aifeedback.LiveEvaluationOptions) aifeedback.LiveEvaluationReport {
		if opts.Timeout != 8*time.Second || opts.RequestInterval != 0 || opts.MaxRetries != 1 {
			t.Fatalf("configured call metadata lost: timeout=%v interval=%v retries=%d", opts.Timeout, opts.RequestInterval, opts.MaxRetries)
		}
		return aifeedback.LiveEvaluationReport{AcceptanceState: "INCOMPLETE", Commit: opts.Commit, AcceptanceGaps: []string{"human review missing"}, CaseEvidence: []aifeedback.EvaluationCaseEvidence{{Case: aifeedback.EvaluationCase{ID: "synthetic-case-1"}, Outcome: "provider_error"}}}
	})
	var stdout, stderr bytes.Buffer
	code := runEvalLive([]string{"--base-url", "http://example.invalid", "--api-key", "test-key", "--format", "json", "--commit", "synthetic-revision"}, &stdout, &stderr, time.Now)
	if code != exitIncomplete {
		t.Fatalf("incomplete JSON report exit: %d", code)
	}
	var report aifeedback.LiveEvaluationReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Commit != "synthetic-revision" || len(report.CaseEvidence) != 1 || report.CaseEvidence[0].Case.ID != "synthetic-case-1" || len(report.AcceptanceGaps) != 1 {
		t.Fatalf("JSON evidence lost: %+v", report)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("synthetic write failure") }

func TestRunEvalLive_PreservesPrivateEvidenceWhenStdoutFails(t *testing.T) {
	withFakeProvider(t, aifeedback.ProviderFeedback{})
	withIncompleteEvaluation(t)
	path := t.TempDir() + "/evidence.json"
	var stderr bytes.Buffer
	code := runEvalLive([]string{"--base-url", "http://example.invalid", "--api-key", "test-key", "--format", "json", "--output", path}, failingWriter{}, &stderr, time.Now)
	if code != exitUsageError {
		t.Fatalf("stdout failure exit: %d", code)
	}
	body, err := os.ReadFile(path)
	if err != nil || !json.Valid(body) {
		t.Fatalf("private evidence lost after stdout failed: %v", err)
	}
}

func TestPacedFeedbackProvider_PreservesProviderIdentity(t *testing.T) {
	provider := &pacedFeedbackProvider{wrapped: aifeedback.NewGeminiFeedbackProvider(aifeedback.GeminiConfig{Model: "synthetic-model"})}
	name, model := aifeedback.ProviderEvaluationIdentity(provider)
	if name != "gemini" || model != "synthetic-model" {
		t.Fatalf("identity changed through pacing: %q, %q", name, model)
	}
}
