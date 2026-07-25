package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// AuthorizeOwner returns a 404 status error when the requester is not the owner
// of a private resource. This implements the deny-by-default pattern for
// requester-scoped private resources: inaccessible resources are
// indistinguishable from missing ones to avoid account or resource enumeration.
//
// Handlers should call this immediately after loading a private resource by its
// identifier. If it returns nil, the handler may return the resource.
func AuthorizeOwner(ctx context.Context, resourceOwnerID uuid.UUID) huma.StatusError {
	if RequesterUserID(ctx) != resourceOwnerID {
		return huma.Error404NotFound("resource not found")
	}
	return nil
}
