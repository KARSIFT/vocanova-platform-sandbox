# VOC-082 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-082-T00 — Fix ReconcileStreak for just-completed today + regression tests

- Requirement source: issue #675 required fix; `VOC-082-D00`;
  `VOC-082-D01`; `VOC-082-DEP-00`
- Acceptance criteria: `VOC-082-AC-00`, `VOC-082-AC-01`, `VOC-082-AC-02`,
  `VOC-082-AC-04`
- Tests: `VOC-082-TEST-00`, `VOC-082-TEST-01`, `VOC-082-TEST-02`,
  `VOC-082-TEST-03`, `VOC-082-TEST-05`
- Evidence: `VOC-082-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. In `apps/api/business/gamification/streak.go`, change the `gap <= 0`
   handling so that when `currentCompletion` is true and `lastGood`
   equals today, the function continues with the current-completion
   reconciliation path instead of returning
   `ErrInvalidStreakSnapshot`. Keep fail-closed rejection when
   `lastGood` is strictly after today.
2. Add a regression unit test that includes today as `completed` in the
   fetched snapshot list with `currentCompletion=true` (mirroring
   `applyP4ReviewWiring`'s mark-then-fetch order) and asserts success
   and expected streak advancement / first-completion behavior.
3. Add or update a negative test that a future `lastGood` / future
   completed snapshot still returns `ErrInvalidStreakSnapshot`.
4. Where the existing reviews/gamification test harness allows without
   unrelated refactors, add coverage that the completing-review write
   path can commit snapshot completion + completion reward + streak
   without rolling back on this error. If a full transactional test is
   not feasible in-repo, document that limitation in `t00-evidence.md`
   and rely on pure unit coverage plus T01 staging proof — do not invent
   harness capability.
5. Re-read `applyP4ReviewWiring` in
   `apps/api/business/reviews/postgres.go`. Edit it only if evidence
   shows a call-site defect beyond `ReconcileStreak`. Prefer the
   minimal streak.go fix named by issue #675.
6. Record commands, results, and AC mapping in `t00-evidence.md`.

### Explicitly out of scope for this task

- Staging live proof (T01).
- VOC-081 monitor / shared-edge / Cloudflare changes.
- Historical backfill migration for stuck 19/20 rows (unless adoption
  expanded open question 2).
- Weakening core-loop E2E gates.

## VOC-082-T01 — Verify staging core-loop through mission completion

- Requirement source: issue #675 verification requirement
- Acceptance criteria: `VOC-082-AC-03`, `VOC-082-AC-04`
- Tests: `VOC-082-TEST-04`, `VOC-082-TEST-05`
- Evidence: `VOC-082-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-082-T00` merging to `develop`

### Required work

1. After T00 is on the staging deploy path, record a real
   `deploy-staging.yml` (or equivalent repository staging deploy) run of
   `tests/staging-e2e/core-loop.staging.spec.ts`.
2. Confirm the journey succeeds through the review that reaches today's
   target without HTTP 500 from this streak defect.
3. Where diagnostics are available, record that today's snapshot is
   completed (not stuck at target-1/`open`) and that completion reward /
   progress reflects success.
4. If the run fails, record the run URL and honest diagnosis; do not
   claim AC-03. Narrow remediation belongs in a follow-up task or
   package if the residual cause is outside T00's accepted fix — do not
   silently expand into VOC-081.

### Explicitly out of scope for this task

- Inventing a green staging run without evidence.
- VOC-081 work.
- Production manual SSH fixes.

## Task ordering notes

- T00 blocks T01.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #675 is gated on AC results with evidence, not on task
  issue closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
