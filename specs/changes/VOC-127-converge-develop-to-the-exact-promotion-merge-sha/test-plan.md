# VOC-127 — Test Plan

## VOC-127-TEST-00 — Happy-path merge advances develop to mergeCommit.oid before audit close

- Covers: `VOC-127-AC-00`
- Preconditions: fixture of an open develop→main promotion PR whose newest
  exact-head checks pass; no secrets or production data
- Procedure: Drive the serialized converge merge-and-sync path (helper unit
  tests and/or `release.yml` policy tests). After a successful
  `--merge --match-head-commit` against `CHECKED_HEAD_SHA`, assert the job
  reads `mergeCommit.oid`, binds parents to checked base + checked head, and
  PATCHes or creates `refs/heads/develop` at that SHA before
  `gh issue close`. Assert the promotion PR remains and is not replaced.
  Assert `test_release_policy.py` no longer requires restoration at
  `CHECKED_HEAD_SHA`.
- Expected result: `develop` equals the promotion merge SHA before audit close;
  the #1033 class (develop left at checked head) fails the new assertion.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-01 — Binding requires checked head, checked base, merged PR, and live tips

- Covers: `VOC-127-AC-00`
- Preconditions: same helper/fixture surface as TEST-00
- Procedure: Assert sync is refused unless merge commit OID is 40-hex, has
  exactly two parents `{checked_base, CHECKED_HEAD_SHA}`, live `main` equals
  that OID, and the PR number/base/head match the evaluated promotion PR.
- Expected result: synchronization cannot target an unbound SHA.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-02 — Unique develop commits fail closed and are not erased

- Covers: `VOC-127-AC-01`
- Preconditions: fixture where `develop` has at least one commit not in the
  merge SHA
- Procedure: Attempt bind-and-sync. Assert the helper/job exits non-zero, does
  not PATCH `develop`, and does not close the audit as success.
- Expected result: concurrent integration work survives; the #1035 "never
  erase" rule holds.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-03 — Changed main or unexpected develop SHA fails closed

- Covers: `VOC-127-AC-01`
- Preconditions: fixtures where live `main` ≠ merge SHA, or live `develop` is
  neither checked head, merge SHA, nor missing
- Procedure: Assert each case fails closed with no ref mutation.
- Expected result: branch races cannot silently converge to the wrong tip.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-04 — Malformed, missing, or unmerged results fail closed

- Covers: `VOC-127-AC-01`
- Preconditions: fixtures for missing `mergeCommit`, non-two-parent commit,
  parent mismatch, missing `main` ref, and PR `CLOSED` without merge
- Procedure: Assert each case fails closed. Distinct from already-`MERGED`
  recovery (TEST-05).
- Expected result: missing/deleted/malformed refs cannot be treated as
  successful promotion.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-05 — Already-merged promotion with open audit syncs idempotently via reconcile-release

- Covers: `VOC-127-AC-02`
- Preconditions: promotion PR `MERGED` at known merge SHA; `develop` still at
  `CHECKED_HEAD_SHA`; release audit still open; `reconcile-release` identity
  present
- Procedure: Assert converge does not no-op as "already terminal", does not
  close as "Already promoted" merely because `ahead_by == 0`, does not open a
  new PR, does not call `gh pr merge` again, advances `develop` to the merge
  SHA, then closes the audit. Second run with `develop` already at that SHA
  performs no ref mutation and still succeeds.
- Expected result: interrupted sync is recoverable; #1035 already-merged
  reconcile class is covered.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-06 — Bound merge does not create a second promotion PR or release loop

- Covers: `VOC-127-AC-02`
- Preconditions: after successful sync, `develop` SHA equals `main` SHA;
  roster still complete; a later check_run/workflow_run/issue wake is
  simulated in policy tests
- Procedure: Assert `release.yml` still contains exactly one `gh pr merge`.
  Assert a wake with equal tips does not `gh pr create` a new promotion PR.
  Assert `ahead_by == 0` with unequal SHAs is not the fully-promoted
  short-circuit.
