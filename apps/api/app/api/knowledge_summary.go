package api

import (
	"context"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/danielgtaylor/huma/v2"
)

type KnowledgeSummaryDTO struct {
	SelfReportedKnown int `json:"selfReportedKnown" doc:"All explicit known-meaning self-assessments; overlaps saved stages and is not a proficiency estimate"`
	Saved             int `json:"saved" doc:"All requester-owned undeleted saved meanings, across every page and status"`
	New               int `json:"new" doc:"Saved meanings in new state"`
	Learning          int `json:"learning" doc:"Saved meanings in learning state"`
	Reviewing         int `json:"reviewing" doc:"Saved meanings in reviewing state"`
	Mastered          int `json:"mastered" doc:"Saved meanings marked mastered by the scheduler; not a proficiency estimate"`
	Ignored           int `json:"ignored" doc:"Saved meanings in ignored state"`
	Archived          int `json:"archived" doc:"Saved meanings in archived state"`
	Due               int `json:"due" doc:"Due meanings eligible for review before any daily mission cap; overlaps new, learning and reviewing"`
}

type GetKnowledgeSummaryOutput struct{ Body KnowledgeSummaryDTO }

func registerKnowledgeSummary(api huma.API, svc *learning.Service) {
	huma.Register(api, huma.Operation{OperationID: "GetKnowledgeSummary", Method: http.MethodGet, Path: "/api/v1/knowledge-summary", Summary: "Count the requester's saved meanings by learning state", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}, Responses: map[string]*huma.Response{"401": {Description: "Authentication is required"}}}, func(ctx context.Context, input *struct{}) (*GetKnowledgeSummaryOutput, error) {
		summary, err := svc.GetKnowledgeSummary(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapLearningError(err)
		}
		return &GetKnowledgeSummaryOutput{Body: KnowledgeSummaryDTO{Saved: summary.Saved, SelfReportedKnown: summary.SelfReportedKnown, New: summary.New, Learning: summary.Learning, Reviewing: summary.Reviewing, Mastered: summary.Mastered, Ignored: summary.Ignored, Archived: summary.Archived, Due: summary.Due}}, nil
	})
}
