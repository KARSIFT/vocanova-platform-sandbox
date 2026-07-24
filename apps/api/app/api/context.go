package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type humaContextKey struct{}

// withHumaContext stores the huma.Context in the request context so handlers
// can access the underlying request and response writer.
func withHumaContext(ctx huma.Context, next func(huma.Context)) {
	next(huma.WithValue(ctx, humaContextKey{}, ctx))
}

// HumaContext retrieves the huma.Context from a handler's context.Context.
func HumaContext(ctx context.Context) huma.Context {
	if c, ok := ctx.Value(humaContextKey{}).(huma.Context); ok {
		return c
	}
	return nil
}
