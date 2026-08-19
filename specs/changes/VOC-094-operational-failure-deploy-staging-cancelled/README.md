# VOC-094 — Fix deploy-staging cancellation from concurrency queue supersession

| Field | Value |
|-------|-------|
| Package | `VOC-094` |
| Title | Fix deploy-staging cancellation from concurrency queue supersession |
| Path | `specs/changes/VOC-094-operational-failure-deploy-staging-cancelled` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#781](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/781) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed operational-failure observer (VOC-088-T02) opened
[issue #781](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/781)
after [deploy-staging run #299](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32290409156)
ended with conclusion `cancelled`.

Public run metadata (2026-08-19, push to `develop` at 18:58 UTC, head
`411fb60157d437e495fc07f599e268669f139e5a`) shows:

- Total duration **~2m 28s** — not a workflow timeout or an in-progress deploy
  cancellation (`cancel-in-progress: false`).
- Workflow annotation: **"Canceling since a higher priority waiting request for
  staging-deploy exists"**.
- GitHub Actions jobs API reports **zero jobs** for the run — the cancellation
  happened while the run was queued behind an in-flight or already-pending
  deploy in the `staging-deploy` concurrency group.
- Display title references the VOC-093 plan merge to `develop`, consistent with
  rapid successive merges to the integration branch during active package work.

`deploy-staging.yml` intentionally sets `cancel-in-progress: false` so an
in-progress deploy is never torn down mid-flight. GitHub's default concurrency
queue, however, allows only **one running** and **one pending** run per group;
a third push cancels the older pending run instead of queueing it. That
supersession is expected platform behavior but currently surfaces as a governed
operational failure with no underlying deploy defect.

## Required outcome (summary)

1. Allow multiple pending staging deploys to queue sequentially (`queue: max`)
   without cancelling superseded commits when `develop` receives rapid merges.
2. Teach the operational-failure observer to **skip** opening issues for
   concurrency-superseded `cancelled` conclusions on deploy workflows, using
   bounded GitHub API metadata only (no job logs or step output in issues).
3. Apply the same queue posture to `deploy-production.yml` for parity on the
   `production-deploy` group.
4. Lock the queue wiring and benign-cancel classifier with deterministic tests.
5. Record live verification that a superseded pending run no longer leaves a
   spurious open `deploy-staging:cancelled` issue, and that the latest commit
   still deploys successfully.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Fix deploy concurrency queue and benign-cancel observer filter | — |
| T01 | Record live verification that superseded cancels stop opening issues | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
