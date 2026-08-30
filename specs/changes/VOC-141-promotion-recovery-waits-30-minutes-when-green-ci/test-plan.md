# VOC-141 — Test Plan

## VOC-141-TEST-00 — Required SUCCESS `ci / ci` plus unattestable parent dispatches dedicated validation

- Covers: `VOC-141-AC-00`
- Preconditions: `apply_promotion_pr_recovery_plan` (or the live recovery
  loop) callable with a mock dispatch/rerun API; no secrets or production data
- Procedure: Construct a fixture modeled on run `33340381776` / job
  `99334840338`: GitHub required PR checks report `ci / ci` SUCCESS;
  `recovery_complete` is false because `promotion_ci_context_is_attestable`
  cannot bind a unique attestable completed non-carrier parent (selected
  SUCCESS row belongs to a release carrier, or the composed attestable
  summary filtered that row out). Call the live promotion recovery planner
  with that required-check SUCCESS view plus composed evidence. Assert it
  dispatches exactly one `pipeline.yml` with `action=recover-promotion-pr-checks`
  and `promotion_pr_number` bound to the promotion PR. Assert it does not
  POST `/actions/runs/<id>/rerun` for the PR `ci / ci` carrier. Assert it
  does not return complete and does not require waiting `1800` seconds to
  decide to dispatch.
- Expected result: the #1109 "SUCCESS so no dispatch" class cannot recur.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-01 — Planner coverage is not satisfied by helper-only SUCCESS-omitting tests

- Covers: `VOC-141-AC-00`
- Preconditions: same harness as TEST-00
- Procedure: Assert the TEST-00 dispatch is produced through
  `apply_promotion_pr_recovery_plan` or the live `main()` recovery loop, not
  only `suppress_active_or_successful_dispatches` with `gate_summary=None`.
  If that suppress helper is reused, feed it GitHub-required SUCCESS rows
  plus unattestable composed evidence and assert `pipeline.yml` remains.
- Expected result: VOC-140 helper tests that omit SUCCESS required rows
  cannot be treated as covering issue #1109.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-02 — Valid completed dedicated parent completes without redispatch

- Covers: `VOC-141-AC-01`
- Preconditions: fixture modeled on successful run `33341923799`
- Procedure: Supply a completed successful `pipeline.yml` `workflow_dispatch`
  run with `display_title` `promotion-pr-validation PR #<n>`, exact head,
  exact PR binding, and path `.github/workflows/pipeline.yml`. Assert
  attestation accepts that run for `ci / ci`. Assert `recovery_complete` is
  true even though the shared workflow definition contains a skipped
  `release` job. Assert the live planner does not dispatch another
  `recover-promotion-pr-checks`. Assert a same-head `display_title: pipeline`
  / `--squash-safe-push` run is still rejected.
- Expected result: dedicated recovery remains the trusted completed identity.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-03 — Active or successful exact dedicated recovery suppresses duplicates

- Covers: `VOC-141-AC-02`
- Preconditions: promotion recovery planner callable
- Procedure: Using the TEST-00 unattestable SUCCESS fixture, add a queued or
  in-progress exact-head dedicated `promotion-pr-validation PR #<n>` run and
  assert a duplicate dedicated dispatch is suppressed. Repeat with a
  completed successful dedicated run and assert redispatch is suppressed.
  Repeat with an in-progress or successful release carrier, including a
  second native carrier shaped like `33340516672`, and assert dedicated
  dispatch is **not** suppressed.
- Expected result: only exact dedicated recovery covers the dispatch.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-04 — Timeout diagnostics name unattestable CI evidence

- Covers: `VOC-141-AC-03`
- Preconditions: `collect_missing` / `format_timeout_diagnostics` (or the
  live timeout path) callable
- Procedure: Feed the TEST-00 snapshot at timeout. Assert diagnostics include
  a stable sanitized token such as `unattestable_ci_evidence` (or the live
  equivalent) and still include mode, target SHA, promotion PR number,
  timeout seconds, pending, failed, and successful counts. Assert the output
  is not only `missing_checks: none` with no unattestable reason. Assert no
  token, secret, or raw provider payload is printed.
