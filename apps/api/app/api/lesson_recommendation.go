package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/lessons"
	"github.com/danielgtaylor/huma/v2"
)

type LessonRecommendationBody lessons.RecommendationResult

// Huma's nullable tag supports scalars only. Describe the actual object-or-null
// response explicitly without making a missing recommendation look like zeros.
func (LessonRecommendationBody) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	schema.Properties["recommendation"] = &huma.Schema{AnyOf: []*huma.Schema{schema.Properties["recommendation"], {Type: "null"}}}
	return schema
}

type LessonRecommendationOutput struct{ Body LessonRecommendationBody }

// RegisterLessonRecommendation only reads requester-owned state; GET awards nothing.
func RegisterLessonRecommendation(api huma.API, svc *lessons.RecommendationService) {
	huma.Register(api, huma.Operation{
		OperationID: "GetLessonRecommendation", Method: http.MethodGet, Path: "/api/v1/lesson-recommendation",
		Summary: "Recommend an unfinished lesson using personal target coverage", Tags: []string{"Lessons"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses:   map[string]*huma.Response{"401": {Description: "Authentication required"}, "404": {Description: "Active learner not found"}, "503": {Description: "Recommendation unavailable"}},
	}, func(ctx context.Context, _ *struct{}) (*LessonRecommendationOutput, error) {
		result, err := svc.Get(ctx, RequesterUserID(ctx))
		if errors.Is(err, lessons.ErrNotFound) {
			return nil, huma.Error404NotFound("Active learner not found")
		}
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("We could not load your next lesson. Please try again.")
		}
		return &LessonRecommendationOutput{Body: LessonRecommendationBody(result)}, nil
	})
}
