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
| Input-limit repair | relocate the five read-only verifier jobs to dedicated caller `pipeline-verify.yml`; do not drop an input; do not pack JSON |
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

**Infrastructure (`KARSIFT/karsift-ai-infra` — local commit `82fe888d7285270060055fbc4a6a46263906f2f0` for source publication):**

- `templates/project-repo/.github/workflows/pipeline.yml` — removed five verifier actions/inputs/jobs; kept `existing_pr_number`; 16 `workflow_dispatch` inputs
- `templates/project-repo/.github/workflows/pipeline-verify.yml` — new dedicated read-only verifier dispatch (17 inputs)
- `tests/test_voc126_workflow_dispatch_input_limit.py` — maximum-input-count regression
- `tests/test_adoption_handoff.py`, `test_auto_advance_ownership.py`, `test_ready_for_review_reuse.py`, `test_remediation_ownership.py` — relocated verifier contract assertions
- `README.md` — verifier dispatch documents `pipeline-verify.yml`

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml` — added `existing_pr_number`; relocated verifiers; 16 inputs
- `.github/workflows/pipeline-verify.yml` — live read-only verifier dispatch mirror
- `tooling/governance/fixtures/karsift-ai-infra/` — template, tests, README pin paragraph, `PINNED_SHA.txt`, supporting config copies
- `tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py`
- `tooling/governance/tests/test_voc080_*`, `test_voc121_implement_policy.py`, `test_voc124_implement_policy.py`
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs`

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | prepared in local `karsift-ai-infra` commit `82fe888d7285270060055fbc4a6a46263906f2f0` for source publication |
| Independently reviewed infra head SHA | pending merge (target `82fe888d7285270060055fbc4a6a46263906f2f0`) |
| Exact infra merge SHA | pending independent merge |
| Pin applicable? | **yes** — project-repo `pipeline.yml` / `pipeline-verify.yml` and related tests consumed |
| Pin must not equal | `1f1705dbad41729563b0ad1e878e4154e5511e93` |
| `PINNED_SHA.txt` after source merge | `82fe888d7285270060055fbc4a6a46263906f2f0` (awaiting exact reviewed merge confirmation) |
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
# Infrastructure (nested checkout)
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
| `python3 -m unittest discover -s tests -p 'test_*.py'` (karsift-ai-infra) | **pass** |
| `bash scripts/governance/validate-governance.sh` | **pass** |
| `bash scripts/governance/classify-change-risk.sh` | **pass** (path floor only in this environment) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **pass** |
| `node scripts/foundation/voc097-fixture-matrix.test.mjs` | **pass** |
| `node scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | **pass** |
| `node scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | **pass** |
| `git diff --check` | **pass** |

**Input-count spot check (caller):** `pipeline.yml` = 16 inputs; `pipeline-verify.yml` = 17 inputs (both ≤ 25).

## Acceptance mapping

- `VOC-126-AC-00` / `VOC-126-EV-00` — implemented; pending independent exact-SHA verification after merge
- `VOC-126-AC-01` / `VOC-126-EV-00` — implemented (`existing_pr_number` on `pipeline.yml`, forwarded on `implement`)
- `VOC-126-AC-02` / `VOC-126-EV-00` — implemented (five verifiers on `pipeline-verify.yml`; mutating actions on `pipeline.yml`)
- `VOC-126-AC-03` / `VOC-126-EV-00` — implemented (read-only verifier workflow; recovery jobs remain on `pipeline.yml`)
- `VOC-126-AC-04` / `VOC-126-EV-00` — implemented; VOC-125 policy tests unchanged in substance
- `VOC-126-AC-05` / `VOC-126-EV-00` — implemented pending infra PR merge confirmation at `82fe888…`; pin advanced; #1024 not merged
- `VOC-126-AC-06` / `VOC-126-EV-00` — handoff recorded; #1022/#1020/#1024 closure and #1003 resume remain post-merge operator steps

## Post-merge operator steps (not executed in T00)

1. After infra PR merge at reviewed SHA `82fe888d7285270060055fbc4a6a46263906f2f0` (or superseding exact reviewed head if amended), confirm `PINNED_SHA.txt` matches that merge.
2. After caller dispatch is merged and promoted, close #1024 as superseded with audit comment naming the VOC-126 exact SHA.
3. Close #1022 / #1020 only when the live route is valid.
4. Resume existing `VOC-122-T00` / #1003 / #1012 at attempt `2` with `existing_pr_number=1012` via the command above.
