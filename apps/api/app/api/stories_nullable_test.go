package api

import (
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/stories"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestStorySessionSchemaAllowsActualNullStepAndFeedback(t *testing.T) {
	r := &storyAPIRepo{owner: uuid.New(), id: uuid.New()}
	a := humachi.New(chi.NewMux(), huma.DefaultConfig("story-null", "test"))
	as := authStubService()
	a.UseMiddleware(withHumaContext)
	a.UseMiddleware(AuthMiddleware(as))
	RegisterStories(a, stories.NewService(r, nil), as)
	req := httptest.NewRequest("GET", "/api/v1/story-sessions"+"/"+r.id.String(), nil)
	req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: r.owner}))
	w := httptest.NewRecorder()
	a.Adapter().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, w.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	properties := a.OpenAPI().Components.Schemas.Map()["StorySession"].Properties
	for _, field := range []string{"currentStep", "feedback"} {
		require.Contains(t, body, field)
		require.Nil(t, body[field])
		require.True(t, a.OpenAPI().Components.Schemas.SchemaFromRef(properties[field].Ref).Nullable, "actual null %s must match its referenced response schema", field)
	}
}
