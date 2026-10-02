package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/stretchr/testify/require"
)

func withFakeOpenAIProvider(t *testing.T, build func(aifeedback.OpenAIConfig) aifeedback.FeedbackProvider) {
	t.Helper()
	previous := newOpenAIProvider
	newOpenAIProvider = build
	t.Cleanup(func() { newOpenAIProvider = previous })
}

func clearOpenAIEvaluationEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"AI_PROVIDER", "AI_PROVIDER_BASE_URL", "AI_PROVIDER_ACCOUNT_ID", "AI_PROVIDER_API_KEY", "AI_PROVIDER_MODEL", "AI_PROVIDER_TIMEOUT", "EVAL_LIVE_REQUEST_INTERVAL", "EVAL_LIVE_COST_USD", "EVAL_LIVE_COST_CEILING_USD", "EVAL_LIVE_OUTPUT", "EVAL_LIVE_OPERATOR_NOTES", "OPENAI_API_KEY"} {
		t.Setenv(name, "")
	}
}

func TestRunEvalLive_OpenAIConfigAndDefaults(t *testing.T) {
	clearOpenAIEvaluationEnvironment(t)
	t.Setenv("OPENAI_API_KEY", "unused-synthetic-openai-key")
	fake := &fakeFeedbackProvider{}
	other := func() aifeedback.FeedbackProvider { t.Fatal("selected another provider"); return fake }
	withFakeProviderPerCall(t, func(aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider { return other() })
	withFakeGeminiProviderPerCall(t, func(aifeedback.GeminiConfig) aifeedback.FeedbackProvider { return other() })
	withFakeCloudflareProviderPerCall(t, func(aifeedback.CloudflareConfig) aifeedback.FeedbackProvider { return other() })
	for _, tc := range []struct {
		name string
		env  map[string]string
		args []string
		want aifeedback.OpenAIConfig
	}{
		{"flag selection and defaults", nil, []string{"--provider", "openai", "--api-key", "synthetic-key"}, aifeedback.OpenAIConfig{APIKey: "synthetic-key", Model: "gpt-5-nano", Timeout: 8 * time.Second, MaxRetries: 1}},
		{"environment selection and configuration", map[string]string{"AI_PROVIDER": "openai", "AI_PROVIDER_API_KEY": "environment-key", "AI_PROVIDER_BASE_URL": "https://api.example.invalid/v1", "AI_PROVIDER_MODEL": "synthetic-model", "AI_PROVIDER_TIMEOUT": "12s"}, nil, aifeedback.OpenAIConfig{BaseURL: "https://api.example.invalid/v1", APIKey: "environment-key", Model: "synthetic-model", Timeout: 12 * time.Second, MaxRetries: 1}},
		{"explicit flags override environment", map[string]string{"AI_PROVIDER": "gemini", "AI_PROVIDER_API_KEY": "environment-key", "AI_PROVIDER_BASE_URL": "https://ignored.example.invalid/v1", "AI_PROVIDER_MODEL": "ignored-model"}, []string{"--provider", "openai", "--api-key", "flag-key", "--base-url", "", "--model", "", "--timeout", "3s"}, aifeedback.OpenAIConfig{APIKey: "flag-key", Model: "gpt-5-nano", Timeout: 3 * time.Second, MaxRetries: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for name, value := range tc.env {
				t.Setenv(name, value)
			}
			constructed := 0
			withFakeOpenAIProvider(t, func(cfg aifeedback.OpenAIConfig) aifeedback.FeedbackProvider {
				constructed++
				require.Equal(t, tc.want, cfg)
				return fake
			})
			withFakeRunLiveEvaluation(t, func(_ context.Context, provider aifeedback.FeedbackProvider, opts aifeedback.LiveEvaluationOptions) aifeedback.LiveEvaluationReport {
				require.Same(t, fake, provider)
				require.Equal(t, tc.want.Timeout, opts.Timeout)
				require.Equal(t, 1, opts.MaxRetries)
				return aifeedback.LiveEvaluationReport{AcceptanceState: "INCOMPLETE"}
			})
			var stdout, stderr bytes.Buffer
			require.Equal(t, exitIncomplete, runEvalLive(tc.args, &stdout, &stderr, time.Now))
			require.Equal(t, 1, constructed)
			require.NotContains(t, stdout.String()+stderr.String(), tc.want.APIKey)
			require.NotContains(t, stdout.String()+stderr.String(), "unused-synthetic-openai-key")
		})
	}
}

func TestRunEvalLive_OpenAIValidationBeforeConstruction(t *testing.T) {
	clearOpenAIEvaluationEnvironment(t)
	t.Setenv("OPENAI_API_KEY", "synthetic-key-not-implicitly-used")
	withFakeOpenAIProvider(t, func(aifeedback.OpenAIConfig) aifeedback.FeedbackProvider {
		t.Fatal("invalid options constructed provider")
		return nil
	})
	withFakeRunLiveEvaluation(t, func(context.Context, aifeedback.FeedbackProvider, aifeedback.LiveEvaluationOptions) aifeedback.LiveEvaluationReport {
		t.Fatal("invalid options reached evaluation")
		return aifeedback.LiveEvaluationReport{}
	})
	for _, tc := range []struct {
		name    string
		args    []string
		message string
	}{
		{"OPENAI_API_KEY alone is not consumed", nil, "AI_PROVIDER_API_KEY"},
		{"explicit empty key overrides environment", []string{"--api-key", ""}, "AI_PROVIDER_API_KEY"},
		{"nonpositive timeout", []string{"--api-key", "synthetic-key", "--timeout", "0s"}, "--timeout"},
		{"unknown capitalization", []string{"--provider", "OpenAI", "--api-key", "synthetic-key"}, "--provider"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if strings.HasPrefix(tc.name, "explicit") {
				t.Setenv("AI_PROVIDER_API_KEY", "synthetic-env-key")
			}
			var stdout, stderr bytes.Buffer
			args := append([]string{"--provider", "openai"}, tc.args...)
			require.Equal(t, exitUsageError, runEvalLive(args, &stdout, &stderr, time.Now))
			require.Contains(t, stderr.String(), tc.message)
			require.NotContains(t, stdout.String()+stderr.String(), "synthetic-key-not-implicitly-used")
			require.NotContains(t, stdout.String()+stderr.String(), "synthetic-env-key")
		})
	}
}

