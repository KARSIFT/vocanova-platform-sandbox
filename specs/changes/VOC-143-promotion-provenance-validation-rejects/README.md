# VOC-143 — Promotion provenance validation rejects legitimate AGENTS.md change under squash-safe-push

| Field | Value |
|-------|-------|
| Package | `VOC-143` |
| Title | Promotion provenance validation rejects legitimate AGENTS.md change under squash-safe-push |
| Path | `specs/changes/VOC-143-promotion-provenance-validation-rejects` |
| Status | `draft` |
| Risk | `R3` (draft proposal; required-check provenance for develop-to-main promotion) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1120](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1120) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The automatic VOC-142 `develop` → `main` promotion PR is blocked because VOC-112
capture provenance still treats a legitimate current `AGENTS.md` documentation
update as a fixture mismatch while the historical capture is intentionally
unmodified.

| Item | Value |
|------|-------|
| Promotion PR | [#1119](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1119) |
| Exact promotion head | `376e00dd769253d7a255660f5391fb208781e2f3` |
| Failing required checks | `validate`, `ci / ci` |
| Release audit | [#1118](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1118), open |
| Prior package | VOC-142 (`AGENTS.md` documentation update; VOC-112 fixtures not recaptured per DEP-07) |
| Live pin | `8993e867640dfb604dec0466c4e0787e68d8e258` (not this defect) |

## Root cause

VOC-142 narrowed the benchmark working-tree `AGENTS.md` exception only for
`pr-validation`. Two promotion-path assertions still require the historical
fixture `agents_sha256` to equal the current tree:

1. `validate` (`.github/workflows/repository-governance.yml`) selects
   `squash-safe-push` for the same-repository `main` ← `develop` promotion PR
   and for push/non-PR events. `assertCapturedRevision` still requires
   `agents_sha256 == sha256("AGENTS.md")` in that mode.
2. `ci / ci` uses `--promotion-pr` → `pr-validation` with
   `VOC112_PROMOTION_PR=true`. `assertPrValidationMergeBase` still requires
   fixture `agents_sha256` to equal `AGENTS.md` at `PR_HEAD_SHA`.

Ordinary task-PR `pr-validation` passed because it is merge-base anchored.
Promotion therefore fails deterministically after a governed `AGENTS.md`
documentation update that correctly leaves historical fixtures unmodified.

## Required outcome (summary)

Use one largest-safe coherent task and one implementation PR:

1. `squash-safe-push` must validate the historical capture against an immutable
   Git ancestor whose `AGENTS.md` matches the fixture hash, not the current
   working tree.
2. Promotion `pr-validation` must use the same historical-ancestor bind for
   `agents_sha256` so `ci / ci` on a promotion PR also passes. Navigator
   HEAD-binding and ordinary merge-base `pr-validation` stay unchanged.
3. `local`, `pr-ancestry`, and ordinary `pr-validation` stay fail-closed as
   today. Fixtures are not recaptured. Promotion check identity is not switched
   or bypassed.
4. After the exact reviewed caller merge, recover #1118 through documented
   `reconcile-release`. Do not snapshot the develop/main gap
   (`karsift-ai-infra#15`). Do not create a duplicate promotion PR or release
   audit.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Bind promotion-path VOC-112 `AGENTS.md` provenance to an immutable historical ancestor | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R3** because it mutates required-check provenance used
before production promotion (CI/CD / governance enforcement). The path
classifier may report R1 for `scripts/foundation/*.mjs`; the builder raises to
R3. Editing `repository-governance.yml` would raise the path floor to R4 and is
out of scope unless SHA passing is proven insufficient. The path-based
classifier and independent verifier remain authoritative; this draft proposal
is not a determination.
