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
| Infrastructure ref | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
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
| Derivation locus | `implement.yml` before `Create implementation branch` |
| Automatic remediate | keep event-derived SHA forwards; also forward `pr_number` as `existing_pr_number` |
| Attempt | `1` or `2` only; existing-carrier resume is attempt `2` |
| VOC-122 | resume existing #1003 / #1012 after exact infra merge and caller dispatch are live; do not create a replacement task or PR |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces

To be filled during implementation. Expected:

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `.github/workflows/implement.yml`
- `.github/workflows/remediate.yml`
- `templates/project-repo/.github/workflows/pipeline.yml`
- `config/` bind helper and/or `verify-expected-head.py`
- `tests/`
- `README.md`

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml`
- `tooling/governance/fixtures/karsift-ai-infra/` when consumed
- `tooling/governance/tests/`
- foundation pin tests if pin literals change
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending implementation |
| Independently reviewed infra head SHA | pending |
| Exact infra merge SHA | pending |
| Pin applicable? | expected yes — `implement.yml` is in the policy fixture subset |
| `PINNED_SHA.txt` after source merge | pending |
| Bootstrap used? | **no** |

## Dependent #1003 / #1012 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; resume through the repaired route; do not merge from this package |
| This package's duty | Repair operator resume identity; record the exact reviewed infra SHA and caller dispatch that #1003 should be resumed against |
| Resume command after T00 is live | `action=implement`, `change_id=VOC-122`, `task_id=VOC-122-T00`, `issue_number=1003`, `attempt=2`, `existing_pr_number=1012` |
| Create a replacement VOC-122 task or PR? | No |
| Treat VOC-125 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (nested checkout at infra main)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` | pending implementation |
| `bash scripts/governance/validate-governance.sh` | pending implementation |
| `bash scripts/governance/classify-change-risk.sh` | pending implementation |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pending implementation |
| `git diff --check` | pending implementation |

## Acceptance mapping

- `VOC-125-AC-00` / `VOC-125-EV-00` — pending
- `VOC-125-AC-01` / `VOC-125-EV-00` — pending
- `VOC-125-AC-02` / `VOC-125-EV-00` — pending
- `VOC-125-AC-03` / `VOC-125-EV-00` — pending
- `VOC-125-AC-04` / `VOC-125-EV-00` — pending
- `VOC-125-AC-05` / `VOC-125-EV-00` — pending
- `VOC-125-AC-06` / `VOC-125-EV-00` — pending
