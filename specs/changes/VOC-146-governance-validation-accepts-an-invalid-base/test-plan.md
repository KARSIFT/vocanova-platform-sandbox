# VOC-146 — Test Plan

## VOC-146-TEST-00 — Nonexistent `--base` fails governance validation before the success line

- Covers: `VOC-146-AC-00`, `VOC-146-AC-03`
- Preconditions: live `validate-governance.sh` and
  `validate-monitoring-impact.sh`; a resolvable `--head` in the object store;
  a well-formed 40-character hex that is not a commit; no secrets or
  production data
- Procedure: Invoke the issue #1127 class:
  `bash scripts/governance/validate-governance.sh --base 376e00dd769afb0fe850052b3a5cb48f729e73ad --head <resolvable-head>`.
  If `376e00dd…` has become a real object, substitute another well-formed
  unresolved SHA and record it. Assert exit status is nonzero. Assert stdout
  does not contain `Governance structure validation passed.` Invoke
  `validate-monitoring-impact.sh` with the same pair and assert nonzero.
- Expected result: the #1127 "Git fatal then exit 0" class cannot recur.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-01 — Nonexistent `--head` fails governance validation before the success line

- Covers: `VOC-146-AC-01`
- Preconditions: same harness as TEST-00
- Procedure: Invoke `validate-governance.sh` and
  `validate-monitoring-impact.sh` with a resolvable `--base` and an
  unresolved `--head`. Assert nonzero exit and no success line.
- Expected result: missing head is fail-closed, not an empty accepted range.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-02 — Unrelated/no-merge-base revisions fail closed

- Covers: `VOC-146-AC-02`
- Preconditions: a disposable Git repository with two unrelated root commits
  is acceptable; do not mutate this repository's history
- Procedure: Create two resolvable commits with no merge base. Invoke the
  live range loader used by `validate-monitoring-impact.sh` (and the
  classifier) with `$base...$head`. Assert Git's three-dot diff fails and
  the script exits nonzero rather than loading an empty file list.
- Expected result: resolving both revisions is not treated as a valid range
  when the symmetric difference cannot be computed.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-03 — Failed `git diff` status is preserved; `mapfile` process substitution is not the load path

- Covers: `VOC-146-AC-03`
- Preconditions: implementation diff includes the governance scripts
- Procedure: Assert `validate-monitoring-impact.sh` and
  `classify-change-risk.sh` do not load a requested `$base...$head` range via
  `mapfile < <(git diff …)`. Assert a nonzero `git diff` status becomes the
  script's nonzero status. Assert a successful diff with zero names is still
  treated as a valid empty change set (classifier may print `No changed
  files to classify.` only in that valid-empty case).
- Expected result: process substitution cannot swallow the Git failure.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-04 — Partial `--base` or `--head` does not fall through to working-tree discovery

- Covers: `VOC-146-AC-04`
- Preconditions: live wrappers; a resolvable commit
- Procedure: Invoke `validate-monitoring-impact.sh` and
  `classify-change-risk.sh` with only `--base` and with only `--head`. Assert
  each case exits nonzero. Assert the scripts do not classify or validate
  using working-tree / index / untracked files for that invocation.
- Expected result: a partial range request is fail-closed.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-05 — Valid ranges, `--files-from`, and VOC-086 missing-range fail-closed remain

- Covers: `VOC-146-AC-05`
- Preconditions: implementation PR base and head both exist; VOC-086 tests
  remain installed
- Procedure: Invoke `validate-governance.sh --base <implementation-pr-base>
  --head <implementation-head>` and assert exit 0 with the success line when
  structure checks pass. Invoke both wrappers with `--files-from` pointing at
  a well-formed list and assert exit 0 for a non-violating list. Re-run
  `voc086-monitoring-impact.test.mjs` missing-range and `--files-from` cases
  and assert they still pass. Assert `--declarations-only` without a range
  still runs.
- Expected result: the repair does not break valid PR CI or VOC-086.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-06 — `classify-change-risk.sh` fail-closed parity

- Covers: `VOC-146-AC-06`
- Preconditions: same invalid-range fixtures as TEST-00 through TEST-04
- Procedure: Repeat nonexistent `--base`, nonexistent `--head`,
  no-merge-base, and partial-range invocations against
  `classify-change-risk.sh`. Assert nonzero exit and that stdout does not
  contain `No changed files to classify.` for those invalid cases. Invoke a
  valid implementation-PR range and assert the script classifies (expect R4
  for this change when `scripts/governance/*` is in the diff).
- Expected result: governance-policy.yml's risk-floor step cannot accept the
  same malformed range the validator used to accept.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-07 — Exhaustive source search and current docs match the live contract

- Covers: `VOC-146-AC-07`
- Preconditions: caller docs in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for
  missing-range-only fail-closed claims, `mapfile < <(git diff` range
  loaders, and success-after-Git-fatal claims; assert evidence disposes
  every match as updated, intentionally historical, or irrelevant. Assert
  `AGENTS.md`'s monitoring-impact paragraph states that an unresolved commit
  or invalid diff range is fail-closed. Assert historical VOC-086 records
  stay clearly historical. Assert `git diff` against the implementation PR
  base does not name files under `specs/changes/VOC-086-…/` or
  `specs/changes/VOC-112-…/`.
- Expected result: no current-state doc claims a contract the code does not
  implement.
- Evidence: `VOC-146-EV-00`

## VOC-146-TEST-08 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-146-AC-08`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert `validate-governance.sh` (valid range),
  `classify-change-risk.sh` (R4), VOC-146 foundation tests, VOC-086
  monitoring-impact tests, and `git diff --check` pass. Assert
  `t00-evidence.md` names the implementation PR base, states that the live
  head is bound by the App-authored independent-review comment/check, and
  does not require a tracked file in the same commit to equal `HEAD`. Assert
  fixture `roles.yml` is unchanged. Assert no App ID/private-key secret
  changed. Assert no `PINNED_SHA.txt` change, no VOC-112 fixture recapture,
  and no develop/main gap snapshot. Assert closed state is not treated as
  completion proof. Assert root issue #1127 is documented to close only
  after allowlisted metadata from a successful implementation/promotion path
  exists.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible; the repair
  is fail-closed range loading plus ordinary release handoff.
- Evidence: `VOC-146-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
