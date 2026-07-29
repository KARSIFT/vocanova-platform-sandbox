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
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	contract "github.com/KARSIFT/vocanova-platform/apps/api/app/api"
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

	api, err := contract.NewProductionAPI(cfg, nil)
	if err != nil {
		return fmt.Errorf("build api: %w", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Adapter(),
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
