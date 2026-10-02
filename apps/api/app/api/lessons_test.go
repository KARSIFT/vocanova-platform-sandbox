package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type lessonAPIRepo struct {
	owner, id uuid.UUID
	calls     int
	err       error
	action    lessons.Action
	key       string
}

func (r *lessonAPIRepo) state(u uuid.UUID) (*lessons.State, error) {
	if r.err != nil {
		return nil, r.err
	}
	if u != r.owner {
		return nil, lessons.ErrNotFound
	}
	return &lessons.State{ID: r.id, UserID: u, Snapshot: lessons.Snapshot{Definition: lessons.Definition{Key: "airport", Version: "1", Title: "Find your flight", SituationSlug: "airport"}, Words: []lessons.Word{}}}, nil
}
func (r *lessonAPIRepo) List(_ context.Context, u uuid.UUID) ([]lessons.State, error) {
	r.calls++
	st, e := r.state(u)
	if e != nil {
		return nil, e
	}
	return []lessons.State{*st}, nil
}
func (r *lessonAPIRepo) Start(_ context.Context, u uuid.UUID, _ lessons.Definition, key string, _ time.Time) (*lessons.State, error) {
	r.calls++
	r.key = key
	return r.state(u)
}
func (r *lessonAPIRepo) Get(_ context.Context, u, id uuid.UUID) (*lessons.State, error) {
	r.calls++
	if id != r.id {
		return nil, lessons.ErrNotFound
	}
	return r.state(u)
}
func (r *lessonAPIRepo) Act(_ context.Context, u, id uuid.UUID, a lessons.Action, key string, _ time.Time) (*lessons.State, error) {
	r.calls++
	r.action = a
	r.key = key
	if id != r.id {
		return nil, lessons.ErrNotFound
	}
	return r.state(u)
}
func lessonTestAPI(t *testing.T) (huma.API, *lessonAPIRepo, *auth.Service) {
	t.Helper()
	repo := &lessonAPIRepo{owner: uuid.New(), id: uuid.New()}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("Lessons", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterLessons(api, lessons.NewService(repo, nil), a)
	return api, repo, a
}
func TestLessonAPIProtectsReadsAndMutations(t *testing.T) {
	for _, route := range []struct{ method, path string }{{"GET", "/api/v1/lessons"}, {"GET", "/api/v1/lesson-sessions/00000000-0000-0000-0000-000000000001"}, {"POST", "/api/v1/lessons/airport/sessions"}, {"POST", "/api/v1/lesson-sessions/00000000-0000-0000-0000-000000000001/actions"}} {
		t.Run(route.method+route.path, func(t *testing.T) {
			api, repo, _ := lessonTestAPI(t)
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
			require.Equal(t, http.StatusUnauthorized, w.Code)
			require.Zero(t, repo.calls)
		})
	}
	api, repo, _ := lessonTestAPI(t)
	req := httptest.NewRequest("POST", "/api/v1/lessons/airport/sessions", nil)
	req.Header.Set("Idempotency-Key", "start")
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: repo.owner}))
	w := httptest.NewRecorder()
	api.Adapter().ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Zero(t, repo.calls)
}
func TestLessonAPIStartActionAndOwnershipContract(t *testing.T) {
	api, repo, a := lessonTestAPI(t)
	send := func(method, path, body, key string, user uuid.UUID) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Idempotency-Key", key)
		req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		addCSRF(req, a)
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	started := send("POST", "/api/v1/lessons/airport/sessions", "", "start", repo.owner)
	require.Equal(t, 200, started.Code, started.Body.String())
	require.Equal(t, "start", repo.key)
	var session lessons.Session
	require.NoError(t, json.Unmarshal(started.Body.Bytes(), &session))
	require.Equal(t, repo.id.String(), session.ID)
	require.Equal(t, "1", session.LessonVersion)
	route := "/api/v1/lesson-sessions/" + repo.id.String()
	require.Equal(t, 404, send("GET", route, "", "", uuid.New()).Code)
	body := `{"stepId":"recall-1","expectedRevision":4,"clientActionId":"action-id","action":"answer","choiceId":"choice-a"}`
	result := send("POST", route+"/actions", body, "action-id", repo.owner)
	require.Equal(t, 200, result.Code, result.Body.String())
	require.Equal(t, lessons.Action{StepID: "recall-1", ExpectedRevision: 4, ClientActionID: "action-id", Action: "answer", ChoiceID: "choice-a"}, repo.action)
	repo.err = lessons.ErrConflict
	require.Equal(t, 409, send("POST", route+"/actions", body, "action-id", repo.owner).Code)
	repo.err = lessons.ErrContentUnavailable
	require.Equal(t, 503, send("POST", "/api/v1/lessons/airport/sessions", "", "start2", repo.owner).Code)
	repo.err = nil
	require.Equal(t, 404, send("POST", "/api/v1/lessons/not-a-lesson/sessions", "", "bad", repo.owner).Code)
	for _, invalid := range []string{`{"stepId":"teach-1","expectedRevision":-1,"clientActionId":"a","action":"continue"}`, `{"stepId":"teach-1","expectedRevision":0,"clientActionId":"a","action":"complete"}`, `{"stepId":"teach-1","expectedRevision":0,"clientActionId":"","action":"continue"}`} {
		before := repo.calls
		w := send("POST", route+"/actions", invalid, "invalid", repo.owner)
		require.Equal(t, 422, w.Code, w.Body.String())
		require.Equal(t, before, repo.calls)
	}
}
