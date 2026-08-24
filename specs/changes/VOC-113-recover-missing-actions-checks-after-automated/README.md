# VOC-113 — Recover missing Actions checks after automated merges and release PR creation

| Field | Value |
|-------|-------|
| Package | `VOC-113` |
| Title | Recover missing Actions checks after automated merges and release PR creation |
| Path | `specs/changes/VOC-113-recover-missing-actions-checks-after-automated` |
| Status | `draft` |
| Risk | `R4` (draft proposal; governance-workflow provenance gate and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#948](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/948) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

App-driven task merge and release handoff can mutate GitHub state successfully
while failing to generate the downstream Actions runs needed to validate that
state.

Sanitized live evidence from issue #948 (2026-08-24):

| Observation | Result |
|-------------|--------|
| Final VOC-112 task PR merged to `develop` | No `push` workflows created for the squash commit |
| Release automation opened promotion PR #947 | No pull-request workflows / required checks created |
| Close/reopen and draft/ready transitions | Did not recover missing checks |
| `reconcile-release` | Safely reused release audit and PR; could not merge because exact-head required checks do not exist |

The active ruleset correctly requires `governance-policy`, `validate`, and
`ci / ci`. Do not weaken it and do not synthesize successful statuses.

## Required outcome (summary)

1. Determine and document the precise trigger/token/event behavior.
2. Preserve GitHub App authentication for automated mutations.
3. Guarantee the exact integration or promotion SHA receives normal required
   validation, independent review where applicable, and merge-gate evidence.
4. Use explicit reusable-workflow or dispatch orchestration where event
   recursion/suppression makes implicit triggers unreliable.
5. Fail closed when evidence is absent, stale, or bound to another SHA.
6. Prevent duplicate promotion PRs, duplicate releases, workflow recursion, and
   check/status fabrication.
7. Preserve the existing ruleset and all governance risk classification.
8. Add deterministic tests for task-merge-to-integration and release-PR recovery.
9. Include a bounded timeout and actionable sanitized diagnostics.
10. Use release PR #947 as the controlled live fixture; complete it only after
    genuine exact-head checks pass.
11. Verify post-promotion workflows and close the remediation only after live
    evidence is complete.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Diagnose, implement durable recovery, deterministic tests, docs | — |
| T01 | Recover and complete promotion PR #947 after genuine exact-head checks | T00 |
| T02 | Verify post-promotion workflows and close remediation | T01 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.

## Risk note

This package **proposes R4** because it changes CI/CD lifecycle orchestration and
must repair the strict provenance gate executed by Repository Governance. The
path-based classifier and independent verifier remain authoritative; this draft
proposal is not a determination.
