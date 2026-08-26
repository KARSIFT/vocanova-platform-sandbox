# VOC-122 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` recovery/release workflows and
  Python policy modules that plan or mutate required Actions runs; caller
  `tooling/governance/` fixtures and tests.
- Prerequisites: confirm the live `main()` still plans
  `plan_required_check_recovery` / exact rerun / absent-context dispatch only
  before the `while` loop, and that the loop only calls `run_metadata_phase`
  and `recovery_complete`. Confirm VOC-121 selected-run validation
  (`validate_selected_workflow_run`, `run_attempt == 1`, required PR view)
  remains the baseline this change must preserve.
- VOC-121's isolated source publisher is already live. Use it. Do not treat an
  untracked local `karsift-ai-infra/` checkout as this repository's tracked
  tree.
- Preserve 1800-second timeout, 30-second poll interval, two-attempt
  implementer bounds, exact-SHA independent review, fail-closed credentials,
  and runner/App-token isolation.

## File reconciliation and implementation sequence

### T00 — Replan newly appearing required-check rows during promotion-recovery polling

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/config/actions-check-recovery-runner.py` | modify | Apply required-check planning to the current PR view during polling with invocation-scoped run-ID and absent-context dedupe |
| `KARSIFT/karsift-ai-infra/config/required_check_satisfaction.py` | modify only if needed | Snapshot classifier already distinguishes absent versus cancelled; change only if invocation-scoped dedupe belongs here |
| `KARSIFT/karsift-ai-infra/config/actions_check_recovery.py` | modify only if needed | Timeout/poll constants and `recovery_complete` stay; no `integration_push` semantic change |
| `KARSIFT/karsift-ai-infra/.github/workflows/recover-actions-checks.yml` | modify comments as needed | Current-state: replan during wait, not one-shot pre-loop planning |
| `KARSIFT/karsift-ai-infra/.github/workflows/release.yml` | modify comments only if they describe one-shot planning | Keep exact-head merge decision and 1800-second recovery call |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state recovery paragraph must describe replan during polling |
| `KARSIFT/karsift-ai-infra/tests/test_voc122_*.py` or extend `test_voc121_actions_check_recovery.py` | create/extend | Time-evolving #1000 class plus idempotence and fail-closed replan cases |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed |
| `tooling/governance/tests/test_voc121_implement_policy.py` and/or new VOC-122 fixture test | modify/extend | Fixture regressions for replan; advance pin literal when consumed |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify if they still assert the previous pin | Same-task pin literals |
| `specs/changes/VOC-122-.../t00-evidence.md` | update | Record mechanism, commands, results, infra SHA, pin applicability |

Ordered steps:

1. In `KARSIFT/karsift-ai-infra`, change promotion recovery so required-check
   planning runs against the current required PR view at start and during
   polling. Persist already-rerun run IDs and already-dispatched contexts for
   the whole invocation. Keep identity validation before every rerun.
2. Do not reset the 1800-second deadline. Fail closed immediately on
   ambiguous, foreign, or mismatched later snapshots. Leave pending rows to
   finish. Preserve VOC-121: a successful alternate run or same-named status
   does not satisfy a cancelled selected row.
3. Add deterministic time-evolving tests for the #1000 class (absent →
   cancelled selected row → one rerun → success), repeated-snapshot
   idempotence, and fail-closed replan cases.
4. Update current-state comments/docs so they describe replan during polling.
5. Run the infra unit/policy suite. Open one reviewed infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first when the caller fixture consumes the
   change.
6. Sync and pin the caller fixture to that exact merge SHA when consumed;
   update caller governance and foundation pin tests; record evidence in
   `t00-evidence.md`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds narrower targeted commands (for example
`python3 -m unittest tests.test_voc122_actions_check_recovery` or the three
foundation pin tests), record the exact commands in `t00-evidence.md` and run
them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- a required row that appears after the initial snapshot is recovered in the
  same invocation;
- reruns remain identity-checked and once-per-run-ID;
- absent contexts dispatch at most once;
- alternate success or same-named status cannot override GitHub's selected
  row;
- ambiguous/foreign/mismatched later snapshots fail closed;
- 1800-second timeout, App-token split, exact-head merge decision, and
  `integration_push` semantics remain;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable;
- the implementer did not approve or merge its own work on either carrier.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future promotion-recovery jobs replan newly
  appearing required rows during the 1800-second wait instead of timing out
  while a recoverable cancelled/failed row is visible.
- **Rollback trigger:** Replan duplicates dispatches/reruns; fails closed on
  legitimate pending rows; or treats alternate success as satisfying a
  selected cancelled row.
- **Rollback mechanism:** Revert the infra and caller fixture/test/doc
  changes to the prior reviewed VOC-121 recovery behavior.
- **Last-known-good reference:** Current recovery runner/modules on
  `main`/`develop` after VOC-121 (infra merge `99476c2a1018e42d4bd442657b5257885ac9f1c9`)
  and before VOC-122 implementation lands.
