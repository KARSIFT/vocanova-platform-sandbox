// Command api starts the production HTTP server: real,
// database-backed, listening on PORT, with the full DOC-12 F2
// "run both apps using only documented commands" path. The
// server refuses to start if the database is unreachable, honors
// every DOC-11 §3 kill switch via the env-driven production
// config, and shuts down gracefully on SIGTERM so containers can
// stop cleanly during rollovers.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	contract "github.com/KARSIFT/vocanova-platform/apps/api/app/api"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api: %v\n", err)
		os.Exit(1)
	}
}

// run is the testable entry point. It returns a non-nil error on
// any startup failure (missing config, database unreachable, etc.)
// so the existing TestRun harness and any future T00
// health-check tests can exercise the wiring without spinning up
// a real process.
func run() error {
	cfg, err := contract.LoadProductionConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.SentryDSN != "" {
		release := cfg.SentryRelease
		if release == "" {
			release = "unversioned"
		}
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDSN,
			Environment: cfg.SentryEnvironment,
			Release:     release,
		}); err != nil {
			return fmt.Errorf("init sentry: %w", err)
		}
		defer sentry.Flush(2 * time.Second)
		fmt.Fprintf(os.Stderr, "api: sentry enabled (env=%s)\n", cfg.SentryEnvironment)
	} else {
		fmt.Fprintln(os.Stderr, "api: sentry disabled (SENTRY_DSN unset)")
	}

	api, db, err := contract.NewProductionAPI(cfg, nil)
	if err != nil {
		return fmt.Errorf("build api: %w", err)
	}
	defer db.Close()

	// A deletion request deliberately deactivates first, then is purged after
	// its grace period. Without this loop, a due row can remain deactivated
	// forever because no request path invokes RunDeletionSweep. Database-level
	// claims make concurrent API replicas safe; each process may run this same
	// bounded loop.
	sweepCtx, cancelSweep := context.WithCancel(context.Background())
	sweepDone := make(chan struct{})
	go func() {
		defer close(sweepDone)
		runDeletionSweepLoop(sweepCtx, newDeletionSweepService(db), cfg.AccountDeletionSweepInterval)
	}()
	defer func() {
		stopDeletionSweep(cancelSweep, sweepDone)
	}()

	// Real unhandled errors/panics from any request, not just the
	// deliberate VOC-037-T04 test endpoint, must reach Sentry - the
	// deliberate test event alone proves the DSN/token wiring, not that
	// "error monitoring active" (DOC-11 §5) is true for real traffic.
	// sentryhttp.Handle is a no-op-safe wrap even when Sentry was never
	// initialized (SENTRY_DSN unset): the default hub's client is nil and
	// events are silently dropped, matching every other Sentry call site
	// in this codebase.
	handler := sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(api.Adapter())

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stderr, "api: listening on %s (env=%s, ai=%s, magic=%s, oauth=%s, signups=%s)\n",
			srv.Addr,
			cfg.Environment,
			boolFlag(cfg.AIEnabled),
			boolFlag(cfg.MagicLinkOn),
			boolFlag(cfg.OAuthOn),
			boolFlag(cfg.NewSignupsOn),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-stop:
		fmt.Fprintf(os.Stderr, "api: %s received, shutting down\n", sig)
	case err := <-serverErr:
		return fmt.Errorf("listen: %w", err)
	}

	// Stop and join background work before beginning the HTTP drain. Otherwise
	// a ticker can start a new irreversible purge during the server's 30-second
	// graceful-shutdown window. The deferred call remains a safety net for an
	// earlier return; cancellation and reads from a closed done channel are both
	// idempotent.
	stopDeletionSweep(cancelSweep, sweepDone)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func boolFlag(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

type deletionSweeper interface {
	RunDeletionSweep(ctx context.Context, clientIP, sessionToken string) (*accounts.SweepResult, error)
}

// stopDeletionSweep cancels the loop and waits until it has exited while the
// database is still open. It is safe to invoke more than once.
func stopDeletionSweep(cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	<-done
}

// newDeletionSweepService builds the narrow production service instance used
// only by the background purge loop. Request-only collaborators are nil: the
// sweep uses just the repository, clock, rate limiter, and deletion settings.
func newDeletionSweepService(db *sql.DB) deletionSweeper {
	clk := clock.Real{}
	return accounts.NewService(
		accounts.NewPostgreSQLRepository(db), nil, nil, nil, clk,
		auth.NewFixedWindowRateLimiter(clk, time.Hour, 60),
		accounts.Config{},
	)
}

// runDeletionSweepLoop performs one pass at startup (so overdue requests do
// not wait a whole cadence) and then at each configured interval. Errors are
// logged without account identifiers; a later pass safely resumes a failed or
// stale claim. Context cancellation stops the ticker promptly during shutdown.
func runDeletionSweepLoop(ctx context.Context, sweeper deletionSweeper, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	run := func() {
		result, err := sweeper.RunDeletionSweep(ctx, "internal", "internal")
		if err != nil {
			fmt.Fprintf(os.Stderr, "api: account-deletion sweep failed: %v\n", err)
			return
		}
		if result.Anonymized > 0 || result.Failed > 0 {
			fmt.Fprintf(os.Stderr, "api: account-deletion sweep processed=%d anonymized=%d failed=%d\n", result.Processed, result.Anonymized, result.Failed)
		}
	}

	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
