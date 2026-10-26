package auth

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
)

func newTestLimiter(window time.Duration, limit, maxKeys int) (*FixedWindowRateLimiter, *clock.Fixed) {
	c := &clock.Fixed{T: time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)}
	l := NewFixedWindowRateLimiter(c, window, limit)
	l.maxKeys = maxKeys
	return l, c
}

func mustAllow(t *testing.T, l *FixedWindowRateLimiter, key string, want bool) {
	t.Helper()
	got, err := l.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("Allow(%q) error: %v", key, err)
	}
	if got != want {
		t.Fatalf("Allow(%q) = %v, want %v", key, got, want)
	}
}

func TestFixedWindowRateLimiterDefaultCapacityIsEnforced(t *testing.T) {
	l := NewFixedWindowRateLimiter(clock.Fixed{T: time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)}, time.Minute, 1)
	for i := 0; i < DefaultRateLimiterMaxKeys; i++ {
		mustAllow(t, l, fmt.Sprintf("default-%d", i), true)
	}
	mustAllow(t, l, "overflow", false)
	mustAllow(t, l, "default-0", false)
}

func TestFixedWindowRateLimiterExistingKeyLimitedAndResets(t *testing.T) {
	l, c := newTestLimiter(time.Minute, 2, 10)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", false)
	c.Advance(59 * time.Second)
	mustAllow(t, l, "a", false)
	c.Advance(time.Second)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", false)
}

func TestFixedWindowRateLimiterRejectsUnknownKeysWhenFull(t *testing.T) {
	l, _ := newTestLimiter(time.Minute, 1, 2)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "b", true)
	mustAllow(t, l, "c", false)
	// Live keys are neither evicted nor reset by the rejection.
	mustAllow(t, l, "a", false)
	mustAllow(t, l, "b", false)
	if _, ok := l.buckets["c"]; ok {
		t.Fatal("rejected key was tracked")
	}
	if len(l.buckets) != 2 {
		t.Fatalf("tracked keys = %d, want 2", len(l.buckets))
	}
}

func TestFixedWindowRateLimiterExistingKeysAllowedWhenFull(t *testing.T) {
	l, _ := newTestLimiter(time.Minute, 3, 2)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "b", true)
	mustAllow(t, l, "c", false)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", true)
	mustAllow(t, l, "a", false)
}

func TestFixedWindowRateLimiterPeriodicSweepReclaimsExpired(t *testing.T) {
	l, c := newTestLimiter(time.Minute, 1, 100)
	for i := 0; i < 50; i++ {
		mustAllow(t, l, fmt.Sprintf("k%d", i), true)
	}
	c.Advance(time.Minute)
	mustAllow(t, l, "fresh", true)
	if len(l.buckets) != 1 {
		t.Fatalf("tracked keys = %d, want 1 after expiry sweep", len(l.buckets))
	}
}

func TestFixedWindowRateLimiterCapacitySweepReclaimsExpired(t *testing.T) {
	l, c := newTestLimiter(time.Minute, 1, 2)
	mustAllow(t, l, "a", true)
	c.Advance(30 * time.Second)
	mustAllow(t, l, "b", true)
	c.Advance(30 * time.Second)
	// Hold off the periodic sweep so only the capacity sweep can reclaim.
	l.nextSweep = c.Now().Add(time.Hour)
	mustAllow(t, l, "c", true)
	if _, ok := l.buckets["a"]; ok {
		t.Fatal("expired key a was not reclaimed")
	}
	// b is still live, keeps its count and its slot.
	mustAllow(t, l, "b", false)
	mustAllow(t, l, "d", false)
}

func TestFixedWindowRateLimiterConcurrent(t *testing.T) {
	const (
		workers   = 32
		perWorker = 200
		limit     = 100
		maxKeys   = 64
	)
	l, _ := newTestLimiter(time.Minute, limit, maxKeys)
	// Track the shared key before unique keys can fill capacity.
	mustAllow(t, l, "shared", true)
	var allowed int64 = 1
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				ok, err := l.Allow(context.Background(), "shared")
				if err != nil {
					t.Errorf("Allow(shared) error: %v", err)
					return
				}
				if ok {
					atomic.AddInt64(&allowed, 1)
				}
				if _, err := l.Allow(context.Background(), fmt.Sprintf("w%d-%d", w, i)); err != nil {
					t.Errorf("Allow(unique) error: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()
	if got := atomic.LoadInt64(&allowed); got != limit {
		t.Fatalf("shared allowed = %d, want %d", got, limit)
	}
	if n := len(l.buckets); n > maxKeys {
		t.Fatalf("tracked keys = %d, want <= %d", n, maxKeys)
	}
}
