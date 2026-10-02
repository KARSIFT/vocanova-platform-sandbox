package api

import (
	"context"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
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

type practiceAPIRepo struct {
	owner, id uuid.UUID
	calls     int
	err       error
	action    practice.Action
	request   practice.StartRequest
}

func (r *practiceAPIRepo) state(u uuid.UUID) (*practice.State, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if u != r.owner {
		return nil, practice.ErrNotFound
	}
	return &practice.State{ID: r.id, UserID: u, Snapshot: practice.Snapshot{Mode: r.request.Mode, ContentVersion: practice.ContentVersion, GradingVersion: practice.GradingVersion}}, nil
}
func (r *practiceAPIRepo) List(_ context.Context, u uuid.UUID) ([]practice.State, int, error) {
	s, e := r.state(u)
	if e != nil {
		return nil, 0, e
	}
	return []practice.State{*s}, 2, nil
}
func (r *practiceAPIRepo) Start(_ context.Context, u uuid.UUID, req practice.StartRequest, _ string, _ time.Time) (*practice.State, error) {
	r.request = req
	return r.state(u)
}
func (r *practiceAPIRepo) Get(_ context.Context, u, id uuid.UUID) (*practice.State, error) {
	if id != r.id {
		return nil, practice.ErrNotFound
	}
	return r.state(u)
}
func (r *practiceAPIRepo) Act(_ context.Context, u, id uuid.UUID, a practice.Action, _ string, _ time.Time) (*practice.State, error) {
	r.action = a
	return r.Get(context.Background(), u, id)
}
func TestPracticeAPIAuthCSRFContractAndErrors(t *testing.T) {
	repo := &practiceAPIRepo{owner: uuid.New(), id: uuid.New()}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("practice", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterPractice(api, practice.NewService(repo, nil), a)
	send := func(method, path, body string, user uuid.UUID, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "request")
		if user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		}
		if csrf {
			addCSRF(req, a)
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	base := "/api/v1/practice-sessions"
	single := base + "/" + repo.id.String()
	body := `{"mode":"typed_recall","lessonKey":"airport"}`
	for _, route := range []struct{ method, path string }{{"GET", base}, {"POST", base}, {"GET", single}, {"POST", single + "/actions"}} {
		require.Equal(t, 401, send(route.method, route.path, body, uuid.Nil, false).Code)
	}
	require.Zero(t, repo.calls)
	require.Equal(t, 403, send("POST", base, body, repo.owner, false).Code)
	require.Zero(t, repo.calls)
	w := send("POST", base, body, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "airport", repo.request.LessonKey)
	require.Equal(t, 404, send("GET", single, "", uuid.New(), false).Code)
	action := `{"stepId":"opaque-step","expectedRevision":2,"clientActionId":"answer-one","action":"answer","typedAnswer":"boarding pass"}`
	w = send("POST", single+"/actions", action, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "boarding pass", repo.action.TypedAnswer)
	require.Equal(t, 2, repo.action.ExpectedRevision)
	for _, tc := range []struct {
		err    error
		status int
	}{{practice.ErrConflict, 409}, {practice.ErrNoMistakes, 409}, {practice.ErrContentUnavailable, 503}} {
		repo.err = tc.err
		require.Equal(t, tc.status, send("POST", base, body, repo.owner, true).Code)
	}
	repo.err = nil
	require.Equal(t, 422, send("POST", base, `{"mode":"invented"}`, repo.owner, true).Code)
	require.Equal(t, 400, send("POST", base, `{"mode":"typed_recall","lessonKey":"invented"}`, repo.owner, true).Code)
	w = send("GET", base, "", repo.owner, false)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"availableMistakes":2`)
}
