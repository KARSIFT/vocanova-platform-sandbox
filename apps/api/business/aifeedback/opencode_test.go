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

func TestOpenCodeFeedbackProviderParsesValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer secret-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req openCodeRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, DefaultOpenCodeModel, req.Model)
		assert.False(t, req.Stream)
		assert.InDelta(t, 0.1, req.Temperature, 0.001)
		assert.Equal(t, 300, req.MaxTokens)
		assert.Len(t, req.Messages, 3)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "developer", req.Messages[1].Role)
		assert.Equal(t, "user", req.Messages[2].Role)
		assert.Contains(t, req.Messages[2].Content, "learner_sentence")

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": `{"status":"correct","target_word_used_correctly":true,"explanation":"Good use of work."}`,
					},
					"finish_reason": "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.APIKey = "secret-key"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":{"message":"upstream unavailable"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"},"finish_reason":"stop"}]}`))
	}))
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid request"}}`))
	}))
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.MaxRetries = 1

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderInvalidInput)
	assert.Equal(t, 1, calls)
}

func TestOpenCodeFeedbackProviderHandlesContentFilterRefusal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":""},"finish_reason":"content_filter"}]}`))
	}))
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderRefusal)
}

func TestOpenCodeFeedbackProviderHandlesTextRefusal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"I'm sorry, I can't help with that."},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	assert.ErrorIs(t, err, ErrProviderRefusal)
}

func TestOpenCodeFeedbackProviderHandlesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"not json"},"finish_reason":"stop"}]}`))
	}))
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := openCodeResponse{
			Choices: []struct {
				Message      openCodeMessage `json:"message"`
				FinishReason string          `json:"finish_reason"`
			}{
				{
					Message:      openCodeMessage{Content: contentJSON},
					FinishReason: "stop",
				},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
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
	var model string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req openCodeRequest
		_ = json.Unmarshal(body, &req)
		model = req.Model
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	config := DefaultOpenCodeConfig()
	config.BaseURL = server.URL
	config.Model = "custom-model"

	provider := NewOpenCodeFeedbackProvider(config)
	_, err := provider.GenerateFeedback(t.Context(), newTestTask())
	require.NoError(t, err)
	assert.Equal(t, "custom-model", model)
}

func TestOpenCodeFeedbackProviderNoAPIKeyOmitsAuthorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"status\":\"correct\",\"target_word_used_correctly\":true,\"explanation\":\"OK.\"}"},"finish_reason":"stop"}]}`))
	}))
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
	assert.Equal(t, DefaultOpenCodeModel, provider.config.Model)
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
