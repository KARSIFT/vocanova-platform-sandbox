package aifeedback

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func openAITestTask() ProviderTask {
	return NewDefaultTaskBuilder().Build(&Target{NormalizedWord: "invite", ShortDefinition: "ask someone to join an event", PartOfSpeech: "verb", LearnerLevel: "A2"}, "I invite my friends to dinner.")
}

func openAITestFeedback() map[string]any {
	return map[string]any{
		"status": "correct", "target_word_used_correctly": true,
		"grammar_acceptable": true, "meaning_clear": true, "naturalness": "natural",
		"headline": "Good invitation!", "explanation": "You use invite clearly.",
		"corrected_sentence": nil, "improvement_tip": nil,
	}
}

func openAITestEnvelope(content string) []byte {
	data, _ := json.Marshal(map[string]any{
		"status": "completed", "error": nil,
		"output": []any{map[string]any{"type": "message", "role": "assistant", "status": "completed",
			"content": []any{map[string]any{"type": "output_text", "text": content}}}},
		"usage": map[string]any{"input_tokens": 123, "output_tokens": 42},
	})
	return data
}

func openAITestResponse() []byte {
	data, _ := json.Marshal(openAITestFeedback())
	return openAITestEnvelope(string(data))
}

func TestOpenAIRequestAndFeedbackContract(t *testing.T) {
	task := openAITestTask()
	before, _ := json.Marshal(task)
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer synthetic-key" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("incorrect request endpoint, method or headers")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Error(err)
		}
		w.Write(openAITestResponse())
	}))
	defer server.Close()
	p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL + "/v1", APIKey: "synthetic-key"})
	feedback, err := p.GenerateFeedback(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewDefaultOutputValidator().Validate(feedback, nil); err != nil {
		t.Fatal(err)
	}
	if feedback.CorrectedSentence != nil || feedback.ImprovementTip != nil || len(feedback.RawJSON) != 9 {
		t.Fatal("nullable fields or raw feedback contract changed")
	}
	if received["model"] != DefaultOpenAIModel || received["store"] != false || received["max_output_tokens"] != float64(300) || len(received["tools"].([]any)) != 0 {
		t.Fatalf("unexpected model, persistence, token or tool configuration: %v", received)
	}
	if received["reasoning"].(map[string]any)["effort"] != "minimal" {
		t.Fatal("default GPT-5 nano requires minimal reasoning")
	}
	for _, field := range []string{"temperature", "previous_response_id", "conversation"} {
		if _, exists := received[field]; exists {
			t.Fatalf("unexpected %s", field)
		}
	}
	input := received["input"].([]any)
	for i, role := range []string{"system", "developer", "user"} {
		if input[i].(map[string]any)["role"] != role {
			t.Fatalf("role %d changed", i)
		}
	}
	var user map[string]any
	if err := json.Unmarshal([]byte(input[2].(map[string]any)["content"].(string)), &user); err != nil {
		t.Fatal(err)
	}
	if user["target_meaning"] != task.UserPayload["target_meaning"] || user["learner_sentence"] != task.UserPayload["learner_sentence"] {
		t.Fatal("lost target meaning or sentence")
	}
	format := received["text"].(map[string]any)["format"].(map[string]any)
	if format["type"] != "json_schema" || format["strict"] != true {
		t.Fatal("missing strict schema")
	}
	schema := format["schema"].(map[string]any)
	if schema["additionalProperties"] != false || len(schema["required"].([]any)) != 9 {
		t.Fatal("schema is not closed/all-required")
	}
	properties := schema["properties"].(map[string]any)
	for _, name := range []string{"corrected_sentence", "improvement_tip"} {
		if !reflect.DeepEqual(properties[name].(map[string]any)["type"], []any{"string", "null"}) {
			t.Fatalf("%s not nullable", name)
		}
	}
	after, _ := json.Marshal(task)
	if string(before) != string(after) {
		t.Fatal("adapter mutated task")
	}
	if provider, model := p.EvaluationIdentity(); provider != "openai" || model != DefaultOpenAIModel {
		t.Fatal("incorrect evaluation identity")
	}
}

