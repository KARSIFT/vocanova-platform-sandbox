package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudflareModerationFailureLogVocabulary(t *testing.T) {
	for _, tc := range []struct {
		category aifeedback.CloudflareFailureCategory
		name     string
	}{
		{aifeedback.CloudflareFailureAuthentication, "authentication"},
		{aifeedback.CloudflareFailureRequestRejected, "request_rejected"},
		{aifeedback.CloudflareFailureRateLimited, "rate_limited"},
		{aifeedback.CloudflareFailureUpstream, "upstream_failure"},
		{aifeedback.CloudflareFailureTimeout, "timeout"},
		{aifeedback.CloudflareFailureCancelled, "cancelled"},
		{aifeedback.CloudflareFailureTransport, "transport_failure"},
		{aifeedback.CloudflareFailureInvalidResponse, "invalid_response"},
		{aifeedback.CloudflareFailureRefusal, "refusal"},
		{aifeedback.CloudflareFailureCategory(255), "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			writeCloudflareModerationFailure(&output, aifeedback.CloudflareModerationFailure{Category: tc.category, HTTPStatus: 400, ProviderCode: 7003})
			assert.Equal(t, "api: ai provider=cloudflare stage=moderation category="+tc.name+" http_status=400 provider_code=7003\n", output.String())
		})
	}
}

func TestProductionCloudflareModerationLogsOnlySanitizedFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"errors":[{"code":7003,"message":"private-key private-account private-model private-learner private-response"}]}`)
	}))
	t.Cleanup(server.Close)
	_, safety := buildAIProviders(ProductionConfig{APIProvider: providerCloudflare, APIKey: "private-key", APIAccountID: "private-account", APIModel: "private-model", APIBaseURL: server.URL})
	// This test is deliberately nonparallel while capturing the existing stderr
	// sink; cleanup restores it even if a later assertion fails.
	output, err := os.CreateTemp(t.TempDir(), "diagnostic-log")
	require.NoError(t, err)
	original := os.Stderr
	os.Stderr = output
	t.Cleanup(func() { os.Stderr = original; _ = output.Close() })
	result, err := safety.Classify(t.Context(), aifeedback.ModerationInput{SentenceText: "private-learner works every day.", TargetWord: "work"})
	os.Stderr = original
	require.NoError(t, err)
	assert.Equal(t, aifeedback.SafetyModerationUnavailable, result.Outcome)
	_, err = output.Seek(0, 0)
	require.NoError(t, err)
	logged, err := io.ReadAll(output)
	require.NoError(t, err)
	assert.Equal(t, "api: ai provider=cloudflare stage=moderation category=request_rejected http_status=400 provider_code=7003\n", string(logged))
	assert.NotContains(t, string(logged), "private-")
	assert.NotContains(t, result.Reason, "private-")
}
