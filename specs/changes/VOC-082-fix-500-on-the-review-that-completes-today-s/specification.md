# VOC-082 — Fix 500 on the review that completes today's mission: Specification

## Objective and requirement source

Close the defect reported in
[GitHub issue #675](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/675):
the staging core-loop fails deterministically when the synthetic account
submits the review that reaches its daily target (20/20). On deploy run
[31886780600](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31886780600),
`POST /api/v1/reviews/submissions` returned HTTP 500 after clicking
`Good`. After rollback, today remained `reviews_completed=19`,
`review_target=20`, `status=open`, with no daily-mission completion
ledger entry.

Primary context (issue #675 and drafting-time repo read):

| Item | Value |
|------|--------|
| Failing request | `POST /api/v1/reviews/submissions` → HTTP 500 |
| Live DB after rollback | `reviews_completed=19`, `review_target=20`, `status=open` |
| Completing path | `applyP4ReviewWiring` → `MarkSnapshotCompleted` → fetch snaps → `ReconcileAndAdvance(..., currentCompletion=true)` |
| Pure-function failure | `ReconcileStreak` selects completed today as `lastGood`, `gap <= 0` → `ErrInvalidStreakSnapshot` |
| Predecessor | [VOC-065](../VOC-065-real-backend-write-path-bug-reviews-completed) (P4 wiring; increments work until completion day) |
| Explicit non-scope | [VOC-081](../VOC-081-route-monitor-vocanova-site-through-the) monitor routing |

**Objective:** after this package's implementation, the review that first
reaches today's `review_target` commits successfully (no HTTP 500 from
this streak defect), today's snapshot is `completed` with a completion
reward ledger entry, streak reconciliation advances under
`currentCompletion=true` even when today's completed snapshot is already
present in the fetched list, genuinely future snapshots remain rejected,
and the staging core-loop re-run passes through that completing review.

## Confirmed findings (issue #675 + drafting-time re-read)

- In `apps/api/business/reviews/postgres.go`, `applyP4ReviewWiring` marks
  today's snapshot completed when `newReviewsCompleted >= review_target`
  and status was `open`, then fetches streak snapshots on the same
  transaction, then calls `ReconcileAndAdvance` with
  `missionCompletedNow` (passed as `currentCompletion`).
- In `apps/api/business/gamification/streak.go`, after locating
  `lastGood` via `mostRecentGood(snapshots)`, `daysBetween(today,
  *lastGood)` returns `0` when `lastGood` is today (because
  `daysBetween` treats non-strictly-after as `0`). The `gap <= 0`
  branch then returns `ErrInvalidStreakSnapshot` as a defensive
  "future lastGood" guard — but it also fires for **today**, which is
  exactly the in-transaction state when `currentCompletion=true`.
- Existing unit tests advance streaks with yesterday (and older) in the
  snapshot list; they do **not** include today-as-completed alongside
  `currentCompletion=true`, so the staging failure was unguarded.
- Rule (a) already no-ops when today is completed and
  `currentCompletion=false`. The missing case is today completed **and**
  `currentCompletion=true` (just completed in this call).

## Scope and non-goals

In scope:

1. **Fix `ReconcileStreak` (T00):** when `currentCompletion` is true and
   `lastGood` equals today, treat that as the current completion path
   (advance/reconcile per existing consecutive/first-completion rules
   using yesterday or prior anchors as appropriate) — not as
   `ErrInvalidStreakSnapshot`. Preserve defensive rejection of
   genuinely future `lastGood` dates (`lastGood` after today).
2. **Regression tests (T00):** add a deterministic test that includes
   today as `completed` in the fetched snapshot list with
   `currentCompletion=true` and expects success (streak advances /
   first-completion behavior as applicable). Add or retain a negative
   case that still rejects a future snapshot. Cover atomic success of
   the completing-review write path via pure unit tests and, where
   existing review/repo test harness allows without expanding scope,
   transactional coverage that snapshot completion + completion reward +
   streak write succeed together when the target is met.
3. **Staging verification (T01):** re-run
   `tests/staging-e2e/core-loop.staging.spec.ts` on a real
   `deploy-staging.yml` path after T00 merges; record PASS evidence
   through the review that completes today's mission (no 500; day
   completes). Honest FAIL with run URL is acceptable evidence of
   unfinished work — never invent a green run.

Non-goals / explicitly excluded:

- VOC-081 monitor.vocanova.site / shared-edge / Cloudflare / monitoring
  Compose changes.
- Schema migrations or historical backfill of stuck 19/20 rows (default;
  see open question 2).
- Unrelated review-UI, seed, or deploy-workflow redesign.
- Snapshot-then-recheck-drift promotion tasks (not applicable).
- Adopting or authorizing this package from within the draft.

## Risk and protected areas

Builder assessment: expected paths are under `apps/api/business/` and
staging E2E tests. Drafting-time
`scripts/governance/classify-change-risk.sh --files-from` against the
expected list reported **Detected path-based risk floor: R1**.

This package **proposes R2** for the change as a whole because the
defect is a production/staging user-facing core-loop correctness failure
(mission completion + streak) for every learner who hits today's target,
and release requires staging evidence plus rollback credibility. This is
a **draft proposal for the reviewing human at adoption time, not a
determination**. The independent verifier may raise to R3 if semantic
review warrants it.

Protected areas: none of auth, secrets, migrations, or production
infra are in the default file list. Gamification ledger / mission
snapshot correctness is still high-value product state — tests must not
weaken idempotency of completion rewards.

Under **active A-004**, engineering-workflow gates require no founder
`approved` comment. EHR is not triggered by this drafting pass.

## Decisions, contradictions, security, and privacy

`VOC-082-D00` (recorded for traceability; formal acceptance at
adoption): A just-completed today snapshot present in the fetched
snapshot list while `currentCompletion=true` is a valid current
completion, not an invalid prior anchor. `ReconcileStreak` must not
return `ErrInvalidStreakSnapshot` solely because `lastGood` equals
today in that case.

`VOC-082-D01`: Genuinely future `lastGood` dates (after today) remain
defensively rejected with `ErrInvalidStreakSnapshot` (or equivalent
fail-closed error).

`VOC-082-D02`: This package's patch stays isolated from VOC-081 monitor
routing (`VOC-082-DEP-01`).

Open questions for the reviewing human:

1. **Risk.** Accept proposed R2 (path floor R1), or elevate to R3 given
   every-user mission-completion impact.
2. **Stuck rows.** Default scope is forward-fix only. Should staging
   (or production) accounts already stuck at target-1/`open` receive an
   ops reset after deploy? If yes, record as a separate ops action or
   expand scope explicitly at adoption — do not invent a migration in
   T00 by default.

Security / privacy: no new PII surfaces; no secrets; no auth change.
Idempotency of daily-mission completion point grants and streak
ledger writes must remain intact (existing unique/idempotency keys).

## Data, migrations, analytics, and accessibility

- **Data / migrations:** No schema migration anticipated. Forward-fix
  only for in-flight stuck snapshots unless adoption expands open
  question 2. Completing-review writes remain single-transaction as
  today (`applyP4ReviewWiring` before commit).
- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** None expected — API/gamification correctness only;
  no product UI change required for the default fix path.
