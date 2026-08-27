# VOC-132 — Release blocker replacement: land infra #165 pin with complete VOC-112 no-change boundary

| Field | Value |
|-------|-------|
| Package | `VOC-132` |
| Title | Release blocker replacement: land infra #165 pin with complete VOC-112 no-change boundary |
| Path | `specs/changes/VOC-132-release-blocker-replacement-land-infra-165-pin` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1057](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1057) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The permanent release synchronization repair is merged in
[KARSIFT/karsift-ai-infra#165](https://github.com/KARSIFT/karsift-ai-infra/pull/165)
at `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, but the caller remains pinned
to infrastructure #164 because two exhausted governed carriers introduced or
retained unrelated VOC-112 changes.

VOC-130-T00 / PR [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051)
exhausted both attempts after exact-SHA review found two VOC-112 evidence JSON
retargets. VOC-131-T00 / PR
[#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) then
exhausted both attempts at head
`c11454e717a6d778143de1f2023acc4480305845`: its pin and mirrored #165 files
are correct, but
`scripts/foundation/voc112-navigation-benchmark.test.mjs` still weakens
missing-subject behavior in a full checkout while its evidence falsely claims
that edit was reverted.

Exact review:
https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439802080.
Supervisor evidence:
https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439768013.

This replacement is the same still-unfinished release-repair outcome. It must
not reopen or reimplement infrastructure #165, and it must not reuse PR #1051
or PR #1056.

The original live defect remains: after caller PR
[#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
merged, `release.yml` needed the shared-policy checkout before caller checkout
so it could resolve an absent-`develop`-safe caller ref, but the root caller
checkout then cleaned nested untracked paths. The task-completion helper was
gone. Release treated that missing validator as a safe no-op, so no release
audit or promotion was created. `converge` has the same latent defect.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| VOC-131 exhausted head | `c11454e717a6d778143de1f2023acc4480305845` |
| VOC-131 exact review | [comment 5439802080](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439802080) |
| VOC-131 supervisor evidence | [comment 5439768013](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439768013) |
| VOC-131 reproduction merge-base | `790ea1b66b096414fe6709e6cd4b8342ff5ac587` |
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

Root cause of VOC-130 exhaustion: both reviewed #1051 heads changed the two
VOC-112 JSON evidence fixtures.

Root cause of VOC-131 exhaustion: replacement packages protected only those
two JSON paths in deterministic diff guards. The implementer could therefore
modify the adjacent VOC-112 provenance test to make local validation pass,
even though the task had no authority to change VOC-112 behavior. On #1056
head `c11454e…`, `local` mode changed from fail-closed on a missing capture
commit to implicit `squash-safe-push`, contrary to VOC-131-D10/R05, while
evidence claimed that edit was reverted.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from current `develop` (do not reuse PR #1051, PR #1056, or their branches):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror only the necessary #165 caller fixture files byte-for-byte,
   preserving the post-caller-checkout shared-policy restore in both release
   jobs and every #164 exact-SHA synchronization/recovery contract.
3. Reuse the valid in-scope implementation shape from #1056 where appropriate
   (pin, mirrored #165 files, restore coverage, pin-literal updates), but on a
   new carrier from current `develop`.
4. Make the complete VOC-112 no-change boundary explicit and deterministic.
   At minimum these paths must be byte-identical to the new carrier base and
   absent from the implementation diff:
   - `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
   - `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
   - `scripts/foundation/voc112-navigation-benchmark.test.mjs`
   - `AGENTS.md`
   - `.agents/skills/vocanova-repo-navigator/SKILL.md`
5. Add a regression that fails if any protected path above differs from the
   immutable carrier base or appears in the PR implementation diff. Do not
   alter VOC-112 provenance behavior, evidence, subject revisions, or hashed
   sources to make a local test pass.
6. Keep current role bindings unchanged and do not add OpenAI. Preserve A-004,
   review separation, retry bounds, required checks, risk floors, secret
   handling, release audit, exact-SHA sync, unique-develop fail-closed
   behavior, idempotent recovery, and tree-equivalent staging suppression.
7. Validate the exact mirrored fixture bytes against infra #165 without
   depending on a machine-specific `/tmp` checkout. Evidence must truthfully
   describe the exact reviewed head.
8. After exact-SHA independent PASS and required checks, merge through
   governance, stage, promote `develop` to `main`, synchronize `develop` to
   the exact promotion merge SHA, deploy production if selected, reconcile
   audit records for VOC-129 and the superseded VOC-130/VOC-131 carriers, and
   close root issue #1057.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #165 from current develop with a complete VOC-112 no-change boundary | — |

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
