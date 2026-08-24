# VOC-114 — Fix VOC-113 recovery runner metadata-read failure before dispatch

| Field | Value |
|-------|-------|
| Package | `VOC-114` |
| Title | Fix VOC-113 recovery runner metadata-read failure before dispatch |
| Path | `specs/changes/VOC-114-voc-113-recovery-runner-fails-before-dispatch` |
| Status | `draft` |
| Risk | `R4` (draft proposal; CI/CD lifecycle orchestration and recovery token contract) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#956](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/956) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The newly merged VOC-113 recovery mechanism fails before dispatch in both
supported live modes because the shared exact-SHA metadata read phase cannot
complete with the minted KARSIFT App installation token.

Sanitized live evidence from issue #956 (2026-08-24):

| Observation | Result |
|-------------|--------|
| PR #954 merged to `develop` at `b97e9575…` | Merge, task-completion publication, and task-issue closure succeeded |
| Pipeline run 32696249484 | Step "Recover missing integration push workflows for merged SHA" failed immediately with `actions-check-recovery: github_metadata_read_failed` |
| `reconcile-release` run 32696549963 for release issue #946 | Step "Recover missing exact-head promotion checks" failed immediately with the same error |
| Recovery dispatch | None in either mode |
| Downstream effect | Blocks VOC-113-T01; promotion PR #947 still lacks genuine exact-head required checks |

App-backed merge/PR/issue operations succeed; both modes initially failed at the shared
metadata read that calls commit check-runs, commit status, and Actions runs
endpoints. The leading hypothesis is missing effective **Checks read** (and
related Actions read) on the App token used for recovery.

Post-carrier release run `32724415871` resolved that hypothesis: the installed
App has mutation/workflow-file permissions but no Actions permission. Recovery
therefore uses a dedicated job `GITHUB_TOKEN`; the App remains mutation-only.

## Required outcome (summary)

1. Ensure merge-gate, release, and `recover-actions-checks` recovery jobs carry
   explicit Actions write plus Checks/Statuses/Contents/Pull requests read on
   their job token, while App tokens preserve only narrow PR/issue/content
   mutation permissions.
2. Localize runner failures to sanitized endpoint classes (`check_runs_read_failed`,
   `workflow_runs_read_failed`, `commit_metadata_read_failed`) instead of one
   generic `github_metadata_read_failed`.
3. Add deterministic policy/runner tests proving both recovery modes can read
   metadata under the declared token contract, fail closed when read capability
   is absent, and never dispatch after a metadata-read failure.
4. Re-run live integration recovery and `reconcile-release`; require genuine
   exact-head contexts before promotion PR #947 can complete (unblocking
   VOC-113-T01), then bind the result to the exact T01 carrier head through the
   existing read-only `verify-promotion-check-recovery / verify` action.

Do not weaken ruleset contexts, fabricate checks/statuses, manually merge #947,
or copy logs/secrets into evidence.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Restore recovery metadata-read token contract, localize errors, add tests | — |
| T01 | Live recovery proof for both modes and unblock promotion PR #947 | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.

## Risk note

This package **proposes R4** because it changes CI/CD lifecycle orchestration
and the repository recovery-token contract for exact-SHA gate metadata. The
path-based classifier and independent verifier remain authoritative; this draft
proposal is not a determination.
