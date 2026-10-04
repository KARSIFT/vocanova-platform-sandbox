package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/stories"
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

type storyAPIRepo struct {
	owner, id uuid.UUID
	calls     int
	err       error
	request   stories.StoryStartRequest
	action    stories.StoryAction
}

func (r *storyAPIRepo) state(u uuid.UUID) (*stories.State, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	if r.owner != u {
		return nil, stories.ErrNotFound
	}
	return &stories.State{ID: r.id, UserID: u, Snapshot: stories.Snapshot{Story: stories.StoryCatalogItem{Key: r.request.StoryKey}, ContentVersion: stories.ContentVersion, GradingVersion: stories.GradingVersion}}, nil
}
func (r *storyAPIRepo) Active(_ context.Context, u uuid.UUID) error { _, e := r.state(u); return e }
func (r *storyAPIRepo) List(_ context.Context, u uuid.UUID) ([]stories.State, error) {
	s, e := r.state(u)
	if e != nil {
		return nil, e
	}
	return []stories.State{*s}, nil
}
func (r *storyAPIRepo) Start(_ context.Context, u uuid.UUID, req stories.StoryStartRequest, _ string, _ time.Time) (*stories.State, error) {
	r.request = req
	return r.state(u)
}
func (r *storyAPIRepo) Get(_ context.Context, u, id uuid.UUID) (*stories.State, error) {
	if id != r.id {
		return nil, stories.ErrNotFound
	}
	return r.state(u)
}
func (r *storyAPIRepo) Act(ctx context.Context, u, id uuid.UUID, a stories.StoryAction, _ string, _ time.Time) (*stories.State, error) {
	r.action = a
	return r.Get(ctx, u, id)
}
func TestStoriesAPIAuthCSRFContractAndOwnership(t *testing.T) {
	repo := &storyAPIRepo{owner: uuid.New(), id: uuid.New()}
	a := authStubService()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("stories", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(a))
	RegisterStories(api, stories.NewService(repo, nil), a)
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
	base := "/api/v1/story-sessions"
	single := base + "/" + repo.id.String()
	body := `{"storyKey":"a-quiet-lunch"}`
	for _, route := range []struct{ method, path string }{{"GET", "/api/v1/stories"}, {"GET", "/api/v1/stories/a-quiet-lunch"}, {"POST", base}, {"GET", single}, {"POST", single + "/actions"}} {
		require.Equal(t, 401, send(route.method, route.path, body, uuid.Nil, false).Code)
	}
	require.Zero(t, repo.calls)
	require.Equal(t, 403, send("POST", base, body, repo.owner, false).Code)
	require.Zero(t, repo.calls)
	w := send("POST", base, body, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "a-quiet-lunch", repo.request.StoryKey)
	require.Equal(t, 404, send("GET", single, "", uuid.New(), false).Code)
	action := `{"stepId":"story-step","expectedRevision":2,"clientActionId":"answer-one","action":"answer","choiceId":"choice-1"}`
	require.Equal(t, 403, send("POST", single+"/actions", action, repo.owner, false).Code)
	w = send("POST", single+"/actions", action, repo.owner, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, "choice-1", repo.action.ChoiceID)
	require.Equal(t, 2, repo.action.ExpectedRevision)
	w = send("GET", "/api/v1/stories?userId="+uuid.NewString(), "", repo.owner, false)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"key":"a-quiet-lunch"`)
	require.NotContains(t, w.Body.String(), "snapshot")
	require.NotContains(t, w.Body.String(), "CorrectChoice")
	w = send("GET", "/api/v1/stories/a-quiet-lunch", "", repo.owner, false)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"lines":[`)
	require.NotContains(t, w.Body.String(), "correctChoice")
	require.Equal(t, 404, send("GET", "/api/v1/stories/unknown", "", repo.owner, false).Code)
	for _, tc := range []struct {
		err    error
		status int
	}{{stories.ErrInvalid, 400}, {stories.ErrConflict, 409}, {stories.ErrContentUnavailable, 503}, {errors.New("private connection detail"), 500}} {
		repo.err = tc.err
		w = send("GET", single, "", repo.owner, false)
		require.Equal(t, tc.status, w.Code)
		require.NotContains(t, w.Body.String(), "private connection detail")
	}
}
