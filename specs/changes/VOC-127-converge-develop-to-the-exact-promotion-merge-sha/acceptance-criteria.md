# VOC-127 — Acceptance Criteria

## VOC-127-AC-00 — Successful promotion advances develop to the exact merge SHA before audit close

- Requirement source: `VOC-127-D01`, `VOC-127-D02`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-00`, `VOC-127-TEST-01`
- Evidence: `VOC-127-EV-00`
- Result: pending

After the single exact-head `gh pr merge --merge --match-head-commit` of a
governed `develop → main` promotion PR, the serialized converge job resolves
that PR's `mergeCommit.oid`, binds it to the checked head, checked production
base, merged PR identity, and live tips, and advances `develop` to that exact
SHA. The release audit is not closed as success while `develop` still resolves
to `CHECKED_HEAD_SHA` or any other SHA. The promotion PR and merge commit
remain as audit evidence.

## VOC-127-AC-01 — Binding fails closed on races, malformed merges, and unique develop commits

- Requirement source: `VOC-127-D02`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-02`, `VOC-127-TEST-03`, `VOC-127-TEST-04`
- Evidence: `VOC-127-EV-00`
- Result: pending

Synchronization refuses and leaves refs unchanged when: live `main` is not the
bound merge SHA; live `develop` is neither the checked head, the merge SHA, nor
missing; the merge commit is missing or not a two-parent merge of the checked
base and checked head; the PR is closed without merge; or `develop` has unique
commits (is not an ancestor of the merge SHA). Concurrent integration work is
never erased.

## VOC-127-AC-02 — Interrupted sync is recoverable through reconcile-release without a second promotion

- Requirement source: `VOC-127-D03`, `VOC-127-D04`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-05`, `VOC-127-TEST-06`
- Evidence: `VOC-127-EV-00`
- Result: pending

`reconcile-release` against an open release audit whose promotion PR is already
merged binds and advances `develop` to that merge SHA, then closes the audit.
It does not open a new promotion PR, does not call `gh pr merge` again, and
does not treat `ahead_by == 0` with unequal SHAs as fully promoted. If `develop`
is already at the merge SHA, the retry is a mutation-free success. A later wake
after equal tips does not create a release loop.

## VOC-127-AC-03 — Auto-deleted develop is restored at the merge SHA

- Requirement source: `VOC-127-D01`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-07`
- Evidence: `VOC-127-EV-00`
- Result: pending

If GitHub deleted the long-lived integration branch as part of the promotion
merge, the job recreates `refs/heads/develop` at the bound merge SHA, not at
`CHECKED_HEAD_SHA`.

## VOC-127-AC-04 — Tree-equivalent sync does not schedule unnecessary staging; path selection remains

- Requirement source: `VOC-127-D05`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-08`
- Evidence: `VOC-127-EV-00`
- Result: pending

A `develop` ref update whose tree is identical to the previous `develop` tip,
or whose changed paths are outside the VOC-111 runtime/deploy allowlist, does
not perform a staging deployment. Pushes that change allowlisted
runtime/deploy paths still select `deploy-staging.yml`. The allowlist is not
broadened.

## VOC-127-AC-05 — Exceptional main-only reconciliation is adopted, exact-SHA, and not the ordinary workflow

- Requirement source: `VOC-127-D06`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-09`
- Evidence: `VOC-127-EV-00`
- Result: pending

A distinct mutating caller action reconciles an already-merged main-targeting
PR back onto `develop` only when the named package/task is adopted and
implementation-authorized, the PR number identifies that merged PR, live `main`
equals that merge commit, and `develop` is a strict ancestor with no unique
commits. Caller `workflow_dispatch` does not accept operator-typed SHAs.
Unadopted packages, develop-targeting PRs, SHA mismatch, and unique develop
commits fail closed. This action does not open a promotion PR and does not run
on a schedule against arbitrary `main`-only deltas. Ordinary post-promotion
sync remains the `release.yml` converge path.

## VOC-127-AC-06 — Existing controls, docs, pin, and roles remain

- Requirement source: `VOC-127-D07`, `VOC-127-D09`
- Tasks: `VOC-127-T00`
- Tests: `VOC-127-TEST-10`, `VOC-127-TEST-11`
- Evidence: `VOC-127-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, and rollback controls remain.
`config/roles.yml` is unchanged. No OpenAI route is added. Current-state docs
that describe release scope or branch behavior state that `develop` advances
to the promotion merge SHA before audit close, that `reconcile-release`
retries that sync, and that main-only reconciliation is exceptional.
Historical A-003 / VOC-075 records are not rewritten. The caller fixture pin
equals the exact reviewed infra merge when consumed.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
