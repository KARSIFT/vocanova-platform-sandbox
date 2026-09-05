package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAIFeedbackAPI(t *testing.T, data aifeedback.MemoryRepositoryData) (huma.API, *aifeedback.Service, *auth.Service) {
	authSvc := authStubService()
	svc := aifeedback.NewService(
		aifeedback.NewMemoryRepository(data),
		aifeedback.NewMockProvider(),
		aifeedback.NewCompositeSafetyClassifier(aifeedback.NewDefaultLocalAbuseChecker(), aifeedback.NewMockProvider()),
		nil,
		learning.NewMemoryIdempotencyStore(),
		nil,
		aifeedback.NewNoopTelemetryRecorder(),
		aifeedback.NewDefaultTaskBuilder(),
		aifeedback.NewDefaultOutputValidator(),
		nil,
		aifeedback.DefaultServiceConfig(),
	)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterAIFeedback(api, svc, authSvc)
	return api, svc, authSvc
}

func authenticatedAIFeedbackRequest(t *testing.T, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback", nil)
	ctx := WithRequester(req.Context(), &auth.User{ID: userID})
	return req.WithContext(ctx)
}

func submitSentenceFeedbackRequest(t *testing.T, userID uuid.UUID, body string, authSvc *auth.Service) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	return req
}

func TestSubmitSentenceFeedbackRequiresAuth(t *testing.T) {
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubmitSentenceFeedbackRequiresCSRF(t *testing.T) {
	api, _, _ := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSubmitSentenceFeedbackRequiresIdempotencyKey(t *testing.T) {
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSubmitSentenceFeedbackSuccess(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, userID, body, authSvc)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var out SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))
	assert.Equal(t, "correct", out.Body.Status)
	assert.Equal(t, "I work every day.", out.Body.OriginalSentence)
	assert.False(t, out.Body.MissionCompleted)
	assert.False(t, out.Body.CanRetry)
}

func TestSubmitSentenceFeedbackValidationTooShort(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, userID, body, authSvc)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var out SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))
	assert.Equal(t, aifeedback.ValidationCodeTooShort, out.Body.ErrorCode)
	assert.True(t, out.Body.CanRetry)
}

func TestSubmitSentenceFeedbackUnknownUserWordReturns404(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, userID, body, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubmitSentenceFeedbackCrossUserCannotSubmit(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: owner, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000005"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, other, body, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSubmitSentenceFeedbackIdempotencyReplay(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	submit := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := submitSentenceFeedbackRequest(t, userID, body, authSvc)
		api.Adapter().ServeHTTP(w, req)
		return w
	}

	first := submit()
	require.Equal(t, http.StatusOK, first.Code)
	second := submit()
	require.Equal(t, http.StatusOK, second.Code)
	var firstBody, secondBody SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstBody.Body))
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondBody.Body))
	assert.Equal(t, firstBody.Body.SentenceID, secondBody.Body.SentenceID)
	assert.Equal(t, firstBody.Body.AttemptID, secondBody.Body.AttemptID)
}

