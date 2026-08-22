# VOC-110 — Restore staging web and gate the shipped container runtime

| Field | Value |
|-------|-------|
| Package | `VOC-110` |
| Title | Restore staging web after Next.js standalone regression and add a container runtime gate |
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

Run metadata and privacy-safe read-only diagnosis (2026-08-22, push to `develop`
at 09:59 UTC, head `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46`) show:

- Total duration **~8m 28s** — the `deploy to staging` job ran **~8m 23s** and
  exited **1** (not a workflow timeout, not a concurrency-queue cancellation).
- Trigger: merge of Dependabot PR **#859** (`dependabot/npm_and_yarn/develop/minor-and-patch`,
  eight npm minor/patch updates).
- The API health step passed, then `Poll staging.vocanova.site/` exhausted its
  five-minute budget while the public web returned HTTP 502.
- `vocanova-web` was restarting because the Next.js 16.3.1 standalone output
  omitted an `@swc/helpers` ESM module on Node 24. Production remained on its
  isolated prior image and healthy.
- The failure matches upstream vercel/next.js issue #97358; stable Next.js 16.3.2
  contains the 16.3 backport from #97372/#97453.

This is distinct from:

- **VOC-094** — `cancelled` pending-run supersession with **zero jobs** started.
- **VOC-095** — workflow **timeout** during unbounded Playwright `--with-deps`
  install after deploy convergence.
- **VOC-090** — scheduled-synthetics **cancellation** (timeout), not deploy-staging.

The merged dependency PR passed source build, lint, type, browser, and governance
checks because none booted the production Docker artifact. That missing runtime
boundary is part of the root cause, not an optional follow-up.

## Required outcome (summary)

1. Upgrade `next` and `@next/eslint-plugin-next` together from 16.3.1 to stable
   16.3.2; use 16.3.0 only as rollback if 16.3.2 cannot boot successfully.
2. Add a merge-gating CI job that builds the real `apps/web/Dockerfile`, starts
   the image, confirms the container remains running, and requires HTTP 2xx.
3. Add deterministic workflow tests that prove relevant web/dependency changes
   execute this runtime check and that merge-gate waits for it.
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
