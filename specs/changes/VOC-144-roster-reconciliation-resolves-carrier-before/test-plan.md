# VOC-144 — Test Plan

## VOC-144-TEST-00 — Stale listed head then exact pushed head reuses the open carrier

- Covers: `VOC-144-AC-00`, `VOC-144-AC-01`, `VOC-144-AC-03`
- Preconditions: live `roster-carrier-runner.py` (or equivalent adapter)
  callable with a mock `gh` API and injected snapshots or a fake clock; no
  secrets or production data
- Procedure: Construct a fixture modeled on reconcile run `33437239322` then
  `33437514152`: `changed == 'true'`; local HEAD is a known 40-character SHA
  (run-2 shape `0206cb70437cb751a19a2e715d8202b672060b50` is sufficient);
  exactly one OPEN PR (#1112-shaped) with matching repository, integration
  branch base, and `karsift/roster-*` head ref; first REST snapshot lists a
  prior SHA (run-1 shape `958e0fedf742173320bb89cfe690ec7070b49e93`); a later
  snapshot lists the exact local HEAD. Invoke the live adapter path used by
  `Open roster PR`. Assert the step records that PR number, action
  `reuse_open`, and does not invoke `gh pr create`. Assert
  `resolve_roster_carrier` on the first snapshot alone would still be
  `MISMATCHED_OPEN_CARRIER`.
- Expected result: the #1122 "push then immediately mismatch, later
  `reuse_open`" class cannot recur.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-01 — Timeout with a still-stale head remains `MISMATCHED_OPEN_CARRIER`

- Covers: `VOC-144-AC-02`, `VOC-144-AC-03`
- Preconditions: same harness as TEST-00; fake clock or bounded snapshot
  sequence that never converges
- Procedure: Supply an otherwise matching OPEN PR whose listed SHA stays
  different from local HEAD across the full named bound. Assert the adapter
  returns `MISMATCHED_OPEN_CARRIER` (or equivalent fail-closed exit). Assert
  `gh pr create` is not called. Assert the wait used named timeout/poll
  constants, did not use 1,800 seconds, and did not depend on wall-clock
  GitHub lag. Assert a still-stale listed head was not reused.
- Expected result: a stable SHA mismatch after the bound is still fail-closed
  and does not create a second PR.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-02 — Durable mismatches do not enter the SHA-lag wait

- Covers: `VOC-144-AC-01`, `VOC-144-AC-04`
- Preconditions: same harness as TEST-00
- Procedure: Repeat with (a) two OPEN PRs for the same head ref, (b) one
  OPEN PR whose base ref differs, (c) one OPEN PR whose head repository
  differs. Assert each case fails closed without additional SHA-lag polls.
  Assert `gh pr create` is not called. Assert an unrelated PR is not closed,
  recreated, or retargeted.
- Expected result: only unique same-repo/same-ref/same-base SHA lag may wait;
  ambiguity and identity mismatch fail immediately.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-03 — Preserved VOC-142 carrier paths still hold

- Covers: `VOC-144-AC-03`, `VOC-144-AC-05`
- Preconditions: same harness as TEST-00 plus roster-wait fixtures from
  VOC-142
- Procedure: Assert an exact first-snapshot OPEN match reuses immediately
  with no extra wait required for correctness. Assert exactly one
  already-merged PR with matching repository, head ref, exact head SHA, and
  base ref is reused and `gh pr create` is not called. Assert zero matching
  OPEN or already-merged carriers may create exactly one new PR. Assert
  roster wait still does not complete while required `ci / ci` is missing or
  IN_PROGRESS. Assert existing-task reuse and single root dispatch remain.
- Expected result: this repair adds SHA-lag wait without regressing VOC-142
  reuse, create-when-zero, or complete-required-set wait.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-04 — GitHub API or parse failure fails closed

- Covers: `VOC-144-AC-04`
- Preconditions: same harness as TEST-00
- Procedure: Make the first `gh api` list (or equivalent) fail or return
  unparseable JSON. Assert the adapter exits fail-closed, does not wait for
  SHA convergence, and does not call `gh pr create`.
- Expected result: API failure is not treated as "no carrier, create" and is
  not retried as if it were SHA lag.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-05 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-144-AC-06`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert SHA-256 of each changed mirrored file equals the same path at that
  merge. Assert the pin is not left at
  `8993e867640dfb604dec0466c4e0787e68d8e258` if that merge still resolves
  from a single post-push snapshot. Discover every live governance test that
  asserts `CURRENT_PIN`, the live fixture pin, or mirrored-file hashes and
  assert it expects the new merge. Assert historical `AUTHORITATIVE_PIN` /
  issue-era constants and package evidence still retain their original
  merge, using separate current and historical constants where needed.
- Expected result: caller fixtures match the repaired infra contract.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-06 — Exhaustive source/pin search and current docs match the live contract

- Covers: `VOC-144-AC-07`
- Preconditions: caller docs in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for claims that
  carrier resolution is a single immediate REST snapshot, that a same-ref
  SHA difference is always a durable mismatch with no wait, and for the old
  pin / `CURRENT_PIN` / mirrored hashes; assert evidence disposes every
  match as updated, intentionally historical, or irrelevant. Assert
  `AGENTS.md`'s reconcile procedure, fixture README, and `adopt.yml` header
  comments state that reconcile waits boundedly for the existing same-ref
  OPEN PR to expose the exact pushed SHA, then reuses it, and still fails
  closed on a stable mismatch. Assert historical VOC-142/VOC-141 records
  stay clearly historical. Assert `git diff` against the implementation PR
  base does not name files under `specs/changes/VOC-142-…/`,
  `specs/changes/VOC-141-…/`, `specs/changes/VOC-140-…/`, or
  `specs/changes/VOC-139-…/`.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-07 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-144-AC-08`, `VOC-144-AC-05`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  `validate-governance.sh`, `classify-change-risk.sh` (R4), and
  `git diff --check` pass. Assert `t00-evidence.md` names the implementation
  PR base, new infra merge, and named timeout/poll constants; states that
  the live head is bound by the App-authored independent-review
  comment/check; and does not require a tracked file in the same commit to
  equal `HEAD`. Assert fixture `roles.yml` is unchanged. Assert no App
  ID/private-key secret changed. Assert roster wait still does not contain
  `statusCheckRollup` or `gh pr checks`. Assert merge-gate/release still
  reject in-progress parents as attestable completed `ci / ci`.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible; attestation
  and VOC-142 wait fail-closed boundaries remain.
- Evidence: `VOC-144-EV-00`

## VOC-144-TEST-08 — No snapshot-gap; reconcile of #1110 can resume #1112

- Covers: `VOC-144-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  duplicate VOC-141 task, roster PR, promotion-PR, or release-audit creation
  helper was added. Assert no helper merges, closes, or recreates #1112.
  Assert evidence states that after this correction merges, ordinary
  `reconcile` for plan PR #1110 may reuse #1112 when it still matches,
  including across post-push REST lag, wait for complete required checks,
  merge, clean up, and dispatch the first not-yet-dispatched task exactly
  once. Assert incident runs `33437239322` and `33437514152` and heads
  `958e0fed…` / `0206cb70…` are preserved as audit evidence, not restaged as
  this package's implementation work. Assert closed state is not treated as
  completion proof. Assert root issue #1122 is documented to close only
  after allowlisted metadata from a successful adopt/reconcile run exists.
  Assert no bootstrap exception was used.
- Expected result: repair is bounded SHA-lag wait plus ordinary release
  handoff, not a snapshot-then-check package and not a manual merge of
  #1112.
- Evidence: `VOC-144-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
