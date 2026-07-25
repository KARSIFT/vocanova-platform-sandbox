package api

import (
	"context"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequesterRoundTrip(t *testing.T) {
	u := &auth.User{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111")}
	ctx := WithRequester(context.Background(), u)

	assert.Equal(t, u, Requester(ctx))
	assert.Equal(t, u.ID, RequesterUserID(ctx))
}

func TestRequesterNilIsNoop(t *testing.T) {
	ctx := WithRequester(context.Background(), nil)

	assert.Nil(t, Requester(ctx))
	assert.Equal(t, uuid.Nil, RequesterUserID(ctx))
}

func TestRequesterMissing(t *testing.T) {
	assert.Nil(t, Requester(context.Background()))
	assert.Equal(t, uuid.Nil, RequesterUserID(context.Background()))
}
