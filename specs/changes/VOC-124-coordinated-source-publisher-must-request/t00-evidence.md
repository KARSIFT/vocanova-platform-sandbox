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
| Caller PR | #1012 — intentionally draft; must remain unmerged until the authoritative source PR merges |
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
| Attempt | 2 of 2 — bootstrap infra PR merged; caller fixture pin and evidence updated on this revision |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`, bootstrap carrier PR #159):**

- `.github/workflows/implement.yml` — VOC-124 current-state comment; `permission-workflows: write` on the `publish-source` mint only; caller `publish` PR body no longer claims a standing human-approval merge gate under active A-004
- `README.md` — source-carrier paragraph describes infrastructure `workflows: write` while caller `publish` still omits it and still refuses caller workflow files
- `tests/test_voc124_workflow_permissions.py` — mint isolation, workflow-file source coverage, A-004 PR-body text, README/CHANGELOG preservation
- `tests/test_voc121_source_carrier_publisher.py` — authorized `.github/workflows/**` source bundle publishes without caller-style workflow-file rejection
- `tests/test_voc121_implement_policy.py` — `publish-source` mint includes `permission-workflows: write`
- `tests/test_live_evidence_reconcile.py` — caller `publish_job` assertions isolated from `publish-source`

**Caller (`vocanova-platform-sandbox`, attempt 2):**

- `tooling/governance/fixtures/karsift-ai-infra/` — mirrored `implement.yml`, VOC-121/124 tests, `CHANGELOG.md` (read-only audit evidence for VOC-124 test), and fixture README pin paragraph advanced to bootstrap merge `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`
- `tooling/governance/tests/test_voc121_implement_policy.py` — pin advanced; `publish-source` mint regression added
- `tooling/governance/tests/test_voc124_implement_policy.py` — fixture regressions for mint isolation, A-004 PR-body text, and VOC-124 current-state comment
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` — pin literals advanced to `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`
- this evidence file

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | https://github.com/KARSIFT/karsift-ai-infra/pull/159 — merged |
| Bootstrap branch | `agent/voc-124-bootstrap` |
| Review base SHA | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` (infra `main` before bootstrap) |
| Independently reviewed infra head SHA | `7d9e72e3149676e004cc36553664686f226080a7` |
| Exact infra merge SHA | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
| Separate merger | `m-e-h-r-d-a-a-d` (merged `2026-08-26`; not the `implementer` model role) |
| `VOC-124-D04` bootstrap status | **exhausted** — one clean branch from infra `main`, one reviewed non-closing infra PR, one merge; caller fixture/pin/evidence resumed only after merge `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` was live on `implement.yml@main` |
| VOC-122 nested head `f90eb63…` published by bootstrap? | **no** |
| Pin applicable? | **yes** — `implement.yml` and VOC-124 tests are in the policy fixture subset |
| `PINNED_SHA.txt` after source merge | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` |
| Fixture/source comparison | **PASS** — `implement.yml`, `test_voc124_workflow_permissions.py`, `test_voc121_source_carrier_publisher.py`, `test_voc121_implement_policy.py`, and `test_live_evidence_reconcile.py` are byte-for-byte identical between infra `main` and the pinned fixture |

## Independent source review

Bootstrap infrastructure PR #159 was reviewed and merged before attempt 2 resumed
caller fixture/pin work. The governed `implementer` model did not approve or merge
that carrier.

## Dependent #1003 / #1012 (not implemented by this task)

| Item | Value |
|------|-------|
| VOC-122-T00 / issue #1003 | Distinct already-authorized promotion-recovery outcome |
| Caller PR #1012 | Existing VOC-122 draft carrier; must remain unmerged until the authoritative VOC-122 source PR merges |
| This package's duty | Repair publisher token permission; exhaust bootstrap; record the exact reviewed infra SHA that #1003 should be re-dispatched or reconciled against |
| Re-dispatch against | `f406cc95a3f853e8aef5bf8bcf22d37a29d64547` (infra `main` after VOC-124 bootstrap merge) |
| Reconstruct `f90eb630743c8c523e2e6e8dff017acbb31a7f43` by hand? | No |
| Treat VOC-124 merge as VOC-122 completion? | No |
| VOC-122 retry handoff | Existing `VOC-122-T00` / #1003 should be re-dispatched or reconciled through the repaired `implement.yml@main` path against `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`; #1012 remains out of scope until VOC-122's authoritative source PR merges |

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
| `python3 -m unittest discover -s tests -p 'test_*.py'` (nested `karsift-ai-infra/`) | **PASS** (344 tests) |
| `bash scripts/governance/validate-governance.sh` | **PASS** |
| `bash scripts/governance/classify-change-risk.sh` | **PASS** — R4 floor |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **PASS** (192 tests) |
| `git diff --check` | **PASS** |

Targeted VOC-124 tests:

```bash
python3 -m unittest tests.test_voc124_workflow_permissions
python3 -m unittest tests.test_voc121_source_carrier_publisher.Voc121SourceCarrierPublisherTests.test_workflow_file_source_bundle_is_not_rejected_by_source_publisher_checks
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc124*.py'
node --test scripts/foundation/voc097-fixture-matrix.test.mjs scripts/foundation/voc104-ready-for-review-reuse.test.mjs scripts/foundation/voc108-authoritative-lifecycle.test.mjs
```

| Command | Result |
|---------|--------|
| `python3 -m unittest tests.test_voc124_workflow_permissions` | **PASS** (7 tests) |
| workflow-file source bundle publisher test | **PASS** |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc124*.py'` | **PASS** (6 tests) |
| foundation pin tests (`voc097`, `voc104`, `voc108`) | **PASS** (16 tests) |

## Acceptance mapping

- `VOC-124-AC-00` / `VOC-124-EV-00` — **pass**: `publish-source` mint requests `permission-workflows: write` on infra `main` and in the pinned fixture
- `VOC-124-AC-01` / `VOC-124-EV-00` — **pass**: caller `publish` still omits that permission and still refuses workflow files
- `VOC-124-AC-02` / `VOC-124-EV-00` — **pass**: authorized `.github/workflows/**` source bundle covered by required permission
- `VOC-124-AC-03` / `VOC-124-EV-00` — **pass**: missing credentials, invalid bundles, stale bases/leases fail closed
- `VOC-124-AC-04` / `VOC-124-EV-00` — **pass**: VOC-121/VOC-123 isolation, named-ref bundle, lease, retry limits preserved
- `VOC-124-AC-05` / `VOC-124-EV-00` — **pass**: A-004 current-state text corrected; historical CHANGELOG unchanged
- `VOC-124-AC-06` / `VOC-124-EV-00` — **pass**: bootstrap exhausted; fixture pin advanced to `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`; #1003 retry handoff recorded against that SHA
