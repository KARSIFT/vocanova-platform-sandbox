# VOC-053-EV-01 — T01 fix-attempt evidence

Evidence for `VOC-053-T01` (`VOC-053-AC-00`, `VOC-053-AC-01`,
`VOC-053-AC-02`). This is attempt 1 of T01. **Outcome: no code change
implemented. T01 cannot complete as scoped from this environment**, for the
reasons recorded honestly below. The implementer role's standing rule is
to "say so plainly instead of producing a partial or scope-expanded
change" when the task cannot be completed as scoped; this evidence file
is that record.

## TL;DR for the reviewer

Per `t00-evidence.md`'s Addendum 1 "T01 entry condition met" section,
T01 was authorized to proceed scoped to candidate (b) first (resolving
`VOC-053-DEP-01`), falling back to candidate (c) if (b)'s trace turned
up no defect. This attempt's complete independent re-trace of (b)'s
HTTP-handler → service → repository → pure-timezone path found **no
defect** that could explain the 2026-08-09 same-day decrease. Candidate
(c) (synthetic-account test-data interaction) **cannot be confirmed from
this environment** for the same reason T00 could not confirm it: the
live verification items in `t00-evidence.md` `findings/VOC-053-required-live-verification`
(items 2, 3, 5) require real staging database/network access that this
implementer session does not have. Per the implementer role's rule
against producing a partial or scope-expanded change, **no source change
is staged for this task**.

## What this attempt did

1. Re-read `apps/api/app/api/missions.go` (the live HTTP handler)
   end-to-end and confirmed T00's `findings/VOC-053-DEP-01` still
   describes the live code (no source change in that file since T00's
   base `509f952`).
2. Re-read `apps/api/business/missions/service.go`
   (`GetDailyMissionView`) end-to-end and confirmed:
   - `time.Now()` is called exactly once per request, in the handler,
     and the same `now` value is passed through the entire service call
     to `LocalDate`, `ReconcileAndAdvance`, and the lazy-creation
     block.
   - The lazy-creation `INSERT ... ON CONFLICT` preserves the existing
     row's `reviews_completed` on conflict (only `timezone`,
     `review_target`, `updated_at` are touched), and the
     `RETURNING` clause returns the existing value, not the INSERT's
     literal `0`. The repository is therefore safe under
     lazy-create races.
   - `loadStreakAndGrace`, `ListRecentCompletionHistory`, and the
     `ReconcileAndAdvance` write path only touch
     `streak_states` / `grace_day_ledger` / `streak_view` columns —
     none of them write to `daily_mission_snapshots.reviews_completed`.
3. Re-read `apps/api/business/missions/repository.go` end-to-end and
   confirmed the only writes to `reviews_completed` are
   `IncrementReviewsCompleted` (`LEAST(reviews_completed + 1,
   review_target)`, an additive cap) and `CreateDailyMissionSnapshot`
   (INSERT specifies `0`, ON CONFLICT preserves existing value). No
   `UPDATE daily_mission_snapshots` anywhere in
   `apps/api/business/missions/repository.go` or
   `apps/api/business/reviews/postgres.go` decrements
   `reviews_completed`.
4. Re-read `apps/api/business/accounts/postgres.go:463` — the only
   `UPDATE daily_mission_snapshots` outside the missions/reviews
   packages — and confirmed it reassigns `user_id` to the
   anonymization placeholder (account-deletion de-identification
   path). It does not touch `reviews_completed` and is not reached
   for the synthetic account (which is never deleted).
5. Re-read `apps/api/business/gamification/timezone.go`
   (`LocalDate`, `ResolveSettings`) end-to-end and confirmed both
   are pure functions of `(now, timezone)` and
   `(stored, clientTimezone)`. For the documented test conditions
   (synthetic account, no `user_settings` row, no client-supplied
   `timezone`), `LocalDate` returns a deterministic UTC-midnight
   `time.Time` and `ResolveSettings` returns
   `ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}`.
6. Confirmed no committed code path in `apps/api/` decrements
   `reviews_completed` — re-verifying the T00 finding.
7. Confirmed the only test-data seeding script
   (`apps/api/scripts/seed-synthetic-smoke-user.{sh,sql}`) only
   touches the `users` table, never `daily_mission_snapshots`.
