# VOC-127 — Converge develop to the exact promotion merge SHA and govern main-only reconciliation

| Field | Value |
|-------|-------|
| Package | `VOC-127` |
| Title | Converge develop to the exact promotion merge SHA and govern main-only reconciliation |
| Path | `specs/changes/VOC-127-converge-develop-to-the-exact-promotion-merge-sha` |
| Status | `draft` |
| Risk | `R4` (draft proposal; protected release-convergence and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1035](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1035) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Successful merge-commit promotions leave the long-lived integration branch at
the pre-promotion checked head while production advances to a main-only merge
commit. Authorized main-only live-evidence carriers have no governed path back
to integration. That produces permanent commit drift and can produce real tree
drift.

This is a KARSIFT automation reliability fix, not product behavior. The
2026-08-27 bounded settings operation already recreated `develop` at exact
`main`; this package must not snapshot that gap as a later drift gate
(`karsift-ai-infra#15`). It must change `release.yml` so the next `--merge`
promotion cannot recreate the gap, and add an exceptional adopted exact-SHA
path for main-only hotfix/evidence reconciliation.

### Live reproduction (before the 2026-08-27 repair)

| Item | Value |
|------|-------|
| `develop` | `883c4a544f24fb9840694a834ebe9c665d4160b5` |
| `main` | promotion merge SHA `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` from PR [#1033](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1033) |
| `git rev-list --left-right --count origin/develop...origin/main` | `0 46` (develop had no unique commits; 46 behind) |
| `git merge-base --is-ancestor origin/develop origin/main` | succeeded |
| Tree diff | exactly the two VOC-113-T01 evidence files from main-targeting PR [#955](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/955) |
| Active PR/issue/workflow | none |
| Post-repair refs | both `develop` and `main` resolve to `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`; compare is zero ahead/behind; trees identical |
| Audit | release issue [#1032](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1032) |

Root cause: `KARSIFT/karsift-ai-infra/.github/workflows/release.yml` performs
the checked `gh pr merge --merge --match-head-commit "$CHECKED_HEAD_SHA"`, then
only preserves or recreates the integration ref at `CHECKED_HEAD_SHA`. It does
not resolve `mergeCommit.oid`, bind it to the checked head/base, or advance
`develop` before closing the release audit. Every `--merge` promotion therefore
creates a main-only merge commit by design. The `ahead_by == 0` "Already
promoted" path then closes a later `reconcile-release` without syncing, because
a merge-commit promotion leaves develop as an ancestor of main.

## Required outcome (summary)

Use one largest-safe coherent task covering infrastructure, caller dispatch,
tests, documentation, fixture pin, and evidence:

1. Preserve the real promotion PR and merge commit as audit evidence.
2. After every successful governed `develop → main` promotion, resolve the
   exact merge SHA and advance `develop` to that SHA before release closure.
3. Bind synchronization to the exact checked promotion head, exact production
   base, exact merged PR identity, and live branch tips.
4. Fail closed if `main` or `develop` moved, if the merge result is malformed,
   or if develop gained unique commits; never erase concurrent integration work.
5. Make interrupted synchronization idempotently recoverable through
   `reconcile-release`, including a promotion already merged but an audit not
   yet closed.
6. Do not create another promotion PR or a release loop.
7. A tree-equivalent synchronization must not schedule an unnecessary staging
   deployment; retain path-selected deployment for actual runtime/deploy tree
   changes.
8. Add an explicit governed, adopted, exact-SHA path for exceptional main-only
   hotfix/evidence changes to reconcile back to develop. It must not normalize
   direct main changes as the ordinary workflow.
9. Preserve all existing roster completion, required-check recovery, independent
   review, retry, risk, secret-redaction, credential-isolation, deployment, and
   rollback controls.
10. Update all documentation that describes release scope or branch behavior.
11. Pin the caller fixture and `PINNED_SHA.txt` to the exact merged
    `KARSIFT/karsift-ai-infra` commit implementing this outcome.
12. Add deterministic regression coverage for the classes listed in issue #1035.

Current `config/roles.yml` bindings are fixed and must not change. No OpenAI
credential or execution route is needed or authorized. Do not print credential
values.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Exact-merge-SHA develop sync after promotion, exceptional main-only reconciliation, tests, docs, and caller pin | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates the
protected release-convergence contract (`.github/workflows/*` is an R3 floor;
semantic mutation of integration-ref identity after promotion is protected).
The path-based classifier and independent verifier remain authoritative; this
draft proposal is not a determination.
