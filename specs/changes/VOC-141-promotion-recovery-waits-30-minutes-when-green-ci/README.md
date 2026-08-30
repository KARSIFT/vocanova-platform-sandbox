# VOC-141 — Promotion recovery waits 30 minutes when green CI has an unattestable parent

| Field | Value |
|-------|-------|
| Package | `VOC-141` |
| Title | Promotion recovery waits 30 minutes when green CI has an unattestable parent |
| Path | `specs/changes/VOC-141-promotion-recovery-waits-30-minutes-when-green-ci` |
| Status | `draft` |
| Risk | `R4` (draft proposal; required-check recovery dispatch and timeout diagnostics) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1109](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1109) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Promotion recovery correctly rejects a green `ci / ci` check whose parent
pipeline is a release carrier, but it does not dispatch dedicated
`promotion-pr-validation` when GitHub's required-check row is already SUCCESS.
The runner polls until its full 1,800-second timeout with no possible state
transition. This is a causal residual of post-merge VOC-140 release
verification.

| Item | Value |
|------|-------|
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Exact promotion head | `c3a53bab3035b7f08c0fb959bdf1b56bf330d291` |
| Infrastructure policy | `KARSIFT/karsift-ai-infra@67bdfd13ef875dead23ce4be01d7d0e8b976e289` |
| Failing release carrier | run `33340381776`, job `99334840338` |
| Step | `Recover missing exact-head promotion checks` |
| Window | `2026-08-30T22:54:51Z` → `2026-08-30T23:25:23Z` |
| Timeout diagnostics | `missing_checks: none`, `pending: 0`, `failed: 0`, `successful: 6`, then exit 1 |
| Duplicate no-progress carrier | run `33340516672` |
| Workaround dedicated SUCCESS | run `33341923799`, titled `promotion-pr-validation PR #1090`, exact head `c3a53bab…` |
| Subsequent reconcile-release | run `33342062118` succeeded, including the production merge guard |

No branch, PR, label, or code fix was created manually for this residual.

## Root cause

`recovery_complete()` in `config/actions_check_recovery.py` correctly returns
false because `promotion_ci_context_is_attestable()` cannot bind the selected
green `ci / ci` row to an attestable completed non-carrier parent.

`apply_promotion_pr_recovery_plan()` in
`config/actions-check-recovery-runner.py` derives `dispatch_contexts` only from
`plan_required_check_recovery()`. GitHub reports the required `ci / ci` row as
SUCCESS, so that planner returns no missing/failed CI context. The dedicated
`pipeline.yml` `action=recover-promotion-pr-checks` dispatch is filtered out.
The polling loop repeats the same metadata until timeout. Diagnostics omit the
unattestable-parent reason.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. When the required PR view has a SUCCESS `ci / ci` row but the authoritative
   composed evidence has no uniquely attestable non-carrier parent, promotion
   recovery must immediately plan exactly one dedicated
   `recover-promotion-pr-checks` dispatch. It must not rerun a doomed PR
   carrier or wait 1,800 seconds without a dispatch.
2. A valid completed dedicated parent completes recovery without redispatch.
   Active or successful exact dedicated recovery suppresses duplicates.
   Release carriers and other `pipeline.yml` runs must not suppress that
   dispatch.
3. Timeout diagnostics must identify unattestable CI evidence rather than
   `missing_checks: none` alone.
4. Production/release-carrier fail-closed boundaries, VOC-138/VOC-139
   exact-identity recovery, doomed-job rerun refusal, and the VOC-140
   two-token production merge guard remain unchanged.
5. Pin the caller fixture to the new independently reviewed infrastructure
   merge. After the exact reviewed caller merge, ordinary `reconcile-release`
   must not hang on this class. Do not snapshot the develop/main gap
   (`karsift-ai-infra#15`). Do not create a duplicate promotion PR or release
   audit.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Dispatch dedicated promotion-pr-validation immediately when green CI is unattestable | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, and required-check recovery updates belong under the R4 path floor, and
because the change mutates promotion recovery dispatch used before production
merge. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
