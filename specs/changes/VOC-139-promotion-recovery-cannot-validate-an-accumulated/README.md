# VOC-139 — Promotion recovery cannot validate an accumulated develop-to-main PR

| Field | Value |
|-------|-------|
| Package | `VOC-139` |
| Title | Promotion recovery cannot validate an accumulated develop-to-main PR |
| Path | `specs/changes/VOC-139-promotion-recovery-cannot-validate-an-accumulated` |
| Status | `draft` |
| Risk | `R4` (draft proposal; provenance hash contract and recovery metadata) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1096](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1096) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed `develop` → `main` release remains fail-closed after VOC-138
merged because both ordinary promotion validation and its recovery entry
point fail on the live accumulated promotion PR
[#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090).

| Item | Value |
|------|-------|
| VOC-138 caller merge | [PR #1095](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1095) → `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090) |
| Release issue | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) |
| PR base (`main`) | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| PR head (`develop`) | `4812fb91ab1b674f9a9ec03906f90c0edf50421d` |
| PR run | `33130426061` |
| Failed PR job | `98718413924` |
| Selected mode | `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` (VOC-138 intent) |
| Failing tests | `VOC-112-TEST-12`, `VOC-112-TEST-13` |
| Fail-closed message | `AGENTS.md hash must be anchored in the PR merge base` |
| Stored/head hash prefix | `b0e65629…` |
| Main/base hash prefix | `5ba216ff…` |
| Recovery dispatch | `gh workflow run pipeline.yml --ref develop -f action=reconcile-release -f release_issue_number=1089` |
| Release run | `33130473438` |
| Recovery run | `33130527834` (`promotion-pr-validation PR #1090`) |
| Failed recovery job | `98718739912` |
| Recovery failure | `gh pr view` → `fatal: not a git repository` (no `-R` / `--repo`, no checkout) |
| Issue-creation pin | `123735c80fec813a5b46a004f3e1122bd425cde2` |

No production promotion occurred. PR #1090 remains open and blocked.

## Root cause

Two independent defects remain in the newly activated promotion path:

1. `pr-validation` always requires stored VOC-112 source hashes to equal the
   PR merge-base files. That is correct for an ordinary feature PR whose
   hashed sources are unchanged between base and head. It rejects an
   accumulated long-lived `main` ← `develop` promotion whenever `AGENTS.md`
   or the navigator skill legitimately changed on `develop`.
2. `.github/workflows/pipeline.yml` job `promotion-pr-metadata` (and the
   matching infrastructure template) invokes `gh pr view` before any
   checkout and without `-R "$GITHUB_REPOSITORY"`, so GitHub CLI cannot
   resolve repository context. Adding `-R` alone would still read the absent
   `.headRepository.nameWithOwner` field as `null`; supported owner/name or
   pull REST full-name fields are required.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. Define and implement a promotion-specific immutable source-hash rule for
   authenticated same-repository `main` ← `develop` `pr-validation`: stored
   capture hashes must bind to the reviewed head/source revision
   (`PR_HEAD_SHA` and the working tree), not to historical `main`. Keep
   supplying exact PR base/head SHAs. Do not switch the promotion PR to
   `--squash-safe-push`.
2. Keep ordinary (non-promotion) `pr-validation` merge-base hash anchoring
   and ordinary fixture-changing `pr-ancestry` fail-closed.
3. Make the no-checkout metadata query repository-explicit in the live caller
   workflow, infrastructure template, and mirrored fixture. Validate
   owner/repository identity from supported live response fields (not absent
   `.headRepository.nameWithOwner`) and exercise the real step in a test that
   has no git repository and includes a same-name fork negative.
4. Keep negatives: malformed SHA, unrelated repository/PR, wrong refs,
   unrelated commits, tampered current hashes, missing/nonancestor subject
   for ordinary `pr-ancestry`.
5. Pin the caller fixture to the new independently reviewed infrastructure
   merge. Keep the seven remaining VOC-112 no-change paths byte-identical to
   `b9e74fc2db4691c48c637639b265d527de9f4505`. The provenance test is in
   scope.
6. After the exact reviewed caller merge, `reconcile-release` for #1089 can
   merge #1090 (or the live promotion at the then-current `develop` head).
   Bind closure of #1096 to allowlisted metadata from the successful
   recovery/release run. Do not snapshot the develop/main gap
   (`karsift-ai-infra#15`).

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Unblock accumulated promotion PR hash validation and no-checkout recovery metadata | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, provenance-test, and recovery-metadata updates belong under the R4 path
floor, and because the change mutates application-check provenance and
required-check recovery. The path-based classifier and independent verifier
remain authoritative; this draft proposal is not a determination.
