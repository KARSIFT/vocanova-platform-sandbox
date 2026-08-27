# VOC-130 — Release blocker: caller checkout deletes shared lifecycle policy

| Field | Value |
|-------|-------|
| Package | `VOC-130` |
| Title | Release blocker: caller checkout deletes shared lifecycle policy |
| Path | `specs/changes/VOC-130-release-blocker-caller-checkout-deletes-shared` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1047](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1047) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

A live VOC-129 task-completion wake-up exposed a fail-closed release defect
after caller PR [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
merged. `release.yml` needed the shared-policy checkout before caller checkout
so it could resolve an absent-`develop`-safe caller ref, but the root caller
checkout then cleaned nested untracked paths. The task-completion helper was
gone. Release treated that missing validator as a safe no-op, so no release
audit or promotion was created. `converge` has the same latent defect.

Infrastructure follow-up [KARSIFT/karsift-ai-infra#165](https://github.com/KARSIFT/karsift-ai-infra/pull/165)
is already merged as `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. The remaining
work is a governed caller pin/fixture/test/doc update that consumes that exact
merge, not a rewrite of VOC-129 and not a snapshot of the develop/main gap.

### Live reproduction

| Item | Value |
|------|-------|
| VOC-129 caller merge | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Failing/no-op release wake-up | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Observed identify behavior | `release-checkout-ref` selected `develop`; caller root checkout followed |
| Observed failure | `python` could not open `karsift-ai-infra/config/task-completion-runner.py` |
| Observed release result | safe no-op; skipped converge; no release audit or promotion |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Infra verification | 429 policy tests plus hosted actionlint, shellcheck, YAML parsing, and policy checks |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |

Root cause: shared policy must be restored from the same immutable
reusable-workflow revision after caller checkout and before later lifecycle
helpers. #164 correctly placed shared-policy checkout before the ref resolver;
it did not restore that nested checkout after the caller owned the workspace
root.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one implementation PR:

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror the exact #165 `release.yml` restore of shared lifecycle policy after
   caller checkout in both `identify` and `converge`.
3. Update every fixture/foundation pin assertion, tests, docs, and task
   evidence required by repository governance.
4. Add deterministic coverage proving both jobs restore the exact shared-policy
   revision after caller checkout and before task-completion helpers.
5. Preserve the #164 missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, review/implementer separation, retry bounds, roles,
   and secret/raw-error controls.
6. Preserve current `config/roles.yml` bindings. Do not add OpenAI execution
   or credentials.
7. After required checks and independent exact-revision review, merge. Then
   complete develop-to-main promotion, exact develop synchronization,
   production deployment where selected, and audit reconciliation for VOC-129
   and this blocker through the ordinary repaired release path. Do not
   snapshot the current develop/main gap (`karsift-ai-infra#15`).

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #165, mirror the shared-policy restore, and cover identify/converge | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates the
protected release checkout-lifetime / task-completion validation contract
(`.github/workflows/*` is an R3 floor if a live caller workflow must change).
The path-based classifier and independent verifier remain authoritative; this
draft proposal is not a determination.
