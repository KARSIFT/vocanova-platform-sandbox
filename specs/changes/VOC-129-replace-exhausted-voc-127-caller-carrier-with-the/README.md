# VOC-129 — Replace exhausted VOC-127 caller carrier with the exact infrastructure #164 fixture and pin

| Field | Value |
|-------|-------|
| Package | `VOC-129` |
| Title | Replace exhausted VOC-127 caller carrier with the exact infrastructure #164 fixture and pin |
| Path | `specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin, live pipeline dispatch, and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1042](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1042) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The already-approved VOC-127 release-synchronization outcome still has no
publishable caller carrier. VOC-127-T00 (#1039 / PR #1041) exhausted attempt
`2/2`. The attempt-2 artifact retained the obsolete infrastructure #163 pin,
the clean publisher refused workflow-file publication, and final-attempt
policy forbids another VOC-127 implementer retry.

Infrastructure follow-up [KARSIFT/karsift-ai-infra#164](https://github.com/KARSIFT/karsift-ai-infra/pull/164)
is already merged as `863fc1f35b1d35e4981a59166b0e939be1a2b681`. The remaining
work is a governed caller replacement that consumes that exact merge, not a
third VOC-127 implementation attempt and not publication of stale PR #1041.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-127 task | [#1039](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1039) (`VOC-127-T00`); origin [#1035](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1035) |
| Unpublishable caller PR | [#1041](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1041) at `9d459813a733f9c6d58ad3352df0db27d33ee7f4` |
| Pipeline run that exhausted attempt 2/2 | [`33058603158`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33058603158) |
| Attempt-2 artifact | `implement-VOC-127-VOC-127-T00-attempt2` (id `9640920826`) |
| Artifact bundle SHA-256 | `724e5547e29283b2701b70b72e000253a5100edf77545807610fce17151f7906` |
| Artifact exact head | `bbdb93aadec461830435771490f5c79ba524fed9` |
| Stale pin on that artifact | `a9df74a63976d5239b84151fd01310835c999e7c` (infrastructure #163) |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| Current `develop` pin at drafting | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| VOC-127 retry budget | final allowed retry already consumed; attempt `3` is forbidden |

Root cause: VOC-127 attempt 2 was bound to original package evidence naming
infrastructure PR #163. A blocking exact-revision review then required causal
infrastructure PR #164 (absent-`develop` preflight / checkout-ref ordering).
The final caller remediation copied only part of the earlier fix and retained
the #163 pin. The deterministic caller suite did not enforce equality with the
newly merged authoritative infrastructure revision. Publication failure made
the bundle inspectable, but final-attempt policy forbids another VOC-127 retry.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one implementation PR:

1. Recreate the complete intended VOC-127 caller diff from current `develop`.
   Do not compose or publish stale #1041 / #1039 / the attempt-2 artifact.
2. Mirror every in-scope infrastructure file needed by the caller from exact
   merge `863fc1f35b1d35e4981a59166b0e939be1a2b681`.
3. Set `PINNED_SHA.txt` and every live pin assertion/evidence statement to
   that exact merge. Do not pin `a9df74a6…` or `60afda3a…`.
4. Carry the caller pipeline, docs, and tests required by the already-approved
   VOC-127 release-sync outcome, including live `reconcile-production-change`
   dispatch from #164.
5. Add deterministic regression coverage for the #164 checkout-ref
   ordering/missing-`develop` path and exact pin consistency.
6. Preserve current `config/roles.yml` bindings. Do not add OpenAI execution
   or credentials.
7. After the replacement merges and is promoted, close #1041 as superseded
   (never merged), then close #1039 and root #1035 with audit comments naming
   the replacement merge. Do not manufacture a VOC-127 completion marker from
   #1041.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Recreate the VOC-127 caller contract from current develop, pin exact infra #164, and supersede unpublishable #1041 | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates the
protected caller release-sync dispatch contract (`.github/workflows/*` is an
R3 floor; semantic mutation of fixture pin identity and exceptional
production-change dispatch is protected). The path-based classifier and
independent verifier remain authoritative; this draft proposal is not a
determination.
