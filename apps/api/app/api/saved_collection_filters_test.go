package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func savedCollectionAPI(t *testing.T) (huma.API, uuid.UUID, uuid.UUID) {
	t.Helper()
	owner, other := uuid.New(), uuid.New()
	now, future := time.Now().UTC(), time.Now().UTC().Add(time.Hour)
	data := learning.MemoryRepositoryData{}
	for i, row := range []struct {
		text, definition, status string
		reviews                  int
		next                     *time.Time
		user                     uuid.UUID
	}{
		{"arrival", "Arriving at a place.", "new", 0, nil, owner},
		{"refund", "Money returned after a purchase.", "new", 2, nil, owner},
		{"refund policy", "Rules for returning money.", "learning", 2, &future, owner},
		{"a_100% offer", "A literal percent and underscore.", "mastered", 5, nil, owner},
		{"refund", "Another learner's private saved entry.", "new", 0, nil, other},
	} {
		word, meaning := uuid.New(), uuid.New()
		data.Words = append(data.Words, learning.MemoryWord{ID: word, Text: row.text, NormalizedText: row.text, Status: "active"})
		data.Meanings = append(data.Meanings, learning.MemoryMeaning{ID: meaning, WordID: word, ShortDefinition: row.definition, PartOfSpeech: "noun", Status: "active"})
		data.UserWords = append(data.UserWords, learning.MemoryUserWord{ID: uuid.New(), UserID: row.user, MeaningID: meaning, Status: row.status, TotalReviewCount: row.reviews, NextReviewAt: row.next, AddedAt: now.Add(-time.Duration(i) * time.Second)})
	}
	authSvc := auth.NewService(auth.NewMemoryRepository(), nil, nil, nil, nil, auth.Config{})
	svc := learning.NewService(learning.NewMemoryRepository(data), learning.NewMemoryIdempotencyStore(), nil)
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Saved collection filters", "1"))
	api.UseMiddleware(withHumaContext)
	RegisterLearning(api, svc, authSvc)
	return api, owner, other
}

func savedCollectionRequest(t *testing.T, api huma.API, owner uuid.UUID, query string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-words"+query, nil)
	if owner != uuid.Nil {
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: owner}))
	}
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return w.Code, body
}

func TestSavedCollectionFiltersSearchBeforePaginationAndBindCursor(t *testing.T) {
	api, owner, other := savedCollectionAPI(t)
	status, body := savedCollectionRequest(t, api, owner, "?q=refund&limit=1")
	require.Equal(t, http.StatusOK, status)
	items := body["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, "refund", items[0].(map[string]any)["wordText"], "query must find a later saved word before limiting the page")
	require.Equal(t, float64(2), body["totalCount"], "count covers every matching owned meaning")
	require.Equal(t, true, body["hasMore"])
	require.Equal(t, "new", items[0].(map[string]any)["status"], "raw status remains compatible")
	require.Equal(t, "learning", items[0].(map[string]any)["reviewState"])
	require.Equal(t, true, items[0].(map[string]any)["due"])
	cursor := body["nextCursor"].(string)
	status, second := savedCollectionRequest(t, api, owner, "?q=%20REFUND%20&limit=1&after="+url.QueryEscape(cursor))
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, float64(2), second["totalCount"])
	require.Equal(t, "refund policy", second["items"].([]any)[0].(map[string]any)["wordText"])
	require.Equal(t, false, second["hasMore"])
	for _, query := range []string{"?q=arrival", "?q=refund&stage=learning", "?q=refund&due=true"} {
		status, _ = savedCollectionRequest(t, api, owner, query+"&after="+url.QueryEscape(cursor))
		require.Equal(t, http.StatusBadRequest, status, "cursor must bind every normalized filter")
	}
	status, _ = savedCollectionRequest(t, api, other, "?q=refund&after="+url.QueryEscape(cursor))
	require.Equal(t, http.StatusBadRequest, status, "cursor cannot silently continue another requester's collection")
}

func TestSavedCollectionFiltersValidateAndProjectLearningState(t *testing.T) {
	api, owner, _ := savedCollectionAPI(t)
	for _, query := range []string{"?q=refund&stage=learning&due=true", "?q=returned&stage=learning&due=true"} {
		status, body := savedCollectionRequest(t, api, owner, query)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, float64(1), body["totalCount"])
		require.Equal(t, "refund", body["items"].([]any)[0].(map[string]any)["wordText"])
	}
	status, body := savedCollectionRequest(t, api, owner, "?q="+url.QueryEscape("_100%"))
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, float64(1), body["totalCount"])
	status, body = savedCollectionRequest(t, api, owner, "?q=missing")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, float64(0), body["totalCount"])
	require.Empty(t, body["items"])
	for _, query := range []string{"?stage=known", "?q=" + url.QueryEscape(strings.Repeat("é", 101)), "?q=%00"} {
		status, _ = savedCollectionRequest(t, api, owner, query)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, status)
	}
	status, _ = savedCollectionRequest(t, api, uuid.Nil, "?q=refund")
	require.Equal(t, http.StatusUnauthorized, status)
}
