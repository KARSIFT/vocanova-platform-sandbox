# VOC-122 — Test Plan

## VOC-122-TEST-00 — Absent then cancelled selected row is rerun once and succeeds

- Covers: `VOC-122-AC-00`, `VOC-122-AC-02`, `VOC-122-AC-06`
- Preconditions: synthetic `promotion_pr` runner fixture modeled on
  `tests/test_voc121_actions_check_recovery.py`, with a time-evolving
  `run_metadata_phase` sequence and a short `--timeout-seconds`
- Procedure: Snapshot 1 required view omits `governance-policy` (other
  required contexts SUCCESS). Snapshot 2 shows that context as `CANCELLED`
  `pull_request` / `Governance policy` with an exact Actions run link (the
  #1000 class). Snapshot 3 shows all required contexts SUCCESS. Assert:
  snapshot 1 produces at most one allowlisted `governance-policy` dispatch and
  no rerun; snapshot 2 produces exactly one `rerun_selected_workflow` of that
  run ID after identity validation; snapshot 3 completes with exit 0. A
  timeout with `dispatched` containing only the snapshot-1 dispatch is a
  failure.
- Expected result: the first recovery invocation recovers the later-appearing
  cancelled row instead of polling until timeout.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-01 — Absent-context dispatch is not repeated on later snapshots

- Covers: `VOC-122-AC-02`, `VOC-122-AC-06`
- Preconditions: same runner fixture; required context remains absent across
  two or more snapshots, then appears SUCCESS from the dispatched work, or
  remains absent until timeout is not the assertion — the assertion is no
  second dispatch
- Procedure: Drive repeated absent snapshots for one required context. Assert
  `dispatch_workflow` is called once for that context and never again.
- Expected result: genuinely absent contexts dispatch at most once per
  recovery invocation.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-02 — Selected run IDs are not rerun again on later snapshots

- Covers: `VOC-122-AC-02`, `VOC-122-AC-06`
- Preconditions: cancelled selected row already rerun once; later snapshots
  still show that row cancelled or pending while attempt 2 is in flight
- Procedure: After the first validated rerun, feed repeated cancelled or
  pending required-view snapshots for the same run ID. Assert
  `rerun_selected_workflow` is not called again and `run_attempt != 1`
  payloads are refused if a replan would try them.
- Expected result: one rerun per run ID per invocation; first-attempt identity
  remains load-bearing.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-03 — Replan reruns still bind PR, SHA, branch, path, event, and attempt

- Covers: `VOC-122-AC-01`
- Preconditions: existing `validate_selected_workflow_run` contract from
  VOC-121 tests
- Procedure: Keep the VOC-121 identity assertions. Add a replan-path case
  where snapshot 2's selected run fails validation on each of `head_sha`,
  `head_branch`, `event`, workflow path, promotion PR number, and
  `run_attempt`. Assert the runner exits non-zero and does not rerun or
  dispatch an override.
- Expected result: every rerun, including during polling, is identity-checked.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-04 — Alternate success or same-named status cannot suppress a newly appeared selected row

- Covers: `VOC-122-AC-03`, `VOC-122-AC-06`
- Preconditions: snapshot 1 absent-context dispatch that later succeeds;
  snapshot 2 cancelled selected pull-request row for the same context name;
  optional successful same-named commit status
- Procedure: Assert recovery still plans and performs the exact selected-run
  rerun. `recovery_complete` remains false while `gh pr checks --required`
  would report the cancelled/failed row. Attestation, if invoked in adjacent
  tests, still refuses that state.
- Expected result: VOC-121 selection survives the replan path; the #1000
  dispatch success does not hide run `32912850066`'s class of evidence.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-05 — Ambiguous, foreign, and mismatched rows fail closed during replan

- Covers: `VOC-122-AC-04`, `VOC-122-AC-06`
- Preconditions: initial snapshot that is recoverable or absent; later
  snapshot that is unsafe
- Procedure: Separate cases for (1) two candidate run IDs for one failed
  required context, (2) foreign workflow name or non-`pull_request` event,
  (3) selected-run payload that fails identity checks. Drive each as snapshot
  2 after a legal snapshot 1. Assert non-zero exit, no override dispatch, and
  no continued poll-as-success.
- Expected result: replan is as fail-closed as the initial plan.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-06 — Timeout, token split, merge decision, and integration_push remain

- Covers: `VOC-122-AC-05`
- Preconditions: final implementation branch
- Procedure: Inspect runner and callers for preserved
  `DEFAULT_TIMEOUT_SECONDS = 1800`, `POLL_INTERVAL_SECONDS = 30`, job-token
  recovery versus App merge, exact-head merge-decision checks in
  `release.yml`, first-attempt rerun validation, and unchanged
  `integration_push` planning semantics (or an equivalence test if a shared
  helper was extracted). Confirm no status fabrication and no ruleset bypass.
- Expected result: reliability fix lands without weakening existing safety
  gates.
- Evidence: `VOC-122-EV-00`

## VOC-122-TEST-07 — Source self-CI, caller suites, docs, and pin match consumption

- Covers: `VOC-122-AC-07`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. Confirm current-state comments no longer claim
  one-shot pre-loop planning. If the fixture consumed the infra change, assert
  the pin equals the exact reviewed infra merge and that
  `test_voc121_implement_policy.py` plus any `scripts/foundation/voc097`,
  `voc104`, and `voc108` pin literals match. If not, assert the pin is
  unchanged and evidence records why.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable.
- Evidence: `VOC-122-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
