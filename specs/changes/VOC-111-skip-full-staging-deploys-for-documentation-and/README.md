# VOC-111 — Skip full staging deploys for documentation and evidence-only pushes

| Field | Value |
|-------|-------|
| Package | `VOC-111` |
| Title | Skip full staging deploys for documentation and evidence-only pushes |
| Path | `specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#920](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/920) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

`deploy-staging.yml` runs the complete image-build, registry-push, SSH deploy, public
probe, OAuth check, and authenticated browser journey on **every** push to `develop`,
even when the merged diff changes only governed plans, task rosters, or evidence.

The workflow header currently describes docs-only deploys as a near-no-op because
Docker layers are cached. Current behavior is materially expensive and can replay
unrelated runtime failures.

Sanitized evidence from issue #920:

| Merge class | Example | deploy-staging run |
|-------------|---------|-------------------|
| Plan-only | `86df6779` | `32568473144` |
| Roster-only | `60822aa5` | `32568622178` |
| Evidence-only | PR #917 | `32572863842` (~3m39s; runtime inputs unchanged) |

Plan/roster pushes also replayed the pre-existing Next.js staging outage before the
VOC-110 runtime fix merged, wasting capacity and creating misleading extra failure
observations.

## Required outcome (summary)

1. Pushes to `develop` that change only documentation, governed change-package
   material, or evidence must **not** start the full staging deployment.
2. Runtime/deployment-relevant changes must continue to trigger fail-closed staging
   deployment (application code, shared packages, infra/deploy assets, root workspace
   manifests/lockfiles, staging e2e tests, and the deploy workflow/selector tests).
3. `workflow_dispatch` remains available for explicit retry/redeploy with existing
   behavior.
4. Staging database, secrets, directories, Docker networks, deploy-user isolation,
   shared-edge nginx architecture, health/OAuth/core-loop gates, and
   operational-failure observer semantics remain unchanged.
5. Add deterministic positive and negative selector tests, including root-file and
   shared-package cases.
6. Update stale workflow documentation that describes docs-only deploys as a
   near-no-op.
7. Live verification: one docs/evidence-only push produces **no** deploy-staging run;
   deterministic tests prove runtime-relevant paths remain selected.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Add fail-closed push selection, selector tests, and doc updates | — |
| T01 | Record live verification that docs/evidence-only pushes skip deploy-staging | T00 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.

## Risk note

This package **proposes R3** because it changes `.github/workflows/deploy-staging.yml`.
The path-based classifier and independent verifier remain authoritative; this draft
proposal is not a determination.
