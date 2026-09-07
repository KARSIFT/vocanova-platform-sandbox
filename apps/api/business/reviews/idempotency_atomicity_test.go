package reviews

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Pause both service prechecks after they read the same absent key. Repository
// serialization must still prevent the losing payload from changing a schedule.
type simultaneousReviewChecks struct {
	*learning.MemoryIdempotencyStore
	arrived sync.WaitGroup
}

func TestMemoryReviewKeyExpiryAndRollback(t *testing.T) {
	userID, meaningID, userWordID := uuid.New(), uuid.New(), uuid.New()
	repo := NewMemoryRepository(MemoryRepositoryData{UserWords: []MemoryUserWord{{
		ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new",
	}}})
	c := &clock.Fixed{T: time.Now().UTC()}
	svc := NewService(repo, nil, c)
	base := SubmitReviewRequest{
		UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		PromptType: PromptTypeSelfCheck, Result: ResultCorrect, Rating: RatingGood,
		AnsweredAt: c.T, IdempotencyKey: "key", ClientAttemptID: "first",
	}
	invalid := base
	invalid.MeaningID = uuid.New()
	_, err := svc.SubmitReview(t.Context(), invalid)
	require.ErrorIs(t, err, ErrUserWordNotFound)
	require.Empty(t, repo.keys, "failed effect must not claim a key")
	first, err := svc.SubmitReview(t.Context(), base)
	require.NoError(t, err)
	created := c.T
	c.T = created.Add(24*time.Hour - time.Nanosecond)
	replayed, err := svc.SubmitReview(t.Context(), base)
	require.NoError(t, err)
	require.Equal(t, first.ID, replayed.ID)
	changed := base
	changed.ClientAttemptID = "second"
	_, err = svc.SubmitReview(t.Context(), changed)
	require.ErrorIs(t, err, ErrIdempotencyConflict)
	c.T = created.Add(24 * time.Hour)
	_, err = svc.SubmitReview(t.Context(), changed)
	require.NoError(t, err, "replay must not extend the original lifetime")
	require.Len(t, repo.attempts, 2)
	base.IdempotencyKey = "fresh-header"
	replayed, err = svc.SubmitReview(t.Context(), base)
	require.NoError(t, err)
	require.Equal(t, first.ID, replayed.ID)
	require.Len(t, repo.attempts, 2, "client attempt guard outlives header key")
}

func (s *simultaneousReviewChecks) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (learning.IdempotencyStatus, error) {
	status, err := s.MemoryIdempotencyStore.Check(ctx, userID, operation, key, fingerprint)
	s.arrived.Done()
	s.arrived.Wait()
	return status, err
}

func TestConcurrentReviewKeyRejectsDifferentAttemptBeforeEffects(t *testing.T) {
	userID, meaningID, userWordID := uuid.New(), uuid.New(), uuid.New()
	repo := NewMemoryRepository(MemoryRepositoryData{UserWords: []MemoryUserWord{{
		ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new",
	}}})
	store := &simultaneousReviewChecks{MemoryIdempotencyStore: learning.NewMemoryIdempotencyStore()}
	store.arrived.Add(2)
	svc := NewService(repo, store, nil)
	base := SubmitReviewRequest{
		UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		PromptType: PromptTypeSelfCheck, Result: ResultCorrect, Rating: RatingGood,
		AnsweredAt: time.Now().UTC(), IdempotencyKey: "same-key",
	}
	results := make(chan error, 2)
	for _, attemptID := range []string{"first-client-attempt", "second-client-attempt"} {
		go func() {
			req := base
			req.ClientAttemptID = attemptID
			_, err := svc.SubmitReview(t.Context(), req)
			results <- err
		}()
	}
	var succeeded, conflicted int
	for range 2 {
		err := <-results
		if err == nil {
			succeeded++
		} else {
			require.ErrorIs(t, err, ErrIdempotencyConflict)
			conflicted++
		}
	}
	require.Equal(t, 1, succeeded, "a reused key cannot advance the schedule twice")
	require.Equal(t, 1, conflicted)
	require.Len(t, repo.attempts, 1)
	require.Equal(t, 1, repo.userWords[0].TotalReviewCount)
}
