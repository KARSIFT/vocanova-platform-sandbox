# VOC-082 — Acceptance Criteria

## VOC-082-AC-00 — Just-completed today is a valid current completion

- Requirement source: issue #675 required fix; `VOC-082-D00`
- Tasks: `VOC-082-T00`
- Tests: `VOC-082-TEST-00`, `VOC-082-TEST-01`
- Evidence: `VOC-082-EV-00`
- Result: pending

Observable outcome:

1. When `ReconcileStreak` is called with `currentCompletion=true` and the
   snapshot list includes today already marked `completed` (as after
   `MarkSnapshotCompleted` in the same transaction), it does **not**
   return `ErrInvalidStreakSnapshot` solely because `lastGood` equals
   today.
2. Streak state advances (or starts) consistently with existing
   consecutive / first-completion rules using prior-day anchors as
   appropriate.
3. `POST /api/v1/reviews/submissions` for the review that first reaches
   `review_target` no longer fails with HTTP 500 from this defect.

## VOC-082-AC-01 — Future snapshots remain defensively rejected

- Requirement source: issue #675 required fix; `VOC-082-D01`
- Tasks: `VOC-082-T00`
- Tests: `VOC-082-TEST-02`
- Evidence: `VOC-082-EV-00`
- Result: pending

Observable outcome:

1. A snapshot (or `lastGood`) dated after today still causes
   `ReconcileStreak` to fail closed with `ErrInvalidStreakSnapshot` (or
   equivalent documented error).
2. The today+`currentCompletion=true` fix does not weaken that guard.

## VOC-082-AC-02 — Completing review commits snapshot, reward, and streak atomically

- Requirement source: issue #675 required fix
- Tasks: `VOC-082-T00`
- Tests: `VOC-082-TEST-00`, `VOC-082-TEST-03`
- Evidence: `VOC-082-EV-00`
- Result: pending

Observable outcome:

1. Deterministic coverage demonstrates that when the review that meets
   today's target succeeds, today's snapshot ends `completed`, the
   daily-mission completion reward path can run, and streak
   reconciliation proceeds without forcing a transaction rollback from
   `ErrInvalidStreakSnapshot`.
2. Existing idempotency of completion rewards / streak ledger keys is
   not weakened.

## VOC-082-AC-03 — Staging core-loop re-run succeeds through mission completion

- Requirement source: issue #675 verification requirement
- Tasks: `VOC-082-T01`
- Tests: `VOC-082-TEST-04`
- Evidence: `VOC-082-EV-01`
- Result: pending

Observable outcome:

1. After T00 merges to `develop`, a real `deploy-staging.yml` run of
   `tests/staging-e2e/core-loop.staging.spec.ts` completes successfully
   through the review that reaches the synthetic account's daily target
   (no HTTP 500 on that submission from this defect).
2. Evidence records the run URL and, where available, that today's
   snapshot is completed (or equivalent Home/mission progress shows
   target reached) rather than stuck at target-1/`open`.
3. Inventing a green run without a URL is a FAIL.

## VOC-082-AC-04 — Scope stays isolated from VOC-081

- Requirement source: issue #675; `VOC-082-D02`; `VOC-082-DEP-01`
- Tasks: `VOC-082-T00`, `VOC-082-T01`
- Tests: `VOC-082-TEST-05`
- Evidence: `VOC-082-EV-00`, `VOC-082-EV-01`
- Result: pending

Observable outcome:

1. Task diffs do not change VOC-081 monitor / shared-edge / monitoring
   Compose / Cloudflare paths.
2. Evidence explicitly notes isolation from VOC-081.
