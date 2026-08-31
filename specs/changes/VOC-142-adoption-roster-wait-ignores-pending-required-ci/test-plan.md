# VOC-142 — Test Plan

## VOC-142-TEST-00 — Continued IN_PROGRESS required `ci / ci` keeps roster wait incomplete

- Covers: `VOC-142-AC-00`, `VOC-142-AC-02`
- Preconditions: live roster-wait path callable with fixture check-runs,
  statuses, pull-request identity, and a required-context set; no secrets or
  production data
- Procedure: Construct a fixture modeled on run `33343125733` / job
  `99342230038` at `2026-08-30T23:58:46Z`: exact roster head
  `98dd0936a73b64a6b548da6cf2000a6d000917ac`; required set includes `ci / ci`,
  `governance-policy`, and `validate`; `governance-policy` and `validate` are
  SUCCESS; `ci / ci` (job `99342299218`) is IN_PROGRESS with a non-terminal
  parent pipeline run `33343147453`. Invoke the live wait-completeness
  predicate used by `Wait for roster PR checks` twice with an unchanged head
  SHA. Assert wait does not complete. Assert merge is not attempted. Assert
  the IN_PROGRESS `ci / ci` row is treated as not-ready, not omitted as if
  pending were 0.
- Expected result: the #1113 "stable partial green so merge" class cannot
  recur.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-01 — Late registration of required `ci / ci` is fail-closed

- Covers: `VOC-142-AC-00`, `VOC-142-AC-01`
- Preconditions: same harness as TEST-00
- Procedure: Supply two consecutive snapshots that contain only SUCCESS
  `governance-policy` and `validate`, with `pending == 0`, `failed == 0`,
  identical `total_count`, and identical head SHA, while `ci / ci` is not
  yet registered. Assert wait does not complete. Assert a later snapshot
  that adds IN_PROGRESS `ci / ci` still does not complete. Assert wait
  completes only after newest logical attempts of `ci / ci`,
  `governance-policy`, and `validate` are all SUCCESS. Assert failure if the
  required set cannot be uniquely resolved.
- Expected result: two stable zero-pending subset snapshots are not treated
  as a complete required set.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-02 — Existing exact open roster PR is reused

- Covers: `VOC-142-AC-03`
- Preconditions: `Open roster PR` path callable with a mock `gh` API; a
  pushed deterministic `karsift/roster-*` ref at a known 40-character SHA
- Procedure: Construct a fixture modeled on reconcile run `33343250178`:
  `changed == 'true'`; exactly one OPEN PR (#1112-shaped) with matching
  repository, integration-branch base, `karsift/roster-*` head ref, and
  exact pushed head SHA. Assert the step records that PR number and does
  not invoke `gh pr create`. Assert wait/merge/cleanup receive that same
  PR number.
- Expected result: the #1113 "PR already exists" create failure cannot
  recur for an exact matching open carrier.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-03 — Mismatched and ambiguous carriers are rejected

- Covers: `VOC-142-AC-04`
- Preconditions: same harness as TEST-02
- Procedure: Repeat TEST-02 with (a) two OPEN PRs for the same head ref,
  (b) one OPEN PR whose head SHA differs from the pushed SHA, (c) one OPEN
  PR whose base ref differs, and (d) one OPEN PR whose head repository
  differs. Assert each case fails closed and does not call `gh pr create`.
  Assert an unrelated PR is not closed, recreated, or retargeted. Assert
  zero matching OPEN or already-merged carriers may create exactly one new
  PR.
- Expected result: only a unique exact identity match may be reused;
  ambiguity and mismatch fail closed.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-04 — Already-merged exact roster carrier is reused, not duplicated

- Covers: `VOC-142-AC-05`
- Preconditions: same harness as TEST-02
- Procedure: Supply exactly one MERGED PR with matching repository, head
  ref, exact head SHA, and base ref, and no OPEN match. Assert `gh pr create`
  is not called. Assert the flow continues into the existing merged-roster
  cleanup/dispatch path rather than opening a second carrier.
- Expected result: idempotent reconcile after a successful roster merge
  does not create a duplicate PR.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-05 — Existing tasks are reused and root dispatch happens exactly once

- Covers: `VOC-142-AC-06`
- Preconditions: adopt issue-list and root-dispatch path callable
- Procedure: With an existing OPEN task issue shaped like #1111 and no
  implementation PR for the deterministic implementation head, assert
  reconcile reuses that issue number and does not open a second task.
  After a successful reused-roster merge, assert root dispatch is needed
  only while the issue is OPEN and `pr_count == 0`. Assert a second
  reconcile after dispatch (`pr_count != 0` or issue not OPEN) sets
  `needed=false`. Assert this package never creates a replacement for
  task #1111.
- Expected result: interrupted adoption resumes without duplicate task or
  duplicate implementation dispatch.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-06 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-142-AC-07`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `67bdfd13ef875dead23ce4be01d7d0e8b976e289` if that merge still completes
  wait without required `ci / ci` or still creates a PR when the exact open
  carrier exists. Discover every live governance test that asserts
  `CURRENT_PIN`, the live fixture pin, or mirrored-file hashes and assert it
  expects the new merge. Assert historical `AUTHORITATIVE_PIN` / issue-era
  constants and package evidence still retain their original merge, using
  separate current and historical constants where needed.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-07 — Exhaustive source/pin search and current docs match the live contract

- Covers: `VOC-142-AC-08`
- Preconditions: caller docs in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for claims that
  two stable zero-pending snapshots complete roster wait, that reconcile
  always opens a new roster PR, and for the old pin / `CURRENT_PIN` /
  mirrored hashes; assert evidence disposes every match as updated,
  intentionally historical, or irrelevant. Assert `AGENTS.md`'s reconcile
  procedure, fixture README, and `adopt.yml` header comments state that wait
  requires the complete required set including `ci / ci` and that reconcile
  reuses a matching open or already-merged carrier. Assert historical
  VOC-141 records stay clearly historical. Assert `git diff` against the
  implementation PR base does not name files under
  `specs/changes/VOC-141-…/`, `specs/changes/VOC-140-…/`, or
  `specs/changes/VOC-139-…/`.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-08 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-142-AC-09`, `VOC-142-AC-02`
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
  Assert roster wait still does not contain `statusCheckRollup` or
  `gh pr checks`. Assert merge-gate/release still reject in-progress parents
  as attestable completed `ci / ci`.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible; attestation
  fail-closed boundaries remain.
- Evidence: `VOC-142-EV-00`

## VOC-142-TEST-09 — No snapshot-gap; reconcile of #1110 can resume #1112

- Covers: `VOC-142-AC-10`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  duplicate VOC-141 task, roster PR, promotion-PR, or release-audit creation
  helper was added. Assert no helper merges, closes, or recreates #1112.
  Assert evidence states that after this correction merges, ordinary
  `reconcile` for plan PR #1110 may reuse #1112 when it still matches, wait
  for complete required checks, merge, clean up, and dispatch the first
  not-yet-dispatched task exactly once. Assert incident runs `33343125733`,
  `33343147453`, `33343250178` and jobs `99342230038`, `99342299218`,
  `99342577393` are preserved as audit evidence, not restaged as this
  package's implementation work. Assert closed state is not treated as
  completion proof. Assert root issue #1113 is documented to close only
  after allowlisted metadata from a successful adopt/reconcile run exists.
  Assert no bootstrap exception was used.
- Expected result: repair is wait-completeness plus carrier reuse plus
  ordinary release handoff, not a snapshot-then-check package and not a
  manual merge of #1112.
- Evidence: `VOC-142-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
