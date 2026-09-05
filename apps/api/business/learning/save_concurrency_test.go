package learning

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type simultaneousSaveChecks struct {
	IdempotencyStore
	checks atomic.Int32
	ready  chan struct{}
}

func (s *simultaneousSaveChecks) ClaimSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (IdempotencyStatus, error) {
	claimer := s.IdempotencyStore.(saveIdempotencyClaimer)
	status, err := claimer.ClaimSave(ctx, userID, operation, key, fingerprint)
	if s.checks.Add(1) == 2 {
		close(s.ready)
	}
	select {
	case <-s.ready:
		return status, err
	case <-ctx.Done():
		return IdempotencyAbsent, ctx.Err()
	}
}

func (s *simultaneousSaveChecks) ReleaseSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) {
	s.IdempotencyStore.(saveIdempotencyClaimer).ReleaseSave(ctx, userID, operation, key, fingerprint)
}

func (s *simultaneousSaveChecks) CompleteSave(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) {
	s.IdempotencyStore.(saveIdempotencyClaimer).CompleteSave(ctx, userID, operation, key, fingerprint)
}

func TestConcurrentSaveSameKeyDifferentMeaningsHasOneEffect(t *testing.T) {
	repo, idem := sampleLearningRepo()
	second := repo.meanings[0]
	second.ID = uuid.New()
	repo.meanings = append(repo.meanings, second)
	gate := &simultaneousSaveChecks{IdempotencyStore: idem, ready: make(chan struct{})}
	svc := NewService(repo, gate, clock.Real{})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	userID := uuid.New()
	errs := make(chan error, 2)
	for _, meaningID := range []uuid.UUID{repo.meanings[0].ID, second.ID} {
		go func(id uuid.UUID) {
			_, err := svc.SaveUserWord(ctx, SaveUserWordRequest{UserID: userID, MeaningID: id, Source: "journey", IdempotencyKey: "same-key"})
			errs <- err
		}(meaningID)
	}
	var successes, conflicts int
	for range 2 {
		err := <-errs
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrIdempotencyConflict):
			conflicts++
		default:
			t.Fatalf("unexpected save error: %v", err)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	require.Len(t, repo.userWords, 1, "the conflicting payload must not save a second word")
}
