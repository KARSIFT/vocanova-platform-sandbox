# VOC-125-T00 — Evidence

Task: `VOC-125-T00` — Bind existing-carrier resume identity for operator
implement dispatch.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1020)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1020 |
| Blocked adopted task | #1003 (`VOC-122-T00`) |
| Existing caller PR | #1012 — draft; must remain the carrier |
| Pipeline run / job | `32966618512` / `98170418081` |
| Caller develop at dispatch | `f0f7a54283c3527f40544fd236516dc3a5f4dc82` |
| Infrastructure ref at failure | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
| Existing branch/head | `agent/voc-122-voc-122-t00@0b7be8c531be8300d5a1d5534acc83bf4d6a1791` |
| Prior App-signed review base | `e910eb4a21d48bbb5b3e0c30b8ee647d64683dbe` |
| Dispatch | `action=implement`, `attempt=2`, no recovery identity |
| Observed failure | `Create implementation branch` finds the existing remote branch, then cannot establish exact-head remediation binding; stop before model resolution or Composer execution |
| Defect | caller `pipeline.yml` exposes `attempt` but not existing-carrier identity; implement job forwards neither `expected_head_sha` nor `expected_base_sha` |
| Why attempt 1 is forbidden | would violate the one-retry limit and rewrite the existing carrier |
| Why a replacement PR is forbidden | #1003 / #1012 remain the only valid authority and carrier |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Operator identity | `existing_pr_number`; no free-form SHA inputs on caller `workflow_dispatch` |
| Derivation locus | `implement.yml` `Bind existing-carrier recovery identity` before `Create implementation branch` |
| Automatic remediate | keep event-derived SHA forwards; also forward `pr_number` as `existing_pr_number` |
| Attempt | `1` or `2` only; existing-carrier resume is attempt `2` |
| VOC-122 | resume existing #1003 / #1012 after exact infra merge and caller dispatch are live; do not create a replacement task or PR |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra` nested checkout):**

- `.github/workflows/implement.yml` — `existing_pr_number` input; bind step; bound SHAs flow to branch/review/publish
- `.github/workflows/remediate.yml` — retry forwards `existing_pr_number: ${{ inputs.pr_number }}`
- `templates/project-repo/.github/workflows/pipeline.yml` — `existing_pr_number` dispatch + forward
- `config/bind_existing_carrier.py`, `config/bind-existing-carrier-runner.py`
- `tests/test_voc125_existing_carrier.py`, `tests/test_remediate_policy.py`, `tests/test_voc113_actions_check_recovery.py`
- `README.md` — operator resume current-state paragraph

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml` — `existing_pr_number` dispatch + forward
- `tooling/governance/fixtures/karsift-ai-infra/` — mirrored `implement.yml`, bind helpers, `test_voc125_existing_carrier.py`; fixture `remediate.yml` retry line only; `PINNED_SHA.txt` advanced
- `tooling/governance/tests/test_voc125_pipeline_dispatch.py`, `test_voc125_implement_fixture.py`
- `tooling/governance/tests/test_voc121_implement_policy.py`, `test_voc124_implement_policy.py` — pin literal updates
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs`
- `scripts/foundation/voc113-actions-check-recovery.test.mjs`, `voc106-remediate-ownership.test.mjs` — dispatch input bound
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending independent review / merge (nested source publication path) |
| Independently reviewed infra head SHA | pending |
| Exact infra git merge SHA | pending — replace content-addressed pin below when merge is live |
| Pin applicable? | **yes** — `implement.yml`, bind helpers, and tests consumed |
| `PINNED_SHA.txt` after fixture sync | `eda07e3af80b929bd6656f6402df2e745d720149` (content-addressed mirror of VOC-125 fixture subset; advance to exact infra merge SHA on publication) |
| Bootstrap used? | **no** |

## Dependent #1003 / #1012 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; resume through the repaired route; do not merge from this package |
| This package's duty | Repair operator resume identity; record the exact reviewed infra SHA and caller dispatch that #1003 should be resumed against |
| Resume command after T00 is live | `gh workflow run pipeline.yml --repo KARSIFT/vocanova-platform-sandbox --ref develop -f action=implement -f change_id=VOC-122 -f package_path=specs/changes/VOC-122-promotion-recovery-must-replan-required-checks -f task_id=VOC-122-T00 -f issue_number=1003 -f attempt=2 -f existing_pr_number=1012` |
| Create a replacement VOC-122 task or PR? | No |
| Treat VOC-125 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (nested checkout)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (karsift-ai-infra) | **OK** (362 tests) |
| `bash scripts/governance/validate-governance.sh` | **OK** |
| `bash scripts/governance/classify-change-risk.sh` | **OK** (path floor only; no PR risk declaration in working tree) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **OK** (199 tests) |
| `git diff --check` | **OK** |

Targeted:

```bash
python3 -m unittest tests.test_voc125_existing_carrier -v
python3 -m unittest tooling.governance.tests.test_voc125_pipeline_dispatch -v
python3 -m unittest tooling.governance.tests.test_voc125_implement_fixture -v
```

## Acceptance mapping

- `VOC-125-AC-00` / `VOC-125-EV-00` — implemented; deterministic tests + dispatch contract
- `VOC-125-AC-01` / `VOC-125-EV-00` — implemented; bind reuses carrier; no replacement PR path
- `VOC-125-AC-02` / `VOC-125-EV-00` — implemented; attempt `1`/`2` only; attempt-1 existing carrier fails closed
- `VOC-125-AC-03` / `VOC-125-EV-00` — implemented; mismatch matrix in `test_voc125_existing_carrier.py`
- `VOC-125-AC-04` / `VOC-125-EV-00` — implemented; remediate retry forwards SHAs + `existing_pr_number`
- `VOC-125-AC-05` / `VOC-125-EV-00` — preserved; existing VOC-121/123/124 policy tests still pass
- `VOC-125-AC-06` / `VOC-125-EV-00` — docs/comments updated; pin recorded; #1003/#1012 handoff recorded above
