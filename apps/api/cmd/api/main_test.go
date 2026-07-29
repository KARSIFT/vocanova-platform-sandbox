package main

import (
	"testing"
)

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
