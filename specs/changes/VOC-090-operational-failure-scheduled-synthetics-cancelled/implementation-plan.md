# VOC-090 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/scheduled-synthetics.yml`,
  `infra/monitoring/synthetics.yaml`, `apps/web/tests/staging-e2e/`,
  `apps/web/playwright.staging.config.ts`.
- Prerequisites: VOC-086-T03 merged (scheduled synthetics exist); VOC-088-T02
  merged (operational-failure observer opened issue #759).

## File reconciliation and implementation sequence

### T00 — Fix staging core-journey CI budget in scheduled-synthetics

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/scheduled-synthetics.yml` | modify | Add pnpm/Playwright caching; raise/adjust `timeout-minutes` on staging core-journey job |
| `infra/monitoring/synthetics.yaml` | modify | Align `timeout_seconds` for staging core journey |
| `infra/monitoring/scheduled-synthetics.mjs` | modify (optional) | Job/registry timeout consistency validator |
| `scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs` | create | Deterministic budget/caching tests |
| `scripts/foundation/voc086-scheduled-synthetics.test.mjs` | inspect | Must remain green |
| `specs/changes/VOC-090-.../t00-evidence.md` | create | Root-cause phase breakdown from run logs |

Ordered steps:

1. Inspect run 32271016931 job log timing; record dominant phase in evidence.
2. Add cache restore/save for pnpm store and Playwright browsers in the staging
   core-journey job.
3. Adjust `timeout-minutes` and registry `timeout_seconds` consistently.
4. Add/extend deterministic validators and tests.
5. Run `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` and
   VOC-090 tests.

### T01 — Record live verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-090-.../t01-evidence.md` | create | Green workflow_dispatch run URL, durations |

Ordered steps:

1. Dispatch `scheduled-synthetics.yml` from `develop` after T00 merge.
2. Record success evidence; confirm no duplicate cancelled fingerprint issue.
3. Optionally note first green schedule-triggered hourly run when available.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs
bash scripts/governance/validate-governance.sh   # if invoked by CI on changed paths
git diff --check
```

Live (T01):

- Operator `workflow_dispatch` of `.github/workflows/scheduled-synthetics.yml`
- Inspect run conclusion and job durations on GitHub Actions UI

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` updates workflow behavior on
  the next Actions run; no separate production deploy step for this package.
- **Rollout:** Immediate effect on next scheduled or dispatched synthetics run.
- **Rollback trigger:** Cached dependencies cause false greens; job still times out;
  staging core-journey regresses.
- **Rollback mechanism:** Revert T00 commits; restore prior timeout and remove
  caching steps.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge.
