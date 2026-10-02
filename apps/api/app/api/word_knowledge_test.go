package api

import (
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMeaningKnowledgeAuthPrivacyCSRFAndReplay(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	m, u, other := uuid.New(), uuid.New(), uuid.New()
	svc := wordknowledge.NewService(wordknowledge.NewMemoryRepository([]uuid.UUID{m}), nil)
	RegisterWordKnowledge(api, svc, authSvc)
	call := func(method, body, key string, user uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/meaning-knowledge/"+m.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		if user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		}
		if csrf {
			addCSRF(req, authSvc)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, call(http.MethodGet, "", "", uuid.Nil, false).Code)
	body := `{"selfReportedKnown":true,"note":"Private <script>not executable</script> text"}`
	require.Equal(t, 403, call(http.MethodPut, body, "one", u, false).Code)
	require.Equal(t, 422, call(http.MethodPut, body, "", u, true).Code)
	require.Equal(t, 422, call(http.MethodPut, `{"note":"missing boolean"}`, "missing", u, true).Code)
	w := call(http.MethodPut, body, "one", u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	var st MeaningKnowledgeDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &st))
	require.True(t, st.SelfReportedKnown)
	require.NotNil(t, st.UpdatedAt)
	w = call(http.MethodGet, "", "", other, false)
	require.Equal(t, 200, w.Code)
	require.NotContains(t, w.Body.String(), "Private")
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &st))
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, 409, call(http.MethodPut, `{"selfReportedKnown":false,"note":""}`, "one", u, true).Code)
	require.Equal(t, 403, call(http.MethodDelete, "", "clear", u, false).Code)
	require.Equal(t, 204, call(http.MethodDelete, "", "clear", other, true).Code)
	require.Contains(t, call(http.MethodGet, "", "", u, false).Body.String(), "Private")
	require.Equal(t, 204, call(http.MethodDelete, "", "clear", u, true).Code)
	w = call(http.MethodPut, body, "one", u, true)
	require.Equal(t, 200, w.Code)
	require.NotContains(t, w.Body.String(), "Private")
	require.Contains(t, w.Body.String(), `"selfReportedKnown":false`)
}

func TestKnowledgeOverlaysDoNotExposePrivateNotes(t *testing.T) {
	api, svc := testContentAPI(t, nil)
	m := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	u := uuid.MustParse("00000000-0000-0000-0000-000000000009")
	knowledge := wordknowledge.NewService(wordknowledge.NewMemoryRepository([]uuid.UUID{m}), nil)
	_, err := knowledge.Write(t.Context(), wordknowledge.WriteRequest{UserID: u, MeaningID: m, SelfReportedKnown: true, Note: "private-marker-do-not-leak", IdempotencyKey: "one"})
	require.NoError(t, err)
	svc.SetKnowledgeReader(knowledge)
	for _, path := range []string{"/api/v1/canonical-words", "/api/v1/canonical-words/boarding-pass", "/api/v1/journey-situations/airport"} {
		for _, user := range []uuid.UUID{u, uuid.New()} {
			req := httptest.NewRequest("GET", path, nil)
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, req)
			require.Equal(t, 200, w.Code, w.Body.String())
			require.NotContains(t, w.Body.String(), "private-marker")
			if user == u {
				require.Contains(t, w.Body.String(), `"selfReportedKnown":true`)
			} else {
				require.Contains(t, w.Body.String(), `"selfReportedKnown":false`)
			}
		}
	}
	for _, tc := range []struct {
		filter string
		user   uuid.UUID
		count  int
		status int
	}{{"known", u, 1, 200}, {"saved", u, 1, 200}, {"unexplored", u, 0, 200}, {"known", uuid.New(), 0, 200}, {"unexplored", uuid.New(), 1, 200}, {"mastered", u, 0, 400}} {
		req := httptest.NewRequest("GET", "/api/v1/canonical-words?knowledge="+tc.filter, nil)
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: tc.user}))
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code, w.Body.String())
		if tc.status == 200 {
			var out SearchCanonicalWordsOutput
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out.Body))
			require.Equal(t, tc.count, out.Body.TotalCount)
		}
	}
	learningAPI, learningSvc, _ := testLearningAPI(t)
	learningSvc.SetKnowledgeReader(knowledge)
	req := httptest.NewRequest("GET", "/api/v1/knowledge-summary", nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: u}))
	w := httptest.NewRecorder()
	learningAPI.Adapter().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	var summary KnowledgeSummaryDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &summary))
	require.Equal(t, 1, summary.SelfReportedKnown)
	require.Zero(t, summary.Mastered)
	require.Zero(t, summary.Saved)
	require.NotContains(t, w.Body.String(), "private-marker")
}

func TestMeaningKnowledgePartialAssessmentRequiresAuthAndPreservesNote(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	m, u, other := uuid.New(), uuid.New(), uuid.New()
	svc := wordknowledge.NewService(wordknowledge.NewMemoryRepository([]uuid.UUID{m}), nil)
	RegisterWordKnowledge(api, svc, authSvc)
	call := func(method, body, key string, user uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/meaning-knowledge/"+m.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		if user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		}
		if csrf {
			addCSRF(req, authSvc)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	patch := `{"selfReportedKnown":true}`
	require.Equal(t, 401, call(http.MethodPatch, patch, "mark", uuid.Nil, true).Code)
	require.Equal(t, 403, call(http.MethodPatch, patch, "mark", u, false).Code)
	require.Equal(t, 422, call(http.MethodPatch, patch, "", u, true).Code)
	for _, body := range []string{`{}`, `{"selfReportedKnown":null}`, `{"selfReportedKnown":"yes"}`} {
		require.Equal(t, 422, call(http.MethodPatch, body, "invalid", u, true).Code)
	}
	require.Equal(t, 200, call(http.MethodPut, `{"selfReportedKnown":false,"note":"Keep this private note"}`, "note", u, true).Code)
	w := call(http.MethodPatch, patch, "mark", u, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	var st MeaningKnowledgeDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &st))
	require.True(t, st.SelfReportedKnown)
	require.Equal(t, "Keep this private note", st.Note)
	w = call(http.MethodPatch, patch, "mark", other, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), "Keep this private note")
	require.Equal(t, 409, call(http.MethodPatch, `{"selfReportedKnown":false}`, "mark", u, true).Code)
	require.Equal(t, 200, call(http.MethodPut, `{"selfReportedKnown":false,"note":"Newer private note"}`, "newer", u, true).Code)
	w = call(http.MethodPatch, patch, "mark", u, true)
	require.Equal(t, 200, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &st))
	require.False(t, st.SelfReportedKnown)
	require.Equal(t, "Newer private note", st.Note)
}
