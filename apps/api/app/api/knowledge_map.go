package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/danielgtaylor/huma/v2"
)

type KnowledgeMapInput struct {
	Band int `query:"band" default:"1" minimum:"1" maximum:"100"`
}
type KnowledgeMapOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         knowledgecatalog.Map
}
type KnowledgeMapDecisionInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           struct {
		MeaningID        string `json:"meaningId" format:"uuid" required:"true"`
		Decision         string `json:"decision" enum:"learn,known" required:"true"`
		ExpectedRevision string `json:"expectedRevision" minLength:"64" maxLength:"64" required:"true"`
	}
}
type KnowledgeMapDecisionOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         knowledgecatalog.DecisionResult
}

func mapKnowledgeCatalogError(err error) error {
	switch {
	case errors.Is(err, knowledgecatalog.ErrInvalid):
		return huma.Error400BadRequest("Invalid knowledge map request")
	case errors.Is(err, knowledgecatalog.ErrNotFound):
		return huma.Error404NotFound("Word or learner not found")
	case errors.Is(err, knowledgecatalog.ErrConflict):
		return huma.Error409Conflict("This action key was already used; refresh before choosing again")
	case errors.Is(err, knowledgecatalog.ErrStale):
		return huma.Error409Conflict("This word changed. Refresh and choose again.")
	default:
		return huma.Error500InternalServerError("Could not access your knowledge map")
	}
}
func RegisterKnowledgeMap(api huma.API, svc *knowledgecatalog.Service, authSvc *auth.Service) {
	private := func(ctx huma.Context, next func(huma.Context)) { ctx.SetHeader("Cache-Control", "no-store"); next(ctx) }
	huma.Register(api, huma.Operation{OperationID: "GetKnowledgeMap", Method: http.MethodGet, Path: "/api/v1/knowledge-map", Summary: "Explore the sourced 10000-word map in bands of 100", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth()}}, func(ctx context.Context, in *KnowledgeMapInput) (*KnowledgeMapOutput, error) {
		state, err := svc.Read(ctx, RequesterUserID(ctx), in.Band)
		if err != nil {
			return nil, mapKnowledgeCatalogError(err)
		}
		return &KnowledgeMapOutput{CacheControl: "no-store", Body: *state}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "DecideKnowledgeMapWord", Method: http.MethodPost, Path: "/api/v1/knowledge-map/decisions", Summary: "Choose Learn or Already known, preserving private notes and review history", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *KnowledgeMapDecisionInput) (*KnowledgeMapDecisionOutput, error) {
		state, err := svc.Decide(ctx, knowledgecatalog.DecisionRequest{UserID: RequesterUserID(ctx), MeaningID: parseUUID(in.Body.MeaningID), Decision: in.Body.Decision, ExpectedRevision: in.Body.ExpectedRevision, IdempotencyKey: in.IdempotencyKey})
		if err != nil {
			return nil, mapKnowledgeCatalogError(err)
		}
		return &KnowledgeMapDecisionOutput{CacheControl: "no-store", Body: *state}, nil
	})
}
