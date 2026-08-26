# VOC-126 — VOC-125 caller pipeline exceeds GitHub workflow_dispatch input limit

| Field | Value |
|-------|-------|
| Package | `VOC-126` |
| Title | VOC-125 caller pipeline exceeds GitHub workflow_dispatch input limit |
| Path | `specs/changes/VOC-126-voc-125-caller-pipeline-exceeds-github-workflow` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller dispatch contract, project-repo template, and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1025](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1025) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-125 added the required `existing_pr_number` operator-resume input to the
caller pipeline and the authoritative project-repo template, but the caller
already had 25 `workflow_dispatch` inputs. The resulting 26-input workflow is
rejected by GitHub before any job can start.

This blocks the adopted VOC-125 caller carrier
[#1024](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1024) at
exact head `8621f12dd466edab37fddb86d4e5e0a348ed3609`, after authoritative
infrastructure PR [KARSIFT/karsift-ai-infra#160](https://github.com/KARSIFT/karsift-ai-infra/pull/160)
merged as `1f1705dbad41729563b0ad1e878e4154e5511e93`.

### Live reproduction (2026-08-26)

| Item | Value |
|------|-------|
| Adopted VOC-125 task blocked | [#1022](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1022) (`VOC-125-T00`); origin [#1020](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1020) |
| Unusable caller PR | [#1024](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1024) at `8621f12dd466edab37fddb86d4e5e0a348ed3609` |
| Infrastructure merge that encoded the invalid template | `KARSIFT/karsift-ai-infra@1f1705dbad41729563b0ad1e878e4154e5511e93` |
| Actions run | [`32977045898`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32977045898) — terminates with no jobs or logs |
| Authenticated annotation | `Invalid workflow file: .github/workflows/pipeline.yml#L1 — you may only define up to 25 inputs for a workflow_dispatch event` |
| Retrigger | draft/ready retriggering on the same exact SHA produces no pipeline run |
| Local suites | fixture, governance, and foundation suites pass because they assert the new input exists and is forwarded, but do not assert GitHub's maximum input count |
| VOC-125 retry budget | the final allowed VOC-125 implementation retry was already consumed remediating the exact source carrier after its first blocking review; attempt `3` is forbidden |
| Downstream carrier still waiting | VOC-122-T00 [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) / draft PR [#1012](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1012) cannot be resumed until a live valid caller `action=implement` route exists |

Root cause: VOC-125 correctly added `existing_pr_number`, but its source/template
and caller tests omitted the platform-level `workflow_dispatch` input-count
constraint. Independent source review also missed it. Removing or repurposing
an unrelated dispatch input is not authorized by VOC-125 and would regress an
existing recovery/verifier contract. PR #1024 cannot reach its required
exact-revision pipeline review in its current form, so the existing carrier is
genuinely unusable rather than merely stale.

## Required outcome (summary)

Use one largest-safe coherent task to restore a live, GitHub-valid caller
dispatch contract that still carries VOC-125's operator-resume identity:

1. Preserve `existing_pr_number` and all existing fail-closed exact-head/base,
   two-attempt, review, risk, protected-check, App-token, and Cursor-only
   contracts. Do not add operator-typed SHA inputs.
2. Relocate the coherent read-only verifier dispatch surface into a dedicated
   caller workflow so every live `workflow_dispatch` block has at most 25
   inputs. Do not delete an active recovery or verifier capability. Do not
   silently drop an input merely to get below 26.
3. Update the authoritative `KARSIFT/karsift-ai-infra` project-repo template
   and deterministic source tests first, including an explicit maximum-input-count
   regression. Independently review and merge that source repair, then record
   its exact merge SHA.
4. Update the live caller workflow, fixture, exact pin, tests, documentation,
   and evidence to that merge. Supersede caller PR #1024 only because its
   definition cannot enter the governed review pipeline; preserve its audit
   trail and close it only after the governed replacement exists and is
   reconciled.
5. Close VOC-125 task #1022 / origin #1020 only when the live caller route is
   valid, reviewed, merged, and promoted. Then resume existing VOC-122
   #1003 / #1012 with `attempt=2` and `existing_pr_number=1012`.
6. Do not implement VOC-122 behavior, create a replacement VOC-122 task or PR,
   enable OpenAI, add credential routes, weaken controls, or exceed the
   one-retry policy.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task. No OpenAI credential or execution
route is needed or authorized. Do not print credential values.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Relocate read-only verifier dispatch, land `existing_pr_number` under the 25-input limit, with tests, docs, caller pin, and VOC-125 / VOC-122 handoff | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates the
protected caller dispatch contract and project-repo pipeline template
(`.github/workflows/*` is an R3 floor; semantic mutation of recovery/verifier
dispatch identity is protected). The path-based classifier and independent
verifier remain authoritative; this draft proposal is not a determination.
