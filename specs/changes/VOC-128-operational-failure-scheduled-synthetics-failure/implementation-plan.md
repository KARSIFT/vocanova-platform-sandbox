# VOC-128 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/scheduled-synthetics.yml`, GitHub
  Environment `production` secret bindings, synthetic session mint secrets
  (read-only use).
- Prerequisites: VOC-086-T03 merged (scheduled synthetics exist); VOC-088-T02
  merged (observer opened issue #1037); VOC-098 merged (observer App must stay
  without Actions write); VOC-090 merged (staging core-journey budget must not
  be doubled by a naive full-suite dispatch).

## File reconciliation and implementation sequence

### T00 — Gate production jobs to main and dispatch production-only checks

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/scheduled-synthetics.yml` | modify | `main`-only production job `if:`; dispatcher job; production-only dispatch input |
| `infra/monitoring/scheduled-synthetics.mjs` | modify | Validators for ref gate, dispatcher, and production-only filter |
| `scripts/foundation/voc086-scheduled-synthetics.test.mjs` | modify | Lock the new contract; keep VOC-086-TEST-11/12 green |
| `docs/operations/monitoring.md` | modify | Document develop-schedule vs `--ref main` production split |
| `specs/changes/VOC-128-.../t00-evidence.md` | update | Public metadata plus implemented diff and test results |

Ordered steps:

1. Confirm run 33039500346 public jobs metadata still matches `VOC-128-D00`.
   Do not copy logs or secrets.
2. Add a `main`-only ref conjunct to each production job `if:` so off-`main`
   those jobs skip. Keep `environment: production` and the existing
   `synthetic_id` filter.
3. Add `workflow_dispatch` boolean input (preferred name
   `production_jobs_only`, default false) that skips staging jobs when true.
4. Add job `dispatch-production-synthetics-on-main` that runs on `schedule`
   when `github.ref != 'refs/heads/main'`, with job-level
   `permissions.actions: write`, and dispatches this workflow at `main` with
   the production-only filter set. Use `GITHUB_TOKEN`. Do not use an App token.
5. Extend `scheduled-synthetics.mjs` validators and `voc086` tests per
   `VOC-128-D06`.
6. Update `docs/operations/monitoring.md` in the same PR.
7. Run the deterministic commands listed in Validation.

### T01 — Record operator-owned live evidence

| File | Action | Notes |
|------|--------|-------|
| `.karsift/live-evidence/VOC-128-T01.yaml` | update | Retarget `sha_lineage` to `exact_sha` of the T00 merge (VOC-128-D07) |
| `specs/changes/VOC-128-.../t01-evidence.md` | update | Allowlisted develop-success and main production-success metadata |

Ordered steps:

1. After T00 merges to `develop`, set the live-evidence contract `exact_sha`
   to that merge (or the develop SHA actually dispatched that contains T00).
2. Enter operator-owned waiting. Repository-controlled reconcile may dispatch
   `scheduled-synthetics.yml` on `develop` using the contract `dispatch` block.
3. Confirm develop-run conclusion `success` with the three required jobs.
   Confirm the dispatched `main` run's three production jobs succeeded.
4. Record only allowlisted metadata. Confirm issue #1037 remains the sole open
   `scheduled-synthetics:failure` fingerprint owner.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Live (T01):

- Operator-owned / repository-controlled `workflow_dispatch` or next hourly
  `schedule` of `.github/workflows/scheduled-synthetics.yml` on `develop`
- Follow-on `main` production-only run created by the dispatcher
- Inspect conclusions and job names on the GitHub Actions UI only; do not
  paste logs

Independent verifier binds to the exact task PR SHA; confirms scope, tests,
docs, and evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` updates the scheduled
  workflow on the repository-controlled promotion path. This package does not
  itself authorize production application deployment.
- **Rollout:** After T00 merges to `develop`, the next default-branch schedule
  uses the new skip-and-dispatch behavior immediately. Production job YAML on
  `main` updates when this package is promoted.
- **Rollback trigger:** Develop schedule still fails, dispatcher fails closed,
  production jobs skip forever with no `main` run, or staging core-journey
  starts twice per hour.
- **Rollback mechanism:** Revert the T00 workflow/docs/test commits. Last
  known good is the pre-T00 `scheduled-synthetics.yml` on `develop`, which
  currently fails production jobs on every hourly schedule.
- **Accountable owner:** unassigned at draft time; implementation owner after
  adoption.
