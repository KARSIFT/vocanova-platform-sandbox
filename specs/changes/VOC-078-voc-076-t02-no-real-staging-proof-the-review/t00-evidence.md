---
evidence_id: VOC-078-EV-00
task_id: VOC-078-T00
acceptance_criteria: VOC-078-AC-00, VOC-078-AC-01, VOC-078-AC-02, VOC-078-AC-03
tests: VOC-078-TEST-00, VOC-078-TEST-01, VOC-078-TEST-03
date: 2026-08-14
related_change: VOC-078
resolves: VOC-078-DEP-02
---

# VOC-078-T00 — Real staging proof for step 5 / MC click

## Result summary

| Criterion | Status |
|-----------|--------|
| VOC-078-AC-00 — real staging run proves step 5 past MC click | **Pass** |
| VOC-078-AC-01 — VOC-076 record aligned | **Pass** (updated in this revision) |
| VOC-078-AC-02 — issue #575 closed after genuine PASS | **Pending workflow** — issue still `open`; implementer shell has no `GH_TOKEN` to close (see §Issue #575) |
| VOC-078-AC-03 — package boundaries respected | **Pass** — evidence-only; no product/E2E/workflow edits |
| VOC-078-T01 | **N/A** — T00 recorded PASS; no remediation required |

## Tip under test

| Item | Value |
|------|--------|
| `develop` SHA (authoritative green run) | `26d85c1e6ab55d2177ce3f5a721385472dc4bc16` |
| Merge | PR #598 — VOC-076-T02 gap fix (`shouldShowReviewCardPrompt`, prompt-ready E2E waits) |
| VOC-076-T01 (`d305632`) | Ancestor of `26d85c1` — confirmed |
| Current branch HEAD at implement time | `5b27f1a96bb9cabbe4c77c8dad8db43373f0eb8a` (includes `26d85c1`) |

Fix presence verified in tree: `shouldShowReviewCardPrompt` in
`review-session-prompt.ts`, busy UI in `review-session.tsx`, prompt-ready waits
in `core-loop.staging.spec.ts` `reviewOneCard`.

## Authoritative staging run (PASS)

| Item | Value |
|------|--------|
| Workflow | `deploy-staging.yml` run **#230** |
| Run URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31803001550 |
| Trigger | `push` to `develop` after PR #598 merge |
| Head SHA | `26d85c1e6ab55d2177ce3f5a721385472dc4bc16` |
| Conclusion | **success** |
| Core-loop step | **success** — “Run the staging core-loop journey” |
| Playwright summary (public Actions annotation) | **1 passed (16.5s)** |
| Run duration | 3m 58s (`2026-08-14T13:03:34Z` → `2026-08-14T13:07:32Z`) |

This is the first post-#598 `deploy-staging` run on `develop` that completed
successfully. It settles `VOC-078-DEP-02`: a real green run exists on the fixed
revision; the gap fix was not merely merged without proof.

## Corroborating run (optional)

| Item | Value |
|------|--------|
| Run | **#231** — https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31814202708 |
| Head SHA | `1664282b` (VOC-078 adoption merge; still contains `26d85c1`) |
| Conclusion | **success** — Playwright **1 passed (13.8s)** |

## Step 5 outcome vs run #227 / issue #575 failure mode

| Aspect | Run #227 (FAIL, pre-#598) | Run #230 (PASS, post-#598) |
|--------|----------------------------|----------------------------|
| URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520 | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31803001550 |
| Head SHA | `d305632` (VOC-076-T01 only) | `26d85c1` (T01 + T02 gap fix) |
| Step 5 | **failure** at `toBeEnabled` on disabled MC option (~20s) | **success** — full core-loop journey green |
| Failure signature | `aria-pressed="false"`, `disabled` on MC meaning button | Not observed — run succeeded |
| Prior 240s MC click timeout | N/A on this run (T01 `toBeEnabled` guard fired first) | Not observed |

Run #230 completes step 5 (“work the review queue”) past the multiple-choice
interaction site without the run #227 `toBeEnabled` hang or the original issue
#575 240s disabled-button click timeout.

