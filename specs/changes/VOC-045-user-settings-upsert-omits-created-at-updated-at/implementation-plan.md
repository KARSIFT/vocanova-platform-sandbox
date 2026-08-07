# VOC-045 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and
implementation is authorized, per this repository's `AGENTS.md` ("a chat
prompt or issue alone is not implementation authority"). `apps/api/business/users/postgres.go`
and `apps/api/business/gamification/repository.go` both write to
`user_settings`, a table with existing production data — see
`specification.md`'s risk section and `impact-analysis.md`'s data section.
Neither has any known in-flight conflicting work at drafting time.

## File reconciliation and implementation sequence

Existing targets (both read in full at drafting time):
- `apps/api/business/users/postgres.go`'s `CompleteOnboarding` (lines 97-179
  at drafting time) — already has a `now time.Time` parameter in scope,
  already uses it for two other statements in the same transaction.
- `apps/api/business/gamification/repository.go`'s `UpsertUserSettings`
  (lines 156-190 at drafting time) — has no `time.Time` parameter in scope;
  its own `ON CONFLICT DO UPDATE` clause already uses SQL `NOW()` directly.
- `apps/api/business/gamification/service.go` (line 58's call site) — only
  needs to change if `VOC-045-T01` resolves open question 1 by threading a
  new parameter through `UpsertUserSettings`, not if it uses SQL `NOW()`
  directly.
- A test file to be identified by the implementer per `test-plan.md` and
  `tasks.md`'s `VOC-045-T02` (e.g. new or existing tests alongside
  `apps/api/business/users/postgres.go` and
  `apps/api/business/gamification/repository.go`, and/or
  `apps/api/migrations/migration_test.go` if integration-style tests already
  live there).

No other file is expected to require changes for this package's primary
fix, pending the reviewing human's decision on `specification.md`'s open
questions 2 and 3.

Ordered steps:

1. `VOC-045-T00`: add `created_at, updated_at` to `CompleteOnboarding`'s
   `user_settings` INSERT column list and VALUES, using the existing `now`
   parameter (e.g. as an additional bound parameter alongside the existing
   `$1, $2, $3` — following this same function's own established pattern of
   reusing one `now` value across multiple statements in the transaction).
   Verify by confirming a fresh insert no longer raises the `NOT NULL`
   violation.
2. `VOC-045-T01`: resolve open question 1 (SQL `NOW()` vs. a threaded `now`
   parameter), then add `created_at, updated_at` to `UpsertUserSettings`'s
   INSERT column list and VALUES accordingly. If a parameter is threaded,
   update `apps/api/business/gamification/service.go`'s call site and any
   other caller. Verify by confirming a fresh insert (no prior row) no
   longer raises the `NOT NULL` violation and returns the expected row.
3. `VOC-045-T02`: add the regression test(s) covering both fixed call sites'
   fresh-insert path (failing-first against pre-fix code, passing after),
   and run the repository-wide `INSERT INTO user_settings` grep, recording
   the full call-site list as evidence.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower apps/api-relevant script (e.g. lint/typecheck/test)
```

Plus this package's own `VOC-045-TEST-00`/`01`/`02` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this
specification, confirm `VOC-045-AC-00` through `AC-03` are each satisfied
with real evidence (not asserted), confirm neither fix altered the existing
`ON CONFLICT DO UPDATE` branches' `daily_review_target`/`timezone`
preservation logic, confirm the call-site inventory (`VOC-045-EV-03`) is
complete, and confirm no self-approval occurred.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. The fix
takes effect for the `apps/api` service the next time it is built/deployed
after merge — implementer/reviewer to confirm the exact deploy/redeploy
mechanism against this repo's existing `apps/api` release process. No
migration is required for this package's primary fix, so no migration-apply
step is part of this deployment (see `specification.md`'s open question 2
for the explicitly-deferred schema-level alternative).

Rollback trigger: the fix does not actually resolve the `NOT NULL`
violation for a genuine fresh insert, or introduces an unintended change to
either upsert's `ON CONFLICT` update behavior for an existing row.

Rollback mechanism: revert the code change (a small, self-contained diff
confined to `apps/api/business/users/postgres.go`,
`apps/api/business/gamification/repository.go`, possibly
`apps/api/business/gamification/service.go` depending on open question 1's
resolution, plus the new tests) — no migration or irreversible state change
is introduced by the primary fix, so rollback is expected to be low-risk and
fast, consistent with VOC-039/VOC-041/VOC-042/VOC-043's small-diff rollback
precedent. Rows successfully created after the fix (before any rollback)
remain valid and are not affected by reverting the code.

Owner: named explicitly in the implementation PR at deploy time, not left
implicit.