8. Ran the local deterministic checks documented in
   `implementation-plan.md` (`bash scripts/governance/validate-governance.sh`,
   `bash scripts/governance/classify-change-risk.sh`,
   `git diff --check`) — all pass. The risk classifier reports
   `R1` because no source change is staged, which matches T00's
   finding that the actual fix's risk class is determined by the
   post-investigation file list (T00 reported the same).

## Why (b) cannot produce the observed decrease

A decrease in `daily_mission_snapshots.reviews_completed` requires
**one** of:

- A committed `UPDATE` or `INSERT ... ON CONFLICT DO UPDATE` that
  writes a value less than the current value of the column for the
  affected row.
- A `DELETE` of the row followed by an `INSERT` that defaults the
  column to `0`.

Neither happens in any committed code path. Specifically:

- `IncrementReviewsCompleted` is strictly additive (`LEAST(current
  + 1, target)`) and is only called by
  `apps/api/business/reviews/postgres.go:374` (the P2 review
  submission path) — and the test failure had `reviewedCards = 0`,
  so this path did not fire in the failing run.
- `CreateDailyMissionSnapshot` is the only path that ever creates a
  row, and it either inserts with `reviews_completed = 0` (a brand
  new row for a user/day that had no prior snapshot, where 0 is
  the correct value) or — on `ON CONFLICT` — preserves the
  existing value (the `ON CONFLICT DO UPDATE SET timezone,
  review_target, updated_at` clause does not include
  `reviews_completed`).
- The streak-reconciliation path in
  `ReconcileAndAdvance` → `ReconcileStreak` → `UpsertStreakState`
  only writes `streak_states` and `grace_day_ledger`, never
  `daily_mission_snapshots`.
- `MarkSnapshotCompleted` only writes `status` and
  `completed_at`.
- `MarkSnapshotProtected` only writes `status`,
  `grace_applied`, and `grace_day_id`.
- `MarkSnapshotMissed` is defined but has no caller anywhere in
  `apps/api/` outside the repository (grep-confirmed).
- The accounts de-identification `UPDATE daily_mission_snapshots
  SET user_id = $2 ...` reassigns ownership and does not touch
  `reviews_completed`, and is not reached for the synthetic
  account.

The two `UPDATE daily_mission_snapshots` statements in
`apps/api/business/missions/repository.go:265` and `:307` only
modify `new_words_completed` and
`sentence_practices_completed` respectively (both
`LEAST(COALESCE(...) + 1, COALESCE(target, 1))`, both additive).

So no committed code path explains the observed
`reviews_completed` decrease.

## Why (c) cannot be confirmed from this environment

T00 already established that candidate (c) — a test-data interaction
specific to the persistent synthetic account being reused across
deploy runs the same night — is the only candidate remaining
after (a) and the "fourth cause" (out-of-band CDN cache) were
ruled out by live evidence, and that (c) "cannot be ruled in or
out from this environment" because the live verification items 2,
3, and 5 in `t00-evidence.md` `findings/VOC-053-required-live-verification`
require direct database / DNS access this implementer session
does not have.

To confirm or rule out (c), the implementer would need:

- Item 2: a direct SQL read of the synthetic account's
  `daily_mission_snapshots` row at the exact instant of the
  step-7 read in a real staging run (and a corresponding
  step-2 read), to see whether the row was decremented in the
  DB itself or whether the API was reading a stale/cached value
  (the latter already ruled out by T00 addendum 1, so the
  remaining open question is the former).
- Item 3: a real staging run against a fresh synthetic
  account/day, to see whether the failure reproduces when no
  prior-run residue is in play (per issue #450's own
  first-thing-to-try).
- Item 5: a network/edge-cache check (e.g. `dig`, `host`, or
  `nslookup` of the staging origin) for any out-of-band cache
  layer not visible in the committed `infra/` configuration.

The addendum 2 dump in T00 covered item 2 in a non-failing run
state (a run where the queue was already empty and the row was
already `0`), so it cannot confirm or refute the decrease
mechanism. None of items 2-in-the-failing-state, 3, or 5 are
reproducible from an implementer session that only has read
access to this repository.

## What T01 cannot legitimately do in this environment

- **Cannot implement a candidate-(a) cache fix.** Candidate (a)
  is conclusively ruled out by T00 addendum 1 (live HTTP
  response headers showed `cache-control: private, no-cache,
  no-store, max-age=0, must-revalidate`,
  `x-nextjs-cache` absent, `cf-cache-status: DYNAMIC` on both
  step-1 and step-7 loads). Adding a `cache: "no-store"` to the
  `/home` `getDailyMission()` fetch is not just unnecessary — it
  would be a change whose only effect is to make the code
  defensively express what the platform is already proving
  empirically, which is exactly the "weakening a check to make
  the change pass" shape the implementer role is forbidden to
  produce.
