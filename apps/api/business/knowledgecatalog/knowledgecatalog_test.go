package knowledgecatalog

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fixtureRepository struct{ calls int }

func (r *fixtureRepository) Read(context.Context, uuid.UUID, int, time.Time) (*Map, error) {
	r.calls++
	return &Map{}, nil
}
func (r *fixtureRepository) Decide(context.Context, DecisionRequest, time.Time) (*DecisionResult, error) {
	r.calls++
	return &DecisionResult{}, nil
}
func TestServiceRejectsInvalidBandsAndDecisionInputs(t *testing.T) {
	repo := &fixtureRepository{}
	svc := NewService(repo, nil)
	u, m := uuid.New(), uuid.New()
	for _, band := range []int{-1, 0, 101, 10000} {
		_, err := svc.Read(t.Context(), u, band)
		require.ErrorIs(t, err, ErrInvalid)
	}
	_, err := svc.Read(t.Context(), uuid.Nil, 1)
	require.ErrorIs(t, err, ErrInvalid)
	valid := DecisionRequest{UserID: u, MeaningID: m, Decision: "learn", ExpectedRevision: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", IdempotencyKey: "one"}
	for _, change := range []func(*DecisionRequest){func(r *DecisionRequest) { r.UserID = uuid.Nil }, func(r *DecisionRequest) { r.MeaningID = uuid.Nil }, func(r *DecisionRequest) { r.Decision = "mastered" }, func(r *DecisionRequest) { r.ExpectedRevision = "" }, func(r *DecisionRequest) {
		r.ExpectedRevision = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	}, func(r *DecisionRequest) { r.IdempotencyKey = " " }, func(r *DecisionRequest) { r.IdempotencyKey = "bad\x00key" }} {
		req := valid
		change(&req)
		_, err = svc.Decide(t.Context(), req)
		require.ErrorIs(t, err, ErrInvalid)
	}
	require.Zero(t, repo.calls)
	_, err = svc.Read(t.Context(), u, 100)
	require.NoError(t, err)
	_, err = svc.Decide(t.Context(), valid)
	require.NoError(t, err)
	require.Equal(t, 2, repo.calls)
}
func TestDecisionFingerprintIncludesOperationMeaningChoiceAndRevision(t *testing.T) {
	req := DecisionRequest{UserID: uuid.New(), MeaningID: uuid.New(), Decision: "learn", ExpectedRevision: "revision", IdempotencyKey: "key"}
	want := fingerprint(req)
	copy := req
	copy.IdempotencyKey = "different-key"
	require.Equal(t, want, fingerprint(copy))
	copy = req
	copy.ExpectedRevision = "other"
	require.NotEqual(t, want, fingerprint(copy))
	copy = req
	copy.Decision = "known"
	require.NotEqual(t, want, fingerprint(copy))
	copy = req
	copy.MeaningID = uuid.New()
	require.NotEqual(t, want, fingerprint(copy))
}
