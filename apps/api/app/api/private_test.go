package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizeOwnerReturns404ForDifferentOwner(t *testing.T) {
	requester := &auth.User{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111")}
	ctx := WithRequester(context.Background(), requester)

	owner := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	err := AuthorizeOwner(ctx, owner)

	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, err.GetStatus())
}

func TestAuthorizeOwnerReturnsNilForSameRequester(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	requester := &auth.User{ID: id}
	ctx := WithRequester(context.Background(), requester)

	err := AuthorizeOwner(ctx, id)

	assert.NoError(t, err)
}

func TestAuthorizeOwnerReturns404WhenUnauthenticated(t *testing.T) {
	owner := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	err := AuthorizeOwner(context.Background(), owner)

	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, err.GetStatus())
}
