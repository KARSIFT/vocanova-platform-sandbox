package aifeedback

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Hold both absent-key reads before either report can be persisted.
type reportCheckBarrier struct {
	learning.IdempotencyStore
	mu    sync.Mutex
	reads int
	both  chan struct{}
}

func (s *reportCheckBarrier) Check(ctx context.Context, userID uuid.UUID, operation, key, fingerprint string) (learning.IdempotencyStatus, error) {
	status, err := s.IdempotencyStore.Check(ctx, userID, operation, key, fingerprint)
	if err != nil {
		return status, err
	}
	s.mu.Lock()
	s.reads++
	if s.reads == 2 {
		close(s.both)
	}
	s.mu.Unlock()
	select {
	case <-s.both:
		return status, nil
	case <-ctx.Done():
		return status, ctx.Err()
	}
}

func TestReportConcurrentSameKeyDifferentAttemptsConflicts(t *testing.T) {
	f := newServiceFixture(t)
	first, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
	require.NoError(t, err)
	second, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work hard every day."))
	require.NoError(t, err)
	require.NotEqual(t, first.AttemptID, second.AttemptID)
	f.service.idem = &reportCheckBarrier{IdempotencyStore: f.service.idem, both: make(chan struct{})}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	results := make(chan error, 2)
	for _, attempt := range []uuid.UUID{first.AttemptID, second.AttemptID} {
		go func(id uuid.UUID) {
			results <- f.service.ReportFeedback(ctx, f.userID, id, ReportReasonAlreadyCorrect, "shared-report-key")
		}(attempt)
	}
	successes, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			successes++
		} else if assert.ErrorIs(t, err, ErrReportIdempotencyConflict) {
			conflicts++
		}
	}
	assert.Equal(t, 1, successes)
	assert.Equal(t, 1, conflicts)
	assert.Len(t, f.repo.QualityReviewReports(), 1)
}
