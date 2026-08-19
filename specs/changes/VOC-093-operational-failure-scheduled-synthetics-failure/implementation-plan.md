# VOC-093 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `infra/scripts/smoke-test-production.sh`,
  `.github/workflows/scheduled-synthetics.yml`, `infra/monitoring/synthetics.yaml`,
  `apps/web/` (if route regression proven).
- Prerequisites: VOC-085-T02 merged (route sweep exists); VOC-086-T03 merged
  (scheduled synthetics exist); VOC-088-T02 merged (operational-failure observer
  opened issue #771).

## File reconciliation and implementation sequence

### T00 — Diagnose and fix production route-sweep failure

| File | Action | Notes |
|------|--------|-------|
| `infra/scripts/smoke-test-production.sh` | modify (likely) | Fix harness if false negative proven |
| `infra/scripts/smoke-test-production.selftest.sh` | modify (likely) | Regression fixture for failing case |
| `infra/scripts/run-scheduled-synthetic.sh` | inspect | Unlikely change; verify profile wiring |
| `.github/workflows/scheduled-synthetics.yml` | inspect | Change only if env miswire proven |
| `infra/monitoring/synthetics.yaml` | inspect | Update only if budget/metadata must change |
| `apps/web/` | modify (conditional) | Production route fix if regression proven |
| `scripts/foundation/voc085-production-route-sweep.test.mjs` | modify (likely) | Lock fix or fixture |
| `scripts/foundation/voc086-scheduled-synthetics.test.mjs` | inspect | Must remain green |
| `specs/changes/VOC-093-.../t00-evidence.md` | create | Failing check + fix rationale |

Ordered steps:

1. Inspect run 32288703894 route-sweep job log; record failing `FAIL:` line(s).
2. Choose harness vs application fix path per evidence.
3. Implement minimal fix; extend selftests and foundation tests.
4. Run deterministic commands listed in Validation.

### T01 — Record live verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-093-.../t01-evidence.md` | create | Green workflow_dispatch run URL |

Ordered steps:

1. Dispatch `scheduled-synthetics.yml` from `develop` after T00 merge.
2. Record success evidence; confirm no duplicate failure fingerprint issue.
3. Optionally note first green schedule-triggered hourly run when available.

## Validation and independent verification

Deterministic (T00):

```bash
bash infra/scripts/smoke-test-production.selftest.sh
node --test scripts/foundation/voc085-production-route-sweep.test.mjs
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
```

If `apps/web/` changes:

```bash
pnpm validate
```

Live (T01):

- Operator `workflow_dispatch` of `.github/workflows/scheduled-synthetics.yml`
  with `synthetic_id: synthetic.production.authenticated-route-content-sweep`
- Inspect run conclusion and job status on GitHub Actions UI

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` updates harness and any app fix
  on the repository-controlled promotion path to `main`.
- **Rollout:** Production route behavior changes deploy with normal promotion;
  workflow/harness changes affect the next scheduled or dispatched synthetics run.
- **Rollback trigger:** Route sweep still fails after T00; fix causes false green;
  production route regression introduced.
- **Rollback mechanism:** Revert T00 commits.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge.