func TestOpenAIResponseFailuresAndSemanticEvidence(t *testing.T) {
	cases := []struct {
		name string
		body []byte
		want error
	}{
		{"malformed", []byte(`not json synthetic-secret`), ErrProviderInvalidResponse},
		{"incomplete", []byte(`{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[]}`), ErrProviderInvalidResponse},
		{"failed", []byte(`{"status":"failed","error":{"message":"synthetic-secret"}}`), ErrProviderInvalidResponse},
		{"empty", []byte(`{"status":"completed","output":[]}`), ErrProviderInvalidResponse},
		{"tool", []byte(`{"status":"completed","output":[{"type":"function_call"}]}`), ErrProviderInvalidResponse},
		{"refusal", []byte(`{"status":"completed","output":[{"type":"message","role":"assistant","status":"completed","content":[{"type":"refusal","refusal":"synthetic-secret"}]}]}`), ErrProviderRefusal},
		{"fenced", openAITestEnvelope("```json\n{}\n```"), ErrProviderInvalidResponse},
		{"trailing", openAITestEnvelope(`{} {}`), ErrProviderInvalidResponse},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseOpenAIResponse(tc.body)
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "synthetic-secret") {
				t.Fatalf("want sanitized %v, got %v", tc.want, err)
			}
		})
	}
	for _, mutation := range []string{"missing_bool", "null_bool", "extra", "bad_enum", "wrong_optional", "long", "missing_optional"} {
		t.Run(mutation, func(t *testing.T) {
			raw := openAITestFeedback()
			switch mutation {
			case "missing_bool":
				delete(raw, "meaning_clear")
			case "null_bool":
				raw["meaning_clear"] = nil
			case "extra":
				raw["private"] = "secret"
			case "bad_enum":
				raw["status"] = "perfect"
			case "wrong_optional":
				raw["corrected_sentence"] = 4
			case "long":
				raw["headline"] = strings.Repeat("é", 61)
			case "missing_optional":
				delete(raw, "improvement_tip")
			}
			data, _ := json.Marshal(raw)
			if _, err := parseOpenAIResponse(openAITestEnvelope(string(data))); !errors.Is(err, ErrProviderInvalidResponse) {
				t.Fatalf("got %v", err)
			}
		})
	}
	raw := openAITestFeedback()
	raw["explanation"] = "I cannot find any grammar mistakes."
	raw["grammar_acceptable"] = false
	raw["corrected_sentence"] = ""
	data, _ := json.Marshal(raw)
	feedback, err := parseOpenAIResponse(openAITestEnvelope(string(data)))
	if err != nil {
		t.Fatal("ordinary feedback was treated as refusal", err)
	}
	if feedback.CorrectedSentence == nil || *feedback.CorrectedSentence != "" {
		t.Fatal("empty correction evidence was erased")
	}
	if NewDefaultOutputValidator().Validate(feedback, nil) == nil {
		t.Fatal("semantic contradiction must remain available for evaluator validation")
	}
}

func TestOpenAIRetryBoundAndSanitizedHTTPErrors(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		status, retries, calls int
		recover                bool
		want                   error
	}{
		{"auth", 401, 1, 1, false, ErrProviderAuth},
		{"forbidden", 403, 1, 1, false, ErrProviderAuth},
		{"bad_input", 400, 1, 1, false, ErrProviderInvalidInput},
		{"retry_429", 429, 1, 2, true, nil},
		{"retry_503", 503, 1, 2, true, nil},
		{"clamped", 503, 99, 2, false, ErrProviderTimeout},
		{"no_retry", 503, 0, 1, false, ErrProviderTimeout},
		{"negative", 503, -4, 1, false, ErrProviderTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) > 1 && tc.recover {
					w.Write(openAITestResponse())
					return
				}
				w.WriteHeader(tc.status)
				io.WriteString(w, `{"error":{"message":"synthetic-secret-and-learner-text"}}`)
			}))
			defer server.Close()
			p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", MaxRetries: tc.retries})
			_, err := p.GenerateFeedback(context.Background(), openAITestTask())
			if !errors.Is(err, tc.want) || int(calls.Load()) != tc.calls {
				t.Fatalf("err=%v calls=%d", err, calls.Load())
			}
			if err != nil && strings.Contains(err.Error(), "synthetic") {
				t.Fatal("raw provider data escaped")
			}
		})
	}
}