- **Cannot implement a candidate-(b) backend-fix without a
  bug.** T00 and this attempt's re-trace both find no
  decrement path. "Hardening" the per-request `now`/timezone
  resolution to prevent a bug that the trace shows does not
  exist would be a speculative change — a fix targeting a
  defect that isn't there. The implementer role is not
  permitted to land a speculative change as if it were a real
  fix; that would mask the real cause by giving a test that
  happens to pass for the wrong reason, and would lock the
  repository into an unjustified design constraint.
- **Cannot implement a candidate-(c) test-data fix without a
  confirmed interaction.** T00 already established that (c)
  cannot be confirmed from this environment. The T01 task
  description's `(c)` clause — "If a test-data interaction
  is confirmed: fix the interaction (e.g. a concurrency-safety
  gap in how the persistent synthetic account's state is
  read/written across simultaneous or near-simultaneous deploy
  runs), not the reuse design itself" — is conditional on
  confirmation, and that confirmation requires the live
  items 2/3/5 above.

## Honest recommendation for the next step

The T00 addendum 1 "T01 entry condition met" decision was
premised on the second bullet — "the live verification has
narrowed the candidates enough that the implementer can pick
the highest-probability candidate and scope the fix to that,
recording the chosen candidate and why it was chosen over the
others in T01's own evidence file." This attempt picked the
highest-probability candidate (b), scoped the trace to it, and
recorded the trace outcome. The outcome is "no defect in (b)'s
code path," which — combined with T00's no-decrement-finding —
means there is no fix this implementer session can produce
without either (i) fabricating a bug that the trace shows
isn't there, or (ii) implementing a speculative "hardening"
change in place of a real fix.

The remaining live-verification items (`t00-evidence.md`
`findings/VOC-053-required-live-verification` items 2-failing-state,
3, and 5) are the smallest, most targeted next steps that could
move (b)/(c) from "inconclusive" to "confirmed." They require
the same class of staging access prior packages (VOC-050,
VOC-052) have used via `deploy-staging.yml`'s real-run path
and the synthetic-smoke-test session mint endpoint. Per the
founder-gate convention that has produced T00's addenda
already, the same operator path that captured the
cf-cache-status and daily_mission_snapshots row data could
capture the missing item-2/3/5 evidence next, and a follow-up
T01 (or a re-scoped T00-extension) could then implement the
fix against a confirmed root cause.

This is not a closure: T02's staging core-loop E2E step-7
verification (`VOC-053-AC-03`/`AC-04`) still depends on a real
fix landing, and the underlying issue #450 failure is still
reproducible in the addendum 1 evidence. The honest output here
is that T01 is blocked on the same live-verification gap that
T00 documented and that PR #460 (the previous T01 attempt, per
T00 addendum 2) also declined without. The proper disposition
is for the operator-run live-verification work to happen next,
not for a speculative source change to be staged as a
stand-in.

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Commands executed in this implementation run:

- `bash scripts/governance/validate-governance.sh` (pass;
  "Repository foundation validation passed. Governance
  structure validation passed.")
- `bash scripts/governance/classify-change-risk.sh` (pass;
  reports path-based floor `R1` because no source change is
  staged — same as T00 reported for the same reason).
- `git diff --check` (pass; no source diff, no whitespace
  errors).

## What changed since T00 (this attempt's base)

This attempt produced **no source change**, **no spec change**,
**no change to `tests/staging-e2e/core-loop.staging.spec.ts`**
(T01's explicit non-goal), and **no commit**. The only file
this attempt writes is this evidence file
(`specs/changes/VOC-053-staging-core-loop-e2e-words-reviewed-today/t01-evidence.md`),
uncommitted, in the working tree as the implementer role
requires.

The branch is `agent/voc-053-voc-053-t01`, based on `develop`
at `509f952a62dec24d827560e38b9f184019c3174c` (T00 addendum 2
merge). No code in
`apps/api/app/api/missions.go`,
`apps/api/business/missions/`,
`apps/api/business/gamification/`, or
`apps/web/` has been modified by this attempt.
