# VOC-052 — Tasks

## VOC-052-T00 — Build and run `apps/api/cmd/seed` against real staging after migrations apply

- Requirement source: `specification.md` scope item 1; issue #437's "suggested
  direction"
- Acceptance criteria: `VOC-052-AC-00`, `VOC-052-AC-01`, `VOC-052-AC-04`
- Tests: `VOC-052-TEST-00`, `VOC-052-TEST-01`, `VOC-052-TEST-04`
- Evidence: `VOC-052-EV-00`
- Status: pending

Add a step to `.github/workflows/deploy-staging.yml`, placed immediately after the
existing `migrate.sh` invocation and its adjacent `seed-synthetic-smoke-user.sh`
call (see the workflow's own SSH step, around the migration-apply and
synthetic-user-seed sequence), that builds and runs `apps/api/cmd/seed` against
the real staging database with `DATABASE_URL` set the same way `migrate.sh`
already resolves it (the private Postgres bridge IP, not the Compose-internal
`postgres` hostname). The implementer must first resolve
`specification.md`'s open question 1 (build-on-runner-and-SCP vs.
run-from-runner-directly) and record which mechanism was chosen and why in this
task's own evidence file, rather than picking one silently. The step must fail
the workflow run (no `continue-on-error`) if the seed fails, before `docker
compose up -d` runs, mirroring the existing fail-closed pattern this workflow
already uses for `migrate.sh` and `seed-synthetic-smoke-user.sh` — see
`specification.md` open question 3 for why this is proposed as the default
rather than assumed without comment. Preserve every existing step in the
workflow unchanged; this is an additive step, not a rewrite of the migration or
deploy sequence.

## VOC-052-T01 — Verify the real staging core-loop E2E check passes against seeded content

- Requirement source: issue #437's reported failure; `specification.md` scope
  item 2
- Acceptance criteria: `VOC-052-AC-02`, `VOC-052-AC-03`
- Tests: `VOC-052-TEST-02`, `VOC-052-TEST-03`
- Evidence: `VOC-052-EV-01`
- Status: pending — depends on `VOC-052-T00` landing and a real staging deploy
  running with it in place

No source change is expected in this task. Its job is to produce and record
verification evidence: after `VOC-052-T00`'s step is live in a real
`deploy-staging.yml` run (triggered by that PR's own merge to `develop`, or by a
subsequent `develop` push), confirm
`tests/staging-e2e/core-loop.staging.spec.ts` passes in full against real
staging, specifically the previously-failing discover-step assertion
(`situationLinks.count()` returning a positive value) and every later step in
the same spec that depends on real word/meaning content. If the check does not
pass on the first real run with `VOC-052-T00`'s step in place, this task's scope
includes diagnosing and fixing that specific gap (e.g. a mismatch between the
seed content and what the discover page or spec expects) — but only that gap,
not unrelated changes.

## VOC-052-T02 (conditional) — Mirror the seed step in `deploy-production.yml`

- Requirement source: issue #437's "suggested direction" section (explicitly
  described there as "presumably" needed, not certain); `specification.md` open
  question 2
- Acceptance criteria: `VOC-052-AC-05`
- Tests: `VOC-052-TEST-05`
- Evidence: `VOC-052-EV-02`
- Status: pending — **not to be dispatched unless the adopting human explicitly
  resolves `VOC-052-DEP-02` in favor of including this task's scope now.** If
  the human instead decides this should wait until production's real-backend
  core-loop smoke gating is separately activated, this task should be closed as
  out-of-scope-for-now rather than implemented, and that decision recorded in
  this package's adoption evidence.

If dispatched, add the same class of step to `.github/workflows/deploy-production.yml`
that `VOC-052-T00` added to `deploy-staging.yml`, in the equivalent position
relative to that workflow's own migration-apply and
`seed-synthetic-smoke-user.sh` sequence, using the same build/run mechanism
`VOC-052-T00` chose (for consistency, unless the implementer records a specific
reason production needs a different mechanism). Verification must not use real
production user data or trigger the check against production before the
founder's own separate decision to activate production's real-backend core-loop
gating — see `test-plan.md`'s constraint on this.
