package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/achievements"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type ListAchievementsOutput struct{ Body achievements.AchievementList }

func RegisterAchievements(api huma.API, svc *achievements.Service) {
	huma.Register(api, huma.Operation{OperationID: "ListAchievements", Method: http.MethodGet, Path: "/api/v1/achievements", Summary: "Learning milestones derived from your confirmed history; no reward mutation", Tags: []string{"Progress"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(ctx context.Context, _ *struct{}) (*ListAchievementsOutput, error) {
		out, err := svc.List(ctx, RequesterUserID(ctx))
		if errors.Is(err, achievements.ErrNotFound) {
			return nil, huma.Error404NotFound("Learner not found")
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("We could not load your achievements. Try again.")
		}
		return &ListAchievementsOutput{Body: *out}, nil
	})
}
