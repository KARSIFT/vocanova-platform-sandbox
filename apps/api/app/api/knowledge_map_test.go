package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type catalogHTTPFixture struct {
	user uuid.UUID
	req  knowledgecatalog.DecisionRequest
	err  error
}

func (r *catalogHTTPFixture) Read(_ context.Context, u uuid.UUID, band int, _ time.Time) (*knowledgecatalog.Map, error) {
	r.user = u
	return &knowledgecatalog.Map{CatalogVersion: knowledgecatalog.CatalogVersion, TotalWords: 10000, BandSize: 100, CurrentBand: knowledgecatalog.CurrentBand{Number: band, Items: []knowledgecatalog.Item{}}}, r.err
}
func (r *catalogHTTPFixture) Decide(_ context.Context, req knowledgecatalog.DecisionRequest, _ time.Time) (*knowledgecatalog.DecisionResult, error) {
	r.req = req
	return &knowledgecatalog.DecisionResult{MeaningID: req.MeaningID, Decision: req.Decision, Revision: req.ExpectedRevision, Saved: req.Decision == "learn", SelfReportedKnown: req.Decision == "known"}, r.err
}
func TestKnowledgeMapAuthenticationValidationCSRFAndPrivateCaching(t *testing.T) {
	api, _, authSvc := testLearningAPI(t)
	repo := &catalogHTTPFixture{}
	RegisterKnowledgeMap(api, knowledgecatalog.NewService(repo, nil), authSvc)
	user, m := uuid.New(), uuid.New()
	revision := strings.Repeat("a", 64)
	body, _ := json.Marshal(map[string]any{"meaningId": m.String(), "decision": "learn", "expectedRevision": revision})
	call := func(method, path, body, key string, actor uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		if actor != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: actor}))
		}
		if csrf {
			addCSRF(req, authSvc)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, call("GET", "/api/v1/knowledge-map", "", "", uuid.Nil, false).Code)
	w := call("GET", "/api/v1/knowledge-map", "", "", user, false)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	require.Contains(t, w.Body.String(), `"number":1`)
	require.Equal(t, user, repo.user)
	require.Equal(t, 422, call("GET", "/api/v1/knowledge-map?band=101", "", "", user, false).Code)
	require.Equal(t, 403, call("POST", "/api/v1/knowledge-map/decisions", string(body), "one", user, false).Code)
	require.Equal(t, 422, call("POST", "/api/v1/knowledge-map/decisions", string(body), "", user, true).Code)
	w = call("POST", "/api/v1/knowledge-map/decisions", string(body), "one", user, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, user, repo.req.UserID)
	require.Equal(t, m, repo.req.MeaningID)
	require.Equal(t, revision, repo.req.ExpectedRevision)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	repo.err = knowledgecatalog.ErrStale
	require.Equal(t, 409, call("POST", "/api/v1/knowledge-map/decisions", string(body), "two", user, true).Code)
	repo.err = knowledgecatalog.ErrNotFound
	require.Equal(t, 404, call("GET", "/api/v1/knowledge-map", "", "", user, false).Code)
}
