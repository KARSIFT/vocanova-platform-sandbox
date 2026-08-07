# VOC-046 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and
implementation is authorized, per this repository's `AGENTS.md` ("a chat
prompt or issue alone is not implementation authority"). All thirteen files
this package's audit touches write to tables with existing production data
— see `specification.md`'s risk section and `impact-analysis.md`'s data
section. None are known at drafting time to have any in-flight conflicting
work.

## File reconciliation and implementation sequence

Existing targets confirmed at drafting time (read in full or via targeted
grep):

- `apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`
  (lines 147-153 at drafting time) — the confirmed defect, this issue's own
  crash.
- The same file's four other `INSERT INTO daily_activity_summaries`
  statements (lines 215-221, 251-256, 293-298, 333-338, 363-368 at drafting
  time) and one further `INSERT INTO daily_mission_snapshots` statement for
  the missed-mission path (lines 450-456 at drafting time) — all confirmed
  present via a targeted grep at drafting time; each needs its own
  cross-check against `created_at`/`updated_at`, since some already carry
  an `ON CONFLICT DO UPDATE` clause (which does not help the first-insert
  branch, the same defect shape as the confirmed crash).
- The twelve other files named in `specification.md`'s in-scope list — not
  yet individually read at drafting time; `VOC-046-T02`'s own first step is
  to read each one and enumerate its `INSERT INTO` statements before fixing
  anything, consistent with this repository's "read every file" expectation
  for template-defining or scope-defining work.
- Test files to be identified by the implementer per `test-plan.md`, likely
  alongside each affected package's existing test suite (e.g.
  `apps/api/business/missions/repository_test.go` if it exists, or a new
  file following that package's convention).
- `apps/api/migrations/` — read-only for this package's primary fix (no
  migration file is expected to change); `VOC-046-T03`, if adopted, may add
  a new test file that reads migration files but does not modify them.

Ordered steps:

1. `VOC-046-T00`: add `created_at, updated_at` to
   `CreateDailyMissionSnapshot`'s `daily_mission_snapshots` INSERT column
   list and VALUES. Verify against a real Postgres instance that the fresh
   insert no longer raises a `NOT NULL` violation, and that
   `GET /api/v1/daily-mission` no longer returns a `500` for a fresh user.
   Dispatch first, given confirmed `P0` production impact.
2. `VOC-046-T01`: check and fix every `daily_activity_summaries` INSERT in
   the same file, following the same pattern.
3. `VOC-046-T02`: read each of the twelve remaining named files, enumerate
   every `INSERT INTO` statement, cross-reference against
   `apps/api/migrations/`'s `NOT NULL`-no-`DEFAULT` columns, fix every
   defect found, and record the complete audit trail as evidence.
4. `VOC-046-T03` (conditional): only if the reviewing human confirms this is
   in scope at adoption, add the schema-scanning detection check.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower apps/api-relevant script (e.g. lint/typecheck/test)
```

Plus this package's own `VOC-046-TEST-00` through `TEST-03` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this
specification, confirm `VOC-046-AC-00` through `AC-04` (where applicable)
are each satisfied with real evidence (not asserted), confirm no fix altered
any existing `ON CONFLICT DO UPDATE` branch's behavior for a row that
already exists, confirm `VOC-046-T02`'s audit trail is genuinely complete
(spot-check at least a few of the thirteen files' actual `INSERT INTO`
statements against the recorded outcome, not just trust the recorded
summary), and confirm no self-approval occurred.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. The
fix takes effect for the `apps/api` service the next time it is
built/deployed after merge — implementer/reviewer to confirm the exact
deploy/redeploy mechanism against this repo's existing `apps/api` release
process. No migration is required for this package's primary fix, so no
migration-apply step is part of this deployment.

Rollback trigger: any fixed call site still raises the `NOT NULL`
violation on fresh insert, or any fix altered an existing `ON CONFLICT DO
UPDATE` branch's behavior for a row that already exists.

Rollback mechanism: revert the code change — expected to be a small,
self-contained diff confined to application code and new tests, no
migration or irreversible state change, consistent with VOC-045's rollback
precedent. Rows successfully created after the fix (before any rollback)
remain valid and are not affected by reverting the code.

Owner: named explicitly in the implementation PR at deploy time, not left
implicit.
