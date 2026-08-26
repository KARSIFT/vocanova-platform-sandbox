# VOC-126-T00 — Evidence

Task: `VOC-126-T00` — Relocate read-only verifier dispatch so VOC-125
`existing_pr_number` can land under GitHub's 25-input limit.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1025)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1025 |
| Blocked adopted VOC-125 task | #1022 (`VOC-125-T00`); origin #1020 |
| Unusable caller PR | #1024 at `8621f12dd466edab37fddb86d4e5e0a348ed3609` |
| Infrastructure merge that encoded the invalid template | `1f1705dbad41729563b0ad1e878e4154e5511e93` |
| Actions run | `32977045898` — terminates with no jobs or logs |
| Authenticated annotation | `Invalid workflow file: .github/workflows/pipeline.yml#L1 — you may only define up to 25 inputs for a workflow_dispatch event` |
| Retrigger | draft/ready retriggering on the same SHA produces no pipeline run |
| VOC-125 template input count | 26, including `existing_pr_number` |
| Live caller input count at drafting | 25; `existing_pr_number` not yet present |
| Why local suites missed it | they assert the new input exists and is forwarded, not GitHub's maximum input count |
| Why VOC-125 cannot retry | final allowed implementation retry already consumed remediating the source carrier; attempt `3` is forbidden |
| Why #1024 cannot be the carrier | GitHub rejects the workflow definition before any job, so exact-revision pipeline review cannot start |
| Downstream blocked carrier | VOC-122-T00 #1003 / draft PR #1012 |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` |

## Attempt history and supervised carrier reconciliation

| Item | Value |
|------|-------|
| Attempt 1 carrier | `103245d20ae58d309a1d2a73da3c8d80427d8ff5`, based on `42c79d44e43e6d99e7dbe5e07f50aa0d2e8bd76e` |
| Exact-SHA review | KARSIFT review run `33018539110` failed attempt 1 for the provisional pin and incomplete VOC-125 carry-forward |
| Attempt 2 carrier | KARSIFT remediation run `33018539110`, artifact `implement-VOC-126-VOC-126-T00-attempt2`, exact artifact head `4d389fa044f864c704eb774c8f53f9c5ce4edbe5` |
| Attempt 2 result | Composer advanced the pin to `20dcf340fa73a36ebc6074442fde79530dfa5871`, synchronized the mirrored source subset, and passed the governed pre-push validation after its single self-correction pass |
| Residual composition | The bounded retry still omitted caller-only `test_voc125_implement_fixture.py`, `test_voc125_pipeline_dispatch.py`, and finalized VOC-125 delivery evidence from unmerged #1024. Those exact governed carrier assets were composed onto the attempt-2 artifact, with only the new pin and GitHub's 25-input bound updated. |
| Retry limit | attempt `2` is final; no attempt `3` dispatched |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Operator identity | `existing_pr_number` remains on `pipeline.yml`; no free-form SHA inputs on caller `workflow_dispatch` |
| Input-limit repair | relocate the five read-only verifier jobs to dedicated caller `pipeline-verify.yml`; do not drop an input; do not pack JSON |
| Mutating actions | `implement`, `plan`, `reconcile`, `reconcile-release`, `reconcile-live-evidence`, `recover-integration-push`, `recover-promotion-pr-checks` stay on `pipeline.yml` |
| Verifier workflow | read-only; no `secrets: inherit`; no `actions: write`; no App-token mint |
| Pin target | this package's exact reviewed infra merge, not `1f1705d…` |
| #1024 | supersede after the governed replacement exists; do not merge |
| VOC-125 #1022 / #1020 | close only when the live caller route is valid; no completion marker bound to #1024 |
| VOC-122 | resume existing #1003 / #1012 after exact infra merge and caller dispatch are live; do not create a replacement task or PR |
| Attempt | VOC-126-T00 attempt `2` on this carrier; do not dispatch VOC-125 as attempt `3` |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra` — directly owner-verified merge `20dcf340fa73a36ebc6074442fde79530dfa5871`, PR #161):**

- `templates/project-repo/.github/workflows/pipeline.yml` — removed five verifier actions/inputs/jobs; kept `existing_pr_number`; 16 `workflow_dispatch` inputs
- `templates/project-repo/.github/workflows/pipeline-verify.yml` — new dedicated read-only verifier dispatch (17 inputs)
- `tests/test_voc126_workflow_dispatch_input_limit.py` — maximum-input-count regression
- `tests/test_adoption_handoff.py`, `test_auto_advance_ownership.py`, `test_ready_for_review_reuse.py`, `test_remediation_ownership.py` — relocated verifier contract assertions
- `README.md` — verifier dispatch documents `pipeline-verify.yml`; operator resume documents `existing_pr_number` on `pipeline.yml`

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml` — added `existing_pr_number`; relocated verifiers; 16 inputs
- `.github/workflows/pipeline-verify.yml` — live read-only verifier dispatch mirror
- `tooling/governance/fixtures/karsift-ai-infra/` — synced to infra merge `20dcf340fa73a36ebc6074442fde79530dfa5871`, including VOC-125 `implement.yml` / `remediate.yml`, bind helpers, and `test_voc125_existing_carrier.py`
- `tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py`
- `tooling/governance/tests/test_voc125_implement_policy.py`, `test_voc125_implement_fixture.py`, `test_voc125_pipeline_dispatch.py`
- `tooling/governance/tests/test_voc080_*`, `test_voc121_implement_policy.py`, `test_voc124_implement_policy.py`
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs`

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | KARSIFT/karsift-ai-infra#161 |
| Directly owner-verified infra head SHA | `9db35bd338313ab3ab8b7903a6c0522383423fa4` |
| Exact infra merge SHA | `20dcf340fa73a36ebc6074442fde79530dfa5871` |
| Pin applicable? | **yes** — project-repo `pipeline.yml` / `pipeline-verify.yml`, VOC-125 implement/remediate surface, bind helpers, and related tests consumed |
| Pin must not equal | `1f1705dbad41729563b0ad1e878e4154e5511e93` |
| `PINNED_SHA.txt` | `20dcf340fa73a36ebc6074442fde79530dfa5871` |
| Bootstrap used? | **no** |

## Dependent carriers (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-125-T00 / issue #1022 / origin #1020 | Remaining caller contract is delivered by this package; close only after the live route is valid, reviewed, merged, and promoted; do not bind a VOC-125 completion marker to #1024 |
| Caller PR #1024 | Unusable definition; supersede after the governed replacement exists and is reconciled; do not merge |
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; resume through the repaired `pipeline.yml` route; do not merge from this package |
| This package's duty | Repair GitHub-valid dispatch while preserving `existing_pr_number`; record the exact reviewed infra SHA and caller dispatch that #1003 should be resumed against |
| Resume command after T00 is live | `gh workflow run pipeline.yml --repo KARSIFT/vocanova-platform-sandbox --ref develop -f action=implement -f change_id=VOC-122 -f package_path=specs/changes/VOC-122-promotion-recovery-must-replan-required-checks -f task_id=VOC-122-T00 -f issue_number=1003 -f attempt=2 -f existing_pr_number=1012` |
| Create a replacement VOC-122 task or PR? | No |
| Dispatch VOC-125 as attempt `3`? | No |
| Treat VOC-126 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (nested checkout at merged SHA)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (karsift-ai-infra source head `9db35bd…`) | **pass** (386 tests; direct owner verification) |
| KARSIFT attempt-2 governed pre-push validation | **pass** after one self-correction |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'` | **pass** (343 tests) |
| `bash scripts/governance/validate-governance.sh` | **pass** |
| `bash scripts/governance/classify-change-risk.sh` | **pass** (R4 path floor) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **pass** (208 tests) |
| `node scripts/foundation/voc097-fixture-matrix.test.mjs` | **pass** (5 tests) |
| `node scripts/foundation/voc102-auto-advance-ownership.test.mjs` | **pass** (5 tests) |
| `node scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | **pass** (6 tests) |
| `node scripts/foundation/voc106-remediate-ownership.test.mjs` | **pass** (3 tests) |
| `node scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | **pass** (5 tests) |
| `node scripts/foundation/voc113-actions-check-recovery.test.mjs` | **pass** (5 tests) |
| `git diff --check` | **pass** |

