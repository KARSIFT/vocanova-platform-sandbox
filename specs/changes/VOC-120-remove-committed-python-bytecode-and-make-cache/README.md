# VOC-120 — Remove committed Python bytecode and make cache ignores portable

| Field | Value |
|-------|-------|
| Package | `VOC-120` |
| Title | Remove committed Python bytecode and make cache ignores portable |
| Path | `specs/changes/VOC-120-remove-committed-python-bytecode-and-make-cache` |
| Status | `draft` |
| Risk | `R4` (draft proposal; governance hygiene validation plus `infra/*` bytecode untrack) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#987](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/987) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Generated Python bytecode remains committed in the caller repository, and the
shared infrastructure repository has no repository-level Python cache ignore
rules. That leaves a generated binary on protected branches and makes fresh
infrastructure clones susceptible to untracked `__pycache__` / `.pyc` clutter.

Issue #987 records exact evidence at current protected refs:

| Observation | Result |
|-------------|--------|
| Caller `origin/develop` | `d67e34b149ad45fd8dcd0e3860a0bbfd1d3da0fb` |
| Caller `origin/main` | `48d7bb5a410035d4d1f12e5c34860399509d9215` |
| `git ls-files infra/scripts/__pycache__` | `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` |
| Tracked blob size | 5,825 bytes |
| Bytecode kind | CPython 3.12 timestamp-based bytecode |
| Introducing commit | `2118865b6acc4bbd533f62b0cf12330f6cf123d0` (`VOC-067: VOC-067-T05 (attempt 2)`) |
| Caller `.gitignore` | already contains `__pycache__/` and `*.py[cod]`; those rules do not untrack an already committed file |
| Shared infra `origin/main` | `37b06aa95030e235b7311b3c14ee23977f62ac76` has no repository `.gitignore` |
| Infra local workaround | private `.git/info/exclude` entries that are not cloned |

No source file was deleted during discovery. The tracked bytecode was restored
byte-for-byte after the audit exposed it.

## Required outcome (summary)

1. Remove the tracked caller `.pyc` without changing its Python source behavior.
2. Add portable Python cache ignore patterns to `KARSIFT/karsift-ai-infra` at
   repository scope.
3. Keep or verify the caller repository-level ignore rules.
4. Add deterministic validation that protected source trees contain no tracked
   Python bytecode/cache artifacts and that the shared-infrastructure ignore rules
   cover `__pycache__/`, `*.pyc`, and related Python bytecode variants.
5. Validate both repositories, independently review exact revisions, merge
   infrastructure first if the caller fixture/pin consumes the change, pin any
   caller mirror to the exact infrastructure merge SHA if applicable, promote, and
   record rollback/evidence.
6. Do not weaken governance, exact-SHA review, protected checks, risk
   classification, or retry limits.

This is cleanup of generated artifacts only. It does not authorize application
behavior, deployment topology, credential, provider, or model-routing changes.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Untrack caller bytecode, add portable infra ignores, and add hygiene validation | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller validation belongs under
`scripts/governance/` and/or `tooling/governance/` (R4 path floors). Untracking
the `infra/scripts/__pycache__/` blob is R3 by path. The path-based classifier
and independent verifier remain authoritative; this draft proposal is not a
determination.
