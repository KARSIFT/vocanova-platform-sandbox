package api

import (
	"context"
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
