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

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Operator identity | `existing_pr_number` remains on `pipeline.yml`; no free-form SHA inputs on caller `workflow_dispatch` |
| Input-limit repair | relocate the five read-only verifier jobs to a dedicated caller workflow (preferred `pipeline-verify.yml`); do not drop an input; do not pack JSON |
| Mutating actions | `implement`, `plan`, `reconcile`, `reconcile-release`, `reconcile-live-evidence`, `recover-integration-push`, `recover-promotion-pr-checks` stay on `pipeline.yml` |
| Verifier workflow | read-only; no `secrets: inherit`; no `actions: write`; no App-token mint |
| Pin target | this package's exact reviewed infra merge, not `1f1705d…` |
| #1024 | supersede after the governed replacement exists; do not merge |
| VOC-125 #1022 / #1020 | close only when the live caller route is valid; no completion marker bound to #1024 |
| VOC-122 | resume existing #1003 / #1012 after exact infra merge and caller dispatch are live; do not create a replacement task or PR |
| Attempt | VOC-126-T00 first run is attempt `1`; do not dispatch VOC-125 as attempt `3` |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces

To be filled during implementation. Expected:

**Infrastructure (`KARSIFT/karsift-ai-infra`):**

- `templates/project-repo/.github/workflows/pipeline.yml`
- `templates/project-repo/.github/workflows/pipeline-verify.yml` (preferred name)
- `tests/` (maximum-input-count regression and action-options updates)
- `README.md`

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml`
- `.github/workflows/pipeline-verify.yml` (preferred name)
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
| Pin applicable? | expected yes — project-repo `pipeline.yml` is in the fixture |
| Pin must not equal | `1f1705dbad41729563b0ad1e878e4154e5511e93` |
| `PINNED_SHA.txt` after source merge | pending |
| Bootstrap used? | **no** |

## Dependent carriers (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-125-T00 / issue #1022 / origin #1020 | Remaining caller contract is delivered by this package; close only after the live route is valid; do not bind a VOC-125 completion marker to #1024 |
| Caller PR #1024 | Unusable definition; supersede after the governed replacement exists; do not merge |
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; resume through the repaired `pipeline.yml` route; do not merge from this package |
| This package's duty | Repair GitHub-valid dispatch while preserving `existing_pr_number`; record the exact reviewed infra SHA and caller dispatch that #1003 should be resumed against |
| Resume command after T00 is live | `action=implement` on `pipeline.yml`, `change_id=VOC-122`, `task_id=VOC-122-T00`, `issue_number=1003`, `attempt=2`, `existing_pr_number=1012` |
| Create a replacement VOC-122 task or PR? | No |
| Dispatch VOC-125 as attempt `3`? | No |
| Treat VOC-126 merge as VOC-122 completion? | No |

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

- `VOC-126-AC-00` / `VOC-126-EV-00` — pending
- `VOC-126-AC-01` / `VOC-126-EV-00` — pending
- `VOC-126-AC-02` / `VOC-126-EV-00` — pending
- `VOC-126-AC-03` / `VOC-126-EV-00` — pending
- `VOC-126-AC-04` / `VOC-126-EV-00` — pending
- `VOC-126-AC-05` / `VOC-126-EV-00` — pending
- `VOC-126-AC-06` / `VOC-126-EV-00` — pending
