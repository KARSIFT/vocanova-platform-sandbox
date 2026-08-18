---
evidence_id: VOC-086-EV-03
task_id: VOC-086-T03
acceptance_criteria:
  - VOC-086-AC-06
  - VOC-086-AC-07
tests:
  - VOC-086-TEST-11
  - VOC-086-TEST-12
  - VOC-086-TEST-13
date: 2026-08-18
related_change: VOC-086
accountable_owner: unassigned
gate_status: repository-complete-live-synthetics-pending
live_synthetics_claimed: false
remediation_of: 9208316e2de1dea53e1e8d0c3dfc5b31ff400628
---

# VOC-086-T03 — Canonical scheduled synthetics workflow/registry

## Scope of this evidence

This task delivers the scheduled synthetics registry wiring, workflow,
harness reuse, mint masking, and deterministic TEST-11/12/13 coverage. It
does **not** claim live green synthetics, monitor-host reachability, or
AC-05/AC-10 closure (`VOC-086-T05`). AC-07 live Sentry health is also T05;
this revision records the repository non-regression half only.

## Remediation (post independent FAIL on `9208316e…`)

| Finding | Fix |
| --- | --- |
| High — production smoke-profile jobs omitted `EXPECT_OAUTH_ENABLED` | `production-journey-content` and `production-authenticated-route-content-sweep` now mirror `deploy-production.yml`: `EXPECT_OAUTH_ENABLED` from the secrets-present expression, plus `EXPECT_MAGIC_LINK_ENABLED=false`, `EXPECT_NEW_SIGNUPS_ENABLED=false`, `EXPECT_AI_ENABLED=true`. Validator + TEST-11 fail closed if the secrets-present expression is missing on either job. |
| Medium — missing `t03-evidence.md` | This file. |
| Low — mint CSRF optional vs deploy-staging | `mint-synthetic-session.sh` now fails closed when `csrf_token` is empty (same signal as deploy-staging). Production jobs still only consume `session_cookie`; the mint endpoint contract returns both. |
| Low — `fetch-depth: 0` on every synthetic job | Not changed in this remediation (non-functional clone cost). Follow-up: drop to default depth 1 if operators want cheaper scheduled clones. |

## Decision records for this task

| Decision | Recorded choice |
| --- | --- |
| `VOC-086-DEP-03` packaging | One workflow (`.github/workflows/scheduled-synthetics.yml`) with **five discrete jobs** whose `name:` is the stable synthetic ID. Optional `workflow_dispatch.inputs.synthetic_id` runs one check; empty / schedule runs all five. Not a job matrix and not `workflow_call` sub-workflows. |
| Harness reuse | Staging OAuth → `verify-staging-oauth-start.sh`; production OAuth → `verify-production-oauth-start.sh`; journey/route-sweep → `smoke-test-production.sh` profiles; staging core-loop → existing Playwright `playwright.staging.config.ts`. |
| Mint secrets | `STAGING_SMOKE_TEST_SESSION_MINT_TOKEN` / `PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN`; values masked with `::add-mask::`. |
| Production kill switches | Same `EXPECT_*` env as `deploy-production.yml` on both smoke-profile jobs, because those profiles still run the healthz kill-switch section. |
| Sentry split (`VOC-086-D06`) | `.github/workflows/error-monitoring.yml` remains the Sentry path; scheduled synthetics neither call nor replace it. |

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Scheduled synthetics workflow | `.github/workflows/scheduled-synthetics.yml` |
| Synthetics registry | `infra/monitoring/synthetics.yaml` |
| Dispatcher | `infra/scripts/run-scheduled-synthetic.sh` |
| Shared mint script | `infra/scripts/mint-synthetic-session.sh` |
| Production OAuth verifier | `infra/scripts/verify-production-oauth-start.sh` |
| Wiring helper | `infra/monitoring/scheduled-synthetics.mjs` |
| Deterministic tests | `scripts/foundation/voc086-scheduled-synthetics.test.mjs` |
| This evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t03-evidence.md` |

## Acceptance mapping

| AC | Repository outcome |
| --- | --- |
| AC-06 (wiring) | Five stable IDs in registry + workflow job names; mint reuse/masking; production sweep `mutating: false` + `SMOKE_TEST_PROFILE=route-sweep`; TEST-11/12. Live green proof remains T05. |
| AC-07 (non-regression) | TEST-13 asserts scheduled synthetics do not call/replace `error-monitoring.yml`; that workflow still names Sentry and keeps its hourly schedule. Live healthy posture remains T05. |

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
```

Result on this remediation working tree:

```
✔ VOC-086-TEST-11: scheduled synthetics map to required stable IDs
✔ VOC-086-TEST-12: synthetic checks reuse mint secrets and mask sessions
✔ VOC-086-TEST-13: error-monitoring.yml remains separate
✔ VOC-086-TEST-12: dispatcher error messages do not echo cookie values
✔ VOC-086-TEST-11: production smoke-profile jobs fail closed without EXPECT_OAUTH_ENABLED
ℹ tests 5
ℹ pass 5
ℹ fail 0
```

Secrets: none committed; no mint token, session cookie, CSRF token, or Kuma password appears in this evidence.

## Live scheduled synthetics

Pending operator `workflow_dispatch` of `.github/workflows/scheduled-synthetics.yml` after merge (`VOC-086-T05`). This revision does not claim live green checks.
