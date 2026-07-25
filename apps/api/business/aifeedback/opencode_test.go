package aifeedback

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCodeFeedbackProviderImplementsInterface(t *testing.T) {
	var _ FeedbackProvider = (*OpenCodeFeedbackProvider)(nil)
}

// newOpenCodeTestServer builds a fake `opencode serve` handling both
// POST /session (returns a session ID) and POST /session/{id}/message
// (handled by messageHandler). This mirrors opencode serve's real API shape
// (confirmed live against its own OpenAPI document, GET /doc) - not an
// OpenAI-compatible chat-completions endpoint.
func newOpenCodeTestServer(t *testing.T, messageHandler func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/session":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ses_test123"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/session/ses_test123/message":
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var req openCodeMessageRequest
			require.NoError(t, json.Unmarshal(body, &req))
			messageHandler(w, r, req)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
}

func TestOpenCodeFeedbackProviderParsesValidResponse(t *testing.T) {
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		assert.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "opencode-go", req.Model.ProviderID)
		assert.Equal(t, "deepseek-v4-pro", req.Model.ModelID)
		assert.Equal(t, "system", req.System)
		require.Len(t, req.Parts, 1)
		assert.Equal(t, "text", req.Parts[0].Type)
		assert.Contains(t, req.Parts[0].Text, "learner_sentence")

		resp := openCodeMessageResponse{
			Parts: []openCodePart{
				{Type: "text", Text: `{"status":"correct","target_word_used_correctly":true,"explanation":"Good use of work."}`},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.APIKey = "secret-key"
	config.Model = "opencode-go/deepseek-v4-pro"
	config.Timeout = 5 * time.Second

	provider := NewOpenCodeFeedbackProvider(config)
	feedback, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, feedback.Status)
	assert.True(t, feedback.TargetWordUsedCorrectly)
	assert.Equal(t, "Good use of work.", feedback.Explanation)
	assert.Nil(t, feedback.CorrectedSentence)
	assert.Nil(t, feedback.ImprovementTip)
	assert.NotNil(t, feedback.RawJSON)
}

func TestOpenCodeFeedbackProviderRetriesTransientError(t *testing.T) {
	calls := 0
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"parts":[{"type":"text","text":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"}]}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.MaxRetries = 1

	provider := NewOpenCodeFeedbackProvider(config)
	feedback, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, feedback.Status)
	assert.Equal(t, 2, calls)
}

func TestOpenCodeFeedbackProviderNoRetryOnAuthError(t *testing.T) {
	calls := 0
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.MaxRetries = 1

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderAuth)
	assert.Equal(t, 1, calls)
}

func TestOpenCodeFeedbackProviderNoRetryOnInvalidInput(t *testing.T) {
	calls := 0
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"Invalid request"}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.MaxRetries = 1

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderInvalidInput)
	assert.Equal(t, 1, calls)
}

func TestOpenCodeFeedbackProviderHandlesTextRefusal(t *testing.T) {
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"parts":[{"type":"text","text":"I'm sorry, I can't help with that."}]}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderRefusal)
}

func TestOpenCodeFeedbackProviderHandlesInvalidJSON(t *testing.T) {
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"parts":[{"type":"text","text":"not json"}]}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderInvalidResponse)
}

func TestOpenCodeFeedbackProviderExtractsJSONFromMarkdown(t *testing.T) {
	contentJSON := "```json\n" +
		`{"status":"incorrect","target_word_used_correctly":false,"explanation":"Wrong.","corrected_sentence":"I work every day.","improvement_tip":"Use the target word."}` +
		"\n```"

	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := openCodeMessageResponse{
			Parts: []openCodePart{{Type: "text", Text: contentJSON}},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL

	provider := NewOpenCodeFeedbackProvider(config)
	feedback, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
	assert.Equal(t, LearningStatusIncorrect, feedback.Status)
	assert.False(t, feedback.TargetWordUsedCorrectly)
	assert.Equal(t, "Wrong.", feedback.Explanation)
	require.NotNil(t, feedback.CorrectedSentence)
	assert.Equal(t, "I work every day.", *feedback.CorrectedSentence)
	require.NotNil(t, feedback.ImprovementTip)
	assert.Equal(t, "Use the target word.", *feedback.ImprovementTip)
}

func TestOpenCodeFeedbackProviderUsesConfiguredModel(t *testing.T) {
	var gotModel *openCodeModelRef
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		gotModel = req.Model
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"parts":[{"type":"text","text":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"}]}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.Model = "custom-provider/custom-model"

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
	require.NotNil(t, gotModel)
	assert.Equal(t, "custom-provider", gotModel.ProviderID)
	assert.Equal(t, "custom-model", gotModel.ModelID)
}

func TestOpenCodeFeedbackProviderNoAPIKeyOmitsAuthorization(t *testing.T) {
	server := newOpenCodeTestServer(t, func(w http.ResponseWriter, r *http.Request, req openCodeMessageRequest) {
		assert.Empty(t, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"parts":[{"type":"text","text":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"}]}`))
	})
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.APIKey = ""

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
}

func TestOpenCodeFeedbackProviderDefaultModel(t *testing.T) {
	config := DefaultOpenCodeConfig()
	provider := NewOpenCodeFeedbackProvider(config)
	assert.Equal(t, DefaultOpenCodeModel, config.Model)
	assert.NotEmpty(t, provider.providerID)
}

func TestOpenCodeFeedbackProviderDefaultTimeout(t *testing.T) {
	config := DefaultOpenCodeConfig()
	provider := NewOpenCodeFeedbackProvider(config)
	assert.Equal(t, 8*time.Second, provider.config.Timeout)
}

func TestOpenCodeFeedbackProviderNoHardcodedCredentials(t *testing.T) {
	config := DefaultOpenCodeConfig()
	assert.Empty(t, config.APIKey, "default config must not contain a hardcoded API key")
	provider := NewOpenCodeFeedbackProvider(config)
	assert.Empty(t, provider.config.APIKey)
}

func TestSplitOpenCodeModel(t *testing.T) {
	providerID, modelID := splitOpenCodeModel("opencode-go/deepseek-v4-pro")
	assert.Equal(t, "opencode-go", providerID)
	assert.Equal(t, "deepseek-v4-pro", modelID)
}

func newTestTask() ProviderTask {
	return ProviderTask{
		PromptVersion:   PromptVersionSentenceFeedbackV1,
		SchemaVersion:   SchemaVersionFeedbackV1,
		SystemPrompt:    "system",
		DeveloperPrompt: "developer",
		UserPayload: map[string]any{
			"learner_level":    "a2",
			"target_word":      "work",
			"part_of_speech":   "verb",
			"target_meaning":   "to do a job",
			"accepted_forms":   []string{"work", "works", "worked", "working"},
			"learner_sentence": "I work every day.",
		},
		OutputSchema:    outputSchema(),
		MaxOutputTokens: 300,
		Temperature:     0.1,
	}
}
