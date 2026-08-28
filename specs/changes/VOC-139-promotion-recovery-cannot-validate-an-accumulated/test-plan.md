# VOC-139 — Test Plan

## VOC-139-TEST-00 — Accumulated promotion with differing AGENTS.md passes under pr-validation

- Covers: `VOC-139-AC-00`
- Preconditions: provenance test callable; a git history where `AGENTS.md`
  (and/or the navigator skill) at `PR_BASE_SHA` differs from `PR_HEAD_SHA`;
  stored capture hashes equal the head/working-tree files; promotion signal
  set; no secrets or production data
- Procedure: Construct or reuse a checkout modeled on #1090 (base
  `0d0b0cdf…` / head a develop revision whose `AGENTS.md` hash prefix is
  `b0e65629…` and whose merge-base hash prefix is `5ba216ff…`). Export
  `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation`, valid `PR_BASE_SHA` /
  `PR_HEAD_SHA`, and the promotion signal. Run
  `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`.
  Assert `VOC-112-TEST-12` and `VOC-112-TEST-13` pass. Assert the runner
  still prints `application-check provenance mode: pr-validation` and still
  exports the exact SHAs.
- Expected result: accumulated promotions no longer fail
  `AGENTS.md hash must be anchored in the PR merge base`.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-01 — Promotion still passes when the capture subject is missing or a non-ancestor

- Covers: `VOC-139-AC-00`
- Preconditions: same promotion signal and exact SHAs as TEST-00; recorded
  `subject_revision` either absent from the object store or resolvable but
  not an ancestor of HEAD
- Procedure: Repeat TEST-00 in both subject states. Assert the named VOC-112
  tests still pass and mode remains `pr-validation`. Assert the runner source
  contains no `git fetch`.
- Expected result: VOC-138 missing-subject / non-ancestor retention holds
  after the hash-rule change.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-02 — Ordinary pr-validation still requires merge-base hashes

- Covers: `VOC-139-AC-01`
- Preconditions: promotion signal absent; `pr-validation`; `AGENTS.md` at
  HEAD differs from the merge-base
- Procedure: Invoke `assertCapturedRevision` (or the provenance tests) with
  valid related SHAs and no promotion signal. Assert failure matching
  `/AGENTS.md hash must be anchored in the PR merge base/` (or the navigator
  equivalent). Repeat with hashes that do match the merge-base and assert
  pass.
- Expected result: "hashes differ from main" is not a general escape hatch.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-03 — Ordinary fixture-changing PRs still select pr-ancestry

- Covers: `VOC-139-AC-01`
- Preconditions: same runner as VOC-138; promotion signal absent
- Procedure: Reuse add/modify/delete fixture transitions without the
  promotion signal. Assert each prints
  `application-check provenance mode: pr-ancestry`. Assert a fork or
  non-`main`/`develop` pair cannot pass the promotion signal. Assert an
  ordinary missing-subject fixture change still fails closed with
  `PR ancestry mode requires every captured commit object`.
- Expected result: ordinary capture PRs still prove ancestry.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-04 — Missing or malformed PR SHAs fail closed

- Covers: `VOC-139-AC-02`
- Preconditions: runner and provenance test callable
- Procedure: Assert invalid or conflicting `--pr-base-sha`/`--pr-head-sha`
  exits non-zero with `invalid or conflicting exact PR validation context`
  (or the live equivalent). Assert unavailable commits or a base that is not
  an ancestor of the head exit
  non-zero with `exact PR validation commits are unavailable or unrelated`.
  Assert the existing VOC-113 missing-SHA messages still apply under
  `pr-validation` with and without the promotion signal.
- Expected result: SHA contract remains fail-closed.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-05 — Tampered current hashes fail closed under both contracts

- Covers: `VOC-139-AC-02`
- Preconditions: existing VOC-113-TEST-10 changed-current-hash case remains
- Procedure: Run the existing tampered current-hash test. Add a promotion-
  signal case whose stored hashes match neither the working tree nor
  `PR_HEAD_SHA` and assert it throws. Add a promotion-signal case whose
  stored hashes match the merge-base but not `PR_HEAD_SHA` and assert it
  throws.
- Expected result: head-binding does not accept main's hashes as a
  substitute for the reviewed head.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-06 — Unrelated repository, PR, and wrong refs fail closed

- Covers: `VOC-139-AC-02`, `VOC-139-AC-03`
- Preconditions: metadata command body and recovery attestation harness
- Procedure: Assert `promotion-pr-metadata` rejects a payload whose base/head
  refs are not `main`/`develop`, whose base/head repository full names do not
  both equal `$GITHUB_REPOSITORY`, whose PR is not open, or whose SHAs are not
  40-character commit IDs. Use a response shaped like live `gh pr view`, where
  `headRepository` has `id` and `name` but no `nameWithOwner`, and assert the
  implementation does not read that absent field. Include a same-name fork
  with a different owner. Assert
  recovery attestation still rejects wrong PR number, wrong repository,
  wrong workflow path, and wrong SHAs (`selected_required_run_mismatch` or
  the live equivalent).
