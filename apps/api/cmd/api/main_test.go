package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
)

type fakeAuthCleaner struct{ calls atomic.Int32 }

func (f *fakeAuthCleaner) Cleanup(context.Context) error {
	f.calls.Add(1)
	return nil
}

func TestRunAuthCleanupLoopRunsImmediatelyAndStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cleaner := &fakeAuthCleaner{}
	done := make(chan struct{})
	go func() {
		runAuthCleanupLoop(ctx, cleaner, time.Hour)
		close(done)
	}()

	deadline := time.After(time.Second)
	for cleaner.calls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("auth cleanup did not run at startup")
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("auth cleanup loop did not stop after cancellation")
	}
}

func TestStopAuthCleanupWaitsForLoopExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()
	stopAuthCleanup(cancel, done)
}

type fakeIdempotencyCleaner struct {
	calls  atomic.Int32
	limits chan int
	err    error
}

func (f *fakeIdempotencyCleaner) CleanupExpired(_ context.Context, limit int) (int, error) {
	f.calls.Add(1)
	if f.limits != nil {
		select {
		case f.limits <- limit:
		default:
		}
	}
	return 0, f.err
}

func TestRunIdempotencyCleanupLoopRetriesAfterFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cleaner := &fakeIdempotencyCleaner{
		limits: make(chan int, 2),
		err:    errors.New("transient database failure"),
	}
	done := make(chan struct{})
	go func() {
		runIdempotencyCleanupLoop(ctx, cleaner, time.Millisecond)
		close(done)
	}()

	for range 2 {
		select {
		case <-cleaner.limits:
		case <-time.After(time.Second):
			t.Fatal("idempotency cleanup failure was not retried")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("idempotency cleanup loop did not stop after retry")
	}
}

func TestRunIdempotencyCleanupLoopRunsImmediatelyAndStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cleaner := &fakeIdempotencyCleaner{limits: make(chan int, 1)}
	done := make(chan struct{})
	go func() {
		runIdempotencyCleanupLoop(ctx, cleaner, time.Hour)
		close(done)
	}()

	select {
	case limit := <-cleaner.limits:
		if limit != idempotencyCleanupBatchSize {
			t.Fatalf("cleanup limit = %d, want %d", limit, idempotencyCleanupBatchSize)
		}
	case <-time.After(time.Second):
		t.Fatal("idempotency cleanup did not run at startup")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("idempotency cleanup loop did not stop after cancellation")
	}
}

func TestStopIdempotencyCleanupWaitsForLoopExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()
	stopIdempotencyCleanup(cancel, done)
}

type fakeDeletionSweeper struct{ calls atomic.Int32 }

func (f *fakeDeletionSweeper) RunDeletionSweep(context.Context, string, string) (*accounts.SweepResult, error) {
	f.calls.Add(1)
	return &accounts.SweepResult{}, nil
}

func TestRunDeletionSweepLoopRunsImmediatelyAndStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sweeper := &fakeDeletionSweeper{}
	done := make(chan struct{})
	go func() {
		runDeletionSweepLoop(ctx, sweeper, time.Hour)
		close(done)
	}()

	deadline := time.After(time.Second)
	for sweeper.calls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("deletion sweep did not run at startup")
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deletion sweep loop did not stop after cancellation")
	}
}

func TestStopDeletionSweepWaitsForLoopExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()
	stopDeletionSweep(cancel, done)
}

// TestRun_RejectsMissingDatabaseURL covers the first
// config-load safety property: a process started with no
// DATABASE_URL must exit non-zero with a clear error message,
// not panic or start a server with a nil pool. This is the
// DOC-11 §3 "no service should start with half-configured
// state" guarantee the founder expects.
func TestRun_RejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("OAUTH_REDIRECT_URI", "")
	t.Setenv("SESSION_COOKIE_DOMAIN", "")

	if err := run(); err == nil {
		t.Fatal("run() must return an error when DATABASE_URL is missing")
	}
}

// TestRun_RejectsMissingBaseURL covers the second required
// env var the production wiring requires: BASE_URL must be
// set so the auth service can build absolute magic-link URLs.
// This guard is independent of the database reachability
// check, so a misconfigured BASE_URL never lets a broken
// auth service start.
func TestRun_RejectsMissingBaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example/db")
	t.Setenv("BASE_URL", "")
	t.Setenv("OAUTH_REDIRECT_URI", "https://api-staging.vocanova.site/auth/oauth/google/callback")
	t.Setenv("SESSION_COOKIE_DOMAIN", "staging.vocanova.site")

	if err := run(); err == nil {
		t.Fatal("run() must return an error when BASE_URL is missing")
	}
}
