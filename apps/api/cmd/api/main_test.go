package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
)

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
