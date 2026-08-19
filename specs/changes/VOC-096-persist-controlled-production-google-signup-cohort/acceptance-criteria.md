# VOC-096 — Acceptance Criteria

## VOC-096-AC-00 — Persistent production allowlist survives automatic deploys

- Requirement source: `VOC-096-D00`, `VOC-096-D01`
- Tasks: `VOC-096-T00`
- Tests: `VOC-096-TEST-00`, `VOC-096-TEST-01`
- Evidence: `VOC-096-EV-00`
- Result: pending

The repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` backs the production
controlled-cohort allowlist. `deploy-production.yml` reads this secret and syncs it
to production `api.env` on every deploy (push-triggered and manual dispatch). A later
normal push deploy preserves the cohort without operator intervention.

## VOC-096-AC-01 — Ephemeral dispatch input cannot silently erase the cohort

- Requirement source: `VOC-096-D02`
- Tasks: `VOC-096-T00`
- Tests: `VOC-096-TEST-02`
- Evidence: `VOC-096-EV-00`
- Result: pending

The `workflow_dispatch` input `new_user_signup_allowlist` is removed. An automatic
push deploy (no dispatch inputs) writes the secret-backed allowlist, not an empty
string.

## VOC-096-AC-02 — Fail closed on malformed production cohort configuration

- Requirement source: `VOC-096-D03`
- Tasks: `VOC-096-T00`
- Tests: `VOC-096-TEST-03`, `VOC-096-TEST-04`
- Evidence: `VOC-096-EV-00`
- Result: pending

When production Google OAuth is enabled and the persistent allowlist secret is
missing, empty, multiline, or malformed, the deploy fails closed (non-zero exit)
before runtime configuration is written or deployment converges. Only a non-sensitive
boolean readiness line is logged on success paths.

## VOC-096-AC-03 — Email addresses never in logs, synthetics, or evidence

- Requirement source: `VOC-096-D04`
- Tasks: `VOC-096-T00`, `VOC-096-T01`, `VOC-096-T02`
- Tests: `VOC-096-TEST-05`
- Evidence: `VOC-096-EV-00`, `VOC-096-EV-01`, `VOC-096-EV-02`
- Result: pending

No deploy log step, synthetic output, test fixture, or evidence file contains a real
email address. Only `@synthetic.vocanova.invalid` addresses appear in test fixtures.

## VOC-096-AC-04 — Production OAuth/readiness synthetic asserts controlled signup readiness

- Requirement source: `VOC-096-D06`
- Tasks: `VOC-096-T01`
- Tests: `VOC-096-TEST-06`, `VOC-096-TEST-07`
- Evidence: `VOC-096-EV-01`
- Result: pending

When production Google OAuth is enabled and `NEW_USER_SIGNUP_ENABLED=false`, the
production OAuth/readiness synthetic asserts that `/healthz` reports
`controlled_signup_ready: true`. The synthetic fails when the controlled cohort is
empty and controlled first-time signup is an expected production capability.

## VOC-096-AC-05 — Production /healthz exposes boolean readiness only

- Requirement source: `VOC-096-D04`, issue #809 AC-5/AC-6
- Tasks: `VOC-096-T01`, `VOC-096-T02`
- Tests: `VOC-096-TEST-08`
- Evidence: `VOC-096-EV-01`, `VOC-096-EV-02`
- Result: pending

After deployment, production `/healthz` reports `controlled_signup_ready: true` when
OAuth is enabled and the controlled cohort is valid. The response never exposes email
values or cohort cardinality.

## VOC-096-AC-06 — Canonical production OAuth start/callback behavior retained

- Requirement source: issue #809 AC-7
- Tasks: `VOC-096-T01`, `VOC-096-T02`
- Tests: `VOC-096-TEST-09`
- Evidence: `VOC-096-EV-01`, `VOC-096-EV-02`
- Result: pending

Production OAuth start still returns HTTP 200 with an `accounts.google.com` authorization
URL and the canonical production callback when OAuth is enabled. Extending the harness
for readiness does not weaken start/callback checks.

## VOC-096-AC-07 — Staging/production secret and tier isolation preserved

- Requirement source: issue #809 AC-9
- Tasks: `VOC-096-T00`
- Tests: `VOC-096-TEST-10`, `VOC-096-TEST-11`
- Evidence: `VOC-096-EV-00`
- Result: pending

`deploy-production.yml` consumes only `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`.
`deploy-staging.yml` consumes only `STAGING_NEW_USER_SIGNUP_ALLOWLIST`. Production and
staging databases, deploy users, Docker networks, shared-edge nginx ownership, and
monitoring isolation remain unchanged.

## VOC-096-AC-08 — Deterministic workflow/config tests cover persistence and failure modes

- Requirement source: issue #809 AC-8
- Tasks: `VOC-096-T00`, `VOC-096-T01`
- Tests: `VOC-096-TEST-00` through `VOC-096-TEST-11`
- Evidence: `VOC-096-EV-00`, `VOC-096-EV-01`
- Result: pending

Foundation tests lock push/manual secret persistence, malformed/empty fail-closed
behavior, boolean-only logging, production readiness harness behavior, and
environment secret isolation without using real secrets or production data.

## VOC-096-AC-09 — Operator procedure documented

- Requirement source: issue #809; `VOC-096-D02`
- Tasks: `VOC-096-T02`
- Tests: `VOC-096-TEST-12`
- Evidence: `VOC-096-EV-02`
- Result: pending

A secure operator procedure documents how to add/remove a production cohort member
(update the GitHub secret) and how a later normal push deploy preserves the cohort.

## VOC-096-AC-10 — Production deploy-and-verify evidence

- Requirement source: issue #809 evidence expectations
- Tasks: `VOC-096-T02`
- Tests: `VOC-096-TEST-13`
- Evidence: `VOC-096-EV-02`
- Result: pending

Evidence shows: successful production deployment from normal workflows; live
production `/healthz` with `controlled_signup_ready: true`; canonical OAuth
start/callback checks green; relevant scheduled synthetic verification green;
production route/content smoke green. No secrets, real identities, OAuth artifacts,
session material, or cohort counts appear in evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
