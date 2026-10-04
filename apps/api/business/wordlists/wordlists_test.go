package wordlists

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

type captureRepository struct {
	calls   int
	request WriteRequest
}

func (r *captureRepository) List(context.Context, uuid.UUID) (*WordListsResponse, error) {
	return &WordListsResponse{}, nil
}
func (r *captureRepository) Get(context.Context, uuid.UUID, uuid.UUID) (*WordListDetail, error) {
	return &WordListDetail{}, nil
}
func (r *captureRepository) Write(_ context.Context, req WriteRequest, _ time.Time) (*WordListDetail, error) {
	r.calls++
	r.request = req
	return &WordListDetail{}, nil
}
func TestListWriteValidationAndNormalizedFingerprint(t *testing.T) {
	repo := &captureRepository{}
	svc := NewService(repo, nil)
	base := WriteRequest{UserID: uuid.New(), ListID: uuid.New(), Operation: "put", Name: " Travel ", IdempotencyKey: "create"}
	_, e := svc.Write(t.Context(), base)
	require.NoError(t, e)
	require.Equal(t, "Travel", repo.request.Name)
	for _, modify := range []func(*WriteRequest){func(r *WriteRequest) { r.UserID = uuid.Nil }, func(r *WriteRequest) { r.ListID = uuid.Nil }, func(r *WriteRequest) { r.Name = " " }, func(r *WriteRequest) { r.Name = strings.Repeat("界", 81) }, func(r *WriteRequest) { r.Name = "line\nnext" }, func(r *WriteRequest) { r.IdempotencyKey = " " }, func(r *WriteRequest) { r.IdempotencyKey = "bad\x00key" }, func(r *WriteRequest) { r.ExpectedRevision = -1 }, func(r *WriteRequest) { r.Operation = "unknown" }, func(r *WriteRequest) { r.Operation = "add"; r.Name = "" }, func(r *WriteRequest) { r.Operation = "remove"; r.MeaningID = uuid.New() }, func(r *WriteRequest) { r.Operation = "delete" }} {
		r := base
		modify(&r)
		_, e := svc.Write(t.Context(), r)
		require.ErrorIs(t, e, ErrInvalid)
	}
	require.Equal(t, 1, repo.calls)
	changed := repo.request
	changed.ExpectedRevision++
	require.NotEqual(t, fingerprint(repo.request), fingerprint(changed))
	changed = repo.request
	changed.Operation = "delete"
	require.NotEqual(t, fingerprint(repo.request), fingerprint(changed))
}
