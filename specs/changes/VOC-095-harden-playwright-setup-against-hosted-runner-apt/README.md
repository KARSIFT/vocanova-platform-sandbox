# VOC-095 — Harden Playwright setup against hosted-runner apt mirror stalls

| Field | Value |
|-------|-------|
| Package | `VOC-095` |
| Title | Harden Playwright setup against hosted-runner apt mirror stalls |
| Path | `specs/changes/VOC-095-harden-playwright-setup-against-hosted-runner-apt` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#792](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/792) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Two consecutive [deploy-staging run attempts](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32299315180)
for merge commit `fdde12daeccf521a5a8be171294eba79a138717b` completed application
convergence but were **cancelled at the workflow timeout** while
`playwright install --with-deps chromium` waited on the hosted Ubuntu package mirror.
The same transient affected an accessibility workflow run; an exact-SHA rerun succeeded.

Drafting-time repo read: four workflows install Chromium inline with `--with-deps`:

| Workflow | Browser cache before install |
|----------|------------------------------|
| `.github/workflows/accessibility.yml` | no |
| `.github/workflows/lighthouse.yml` | no |
| `.github/workflows/deploy-staging.yml` (staging core-loop) | no |
| `.github/workflows/scheduled-synthetics.yml` | yes (`~/.cache/ms-playwright`) |

The `--with-deps` path invokes `apt-get` without a repository-managed timeout or retry
contract, so a slow or stalled mirror can consume the entire job budget and cancel
deploy-staging after successful SSH convergence — a false operational failure unrelated
to application health.

## Required outcome (summary)

1. Introduce a **single repository-managed install contract** with bounded duration,
   limited retries, and post-install verification — fail-closed, never skip browser tests.
2. Apply the contract to **every** workflow that installs Chromium for accessibility,
   Lighthouse, staging core-loop, or scheduled core-loop gates.
3. Restore Playwright browser cache (`~/.cache/ms-playwright`) **before** install in
   all four workflows, matching the scheduled-synthetics key scheme.
4. Preserve existing workflow timeouts, staging/production isolation, deploy ordering,
   OAuth controls, and operational-failure monitoring semantics.
5. Lock the contract with deterministic foundation tests.
6. Record live evidence of a successful staging core-loop after merge before closure.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Add bounded Playwright Chromium install script and foundation tests | — |
| T01 | Wire all four workflows and monitoring validator to the shared contract | T00 |
| T02 | Record live staging core-loop verification after deploy-staging merge | T01 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
