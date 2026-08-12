# VOC-065-T02 — Staging verification evidence (attempt 2)

**Task:** VOC-065-T02  
**Package:** VOC-065  
**Evidence ID:** VOC-065-EV-02  
**Verification date:** 2026-08-12 (UTC)  
**Implementer attempt:** 2 of 2 (prior attempt 1 on branch `agent/voc-065-voc-065-t02`
documented deploy-blocked verification; this attempt records post-recovery staging
passes)  
**Reviewed revision:** `develop` tip `bebff72b` (includes VOC-065-T01 via merge
`8962c86`)

## Summary

**VOC-065-AC-03 is satisfied.** After VOC-065-T01 merged to `develop`, real
`deploy-staging.yml` runs deployed the T01-fixed API and executed
`tests/staging-e2e/core-loop.staging.spec.ts` through step 7 successfully.

The pre-fix failure shape from issue #482 / run
[31429774964](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31429774964)
(`observed=[0, 0, 0, 0]` while `reviewedCards=2`) does **not** recur on the
post-T01 staging stack: the core-loop journey passes and step 7's invariant holds.

## T01 dependency (satisfied)

| Item | Value |
|------|-------|
| T01 merge PR | [#523](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/523) |
| T01 merge commit | `8962c869` (2026-08-11 ~21:22 UTC) |
| Fix | `newProductionReviewsRepository` wires `WithGamificationService` / `WithMissionsService` in `apps/api/app/api/production.go` |
| T01 evidence | `t01-evidence.md` |

`bebff72b` is a descendant of `8962c869`; the successful runs below built and
deployed API images from commits that include the wiring fix.

## Primary verification run

| Item | Value |
|------|-------|
| Workflow run | [31579570063](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31579570063) |
| Job | [deploy to staging](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31579570063/job/94059331956) |
| Head SHA | `bebff72b` (`develop` tip at run time) |
| Trigger | push — merge PR #533 |
| Time | 2026-08-12 ~08:42 UTC |
| Overall conclusion | **success** (3m 16s) |

### Post-deploy gate steps (public jobs API)

| Step | Conclusion |
|------|------------|
| Deploy to staging host | success |
| Poll api-staging /healthz and staging web | success |
| Mint synthetic smoke-test session | success |
| **Run the staging core-loop journey** | **success** |
| **DIAGNOSTIC — dump synthetic account's daily_mission_snapshots rows** | **success** |

Playwright GitHub check annotation on this commit:

```
🎭 Playwright Run Summary
  1 passed (9.8s)
```

## Confirmatory run (first full post-T01 staging deploy)

| Item | Value |
|------|-------|
| Workflow run | [31575459316](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31575459316) |
| Head SHA | `372ab8b5` (descendant of `8962c869`) |
| Time | 2026-08-12 ~07:47 UTC |
| Overall conclusion | **success** |
| Core-loop step | success |
| Diagnostic dump step | success |
| Playwright annotation | `1 passed (11.3s)` |

Two independent successful deploys on the post-T01 `develop` line both pass the
full staging core-loop gate.

## Core-loop step 7 — passed with real increment invariant

### Observed values

The staging spec logs explicit counter values only when step 7 **fails** (see
`readReviewedTodayCountAfterReviews` throw text in
`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`) or when VOC-063's retry
loop needs more than one attempt (`step-7-retry` annotation). On a first-attempt
pass, Playwright's GitHub reporter surfaces only the suite summary (`1 passed`).

Authenticated workflow job logs for runs `31579570063` / `31575459316` contain
the per-step Playwright `list` reporter output (step names `2. read the
daily-mission baseline`, `5. work the review queue`, `7. progress reflects the
completed reviews`). This implementer environment had no `GITHUB_TOKEN` /
`GH_TOKEN` with Actions log-read scope, so exact `reviewedBefore`,
`reviewedCards`, and `reviewedAfter` integers were not copied here.

**What is directly evidenced without those integers:**

| Requirement | Evidence |
|-------------|----------|
| Step 7 executes | Core-loop step succeeded; the single test includes step `7. progress reflects the completed reviews` |
| Step 7 invariant (`reviewedAfter >= reviewedBefore + reviewedCards`) | Suite passed — failure throws with the issue #482 diagnostic shape (`observed=[…]`) |
| VOC-063 retry bounds unchanged | No `step-7-retry` annotation on either successful run (first-attempt pass, not a weakened bound) |
| Not the pre-fix never-increments defect | Pre-fix run `31429774964` failed step 7 with `observed=[0,0,0,0]` at `reviewedCards=2`; post-fix runs do not fail step 7 |
| `reviewedCards >= 1` (AC-03 preference) | Not exported to public check annotations on success. The journey saves a word (step 4) before the review loop (step 5); an all-zero review loop would typically emit a `skipped-step` annotation on step 6 (`sentence feedback not reachable… reviewed N card(s)`). Neither successful run exported that annotation. **Independent reviewer:** confirm `reviewedCards >= 1` from the job log step-5 return if strict numeric evidence is required |

### Contrast with issue #482 failure (pre-T01 stack)

| Signal | Run 31429774964 (pre-fix) | Runs 31575459316 / 31579570063 (post-T01) |
|--------|---------------------------|-------------------------------------------|
| Core-loop step | **failure** at step 7 | **success** |
| Step 7 error shape | `observed=[0, 0, 0, 0]` at `reviewedCards=2` | no step-7 failure |
| Diagnostic `reviews_completed` (today) | `0`, `updated_at` hours before test | dump step **success** (row contents in job log) |

## Diagnostic dump

The `DIAGNOSTIC - dump synthetic account's daily_mission_snapshots rows` step
concluded **success** on both verification runs immediately after the core-loop
journey. That step prints today's `reviews_completed`, `local_date`, `timezone`,
`status`, and `updated_at` for `smoke-test-bot@synthetic.vocanova.invalid` via
`docker compose exec postgres psql` (see `deploy-staging.yml`).

Dump **stdout is only in the authenticated job log**, not in the public Actions
API. **Independent reviewer:** confirm from the run log that today's row shows
`reviews_completed > 0` and `updated_at` advanced to the deploy/test window (the
pre-fix dump in issue #482 showed `reviews_completed=0` with `updated_at` hours
stale).

## Attempt 1 follow-up resolved

Attempt 1 (`t02-evidence.md` on commit `b016ed9`) correctly recorded that
post-T01 deploys on 2026-08-11 ~22:03 UTC failed at **Deploy to staging host**
before core-loop ran. Staging deploy recovered on 2026-08-12:

| Run | Deploy step | Core-loop |
|-----|-------------|-----------|
| [31549432988](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31549432988) | failure | skipped |
| [31575459316](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31575459316) | success | **pass** |
| [31579570063](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31579570063) | success | **pass** |

## VOC-063 step-7 hardening (unchanged)

No edits to `apps/web/tests/staging-e2e/core-loop.staging.spec.ts` in VOC-065-T01
or this task (`git diff ec15fca..bebff72b --` that path is empty). Constants
remain `STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS = 4` and
`STEP_7_REVIEWED_COUNT_RETRY_DELAY_MS = 1500`.

## Acceptance criterion mapping

| Criterion | Result |
|-----------|--------|
| **VOC-065-AC-03** | **Satisfied** — real post-T01 `deploy-staging.yml` runs pass core-loop step 7; pre-fix never-increments failure does not recur |
| **VOC-065-AC-04** | **Satisfied** — VOC-063 retry/invariant untouched; VOC-053/issue #450 not re-scoped |
| **VOC-065-TEST-04** | **Pass** — runs `31575459316`, `31579570063` |

## Commands / API queries used

```bash
# Public GitHub Actions API (no auth) — workflow runs after T01 merge
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-staging.yml/runs?per_page=25"

# Job step timelines
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31579570063/jobs"
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31575459316/jobs"

# Playwright check annotations
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/check-runs/94059331956/annotations"

# Confirm T01 fix present on deployed SHA
git show bebff72:apps/api/app/api/production.go  # newProductionReviewsRepository with P4 options

# Staging reachability
curl -sS https://api-staging.vocanova.site/healthz  # HTTP 200 at verification time
```

Job log download returned 403 without admin/token scope — numeric step-5/7
values and diagnostic SQL output are cited as reviewer-accessible from the run
URLs above.

## Out of scope (unchanged)

- No product code changes in this task (verification-only).
- Issue #450 remains open; VOC-053 not re-scoped.
- No historical `daily_mission_snapshots` backfill (`VOC-065-DEP-01`).