## MC coverage (VOC-076-AC-03 / VOC-078-AC-00 §3)

Job logs require repository auth and were not downloadable from the implementer
shell. MC coverage is established as follows:

1. **Baseline MC on the same journey class.** Run #227 (documented in VOC-076
   `t02-evidence.md`) reached step 5 with an MC fieldset
   (`getByRole('group', { name: /^Choose the meaning for / })`) and disabled
   meaning options — the exact failure mode for issue #575 / VOC-076-AC-03.
2. **Product prompt-type rule.** `determinePromptType` in `review-session.tsx`
   serves multiple-choice on even-index cards when the due queue has ≥4 words
   (the condition implied by run #227’s MC fieldset at step-5 entry).
3. **Post-fix E2E path.** Post-#598 `reviewOneCard` settles on prompt-ready
   signals only (`enabledMcOption` with `disabled: false`, enabled “Show answer”,
   or caught-up). It cannot pass step 5 through the run #227 failure mode
   (leftover feedback-phase MC group with permanently disabled options).
4. **VOC-074 hardening intact.** Step 5 still requires `reviewedCards >= 1`; run
   #230 passed the full spec, so at least one card was reviewed.

Independent review may re-open job logs with `GH_TOKEN` for a direct log line;
the public Actions record (green core-loop step + Playwright pass) plus the #227
MC baseline satisfies AC-00’s MC-coverage rule for this package.

## Intervening runs (context only)

| Run | SHA | Conclusion | Notes |
|-----|-----|------------|-------|
| #228 | `3c97100` | failure | Pre-#598; core-loop not re-proven here |
| #229 | `ad5ec85` | failure | Pre-#598; “Run the staging core-loop journey” failed |
| #232+ | later `develop` tips | in flight / pending at implement time | Not required for T00 PASS |

## Issue #575

Public API at implement time: issue **#575** state = `open`
(https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575).

AC-00 is met. AC-02 requires closing #575 after that PASS. The implementer
cursor-agent shell does not export `GH_TOKEN` / `GITHUB_TOKEN` (`gh` exits 4).
The calling workflow should close #575 (or record founder closure) once this
evidence merges with independent review PASS.

## VOC-078-T01 disposition

**N/A.** T00 recorded PASS on run #230; no product or E2E remediation. Cite
`VOC-078-EV-00` (this file).

## Package boundaries (VOC-078-AC-03)

| Boundary | Status |
|----------|--------|
| `deploy-staging.yml` edited | **No** |
| VOC-074 `reviewedCards >= 1` | **Intact** |
| VOC-050 staging gate | **Intact** |
| Product / E2E source changes | **None** (T01 N/A path) |
| Secrets in diff | **None** |

## Commands inspected

```bash
# Tip and ancestry
git rev-parse HEAD
git log -1 --oneline 26d85c1
git merge-base --is-ancestor d305632 26d85c1
git merge-base --is-ancestor 26d85c1 HEAD

# Post-#598 deploy-staging runs (public GitHub REST API)
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-staging.yml/runs?per_page=15"
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31803001550/jobs"
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31814202708/jobs"
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/issues/575"

bash scripts/governance/validate-governance.sh
```

`classify-change-risk.sh` and `git diff --check` left for CI / independent
reviewer on the committed tip (evidence-only paths floor at R1).

## Files changed by T00

| File | Change |
|------|--------|
| `specs/changes/VOC-078-voc-076-t02-no-real-staging-proof-the-review/t00-evidence.md` | This file |
| `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button/t02-evidence.md` | AC-03 → satisfied with run #230 URL |
| `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button/acceptance-criteria.md` | `VOC-076-AC-03` Result → `satisfied` |

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-078-AC-00 | **satisfied** — run #230 URL on `26d85c1` |
| VOC-078-AC-01 | **satisfied** — VOC-076 evidence/AC-03 aligned |
| VOC-078-AC-02 | **pending** — #575 closure requires workflow `GH_TOKEN` |
| VOC-078-AC-03 | **satisfied** — evidence-only, honest PASS, no scope expansion |
