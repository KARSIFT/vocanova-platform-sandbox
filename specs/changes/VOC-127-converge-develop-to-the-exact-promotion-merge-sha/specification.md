# VOC-127 — Converge develop to the exact promotion merge SHA and govern main-only reconciliation: Specification

## Objective and requirement source

After every successful governed `develop → main` promotion, keep the real
promotion PR and merge commit as audit evidence, resolve that exact merge SHA,
and advance `develop` to it before the release audit closes — bound to the
checked promotion head, checked production base, merged PR identity, and live
tips, fail-closed on races or unique develop commits, and idempotently
recoverable through `reconcile-release`. Add a separate exceptional adopted
exact-SHA path that can reconcile authorized main-only hotfix/evidence changes
back to `develop` without making direct-to-main the ordinary workflow.

**Requirement source:** [GitHub issue #1035](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1035).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1035)

| Item | Value |
|------|-------|
| Pre-repair `develop` | `883c4a544f24fb9840694a834ebe9c665d4160b5` |
| Pre-repair `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` (PR #1033 merge) |
| Ahead/behind | `0 46` |
| Ancestor | `develop` was an ancestor of `main` |
| Tree drift | VOC-113-T01 evidence files from main-targeting PR #955 |
| Root cause | `release.yml` restores `CHECKED_HEAD_SHA`, not `mergeCommit.oid` |
| Post-repair | both refs at `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` (settings operation on #1032; not reusable implementation authority) |

## Scope and non-goals

### In scope

1. Change the authoritative `KARSIFT/karsift-ai-infra` `release.yml` serialized
   converge job so that, after the single exact-head `--merge` (or when that
   promotion PR is already merged), it resolves the merge commit SHA, binds it,
   and advances `develop` to that exact SHA before closing the release audit.
2. Preserve the real promotion PR and merge commit. Do not squash, rebase, or
   replace that merge with a fast-forward of `main`, and do not delete the PR
   as the audit record.
3. Bind synchronization to all of: checked promotion head (`CHECKED_HEAD_SHA`),
   checked production base at merge time, merged PR number/identity, live
   `main` tip, and live `develop` tip (or a proven missing integration ref
   after auto-delete).
4. Fail closed if `main` moved away from the expected merge SHA, if `develop`
   moved to an unexpected SHA, if the merge commit is missing or malformed
   (not a two-parent merge whose parents are the checked base and checked
   head), or if `develop` gained unique commits. Never rewind or erase
   concurrent integration work.
5. Make the same bind-and-sync path idempotent via `reconcile-release`,
   including the case where the promotion PR is already merged and the audit
   issue is still open. Do not treat `ahead_by == 0` alone as "already
   promoted" while `develop` and `main` resolve to different SHAs.
6. Do not open another promotion PR once the bound merge exists. A later wake
   that finds `develop` already at that merge SHA, or `develop == main`, must
   close or no-op the audit without creating a release loop.
7. Ensure a tree-equivalent `develop` update (identical tree versus the
   previous `develop` tip) does not schedule an unnecessary staging deployment.
   Retain VOC-111 path-selected deployment for actual runtime/deploy tree
   changes. If GitHub's `on.push.paths` filter is not sufficient for
   merge-commit fast-forwards, add a fail-closed job-level selector in
   `deploy-staging.yml` that skips when no allowlisted path changed; do not
   broaden the allowlist.
8. Add an explicit exceptional caller action (preferred name
   `reconcile-main-to-develop`; exact name is `VOC-127-DEP-07`) that, for an
   adopted package/task and a merged main-targeting PR number, derives the
   merge SHA, applies the same fail-closed bind-and-sync to `develop`, and
   does not open a promotion PR. Operator identity is that PR number plus
   adopted package bindings — not a free-form SHA on `workflow_dispatch`.
   This path must not run on a schedule against arbitrary `main`-only deltas
   and must not be the ordinary post-promotion route.
9. Preserve existing roster-completion markers, required-check recovery,
   independent review, retry caps, risk floors, secret redaction, App-token
   isolation, deployment, and rollback controls from VOC-113 through VOC-126.
10. Update every current-state document that would otherwise still describe
    `release.yml` as leaving `develop` at the checked head, treating
    `ahead_by == 0` as fully converged, or omitting post-promotion
    develop-sync / exceptional main-only reconciliation. Do not rewrite
    historical A-003 / VOC-075 / CHANGELOG audit records except for a new
    current-state note where that file's current-state section is the live
    contract.
11. After the independently reviewed infrastructure merge, pin
    `tooling/governance/fixtures/karsift-ai-infra/` and `PINNED_SHA.txt` to
    that exact merge SHA when the fixture consumes the change, and advance
    matching caller pin assertions.
12. Add deterministic tests for the classes in `VOC-127-D08`.

### Non-goals / explicitly excluded

- Application runtime behavior, deployment topology, product permissions, or
  monitor inventory changes, except the staging skip for tree-equivalent
  develop-sync described above.
- Re-performing the 2026-08-27 settings repair, snapshotting the current
  (already repaired) develop/main gap as package evidence, or gating a later
  task on drift against such a snapshot.
- Fast-forwarding `main` instead of preserving `--merge` promotion commits.
- Normalizing direct pushes or unreviewed PRs to `main` as the ordinary
  workflow.
- Operator-typed SHA inputs on any caller `workflow_dispatch`.
- Overloading `existing_pr_number` (that input remains implement-only).
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease / compare-and-swap, retry caps, or fail-closed
  missing-binding behavior.
- Implementing VOC-122 promotion-recovery replan, merging unrelated open
  carriers, or rotating `KARSIFT_BOT_*` secrets / App installation permissions.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG, A-003, or VOC-075 audit records.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. The #1032 repair and #1033/#955 identities are
  planning evidence, not T00 live-evidence gates.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: release-convergence and integration-ref mutation in
  `KARSIFT/karsift-ai-infra` `release.yml` and helpers; caller `pipeline.yml`
  exceptional reconcile dispatch; `deploy-staging.yml` path selection;
  caller `tooling/governance/` fixtures and tests; current-state release and
  branch-behavior documentation.
- Protected technical effect: whether `develop` remains a permanent ancestor
  of `main` after every `--merge` promotion; whether concurrent unique
  develop commits can be erased; whether a second promotion PR / release loop
  can fire; whether tree-equivalent ref sync deploys staging. No application
  runtime effect is intended for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-127-D00`: This is one outcome-sized release-convergence change. Use one
end-to-end implementation task covering infrastructure `release.yml` and
helpers, source tests, exceptional caller dispatch, staging path-selection
proof or narrow skip, current-state docs, caller fixture/pin, and evidence.
Coordinated pull requests in `KARSIFT/karsift-ai-infra` and this caller remain
one task. Repository count, file count, and workflow-versus-tests-versus-docs
are not split reasons. Do not add a snapshot-then-drift-check task for the
already-repaired #1033 gap.

`VOC-127-D01`: Ordinary post-promotion synchronization is mandatory in the
existing serialized `converge` job. After `gh pr merge --merge --match-head-commit
"$CHECKED_HEAD_SHA"` succeeds, or when the same promotion PR is already
`MERGED`, resolve `mergeCommit.oid`, bind it, fast-forward or recreate
`develop` at that exact SHA using compare-and-swap / non-force ref update
against the expected current SHA, then close the release audit. Preserve the
promotion PR and merge commit. Recreating an auto-deleted integration ref
must use the merge SHA, not `CHECKED_HEAD_SHA`. If `develop` is already at
that SHA, skip mutation and close (idempotent).

`VOC-127-D02`: Binding must prove all of the following before mutating
`develop`:

1. merged PR number is the evaluated promotion PR;
2. PR base is `production_branch` and head is `integration_branch`;
3. merge commit OID is a 40-hex SHA;
4. that commit has exactly two parents: the checked production base and
   `CHECKED_HEAD_SHA` (GitHub merge first parent is the base);
5. live `main` tip equals that merge SHA;
6. live `develop` is exactly `CHECKED_HEAD_SHA`, exactly the merge SHA, or
   missing (auto-deleted). Any other `develop` SHA is a race.

If `develop` is not an ancestor of the merge SHA (unique develop commits),
fail closed and do not move the ref. If `main` is not the merge SHA, fail
closed. If the PR is `CLOSED` without merge, fail closed. Do not close the
audit as success unless the live `develop` tip equals the bound merge SHA.

`VOC-127-D03`: `reconcile-release` retries the same serialized converge job.
When the promotion PR is already merged and the audit is still open, do not
take the current "already terminal; successful no-op" or `ahead_by == 0`
short-circuit that skips sync. Instead bind-and-sync, then close. "Already
promoted" means live `develop` SHA equals live `main` SHA (fully converged),
not merely `ahead_by == 0`. Diverged history (`ahead_by > 0` and
`behind_by > 0`) fails closed and never erases unique develop commits.

`VOC-127-D04`: After a bound merge exists, later wakes must not open a new
promotion PR or merge again. One `gh pr merge` call remains the sole merge
mutation. A develop-ref update that wakes `check_run` / `workflow_run` /
issue events must evaluate, see equal tips or a completed bound sync, and
no-op or close the audit. Do not create a release loop.

`VOC-127-D05`: Staging. Retain the VOC-111 fail-closed runtime/deploy
allowlist on `deploy-staging.yml`. A tree-equivalent develop synchronization
(empty `git diff --name-only` versus the previous `develop` tip, or only
non-allowlisted paths such as `specs/**` evidence) must not schedule or must
exit successfully without deploying. Actual allowlisted runtime/deploy tree
changes must still select a staging deploy. If `on.push.paths` cannot be
shown to skip merge-commit fast-forwards, add a job-level selector; do not
broaden `"*"`, `apps/**`, `packages/**`, `infra/**`, or the other allowlisted
globs.

`VOC-127-D06`: Exceptional main-only reconciliation is a distinct mutating
caller action on `pipeline.yml` (not `pipeline-verify.yml`). It requires an
adopted package, implementation-authorized task identity, and a merged PR
whose base is `production_branch` and whose merge commit is the live `main`
tip. It derives the SHA from that PR; caller `workflow_dispatch` still must
not expose operator-typed SHA inputs. It applies `VOC-127-D02` ancestor and
race rules, advances `develop` to that merge SHA when `develop` is a strict
ancestor with no unique commits, and does not open a promotion PR. It must
not run automatically on every `main` push. Direct-to-main remains
exceptional; this path is how an already-authorized main-only change (for
example operator-owned evidence like VOC-113-T01 PR #955) is reconciled into
`develop`. Unadopted, unmerged, develop-targeting, or SHA-mismatched PRs fail
closed. Keep every live `workflow_dispatch` block at most 25 inputs.

`VOC-127-D07`: Preserve VOC-113 through VOC-126 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer implementer,
Cursor Grok exact-revision review, no OpenAI route, `roles.yml` unchanged, no
secrets in bundles/logs/fixtures, no credential values printed. Source PRs
`Relates to OWNER/CALLER#N` with no closing keyword. This package's caller PR
`Closes` only its own VOC-127 task issue.

`VOC-127-D08`: Deterministic tests must prove:

1. happy-path: after `--merge`, `develop` is advanced to `mergeCommit.oid`
   before audit close, and the promotion PR/merge commit remain;
2. already-merged reconcile: open audit + merged PR + develop still at
   checked head → sync to merge SHA, no new PR;
3. idempotent retry: develop already at merge SHA → no ref mutation, close;
4. unique develop commits → fail closed, ref unchanged;
5. `main` moved away from the merge SHA → fail closed;
6. malformed/missing merge commit, missing `main`, or unexpected `develop`
   SHA → fail closed;
7. auto-deleted `develop` is recreated at the merge SHA, not `CHECKED_HEAD_SHA`;
8. `ahead_by == 0` with unequal SHAs is not treated as fully promoted;
9. no second `gh pr merge` and no new promotion PR after a bound merge;
10. exceptional main-only path requires adopted package + merged
    main-targeting PR number, rejects operator SHAs, rejects unadopted or
    develop-targeting PRs, and still never erases unique develop commits;
11. tree-equivalent develop sync does not keep a staging deploy scheduled;
    allowlisted runtime/deploy path changes still select deploy;
12. existing recovery, review, retry, and credential-isolation contracts
    remain.

Tests must not mint real App tokens, use secrets, or use production data.

`VOC-127-D09`: Current-state comments in `release.yml`, the pipeline template,
`karsift-ai-infra/README.md`, `AGENTS.md` (reconcile-release and release
authority), `docs/operations/11-devops-and-ci-cd.md`,
`docs/operations/10-development-workflow.md`, current-state branch/release
paragraphs in `docs/operations/15-ai-native-product-and-engineering-operating-model.md`
and `docs/governance/16-autonomous-development-operating-model.md`, and any
other live contract that would become false, must say that after a successful
promotion merge, `develop` is advanced to that exact merge SHA before audit
close, that `reconcile-release` retries that sync, and that exceptional
main-only reconciliation is a separate adopted exact-SHA path. Historical
amendment records stay unchanged. After the exact reviewed infrastructure
merge SHA is known, pin the caller fixture when consumed.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Synchronization uses the
existing release App token with contents write; it must not broaden
installation permissions, must not grant the model-controlled runner that
token, and must not accept free-form SHAs as authority.

Abuse/process risks:

1. Leaving `develop` at `CHECKED_HEAD_SHA` after `--merge` — mitigated by
   `VOC-127-D01` / `VOC-127-D03`.
2. Erasing unique develop commits by forcing the merge SHA —
   forbidden by `VOC-127-D02`.
3. Opening a second promotion PR or looping release after the develop-ref
   update — forbidden by `VOC-127-D04`.
4. Treating arbitrary main-only commits as ordinary workflow —
   forbidden by `VOC-127-D06`.
5. Unnecessary staging deploys from tree-equivalent ref sync —
   mitigated by `VOC-127-D05`.
6. Printing App tokens, private keys, or secret values — forbidden.

## Contradictions and open questions

1. **Helper and action names (`VOC-127-DEP-07`):** required behavior is
   settled; exact Python module name and exceptional `pipeline.yml` action /
   input names are implementation choices. Preferred: `reconcile-main-to-develop`
   plus a dedicated merged-PR input that is not `existing_pr_number`.
2. **Staging skip mechanism:** required outcome is settled (tree-equivalent
   sync must not deploy; allowlisted runtime changes must). Whether VOC-111
   `on.push.paths` already skips merge-commit fast-forwards, or a job-level
   selector must be added, is proven at implementation time — not guessed
   here. Do not broaden the allowlist.
3. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes `release.yml`,
   helpers, tests, or comments. Consumption is expected. Pin to this
   package's infra merge, not to the drafting-time pin
   `60afda3a44fd06b8c00b219771de7112f1aded6e`. If some files are not in that
   subset, do not copy them merely to force a pin; record non-consumption.
4. **AGENTS.md / DOC-15 / DOC-16:** update current-state sentences that would
   become false. Do not expand them into a new release-internals runbook, and
   do not rewrite historical correction notes.
5. **Pre-existing 2026-08-27 repair:** both refs currently match. This package
   must not treat that settings operation as a template for implementation,
   must not open a promotion of "current develop to main" as its work, and
   must not add a snapshot commit under this package directory as a drift
   baseline.
