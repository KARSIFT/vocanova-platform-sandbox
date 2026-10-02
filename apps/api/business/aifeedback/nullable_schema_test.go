package aifeedback

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeedbackSchemaNullableStrings(t *testing.T) {
	schema := outputSchema()
	properties := schema["properties"].(map[string]any)
	for name, limit := range map[string]int{"corrected_sentence": 300, "improvement_tip": 160} {
		t.Run(name, func(t *testing.T) {
			field := properties[name].(map[string]any)
			require.Equal(t, []string{"string", "null"}, field["type"])
			require.Equal(t, limit, field["maxLength"])
			require.NotContains(t, schema["required"], name)
		})
	}
	require.Equal(t, []string{"status", "target_word_used_correctly", "grammar_acceptable", "meaning_clear", "naturalness", "headline", "explanation"}, schema["required"])
}

// Inspect the actual serialized request schema, not a second implementation of
// JSON Schema validation. These builders are also used by the live adapters.
func TestFeedbackWireSchemasPreserveNullableStrings(t *testing.T) {
	original := newTestTask()
	repair := NewDefaultTaskBuilder().BuildRepair(original, "invalid feedback", nil)
	originalBefore := nullableSchemaJSON(t, original)
	providers := []string{"cloudflare", "opencode", "gemini", "openai"}
	for _, taskCase := range []struct {
		name string
		task ProviderTask
	}{{"initial", original}, {"repair", repair}} {
		for _, provider := range providers {
			t.Run(taskCase.name+"/"+provider, func(t *testing.T) {
				before := nullableSchemaJSON(t, taskCase.task)
				got := feedbackRequestSchema(t, provider, taskCase.task)
				require.Equal(t, before, nullableSchemaJSON(t, taskCase.task), "building a request must not mutate the task")
				require.Equal(t, originalBefore, nullableSchemaJSON(t, original), "repair and original share a schema")

				// Start with a JSON copy so numeric values match the decoded wire shape.
				var want map[string]any
				require.NoError(t, json.Unmarshal(nullableSchemaJSON(t, outputSchema()), &want))
				properties := want["properties"].(map[string]any)
				for _, name := range []string{"corrected_sentence", "improvement_tip"} {
					field := properties[name].(map[string]any)
					field["type"] = []any{"string", "null"}
					if provider == "gemini" {
						field["type"] = "string"
						field["nullable"] = true
					}
				}
				if provider == "openai" {
					want["additionalProperties"] = false
					want["required"] = []any{"corrected_sentence", "explanation", "grammar_acceptable", "headline", "improvement_tip", "meaning_clear", "naturalness", "status", "target_word_used_correctly"}
				}
				require.Equal(t, want, got, "preserve all limits, enums and required fields while representing nullable strings")
			})
		}
	}
}

func TestGeminiNullableSchemaTranslationDoesNotMutateCanonicalUnion(t *testing.T) {
	task := newTestTask()
	// Supply the intended union directly so this regression independently detects
	// missing Gemini adaptation even before the canonical schema is corrected.
	properties := task.OutputSchema["properties"].(map[string]any)
	for _, name := range []string{"corrected_sentence", "improvement_tip"} {
		properties[name].(map[string]any)["type"] = []string{"string", "null"}
	}
	before := nullableSchemaJSON(t, task)
	got := feedbackRequestSchema(t, "gemini", task)
	require.Equal(t, before, nullableSchemaJSON(t, task))
	for name, limit := range map[string]float64{"corrected_sentence": 300, "improvement_tip": 160} {
		field := got["properties"].(map[string]any)[name].(map[string]any)
		require.Equal(t, map[string]any{"type": "string", "nullable": true, "maxLength": limit}, field)
	}
}

func TestGeminiNullableSchemaTranslationPreservesModeration(t *testing.T) {
	schema := moderationOutputSchema()
	before := nullableSchemaJSON(t, schema)
	body, err := buildGeminiGenerateContentBody("system", "developer", map[string]any{"sentence": "Synthetic text."}, schema)
	require.NoError(t, err)
	var request geminiGenerateContentRequest
	require.NoError(t, json.Unmarshal(body, &request))
	require.JSONEq(t, string(before), string(nullableSchemaJSON(t, request.GenerationConfig.ResponseSchema)))
	require.Equal(t, before, nullableSchemaJSON(t, schema))
}

func feedbackRequestSchema(t *testing.T, provider string, task ProviderTask) map[string]any {
	t.Helper()
	switch provider {
	case "cloudflare":
		body, err := buildCloudflareRequestBody(task.SystemPrompt, task.DeveloperPrompt, task.UserPayload, task.OutputSchema)
		require.NoError(t, err)
		var request cloudflareRunRequest
		require.NoError(t, json.Unmarshal(body, &request))
		require.Equal(t, "json_schema", request.ResponseFormat.Type)
		return request.ResponseFormat.JSONSchema
	case "opencode":
		provider := NewOpenCodeFeedbackProvider(OpenCodeConfig{BaseURL: "http://127.0.0.1:1", Model: "synthetic/schema-test"})
		body, err := provider.buildMessageRequestBody(task)
		require.NoError(t, err)
		var request openCodeMessageRequest
		require.NoError(t, json.Unmarshal(body, &request))
		require.Len(t, request.Parts, 1)
		_, schemaText, found := strings.Cut(request.Parts[0].Text, "Respond with a single JSON object matching this schema exactly, and nothing else:\n")
		require.True(t, found)
		var schema map[string]any
		require.NoError(t, json.Unmarshal([]byte(schemaText), &schema))
		return schema
	case "gemini":
		body, err := buildGeminiGenerateContentBody(task.SystemPrompt, task.DeveloperPrompt, task.UserPayload, task.OutputSchema)
		require.NoError(t, err)
		var request geminiGenerateContentRequest
		require.NoError(t, json.Unmarshal(body, &request))
		require.Equal(t, "application/json", request.GenerationConfig.ResponseMimeType)
		return request.GenerationConfig.ResponseSchema
	case "openai":
		body, err := buildOpenAIRequest(task, DefaultOpenAIModel)
		require.NoError(t, err)
		var request map[string]any
		require.NoError(t, json.Unmarshal(body, &request))
		format := request["text"].(map[string]any)["format"].(map[string]any)
		require.Equal(t, true, format["strict"])
		return format["schema"].(map[string]any)
	default:
		t.Fatalf("unsupported test provider %q", provider)
		return nil
	}
}

func nullableSchemaJSON(t *testing.T, value any) []byte {
	t.Helper()
	result, err := json.Marshal(value)
	require.NoError(t, err)
	return result
}
