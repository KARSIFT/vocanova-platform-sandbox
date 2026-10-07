package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dictionary"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type dictionaryStub struct {
	calls int
	err   error
	word  string
}

func (s *dictionaryStub) Lookup(_ context.Context, word string) (dictionary.Entry, error) {
	s.calls++
	s.word = word
	return dictionary.Entry{Word: word, Meanings: []dictionary.Meaning{{PartOfSpeech: "noun", Definitions: []dictionary.Definition{{Definition: "A test definition.", Example: "A test example."}}}}, License: "Original notice fixture"}, s.err
}

func dictionaryTestAPI(svc dictionaryLookup, limit int) huma.API {
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Test", "1"))
	api.UseMiddleware(withHumaContext)
	registerDictionary(api, svc, auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, limit))
	return api
}

func dictionaryRequest(api huma.API, query string, user uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/api/v1/dictionary?q="+url.QueryEscape(query), nil)
	if user != uuid.Nil {
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
	}
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	return w
}

func TestDictionaryRequiresAuthenticationAndExplicitValidWord(t *testing.T) {
	stub := &dictionaryStub{}
	api := dictionaryTestAPI(stub, 20)
	require.Equal(t, 401, dictionaryRequest(api, "book", uuid.Nil).Code)
	for _, q := range []string{"", "  ", "book sentence", "https://example.com", "word%2fsecret"} {
		require.Equal(t, 400, dictionaryRequest(api, q, uuid.New()).Code)
	}
	require.Zero(t, stub.calls)
}

func TestDictionaryProjectsAttributionWithoutCanonicalIdentity(t *testing.T) {
	stub := &dictionaryStub{}
	w := dictionaryRequest(dictionaryTestAPI(stub, 20), " Book ", uuid.New())
	require.Equal(t, 200, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.Equal(t, "book", stub.word)
	var out DictionaryEntryDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	require.Equal(t, "book", out.Word)
	require.Equal(t, "WordNet 3.0", out.Attribution.Provider)
	require.Equal(t, "Original notice fixture", out.Attribution.Licenses[0].Text)
	require.Equal(t, "A test example.", out.Meanings[0].Definitions[0].Example)
	for _, private := range []string{"meaningId", "wordId", "userWordId", "saved", "phonetic"} {
		require.NotContains(t, w.Body.String(), `"`+private+`"`)
	}
}

func TestDictionaryMissingAndFailureHaveDistinctSafeResponses(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{dictionary.ErrNotFound, 404}, {dictionary.ErrUnavailable, 503}, {errors.New("private failure detail"), 503}} {
		w := dictionaryRequest(dictionaryTestAPI(&dictionaryStub{err: tc.err}, 20), "book", uuid.New())
		require.Equal(t, tc.status, w.Code)
		require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		require.NotContains(t, w.Body.String(), "private failure detail")
	}
}

func TestDictionaryRateLimitIsRequesterScoped(t *testing.T) {
	stub := &dictionaryStub{}
	api := dictionaryTestAPI(stub, 1)
	user := uuid.New()
	require.Equal(t, 200, dictionaryRequest(api, "book", user).Code)
	w := dictionaryRequest(api, "apple", user)
	require.Equal(t, 429, w.Code)
	require.Equal(t, "60", w.Header().Get("Retry-After"))
	require.Equal(t, 200, dictionaryRequest(api, "apple", uuid.New()).Code)
	require.Equal(t, 2, stub.calls)
}

func TestDictionaryActualBundleIsExposedByContract(t *testing.T) {
	w := dictionaryRequest(NewContractAPI(), "serendipity", uuid.New())
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "fortunate discoveries")
	require.Contains(t, w.Body.String(), "Copyright 2006")
}
