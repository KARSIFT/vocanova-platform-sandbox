package aifeedback

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type diagnosticRoundTripper func(*http.Request) (*http.Response, error)

func (f diagnosticRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCloudflareModerationDiagnosticsFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		body     string
		category CloudflareFailureCategory
		code     int
	}{
		{"unauthorized", 401, `{"errors":[{"code":10000,"message":"private-provider-text"}]}`, CloudflareFailureAuthentication, 10000},
		{"forbidden", 403, `{}`, CloudflareFailureAuthentication, 0},
		{"invalid request", 400, `{}`, CloudflareFailureRequestRejected, 0},
		{"missing model", 404, `{}`, CloudflareFailureRequestRejected, 0},
		{"rate limit", 429, `{}`, CloudflareFailureRateLimited, 0},
		{"upstream", 503, `{}`, CloudflareFailureUpstream, 0},
		{"malformed", 200, `private-provider-text`, CloudflareFailureInvalidResponse, 0},
		{"unsuccessful envelope", 200, `{"success":false,"errors":[{"code":5007,"message":"private-provider-text"}]}`, CloudflareFailureInvalidResponse, 5007},
		{"invalid outcome", 200, `{"success":true,"result":{"response":{"outcome":"private-provider-text"}}}`, CloudflareFailureInvalidResponse, 0},
		{"refusal", 200, `{"success":true,"result":{"response":"I cannot assist with that"}}`, CloudflareFailureRefusal, 0},
		{"string code ignored", 400, `{"errors":[{"code":"private-provider-text"}]}`, CloudflareFailureRequestRejected, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []CloudflareModerationFailure
			provider := NewCloudflareModerationProvider(CloudflareConfig{APIToken: "private-key", AccountID: "private-account", Model: "private-model", MaxRetries: 9, OnModerationFailure: func(event CloudflareModerationFailure) { events = append(events, event) }})
			calls := 0
			provider.client.Transport = diagnosticRoundTripper(func(r *http.Request) (*http.Response, error) {
				calls++
				var body cloudflareRunRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "json_schema", body.ResponseFormat.Type)
				expected, _ := json.Marshal(moderationOutputSchema())
				actual, _ := json.Marshal(body.ResponseFormat.JSONSchema)
				assert.JSONEq(t, string(expected), string(actual))
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})
			f := newServiceFixture(t)
			f.service.safety = NewCompositeSafetyClassifier(NewDefaultLocalAbuseChecker(), provider)
			mission := &countedMissionAccounting{}
			f.service.mission = mission
			result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, ErrorCodeSafetyModerationUnavailable, result.ErrorCode)
			assert.True(t, result.CanRetry)
			assert.False(t, result.MissionCompleted)
			assert.Zero(t, mission.calls)
			assert.Empty(t, f.repo.attempts)
			assert.Empty(t, f.repo.sentences)
			assert.Zero(t, f.provider.calls)
			assert.Equal(t, 1, calls)
			require.Equal(t, []CloudflareModerationFailure{{Category: tc.category, HTTPStatus: tc.status, ProviderCode: tc.code}}, events)
			encoded, _ := json.Marshal(events)
			assert.NotContains(t, string(encoded), "private-")
		})
	}
}

func TestCloudflareModerationDiagnosticsTransport(t *testing.T) {
	for _, tc := range []struct {
		name     string
		err      error
		category CloudflareFailureCategory
	}{
		{"deadline", context.DeadlineExceeded, CloudflareFailureTimeout},
		{"cancelled", context.Canceled, CloudflareFailureCancelled},
		{"network", errors.New("private-key private-request private-host"), CloudflareFailureTransport},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []CloudflareModerationFailure
			provider := NewCloudflareModerationProvider(CloudflareConfig{AccountID: "private-account", OnModerationFailure: func(event CloudflareModerationFailure) { events = append(events, event) }})
			provider.client.Transport = diagnosticRoundTripper(func(*http.Request) (*http.Response, error) { return nil, tc.err })
			_, err := provider.Classify(t.Context(), ModerationInput{SentenceText: "I work every day."})
			require.Error(t, err)
			assert.Equal(t, []CloudflareModerationFailure{{Category: tc.category}}, events)
		})
	}
}

func TestCloudflareModerationSuccessDoesNotReportFailure(t *testing.T) {
	provider := NewCloudflareModerationProvider(CloudflareConfig{AccountID: "synthetic", OnModerationFailure: func(CloudflareModerationFailure) { t.Error("unexpected failure event") }})
	provider.client.Transport = diagnosticRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"success":true,"result":{"response":{"outcome":"allowed","reason":"synthetic"}}}`)), Header: make(http.Header)}, nil
	})
	result, err := provider.Classify(t.Context(), ModerationInput{SentenceText: "I work every day."})
	require.NoError(t, err)
	assert.Equal(t, SafetyAllowed, result.Outcome)
}

type diagnosticBrokenBody struct{}

func (diagnosticBrokenBody) Read([]byte) (int, error) {
	return 0, errors.New("private-response-read-error")
}
func (diagnosticBrokenBody) Close() error { return nil }

func TestCloudflareModerationDiagnosticsBodyReadFailure(t *testing.T) {
	var events []CloudflareModerationFailure
	provider := NewCloudflareModerationProvider(CloudflareConfig{AccountID: "synthetic", OnModerationFailure: func(event CloudflareModerationFailure) { events = append(events, event) }})
	provider.client.Transport = diagnosticRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: diagnosticBrokenBody{}, Header: make(http.Header)}, nil
	})
	_, err := provider.Classify(t.Context(), ModerationInput{SentenceText: "I work every day."})
	require.Error(t, err)
	assert.Equal(t, []CloudflareModerationFailure{{Category: CloudflareFailureTransport, HTTPStatus: 200}}, events)
}
