# VOC-107 — Isolated implementer publisher cannot verify incremental remediation bundles

| Field                     | Value                                                                                        |
| ------------------------- | -------------------------------------------------------------------------------------------- |
| Package                   | `VOC-107`                                                                                    |
| Title                     | Isolated implementer publisher cannot verify incremental remediation bundles                 |
| Path                      | `specs/changes/VOC-107-isolated-implementer-publisher-cannot-verify`                         |
| Status                    | `draft`                                                                                      |
| Risk                      | `R3` (draft proposal; independently classified per task)                                     |
| Authority model           | A-004 active                                                                                 |
| Requirement source        | GitHub issue [#891](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/891)         |
| Target branch             | `develop`                                                                                    |
| Approval                  | `not-approved`                                                                               |
| Implementation authorized | `false`                                                                                      |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule)                                                   |

## Problem

A supervised remediation run produced a valid committed fix and recovery bundle,
but the clean isolated publisher rejected the bundle before publication. The
publisher fetched only the integration branch while the incremental bundle
declared the existing task-PR head (or a locally rebased derivative) as a
prerequisite. That prerequisite was therefore absent from the clean bare
repository. The implementation work had to be recovered manually, wasting one
model run and several minutes.

Sanitized observed evidence from issue #891: Actions run `32539352323`. No logs,
artifacts, credentials, or user data are attached.

Drafting-time read of `karsift-ai-infra/.github/workflows/implement.yml` confirms
the bundle step creates `base_sha..HEAD`, and on attempt 2+ `base_sha` is set to
`HEAD` after checkout/rebase of the existing agent branch — not necessarily an
object present when the publisher fetches only `integration_branch`.
`git bundle verify` correctly refuses that thin bundle.

## Required outcome (summary)

1. Make the implementer bundle contain the complete task-branch-only lineage
   needed by the clean publisher, anchored to the reviewed integration lineage
   rather than only the last incremental delta.
2. Preserve the isolated clean publisher, exact published SHA check, integration
   ancestry check, workflow-path deny rule, and SHA-valued force-with-lease.
3. Preserve the two-attempt cap and never rerun the model merely because a valid
   committed bundle lacks a publisher prerequisite.
4. Add a deterministic end-to-end Git fixture that creates integration, attempt-1,
   and attempt-2 commits, then proves the attempt-2 artifact verifies/imports in
   a clean bare repository containing only integration. Include a negative
   malformed/stale lineage case.
5. Cover rebase-derived remediation lineage so a locally recreated base cannot
   strand publication.
6. Keep logs and artifacts out of issues and review evidence.
7. Keep this root focused; Node runtime deprecations, Go cache warnings,
   dependency updates, OAuth/application behavior, and deployment changes are
   separate follow-ups.

## Tasks

| Task | Summary                                                                                          | Depends on |
| ---- | ------------------------------------------------------------------------------------------------ | ---------- |
| T00  | Integration-anchored implementer bundle lineage, publisher guards, docs, deterministic Git fixture | —          |

See `tasks.md` for the full task definition.

## What this package deliberately does NOT do

- Weaken or remove the isolated clean publisher, exact published SHA check,
  integration ancestry check, workflow-path deny rule, or SHA-valued
  force-with-lease.
- Expand the two-attempt cap or rerun the model solely for a publisher
  prerequisite miss on an otherwise valid committed bundle.
- Soft-reset attempt-2 commits against the integration tip (that would squash
  prior task commits); soft-reset must keep using the pre-model tip.
- Change `plan.yml` planner-bundle behavior (related pattern; explicit follow-up
  if needed).
- Address Node runtime deprecations, Go cache warnings, dependency updates,
  OAuth/application behavior, or deployment changes.
- Change application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID behavior.
- Attach logs, artifacts, credentials, or user data to issues or review evidence.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
