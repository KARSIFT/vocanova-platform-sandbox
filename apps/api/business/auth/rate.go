package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
)

// RateLimiter allows keyed requests up to a fixed count within a fixed window.
// It is an interface so production can use Redis without changing the service.
type RateLimiter interface {
	// Allow returns true if the request is within the limit for key.
	Allow(ctx context.Context, key string) (bool, error)
}

// DefaultRateLimiterMaxKeys bounds how many distinct keys an in-memory
// FixedWindowRateLimiter tracks at once.
const DefaultRateLimiterMaxKeys = 10000

// FixedWindowRateLimiter is a simple in-memory fixed-window rate limiter.
// It is used by each single-process environment. Limits are not shared across
// replicas; use a shared backend before scaling horizontally. Memory is bounded: expired buckets are reclaimed by a
// periodic sweep (at most once per window) and by a capacity sweep when the
// key limit is reached; if no bucket has expired, unknown keys are rejected
// rather than evicting keys that are still being limited.
type FixedWindowRateLimiter struct {
	mu      sync.Mutex
	clock   clock.Clock
	window  time.Duration
	limit   int
	maxKeys int
	buckets map[string]*bucket

	// nextSweep is when the next periodic sweep is due. earliestReset is a
	// lower bound on every tracked bucket's resetAt, zero when empty.
	nextSweep     time.Time
	earliestReset time.Time
}

type bucket struct {
	count   int
	resetAt time.Time
}

// NewFixedWindowRateLimiter creates an in-memory rate limiter that tracks at
// most DefaultRateLimiterMaxKeys keys.
func NewFixedWindowRateLimiter(c clock.Clock, window time.Duration, limit int) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clock:   c,
		window:  window,
		limit:   limit,
		maxKeys: DefaultRateLimiterMaxKeys,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the key has not exceeded its limit. Unknown keys are
// rejected while the limiter is full of unexpired buckets.
func (l *FixedWindowRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock.Now()
	if !now.Before(l.nextSweep) {
		l.sweep(now)
		l.nextSweep = now.Add(l.sweepInterval())
	}
	b, ok := l.buckets[key]
	switch {
	case ok && !now.Before(b.resetAt):
		b.count = 0
		b.resetAt = now.Add(l.window)
	case !ok:
		if len(l.buckets) >= l.maxKeys {
			if !l.earliestReset.IsZero() && !now.Before(l.earliestReset) {
				l.sweep(now)
			}
			if len(l.buckets) >= l.maxKeys {
				return false, nil
			}
		}
		b = &bucket{resetAt: now.Add(l.window)}
		l.buckets[key] = b
		if l.earliestReset.IsZero() || b.resetAt.Before(l.earliestReset) {
			l.earliestReset = b.resetAt
		}
	}
	if b.count >= l.limit {
		return false, nil
	}
	b.count++
	return true, nil
}

// sweep deletes expired buckets and recomputes earliestReset. It is O(n) in
// the tracked keys, which maxKeys bounds. Callers must hold l.mu.
func (l *FixedWindowRateLimiter) sweep(now time.Time) {
	var earliest time.Time
	for k, b := range l.buckets {
		if !now.Before(b.resetAt) {
			delete(l.buckets, k)
			continue
		}
		if earliest.IsZero() || b.resetAt.Before(earliest) {
			earliest = b.resetAt
		}
	}
	l.earliestReset = earliest
}

func (l *FixedWindowRateLimiter) sweepInterval() time.Duration {
	if l.window > 0 {
		return l.window
	}
	return time.Second
}

// KeyForIP builds a key for a client IP address and action.
func KeyForIP(action, ip string) string { return fmt.Sprintf("ip:%s:%s", action, ip) }

// KeyForEmail scopes anonymous authentication requests without retaining the
// address itself in a limiter backend or its operational logs.
func KeyForEmail(action, address string) string {
	return fmt.Sprintf("email:%s:%s", action, hashTokenString(normalizeEmail(address)))
}

// KeyForSession builds a key for a session token and action.
func KeyForSession(action, token string) string {
	return fmt.Sprintf("session:%s:%s", action, hashTokenString(token))
}
