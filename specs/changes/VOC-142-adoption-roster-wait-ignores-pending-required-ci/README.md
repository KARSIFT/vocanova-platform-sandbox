# VOC-142 — Adoption roster wait ignores pending required CI and reconcile cannot reuse the open roster PR

| Field | Value |
|-------|-------|
| Package | `VOC-142` |
| Title | Adoption roster wait ignores pending required CI and reconcile cannot reuse the open roster PR |
| Path | `specs/changes/VOC-142-adoption-roster-wait-ignores-pending-required-ci` |
| Status | `draft` |
| Risk | `R4` (draft proposal; required-check completeness before roster merge, and idempotent roster-PR reuse) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1113](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1113) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Native adoption can declare a generated roster PR's checks complete while the
ruleset-required `ci / ci` check is still running. The subsequent protected
merge correctly fails. The documented idempotent `reconcile` dispatch then
cannot resume because `adopt.yml` unconditionally tries to open a new roster
PR even when the exact open roster PR already exists.

These are causally linked recovery defects: the first creates the
interrupted-adoption state and the second makes the documented recovery path
fail in that state.

| Item | Value |
|------|-------|
| Source issue | [#1109](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1109) |
| Merged plan PR | [#1110](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1110) |
| Reviewed plan head | `178ae81ed4f0224fbb12359c90ddb67c6687ca9a` |
| Plan merge | `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2` |
| Adoption run | `33343125733`, job `99342230038` |
| Generated task | [#1111](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1111) |
| Generated roster PR | [#1112](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1112) |
| Exact roster head | `98dd0936a73b64a6b548da6cf2000a6d000917ac` |
| Roster pipeline run | `33343147453` |
| Required CI job | `99342299218` |
| Wait SUCCESS | `2026-08-30T23:57:08Z` → `2026-08-30T23:58:46Z` |
| Merge failure | `2026-08-30T23:58:46Z` → `2026-08-30T23:58:47Z` |
| `ci / ci` SUCCESS | `2026-08-30T23:59:16Z` (still IN_PROGRESS at merge) |
| Reconcile run | `33343250178`, job `99342577393`, failed at `Open roster PR` |

No roster PR was manually merged, closed, recreated, or bypassed. Issue #1112
remains the single exact VOC-141 carrier.

## Root cause

Current `karsift-ai-infra` `adopt.yml` at issue-creation pin
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` has two gaps:

1. `Wait for roster PR checks` accepts two stable, zero-pending snapshots
   from `authoritative-checks-runner.py`. That runner, without
   `--ordinary-pr-gate`, omits GitHub Actions check-runs whose parent is not
   attestable. `parent_run_is_attestable` is false while the parent is still
   in progress, so IN_PROGRESS `ci / ci` is dropped. A partial green set
   (`governance-policy` and `validate` at the incident) can stabilize and
   complete the wait.
2. When `steps.commit.outputs.changed == 'true'`, `Open roster PR` always
   calls `gh pr create` for the deterministic roster branch. It does not
   first resolve and validate an existing open PR for the same base, head
   ref, and exact head SHA. This contradicts the workflow header's promise
   that reconcile safely reuses existing artifacts.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. Roster adoption must not attempt merge until the complete required check
   set for the exact roster head is registered and the newest logical attempt
   of every required row, including `ci / ci`, is SUCCESS.
2. Registration stability must be fail-closed; a partial green snapshot must
   not count as complete merely because its current count is stable.
3. Reconcile must find and validate an existing open roster PR for the
   deterministic branch/base/exact head and resume waiting/merge/cleanup from
   that carrier.
4. Reconcile must create a PR only when no matching open or already-merged
   carrier exists, and must reject ambiguous/mismatched carriers.
5. Existing task issues must be reused and implementation must dispatch
   exactly once after the checked roster merges.
6. Pin the caller fixture to the new independently reviewed infrastructure
   merge. After the exact reviewed caller merge, ordinary `reconcile` for
   plan PR #1110 must be able to resume #1112. Do not snapshot the
   develop/main gap (`karsift-ai-infra#15`). Do not manually merge #1112.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Wait for the complete required roster-check set and reuse the exact open roster PR on reconcile | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, and required-check wait/reuse updates belong under the R4 path floor, and
because the change mutates adoption merge readiness used before implementation
dispatch. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
