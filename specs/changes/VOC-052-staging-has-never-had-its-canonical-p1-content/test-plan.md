# VOC-052 — Test Plan

## VOC-052-TEST-00 — Seed step runs and reports expected row counts on a fresh staging deploy

- Covers: `VOC-052-AC-00`
- Preconditions: A `deploy-staging.yml` run with `VOC-052-T00`'s step in place,
  triggered against a staging database that has the VOC-026 schema migrated but
  no P1 content rows yet (or has previously been seeded — idempotency is covered
  separately in `VOC-052-TEST-01`).
- Procedure: Trigger a staging deploy (a real merge to `develop`, or the
  workflow's own `workflow_dispatch` path if applicable). Inspect the workflow
  run log for the new seed step's output.
- Expected result: The step exits 0 and its log includes `apps/api/cmd/seed`'s
  own summary line (`seed applied: N situations, N words, N meanings, N
  examples, N notes, N journey_words`) with every count matching
  `voc026-p1.json`'s actual content counts (not zero).
- Evidence: `VOC-052-EV-00`

## VOC-052-TEST-01 — Seed step is idempotent across two consecutive deploys

- Covers: `VOC-052-AC-01`
- Preconditions: A staging database already seeded once by `VOC-052-T00`'s step.
- Procedure: Trigger a second staging deploy with no change to
  `voc026-p1.json` or the migration schema. Query the seeded tables' row counts
  and primary-key sets before and after the second run.
- Expected result: The second run's seed step exits 0, row counts are unchanged,
  no duplicate-key error occurs, and no row's primary key set differs from
  before the second run.
- Evidence: `VOC-052-EV-00`, `VOC-052-EV-01`

## VOC-052-TEST-02 — Real staging `/discover` renders journey-situation links

- Covers: `VOC-052-AC-02`
- Preconditions: `VOC-052-T00`'s step has run successfully against staging at
  least once.
- Procedure: Load `https://staging.vocanova.site/discover` (manually, or via
  `tests/staging-e2e/core-loop.staging.spec.ts`'s own navigation) and count
  rendered journey-situation links; open one and confirm it reaches at least one
  word.
- Expected result: At least one journey-situation link renders; opening it
  reaches at least one word — matching the spec's existing assertions at line
  239 and its neighboring steps.
- Evidence: `VOC-052-EV-01`

## VOC-052-TEST-03 — Full real staging core-loop E2E spec passes

- Covers: `VOC-052-AC-03`
- Preconditions: `VOC-052-T00`'s step is live in `deploy-staging.yml`.
- Procedure: Run `tests/staging-e2e/core-loop.staging.spec.ts` (as
  `deploy-staging.yml` already does, post-deploy) against real staging.
- Expected result: The spec passes in full — every step, not only the
  previously-failing discover step, since later steps in the same journey (e.g.
  reviewing or practicing a discovered word) also depend on real seeded content
  existing.
- Evidence: `VOC-052-EV-01`

## VOC-052-TEST-04 — A failing seed step fails the deploy without tearing down running containers (negative/failure test)

- Covers: `VOC-052-AC-04`
- Preconditions: A disposable, non-staging test environment (e.g. a local or
  ephemeral Postgres instance the seed step targets) rigged to make the seed
  step fail — e.g. by pointing `DATABASE_URL` at a database missing the VOC-026
  schema, so the seed's `INSERT` statements fail against nonexistent tables.
- Procedure: Run the seed step's exact command (extracted from the workflow
  step, not a paraphrase) against the rigged target and observe the exit code
  and workflow behavior in an isolated rehearsal (not against real staging or
  production).
- Expected result: The command exits non-zero; when run as an actual workflow
  step (no `continue-on-error`), the job fails before any subsequent
  `docker compose up -d` step runs, leaving previously-running containers
  untouched, consistent with `migrate.sh` and `seed-synthetic-smoke-user.sh`'s
  existing fail-closed behavior.
- Evidence: `VOC-052-EV-00`

## VOC-052-TEST-05 (conditional) — `deploy-production.yml` parity, only if `VOC-052-T02` is dispatched

- Covers: `VOC-052-AC-05`
- Preconditions: `VOC-052-DEP-02` resolved in favor of including `VOC-052-T02`;
  the equivalent step added to `deploy-production.yml`.
- Procedure: Verify the same class of evidence as `VOC-052-TEST-00`–`TEST-01`
  but for the production workflow, using a rehearsal or disposable target — not
  real production data or an unauthorized run against the live production
  database ahead of the founder's own separate activation decision for
  production's real-backend core-loop gating.
- Expected result: Same idempotent, fail-closed behavior as staging, confirmed
  without touching real production data or triggering an unauthorized
  production check.
- Evidence: `VOC-052-EV-02`

## Rollback coverage

Rolling back this package means reverting the added workflow step(s). Since the
seed step is additive and idempotent, no data rollback is required (see
`impact-analysis.md`'s "Data and migrations" section) — reverting the workflow
diff is sufficient and leaves any already-seeded rows in place harmlessly.

## Constraints

No test in this plan uses secrets or production data. `VOC-052-TEST-04` and
`VOC-052-TEST-05` explicitly require a disposable or rehearsal target, never real
staging or production, for their failure-mode and production-parity checks
respectively.