- Expected result: develop-ref update cannot restart promotion.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-07 — Auto-deleted develop is recreated at the merge SHA

- Covers: `VOC-127-AC-03`
- Preconditions: integration ref missing after merge; merge SHA bound
- Procedure: Assert `POST git/refs` uses `sha=<mergeCommit.oid>`, not
  `CHECKED_HEAD_SHA`. Update `test_release_policy.py`
  `test_promotion_preserves_long_lived_integration_branch` accordingly.
- Expected result: restoring a deleted `develop` cannot recreate the #1033
  ancestor gap.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-08 — Tree-equivalent sync does not schedule staging; allowlisted paths still do

- Covers: `VOC-127-AC-04`
- Preconditions: live `.github/workflows/deploy-staging.yml` and
  `scripts/foundation/voc111-deploy-staging-paths.test.mjs`
- Procedure: Assert an empty changed-path set (tree-equivalent develop update)
  does not select a staging deploy. Assert `specs/**` evidence-only paths do
  not select deploy. Assert existing VOC-111 allowlisted paths (`apps/**`,
  `packages/**`, `infra/**`, root files matched by `"*"`, and the other
  listed globs) still select deploy. If implementation adds a job-level skip,
  assert it exits successfully without SSH/deploy steps when no allowlisted
  path changed, and does not skip when an allowlisted path did change. Do not
  treat a broadened allowlist as a pass.
- Expected result: merge-SHA develop sync does not waste a staging deploy;
  real runtime/deploy tree changes still deploy.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-09 — Exceptional main-only path is authorized, exact-SHA, and not ordinary

- Covers: `VOC-127-AC-05`
- Preconditions: live and template `pipeline.yml`; helper for the exceptional
  action (preferred `reconcile-main-to-develop`)
- Procedure: Assert the mutating action exists on `pipeline.yml`, not
  `pipeline-verify.yml`. Assert it requires adopted package/task identity and
  a merged main-targeting PR number, forwards those bindings, and does not
  declare operator-typed SHA inputs. Assert it fails closed for unadopted
  packages, develop-targeting PRs, live `main` ≠ merge commit, and unique
  develop commits. Assert it does not open a promotion PR and is not a
  `schedule` trigger. Assert `existing_pr_number` remains implement-only.
  Assert every live and template `workflow_dispatch` still has at most 25
  inputs (reuse VOC-126 TEST-00 style).
- Expected result: main-only reconciliation cannot be used as the ordinary
  workflow and cannot paste a SHA as authority.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-10 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-127-AC-06`
- Preconditions: final `release.yml`, `implement.yml`, caller pipeline
- Procedure: Reuse or extend existing policy tests to assert roster-marker
  validation, exact-head `--match-head-commit`, ruleset attestation before
  merge, required-check recovery, App-token on merge/sync mutation and
  job-token on recovery reads, two-attempt implementer bound, Cursor Composer
  implementer, Cursor Grok reviewer, no OpenAI path, and unchanged
  `config/roles.yml`. Confirm this package's caller PR `Closes` only its own
  VOC-127 task issue and source PRs `Relates to` without a closing keyword.
- Expected result: merge-SHA sync lands without weakening prior safety gates.
- Evidence: `VOC-127-EV-00`

## VOC-127-TEST-11 — Source self-CI, caller suites, docs, and pin

- Covers: `VOC-127-AC-06`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. If the fixture consumed the infra change,
  assert the pin equals the exact reviewed infra merge and is not
  `60afda3a44fd06b8c00b219771de7112f1aded6e`, and that matching foundation
  pin literals match. If not, assert the pin is unchanged and evidence
  records why. Assert current-state README/`release.yml` comments/`AGENTS.md`
  and the listed operations/governance docs describe exact-merge-SHA develop
  sync, `reconcile-release` retry, and exceptional main-only reconciliation,
  and that historical A-003/VOC-075 records were not rewritten. Confirm
  `t00-evidence.md` records that no snapshot-gap task was added and no
  bootstrap exception was used.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable.
- Evidence: `VOC-127-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
