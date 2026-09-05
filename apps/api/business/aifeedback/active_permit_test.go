package aifeedback

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeniedGenerationCannotReleaseAnotherRequestsPermit(t *testing.T) {
	f := newServiceFixture(t)
	// Hold the same real limiter permit a running generation owns. Rejected
	// requests must not release it, even when multiple requests arrive.
	require.NoError(t, f.service.rateLimiter.Allow(t.Context(), f.userID))
	t.Cleanup(func() { f.service.rateLimiter.Release(f.userID) })
	for range 3 {
		result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
		require.NoError(t, err)
		assert.Equal(t, ErrorCodeRateLimited, result.ErrorCode)
		assert.Equal(t, 0, f.provider.calls)
	}

	// Only the permit owner releasing its slot permits a new generation.
	f.service.rateLimiter.Release(f.userID)
	result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
	require.NoError(t, err)
	assert.Empty(t, result.ErrorCode)
	assert.Equal(t, 1, f.provider.calls)
	// Successful service completion releases its own permit as before.
	require.NoError(t, f.service.rateLimiter.Allow(t.Context(), f.userID))
}