func TestOpenAIRequestLimitsAndGuards(t *testing.T) {
	for _, limit := range []int{-1, 0, 64, 99999} {
		task := openAITestTask()
		task.MaxOutputTokens = limit
		data, err := buildOpenAIRequest(task, "gpt-6-luna")
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		json.Unmarshal(data, &raw)
		want := 300
		if limit == 64 {
			want = 64
		}
		if raw["max_output_tokens"] != float64(want) || raw["model"] != "gpt-6-luna" {
			t.Fatal("output bound/model not respected")
		}
	}
	for _, field := range []string{"tools", "web", "memory", "schema", "payload"} {
		task := openAITestTask()
		switch field {
		case "tools":
			task.EnableTools = true
		case "web":
			task.EnableWebSearch = true
		case "memory":
			task.EnableMemory = true
		case "schema":
			task.OutputSchema = nil
		case "payload":
			task.UserPayload["invalid"] = make(chan int)
		}
		if _, err := buildOpenAIRequest(task, DefaultOpenAIModel); !errors.Is(err, ErrProviderInvalidInput) {
			t.Fatalf("%s accepted", field)
		}
	}
	for _, base := range []string{"http://example.com/v1", "https://user:secret@example.com", "https://example.com?key=secret", "https://example.com#secret", "://"} {
		p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: base, APIKey: "synthetic"})
		if _, err := p.GenerateFeedback(context.Background(), openAITestTask()); !errors.Is(err, ErrProviderInvalidInput) {
			t.Fatalf("unsafe base accepted: %v", err)
		}
	}
	p := NewOpenAIFeedbackProvider(OpenAIConfig{})
	if p.config.Timeout != 8*time.Second || p.config.BaseURL != "https://api.openai.com/v1" {
		t.Fatal("incorrect defaults")
	}
	if _, err := p.GenerateFeedback(context.Background(), openAITestTask()); !errors.Is(err, ErrProviderAuth) {
		t.Fatal("missing key not rejected")
	}
}

func TestOpenAIModelProfilesPreserveCallerCeilings(t *testing.T) {
	profiles := []struct {
		model  string
		effort string
		limit  int
	}{
		{"gpt-5-nano", "minimal", 1024},
		{"gpt-5-nano-2025-08-07", "minimal", 1024},
		{"gpt-6-luna", "none", 300},
		{"gpt-6-luna-2026-09-01", "none", 300},
		{"gpt-4.1-nano", "", 300},
		{"gpt-4.1-nano-2025-04-14", "", 300},
		{"gpt-4o-mini", "", 300},
		{"gpt-4o-mini-2024-07-18", "", 300},
	}
	for _, profile := range profiles {
		t.Run(profile.model, func(t *testing.T) {
			for _, ceiling := range []int{-1, 0, 64, 300, 512, 1024, 99999} {
				task := openAITestTask()
				task.MaxOutputTokens = ceiling
				body, err := buildOpenAIRequest(task, profile.model)
				if err != nil {
					t.Fatal(err)
				}
				var request map[string]any
				if err := json.Unmarshal(body, &request); err != nil {
					t.Fatal(err)
				}
				want := ceiling
				if want <= 0 {
					want = 300
				}
				if want > profile.limit {
					want = profile.limit
				}
				if request["max_output_tokens"] != float64(want) {
					t.Fatalf("caller ceiling %d: got %v, want %d", ceiling, request["max_output_tokens"], want)
				}
				if task.MaxOutputTokens != ceiling {
					t.Fatal("caller task was changed")
				}
				reasoning, present := request["reasoning"]
				if profile.effort == "" {
					if present {
						t.Fatal("non-reasoning model received reasoning field")
					}
				} else if !present || reasoning.(map[string]any)["effort"] != profile.effort {
					t.Fatalf("incorrect effort: %v", reasoning)
				}
				if request["model"] != profile.model {
					t.Fatal("requested model changed")
				}
			}
		})
	}
	if DefaultOpenAIModel != "gpt-5-nano" {
		t.Fatal("default must be cheapest-first candidate")
	}
	if task := openAITestTask(); task.MaxOutputTokens != 300 {
		t.Fatal("canonical task budget changed")
	}
}

