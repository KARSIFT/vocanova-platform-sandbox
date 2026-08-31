# VOC-144 — Roster reconciliation resolves carrier before pushed PR head metadata converges

| Field | Value |
|-------|-------|
| Package | `VOC-144` |
| Title | Roster reconciliation resolves carrier before pushed PR head metadata converges |
| Path | `specs/changes/VOC-144-roster-reconciliation-resolves-carrier-before` |
| Status | `draft` |
| Risk | `R4` (draft proposal; exact-SHA roster-carrier reuse after post-push REST lag) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1122](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1122) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Documented `pipeline.yml` reconciliation for merged plan PR #1110 cannot reuse
the existing deterministic roster PR #1112 after VOC-142. The adopt job pushes
a new exact commit to `karsift/roster-voc-141`, then immediately fails in
`Open roster PR` with `MISMATCHED_OPEN_CARRIER`. After the job fails, GitHub
exposes the just-pushed head and the same resolver returns `reuse_open`.
Every retry creates a fresh roster commit/head and fails at the same
boundary.

| Item | Value |
|------|-------|
| Source issue | [#1122](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1122) |
| Merged plan PR | [#1110](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1110) |
| Generated roster PR | [#1112](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1112) |
| Deterministic branch | `karsift/roster-voc-141` |
| Issue-creation pin | VOC-142 merge `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Reconcile run 1 | `33437239322` pushed `958e0fedf742173320bb89cfe690ec7070b49e93`, then `MISMATCHED_OPEN_CARRIER`; afterward PR #1112 exposed that head and local resolution returned `reuse_open` |
| Idempotent retry run 2 | `33437514152` pushed `0206cb70437cb751a19a2e715d8202b672060b50`, then the same failure; afterward PR #1112 exposed that new head |

No roster PR was manually merged, closed, recreated, or bypassed. Issue #1112
remains the single exact VOC-141 carrier.

## Root cause

`karsift-ai-infra` `adopt.yml` at issue-creation pin
`8993e867640dfb604dec0466c4e0787e68d8e258` runs `Push roster branch` and then
invokes `roster-carrier-runner.py` without a bounded metadata-convergence
wait. `resolve_roster_carrier` is correctly fail-closed for a stable
same-ref SHA mismatch, but the GitHub adapter does not distinguish transient
post-push PR-head REST lag from a durable identity mismatch.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. After a roster-branch push, carrier resolution must boundedly poll the
   existing same-repository, same-head-ref, same-base OPEN PR until its
   listed head equals the locally pushed exact SHA.
2. Keep a finite timeout. Fail closed on a stable SHA mismatch after the
   bound, on ambiguous carriers, on repository/base mismatch, or on API
   failure.
3. Do not weaken exact-SHA identity, create a duplicate PR, or manually
   merge #1112.
4. Add a deterministic regression for stale-old-head then exact-new-head
   convergence (injected snapshots / fake clock, not wall-clock GitHub lag).
5. Pin the caller fixture to the new independently reviewed infrastructure
   merge. After the exact reviewed caller merge, ordinary `reconcile` for
   plan PR #1110 must be able to resume #1112. Do not snapshot the
   develop/main gap (`karsift-ai-infra#15`).

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Preserve VOC-142 complete-required-set roster wait.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Boundedly wait for existing roster-PR head metadata to expose the pushed SHA, then reuse that carrier | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, and exact-SHA carrier-reuse updates belong under the R4 path floor, and
because the change mutates whether adopt may treat a post-push REST SHA lag
as a permanent mismatched carrier. The path-based classifier and independent
verifier remain authoritative; this draft proposal is not a determination.
