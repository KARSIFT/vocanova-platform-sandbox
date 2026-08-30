# VOC-131 — Replacement release blocker: consume infra #165 without rewriting VOC-112 evidence

| Field | Value |
|-------|-------|
| Package | `VOC-131` |
| Title | Replacement release blocker: consume infra #165 without rewriting VOC-112 evidence |
| Path | `specs/changes/VOC-131-replacement-release-blocker-consume-infra-165` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1052](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1052) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed VOC-130 implementation exhausted both attempts without merging.
PR [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051)
correctly implemented the exact infrastructure #165 pin and release fixture,
but both reviewed heads retained two unrelated VOC-112 evidence retargets.
Attempt 2 claimed those retargets were reverted, while exact head
`e846cc2f62556c2d6616282f2e9c4929c00655e7` still contained them. Independent
exact-SHA review therefore failed closed twice.

This replacement is the same still-unfinished release-repair outcome. It must
not reopen or reimplement infrastructure #165, and it must not reuse PR #1051.

The original live defect remains: after caller PR
[#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
merged, `release.yml` needed the shared-policy checkout before caller checkout
so it could resolve an absent-`develop`-safe caller ref, but the root caller
checkout then cleaned nested untracked paths. The task-completion helper was
gone. Release treated that missing validator as a safe no-op, so no release
audit or promotion was created. `converge` has the same latent defect.

Infrastructure follow-up [KARSIFT/karsift-ai-infra#165](https://github.com/KARSIFT/karsift-ai-infra/pull/165)
is already merged as `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. The remaining
work is a governed caller pin/fixture/test/doc update that consumes that exact
merge, without rewriting VOC-112 evidence, and without snapshotting the
develop/main gap.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Attempt-1 reviewed head | `a04a41a19176d1322faee1a3365d82153d61af1b` |
| Attempt-1 review | [comment 5438961796](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5438961796) |
| Attempt-2 reviewed head | `e846cc2f62556c2d6616282f2e9c4929c00655e7` |
| Attempt-2 review | [comment 5439141140](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5439141140) |
| Durable exhaustion marker | [issue #1049 comment 5439143793](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049#issuecomment-5439143793) |
| VOC-129 caller merge | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Observed identify behavior | `release-checkout-ref` selected `develop`; caller root checkout followed |
| Observed failure | `python` could not open `karsift-ai-infra/config/task-completion-runner.py` |
| Observed release result | safe no-op; skipped converge; no release audit or promotion |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |

Root cause of the release no-op: shared policy must be restored from the same
immutable reusable-workflow revision after caller checkout and before later
lifecycle helpers. #164 correctly placed shared-policy checkout before the ref
resolver; it did not restore that nested checkout after the caller owned the
workspace root.

Root cause of VOC-130 exhaustion: both reviewed #1051 heads changed

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`

Those files are out of scope for this pin. They must remain byte-for-byte
identical to the new carrier's `develop` base.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from current `develop` (do not reuse PR #1051 or its branch):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror from exact infra merge #165 only the two changed authoritative
   files: `.github/workflows/release.yml` and `tests/test_release_policy.py`,
   byte-for-byte.
3. Prove both `identify` and `converge` restore
   `${{ job.workflow_repository }}` at `${{ job.workflow_sha }}` into
   `karsift-ai-infra` after caller checkout and before lifecycle helpers, with
   credentials not persisted.
4. Advance the existing pin assertions and fixture README/evidence needed for
   the #165 contract.
5. Add a deterministic regression that fails if either named VOC-112 fixture
   differs from the adopted package / carrier `develop` base, and ensure
   neither file appears in the implementation diff.
6. Preserve all #164 missing-`develop` recovery, exact promotion-merge
   synchronization, unique-develop fail-closed behavior, release
   serialization, promotion checks, retry bounds, review separation,
   `reconcile-production-change`, App-token isolation, and sanitized error
   handling.
7. Keep live `.github/workflows/pipeline.yml` unchanged unless exact source
   comparison proves a required change; it already calls `release.yml@main`.
8. Keep current role bindings unchanged and add no OpenAI execution or
   `OPENAI_API_KEY` request.
9. Run caller governance, full fixture tests, relevant foundation tests,
   exact-pin/byte-identity checks, and `git diff --check`; require independent
   exact-revision review and protected checks before merge.
10. After merge, complete ordinary governed promotion of the outstanding
    VOC-129 work plus this replacement, exact develop synchronization to the
    resulting main merge SHA, production deployment where selected, evidence
    reconciliation, and closure of VOC-129, VOC-130, and this replacement with
    precise audit links.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #165 from current develop, restore shared policy after caller checkout, and keep VOC-112 fixtures identical to the develop base | — |

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
