package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/stretchr/testify/require"
)

func TestProductionOpenAIExplicitConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "https://staging.vocanova.site")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")
	t.Setenv("AI_PROVIDER", "openai")
	t.Setenv("AI_PROVIDER_API_KEY", "synthetic-key")
	t.Setenv("AI_PROVIDER_BASE_URL", "")
	t.Setenv("AI_PROVIDER_MODEL", "")
	cfg, err := LoadProductionConfig()
	require.NoError(t, err)
	require.Equal(t, "openai", cfg.APIProvider)
	require.Empty(t, cfg.APIBaseURL)
	require.Equal(t, "gpt-5-nano", cfg.APIModel)
	cfg.AIEnabled = true
	require.True(t, productionAIGenerationEnabled(cfg))
	feedback, safety := buildAIProviders(cfg)
	require.IsType(t, &aifeedback.OpenAIFeedbackProvider{}, feedback)
	require.IsType(t, &aifeedback.OpenAIModerationProvider{}, safety.(*aifeedback.CompositeSafetyClassifier).Provider())
	cfg.APIKey = ""
	require.False(t, productionAIGenerationEnabled(cfg))
	cfg.APIKey = "synthetic-key"
	cfg.AIEnabled = false
	require.False(t, productionAIGenerationEnabled(cfg))
	t.Setenv("AI_PROVIDER_MODEL", "gpt-4.1-nano")
	t.Setenv("AI_PROVIDER_BASE_URL", "https://example.invalid/v1")
	cfg, err = LoadProductionConfig()
	require.NoError(t, err)
	require.Equal(t, "gpt-4.1-nano", cfg.APIModel)
	require.Equal(t, "https://example.invalid/v1", cfg.APIBaseURL)
	t.Setenv("AI_PROVIDER_BASE_URL", "")
	t.Setenv("AI_PROVIDER_MODEL", "")
	require.Empty(t, aiProviderBaseURL(providerCloudflare))
	require.Equal(t, "@cf/meta/llama-3.3-70b-instruct-fp8-fast", aiProviderModel(providerCloudflare))
	require.Equal(t, defaultOpenCodeBaseURL, aiProviderBaseURL("opencode"))
}

func TestProductionOpenAIRuntimeZeroRetries(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(503)
		_, _ = io.WriteString(w, "private-response")
	}))
	t.Cleanup(server.Close)
	feedback, safety := buildAIProviders(ProductionConfig{APIProvider: providerOpenAI, APIKey: "private-key", APIBaseURL: server.URL, APIModel: "gpt-5-nano"})
	_, err := feedback.GenerateFeedback(t.Context(), aifeedback.NewDefaultTaskBuilder().Build(&aifeedback.Target{NormalizedWord: "work", PartOfSpeech: "verb", ShortDefinition: "do a job", LearnerLevel: "A2"}, "I work here."))
	require.ErrorIs(t, err, aifeedback.ErrProviderTimeout)
	require.EqualValues(t, 1, calls.Load())
	result, err := safety.Classify(t.Context(), aifeedback.ModerationInput{SentenceText: "I work here."})
	require.NoError(t, err)
	require.Equal(t, aifeedback.SafetyModerationUnavailable, result.Outcome)
	require.EqualValues(t, 2, calls.Load())
	// Deterministic local safety still prevents provider calls.
	result, err = safety.Classify(t.Context(), aifeedback.ModerationInput{SentenceText: "I want to kill myself."})
	require.NoError(t, err)
	require.Equal(t, aifeedback.SafetySelfHarmIntervention, result.Outcome)
	require.EqualValues(t, 2, calls.Load())
}

func TestProductionOpenAIModerationPrivateLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = io.WriteString(w, `{"error":{"message":"private-key private-learner private-provider-body"}}`)
	}))
	t.Cleanup(server.Close)
	_, safety := buildAIProviders(ProductionConfig{APIProvider: providerOpenAI, APIKey: "private-key", APIBaseURL: server.URL, APIModel: "gpt-5-nano"})
	output, err := os.CreateTemp(t.TempDir(), "diagnostics")
	require.NoError(t, err)
	original := os.Stderr
	os.Stderr = output
	t.Cleanup(func() { os.Stderr = original; _ = output.Close() })
	result, err := safety.Classify(t.Context(), aifeedback.ModerationInput{SentenceText: "private-learner works here."})
	os.Stderr = original
	require.NoError(t, err)
	require.Equal(t, aifeedback.SafetyModerationUnavailable, result.Outcome)
	_, err = output.Seek(0, 0)
	require.NoError(t, err)
	data, err := io.ReadAll(output)
	require.NoError(t, err)
	require.Equal(t, "api: ai provider=openai stage=moderation failure=request_rejected http_status=400\n", string(data))
	require.NotContains(t, result.Reason, "private-")
	var fixed bytes.Buffer
	writeOpenAIFailure(&fixed, aifeedback.OpenAIFailure{Stage: 255, Category: 255, HTTPStatus: 999})
	require.Equal(t, "api: ai provider=openai stage=unknown failure=unknown http_status=0\n", fixed.String())
}

func TestProductionOpenAIModelOverrideReachesBothAdapters(t *testing.T) {
	var models []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		models = append(models, body["model"].(string))
		w.WriteHeader(400)
	}))
	t.Cleanup(server.Close)
	f, s := buildAIProviders(ProductionConfig{APIProvider: providerOpenAI, APIKey: "synthetic-key", APIBaseURL: server.URL, APIModel: "gpt-4.1-nano"})
	_, _ = f.GenerateFeedback(t.Context(), aifeedback.NewDefaultTaskBuilder().Build(&aifeedback.Target{NormalizedWord: "work"}, "I work here."))
	_, _ = s.Classify(t.Context(), aifeedback.ModerationInput{SentenceText: "I work here."})
	require.Equal(t, []string{"gpt-4.1-nano", "gpt-4.1-nano"}, models)
}
