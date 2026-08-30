# VOC-140 — Test Plan

## VOC-140-TEST-00 — Newest SUCCESS `ci / ci` on a still-running release carrier is not attestable

- Covers: `VOC-140-AC-00`, `VOC-140-AC-05`
- Preconditions: recovery, authoritative selection, and attestation harnesses
  callable; no secrets or production data
- Procedure: Construct a fixture modeled on run `33136633666` / job
  `98738317266`: required PR checks show `ci / ci` SUCCESS; the newest
  `ci / ci` check run is completed/success; its parent workflow run is
  `status=in_progress`, path `.github/workflows/pipeline.yml`, and actually
  executes a non-skipped `release / converge` job (or has equivalent
  `reconcile-release` event/input/title carrier identity). Assert
  `recovery_complete` is false. Assert
  `verify_promotion_required_run_semantics` / attestation fails
  `untrusted_ci_recovery_identity` (or the live equivalent) for that run_id.
  Assert authoritative selection does not choose that run as attestable
  `ci / ci`.
- Expected result: the circular-CI class cannot be treated as completed CI.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-01 — Failed or non-terminal release carriers remain untrusted

- Covers: `VOC-140-AC-00`
- Preconditions: same harness as TEST-00
- Procedure: Repeat TEST-00 with parent workflow `conclusion=failure`,
  `cancelled`, or `status=queued`. Assert each remains untrusted and cannot
  suppress dedicated recovery dispatch.
- Expected result: only completed successful non-carrier runs may attest.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-02 — Recovery dispatches dedicated `promotion-pr-validation` when no completed non-carrier `ci / ci` exists

- Covers: `VOC-140-AC-01`
- Preconditions: promotion recovery planner callable
- Procedure: Using the TEST-00 fixture, assert the remaining dispatch plan
  includes `pipeline.yml` with `action=recover-promotion-pr-checks` and
  `promotion_pr_number` bound to the promotion PR. Assert an in-progress
  release-carrier `pipeline.yml` run does not suppress that plan. Assert a
  queued/in-progress dedicated run with `display_title`
  `promotion-pr-validation PR #<n>` does suppress a duplicate dispatch.
- Expected result: job `98738317266`'s "complete without dispatch" class
  cannot recur when the only SUCCESS `ci / ci` belongs to the release
  carrier.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-03 — Completed dedicated `promotion-pr-validation` is attestable and suppresses redispatch

- Covers: `VOC-140-AC-01`
- Preconditions: fixture modeled on successful run `33136865709`
- Procedure: Supply a completed successful `pipeline.yml` `workflow_dispatch`
  run with `display_title` `promotion-pr-validation PR #1090`, exact head
  `21eef755…` (or the fixture SHA), exact PR binding, and path
  `.github/workflows/pipeline.yml`. Assert attestation accepts that run for
  `ci / ci`. Assert recovery_complete is true even though the shared workflow
  definition contains a skipped `release` job. Assert redispatch is suppressed.
  Assert a same-head `display_title: pipeline` /
  `--squash-safe-push` run is still rejected.
- Expected result: dedicated recovery remains the trusted dispatch identity.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-04 — VOC-138/VOC-139 exact-identity recovery semantics are retained

- Covers: `VOC-140-AC-02`
- Preconditions: existing VOC-138 recovery attestation tests
- Procedure: Keep rejecting generic `display_title: pipeline` /
  `--squash-safe-push` same-head evidence. Keep requiring PR number,
  repository, immutable base/head, `main` ← `develop`, expected
  workflow/path, and `pr-validation`. Assert recovery still does not POST
  `/actions/runs/<doomed-id>/rerun` for structurally doomed `pull_request`
  `ci / ci` jobs. Assert a completed exact-bound `pull_request` `ci / ci`
  whose parent run is completed/successful and is not a release carrier may
  still attest, and that it does not win over a newer in-progress carrier
  check.
- Expected result: prior recovery semantics are retained, not replaced by a
  weaker path.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-05 — Empty-bypass qualifying ruleset still passes the guard

- Covers: `VOC-140-AC-03`
- Preconditions: verifier and runner callable
- Procedure: Feed a payload shaped like administrator-visible ruleset
  `20575146`: repository-owned, active, production-branch-effective, strict,
  non-empty required checks, pull-request rule, `bypass_actors: []`. Assert
  the verifier prints `production-merge-guard: ok ruleset_id=…` and returns
  that id.
- Expected result: the live guard continues to accept a real empty-bypass
  ruleset.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-06 — Non-empty bypass and incomplete rulesets still fail closed

- Covers: `VOC-140-AC-03`
- Preconditions: existing production-merge-guard negatives remain
- Procedure: Keep rejecting non-strict checks, empty/malformed required-check
  lists, inactive enforcement, missing pull-request rule, source mismatch,
  and any non-empty `bypass_actors`. Assert the implementation never writes
  bypass actors into a ruleset and never posts fabricated statuses.
- Expected result: guard strength is unchanged.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-07 — Omitted `bypass_actors` fails distinctly from a missing guard

- Covers: `VOC-140-AC-04`, `VOC-140-AC-05`
- Preconditions: a mock `gh` on `PATH`; no secrets
- Procedure: Execute the real `verify-production-merge-guard.sh` (or the
  helper it actually calls) against a ruleset JSON that matches GitHub's
  App-token-visible shape: the HTTP fetch succeeds, `bypass_actors` is
  omitted or a non-array. Assert failure matching the new sanitized class
  (`production_merge_guard_payload_incomplete` or the live equivalent) and
  not `production_merge_guard_missing`. Assert stderr names the operator
  action to set `karsift-ai-infra-bot` Administration: Read and write, obtain
  owner approval on KARSIFT organization installation `148001476`, retain the
  explicit current-repository token restriction for
  `KARSIFT/vocanova-platform-sandbox`, avoid secret rotation, and rerun. Assert
  it does not print tokens or private-key material. Repeat with a full
  administrator-visible payload including `bypass_actors: []` and assert
  `ok`.
