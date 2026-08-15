---
evidence_id: VOC-082-EV-01
task_id: VOC-082-T01
acceptance_criteria:
  - VOC-082-AC-03
  - VOC-082-AC-04
tests:
  - VOC-082-TEST-04
  - VOC-082-TEST-05
date: 2026-08-15
related_change: VOC-082
cites: VOC-082-EV-00
gate_status: streak-defect-resolved-e2e-cap-assertion-fixed-awaiting-rerun
---

# VOC-082-T01 — Staging core-loop through mission completion

**Task:** VOC-082-T01  
**Package:** VOC-082  
**Evidence ID:** VOC-082-EV-01  
**T00 dependency:** merged to `develop` at `3964f08003ceea9c4be314724ad305ab43851714` (PR #681)

## Summary

| Item | Result |
|------|--------|
| Authoritative staging run on T00 tip | **FAIL** (run #258) — step 7 only |
| Streak / completing-review HTTP 500 defect | **Resolved** on run #258 (see below) |
| Mission progress at target | **20 of 20** observed (not stuck at 19/open) |
| Narrow E2E step-7 cap assertion | **Applied** in this revision |
| Full AC-03 green deploy | **Pending** — requires `deploy-staging` re-run after this E2E fix merges |

## VOC-082-TEST-04 — Real staging core-loop through daily target

### Run on T00 merge (`3964f08`) — pre step-7 E2E fix

| Field | Value |
|-------|-------|
| Workflow | `deploy-staging.yml` |
| Run number | **#258** |
| Run ID | `31888014989` |
| URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31888014989 |
| Head SHA | `3964f08003ceea9c4be314724ad305ab43851714` |
| Conclusion | **failure** |
| Failed step | `Run the staging core-loop journey` (step 7) |

GitHub Actions annotation (public run summary):

```
reviewedBefore=19, reviewedCards=2, minimumExpected=21, observed=[20, 20, 20, 20]
Error: Step 7 reviewed-today counter did not reach the expected minimum after 4 attempt(s)
```

**Diagnosis:**

1. **Step 5 succeeded with `reviewedCards=2`.** Both review submissions in the
   queue loop completed without HTTP 500. The account started at **19 of 20**
   (`reviewedBefore=19`); the first submission was the completing review that
   reaches today's target. Under the pre-T00 defect (issue #675 / run #253), that
   submission returned HTTP 500 and left the counter at 19 with snapshot `open`.
   Run #258 advanced the counter to **20** and allowed a second review — the
   streak reconciliation path no longer rolls back the transaction.

2. **Step 7 failed on an E2E assertion gap, not the API defect.** The step-7
   helper required `reviewedAfter >= reviewedBefore + reviewedCards` (21) but
   `reviews_completed` caps at `review_target` (20). After mission completion,
   additional reviews do not raise the home counter. This is expected product
   behavior, not a regression of VOC-082-T00.

3. **Contrast with pre-fix baseline (issue #675).** Run **#253**
   (`e11fb4a6`, https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31886780600)
   failed the core-loop journey on the completing submission path (HTTP 500 from
   `ErrInvalidStreakSnapshot`). Run #258's failure mode is different and occurs
   only after the counter shows target reached.

### Narrow E2E remediation (this revision)

`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`:

- Parse both reviewed count and `review_target` from the home counter.
- Step 7 expects `min(reviewedBefore + reviewedCards, reviewTarget)` so a
  journey that completes today's mission mid-run does not false-fail when extra
  reviews are capped at the target.

This edit is scoped per `change.yaml` ("edit only if a narrow readiness assertion
is required after the API fix") and does not substitute for the T00 streak fix.

### Re-run requirement

A **green** `deploy-staging` run on a revision that includes this step-7 fix is
still required to close **VOC-082-AC-03** with a PASS conclusion. The implementer
shell has no `GH_TOKEN` / `GITHUB_TOKEN` (`gh` exits 4); the calling workflow
should trigger or record the post-merge re-run.

## Mission completion evidence (run #258)

| Signal | Value | Interpretation |
|--------|-------|----------------|
| `reviewedBefore` | 19 | One review short of target at journey start |
| `reviewedCards` | 2 | Completing review + one post-target review both submitted |
| `observed` counter | 20, 20, 20, 20 | Target reached; not stuck at 19/open |
| Step 5 HTTP 500 | None | Completing submission path succeeded |

Equivalent Home progress: **20 of 20 words reviewed today** (from annotation).

## VOC-082-TEST-05 — Diff excludes VOC-081 monitor paths

| Boundary | Status |
|----------|--------|
| `infra/docker-compose.monitoring.yml` | **Not touched** |
| Shared-edge monitor vhosts / Cloudflare | **Not touched** |
| VOC-081 deploy topology | **Not touched** |
| T01 diff scope | `core-loop.staging.spec.ts`, `t01-evidence.md` |

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-082-AC-03 | **Pending re-run** | Run #258 proves streak fix + target reached; overall conclusion FAIL on pre-fix step 7; green run pending after E2E cap fix merges |
| VOC-082-AC-04 | **PASS** | No VOC-081 paths in T01 diff |

## Commands inspected

```bash
bash scripts/governance/validate-governance.sh
# Repository foundation validation passed.
# Governance structure validation passed.

bash scripts/governance/classify-change-risk.sh
# Detected path-based risk floor: R1 (core-loop.staging.spec.ts)

git diff --check
# (no whitespace errors)

curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31888014989"
# status=completed, conclusion=failure, head_sha=3964f08...

# pnpm typecheck:e2e — not run (pnpm unavailable in implementer shell)
```

## Files changed (T01)

- `apps/web/tests/staging-e2e/core-loop.staging.spec.ts` — step-7 cap-aware minimum
- `specs/changes/VOC-082-fix-500-on-the-review-that-completes-today-s/t01-evidence.md` — this file
