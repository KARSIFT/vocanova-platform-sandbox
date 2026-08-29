# VOC-140 — Release convergence cannot trust its own CI run or verify the production merge guard with the App token

| Field | Value |
|-------|-------|
| Package | `VOC-140` |
| Title | Release convergence cannot trust its own CI run or verify the production merge guard with the App token |
| Path | `specs/changes/VOC-140-release-convergence-cannot-trust-its-own-ci-run` |
| Status | `draft` |
| Risk | `R4` (draft proposal; recovery identity and production-merge-guard token/API contract) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1102](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1102) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed `develop` → `main` release for release audit
[#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089) /
promotion PR [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090)
remains fail-closed after VOC-139 repaired promotion validation. Two causally
related release-convergence defects now block the same end-to-end outcome.

| Item | Value |
|------|-------|
| `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| `develop` / PR #1090 head | `21eef75549226766fc4f78f62f232ee5fbdb8d6d` |
| Promotion PR | [#1090](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1090), open, `main` ← `develop` |
| Release audit | [#1089](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1089), open |
| Issue-creation pin | `599436835371f27fac52ec6b47a18b36257366ac` |
| Dedicated recovery (success) | run `33136865709`, exact-head/PR-bound |
| Circular-CI run / job | `33136633666` / `98738317266` |
| Circular-CI fail-closed | `untrusted_ci_recovery_identity` |
| Guard-fail release runs / jobs | `33136984634` / `98739074178`; `33137091931` / `98739420310` |
| Guard fail-closed | `production-merge-guard: production_merge_guard_missing` |
| Admin-visible verifier | `production-merge-guard: ok ruleset_id=20575146` |
| Live ruleset | `20575146`, repository-owned, active, `main`-only, strict, required PRs, `governance-policy` / `validate` / `ci`, `bypass_actors: []` |
| Production promotion | none |

No duplicate promotion PR or audit issue may be created.

## Root cause

1. **Circular CI recovery identity.** Promotion recovery reports complete
   without dispatch when required PR checks already show `ci / ci` SUCCESS.
   The authoritative selector then chooses the newest `ci / ci` check, which
   can belong to the still-running `pipeline.yml` carrier that also contains
   `release / converge`. Status attestation requires that workflow run to be
   `completed`/`success`, so it fails `untrusted_ci_recovery_identity` until
   the release job ends — a run that cannot complete until attestation
   succeeds.
2. **Production guard visibility under the release App token.** After dedicated
   recovery `33136865709` made required-context attestation succeed, the merge
   step calls `verify-production-merge-guard.sh` with the mutation App token
   minted as `permission-contents: write`, `permission-issues: write`, and
   `permission-pull-requests: write`. GitHub omits `bypass_actors` unless the
   caller has ruleset write access, while `GET /repos/{owner}/{repo}/rulesets/{id}`
   is reachable under Metadata:read, so the fetch can succeed and the validator
   still raises `production_merge_guard_missing`. The same revision is
   `ok ruleset_id=20575146` under an administrator-visible token.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. Never select an in-progress or failed release carrier as trusted recovered
   `ci / ci`. Recovery must dispatch or select the dedicated
   `promotion-pr-validation PR #1090` workflow and require that run to be
   completed/successful before attestation.
2. Preserve the mutation token at exactly Contents/Issues/Pull requests write
   and as the sole App token for `gh pr merge` and mutations. Separately mint
   an ephemeral, current-caller-repository-scoped guard token with only
   Administration write, inject it only into guard verification immediately
   before merge, and use the same two-token separation in the production-branch
   path of `merge-gate.yml`. Omitted/non-array `bypass_actors` fails distinctly.
   Activation requires `karsift-ai-infra-bot` Administration: Read and write
   plus owner approval on KARSIFT organization installation `148001476`; it
   requires no secret rotation and must produce hosted proof of explicit
   `bypass_actors: []`. That installation currently selects all repositories,
   while each guard token remains explicitly scoped to this caller repository.
3. Exercise the real token-visible payload shape and parse both workflow token
   mints in regression tests, proving exact permissions, repository scope,
   guard-before-merge order, and that the guard token never reaches merge,
   mutation, status, issue, or PR operations.
4. Do not weaken the guard, add bypass actors, fabricate statuses, or manually
   merge. Preserve exact PR base/head/refs/repository binding, ruleset
   enforcement, no founder-comment gate, and idempotent `reconcile-release`.
5. Pin the caller fixture to the new independently reviewed infrastructure
   merge. After the exact reviewed caller merge, rerun dedicated promotion
   recovery if necessary, `reconcile-release` for #1089, verify #1090 merges,
   verify `develop` synchronizes to the promotion merge SHA, and verify the
   exact `main` push production deployment succeeds. Do not snapshot the
   develop/main gap (`karsift-ai-infra#15`). Exhaustively search current source
   and pin documentation; reconcile at least the fixture README,
   `docs/operations/11-devops-and-ci-cd.md`,
   `docs/governance/repository-settings.md`, the activation checklist, and
   DOC-19 to active A-004/current automatic release and the two-token contract.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Both tokens still derive from the same App/private key and installation
`148001476` currently uses `repository_selection: all`, so record that
organization-installation permission-ceiling risk; a dedicated
single-repository guard App is optional future hardening, not required or
authorized by this package.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Unblock release convergence CI identity and production-merge-guard App-token visibility | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture,
pin, recovery-identity, attestation, and production-merge-guard updates belong
under the R4 path floor, and because the change mutates required-check recovery
and the App identity that proves production protection. The path-based
classifier and independent verifier remain authoritative; this draft proposal
is not a determination.