func TestRunEvalLive_OpenAIHelpAndExclusivePrivateEvidence(t *testing.T) {
	clearOpenAIEvaluationEnvironment(t)
	t.Setenv("AI_PROVIDER_API_KEY", "synthetic-private-key")
	t.Setenv("OPENAI_API_KEY", "synthetic-other-private-key")
	var stdout, stderr bytes.Buffer
	require.Equal(t, exitSuccess, runEvalLive([]string{"--help"}, &stdout, &stderr, time.Now))
	require.Contains(t, stderr.String(), "openai")
	require.Contains(t, stderr.String(), "gpt-5-nano")
	require.Contains(t, stderr.String(), `export AI_PROVIDER_API_KEY="$OPENAI_API_KEY"`)
	require.NotContains(t, stderr.String(), "synthetic-private-key")
	require.NotContains(t, stderr.String(), "synthetic-other-private-key")
	constructed := 0
	withFakeOpenAIProvider(t, func(aifeedback.OpenAIConfig) aifeedback.FeedbackProvider {
		constructed++
		return &fakeFeedbackProvider{}
	})
	withIncompleteEvaluation(t)
	path := filepath.Join(t.TempDir(), "evidence.json")
	args := []string{"--provider", "openai", "--format", "json", "--output", path}
	stdout.Reset()
	stderr.Reset()
	require.Equal(t, exitIncomplete, runEvalLive(args, &stdout, &stderr, time.Now))
	require.Equal(t, 1, constructed)
	original, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, stdout.Bytes(), original)
	require.NotContains(t, string(original), "synthetic-private-key")
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Zero(t, info.Mode().Perm()&0077)
	stdout.Reset()
	stderr.Reset()
	require.Equal(t, exitUsageError, runEvalLive(args, &stdout, &stderr, time.Now))
	require.Equal(t, 1, constructed, "existing output must be rejected before construction")
	retained, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, original, retained)
}
