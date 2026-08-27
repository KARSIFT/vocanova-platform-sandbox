# VOC-136 — Complete infra #167 caller pin with exhaustive executable bypass scanning

| Field | Value |
|-------|-------|
| Package | `VOC-136` |
| Title | Complete infra #167 caller pin with exhaustive executable bypass scanning |
| Path | `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive` |
| Status | `draft` |
| Risk | `R4` (draft proposal; caller fixture/pin and `tooling/governance/` tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1076](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1076) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The release-synchronization infrastructure fix is merged in
`KARSIFT/karsift-ai-infra` as exact merge
`b263c0c110591cc798b89277dfc35542abb1597b` (PR
[#167](https://github.com/KARSIFT/karsift-ai-infra/pull/167); independently
reviewed head `eb11c4fc6841ec73816e2e064dcd449d98c1e933`). The caller is
still pinned to infra #164 `863fc1f35b1d35e4981a59166b0e939be1a2b681`
because VOC-135 / PR
[#1075](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075) was
deliberately stopped unmerged after its second attempt weakened a required
regression.

Current caller refs at issue creation:

| Ref | SHA |
|-----|-----|
| `develop` | `5044d62f6f35069412e8ffcf80a0682e7847d1c1` |
| `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |

PR #1075 is closed/unmerged; its remote branch is deleted. VOC-135 task
[#1073](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1073)
and root [#1071](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1071)
are closed not-planned with audit; no completion marker exists.

### Exact VOC-135 failure history (PR #1075)

| Attempt | Head | Result |
|---------|------|--------|
| 1 | `7c3e821bb3228c0fd2c7279b6d577a388148cb18` | Self-failed TEST-12 because the complete-diff scan matched its own assertion literal. Hosted validate and supervisor execution both ran 240 tests with 1 failure. Independent exact review FAIL: [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444183121) |
| 2 | `dcd2da9ccbdf471c5294db30f09c9bb99aacdcd1` | Passed existing suites but added `"tooling/governance/tests/"` to `SCAN_EXCLUDE_PREFIXES`. That excludes every changed/added caller test Python executable from scanning. Those modules execute during hosted unittest discovery and could perform an import-time capture fetch, evidence hydration/materialization, or provenance-environment override before later checks. Supervisor finding: [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444311947) |

The exact-head reviewer acknowledged this residual path but classified it
Low and returned PASS WITH NON-BLOCKING FINDINGS:
[comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444342914).
Operator supervision converted the PR to draft before auto-merge, then
closed it unmerged. **Do not treat that PASS as sufficient for a replacement
that retains the exclusion.**

### Prior exhausted carriers still in the audit trail

VOC-134-T00 ([#1068](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1068)
/ PR [#1070](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1070))
exhausted attempts 1 and 2 on an import-time capture fetch and a relocated
hydrate helper. VOC-133 exhausted on a `package.json` provenance wrapper
plus Git-impossible self-SHA evidence. VOC-130 / VOC-131 exhausted on
VOC-112 JSON retargets and provenance-test weakening. VOC-132 produced no
publishable caller carrier. Those packages remain closed/not-retried
history. This package is the governed replacement for the still-unfinished
#167 pin plus the exhaustive scan VOC-135 failed to land.

Live tracked `PINNED_SHA.txt` at drafting still equals infrastructure #164
(`863fc1f35b1d35e4981a59166b0e939be1a2b681`) because neither VOC-134 nor
VOC-135 merged. This package pins to #167, which already contains the #166
release-synchronization files and the #167 immutable PR-context contract.

Root cause of the original pre-push failure remains truthful: in a full
checkout, missing squash-era subject
`f9d11e232a07c7d7a9c433d02c9267912543ba10` caused the protected test to fail
with `a full local checkout must already contain the captured commit`.
Infrastructure #167 supplies the missing immutable PR context so that path
is no longer taken for an unchanged capture fixture. The caller still has
to consume that merge without weakening the complete-diff scan.

## Required outcome (summary)

Use one largest-safe coherent replacement task and one fresh implementation
PR from freshly revalidated current `develop` (do not reuse PR #1051, #1056,
#1065, #1070, or #1075, and do not redispatch VOC-132-T00 through
VOC-135-T00):

1. Pin `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` exactly to
   infra #167 merge `b263c0c110591cc798b89277dfc35542abb1597b`. It must equal
   neither #164 `863fc1f…` nor #165 `8ce2b77…` nor #166 `f3d791…`.
2. Mirror every necessary authoritative #164→#167 caller fixture surface
   byte-for-byte, including `.github/workflows/ci.yml`,
   `.github/workflows/implement.yml`, `.github/workflows/release.yml`,
   `config/run-app-checks.sh`, `config/implementer_nested_checkout.py`,
   `tests/test_app_check_context.py`, `tests/test_release_policy.py`,
   `tests/test_voc121_implement_policy.py`,
   `tests/test_voc123_source_bundle.py`, fixture `CHANGELOG.md`, and
   current-state caller fixture README documentation, with committed
   SHA-256 constants and no runtime `/tmp` checkout dependency.
3. Preserve #164 missing-`develop` recovery and exact promotion merge
   synchronization, #165 post-caller-checkout restore, #166 helper/nested-
   checkout lifetime, and #167 immutable PR base/head validation.
4. Keep the eight no-change paths byte-identical to protected comparison
   anchor `b9e74fc2db4691c48c637639b265d527de9f4505` and absent from the
   implementation diff against that anchor. JSON `subject_revision` remains
   `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Record the implementation PR
   base separately. Fail closed on unrelated/material movement of `develop`.
   This package's own plan/adoption/roster commits do not count as
   protected-file drift.
5. Do not add any caller-side capture/evidence Git-object fetch,
   hydrate/materialize helper, package/test wrapper, import side effect,
   provenance-mode override, environment override, evidence mutation/stamping
   helper, skip, or equivalent under any filename.
6. Add an **exhaustive** complete-diff scan of every added or modified
   executable path against the protected comparison anchor. Scan all
   `scripts/**`, `package.json`, every added/changed `*.mjs` / `*.js` /
   `*.sh` / `*.py` anywhere in the caller tree except the mirrored
   `tooling/governance/fixtures/karsift-ai-infra/**` subtree, and any other
   newly executable caller file. **Do not** exclude
   `tooling/governance/tests/**`, this regression's own module, or another
   executable directory wholesale. Represent test/assertion literals as
   non-contiguous source-safe values so self-scanning a tracked committed
   module does not false-positive; actual semantic setters/commands must
   fail. Add the required negative unit cases. Prove benign
   assertion/test-data mentions do not false-positive. The regression must
   pass after it is tracked and committed, not merely while untracked.
7. Bind exact-revision evidence as App-authored independent-review
   comment/check metadata. Do not require a commit to contain its own SHA.
8. Leave `config/roles.yml` unchanged. Add no OpenAI route.
9. After exact-SHA PASS and protected checks, merge through KARSIFT, continue
   through staging only for the real tree change, promote `develop` to
   `main`, advance `develop` to the exact resulting main merge SHA (0
   ahead / 0 behind, identical tree) without a staging redeploy for
   tree-equivalent sync, complete production deployment where selected, and
   reconcile/close superseded VOC-127 / VOC-130 through VOC-135 carriers
   without manufactured task markers. Clean only obsolete remote automation
   branches/PRs/issues; preserve unrelated VOC-128 and all user worktrees.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task. Infrastructure #167 is already
merged; do not open another infra PR.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Pin exact infra #167 from current develop with immutable PR-context validation, exhaustive executable bypass scanning, and complete release repair | — |

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
