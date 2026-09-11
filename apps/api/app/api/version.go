package api

import (
	"context"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/buildinfo"
	"github.com/danielgtaylor/huma/v2"
)

func RegisterVersion(api huma.API) {
	type output struct {
		CacheControl string `header:"Cache-Control"`
		Body         buildinfo.Info
	}
	huma.Register(api, huma.Operation{
		OperationID: "GetBuildVersion",
		Method:      http.MethodGet,
		Path:        "/version",
		Summary:     "Identify the deployed API build",
		Tags:        []string{"Operations"},
	}, func(context.Context, *struct{}) (*output, error) {
		return &output{CacheControl: "no-store", Body: buildinfo.Current()}, nil
	})
}
