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

**Infrastructure (`KARSIFT/karsift-ai-infra`) — nested source carrier only until merge:**

- `config/actions-check-recovery-runner.py` — promotion replan during polling with invocation-scoped dedupe
- `.github/workflows/recover-actions-checks.yml` — current-state comment
- `.github/workflows/release.yml` — recovery-step comment
- `README.md` — promotion recovery paragraphs
- `tests/test_voc122_actions_check_recovery.py` — time-evolving #1000 class and fail-closed replan cases

**Caller (`vocanova-platform-sandbox`) — this attempt:**

- `specs/changes/VOC-122-promotion-recovery-must-replan-required-checks/t00-evidence.md` — this file
- caller fixture files intentionally **unchanged** at pin `20dcf340fa73a36ebc6074442fde79530dfa5871` until the reviewed infrastructure merge lands

Attempt 1 incorrectly copied infrastructure changes into the pinned fixture before
the coordinated infra merge. Attempt 2 reverted that and only updated evidence.
This remediation keeps the fixture byte-for-byte aligned with the current pin and
publishes the VOC-122 source only through the nested `karsift-ai-infra/` checkout for
`publish-source`.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | opened/updated by `publish-source` from the nested source carrier on this attempt |
| Independently reviewed infra head SHA | pending independent review of the carrier branch |
| Exact infra merge SHA | pending merge of the coordinated infrastructure PR |
| Pin applicable? | **no until merge** — fixture must remain at `20dcf340fa73a36ebc6074442fde79530dfa5871` until the reviewed infrastructure merge is known |
| `PINNED_SHA.txt` after source merge | unchanged at `20dcf340fa73a36ebc6074442fde79530dfa5871`; advance to the exact reviewed infra merge SHA in a follow-up caller change after merge when syncing runner, tests, and recovery comments |

## Validation commands

```bash
# Infrastructure (nested source checkout)
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*.py'
python3 -m unittest tests.test_voc122_actions_check_recovery

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (karsift-ai-infra) | **PASS** — 395 tests |
| `python3 -m unittest tests.test_voc122_actions_check_recovery` | **PASS** — 9 tests |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** (path floor R1 for untracked nested carrier; semantic R4 remains for recovery mutation) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** — 208 tests |
| `git diff --check` | **PASS** |

## Acceptance mapping

- `VOC-122-AC-00` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-01` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-02` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-03` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-04` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-05` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-06` / `VOC-122-EV-00` — implemented in nested carrier; caller fixture pin pending infra merge
- `VOC-122-AC-07` / `VOC-122-EV-00` — infra docs/comments updated in nested carrier; caller fixture/docs pin pending infra merge
