package aifeedback

import (
	"context"
	"sync"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// RateLimiter scopes generation requests per learner.
type RateLimiter interface {
	// Allow acquires a permit only on success. An error grants no permit.
	Allow(ctx context.Context, userID uuid.UUID) error
	// Release must be called only by a request that successfully called Allow.
	Release(userID uuid.UUID)
}

// RateLimitConfig is the configurable starting limit set (DOC-09 §19).
type RateLimitConfig struct {
	MaxActivePerLearner int
	MaxPerMinute        int
	MaxPerDay           int
}

// DefaultRateLimitConfig returns the approved starting values.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxActivePerLearner: 1,
		MaxPerMinute:        5,
		MaxPerDay:           30,
	}
}

// MemoryRateLimiter is a deterministic, in-process rate limiter for tests and
// single-process deployments. It is not safe across processes.
type MemoryRateLimiter struct {
	mu      sync.Mutex
	config  RateLimitConfig
	clock   clock.Clock
	active  map[uuid.UUID]bool
	history map[uuid.UUID][]time.Time
}

// NewMemoryRateLimiter creates an in-memory rate limiter.
func NewMemoryRateLimiter(config RateLimitConfig, c clock.Clock) *MemoryRateLimiter {
	if c == nil {
		c = clock.Real{}
	}
	return &MemoryRateLimiter{
		config:  config,
		clock:   c,
		active:  make(map[uuid.UUID]bool),
		history: make(map[uuid.UUID][]time.Time),
	}
}

// Allow checks the active-generation, per-minute, and per-day limits.
func (r *MemoryRateLimiter) Allow(ctx context.Context, userID uuid.UUID) error {
	if r.config.MaxActivePerLearner <= 0 && r.config.MaxPerMinute <= 0 && r.config.MaxPerDay <= 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.active[userID] {
		return ErrRateLimited
	}

	now := r.clock.Now().UTC()
	minuteAgo := now.Add(-time.Minute)
	dayAgo := now.Add(-24 * time.Hour)

	var inMinute, inDay int
	kept := r.history[userID][:0]
	for _, t := range r.history[userID] {
		if t.After(dayAgo) || t.Equal(dayAgo) {
			inDay++
			if t.After(minuteAgo) || t.Equal(minuteAgo) {
				inMinute++
			}
			kept = append(kept, t)
		}
	}
	r.history[userID] = kept

	if r.config.MaxPerDay > 0 && inDay >= r.config.MaxPerDay {
		return ErrRateLimited
	}
	if r.config.MaxPerMinute > 0 && inMinute >= r.config.MaxPerMinute {
		return ErrRateLimited
	}

	r.active[userID] = true
	r.history[userID] = append(r.history[userID], now)
	return nil
}

// Release frees the active-generation slot.
func (r *MemoryRateLimiter) Release(userID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.active, userID)
}
