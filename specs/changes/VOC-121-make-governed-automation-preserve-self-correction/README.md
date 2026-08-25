# VOC-121 — Make governed automation preserve self-correction helpers and finish coordinated cross-repository tasks

| Field | Value |
|-------|-------|
| Package | `VOC-121` |
| Title | Make governed automation preserve self-correction helpers and finish coordinated cross-repository tasks |
| Path | `specs/changes/VOC-121-make-governed-automation-preserve-self-correction` |
| Status | `draft` |
| Risk | `R4` (draft proposal; CI/CD orchestration, second-repository publication, and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#994](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/994) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-120 remained a valid adopted task and was eventually completed, but KARSIFT
could not finish its own authorized work without manual recovery. Issue #994
records three related autonomous-delivery failures on the live governed path.

### 1. Authorized infrastructure edits are silently discarded

| Observation | Result |
|-------------|--------|
| Adopted task | `VOC-120-T00` dispatched through `pipeline.yml` / reusable `implement.yml` |
| Implementation run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32899479806 (job `97970681892`) |
| Implementer behavior | Correctly edited both the caller and the nested `karsift-ai-infra` checkout |
| Publishing contract | Copies only `config/run-app-checks.sh`, runs `rm -rf karsift-ai-infra`, stages a caller-only Git bundle |
| Result | Caller PR published; every authorized infrastructure edit disappeared; no infrastructure PR existed until manual recovery under the same task |

Root cause: `prompts/implement.md` already permits coordinated repository
carriers, but `implement.yml` treats the nested infrastructure checkout only as
disposable policy input. There is no governed representation, isolated
commit/bundle, publisher, PR, review, merge, or exact-SHA reconciliation for a
second repository.

### 2. Pre-push self-correction deletes the helper it later invokes

From the same job: `Commit implementer's work` copies `/tmp/run-app-checks.sh`
and removes `karsift-ai-infra`. After pre-push validation fails, `Self-correct
failed pre-push validation` runs
`python3 karsift-ai-infra/config/prepare_cursor_model.py ...` and fails closed
with `can't open file .../karsift-ai-infra/config/prepare_cursor_model.py`.

Root cause: step ordering/lifetime mismatch. Only the app-check script is
preserved before deleting the nested checkout.

### 3. Promotion recovery mistakes an alternate successful status for a required failed check-run

| Observation | Result |
|-------------|--------|
| Promotion PR | [#993](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/993) |
| Exact-head `governance-policy` pull-request check-run | cancelled by concurrency |
| Alternate evidence | another workflow-dispatch run and a published status of the same context were successful |
| Release run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32902418610 reported `actions-check-recovery: ... dispatched=none`, attested successful statuses, then failed three merge attempts |
| `gh pr checks --required` | continued to report `governance-policy` failed |
| Actual recovery that worked | rerunning the cancelled exact-head governance run; the App controller then merged #993 without an admin bypass |

Root cause: recovery/attestation considered the context satisfied even though
GitHub ruleset evaluation still had a non-successful required check-run that a
same-named commit status did not override.

## Required outcome (summary)

1. A single adopted task can safely complete all authorized coordinated
   repository carriers end to end, including separate source and caller PRs when
   the repository boundary requires them.
2. Source work is committed and published from isolated repository state; it is
   never smuggled into the caller as a gitlink or silently discarded. If general
   multi-carrier support cannot be made safe in this change, authorized nested
   edits must fail loudly with precise recovery instructions rather than being
   deleted.
3. Cross-repository PR bodies use fully qualified non-closing caller references.
4. Independent review, exact-head checks, merge order, exact source merge SHA
   capture, caller fixture/pin reconciliation, evidence, bounded retry, and
   caller completion remain fail-closed.
5. Self-correction retains immutable access to every model-resolution, retry, and
   check helper it needs after caller staging, without allowing a nested
   checkout/gitlink into caller commits.
6. Required-check recovery uses GitHub's actual branch/ruleset satisfaction
   state. A cancelled or failed required check-run is rerun or redispatched on
   the unchanged exact head even when another run or same-named status is
   successful.
7. Deterministic tests reproduce all three live failures. Current-state
   comments/docs and the caller fixture/pin follow the exact reviewed
   infrastructure merge.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, one bounded
remediation retry, and release gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Preserve self-correction helpers, finish coordinated carriers without silent loss, and recover required checks from GitHub satisfaction state | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because implementer publication,
App-token scope, and required-check recovery change protected CI/CD and agent
authority. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