- Expected result: job `99334840338`'s opaque `missing_checks: none` timeout
  cannot recur without naming the unattestable CI reason.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-05 — Production and release-carrier fail-closed boundaries remain unchanged

- Covers: `VOC-141-AC-04`
- Preconditions: existing VOC-140 carrier/guard tests remain
- Procedure: Keep rejecting in-progress, failed, cancelled, and queued
  release carriers as attestable `ci / ci`. Keep rejecting generic
  `display_title: pipeline` / `--squash-safe-push` same-head evidence. Keep
  requiring PR number, repository, immutable base/head, `main` ← `develop`,
  expected workflow/path, and `pr-validation`. Assert recovery still does not
  rerun structurally doomed `pull_request` `ci / ci` jobs. Assert the
  production merge guard still requires `bypass_actors: []`, still fails
  distinctly on omitted/non-array `bypass_actors`, and that no App-token mint
  permission set changed. Assert the implementation never writes bypass
  actors and never posts fabricated statuses.
- Expected result: prior recovery and guard semantics are retained, not
  replaced by a weaker path.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-06 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-141-AC-05`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `67bdfd13ef875dead23ce4be01d7d0e8b976e289` if that merge still omits
  dedicated dispatch for unattestable SUCCESS CI. Discover every live
  governance test that asserts `CURRENT_PIN`, the live fixture pin, or
  mirrored-file hashes and assert it expects the new merge. Assert historical
  `AUTHORITATIVE_PIN` / issue-era constants and package evidence still retain
  their original merge, using separate current and historical constants
  where needed.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-07 — Exhaustive source/pin search and current docs match the live contract

- Covers: `VOC-141-AC-06`
- Preconditions: caller docs in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for claims that
  required-check SUCCESS completes promotion recovery, that dedicated
  validation is dispatched whenever no completed non-carrier run exists, and
  for the old pin / `CURRENT_PIN` / mirrored hashes; assert evidence disposes
  every match as updated, intentionally historical, or irrelevant. Assert
  fixture README and `docs/operations/11-devops-and-ci-cd.md` state that
  GitHub-required SUCCESS is not sufficient when composed CI is unattestable
  and that recovery dispatches dedicated validation immediately rather than
  polling to timeout. Assert historical VOC-140 records stay clearly
  historical. Assert `git diff` against the implementation PR base does not
  name files under `specs/changes/VOC-140-…/`, `specs/changes/VOC-139-…/`, or
  `specs/changes/VOC-138-…/`.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-08 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-141-AC-07`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  `validate-governance.sh`, `classify-change-risk.sh` (R4), and
  `git diff --check` pass. Assert `t00-evidence.md` names the implementation
  PR base and new infra merge, states that the live head is bound by the
  App-authored independent-review comment/check, and does not require a
  tracked file in the same commit to equal `HEAD`. Assert fixture
  `roles.yml` is unchanged. Assert no App ID/private-key secret changed.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible.
- Evidence: `VOC-141-EV-00`

## VOC-141-TEST-09 — No snapshot-gap; reconcile-release handoff

- Covers: `VOC-141-AC-08`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  duplicate promotion PR or release-audit creation helper was added. Assert
  evidence states that after this correction merges, ordinary
  `reconcile-release` may recover the live same-repository promotion without
  waiting 1,800 seconds on unattestable SUCCESS `ci / ci`, converges
  `develop` to the exact main merge SHA, and does not keep staging scheduled
  for tree-equivalent sync. Assert incident runs `33340381776`,
  `33340516672`, `33341923799`, `33342062118` and job `99334840338` are
  preserved as audit evidence, not restaged as this package's implementation
  work. Assert closed state is not treated as completion proof. Assert root
  issue #1109 is documented to close only after allowlisted metadata from a
  successful recovery/release run exists.
- Expected result: repair is dispatch-plus-diagnostics plus ordinary release
  handoff, not a snapshot-then-check package.
- Evidence: `VOC-141-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
