package aifeedback

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAI4oMiniExplicitProfileForFeedbackAndModeration(t *testing.T) {
	for _, model := range []string{"gpt-4o-mini", "gpt-4o-mini-2024-07-18"} {
		t.Run(model, func(t *testing.T) {
			stages := []string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, model, body["model"])
				require.Equal(t, false, body["store"])
				require.Equal(t, float64(300), body["max_output_tokens"])
				require.NotContains(t, body, "reasoning", "non-reasoning profile must omit the field entirely")
				require.Empty(t, body["tools"])
				format := body["text"].(map[string]any)["format"].(map[string]any)
				require.Equal(t, true, format["strict"])
				require.Equal(t, false, format["schema"].(map[string]any)["additionalProperties"])
				stage := format["name"].(string)
				stages = append(stages, stage)
				if stage == "sentence_moderation" {
					_, _ = w.Write(openAITestEnvelope(`{"outcome":"allowed","reason":"Ordinary learning sentence."}`))
				} else {
					_, _ = w.Write(openAITestResponse())
				}
			}))
			defer server.Close()
			config := OpenAIConfig{APIKey: "synthetic", BaseURL: server.URL, Model: model, MaxRetries: 0}
			feedback, err := NewOpenAIFeedbackProvider(config).GenerateFeedback(t.Context(), openAITestTask())
			require.NoError(t, err)
			require.NotNil(t, feedback)
			moderation, err := NewOpenAIModerationProvider(config).Classify(t.Context(), ModerationInput{SentenceText: "I work in a shop.", TargetWord: "work", LearnerLevel: "A2"})
			require.NoError(t, err)
			require.Equal(t, SafetyAllowed, moderation.Outcome)
			require.Equal(t, []string{"sentence_feedback", "sentence_moderation"}, stages)
		})
	}
}
