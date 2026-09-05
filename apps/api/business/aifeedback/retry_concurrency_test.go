package aifeedback

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepositoryCreateRetryAttemptReturnsExtantActiveGeneration(t *testing.T) {
	repo := NewMemoryRepository(MemoryRepositoryData{})
	failed := &StoredFeedbackAttempt{
		ID:                uuid.New(),
		LearnerSentenceID: uuid.New(),
		Status:            AttemptStatusFailed,
		RequestHash:       "concurrent-retry-hash",
	}
	repo.attempts = append(repo.attempts, MemoryAIFeedbackAttempt{
		ID: failed.ID, LearnerSentenceID: failed.LearnerSentenceID, Status: AttemptStatusFailed, RequestHash: failed.RequestHash,
	})

	start := make(chan struct{})
	results := make(chan *RetryAttempt, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			retry, err := repo.CreateRetryAttempt(t.Context(), failed, ProviderMock, "mock", time.Now().UTC())
			results <- retry
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	var created, extant int
	for retry := range results {
		if retry.Pending != nil {
			created++
		} else {
			require.NotNil(t, retry.Existing)
			assert.Equal(t, AttemptStatusPending, retry.Existing.Status)
			extant++
		}
	}
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, extant)
}
