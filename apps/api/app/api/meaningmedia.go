package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/danielgtaylor/huma/v2"
)

type MeaningImagesInput struct {
	MeaningIDs string `query:"meaningIds" required:"true" maxLength:"3699" doc:"Up to 100 comma-separated canonical meaning UUIDs"`
}
type MeaningImagesOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         struct {
		Items []meaningmedia.Image `json:"items"`
	}
}

func RegisterMeaningImages(api huma.API, svc *meaningmedia.Service) {
	huma.Register(api, huma.Operation{OperationID: "GetMeaningImages", Method: http.MethodGet, Path: "/api/v1/meaning-images", Summary: "Find approved illustrations for a bounded set of exact meanings", Tags: []string{"Discovery"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(ctx context.Context, in *MeaningImagesInput) (*MeaningImagesOutput, error) {
		images, err := svc.Read(ctx, strings.Split(in.MeaningIDs, ","))
		if errors.Is(err, meaningmedia.ErrInvalid) {
			return nil, huma.Error400BadRequest("Choose up to 100 canonical meanings.")
		}
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Word pictures are temporarily unavailable.")
		}
		out := &MeaningImagesOutput{CacheControl: "private, max-age=60"}
		out.Body.Items = images
		return out, nil
	})
}
