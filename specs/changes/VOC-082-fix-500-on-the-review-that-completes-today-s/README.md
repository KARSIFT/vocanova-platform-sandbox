# VOC-082 — Fix 500 on the review that completes today's mission

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #675](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/675),
prepared for plan review and adoption under **active A-004** (proposed **R2**).

## Identity and lifecycle

- Package ID: VOC-082
- Title: Fix 500 on the review that completes today's mission
- Canonical path:
  `specs/changes/VOC-082-fix-500-on-the-review-that-completes-today-s`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R2` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R1** for `apps/api/business/gamification/*` and related API/test paths)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`, `implementation.authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#675](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/675)
    (this package's requirement source)
- Related but distinct packages:
  - [VOC-065](../VOC-065-real-backend-write-path-bug-reviews-completed)
    — P4 wiring so `reviews_completed` increments; predecessor, not the
    completion-day 500
  - [VOC-081](../VOC-081-route-monitor-vocanova-site-through-the)
    — monitor.vocanova.site / shared-edge; **out of scope** (issue #675)

## Why this exists

Staging core-loop fails deterministically when the synthetic account
submits the review that reaches its daily target (20/20). On deploy run
[31886780600](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31886780600),
`POST /api/v1/reviews/submissions` returned HTTP 500 after clicking
`Good`. After transaction rollback, today's
`daily_mission_snapshots` row stayed at `reviews_completed=19`,
`review_target=20`, `status=open`, with no daily-mission completion
ledger entry.

Issue #675 verified the code path:

1. `applyP4ReviewWiring` marks today complete when the review hits target.
2. It then fetches streak snapshots inside the same transaction (today is
   already `completed` in that list).
3. It calls `ReconcileAndAdvance(..., currentCompletion=true)`.
4. `ReconcileStreak` picks the newly completed today snapshot as
   `lastGood`, computes `gap <= 0`, and returns
   `ErrInvalidStreakSnapshot`.

That error aborts the transaction, so the completing review never commits.

## What this package does

1. **Fix streak reconciliation + regression tests** (`VOC-082-T00`): treat a
   just-completed today snapshot as the current completion when
   `currentCompletion=true`; keep rejecting genuinely future snapshots;
   add a unit regression that includes today as completed in the fetched
   list with `currentCompletion=true`; prove (via deterministic tests
   and/or transactional coverage) that the completing review can commit
   snapshot completion, completion reward, and streak atomically.
2. **Verify on real staging** (`VOC-082-T01`): re-run the staging
   core-loop successfully through the review that reaches today's target,
   with evidence that the 20th review no longer 500s and the day
   completes.

## What this package deliberately does NOT do

- Not VOC-081 monitor / shared-edge / Cloudflare work.
- Not historical backfill of stuck 19/20 snapshots (forward-fix only
  unless adoption expands scope — see open questions).
- Not weakening staging core-loop gates or inventing a green staging run.
- Not adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed **R2** (path floor R1; semantic elevation for
   every-user mission-completion 500), or raise to R3.
2. Whether any stuck staging/production snapshots already at target-1
   need an ops reset (default: out of scope; forward-fix only).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R2 still requires staged evidence and rollback
credibility. `automatic_merge_allowed: true` is set per AGENTS.md. This
draft still carries no adoption or implementation authority.