func TestOpenAIUnsupportedProfilesNeverSendRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write(openAITestResponse())
	}))
	defer server.Close()
	for _, model := range []string{
		"gpt-5", "gpt-5-mini", "gpt-5-mini-2025-08-07", "gpt-6-astra",
		"gpt-5-nano-custom", "gpt-5-nano-2025-99-99", "gpt-5-nano-2025-08-07-extra",
		"gpt-4.1-nano-preview", "gpt-6-luna-fast", "unreviewed-synthetic-model",
		"gpt-4o-mini-preview", "gpt-4o-mini-2024-02-30", "gpt-4o-mini-2024-07-18-extra",
	} {
		t.Run(model, func(t *testing.T) {
			p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", Model: model, MaxRetries: 1})
			feedback, err := p.GenerateFeedback(context.Background(), openAITestTask())
			if feedback != nil || !errors.Is(err, ErrProviderInvalidInput) {
				t.Fatalf("unsupported profile accepted: feedback=%v error=%v", feedback, err)
			}
			if strings.Contains(err.Error(), model) {
				t.Fatal("model value escaped into error")
			}
		})
	}
	if calls.Load() != 0 {
		t.Fatal("unsupported model caused an HTTP request")
	}
}

func TestOpenAINeverFollowsRedirectsOrAcceptsOversizedBody(t *testing.T) {
	var redirected atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected.Add(1) }))
	defer destination.Close()
	for _, redirect := range []bool{true, false} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redirect {
				http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
				return
			}
			io.WriteString(w, strings.Repeat(" ", openAIMaxResponseBytes+1))
		}))
		p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", MaxRetries: 1})
		_, err := p.GenerateFeedback(context.Background(), openAITestTask())
		server.Close()
		if !errors.Is(err, ErrProviderInvalidResponse) {
			t.Fatalf("got %v", err)
		}
	}
	if redirected.Load() != 0 {
		t.Fatal("redirect leaked request")
	}
}

func TestOpenAITimeoutAndCancellation(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", Timeout: 100 * time.Millisecond})
	if _, err := p.GenerateFeedback(context.Background(), openAITestTask()); !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p.config.MaxRetries = 1
	if _, err := p.GenerateFeedback(ctx, openAITestTask()); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("cancelled input made an HTTP attempt")
	}
}

func TestOpenAIActiveCancellationDoesNotRetry(t *testing.T) {
	started := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(started)
		}
		io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer server.Close()
	p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", Timeout: time.Second, MaxRetries: 1})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := p.GenerateFeedback(ctx, openAITestTask()); done <- err }()
	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("request did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation did not return")
	}
	if calls.Load() != 1 {
		t.Fatal("active cancellation retried")
	}
}

func TestOpenAIResponseBodyTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	p := NewOpenAIFeedbackProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "synthetic", Timeout: 100 * time.Millisecond})
	if _, err := p.GenerateFeedback(context.Background(), openAITestTask()); !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("body timeout classified as %v", err)
	}
}
