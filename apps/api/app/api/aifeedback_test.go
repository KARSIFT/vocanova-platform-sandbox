package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestReportSentenceFeedbackRequiresCSRF(t *testing.T) {
	api, _, _ := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestReportSentenceFeedbackUnknownAttemptReturns404(t *testing.T) {
	api, _, authSvc := testAIFeedbackAPI(t, aifeedback.MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sentence-feedback/00000000-0000-0000-0000-000000000001/reports", bytes.NewReader([]byte(`{"reason":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
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
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/sentence-feedback/%s/reports", attemptID), bytes.NewReader([]byte(`{"reason":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
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
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/sentence-feedback/%s/reports", attemptID), bytes.NewReader([]byte(`{"reason":"bad","classification":"incorrect"}`)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: owner}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
