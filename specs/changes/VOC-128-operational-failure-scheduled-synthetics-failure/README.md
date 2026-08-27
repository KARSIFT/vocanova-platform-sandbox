# VOC-128 — Fix scheduled-synthetics production jobs failing on the develop default-branch schedule

| Field | Value |
|-------|-------|
| Package | `VOC-128` |
| Title | Fix scheduled-synthetics production jobs failing on the develop default-branch schedule |
| Path | `specs/changes/VOC-128-operational-failure-scheduled-synthetics-failure` |
| Status | `draft` |
| Risk | `R3` (draft proposal; `.github/workflows/` and `infra/` path floor) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1037](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1037) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Hourly `scheduled-synthetics` fails on this repository because GitHub cron
runs on the default branch `develop`, while the three production synthetic
jobs declare `environment: production`. Those jobs are rejected before a
runner starts, so the workflow conclusion is `failure` even though both
staging checks succeed.

This is the scheduled-synthetics branch-selection footgun that VOC-108
explicitly left out of its own scope. It is not a repeat of VOC-093
(production route-sweep smoke exit 1 on `main` with real steps).

### Live reproduction (2026-08-27)

| Item | Value |
|------|-------|
| Issue | [#1037](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1037) — observer marker `operational-failure:scheduled-synthetics:failure` |
| Actions run | [`33039500346`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33039500346) — scheduled-synthetics run 200 |
| Event | `schedule` at `2026-08-27T04:28:07Z` |
| Head | `develop` @ `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Workflow conclusion | `failure` after about 54 seconds wall clock |
| `synthetic.staging.oauth-expected-state` | success (~7s, runner assigned) |
| `synthetic.staging.authenticated-core-journey` | success (~51s, runner assigned, Playwright steps ran) |
| `synthetic.production.oauth-expected-state` | failure (~2s, empty steps, `runner_id` 0) |
| `synthetic.production.journey-content` | failure (~2s, empty steps, `runner_id` 0) |
| `synthetic.production.authenticated-route-content-sweep` | failure (~2s, empty steps, `runner_id` 0) |
| Default branch | `develop` (public repository metadata) |

No job logs, secrets, session cookies, OAuth state, or user identifiers are
copied here. Public jobs-API metadata is enough to bound the class: production
jobs never started.

## Required outcome (summary)

1. A `schedule` run on `develop` must not fail because production jobs hit the
   GitHub `production` environment gate. Those jobs must skip, not fail, when
   the workflow ref is not `main`.
2. Hourly production synthetics must still execute on `main`, where
   `environment: production` is valid. Do not widen production environment
   branch policy to `develop`.
3. Staging synthetics continue to run from the default-branch schedule that
   matches staging deploys. The develop-default schedule must not start a
   second staging core-journey on `main`.
4. Preserve VOC-086 stable IDs, VOC-088 observer sanitization, VOC-098
   observer-App least privilege, and existing mint-token / kill-switch
   contracts.
5. Prove the repair with deterministic workflow tests and operator-owned live
   evidence after T00 merges to `develop`.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Gate production jobs to `main`, dispatch production-only checks onto `main` from the develop schedule, with tests and docs | — |
| T01 | Record operator-owned live evidence that the develop schedule completes success and production jobs run on `main` | T00 |

See `tasks.md` for full task definitions. T01 split reason is
`post-merge-evidence-not-in-carrier`.

## What this package deliberately does NOT do

- Change production environment deployment-branch settings to allow `develop`.
- Give the operational-failure observer App `actions: write`.
- Weaken route-sweep coverage, OAuth expected-state assertions, or kill switches.
- Re-open VOC-093 harness work or VOC-090 staging timeout work unless T00
  proves a shared regression (default: out of scope).
- Change application product behavior, schema, secrets, or Kuma inventory IDs.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R3** because `.github/workflows/*` and `infra/` are R3
path floors. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
