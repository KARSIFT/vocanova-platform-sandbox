package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLearningAPI(t *testing.T) (huma.API, *learning.Service, *auth.Service) {
	repo := auth.NewMemoryRepository()
	authSvc := auth.NewService(repo, nil, nil, nil, nil, auth.Config{
		Cookie: auth.CookieConfig{Name: "vocanova_session", CSRName: "vocanova_csrf"},
	})

	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	learningRepo := learning.NewMemoryRepository(learning.MemoryRepositoryData{
		Words: []learning.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []learning.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
	})
	learningSvc := learning.NewService(learningRepo, learning.NewMemoryIdempotencyStore(), nil)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterLearning(api, learningSvc, authSvc)
	return api, learningSvc, authSvc
}

func authenticatedLearningRequest(t *testing.T, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-words", nil)
	ctx := WithRequester(req.Context(), &auth.User{ID: userID})
	return req.WithContext(ctx)
}

func addCSRF(req *http.Request, authSvc *auth.Service) {
	csrfToken, csrfCookie := authSvc.IssueCSRFCookie()
	req.AddCookie(csrfCookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
}

func TestListSavedWordsRequiresAuth(t *testing.T) {
	api, _, _ := testLearningAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-words", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListSavedWordsReturnsSavedMeanings(t *testing.T) {
	api, svc, _ := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), learning.SaveUserWordRequest{
		UserID: userID, MeaningID: meaningID, Source: "journey", IdempotencyKey: "k1",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := authenticatedLearningRequest(t, userID)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListSavedWordsOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Items, 1)
	assert.Equal(t, "boarding-pass", body.Body.Items[0].WordSlug)
	assert.Equal(t, "A document.", body.Body.Items[0].ShortDefinition)
	assert.True(t, body.Body.Items[0].Saved)
}

func TestSaveUserWordRequiresAuth(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSaveUserWordRequiresCSRF(t *testing.T) {
	api, _, _ := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := authenticatedLearningRequest(t, userID)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSaveUserWordRequiresIdempotencyKey(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestSaveUserWordSavesMeaning(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "save-key-1")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body SaveUserWordOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", body.Body.MeaningID)
	assert.Equal(t, "boarding-pass", body.Body.WordSlug)
	assert.Equal(t, "journey", body.Body.Source)
	assert.True(t, body.Body.Saved)
}

func TestSaveUserWordUnknownMeaningReturns404(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000099","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "save-key-1")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSaveUserWordIdempotencyKeyReplay(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	key := "replay-key"
	body := `{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`

	save := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
		addCSRF(req, authSvc)
		api.Adapter().ServeHTTP(w, req)
		return w
	}

	first := save()
	require.Equal(t, http.StatusOK, first.Code)

	second := save()
	require.Equal(t, http.StatusOK, second.Code)
	var firstBody, secondBody SaveUserWordOutput
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstBody.Body))
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondBody.Body))
	assert.Equal(t, firstBody.Body.UserWordID, secondBody.Body.UserWordID)
}

func TestSaveUserWordIdempotencyKeyConflict(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	key := "conflict-key"

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(`{"meaningId":"00000000-0000-0000-0000-000000000002","source":"manual"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSaveUserWordIdempotencyKeyIsolatedByUser(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	key := "shared-key"
	body := `{"meaningId":"00000000-0000-0000-0000-000000000002","source":"journey"}`

	for _, userID := range []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		uuid.MustParse("00000000-0000-0000-0000-000000000004"),
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/user-words", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
		addCSRF(req, authSvc)
		api.Adapter().ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, userID)
	}
}

func TestUnsaveUserWordRequiresAuth(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", nil)
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnsaveUserWordRequiresCSRF(t *testing.T) {
	api, _, _ := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUnsaveUserWord(t *testing.T) {
	api, svc, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), learning.SaveUserWordRequest{
		UserID: userID, MeaningID: meaningID, Source: "journey", IdempotencyKey: "k1",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)

	states, err := svc.IsSaved(t.Context(), userID, []uuid.UUID{meaningID})
	require.NoError(t, err)
	assert.False(t, states[meaningID])
}

func TestUnsaveUserWordUnknownReturns404(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUnsaveUserWordCannotTouchOtherUser(t *testing.T) {
	api, svc, authSvc := testLearningAPI(t)
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), learning.SaveUserWordRequest{
		UserID: owner, MeaningID: meaningID, Source: "journey", IdempotencyKey: "k1",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-words/00000000-0000-0000-0000-000000000002", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: other}))
	addCSRF(req, authSvc)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListSavedWordsCrossUserIsolation(t *testing.T) {
	api, svc, _ := testLearningAPI(t)
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	_, err := svc.SaveUserWord(t.Context(), learning.SaveUserWordRequest{
		UserID: owner, MeaningID: meaningID, Source: "journey", IdempotencyKey: "k1",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := authenticatedLearningRequest(t, other)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListSavedWordsOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Empty(t, body.Body.Items)
}
