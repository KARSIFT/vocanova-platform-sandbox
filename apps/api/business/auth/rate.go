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

// FixedWindowRateLimiter is a simple in-memory fixed-window rate limiter.
// It is suitable for tests and single-process deployments; production uses
// a shared backend.
type FixedWindowRateLimiter struct {
	mu      sync.Mutex
	clock   clock.Clock
	window  time.Duration
	limit   int
	buckets map[string]*bucket
}

type bucket struct {
	count   int
	resetAt time.Time
}

// NewFixedWindowRateLimiter creates an in-memory rate limiter.
func NewFixedWindowRateLimiter(c clock.Clock, window time.Duration, limit int) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clock:   c,
		window:  window,
		limit:   limit,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the key has not exceeded its limit.
func (l *FixedWindowRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock.Now()
	b, ok := l.buckets[key]
	if !ok || !now.Before(b.resetAt) {
		b = &bucket{count: 0, resetAt: now.Add(l.window)}
		l.buckets[key] = b
	}
	if b.count >= l.limit {
		return false, nil
	}
	b.count++
	return true, nil
}

// KeyForIP builds a key for a client IP address and action.
func KeyForIP(action, ip string) string { return fmt.Sprintf("ip:%s:%s", action, ip) }

// KeyForSession builds a key for a session token and action.
func KeyForSession(action, token string) string {
	return fmt.Sprintf("session:%s:%s", action, hashTokenString(token))
}
