# VOC-138 — Test Plan

## VOC-138-TEST-00 — Promotion signal deterministically selects pr-validation

- Covers: `VOC-138-AC-00`
- Preconditions: fixture `run-app-checks.sh` callable; temporary git repo;
  no secrets or production data
- Procedure: Construct a checkout whose capture fixture differs between base
  and head, whose recorded `subject_revision` is a 40-character SHA that
  `git cat-file` cannot resolve, and whose exact PR SHAs are valid. Invoke
  the runner with the promotion signal and those SHAs. Assert stdout contains
  `application-check provenance mode: pr-validation`. Assert `PR_BASE_SHA`
  and `PR_HEAD_SHA` are exported. Assert the runner source contains no
  `git fetch`.
- Expected result: the #1090 class no longer selects `pr-ancestry`.
- Repeat with a resolvable but non-ancestor subject and assert the same mode;
  object availability must not change promotion classification.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-01 — VOC-112-TEST-12 and VOC-112-TEST-13 pass under pr-validation without the subject

- Covers: `VOC-138-AC-01`
- Preconditions: committed `voc112-navigation-benchmark.test.mjs` unchanged
  versus `b9e74fc2…`; `pr-validation` env with valid merge-base SHAs
- Procedure: With `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` and valid
  `PR_BASE_SHA` / `PR_HEAD_SHA`, run
  `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`.
  Assert `VOC-112-TEST-12` and `VOC-112-TEST-13` pass while
  `git cat-file -e f9d11e232a07c7d7a9c433d02c9267912543ba10^{commit}` is
  allowed to fail in the constructed missing-subject fixture used by TEST-00.
  Assert the provenance test source still fail-closes `pr-ancestry` when the
  subject is missing.
- Expected result: named tests pass under the correct exact-base/head mode;
  `pr-ancestry` fail-closed text remains.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-02 — Eight VOC-112 no-change paths stay byte-identical to the anchor

- Covers: `VOC-138-AC-01`
- Preconditions: protected comparison anchor (not the implementation PR base)
  `b9e74fc2db4691c48c637639b265d527de9f4505`
- Procedure: For each of the eight VOC-112 no-change paths, assert
  working-tree bytes equal `git show b9e74fc2…:path` and that
  `git diff b9e74fc2…` does not name the path. If the commit or path cannot
  resolve, fail — do not skip. Assert JSON `subject_revision` remains
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
- Expected result: VOC-112 identity remains; the repair is mode selection,
  not a recapture.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-03 — Ordinary fixture-changing PRs still select pr-ancestry

- Covers: `VOC-138-AC-02`
- Preconditions: same runner as TEST-00; promotion signal absent
- Procedure: Reuse the existing add/modify/delete fixture transitions
  without the promotion signal. Assert each prints
  `application-check provenance mode: pr-ancestry`. Assert a fork or
  non-`main`/`develop` pair cannot pass the promotion signal.
- Expected result: ordinary capture PRs still prove ancestry.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-04 — Ordinary missing-subject fixture change still fails closed

- Covers: `VOC-138-AC-02`
- Preconditions: capture fixture changed; subject commit object absent;
  promotion signal absent; mode `pr-ancestry`
- Procedure: Assert `assertCapturedRevision` (or the runner-selected mode
  plus the existing provenance test) throws
  `/PR ancestry mode requires every captured commit object/`.
- Expected result: missing-subject fallback is not a general escape hatch.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-05 — Missing or malformed PR SHAs fail closed

- Covers: `VOC-138-AC-03`
- Preconditions: runner callable
- Procedure: Assert invalid or conflicting `--pr-base-sha`/`--pr-head-sha`
  exits non-zero with `invalid or conflicting exact PR validation context`
  (or the live equivalent). Assert unavailable or unrelated commits exit
  non-zero with `exact PR validation commits are unavailable or unrelated`.
  Assert the existing VOC-113 missing-SHA messages still apply under
  `pr-validation`.
- Expected result: SHA contract remains fail-closed.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-06 — Tampered merge-base and current hashes fail closed under pr-validation

- Covers: `VOC-138-AC-03`
- Preconditions: existing VOC-113-TEST-10 cases remain in
  `voc112-navigation-benchmark.test.mjs`
- Procedure: Run the existing tampered merge-base and changed-current-hash
  tests. Assert they still throw. Add a runner-level assertion that
  unchanged fixtures with valid SHAs still print `pr-validation`.