- Expected result: identity mismatches stay fail-closed after `-R` is added.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-07 — No-checkout metadata depends on an explicit repo flag

- Covers: `VOC-139-AC-03`
- Preconditions: live caller `pipeline.yml`, infra template, and fixture
  mirror; a mock `gh` on `PATH`; a temporary directory that is not a git
  repository; no secrets
- Procedure: Extract and execute the real `promotion-pr-metadata` `run`
  body (or the helper it actually calls). Set `GITHUB_REPOSITORY`,
  `PROMOTION_PR_NUMBER`, and `GH_TOKEN` to fixture values. Run with `cwd`
  equal to the non-git directory. Assert the invoked `gh` arguments address
  `$GITHUB_REPOSITORY` explicitly, either through the REST path or `-R` /
  `--repo`, and return a realistic identity shape. Assert the job YAML has no
  `actions/checkout` step. Repeat without explicit repository context (or
  with `gh` configured to require git context when it is absent) and assert
  failure matching
  `/not a git repository/` or a deterministic equivalent that proves
  repository-context dependence. Assert live caller, template, and fixture
  bytes all contain the repository-explicit invocation.
- Expected result: recovery job `98718739912`'s class cannot recur; tests
  detect missing repository context and unsupported identity projections,
  not just missing substring comments.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-08 — Recovery still requires PR-bound pr-validation equivalence

- Covers: `VOC-139-AC-04`
- Preconditions: existing VOC-138 recovery attestation tests
- Procedure: Keep rejecting generic `display_title: pipeline` /
  `--squash-safe-push` same-head evidence. Keep accepting a genuine recovery
  execution bound to PR #1090, immutable base/head, repository,
  `main` ← `develop`, expected workflow/path, and `pr-validation`. Assert
  recovery still does not POST `/actions/runs/<doomed-id>/rerun` for
  structurally doomed `pull_request` `ci / ci` jobs.
- Expected result: VOC-138 recovery semantics are retained, not replaced by
  a weaker path.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-09 — Seven VOC-112 paths stay frozen; provenance test is the allowed exception

- Covers: `VOC-139-AC-05`
- Preconditions: protected comparison anchor
  `b9e74fc2db4691c48c637639b265d527de9f4505`
- Procedure: For each of the seven remaining no-change paths, assert
  working-tree bytes equal `git show b9e74fc2…:path` and that
  `git diff b9e74fc2…` does not name the path. If the commit or path cannot
  resolve, fail — do not skip. Assert JSON `subject_revision` remains
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Assert
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` is present in
  the implementation diff and implements D02. Assert live VOC-138
  `NO_CHANGE_PATHS` no longer includes that provenance test. Assert
  `git diff` against the implementation PR base does not name files under
  `specs/changes/VOC-138-promotion-pr-ci-fails-voc-112-ancestry-when/`.
  Assert `run-app-checks.sh` and reusable `ci.yml` contain no `git fetch`.
- Expected result: VOC-112 identity remains except for the in-scope hash
  rule; VOC-138 audit records remain.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-10 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-139-AC-06`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `123735c80fec813a5b46a004f3e1122bd425cde2` if that merge still lacks
  repository-explicit metadata or the promotion hash-anchor export.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-11 — Current-state docs match the live contract

- Covers: `VOC-139-AC-07`
- Preconditions: caller docs in the implementation diff
- Procedure: Assert `docs/operations/11-devops-and-ci-cd.md` no longer says
  that accumulated promotion PRs are merge-base/hash-bound for source hashes.
  Assert it states authenticated same-repository `main` ← `develop`
  promotion PRs use `pr-validation` with hashes bound to the reviewed
  head/source revision, ordinary unchanged-fixture PRs remain merge-base
  anchored, and recovery metadata is repository-explicit without a checkout.
  Assert `docs/development/agent-skills.md` and the fixture README match.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-12 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-139-AC-08`
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
- Evidence: `VOC-139-EV-00`

## VOC-139-TEST-13 — No snapshot-gap; reconcile-release handoff

- Covers: `VOC-139-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert
  evidence states that after this correction merges, `reconcile-release` for
  #1089 may merge #1090 (or the live promotion at the then-current `develop`
  head), converges `develop` to the exact main merge SHA, and does not keep
  staging scheduled for tree-equivalent sync. Assert incident runs
  `33130426061`, `33130473438`, `33130527834` and jobs `98718413924`,
  `98718739912` are preserved as audit evidence, not restaged as this
  package's implementation work. Assert closed state is not treated as
  completion proof. Assert root issue #1096 is documented to close only after
  allowlisted metadata from the successful recovery/release run exists.
- Expected result: repair is hash-rule plus metadata plus ordinary release
  handoff, not a snapshot-then-check package.
- Evidence: `VOC-139-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