**Input-count spot check (caller):** `pipeline.yml` = 16 inputs; `pipeline-verify.yml` = 17 inputs (both ≤ 25).

## Acceptance mapping

- `VOC-126-AC-00` / `VOC-126-EV-00` — implemented; pending independent exact-SHA verification of the final caller head
- `VOC-126-AC-01` / `VOC-126-EV-00` — implemented (`existing_pr_number` on `pipeline.yml`, forwarded on `implement`)
- `VOC-126-AC-02` / `VOC-126-EV-00` — implemented (five verifiers on `pipeline-verify.yml`; mutating actions on `pipeline.yml`)
- `VOC-126-AC-03` / `VOC-126-EV-00` — implemented (read-only verifier workflow; recovery jobs remain on `pipeline.yml`)
- `VOC-126-AC-04` / `VOC-126-EV-00` — implemented; fixture carries VOC-125 `implement.yml` / `remediate.yml` / bind helpers; caller `test_voc125_implement_policy.py` added
- `VOC-126-AC-05` / `VOC-126-EV-00` — implemented; pin equals directly owner-verified infra merge `20dcf340fa73a36ebc6074442fde79530dfa5871`; #1024 not merged
- `VOC-126-AC-06` / `VOC-126-EV-00` — handoff recorded; #1022/#1020/#1024 closure and #1003 resume remain post-merge operator steps

## Post-merge operator steps (not executed in T00)

1. After caller dispatch is merged and promoted, close #1024 as superseded with audit comment naming the VOC-126 exact SHA.
2. Close #1022 / #1020 only when the live route is valid.
3. Resume existing `VOC-122-T00` / #1003 / #1012 at attempt `2` with `existing_pr_number=1012` via the command above.
