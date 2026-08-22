# VOC-110 — Fix deploy-staging failure after dependabot integration merge

| Field | Value |
|-------|-------|
| Package | `VOC-110` |
| Title | Fix deploy-staging failure after dependabot integration merge |
| Path | `specs/changes/VOC-110-operational-failure-deploy-staging-failure` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#911](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/911) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed operational-failure observer (VOC-088-T02) opened
[issue #911](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/911)
after [deploy-staging run #364](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32566405628)
ended with conclusion `failure`.

Public run metadata (2026-08-22, push to `develop` at 09:59 UTC, head
`f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46`) shows:

- Total duration **~8m 28s** — the `deploy to staging` job ran **~8m 23s** and
  exited **1** (not a workflow timeout, not a concurrency-queue cancellation).
- Trigger: merge of Dependabot PR **#859** (`dependabot/npm_and_yarn/develop/minor-and-patch`,
  eight npm minor/patch updates).
- Public annotations include missing Playwright report paths on the failure
  artifact-upload steps (`apps/web/playwright-report-staging/`,
  `apps/web/test-results-staging/`) and a benign Go cache restore warning
  (`go.mod` not found at repo root).
- Two Docker build attestations were produced — build/push steps likely completed
  before the failing step.

This is distinct from:

- **VOC-094** — `cancelled` pending-run supersession with **zero jobs** started.
- **VOC-095** — workflow **timeout** during unbounded Playwright `--with-deps`
  install after deploy convergence.
- **VOC-090** — scheduled-synthetics **cancellation** (timeout), not deploy-staging.

The exact failing workflow step is **not determined at drafting time** (job logs
require authenticated access). T00 must identify it and apply the smallest correct
fix.

## Required outcome (summary)

1. Record run 32566405628 root-cause metadata and the failing step from job logs
   at implementation time (sanitized — no secrets, sessions, or personal data).
2. Fix the identified defect in the smallest correct surface (application,
   staging core-loop test, deploy script, or workflow wiring — bounded by evidence).
3. Extend deterministic tests to lock the fix or a regression fixture matching
   the failure mode.
4. Record operator-owned live evidence that a post-fix `deploy-staging` run on
   `develop` reaches conclusion `success` including the staging core-loop gate.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Diagnose and fix deploy-staging failure from run 32566405628 | — |
| T01 | Record live verification that deploy-staging succeeds on develop | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
