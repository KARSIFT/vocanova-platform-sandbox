package api

import (
	"context"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/google/uuid"
)

type requesterKey struct{}

// WithRequester stores the authenticated user in the context. The requester
// is set by AuthMiddleware and consumed by handlers and requester-scoped
// services.
func WithRequester(ctx context.Context, u *auth.User) context.Context {
	if u == nil {
		return ctx
	}
	return context.WithValue(ctx, requesterKey{}, u)
}

// Requester returns the authenticated user from the context, or nil when the
// caller is unauthenticated.
func Requester(ctx context.Context) *auth.User {
	if u, ok := ctx.Value(requesterKey{}).(*auth.User); ok {
		return u
	}
	return nil
}

// RequesterUserID returns the authenticated user ID or uuid.Nil.
func RequesterUserID(ctx context.Context) uuid.UUID {
	if u := Requester(ctx); u != nil {
		return u.ID
	}
	return uuid.Nil
}
