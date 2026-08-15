# VOC-082-T00 — ReconcileStreak fix evidence

**Task:** VOC-082-T00  
**Package:** VOC-082  
**Evidence ID:** VOC-082-EV-00  
**Date:** 2026-08-15  
**Reviewed revision:** working tree on branch `agent/voc-082-voc-082-t00`

## Root cause (confirmed, matches VOC-082-DEP-00)

`applyP4ReviewWiring` marks today's snapshot `completed`, then fetches streak
snapshots on the same transaction (including today as `completed`), then calls
`ReconcileAndAdvance(..., missionCompletedNow=true)`. `ReconcileStreak` selected
today via `mostRecentGood`, computed `gap = daysBetween(today, today) = 0`, and
returned `ErrInvalidStreakSnapshot` — rolling back snapshot completion, the +10
completion reward, and streak writes.

`apps/api/business/reviews/postgres.go` call order is correct; no call-site edit
was required.

## Fix

In `apps/api/business/gamification/streak.go`, when `currentCompletion=true`
and `lastGood` equals today, the function now anchors reconciliation against the
most recent good day **before** today (or `state.LastCompletedLocalDate` when
older snapshots are outside the fetch window). Genuinely future `lastGood` dates
remain rejected via the existing `gap <= 0` fail-closed path.

Added helper `mostRecentGoodBefore` to locate the prior anchor without treating
today's just-completed snapshot as invalid.

## Validation commands and results

```bash
bash scripts/governance/validate-governance.sh
# Repository foundation validation passed.
# Governance structure validation passed.

bash scripts/governance/classify-change-risk.sh
# Detected path-based risk floor: R1 (streak.go, streak_test.go, postgres_p4_test.go)

git diff --check
# (no whitespace errors)

cd apps/api && go test ./business/gamification/ ./business/reviews/ -count=1
# ok  github.com/KARSIFT/vocanova-platform/apps/api/business/gamification
# ok  github.com/KARSIFT/vocanova-platform/apps/api/business/reviews
```

## Test coverage mapping

| Test ID | Test | Result |
|---------|------|--------|
| VOC-082-TEST-00 | `TestReconcileStreakAdvancesWhenTodayAlreadyCompletedInSnapshots` | PASS |
| VOC-082-TEST-01 | `TestReconcileStreakFirstCompletionWhenTodayInSnapshots` | PASS |
| VOC-082-TEST-02 | `TestReconcileStreakRejectsFutureLastGood` | PASS |
| VOC-082-TEST-03 | `TestPostgreSQLRepositorySubmitReviewP4MissionCompletion` (updated: fetch returns today+yesterday completed, streak state 19→20 path) | PASS |
| VOC-082-TEST-05 | No VOC-081 monitor/shared-edge paths touched | PASS (scope isolation) |

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-082-AC-00 | PASS | Just-completed today no longer returns `ErrInvalidStreakSnapshot`; streak advances per consecutive rules |
| VOC-082-AC-01 | PASS | Future `lastGood` still rejected (`TestReconcileStreakRejectsFutureLastGood`) |
| VOC-082-AC-02 | PASS | Mission-completion sqlmock test commits through streak reconciliation with today in fetched snapshots |
| VOC-082-AC-04 | PASS | Diff limited to `streak.go`, `streak_test.go`, `postgres_p4_test.go`, and this evidence file |

## Staging / T01

Staging core-loop proof is **out of scope for T00** (`VOC-082-AC-03` /
`VOC-082-EV-01` deferred to VOC-082-T01 after merge to `develop`).

## Files changed

- `apps/api/business/gamification/streak.go` — fix + `mostRecentGoodBefore`
- `apps/api/business/gamification/streak_test.go` — regression and negative tests
- `apps/api/business/reviews/postgres_p4_test.go` — mission-completion harness mirrors post-mark fetch
- `specs/changes/VOC-082-fix-500-on-the-review-that-completes-today-s/t00-evidence.md` — this file
