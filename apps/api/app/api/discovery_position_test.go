package api

import (
	"context"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/discoveryposition"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type positionFixture struct {
	positions map[uuid.UUID]*discoveryposition.Position
}

func (r *positionFixture) Get(_ context.Context, u uuid.UUID) (*discoveryposition.Position, error) {
	return r.positions[u], nil
}
func (r *positionFixture) Put(_ context.Context, u uuid.UUID, v discoveryposition.Update) (*discoveryposition.Position, error) {
	current := r.positions[u]
	revision := int64(0)
	if current != nil {
		revision = current.Revision
	}
	if v.ExpectedRevision != revision {
		return nil, discoveryposition.ErrConflict
	}
	p := &discoveryposition.Position{CatalogVersion: v.CatalogVersion, Band: v.Band, Rank: v.Rank, Active: v.Active, Revision: revision + 1}
	r.positions[u] = p
	return p, nil
}
func TestDiscoveryPositionHTTPContract(t *testing.T) {
	api, authSvc, _ := testOnboardingAPI(t)
	repo := &positionFixture{map[uuid.UUID]*discoveryposition.Position{}}
	RegisterDiscoveryPosition(api, discoveryposition.NewService(repo), authSvc)
	uid, other := uuid.New(), uuid.New()
	request := func(method, body string, csrf bool, user uuid.UUID) *httptest.ResponseRecorder {
		if body != "" {
			var value map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &value))
			value["expectedAccountId"] = user.String()
			encoded, err := json.Marshal(value)
			require.NoError(t, err)
			body = string(encoded)
		}
		req := onbRequesterRequest(t, method, "/api/v1/discovery-position", body, user)
		if csrf {
			token, cookie := authSvc.IssueCSRFCookie()
			req.AddCookie(cookie)
			req.Header.Set("X-CSRF-Token", token)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		return w
	}
	w := request(http.MethodGet, "", false, uid)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"position":null`)
	body := `{"catalogVersion":"english-10000-v1","band":44,"rank":4330,"active":true,"expectedRevision":0}`
	require.Equal(t, 403, request(http.MethodPut, body, false, uid).Code)
	w = request(http.MethodPut, body, true, uid)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"rank":4330`)
	require.Equal(t, 409, request(http.MethodPut, body, true, uid).Code)
	w = request(http.MethodGet, "", false, other)
	require.Contains(t, w.Body.String(), `"position":null`)
	for _, invalid := range []string{`{"catalogVersion":"old","band":44,"rank":4330,"active":true,"expectedRevision":1}`, `{"catalogVersion":"english-10000-v1","band":43,"rank":4330,"active":true,"expectedRevision":1}`} {
		require.Equal(t, 400, request(http.MethodPut, invalid, true, uid).Code)
	}
	for _, invalid := range []string{`{"catalogVersion":"english-10000-v1","band":44,"rank":4330,"expectedRevision":1}`, `{"catalogVersion":"english-10000-v1","band":44,"rank":4330,"active":true}`} {
		require.Equal(t, 422, request(http.MethodPut, invalid, true, uid).Code)
	}
	w = httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/discovery-position", nil))
	require.Equal(t, 401, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
}

func TestDiscoveryPositionRejectsDifferentRenderedOwner(t *testing.T) {
	api, authSvc, _ := testOnboardingAPI(t)
	a, b := uuid.New(), uuid.New()
	repo := &positionFixture{map[uuid.UUID]*discoveryposition.Position{
		a: {CatalogVersion: "english-10000-v1", Band: 44, Rank: 4330, Revision: 1},
		b: {CatalogVersion: "english-10000-v1", Band: 12, Rank: 1150, Revision: 1},
	}}
	RegisterDiscoveryPosition(api, discoveryposition.NewService(repo), authSvc)
	// Owner identity is mandatory; missing identity cannot fall back to the
	// current cookie and thereby acquire a different account during a retry.
	missing := onbRequesterRequest(t, http.MethodPut, "/api/v1/discovery-position", `{"catalogVersion":"english-10000-v1","band":44,"rank":4331,"active":true,"expectedRevision":1}`, b)
	token, cookie := authSvc.IssueCSRFCookie()
	missing.AddCookie(cookie)
	missing.Header.Set("X-CSRF-Token", token)
	missingResponse := httptest.NewRecorder()
	api.Adapter().ServeHTTP(missingResponse, missing)
	require.Equal(t, 422, missingResponse.Code)
	for range 2 {
		req := onbRequesterRequest(t, http.MethodPut, "/api/v1/discovery-position", `{"catalogVersion":"english-10000-v1","band":44,"rank":4331,"active":true,"expectedRevision":1,"expectedAccountId":"`+a.String()+`"}`, b)
		token, cookie := authSvc.IssueCSRFCookie()
		req.AddCookie(cookie)
		req.Header.Set("X-CSRF-Token", token)
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		require.Equal(t, 403, w.Code, w.Body.String())
		require.Contains(t, w.Body.String(), "Discovery account changed")
		require.Equal(t, 1150, repo.positions[b].Rank)
		require.EqualValues(t, 1, repo.positions[b].Revision)
	}
}
