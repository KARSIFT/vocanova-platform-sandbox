package aifeedback

import (
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRateLimiterAllowsRequest(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	lim := NewMemoryRateLimiter(DefaultRateLimitConfig(), c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	require.NoError(t, lim.Allow(t.Context(), userID))
	lim.Release(userID)
}

func TestMemoryRateLimiterBlocksConcurrentActive(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	lim := NewMemoryRateLimiter(DefaultRateLimitConfig(), c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	require.NoError(t, lim.Allow(t.Context(), userID))
	assert.ErrorIs(t, lim.Allow(t.Context(), userID), ErrRateLimited)
	lim.Release(userID)
}

func TestMemoryRateLimiterEnforcesPerMinuteLimit(t *testing.T) {
	start := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	c := clock.Fixed{T: start}
	cfg := RateLimitConfig{MaxActivePerLearner: 1, MaxPerMinute: 2, MaxPerDay: 10}
	lim := NewMemoryRateLimiter(cfg, &c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	require.NoError(t, lim.Allow(t.Context(), userID))
	lim.Release(userID)
	c.Advance(time.Second)
	require.NoError(t, lim.Allow(t.Context(), userID))
	lim.Release(userID)
	c.Advance(time.Second)
	assert.ErrorIs(t, lim.Allow(t.Context(), userID), ErrRateLimited)
}

func TestMemoryRateLimiterReleaseFreesSlot(t *testing.T) {
	c := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	lim := NewMemoryRateLimiter(DefaultRateLimitConfig(), c)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	require.NoError(t, lim.Allow(t.Context(), userID))
	lim.Release(userID)
	require.NoError(t, lim.Allow(t.Context(), userID))
}
