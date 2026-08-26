# VOC-124-T00 — Evidence

Task: `VOC-124-T00` — Request workflow-write on the coordinated source publisher
token.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1013)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1013 |
| Blocked adopted task | #1003 (`VOC-122-T00`) |
| Pipeline run / job | `32958526215` / `98147443377` |
| Nested bundle head | `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| Infrastructure base | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Bundle download / verify / token mint | succeeded |
| Error | `refusing to allow a GitHub App to create or update workflow ... without workflows permission` |
| Changed workflow file | `.github/workflows/recover-actions-checks.yml` |
| Caller PR | #1012 — intentionally draft; staged mirror/evidence; still pins the prior infrastructure SHA; must remain unmerged until the authoritative source PR merges |
| Defect | `publish-source` mint requests `permission-contents`, `permission-issues`, and `permission-pull-requests`, and omits `permission-workflows: write` |
| Installation | `karsift-ai-infra-bot` installation `148001476` already has repository `workflows: write` |
| Why a VOC-122 retry cannot self-repair | caller executes `implement.yml@main`; the same mint would reject its own workflow-file fix |
| Why bootstrap is required | the repair itself changes `.github/workflows/implement.yml` and must land through a bounded supervised infra PR, then be exhausted |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Token change | `permission-workflows: write` on the `publish-source` mint only |
| Caller publisher | keep omitting `permission-workflows`; keep rejecting `.github/workflows/**` |
| Bootstrap | `VOC-124-D04` one-time supervised infra PR from current `main`; no PATH/Git interception; no direct `main` push; no VOC-122 bundle publication |
| VOC-122 | retry existing #1003 / #1012 after the exact infra merge is live; do not create a replacement task or PR; do not hand-push `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| A-004 text | correct `implement.yml` caller PR body / current-state comments; do not rewrite historical CHANGELOG |
| Secrets | do not print credential values; do not rotate App secrets; do not change installation permissions |
| Attempt | 1 of 2 — infrastructure source repair and deterministic tests prepared; bootstrap infra PR and caller fixture pin remain pending |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`, bootstrap carrier — attempt 1 source):**

- `.github/workflows/implement.yml` — VOC-124 current-state comment; `permission-workflows: write` on the `publish-source` mint only; caller `publish` PR body no longer claims a standing human-approval merge gate under active A-004
- `README.md` — source-carrier paragraph describes infrastructure `workflows: write` while caller `publish` still omits it and still refuses caller workflow files
- `tests/test_voc124_workflow_permissions.py` — mint isolation, workflow-file source coverage, A-004 PR-body text, README/CHANGELOG preservation
- `tests/test_voc121_source_carrier_publisher.py` — authorized `.github/workflows/**` source bundle publishes without caller-style workflow-file rejection
- `tests/test_voc121_implement_policy.py` — `publish-source` mint includes `permission-workflows: write`
- `tests/test_live_evidence_reconcile.py` — caller `publish_job` assertions isolated from `publish-source`

**Caller (`vocanova-platform-sandbox`, attempt 1):**

- `tooling/governance/tests/test_voc124_implement_policy.py` — nested-checkout regressions; fixture pin intentionally unchanged until bootstrap merge
- this evidence file

**Not changed on attempt 1 (pending bootstrap merge):**

- `tooling/governance/fixtures/karsift-ai-infra/` mirror and `PINNED_SHA.txt` — remain at reviewed merge `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` until the independently reviewed bootstrap infra PR merges and attempt 2 (or the resumed carrier) syncs the exact merge SHA
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` pin literals — unchanged until fixture pin advances

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | pending — bootstrap branch to open from infra `main` with attempt-1 source |
| Bootstrap branch | pending implementation handoff (`agent/voc-124-voc-124-t00` or equivalent) |
| Review base SHA | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` (current infra `main`) |
| Independently reviewed infra head SHA | pending bootstrap PR |
| Exact infra merge SHA | pending bootstrap merge |
| Separate merger | pending |
| `VOC-124-D04` bootstrap status | **prepared on attempt 1** — source repair and tests ready; not yet opened/merged/exhausted |
| VOC-122 nested head `f90eb63…` published by bootstrap? | **must remain no** |
| Pin applicable? | **yes, when bootstrap merges** — `implement.yml` and VOC-124 tests are in the policy fixture subset |
| `PINNED_SHA.txt` after source merge | pending exact bootstrap merge SHA |

## Independent source review

To be recorded when the bootstrap infrastructure PR is opened. Bind review to
the exact final infrastructure head. The implementer must not approve or merge
that carrier.

## Dependent #1003 / #1012 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; must remain unmerged until the authoritative VOC-122 source PR merges |
| This package's duty | Repair publisher token permission; exhaust bootstrap; record the exact reviewed infra SHA that #1003 should be re-dispatched or reconciled against |
| Re-dispatch against | pending exact VOC-124 infra merge SHA |
| Reconstruct `f90eb630743c8c523e2e6e8dff017acbb31a7f43` by hand? | No |
| Treat VOC-124 merge as VOC-122 completion? | No |

## Validation commands

```bash
# Infrastructure (nested checkout / bootstrap source)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (nested `karsift-ai-infra/`) | **PASS** (344 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** — R4 floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (191 tests) |
| `git diff --check` | **PASS** |

Targeted VOC-124 tests:

```bash
python3 -m unittest tests.test_voc124_workflow_permissions
python3 -m unittest tests.test_voc121_source_carrier_publisher.Voc121SourceCarrierPublisherTests.test_workflow_file_source_bundle_is_not_rejected_by_source_publisher_checks
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc124*.py'
```

| Command | Result |
|---------|--------|
| `python3 -m unittest tests.test_voc124_workflow_permissions` | **PASS** (7 tests) |
| workflow-file source bundle publisher test | **PASS** |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc124*.py'` | **PASS** (6 tests) |

Foundation pin tests were not rerun on attempt 1 because `PINNED_SHA.txt` is
intentionally unchanged until bootstrap merge.

## Acceptance mapping

- `VOC-124-AC-00` / `VOC-124-EV-00` — **pass on bootstrap source**: `publish-source` mint requests `permission-workflows: write`
- `VOC-124-AC-01` / `VOC-124-EV-00` — **pass on bootstrap source**: caller `publish` still omits that permission and still refuses workflow files
- `VOC-124-AC-02` / `VOC-124-EV-00` — **pass on bootstrap source**: authorized `.github/workflows/**` source bundle covered by required permission
- `VOC-124-AC-03` / `VOC-124-EV-00` — **pass on bootstrap source**: missing credentials, invalid bundles, stale bases/leases fail closed
- `VOC-124-AC-04` / `VOC-124-EV-00` — **pass on bootstrap source**: VOC-121/VOC-123 isolation, named-ref bundle, lease, retry limits preserved
- `VOC-124-AC-05` / `VOC-124-EV-00` — **pass on bootstrap source**: A-004 current-state text corrected; historical CHANGELOG unchanged
- `VOC-124-AC-06` / `VOC-124-EV-00` — **pending attempt 2 / post-bootstrap**: bootstrap exhaustion, fixture pin, foundation pin literals, and #1003 re-dispatch against exact infra merge SHA
