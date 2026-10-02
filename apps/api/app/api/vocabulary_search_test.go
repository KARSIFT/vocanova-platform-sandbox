package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCanonicalWordsAuthenticatedFilteredOverlay(t *testing.T) {
	api, _ := testContentAPI(t, nil)
	owner := uuid.MustParse("00000000-0000-0000-0000-000000000009")
	for _, tc := range []struct {
		name   string
		user   uuid.UUID
		path   string
		status int
		saved  bool
		count  int
	}{
		{"unauthenticated", uuid.Nil, "/api/v1/canonical-words", 401, false, 0},
		{"owner", owner, "/api/v1/canonical-words?q=%20DOCUMENT%20&category=travel", 200, true, 1},
		{"other learner", uuid.New(), "/api/v1/canonical-words?q=document", 200, false, 1},
		{"empty", owner, "/api/v1/canonical-words?q=absent", 200, false, 0},
		{"unknown category", owner, "/api/v1/canonical-words?category=secret", 400, false, 0},
		{"unknown level", owner, "/api/v1/canonical-words?level=c2", 400, false, 0},
		{"long query", owner, "/api/v1/canonical-words?q=" + url.QueryEscape(strings.Repeat("é", 101)), 400, false, 0},
		{"bad cursor", owner, "/api/v1/canonical-words?after=bad", 400, false, 0},
		{"invalid database text", owner, "/api/v1/canonical-words?q=%00", 400, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.user != uuid.Nil {
				req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: tc.user}))
			}
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, req)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			if tc.status != 200 {
				return
			}
			var output SearchCanonicalWordsOutput
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &output.Body))
			assert.Len(t, output.Body.Items, tc.count)
			assert.Equal(t, tc.count, output.Body.TotalCount)
			assert.False(t, output.Body.HasMore)
			if tc.count == 0 {
				assert.Contains(t, w.Body.String(), `"items":[]`)
				return
			}
			assert.Equal(t, "boarding-pass", output.Body.Items[0].WordSlug)
			assert.Equal(t, tc.saved, output.Body.Items[0].Saved)
			if tc.saved {
				assert.NotEmpty(t, output.Body.Items[0].UserWordID)
				assert.Equal(t, "learning", output.Body.Items[0].ReviewState)
				assert.True(t, output.Body.Items[0].Due)
			} else {
				assert.Empty(t, output.Body.Items[0].UserWordID)
				assert.Empty(t, output.Body.Items[0].ReviewState)
				assert.False(t, output.Body.Items[0].Due)
			}
		})
	}
}

func TestKnowledgeSummaryAuthenticatedRequesterCounts(t *testing.T) {
	api, svc, _ := testLearningAPI(t)
	owner, other := uuid.New(), uuid.New()
	_, err := svc.SaveUserWord(t.Context(), learning.SaveUserWordRequest{UserID: owner, MeaningID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Source: "search", IdempotencyKey: "summary-save"})
	require.NoError(t, err)
	for _, tc := range []struct {
		user          uuid.UUID
		status, count int
	}{{uuid.Nil, 401, 0}, {owner, 200, 1}, {other, 200, 0}} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge-summary", nil)
		if tc.user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: tc.user}))
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code, w.Body.String())
		if tc.status != 200 {
			continue
		}
		var summary KnowledgeSummaryDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &summary))
		assert.Equal(t, KnowledgeSummaryDTO{Saved: tc.count, New: tc.count, Due: tc.count}, summary)
	}
}
