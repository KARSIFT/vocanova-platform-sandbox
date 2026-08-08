# VOC-050 — Acceptance Criteria

## VOC-050-AC-00 — Synthetic test account exists, is obviously synthetic, and is seeded idempotently

- Requirement source: `specification.md`'s "Decisions, contradictions,
  security, and privacy" section
- Tasks: `VOC-050-T00`
- Tests: `VOC-050-TEST-00`
- Evidence: `VOC-050-EV-00`
- Result: pending

A dedicated test-account row exists on both staging and production,
distinguishable from every real account at the database and session layer
(e.g. reserved email domain and/or an explicit synthetic-account flag).
Re-running the seed mechanism against a database where the account already
exists does not create a duplicate or error.

## VOC-050-AC-01 — Synthetic account is not reachable via any real signup path

- Requirement source: `specification.md`'s "Decisions, contradictions,
  security, and privacy" section (issue #391's constraint)
- Tasks: `VOC-050-T00`
- Tests: `VOC-050-TEST-01`
- Evidence: `VOC-050-EV-00`
- Result: pending

A real user cannot cause an account with the synthetic account's reserved
identity (or any account bearing its synthetic-marker) to be created or
authenticated through the real magic-link or OAuth signup/sign-in paths.

## VOC-050-AC-02 — Session minting is narrowly scoped, secret-gated, and fails closed

- Requirement source: `specification.md`'s "Decisions, contradictions,
  security, and privacy" section
- Tasks: `VOC-050-T01`
- Tests: `VOC-050-TEST-02`
- Evidence: `VOC-050-EV-01`
- Result: pending

The session-minting mechanism only ever mints a session for the one
synthetic account, requires a secret gating token to invoke, and is a no-op
(does not register/expose any reachable behavior) whenever that token is
unset - mirroring `RegisterMonitoringSentryTest`'s existing fail-closed
pattern (`apps/api/app/api/production.go`).

## VOC-050-AC-03 — `deploy-staging.yml` runs the real core-loop journey against staging and fails closed on any step failure

- Requirement source: `specification.md`'s objective (issue #391 requirement 2)
- Tasks: `VOC-050-T02`
- Tests: `VOC-050-TEST-03`
- Evidence: `VOC-050-EV-02`
- Result: pending

Immediately after `deploy-staging.yml`'s existing `/healthz` poll passes, a
new step runs the real, authenticated core-loop journey against
`https://staging.vocanova.site` using the minted synthetic-account session.
Any journey-step failure fails the workflow job itself (no
`continue-on-error`), matching the file's existing fail-closed design for
its other steps.

## VOC-050-AC-04 — A failed staging core-loop run is recorded as a blocking signal for the downstream promotion decision

- Requirement source: `specification.md`'s objective and open question 1
  (issue #391 requirement 3)
- Tasks: `VOC-050-T03`
- Tests: `VOC-050-TEST-04`
- Evidence: `VOC-050-EV-03`
- Result: pending

`deploy-staging.yml`'s job conclusion is failure whenever the core-loop
check fails, with no path that reports success despite that failure. Because
this package cannot modify `karsift-ai-infra`'s `release.yml` (a different
repository - see `specification.md`'s open question 1), this criterion is
satisfied by confirming the failure signal exists and is accurate on this
repository's side; it does not, by itself, prove `release.yml` acts on that
signal. `VOC-050-T03`'s pull request must explicitly record this limitation
and the open cross-repo item for the reviewing human, rather than silently
implying the gate is proven end-to-end.

## VOC-050-AC-05 — Production core-loop smoke check runs for real instead of SKIPping, with no state-mutating side effects

- Requirement source: `specification.md`'s objective (issue #391 requirement 4)
- Tasks: `VOC-050-T04`
- Tests: `VOC-050-TEST-05`
- Evidence: `VOC-050-EV-04`
- Result: pending

`deploy-production.yml` provides a real, non-empty `SMOKE_TEST_SESSION_COOKIE`
to `infra/scripts/smoke-test-production.sh`, whose core-loop section (its own
`# 5.` section) then runs its `GET /api/v1/me` and
`GET /api/v1/journey-situations` checks instead of printing `SKIP:`. No new
request in the activated check performs a state-mutating action against a
real target (no real magic-link request, no real OAuth flow completion,
matching the script's existing documented design).

Acceptance criteria remain observable, stable, security-aware, and
bidirectionally traceable to requirements, tasks, tests, and evidence. None
of these criteria may be marked satisfied until this package is adopted and
each task is actually implemented.
