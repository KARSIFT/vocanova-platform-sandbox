# VOC-122-T00 — Evidence

Task: `VOC-122-T00` — Replan newly appearing required-check rows during
promotion-recovery polling.

Do not record secrets, credentials, session values, OAuth material, personal
data, or complete CI logs.

## Discovery recorded at planning time (issue #1001)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1001 |
| Promotion PR | #1000 |
| `develop` head that opened it | `bb531fc211e91e40cec1847e1e28d98beaedacff` |
| First recovery run / job | `32912818046` / `98010275057` at `2026-08-25T23:54:50Z` |
| Snapshot-1 action | absent-context `governance-policy` dispatch `32912851134` (passed) |
| Later-appearing selected row | pull-request `governance-policy` run `32912850066` at `2026-08-25T23:54:52Z`, `CANCELLED` |
| Required view | `gh pr checks 1000 --required` continued to select that row as `CANCELLED` |
| Defect | `plan_required_check_recovery` / exact rerun / dispatch planning only before the `while` loop in `config/actions-check-recovery-runner.py` |
| Later recovery that worked | convergence run `32912851044` / job `98010820950` reran `32912850066` as attempt 2 |
| Merge | App merged #1000 at `70db9d7dbc4789264732e4fe4f347914dd6f764e` with no ruleset bypass |

## Chosen delivery path

| Item | Value |
|------|-------|
| Replan mechanism | Extracted `apply_promotion_pr_recovery_plan(...)` and call it from the initial snapshot and inside the promotion polling loop; `apply_integration_push_recovery_plan(...)` remains initial-snapshot-only |
| Run-ID dedupe | Invocation-scoped `rerun_ids: set[int]` passed into every promotion plan application |
| Absent-context dedupe | Invocation-scoped `dispatched_contexts: set[str]` keyed by required context name |
| Ruleset satisfaction probe | Unchanged GitHub required PR view via `load_required_pr_checks` / `gh pr checks --required --json` and `plan_required_check_recovery` |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra` — merge PR #162):**

- `config/actions-check-recovery-runner.py` — promotion replan during polling with invocation-scoped dedupe
- `.github/workflows/recover-actions-checks.yml` — current-state comment
- `.github/workflows/release.yml` — recovery-step comment
- `README.md` — promotion recovery paragraphs
- `tests/test_voc122_actions_check_recovery.py` — time-evolving #1000 class and fail-closed replan cases

**Caller (`vocanova-platform-sandbox`):**

- `tooling/governance/fixtures/karsift-ai-infra/` — synced to infra merge `60afda3a44fd06b8c00b219771de7112f1aded6e` (runner, recovery workflow comments, release recovery comment, README pin paragraph, VOC-122 tests)
- `tooling/governance/tests/test_voc122_implement_policy.py` — fixture regressions for replan contract and pin
- `tooling/governance/tests/test_voc121_implement_policy.py`, `test_voc124_implement_policy.py`, `test_voc125_implement_policy.py`, `test_voc125_implement_fixture.py`, `test_voc126_workflow_dispatch_input_limit.py` — advanced pin literals
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` — advanced pin literals
- `specs/changes/VOC-122-promotion-recovery-must-replan-required-checks/t00-evidence.md` — this file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | KARSIFT/karsift-ai-infra#162 |
| Exact infra merge SHA | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Pin applicable? | **yes** — runner, recovery workflow comments, release recovery comment, README recovery paragraphs, and VOC-122 tests consumed |
| `PINNED_SHA.txt` | `60afda3a44fd06b8c00b219771de7112f1aded6e` |

## Validation commands

```bash
# Infrastructure (nested source checkout at merge SHA)
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'
python3 -m unittest tests.test_voc122_actions_check_recovery

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
node --test scripts/foundation/voc097-fixture-matrix.test.mjs
node --test scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node --test scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (karsift-ai-infra) | **PASS** — 395 tests |
| `python3 -m unittest tests.test_voc122_actions_check_recovery` | **PASS** — 9 tests |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** — 213 tests |
| `node --test scripts/foundation/voc097-fixture-matrix.test.mjs` | **PASS** |
| `node --test scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | **PASS** |
| `node --test scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | **PASS** |
| `git diff --check` | **PASS** |

## Acceptance mapping

- `VOC-122-AC-00` / `VOC-122-EV-00` — `test_absent_then_cancelled_selected_row_is_rerun_once_and_succeeds`; promotion loop calls `apply_promotion_pr_recovery_plan` on later snapshots
- `VOC-122-AC-01` / `VOC-122-EV-00` — `test_replan_rerun_still_binds_pr_head_branch_event_and_workflow`; `validate_selected_workflow_run` before every rerun
- `VOC-122-AC-02` / `VOC-122-EV-00` — `test_absent_context_dispatch_is_not_repeated_on_later_snapshots`, `test_selected_run_ids_are_not_rerun_again_on_later_snapshots`, `test_dispatch_success_does_not_suppress_later_cancelled_selected_row`
- `VOC-122-AC-03` / `VOC-122-EV-00` — `test_dispatch_success_does_not_suppress_later_cancelled_selected_row`
- `VOC-122-AC-04` / `VOC-122-EV-00` — `test_ambiguous_required_check_run_fails_closed_on_replan`, `test_foreign_required_check_fails_closed_on_replan`, `test_replan_rerun_refuses_second_attempt_selected_run`
- `VOC-122-AC-05` / `VOC-122-EV-00` — `test_timeout_poll_interval_and_integration_push_planning_remain`; `integration_push` planning stays initial-snapshot-only
- `VOC-122-AC-06` / `VOC-122-EV-00` — nine deterministic time-evolving tests in `tests/test_voc122_actions_check_recovery.py`
- `VOC-122-AC-07` / `VOC-122-EV-00` — fixture pin `60afda3a44fd06b8c00b219771de7112f1aded6e`; current-state comments/docs describe replan during polling; `test_voc122_implement_policy.py` and foundation pin literals advanced
