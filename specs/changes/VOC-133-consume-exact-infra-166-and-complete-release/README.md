# VOC-133 — Consume exact infra #166 and complete release repair after VOC-132 infrastructure failure

| Field | Value |
|-------|-------|
| Package | `VOC-133` |
| Title | Consume exact infra #166 and complete release repair after VOC-132 infrastructure failure |
| Path | `specs/changes/VOC-133-consume-exact-infra-166-and-complete-release` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1061](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1061) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-132 was adopted to pin exact infrastructure
[KARSIFT/karsift-ai-infra#165](https://github.com/KARSIFT/karsift-ai-infra/pull/165)
(`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`). Its first implementation run
[`33079499176`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33079499176)
produced correct caller work but failed before commit/publication because the
implementer removed the untracked nested `karsift-ai-infra/` checkout and the
later shared commit step still copied helpers from that deleted directory. No
caller branch, PR, or completion marker was created.

The lifecycle defect is now fixed through directly reviewed infrastructure PR
[KARSIFT/karsift-ai-infra#166](https://github.com/KARSIFT/karsift-ai-infra/pull/166).
Exact independent Terra review failed initial head
`ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f`, remediation landed on reviewed PASS
head `1488619d0d37aaa179d8e739bfe931881d6c51aa`, all 433 policy tests and hosted
actionlint/shellcheck/YAML checks passed, and the authoritative merge is
`f3d79177bf8a9abe0dae550f39502165d494c576`.

VOC-132 cannot silently change its exact pin from #165 to #166. Do not
redispatch VOC-132-T00
([#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059)), and
do not manufacture a VOC-132 completion marker. This package is the governed
replacement authority for consuming #166 directly.

The caller remains pinned to infrastructure #164 because VOC-130 / VOC-131
exhausted on unrelated VOC-112 changes and VOC-132 never published a caller PR.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| Adopted VOC-132 plan | [#1058](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1058) |
| VOC-132 task that must not be redispatched | [#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059) (`VOC-132-T00`) |
| VOC-132 failed implementer run | [`33079499176`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33079499176), job `98543693163` |
| Observed VOC-132 failure | bounded `cp: cannot stat karsift-ai-infra/config/run-app-checks.sh` after the implementer correctly reported removal of the untracked nested checkout |
| VOC-132 publication result | no caller branch, PR, or completion marker |
| VOC-129 caller merge | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@f3d79177bf8a9abe0dae550f39502165d494c576` (#166) |
| Failed initial infra review head | `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f` |
| Independently reviewed PASS infra head | `1488619d0d37aaa179d8e739bfe931881d6c51aa` |
| Infra initial review | [comment 5440485626](https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440485626) |
| Infra final PASS | [comment 5440515073](https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440515073) |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |

Root cause of the VOC-132 publication failure: post-model lifecycle helpers were
still read from the nested checkout after the implementer had legitimately
removed it. #166 copies those helpers to immutable `/tmp` paths before the
unrestricted model runs, treats an absent nested checkout as no
infrastructure-source change while caller changes continue, and fails closed on
a surviving path that is not a distinct nested Git checkout.

Root cause of the earlier release no-op (run `33066533397`): shared policy must
be restored from the same immutable reusable-workflow revision after caller
checkout and before later lifecycle helpers. That #165 restore remains required
and is included in the #164→#166 fixture delta.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from current `develop` (do not reuse PR #1051 or PR #1056, and do not
redispatch VOC-132-T00):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   infra #166 merge `f3d79177bf8a9abe0dae550f39502165d494c576`. It must equal
   neither #164 `863fc1f35b1d35e4981a59166b0e939be1a2b681` nor #165
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`.
2. Mirror every necessary authoritative #164→#166 caller fixture surface
   byte-for-byte. The infra delta contains `.github/workflows/implement.yml`,
   `.github/workflows/release.yml`, `config/implementer_nested_checkout.py`,
   `tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`, and
   `tests/test_voc123_source_bundle.py`, plus documentation. Update caller
   current-state fixture docs accurately; preserve historical audit content
   unless a true current-state statement requires change.
3. Preserve #165 post-caller-checkout release-policy restoration in both
   identify and converge and all #164 missing-develop / exact-promotion-SHA /
   unique-develop fail-closed / idempotent recovery contracts.
4. Preserve #166 post-implementer behavior: helpers copied before the
   unrestricted model; absent nested checkout means no infrastructure source
   carrier while caller changes continue; a plain subdirectory inheriting
   caller Git, non-directory, or symlink fails closed; a distinct nested Git
   checkout preserves exact-head bundle/ancestry/remote/lease publication.
5. Add deterministic caller fixture regressions and recorded exact hashes for
   all newly mirrored authoritative files without any machine-specific `/tmp`
   checkout dependency.
6. Keep the complete VOC-112 no-change boundary byte-identical to the immutable
   carrier base and absent from the implementation diff.
7. Do not reuse, modify, or merge exhausted PR #1051 or #1056. Do not
   redispatch VOC-132 task #1059. Do not manufacture completion markers for
   VOC-129, VOC-130, VOC-131, or VOC-132. This new PR closes only its own
   generated task issue.
8. Do not change role bindings or add OpenAI. Preserve A-004, R4
   classification, independent exact-SHA review, protected checks, retry caps,
   task/roster marker controls, App-token isolation, sanitized errors, release
   audit, promotion merge commit, exact develop synchronization, and
   tree-equivalent staging suppression.
9. After the exact reviewed caller merge, allow ordinary release (or
   idempotent `reconcile-release`) to promote all qualified work, synchronize
   `develop` to the exact `main` promotion merge SHA, deploy production where
   selected, and reconcile/close VOC-129 and superseded VOC-130 / VOC-131 /
   VOC-132 records plus this root issue with truthful audit evidence.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #166 from current develop with complete VOC-112 no-change boundary and complete the release repair | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/pin and
`tooling/governance/` test updates belong under the R4 path floor, and because
the change mutates the protected implementer nested-checkout / source-carrier
and release checkout-lifetime contracts (`.github/workflows/*` is an R3 floor
if a live caller workflow must change). The path-based classifier and
independent verifier remain authoritative; this draft proposal is not a
determination.