- Expected result: run `33136984634`'s class is diagnosed as token-visible
  payload incompleteness, not as a missing live ruleset.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-08 — Both workflow paths enforce exact two-token isolation

- Covers: `VOC-140-AC-04`
- Preconditions: live `release.yml` and `merge-gate.yml` after the infra
  merge
- Procedure: Parse `release.yml` and `merge-gate.yml`. For each production
  merge path, assert the existing mutation mint has exactly
  `permission-contents: write`, `permission-issues: write`, and
  `permission-pull-requests: write`, has no Administration permission, and is
  the only App-token output supplied to `gh pr merge` or mutation steps.
  Assert a distinct mint appears immediately before guard verification, has
  exactly `permission-administration: write`, explicit `owner`, and
  `repositories` limited to the current caller repository; it has no Contents,
  Issues, Pull requests, Statuses, Actions, or Workflows grant. Assert its
  output is injected only into `verify-production-merge-guard.sh`, that guard
  success precedes a fresh exact-head/base/ref revalidation and merge, and that
  no guard-token output/expression appears in `gh pr merge`, mutation, status,
  issue, PR, content, or completion-marker steps. Assert neither identity is
  replaced by `github.token`.
- Expected result: exact permissions, repository scoping, step order, and
  credential-use isolation are machine-proven in both workflows.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-09 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-140-AC-06`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `599436835371f27fac52ec6b47a18b36257366ac` if that merge still selects a
  still-running release carrier as `ci / ci` or still cannot prove
  `bypass_actors` with the isolated guard token. Discover every live governance
  test that asserts `CURRENT_PIN`, the live fixture pin, or mirrored-file
  hashes and assert it expects the new merge. Assert historical
  `AUTHORITATIVE_PIN` / issue-era constants and package evidence still retain
  their original merge, using separate current and historical constants
  where needed.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-10 — Exhaustive source/pin search and current docs match the live contract

- Covers: `VOC-140-AC-07`
- Preconditions: caller docs in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for the old pin,
  `CURRENT_PIN`, mirrored hashes, mutation-only/token permission claims,
  active-A-003 claims, and disabled/unimplemented automatic-release or
  production-deploy claims; assert evidence disposes every match as updated,
  intentionally historical, or irrelevant. Assert fixture README and
  `docs/operations/11-devops-and-ci-cd.md` distinguish the unchanged mutation
  token from the guard-only token. Assert
  `docs/governance/repository-settings.md`, the activation checklist, and
  DOC-19 identify active A-004/current enabled repository-controlled release
  and production deployment while retaining RL1/RL2 disabled. Assert docs state
  recovery/selection never
  treats a still-running release carrier as attestable `ci / ci`, dedicated
  `promotion-pr-validation` must be completed/successful, and the merge App
  guard token is Administration-write-only, current-repository scoped, and
  never used for merge/mutations. Assert historical A-003 and issue-era pin
  records stay clearly historical. Assert
  governance validators/tests no longer require stale current-state wording
  and retain unrelated fail-closed invariants. Assert
  `git diff` against the implementation PR base does not name files under
  `specs/changes/VOC-139-…/` or `specs/changes/VOC-138-…/`.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-11 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-140-AC-08`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  `validate-governance.sh`, `classify-change-risk.sh` (R4), and
  `git diff --check` pass. Assert `t00-evidence.md` names the implementation
  PR base and new infra merge, states that the live head is bound by the
  App-authored independent-review comment/check, and does not require a
  tracked file in the same commit to equal `HEAD`. Assert fixture
  `roles.yml` is unchanged. Assert deterministic fixtures prove omission,
  non-array, and non-empty values fail closed and that T00 records the precise
  post-merge operator action without requiring owner approval or a hosted probe
  to close the implementation task. Assert no App ID/private-key secret changed.
  Assert evidence records the shared App/private-key residual risk and labels a
  dedicated guard App as optional future hardening only.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible.
- Evidence: `VOC-140-EV-00`

## VOC-140-TEST-12 — No snapshot-gap; reconcile-release handoff

- Covers: `VOC-140-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  duplicate promotion PR or release-audit creation helper was added. Assert
  evidence states that after this correction merges, dedicated promotion
  recovery may be rerun if necessary, then `reconcile-release` for #1089 may
  merge #1090 (or the live promotion at the then-current `develop` head),
  converges `develop` to the exact main merge SHA, and does not keep staging
  scheduled for tree-equivalent sync. Assert incident runs `33136633666`,
  `33136865709`, `33136984634`, `33137091931` and jobs `98738317266`,
  `98739074178`, `98739420310` are preserved as audit evidence, not restaged
  as this package's implementation work. Assert closed state is not treated
  as completion proof. After T00 merges, assert the KARSIFT installation owner
  approved Administration: Read and write and a hosted guard-only token scoped
  to this caller repository received explicit `bypass_actors: []`; omission,
  non-array, and non-empty values remain fail closed. Assert root issue #1102
  is documented to close only after this activation proof and allowlisted
  metadata from the successful recovery/release run exist.
- Expected result: repair is identity plus token/API plus ordinary release
  handoff, not a snapshot-then-check package.
- Evidence: `VOC-140-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
