# VOC-052 — Acceptance Criteria

## VOC-052-AC-00 — Staging deploy runs the P1 content seed after migrations apply

- Requirement source: issue #437, `specification.md` scope item 1
- Tasks: `VOC-052-T00`
- Tests: `VOC-052-TEST-00`
- Evidence: `VOC-052-EV-00`
- Result: pending

`deploy-staging.yml` builds and runs `apps/api/cmd/seed` against the real staging
database, after `migrate.sh` applies pending migrations and before (or as part of)
`docker compose up -d`, on every run that reaches that point in the workflow (not
gated behind a manual dispatch flag or opt-in input).

## VOC-052-AC-01 — Seed step is idempotent across repeated deploys

- Requirement source: issue #437 ("safe to rerun" property of `apps/api/cmd/seed`)
- Tasks: `VOC-052-T00`
- Tests: `VOC-052-TEST-01`
- Evidence: `VOC-052-EV-00`, `VOC-052-EV-01`
- Result: pending

Running the seed step on two consecutive staging deploys (with no seed-content
change between them) produces the same row counts and the same primary-key set
both times — no duplicate rows, no error on the second run.

## VOC-052-AC-02 — Real staging `/discover` page renders journey-situation links

- Requirement source: issue #437's reported failure
- Tasks: `VOC-052-T00`, `VOC-052-T01`
- Tests: `VOC-052-TEST-02`
- Evidence: `VOC-052-EV-01`
- Result: pending

After the seed step runs, `https://staging.vocanova.site/discover` renders at
least one journey-situation link, and opening one reaches at least one word —
matching `tests/staging-e2e/core-loop.staging.spec.ts`'s existing discover-step
assertions.

## VOC-052-AC-03 — Real staging core-loop E2E check passes end-to-end

- Requirement source: issue #437 (the specific failing check)
- Tasks: `VOC-052-T00`, `VOC-052-T01`
- Tests: `VOC-052-TEST-03`
- Evidence: `VOC-052-EV-01`
- Result: pending

`tests/staging-e2e/core-loop.staging.spec.ts`, run by `deploy-staging.yml` after a
real deploy with the new seed step in place, passes in full — not just at the
previously-failing discover step, since the seeded content also feeds any later
step in the same journey that depends on real words/meanings existing.

## VOC-052-AC-04 — A seed-step failure fails the staging deploy without tearing down running containers

- Requirement source: `specification.md` open question 3 (proposed fail-closed
  default, mirroring `seed-synthetic-smoke-user.sh`)
- Tasks: `VOC-052-T00`
- Tests: `VOC-052-TEST-04`
- Evidence: `VOC-052-EV-00`
- Result: pending

If the seed step exits non-zero, the workflow run fails before `docker compose up
-d` runs, and the previously-running containers (if any) remain untouched — the
same fail-closed shape `deploy-staging.yml` already uses for `migrate.sh` and
`seed-synthetic-smoke-user.sh`.

## VOC-052-AC-05 (conditional) — Production deploy parity, only if `VOC-052-DEP-02` resolves to "in scope now"

- Requirement source: issue #437's "suggested direction" section;
  `specification.md` open question 2
- Tasks: `VOC-052-T02`
- Tests: `VOC-052-TEST-05`
- Evidence: `VOC-052-EV-02`
- Result: pending — conditional; not evaluated unless the reviewing human
  resolves `VOC-052-DEP-02` in favor of including T02 in this package's adopted
  scope

If adopted, `deploy-production.yml` gains the same seed step in the same
placement and idempotency posture as T00, and this criterion's evidence must show
it exercised against a real (non-production, or explicitly authorized
production-safe) target before being claimed as passing — see `test-plan.md` for
the constraint that tests must not use production data.
