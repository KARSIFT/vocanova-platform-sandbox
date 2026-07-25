package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testReviewsAPI(t *testing.T, data reviews.MemoryRepositoryData) (huma.API, *reviews.Service) {
	repo := reviews.NewMemoryRepository(data)
	svc := reviews.NewService(repo)

	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authStubService()))
	RegisterReviews(api, svc)
	return api, svc
}

func authenticatedReviewsRequest(t *testing.T, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/due", nil)
	ctx := WithRequester(req.Context(), &auth.User{ID: userID})
	return req.WithContext(ctx)
}

func TestListReviewsDueRequiresAuth(t *testing.T) {
	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/due", nil)
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListReviewsDueReturnsSavedNeverReviewed(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{
		Words: []reviews.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []reviews.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []reviews.MemoryUserWord{
			{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
		},
	})

	w := httptest.NewRecorder()
	req := authenticatedReviewsRequest(t, userID)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Items, 1)
	assert.Equal(t, userWordID.String(), body.Body.Items[0].UserWordID)
	assert.Equal(t, "boarding-pass", body.Body.Items[0].WordSlug)
	assert.Equal(t, "new", body.Body.Items[0].Status)
	assert.Equal(t, 1, body.Body.TotalCount)
}

func TestListReviewsDueExcludesFutureNextReviewAt(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	future := time.Now().UTC().Add(time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{
		Words: []reviews.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []reviews.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []reviews.MemoryUserWord{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), UserID: userID, MeaningID: meaningID, Status: "learning", Source: "journey", ReviewStep: 1, NextReviewAt: &future},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), UserID: userID, MeaningID: meaningID, Status: "learning", Source: "journey", ReviewStep: 2, NextReviewAt: &past},
		},
	})

	w := httptest.NewRecorder()
	req := authenticatedReviewsRequest(t, userID)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Items, 1)
	assert.Equal(t, 2, body.Body.Items[0].ReviewStep)
	assert.Equal(t, 1, body.Body.TotalCount)
}

func TestListReviewsDueExcludesNonDueStatusesAndDeleted(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	deletedAt := time.Now().UTC().Add(-time.Hour)

	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{
		Words: []reviews.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []reviews.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []reviews.MemoryUserWord{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), UserID: userID, MeaningID: meaningID, Status: "mastered", Source: "journey"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), UserID: userID, MeaningID: meaningID, Status: "ignored", Source: "journey"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000006"), UserID: userID, MeaningID: meaningID, Status: "archived", Source: "journey"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000007"), UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", DeletedAt: &deletedAt},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000008"), UserID: userID, MeaningID: meaningID, Status: "reviewing", Source: "journey"},
		},
	})

	w := httptest.NewRecorder()
	req := authenticatedReviewsRequest(t, userID)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	require.Len(t, body.Body.Items, 1)
	assert.Equal(t, "reviewing", body.Body.Items[0].Status)
	assert.Equal(t, 1, body.Body.TotalCount)
}

func TestListReviewsDueCrossUserIsolation(t *testing.T) {
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	other := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{
		Words: []reviews.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []reviews.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []reviews.MemoryUserWord{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), UserID: owner, MeaningID: meaningID, Status: "new", Source: "journey"},
		},
	})

	w := httptest.NewRecorder()
	req := authenticatedReviewsRequest(t, other)
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body.Body))
	assert.Empty(t, body.Body.Items)
	assert.Equal(t, 0, body.Body.TotalCount)
}

func TestListReviewsDueInvalidCursor(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/due?after=not-valid", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListReviewsDuePagination(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	past1 := time.Now().UTC().Add(-1 * time.Hour)
	past2 := time.Now().UTC().Add(-2 * time.Hour)
	past3 := time.Now().UTC().Add(-3 * time.Hour)

	api, _ := testReviewsAPI(t, reviews.MemoryRepositoryData{
		Words: []reviews.MemoryWord{
			{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"},
		},
		Meanings: []reviews.MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []reviews.MemoryUserWord{
			{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000a"), UserID: userID, MeaningID: meaningID, Status: "reviewing", Source: "journey", ReviewStep: 1, NextReviewAt: &past1},
			{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000b"), UserID: userID, MeaningID: meaningID, Status: "reviewing", Source: "journey", ReviewStep: 2, NextReviewAt: &past2},
			{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000c"), UserID: userID, MeaningID: meaningID, Status: "reviewing", Source: "journey", ReviewStep: 3, NextReviewAt: &past3},
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/due?limit=2", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var first ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &first.Body))
	require.Len(t, first.Body.Items, 2)
	assert.Equal(t, 3, first.Body.TotalCount)
	assert.NotEmpty(t, first.Body.NextCursor)
	// Ascending by next_review_at means past3 (oldest) first, then past2.
	assert.Equal(t, 3, first.Body.Items[0].ReviewStep)
	assert.Equal(t, 2, first.Body.Items[1].ReviewStep)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/reviews/due?limit=2&after="+first.Body.NextCursor, nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
	api.Adapter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var second ListReviewsDueOutput
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &second.Body))
	require.Len(t, second.Body.Items, 1)
	assert.Equal(t, 1, second.Body.Items[0].ReviewStep)
	assert.Equal(t, 3, second.Body.TotalCount)
	assert.Empty(t, second.Body.NextCursor)
}
