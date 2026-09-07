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
		cancelCleanup()
		<-cleanupDone
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

// newAuthCleanupService builds the narrow production service instance used by
// the cleanup loop. Cleanup only needs the PostgreSQL repository and clock;
// delivery, OAuth, and request-rate-limit collaborators are not involved.
func newAuthCleanupService(db *sql.DB) authCleaner {
	return auth.NewService(auth.NewPostgreSQLRepository(db), nil, nil, clock.Real{}, nil, auth.Config{})
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
