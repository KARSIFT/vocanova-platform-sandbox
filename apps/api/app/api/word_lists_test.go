package api

import (
	"context"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordlists"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type listAPIRepo struct {
	owner, id uuid.UUID
	calls     int
	err       error
	req       wordlists.WriteRequest
}

func (r *listAPIRepo) Get(_ context.Context, u, id uuid.UUID) (*wordlists.WordListDetail, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if u != r.owner || id != r.id {
		return nil, wordlists.ErrNotFound
	}
	return &wordlists.WordListDetail{WordListSummary: wordlists.WordListSummary{ID: id.String(), Name: "Travel", Revision: 1}, Members: []wordlists.WordListMember{}}, nil
}
func (r *listAPIRepo) List(ctx context.Context, u uuid.UUID) (*wordlists.WordListsResponse, error) {
	st, e := r.Get(ctx, u, r.id)
	if e != nil {
		return nil, e
	}
	return &wordlists.WordListsResponse{Items: []wordlists.WordListSummary{st.WordListSummary}}, nil
}
func (r *listAPIRepo) Write(ctx context.Context, req wordlists.WriteRequest, _ time.Time) (*wordlists.WordListDetail, error) {
	r.req = req
	return r.Get(ctx, req.UserID, req.ListID)
}
func TestWordListsAPIAuthCSRFRequesterRevisionAndErrors(t *testing.T) {
	repo := &listAPIRepo{owner: uuid.New(), id: uuid.New()}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("lists", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterWordLists(api, wordlists.NewService(repo, nil), a)
	send := func(method, path, body string, u uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "request")
		if u != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: u}))
		}
		if csrf {
			addCSRF(req, a)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	base := "/api/v1/word-lists"
	single := base + "/" + repo.id.String()
	member := single + "/members/" + uuid.NewString()
	for _, route := range []struct{ method, path, body string }{{"GET", base, ""}, {"GET", single, ""}, {"PUT", single, `{"name":"Travel","expectedRevision":0}`}, {"DELETE", single + "?expectedRevision=1", ""}, {"PUT", member, `{"expectedRevision":1}`}, {"DELETE", member + "?expectedRevision=1", ""}} {
		require.Equal(t, 401, send(route.method, route.path, route.body, uuid.Nil, false).Code)
		if route.method != "GET" {
			require.Equal(t, 403, send(route.method, route.path, route.body, repo.owner, false).Code)
		}
	}
	require.Zero(t, repo.calls)
	require.Equal(t, 404, send("GET", single, "", uuid.New(), false).Code)
	w := send("PUT", single, `{"name":"Travel","expectedRevision":0}`, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, repo.owner, repo.req.UserID)
	require.Equal(t, "put", repo.req.Operation)
	require.Equal(t, "request", repo.req.IdempotencyKey)
	w = send("PUT", member, `{"expectedRevision":1}`, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "add", repo.req.Operation)
	require.Equal(t, 1, repo.req.ExpectedRevision)
	w = send("DELETE", member+"?expectedRevision=3", "", repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "remove", repo.req.Operation)
	require.Equal(t, 3, repo.req.ExpectedRevision)
	w = send("DELETE", single+"?expectedRevision=3", "", repo.owner, true)
	require.Equal(t, 204, w.Code, w.Body.String())
	require.Equal(t, "delete", repo.req.Operation)
	require.Equal(t, 422, send("DELETE", single, "", repo.owner, true).Code)
	for _, tc := range []struct {
		err    error
		status int
	}{{wordlists.ErrInvalid, 400}, {wordlists.ErrNotFound, 404}, {wordlists.ErrConflict, 409}, {wordlists.ErrLimit, 409}} {
		repo.err = tc.err
		require.Equal(t, tc.status, send("GET", single, "", repo.owner, false).Code)
	}
}
