# VOC-135 — Pin exact infrastructure PR-context validation and remove all VOC-112 hydration bypasses

| Field | Value |
|-------|-------|
| Package | `VOC-135` |
| Title | Pin exact infrastructure PR-context validation and remove all VOC-112 hydration bypasses |
| Path | `specs/changes/VOC-135-pin-exact-infrastructure-pr-context-validation` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1071](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1071) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-134 was adopted to pin exact infrastructure
[KARSIFT/karsift-ai-infra#166](https://github.com/KARSIFT/karsift-ai-infra/pull/166)
(`f3d79177bf8a9abe0dae550f39502165d494c576`) from current `develop` with a
complete VOC-112 no-change boundary. VOC-134-T00 issue
[#1068](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1068) /
PR [#1070](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1070)
exhausted attempts 1 and 2. Exact reviewed heads:

| Attempt | Head | Independent result |
|---------|------|--------------------|
| 1 | `f373d9a61f90659e982173395a96ff54776fc945` | FAIL: added an import-time capture fetch in `scripts/foundation/voc112-navigation-benchmark-run.mjs` |
| 2 | `718bfdcdaec65e5df6e918fbc25f0fe64f6a2cad` | FAIL: restored that runner but added `scripts/foundation/hydrate-voc112-git-objects.mjs` and invoked it from `scripts/foundation/validate-workspace.mjs` |

The original pre-push failure was truthful: in a full checkout, missing
squash-era subject `f9d11e232a07c7d7a9c433d02c9267912543ba10` caused the
protected test to fail with `a full local checkout must already contain the
captured commit`. The old implementation pre-push boundary ran full PR-style
checks under default local provenance, so self-correction repeatedly
introduced forbidden evidence-object hydration helpers.

Infrastructure PR
[KARSIFT/karsift-ai-infra#167](https://github.com/KARSIFT/karsift-ai-infra/pull/167)
fixes the authoritative boundary without fetching evidence:
`run-app-checks.sh` accepts and validates an immutable PR base/head pair,
selects `pr-validation` for an unchanged capture fixture, `pr-ancestry` for
add/modify/delete, fails closed on comparison errors, reusable CI checks out
full reachable history, and implementation pre-push uses the integration
anchor plus current committed HEAD including after self-correction.
Independent Terra exact review PASS for infra head
`eb11c4fc6841ec73816e2e064dcd449d98c1e933`; merge SHA is
`b263c0c110591cc798b89277dfc35542abb1597b`. Hosted actionlint, policy-tests,
shellcheck, and YAML parse passed; full infra suite passed 442/442.

VOC-134 cannot be redispatched. Do not reuse PR #1070. This package is the
governed replacement authority for the still-unfinished pin plus release
repair, now consuming exact #167 so the caller no longer needs hydration
bypasses.

Live tracked `PINNED_SHA.txt` at drafting still equals infrastructure #164
(`863fc1f35b1d35e4981a59166b0e939be1a2b681`) because VOC-134 never merged.
Issue #1071's "#166 pin" names VOC-134's unmerged target. This package pins
to #167, which already contains the #166 release-synchronization files.

### Live reproduction

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| VOC-132 task that must not be redispatched | [#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059) (`VOC-132-T00`) |
| Exhausted VOC-133 task | [#1063](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1063) (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | [#1065](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065) |
| Adopted VOC-134 plan | [#1067](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1067) |
| Exhausted VOC-134 task | [#1068](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1068) (`VOC-134-T00`) |
| Exhausted VOC-134 carrier | [#1070](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1070) |
| VOC-134 attempt-1 FAIL head | `f373d9a61f90659e982173395a96ff54776fc945` |
| VOC-134 attempt-2 FAIL head | `718bfdcdaec65e5df6e918fbc25f0fe64f6a2cad` |
| VOC-129 caller merge | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Authoritative infra merge | `KARSIFT/karsift-ai-infra@b263c0c110591cc798b89277dfc35542abb1597b` (#167) |
| Independently reviewed PASS infra head | `eb11c4fc6841ec73816e2e064dcd449d98c1e933` |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| VOC-134 unmerged pin target | `f3d79177bf8a9abe0dae550f39502165d494c576` (infrastructure #166) |
| Expected immutable carrier base at drafting | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |

Root cause of the VOC-134 exhaustion: hosted `implement.yml` pre-push still
invoked `run-app-checks.sh` with no PR base/head pair, so default `local`
provenance ran against a full checkout missing squash-era subject
`f9d11e23…`. Self-correction then added fetch/hydrate helpers instead of
leaving the protected fail-closed test alone. #167 supplies the missing
immutable PR context so that path is no longer taken for an unchanged
capture fixture.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from current `develop` (do not reuse PR #1051, #1056, #1065, or #1070, and
do not redispatch VOC-132-T00, VOC-133-T00, or VOC-134-T00):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   infra #167 merge `b263c0c110591cc798b89277dfc35542abb1597b`. It must equal
   neither #164 `863fc1f…` nor #165 `8ce2b77…` nor #166 `f3d791…`.
2. Mirror every necessary authoritative #164→#167 caller fixture surface
   byte-for-byte, including the named #167 files
   (`.github/workflows/ci.yml`, `.github/workflows/implement.yml`,
   `config/run-app-checks.sh`, `tests/test_app_check_context.py`, and the
   infra README PR-context documentation contract) plus every #166 file still
   required by the release synchronization outcome (`release.yml`,
   `config/implementer_nested_checkout.py`, `tests/test_release_policy.py`,
   `tests/test_voc121_implement_policy.py`,
   `tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md`), with
   committed SHA-256 constants and no runtime `/tmp` checkout dependency.
3. Preserve #165 post-caller-checkout release-policy restoration and #166
   helper-copy-before-model / nested-checkout classification.
4. Keep the eight no-change paths byte-identical to the immutable
   carrier-base SHA and absent from the implementation diff. JSON
   `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
5. Do not add any caller-side capture/evidence Git-object fetch,
   hydrate/materialize helper, package/test wrapper, import side effect,
   provenance-mode override, environment override, evidence mutation/stamping
   helper, skip, or equivalent under any filename.
6. Add a complete caller-diff scan of changed executable paths, not only one
   prohibited filename. It must fail if any new or changed caller script
   fetches capture subjects, sets VOC-112 provenance variables around tests,
   hydrates evidence objects, or makes local fail-closed unreachable.
7. Bind the no-change regression to the immutable carrier-base SHA selected
   before implementation (expected `b9e74fc2db4691c48c637639b265d527de9f4505`).
   Revalidate current `develop` at dispatch and fail closed if it moved.
8. Define feasible exact-revision evidence: committed package evidence records
   carrier base, exact infra merge, hashes, validation, and the contract that
   the final implementation head is bound by the App-authored
   independent-review comment/check. Do not require a commit to contain its
   own SHA.
9. Leave `config/roles.yml` unchanged. Add no OpenAI route.
10. After exact-SHA PASS and protected checks, merge through KARSIFT, continue
    through staging only for the real tree change, promote `develop` to
    `main`, advance `develop` to the exact resulting main merge SHA without a
    staging redeploy for tree-equivalent sync, complete production deployment
    where selected, and reconcile/close superseded VOC-127 / VOC-130 /
    VOC-131 / VOC-132 / VOC-133 / VOC-134 carriers without manufactured task
    markers.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #167 from current develop with immutable PR-context validation, no VOC-112 hydration bypasses, and complete release repair | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable caller fixture/pin and
`tooling/governance/` test updates belong under the R4 path floor, and
because the change mutates the protected application-check provenance /
pre-push PR-context, implementer nested-checkout / source-carrier, and
release checkout-lifetime contracts (`.github/workflows/*` is an R3 floor if
a live caller workflow must change). The path-based classifier and
independent verifier remain authoritative; this draft proposal is not a
determination.
