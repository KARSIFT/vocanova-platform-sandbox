# VOC-090 — Fix scheduled-synthetics cancellation from staging core-journey timeout

| Field | Value |
|-------|-------|
| Package | `VOC-090` |
| Title | Fix scheduled-synthetics cancellation from staging core-journey timeout |
| Path | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed operational-failure observer (VOC-088-T02) opened
[issue #759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759)
after [scheduled-synthetics run #22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931)
ended with conclusion `cancelled`.

Public run metadata (2026-08-19, triggered on schedule at 15:35 UTC) shows:

- Total duration **30m 19s** — the workflow hit its wall clock.
- Job `synthetic.staging.authenticated-core-journey` failed with:
  **"The job has exceeded the maximum execution time of 30m0s"** and **"The
  operation was canceled."**
- Other synthetic jobs in the same run completed; only the staging
  authenticated core-loop journey blocked a green hourly suite.

The job currently performs a full cold `pnpm install --frozen-lockfile` and
`playwright install --with-deps chromium` on every hourly schedule tick with
no dependency caching, then runs the real-backend Playwright journey (SSH
seed, session mint, up to eight review cards). That setup overhead plus the
journey exceeds the job's `timeout-minutes: 30` budget. By contrast,
`deploy-staging.yml`'s post-deploy core-loop gate allows **40 minutes** for
the same dependency setup and journey pattern.

## Required outcome (summary)

1. Keep the staging authenticated core-loop synthetic functionally equivalent —
   no weakened assertions, no reduced review coverage, no mutating changes to
   production synthetics.
2. Make the hourly scheduled job complete reliably within its declared budget
   by reducing cold-start overhead (pnpm store and Playwright browser caching)
   and aligning the workflow job timeout with the registry synthetic budget
   plus realistic setup time.
3. Update `infra/monitoring/synthetics.yaml` timeout metadata for
   `synthetic.staging.authenticated-core-journey` when the corrected budget
   is known.
4. Add deterministic tests that lock the caching wiring, timeout relationship,
   and registry/workflow consistency.
5. Record live verification that a full `scheduled-synthetics` run completes
   green after merge.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Fix staging core-journey CI budget in scheduled-synthetics | — |
| T01 | Record live verification that the hourly suite completes green | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
