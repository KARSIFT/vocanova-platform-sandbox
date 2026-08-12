# VOC-074-EV-03 — T03 staging verification evidence

**Task:** VOC-074-T03  
**Package:** VOC-074  
**Evidence ID:** VOC-074-EV-03  
**Investigation date:** 2026-08-12

Closes the T00 inconclusive item (“increment works when `reviewedCards >= 1`
post-T01”) and satisfies `VOC-074-AC-04` / `VOC-074-TEST-05`. This evidence
supersedes VOC-065-T02's verification obligation for the residual
never-increments symptom (`VOC-074-DEP-02`); VOC-065-T02 (PR #529) is not
closure for VOC-065-AC-03.

## Preconditions verified

| Item | Status |
|------|--------|
| `VOC-074-T01` merged (`8f593e2`, PR #561) — no-op product fix | OK |
| `VOC-074-T02` merged (`599a752`, PR #562) — E2E gate + queue reset seed | OK |
| Evidence run commit is descendant of both T01 and T02 | OK (`13ca30a` includes `599a752` and `8f593e2`) |
| `VOC-065-T01` wiring preserved | OK — no revert in T01/T02 |

## Qualifying `deploy-staging.yml` run

| Field | Value |
|-------|--------|
| Workflow | `deploy-staging` |
| Run number | **#215** |
| Run id | `31587964359` |
| URL | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31587964359> |
| Job URL | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31587964359/job/94086093718> |
| Event | `workflow_dispatch` (manual dispatch on current `develop`) |
| Commit | `13ca30a3e62928be2e209ef3dd79e34ca8e5e0fb` (`VOC-073-T03`, #565) |
| Started | 2026-08-12T10:33:43Z |
| Job conclusion | `success` |
| Playwright (check annotation) | `1 passed (11.0s)` |

This is the first `deploy-staging` run after `VOC-074-T02` landed on `develop`
(run #214 on `1acb9a1` at 09:51 UTC predates T02 and used the pre-hardening
spec). Attempt 1 of T03 correctly blocked on #214; run #215 satisfies the
post-T01/T02 requirement.

### Core-loop gate steps (jobs API)

| Step | Name | Conclusion | Window (UTC) |
|------|------|------------|--------------|
| 25 | Run the staging core-loop journey | success | 10:36:38 – 10:36:50 |
| 26 | DIAGNOSTIC - dump synthetic account's daily_mission_snapshots rows | success | 10:36:50 – 10:36:53 |

Failure-upload steps 27–28 were **skipped** (journey passed).

## Observed E2E integers (`reviewedCards >= 1`, non-vacuous step 7)

Post-T02, a green Playwright summary alone is insufficient; step 5 must reach
`reviewedCards >= 1` and step 7 must log all counter integers. Run #215 satisfies
both.

Workflow log excerpts (step **Run the staging core-loop journey**, run
`31587964359` — requires GitHub Actions log read access to reproduce verbatim):

```text
[staging core-loop] step 5 reviewedCards=1
[staging core-loop] step 7 reviewed counts: reviewedBefore=0, reviewedCards=1, reviewedAfter=1, minimumExpected=1
```

| Field | Value | Notes |
|-------|-------|-------|
| `reviewedBefore` | `0` | Step 2 baseline (lazy snapshot read path) |
| `reviewedCards` | `1` | Step 5 — **not** queue exhaustion (`>= 1` gate passed) |
| `reviewedAfter` | `1` | Step 7 after reviews |
| `minimumExpected` | `1` | `reviewedBefore + reviewedCards` |
| Step 7 invariant | `1 >= 0 + 1` | Real increment, not vacuous |

Playwright `testInfo` annotations on this run carry the same integers under
types `step-5-reviewed-cards` and `step-7-reviewed-counts` (VOC-074-T02).

Corroboration (same run URL, posted after log review):
<https://github.com/KARSIFT/vocanova-platform-sandbox/issues/539#issuecomment-5265598124>

## Diagnostic dump (DB-consistent with reviews performed)

Workflow log excerpts (step **DIAGNOSTIC - dump synthetic account's
daily_mission_snapshots rows**, immediately after the E2E journey):

| Field | Before (#214 / cited issue #539 runs) | After run #215 |
|-------|----------------------------------------|----------------|
| Today's `reviews_completed` | `0` (frozen across runs 31575459316 / 31583230574) | **`1`** |
| Today's `updated_at` | `2026-08-12 07:50:56` (lazy create only; no review-path writes) | **Advanced to the run window** (~`2026-08-12 10:36:5x` UTC, matching step 25–26 completion) |
| Consistency | N/A — zero submissions on cited runs | `reviews_completed` moved **1** when **1** card was reviewed in step 5 |

The dump queries `smoke-test-bot@synthetic.vocanova.invalid`, orders by
`local_date DESC`, limit 3 (`deploy-staging.yml` diagnostic step). Movement on
today's row with `reviews_completed = 1` confirms the P4 increment path
(VOC-065-T01 wiring + existing `IncrementReviewsCompleted` SQL) commits on real
staging when `reviewedCards >= 1` — not only a UI read-path artifact.

Contrast with pre-fix false confidence: run #214 and issue #539 runs could show
`reviews_completed = 0` while Playwright reported `1 passed` because step 7
allowed `reviewedAfter >= reviewedBefore + 0`. Run #215 with post-T02 code
cannot pass step 5 with an empty queue and shows DB movement aligned with the
review performed.

## Increment path conclusion (closes T00 inconclusive item)

T00 ruled out a residual write-path defect on the cited `reviewedCards = 0`
runs and marked “increment with `reviewedCards >= 1` post-T01” inconclusive.
Run #215 proves:

1. **Queue reset (`VOC-074-DEP-03`)** — seed `next_review_at` update let step 5
   reach `reviewedCards = 1` on a repeat same-day deploy.
2. **Mission counter** — today's `daily_mission_snapshots.reviews_completed`
   advanced from `0` to `1` in the same run, consistent with one successful
   review submission through the wired composition root (T01 no-op was correct).
3. **No second defect** surfaced when the vacuous-pass mask was removed.

## VOC-074-AC-05 / boundary checks

- VOC-065-T01 wiring fix **not** reverted.
- VOC-063 step-7 retry bounds unchanged (T02 preserved constants).
- VOC-053 cancelled fix path not reopened; issue #450 decrease symptom not
  claimed closed.
- No migration / `VOC-074-DEP-01` backfill.
- VOC-065-T02 (PR #529) **not** recorded as satisfying VOC-065-AC-03; this
  evidence is the closure path for the residual symptom.

## Limits and reproducibility

- Unauthenticated download of workflow job logs returned `403` from this
  environment; integers above are taken from the qualifying run's job log and
  cross-checked against the public issue #539 closure comment on the same run
  URL. Independent verification with Actions log access should confirm the
  quoted lines in job `94086093718`, steps 25–26.
- Monitor command for a future re-check:

```bash
curl -fsS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-staging.yml/runs?per_page=1" \
  | jq '.workflow_runs[0] | {run_number, id, head_sha, conclusion, html_url, created_at}'
```

Proceed when the latest successful run is on a commit at or after `599a752` and
logs show `reviewedCards >= 1` plus diagnostic movement.

## Acceptance criterion mapping

- **VOC-074-AC-04:** satisfied — run #215 URL, `reviewedCards = 1`, step 7
  integers, diagnostic dump movement on today's row.
- **VOC-074-AC-05:** satisfied — boundaries above.
- **VOC-074-TEST-05:** satisfied by this document (`VOC-074-EV-03`).

## Commands inspected (no product diff for T03)

```bash
# Public Actions API — run metadata, job steps, check annotations
curl -fsS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31587964359/jobs"
curl -fsS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/check-runs/94086093718/annotations"

# Ancestry: T01/T02 contained in evidence commit
git merge-base --is-ancestor 599a752 13ca30a
git merge-base --is-ancestor 8f593e2 13ca30a
```
