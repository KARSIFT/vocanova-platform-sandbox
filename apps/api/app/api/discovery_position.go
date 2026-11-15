package api

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/discoveryposition"
	"github.com/danielgtaylor/huma/v2"
)

type DiscoveryPositionBody struct {
	Position *discoveryposition.Position `json:"position"`
}

func (DiscoveryPositionBody) Schema(registry huma.Registry) *huma.Schema {
	type alias DiscoveryPositionBody
	schema := huma.SchemaFromType(registry, reflect.TypeOf(alias{}))
	schema.Properties["position"] = &huma.Schema{AnyOf: []*huma.Schema{schema.Properties["position"], {Type: "null"}}}
	return schema
}

type DiscoveryPositionOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         DiscoveryPositionBody
}
type PutDiscoveryPositionInput struct{ Body discoveryposition.Update }

func RegisterDiscoveryPosition(api huma.API, svc *discoveryposition.Service, authSvc *auth.Service) {
	private := func(ctx huma.Context, next func(huma.Context)) { ctx.SetHeader("Cache-Control", "no-store"); next(ctx) }
	huma.Register(api, huma.Operation{OperationID: "GetDiscoveryPosition", Method: http.MethodGet, Path: "/api/v1/discovery-position", Summary: "Read your optional discovery position", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth()}, Errors: []int{401, 404, 500}}, func(ctx context.Context, _ *struct{}) (*DiscoveryPositionOutput, error) {
		p, err := svc.Get(ctx, RequesterUserID(ctx))
		if err != nil {
			return nil, mapDiscoveryPositionError(err)
		}
		return &DiscoveryPositionOutput{CacheControl: "no-store", Body: DiscoveryPositionBody{Position: p}}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "PutDiscoveryPosition", Method: http.MethodPut, Path: "/api/v1/discovery-position", Summary: "Save your optional discovery position", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){private, RequireAuth(), CSRFMiddleware(authSvc)}, Errors: []int{400, 401, 403, 404, 409, 422, 500}}, func(ctx context.Context, in *PutDiscoveryPositionInput) (*DiscoveryPositionOutput, error) {
		p, err := svc.Put(ctx, RequesterUserID(ctx), in.Body)
		if err != nil {
			return nil, mapDiscoveryPositionError(err)
		}
		return &DiscoveryPositionOutput{CacheControl: "no-store", Body: DiscoveryPositionBody{Position: p}}, nil
	})
}
func mapDiscoveryPositionError(err error) error {
	switch {
	case errors.Is(err, discoveryposition.ErrInvalid):
		return huma.Error400BadRequest("Invalid discovery position")
	case errors.Is(err, discoveryposition.ErrNotFound):
		return huma.Error404NotFound("Learner not found")
	case errors.Is(err, discoveryposition.ErrConflict):
		return huma.Error409Conflict("Discovery position changed; load your saved position")
	default:
		return huma.Error500InternalServerError("Discovery position unavailable")
	}
}
