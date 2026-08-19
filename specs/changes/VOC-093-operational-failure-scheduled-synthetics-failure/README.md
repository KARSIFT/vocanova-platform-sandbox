# VOC-093 — Fix scheduled-synthetics production route-sweep failure

| Field | Value |
|-------|-------|
| Package | `VOC-093` |
| Title | Fix scheduled-synthetics production route-sweep failure |
| Path | `specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#771](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/771) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed operational-failure observer (VOC-088-T02) opened
[issue #771](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/771)
after [scheduled-synthetics run #27](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32288703894)
ended with conclusion `failure`.

Public run metadata (2026-08-19, triggered on schedule at 18:40 UTC on `main` at
`7851d570f1dc245a3b310b9705b54b10585c9ea0`) shows:

- Total duration **9m 10s** — not a workflow timeout or cancellation.
- Job `synthetic.production.authenticated-route-content-sweep` failed with
  **Process completed with exit code 1** on step
  **Run production authenticated route/content sweep synthetic** (~5 seconds).
- Job `synthetic.production.journey-content` **succeeded in the same run** using
  the same `smoke-test-production.sh` harness with profile `journey-content`.
- Other jobs (`synthetic.staging.oauth-expected-state`,
  `synthetic.production.oauth-expected-state`,
  `synthetic.staging.authenticated-core-journey`) completed successfully.

Because journey-content passed, healthz, kill-switch assertions, session mint,
`/api/v1/me`, and journey-situations API checks were healthy at observation time.
The failure is bounded to the **route-sweep web GET section** (ten fixed routes
plus two API-derived discover routes) or to harness logic specific to that
section.

## Required outcome (summary)

1. Identify the failing route-sweep check from run 32288703894 job logs at
   implementation time (no secrets or personal data in evidence).
2. Fix the root cause — production route regression, harness false negative, or
   documented production-data edge case — without weakening the non-mutating
   GET-only contract.
3. Extend or update deterministic tests so the corrected behavior is locked.
4. Record live verification that `scheduled-synthetics` completes green for the
   production route sweep after merge.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Diagnose and fix production route-sweep failure | — |
| T01 | Record live verification that route sweep completes green | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
