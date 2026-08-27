# VOC-138 — Promotion PR CI fails VOC-112 ancestry when captured commit is not reachable

| Field | Value |
|-------|-------|
| Package | `VOC-138` |
| Title | Promotion PR CI fails VOC-112 ancestry when captured commit is not reachable |
| Path | `specs/changes/VOC-138-promotion-pr-ci-fails-voc-112-ancestry-when` |
| Status | `draft` |
| Risk | `R4` (draft proposal; provenance mode selection and required-check recovery) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1091](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1091) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Fresh promotion PR [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090)
(`main` <- `develop`) at exact head
`87f0efcb94a213a0ede9fdbca94a707a22d42b86` cannot pass required `ci / ci`.

| Item | Value |
|------|-------|
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Release issue | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `87f0efcb94a213a0ede9fdbca94a707a22d42b86` |
| PR run | `33122154521` |
| Failed jobs | `98691441027`, rerun `98692552949` |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| Fail-closed message | `PR ancestry mode requires every captured commit object` |
| Captured subject | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Dispatch recovery (pass) | run `33122158425` (339 foundation tests) |
| reconcile-release (blocked) | `33122099253`, `33122436137` |
| First recovery class | `selected_required_run_mismatch` |
| Issue-creation pin | `b263c0c110591cc798b89277dfc35542abb1597b` (#167) |

No production promotion occurred. PR #1090 remains open and blocked.

## Root cause

Reusable CI checks out `pull/1090/merge` with `fetch-depth: 0` and
`run-app-checks.sh` receives the exact PR base/head pair. Because the VOC-112
capture fixture differs between `main` and `develop`, the script selects
`pr-ancestry`. That mode requires the historical capture subject commit
object. Squash-era subject `f9d11e23…` is not reachable in the synthetic
promotion checkout. A failed-job rerun cannot change that
checkout/provenance decision, so `reconcile-release` repeats a guaranteed
failure. Workflow-dispatch recovery passes only because it takes the non-PR
`squash-safe-push` path; it does not repair the failing required PR check.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. Make same-repository `main` <- `develop` promotion PR validation use the
   existing merge-base/hash-bound `pr-validation` contract when the recorded
   subject commit object cannot be resolved. Keep supplying exact PR base/head
   SHAs. Do not switch the promotion PR to `--squash-safe-push`.
2. Do not fetch or hydrate evidence commits. Do not weaken ordinary
   (non-promotion) PR `pr-ancestry` fail-closed behavior when the capture
   fixture is added, modified, or deleted.
3. Keep negative tests that reject tampered merge-base/current hashes and
   missing or malformed PR SHAs.
4. Make exact-head check recovery select or publish one unambiguous successful
   validation at the promotion head instead of rerunning a structurally doomed
   `pull_request` job.
5. Add a deterministic regression for a `main` <- `develop` promotion whose
   recorded subject commit object is absent, proving `VOC-112-TEST-12` and
   `VOC-112-TEST-13` pass under `pr-validation`.
6. Pin the caller fixture to the new independently reviewed infrastructure
   merge. Keep the eight VOC-112 no-change paths byte-identical to
   `b9e74fc2db4691c48c637639b265d527de9f4505`.
7. After the exact reviewed caller merge, `reconcile-release` for #1089 can
   merge #1090 (or the live promotion at the then-current `develop` head),
   synchronize `develop` to that merge SHA, and allow the normal production
   deployment gate to run.
8. Preserve the named run/job IDs as audit evidence. Do not snapshot the
   develop/main gap (`karsift-ai-infra#15`).

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Unblock promotion PR CI and exact-head recovery when the VOC-112 subject is unreachable | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, and recovery-contract updates belong under the R4 path floor, and
because the change protects application-check provenance and required-check
recovery. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
