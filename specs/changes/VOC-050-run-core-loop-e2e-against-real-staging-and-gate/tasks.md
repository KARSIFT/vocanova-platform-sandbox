# VOC-050 — Tasks

## VOC-050-T00 — Seed a dedicated, obviously-synthetic test user automatically on staging and production deploy

- Requirement source: `specification.md`'s "Scope and non-goals" and
  "Decisions, contradictions, security, and privacy" sections
- Acceptance criteria: `VOC-050-AC-00`, `VOC-050-AC-01`
- Tests: `VOC-050-TEST-00`, `VOC-050-TEST-01`
- Evidence: `VOC-050-EV-00`
- Status: pending

Add a migration under `apps/api/migrations` (following `.karsift/lessons.md`'s
Atlas-directive and duplicate-unique-index lessons) and an idempotent seed
mechanism (new command or extension of `apps/api/cmd/seed`, implementer's
choice) that creates a single, obviously-synthetic test user account -
distinguishable via a reserved, non-deliverable email domain and/or an
explicit boolean flag - safe to rerun on every deploy without duplicating or
erroring. Resolve `change.yaml`'s `VOC-050-DEP-02` by confirming the current
signup/magic-link code's actual behavior and choosing a concrete mechanism
that makes the account unreachable via any real signup path, documented
explicitly in this task's pull request rather than assumed. Wire the seed
step into whichever of `deploy-staging.yml`/`deploy-production.yml` (or
both) is the right place per this task's own findings about the deploy
sequence, gated behind the existing migration-apply step's location in each
file.

## VOC-050-T01 — Mint the synthetic account's session server-side, without a real magic-link/OAuth round-trip

- Requirement source: `specification.md`'s "Scope and non-goals" and
  "Decisions, contradictions, security, and privacy" sections; depends on
  `VOC-050-T00`
- Acceptance criteria: `VOC-050-AC-02`
- Tests: `VOC-050-TEST-02`
- Evidence: `VOC-050-EV-01`
- Status: pending

Add a narrowly-scoped, secret-gated mechanism in `apps/api/business/auth`
(or a small CLI/admin entry point calling into it) that mints a valid
session for `VOC-050-T00`'s synthetic account only, mirroring
`apps/api/app/api/production.go`'s `RegisterMonitoringSentryTest` fail-closed
pattern (a no-op, not-reachable behavior whenever the gating token is
unset). This mechanism must never mint a session for any other account.
Document the exact invocation this mechanism expects (CLI flag, HTTP
endpoint + header, etc.) in this task's pull request, since `VOC-050-T02` and
`VOC-050-T04` both depend on invoking it from CI/deploy-workflow steps.

## VOC-050-T02 — Run core-loop.spec.ts (or a staging-targeted variant) against real staging in deploy-staging.yml

- Requirement source: `specification.md`'s "Scope and non-goals" section and
  open questions 2/3; depends on `VOC-050-T00`, `VOC-050-T01`
- Acceptance criteria: `VOC-050-AC-03`
- Tests: `VOC-050-TEST-03`
- Evidence: `VOC-050-EV-02`
- Result: pending

Resolve `specification.md`'s open questions 2 and 3 explicitly in this
task's pull request: decide whether the staging-targeted journey reuses
`apps/web/tests/e2e/core-loop.spec.ts` with environment-aware setup or is a
separate sibling spec file, and decide how the journey reaches an
authenticated starting state against the real backend (e.g. the synthetic
account is seeded already past onboarding by `T00`, using `T01`'s minted
session, rather than relying on the mock-only `e2e_onboarding_status`/
`e2e_unauthenticated` override cookies against a real backend that does not
recognize them). Add a new step to `.github/workflows/deploy-staging.yml`,
immediately after the existing "Poll api-staging.vocanova.site/healthz"
step, that runs this journey against `https://staging.vocanova.site`. The
step must have no `continue-on-error` and must fail the job on any
journey-step failure, matching the file's existing fail-closed style.

## VOC-050-T03 — Make a failed staging core-loop run fail deploy-staging.yml closed, and document the cross-repo release-gating dependency

- Requirement source: `specification.md`'s objective and open question 1;
  depends on `VOC-050-T02`
- Acceptance criteria: `VOC-050-AC-04`
- Tests: `VOC-050-TEST-04`
- Evidence: `VOC-050-EV-03`
- Status: pending

Confirm, with a live-observed run (a deliberately-broken target or a
disposable test branch - not just an inspection of the YAML), that
`deploy-staging.yml`'s job conclusion is failure whenever `T02`'s core-loop
step fails. This task may not modify anything in the `karsift-ai-infra`
repository - it is out of reach for this package (see `change.yaml`'s
`VOC-050-DEP-01`). This task's pull request must explicitly document, as an
open item for the reviewing human, that `karsift-ai-infra`'s `release.yml`
must separately be confirmed (or changed, in a companion package against
that repository) to actually gate its `develop` → `main` auto-promote
decision on this workflow's conclusion for the relevant commit - this task
does not claim to have proven that cross-repo behavior itself.

## VOC-050-T04 — Activate the production core-loop smoke check

- Requirement source: `specification.md`'s "Scope and non-goals" section;
  depends on `VOC-050-T00`, `VOC-050-T01`
- Acceptance criteria: `VOC-050-AC-05`
- Tests: `VOC-050-TEST-05`
- Evidence: `VOC-050-EV-04`
- Status: pending

Update `.github/workflows/deploy-production.yml`'s "Run production
smoke-test suite" step to set `SMOKE_TEST_SESSION_COOKIE` from `T01`'s
minting mechanism (invoked in a preceding step, following the file's
existing SSH-command pattern for other pre-deploy configuration writes).
Re-verify `infra/scripts/smoke-test-production.sh`'s core-loop section
(`# 5.`) against the synthetic account's actual seeded data (e.g. that
`GET /api/v1/journey-situations` genuinely returns 200 for this account),
adjusting the script only if a real mismatch is found - do not change its
no-side-effects design. Confirm with a live-observed run that the section's
output changes from `SKIP:` to `PASS:` lines.

Tasks preserve scope, separation of duties, and rollback safety. No task may
be dispatched before this package is adopted. `T00` must complete before
`T01`; `T01` must complete before `T02` and `T04`; `T02` must complete
before `T03`.