func TestSubmitSentenceFeedbackIdempotencyKeyConflict(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning",
		}},
	})

	body1 := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	body2 := `{"sentenceText":"She works hard.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`

	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, userID, body1, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = submitSentenceFeedbackRequest(t, userID, body2, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSubmitSentenceFeedbackCrossUserIdempotencyKeyIsolated(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	ownerWordID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	otherWordID := uuid.MustParse("00000000-0000-0000-0000-000000000006")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{
			{ID: ownerWordID, UserID: owner, MeaningID: meaningID, Status: "learning"},
			{ID: otherWordID, UserID: other, MeaningID: meaningID, Status: "learning"},
		},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"` + ownerWordID.String() + `"}`
	otherBody := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"` + otherWordID.String() + `"}`

	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, owner, body, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = submitSentenceFeedbackRequest(t, other, otherBody, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReportSentenceFeedbackRequiresAuth(t *testing.T) {
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"already_correct"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReportSentenceFeedbackRequiresCSRF(t *testing.T) {
	api, _, _ := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"already_correct"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestReportSentenceFeedbackUnknownAttemptReturns404(t *testing.T) {
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"already_correct"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportSentenceFeedbackCrossUserReturns404(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: owner, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000005"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, owner, body, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var out SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))
	attemptID := out.Body.AttemptID

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/sentence-feedback/%s/reports", attemptID), bytes.NewReader([]byte(`{"reason":"already_correct"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: other}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportSentenceFeedbackSucceedsForOwner(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{
		Words: []aifeedback.MemoryWord{{
			ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
		}},
		Meanings: []aifeedback.MemoryMeaning{{
			ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
		}},
		UserWords: []aifeedback.MemoryUserWord{{
			ID: userWordID, UserID: owner, MeaningID: meaningID, Status: "learning",
		}},
	})

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000005"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, owner, body, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var out SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))
	attemptID := out.Body.AttemptID

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/sentence-feedback/%s/reports", attemptID), bytes.NewReader([]byte(`{"reason":"already_correct"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: owner}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Reusing the operation key with a different report reason is a conflict.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/sentence-feedback/%s/reports", attemptID), bytes.NewReader([]byte(`{"reason":"correction_changed_meaning"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-key")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: owner}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------------------------------------------------------------------
// VOC-034-T02 regression test: prove the production-wiring defect from
// issue #216 (the literal-nil safety classifier) is fixed at the route
// level, not just at the helper level. An ordinary safe sentence submitted
// via POST /api/v1/sentence-feedback must reach BOTH the moderation
// provider AND the feedback provider through the same real-provider
// construction path production.go's buildAIProviders uses, not
// MockProvider. The fake server's per-message counters prove both
// network calls were actually made.
// ---------------------------------------------------------------------------

// openCodeE2EMessageRequest mirrors the unexported aifeedback.openCodeMessageRequest
// shape (System / Model / Parts) so this test file can decode the request
// body the real provider sends without exporting the type. Field names match
// the JSON tags aifeedback/opencode.go uses; no field is added or removed.
type openCodeE2EMessageRequest struct {
	System string            `json:"system,omitempty"`
	Model  *openCodeE2EModel `json:"model,omitempty"`
	Parts  []openCodeE2EPart `json:"parts"`
}

type openCodeE2EModel struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

type openCodeE2EPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// openCodeE2EMessageResponse mirrors the unexported aifeedback.openCodeMessageResponse
// shape (Info / Parts) so the fake server's response matches what the real
// adapter's parseMessageResponse / parseModerationResponse unmarshal into.
type openCodeE2EMessageResponse struct {
	Info  json.RawMessage   `json:"info"`
	Parts []openCodeE2EPart `json:"parts"`
}

// openCodeE2ECallCounters holds the per-message counters the fake server
// records. Both providers POST to the same /session and
// /session/{id}/message endpoints, so the test distinguishes them by
// inspecting the system prompt the request body carries (moderation
// system prompt starts with "You are a content-moderation classifier";
// feedback system prompt starts with "You are a concise, supportive
// English-learning tutor"). This is the same architecture
// aifeedback.opencode.go carries: a single set of transport endpoints,
// a system-prompt-shaped distinction per adapter.
type openCodeE2ECallCounters struct {
	sessionCalls    int32
	moderationCalls int32
	feedbackCalls   int32
}

// newOpenCodeE2ETestServer stands in for a real `opencode serve` for the
// end-to-end route-level regression. It serves POST /session (always
// returns the same session id, since the real adapter creates a new
// session per call) and POST /session/{id}/message. The message handler
// inspects req.System to decide which provider is being called: the
// moderation message returns the strict four-value outcome JSON the
// parser expects, the feedback message returns a status="correct" JSON
// that passes DefaultOutputValidator. Per-attempt call counters let the
// test assert both round trips actually happened.
func newOpenCodeE2ETestServer(t *testing.T) (*httptest.Server, *openCodeE2ECallCounters) {
	t.Helper()
	counters := &openCodeE2ECallCounters{}
	const sessionID = "ses_e2e_regression"
	const authToken = "test-e2e-opencode-key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/session":
			atomic.AddInt32(&counters.sessionCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"` + sessionID + `"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/session/"+sessionID+"/message":
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var req openCodeE2EMessageRequest
			require.NoError(t, json.Unmarshal(body, &req))
			assert.Equal(t, "Bearer "+authToken, r.Header.Get("Authorization"),
				"real provider must forward the configured API key as Bearer")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "opencode-go", req.Model.ProviderID)
			assert.Equal(t, "test-model", req.Model.ModelID)
			require.NotEmpty(t, req.Parts)
			assert.Equal(t, "text", req.Parts[0].Type)

			// Distinguish the two adapters by their system-prompt prefix.
			// This is a stable, version-controlled distinction: the
			// moderation and feedback system prompts are both defined in
			// version-controlled code (moderation.go's
			// moderationSystemPrompt() and task.go's systemPrompt()).
			switch {
			case strings.Contains(req.System, "content-moderation classifier"):
				atomic.AddInt32(&counters.moderationCalls, 1)
				resp := openCodeE2EMessageResponse{
					Info:  json.RawMessage(`{}`),
					Parts: []openCodeE2EPart{{Type: "text", Text: `{"outcome":"allowed","reason":"ordinary language"}`}},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(resp))
			case strings.Contains(req.System, "English-learning tutor"):
				atomic.AddInt32(&counters.feedbackCalls, 1)
				resp := openCodeE2EMessageResponse{
					Info:  json.RawMessage(`{}`),
					Parts: []openCodeE2EPart{{Type: "text", Text: `{"status":"correct","target_word_used_correctly":true,"explanation":"Good use of the target word."}`}},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(resp))
			default:
				t.Fatalf("unrecognized system prompt (cannot tell moderation from feedback): %q", req.System)
			}
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	return server, counters
}

// TestSubmitSentenceFeedback_RealProviders_ReachesFeedbackProviderSeam is the

// TestSubmitSentenceFeedback_RealProviders_ReachesFeedbackProviderSeam is the
// VOC-034-T02 / VOC-034-AC-06 / VOC-034-AC-09 regression: an ordinary safe
// sentence submitted via the real production route must (a) NOT return
// SAFETY_MODERATION_UNAVAILABLE, (b) actually invoke the moderation
// provider's transport, AND (c) actually invoke the feedback provider's
// transport - the failure sequence from issue #216 was "moderation never
// reached, feedback never called, fail-closed response returned". This
// test wires the real OpenCode providers the same way production.go's
// buildAIProviders does (CompositeSafetyClassifier over a real
// OpenCodeModerationProvider, not MockProvider), points them at a fake
// `opencode serve` httptest.Server, and asserts the exact failure sequence
// is no longer present.
func TestSubmitSentenceFeedback_RealProviders_ReachesFeedbackProviderSeam(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	server, counters := newOpenCodeE2ETestServer(t)

	// Build the exact same construction production.go's buildAIProviders
	// builds for the "fully configured" branch (the literal-nil-defect
	// branch's replacement). NOT MockProvider - the whole point of this
	// test is to prove a real, non-nil moderation provider is reached.
	openCodeCfg := aifeedback.OpenCodeConfig{
		BaseURL:    server.URL,
		APIKey:     "test-e2e-opencode-key",
		Model:      "opencode-go/test-model",
		Timeout:    2 * time.Second,
		MaxRetries: 1,
	}
	feedbackProvider := aifeedback.NewOpenCodeFeedbackProvider(openCodeCfg)
	moderationProvider := aifeedback.NewOpenCodeModerationProvider(openCodeCfg)
	safetyClassifier := aifeedback.NewCompositeSafetyClassifier(
		aifeedback.NewDefaultLocalAbuseChecker(),
		moderationProvider,
	)

	authSvc := authStubService()
	svc := aifeedback.NewService(
		aifeedback.NewMemoryRepository(aifeedback.MemoryRepositoryData{
			Words: []aifeedback.MemoryWord{{
				ID: wordID, Text: "work", NormalizedText: "work", WordType: "word", DifficultyLevel: "a2", Status: "active",
			}},
			Meanings: []aifeedback.MemoryMeaning{{
				ID: meaningID, WordID: wordID, PartOfSpeech: "verb", ShortDefinition: "to do a job", Status: "active",
			}},
			UserWords: []aifeedback.MemoryUserWord{{
				ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning",
			}},
		}),
		feedbackProvider,
		safetyClassifier,
		nil,
		learning.NewMemoryIdempotencyStore(),
		nil,
		aifeedback.NewNoopTelemetryRecorder(),
		aifeedback.NewDefaultTaskBuilder(),
		aifeedback.NewDefaultOutputValidator(),
		nil,
		aifeedback.DefaultServiceConfig(),
	)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterAIFeedback(api, svc, authSvc)

	body := `{"sentenceText":"I work every day.","source":"word_detail","attemptId":"00000000-0000-0000-0000-000000000004"}`
	w := httptest.NewRecorder()
	req := submitSentenceFeedbackRequest(t, userID, body, authSvc)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "the real route must return 200 for an ordinary safe sentence when a real moderation provider is wired")
	var out SubmitSentenceFeedbackOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))

	// VOC-034-AC-06: the body must NOT carry the fail-closed business
	// error code that was the user-visible symptom of issue #216. A
	// successful (status=correct) or needs_improvement outcome is fine;
	// the test is only asserting moderation did not fail closed.
	assert.NotEqual(t, aifeedback.ErrorCodeSafetyModerationUnavailable, out.Body.ErrorCode,
		"the fail-closed SAFETY_MODERATION_UNAVAILABLE error code from issue #216 must not appear when the real moderation provider is wired")
	assert.Equal(t, "correct", out.Body.Status, "the real feedback provider must have produced a real, non-fail-closed outcome")
	assert.Equal(t, "I work every day.", out.Body.OriginalSentence)
	assert.False(t, out.Body.MissionCompleted)
	assert.False(t, out.Body.CanRetry)

	// VOC-034-AC-06 (call-count proof): BOTH the moderation transport
	// and the feedback transport must have been hit. A pre-fix wiring
	// (the literal nil) would have caused moderationCalls=0 and
	// feedbackCalls=0 and the test would fail with "moderation
	// provider was never reached" / "feedback provider was never
	// reached" - the exact two phrases the issue's evidence describes.
	assert.GreaterOrEqual(t, atomic.LoadInt32(&counters.moderationCalls), int32(1),
		"moderation provider's transport must be invoked at least once (issue #216 was: it was never reached)")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&counters.feedbackCalls), int32(1),
		"feedback provider's transport must be invoked at least once after moderation allowed the content (issue #216 was: control never reached the feedback provider)")
	assert.GreaterOrEqual(t, atomic.LoadInt32(&counters.sessionCalls), int32(2),
		"two distinct sessions must have been created - one for moderation, one for feedback")
}
