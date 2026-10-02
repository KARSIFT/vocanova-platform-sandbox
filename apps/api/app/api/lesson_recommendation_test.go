package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type recommendationAPIRepo struct {
	owner uuid.UUID
	calls int
	err   error
	empty bool
}

func (r *recommendationAPIRepo) ReadRecommendation(_ context.Context, user uuid.UUID) (lessons.RecommendationData, error) {
	r.owner = user
	r.calls++
	return lessons.RecommendationData{Available: map[string]bool{"conversation-start": !r.empty}}, r.err
}

func TestLessonRecommendationAPIAuthenticationOwnershipAndFailure(t *testing.T) {
	r := &recommendationAPIRepo{}
	a := humachi.New(chi.NewMux(), huma.DefaultConfig("Recommendations", "test"))
	a.UseMiddleware(withHumaContext)
	RegisterLessonRecommendation(a, lessons.NewRecommendationService(r))
	request := func(user uuid.UUID) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/lesson-recommendation?userId="+uuid.NewString(), nil)
		if user != uuid.Nil {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: user}))
		}
		w := httptest.NewRecorder()
		a.Adapter().ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, request(uuid.Nil).Code)
	require.Zero(t, r.calls)
	u := uuid.New()
	w := request(u)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, u, r.owner, "query user ID never selects someone else's data")
	require.Contains(t, w.Body.String(), `"usefulTargetCount":3`)
	require.NotContains(t, w.Body.String(), "correctChoiceId")
	require.NotContains(t, w.Body.String(), "snapshot")
	r.empty = true
	w = request(u)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), `"recommendation":null`)
	require.Contains(t, w.Body.String(), `"status":"content_unavailable"`)
	property := a.OpenAPI().Components.Schemas.Map()["LessonRecommendationBody"].Properties["recommendation"]
	require.Len(t, property.AnyOf, 2)
	require.Equal(t, "null", property.AnyOf[1].Type)
	r.err = errors.New("private database detail")
	w = request(u)
	require.Equal(t, 503, w.Code)
	require.NotContains(t, w.Body.String(), "private database detail")
	require.NotContains(t, w.Body.String(), `"usefulTargetCount":0`)
	r.err = lessons.ErrNotFound
	require.Equal(t, 404, request(u).Code)
}