- Expected result: hash binding remains fail-closed.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-07 — No fetch, hydrate, wrapper, or skip in the implementation diff

- Covers: `VOC-138-AC-04`
- Preconditions: implementation PR diff versus the implementation PR base
- Procedure: Assert `run-app-checks.sh` and reusable `ci.yml` contain no
  `git fetch`. Assert the implementation diff does not add a capture-commit
  fetch helper, hydrate/materialize helper, provenance-mode wrapper, or skip
  of `VOC-112-TEST-12` / `VOC-112-TEST-13`. Assert the complete-diff
  executable scan (if still in the suite) does not exclude
  `tooling/governance/tests/`.
- Expected result: VOC-134/VOC-135 hydration classes cannot recur.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-08 — Recovery requires PR-bound pr-validation equivalence

- Covers: `VOC-138-AC-04`, `VOC-138-AC-05`
- Preconditions: recovery runner test harness; no secrets or production data
- Procedure: Feed a required `ci / ci` row modeled on failed `pull_request`
  run `33122154521` / job `98691441027`. First provide same-head dispatch
  `33122158425` with `squash-safe-push`; assert it is rejected as insufficient
  and never attested. Then provide a genuine recovery execution bound to PR
  #1090, its immutable base/head SHAs, repository, `main` <- `develop`, the
  expected workflow/path, and `pr-validation`; assert recovery does not POST
  `/actions/runs/<doomed-id>/rerun`, waits for success, and only then publishes
  the equivalent result.
- Expected result: `reconcile-release` cannot repeat the doomed rerun or mask
  PR-only SHA/hash failures with a weaker same-head check.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-09 — selected_required_run_mismatch remains for other identity failures

- Covers: `VOC-138-AC-05`
- Preconditions: existing VOC-121/VOC-122 identity checks
- Procedure: Assert wrong SHA, wrong branch, wrong workflow path, wrong PR
  number, wrong validation mode (including `squash-safe-push`), or
  non-matching conclusion still raise
  `selected_required_run_mismatch` (or the live equivalent). Assert
  unbacked status fabrication is still forbidden.
- Expected result: D07 extension is narrow to the doomed-job class.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-10 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-138-AC-06`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `b263c0c110591cc798b89277dfc35542abb1597b` if that merge still selects
  `pr-ancestry` for the TEST-00 checkout.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-11 — Current-state docs match the live contract

- Covers: `VOC-138-AC-07`
- Preconditions: caller docs in the implementation diff
- Procedure: Assert `docs/operations/11-devops-and-ci-cd.md` no longer says
  the canonical promotion PR validates with `squash-safe-push`. Assert it
  states authenticated same-repository `main` <- `develop` promotion PRs
  deterministically use `pr-validation`, independent of subject availability,
  and ordinary fixture-changing PRs retain `pr-ancestry`. Assert
  `docs/development/agent-skills.md` and the fixture README match. Assert
  `git diff` against the implementation PR base does not name files under
  VOC-135, VOC-136, or VOC-137 package directories.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-12 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-138-AC-08`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  VOC-112 navigation-benchmark tests, `validate-governance.sh`,
  `classify-change-risk.sh` (R4), and `git diff --check` pass. Assert
  `t00-evidence.md` names the implementation PR base and new infra merge,
  states that the live head is bound by the App-authored independent-review
  comment/check, and does not require a tracked file in the same commit to
  equal `HEAD`. Assert fixture `roles.yml` is unchanged.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible.
- Evidence: `VOC-138-EV-00`

## VOC-138-TEST-13 — No snapshot-gap; reconcile-release handoff

- Covers: `VOC-138-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert
  evidence states that after this correction merges, `reconcile-release` for
  #1089 may merge #1090 (or the live promotion at the then-current `develop`
  head), converges `develop` to the exact main merge SHA, and does not keep
  staging scheduled for tree-equivalent sync. Assert incident runs
  `33122154521`, `33122158425`, `33122099253`, `33122436137` and jobs
  `98691441027`, `98692552949` are preserved as audit evidence, not
  restaged as this package's implementation work. Assert closed state is not
  treated as completion proof.
- Expected result: repair is provenance/recovery plus ordinary release
  handoff, not a snapshot-then-check package.
- Evidence: `VOC-138-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
