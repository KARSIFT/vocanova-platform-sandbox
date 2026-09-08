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
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
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

	// Auth.Cleanup owns the bounded deletion of expired sessions, magic links,
	// and OAuth states. It has no HTTP caller, so run it independently of
	// request traffic and stop it before releasing the database on shutdown.
	cleanupCtx, cancelCleanup := context.WithCancel(context.Background())
	cleanupDone := make(chan struct{})
	go func() {
		defer close(cleanupDone)
		runAuthCleanupLoop(cleanupCtx, newAuthCleanupService(db), cfg.AuthCleanupInterval)
	}()
	defer func() {
		stopAuthCleanup(cancelCleanup, cleanupDone)
	}()

	// Idempotency records stop affecting requests after 24 hours. Delete old
	// rows in bounded batches so normal traffic cannot grow the table forever.
	idempotencyCtx, cancelIdempotencyCleanup := context.WithCancel(context.Background())
	idempotencyCleanupDone := make(chan struct{})
	go func() {
		defer close(idempotencyCleanupDone)
		runIdempotencyCleanupLoop(idempotencyCtx, newIdempotencyCleaner(db), cfg.IdempotencyCleanupInterval)
	}()
	defer func() {
		stopIdempotencyCleanup(cancelIdempotencyCleanup, idempotencyCleanupDone)
	}()

	// Email-change confirmation links are account-owned rather than auth-owned,
	// but have the same short-lived credential lifecycle. Run their cleanup on
	// the configured auth-cleanup cadence so consumed, revoked, and expired
	// links do not retain a pending email address or token hash indefinitely.
	emailChangeCleanupCtx, cancelEmailChangeCleanup := context.WithCancel(context.Background())
	emailChangeCleanupDone := make(chan struct{})
	go func() {
		defer close(emailChangeCleanupDone)
		runEmailChangeCleanupLoop(emailChangeCleanupCtx, newEmailChangeCleanupService(db), cfg.AuthCleanupInterval)
	}()
	defer func() {
		stopEmailChangeCleanup(cancelEmailChangeCleanup, emailChangeCleanupDone)
	}()

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

	// Stop and join all background jobs before beginning the HTTP drain. This
	// prevents a ticker from beginning a new credential-delete, idempotency-
	// cleanup, email-change cleanup, or account-purge pass during the graceful-
	// shutdown window. The deferred calls remain
	// safety nets for earlier returns; cancellation and reads from closed done
	// channels are idempotent.
	stopAuthCleanup(cancelCleanup, cleanupDone)
	stopIdempotencyCleanup(cancelIdempotencyCleanup, idempotencyCleanupDone)
	stopEmailChangeCleanup(cancelEmailChangeCleanup, emailChangeCleanupDone)
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

type authCleaner interface {
	Cleanup(context.Context) error
}

type emailChangeCleaner interface {
	CleanupExpiredEmailChangeLinks(context.Context) error
}

// stopAuthCleanup cancels the loop and waits until it has exited while the
// database remains open. It is safe to invoke more than once.
func stopAuthCleanup(cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	<-done
}

// stopEmailChangeCleanup cancels the loop and waits until it has exited while
// the database remains open. It is safe to invoke more than once.
func stopEmailChangeCleanup(cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	<-done
}

type deletionSweeper interface {
	RunDeletionSweep(ctx context.Context, clientIP, sessionToken string) (*accounts.SweepResult, error)
}

const idempotencyCleanupBatchSize = 1_000

type idempotencyCleaner interface {
	CleanupExpired(context.Context, int) (int, error)
}

// stopIdempotencyCleanup cancels the loop and joins it while the database is
// still open. It is safe to invoke more than once.
func stopIdempotencyCleanup(cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	<-done
}

// stopDeletionSweep cancels the loop and waits until it has exited while the
// database is still open. It is safe to invoke more than once.
func stopDeletionSweep(cancel context.CancelFunc, done <-chan struct{}) {
	cancel()
	<-done
}

// newAuthCleanupService builds the narrow production service instance used by
// the cleanup loop. Cleanup only needs the PostgreSQL repository and clock;
// delivery, OAuth, and request-rate-limit collaborators are not involved.
func newAuthCleanupService(db *sql.DB) authCleaner {
	return auth.NewService(auth.NewPostgreSQLRepository(db), nil, nil, clock.Real{}, nil, auth.Config{})
}

// newEmailChangeCleanupService builds the narrow account-owned service used
// only by the background credential-retention loop.
func newEmailChangeCleanupService(db *sql.DB) emailChangeCleaner {
	return accounts.NewService(accounts.NewPostgreSQLRepository(db), nil, nil, nil, clock.Real{}, nil, accounts.Config{})
}

// runAuthCleanupLoop cleans once at startup, then at the configured cadence.
// A failed pass is logged without credential or learner data and a later pass
// retries it. Cancellation stops the ticker promptly during shutdown.
func runAuthCleanupLoop(ctx context.Context, cleaner authCleaner, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	run := func() {
		if err := cleaner.Cleanup(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "api: auth cleanup failed: %v\n", err)
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

func newIdempotencyCleaner(db *sql.DB) idempotencyCleaner {
	return learning.NewPostgreSQLIdempotencyStore(db)
}

// runIdempotencyCleanupLoop deletes one bounded batch at startup and at each
// cadence. Failures contain no record contents and are retried by a later pass.
func runIdempotencyCleanupLoop(ctx context.Context, cleaner idempotencyCleaner, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	run := func() {
		deleted, err := cleaner.CleanupExpired(ctx, idempotencyCleanupBatchSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "api: idempotency cleanup failed: %v\n", err)
			return
		}
		if deleted > 0 {
			fmt.Fprintf(os.Stderr, "api: idempotency cleanup deleted=%d\n", deleted)
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

// runEmailChangeCleanupLoop removes expired, consumed, and revoked email
// confirmation links once at startup and then at the auth-cleanup cadence.
// Errors intentionally contain no link or account identifiers, and a failed
// pass does not prevent the next one.
func runEmailChangeCleanupLoop(ctx context.Context, cleaner emailChangeCleaner, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	run := func() {
		if err := cleaner.CleanupExpiredEmailChangeLinks(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "api: email-change cleanup failed: %v\n", err)
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
