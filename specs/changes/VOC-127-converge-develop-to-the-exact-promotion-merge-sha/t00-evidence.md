# VOC-127-T00 — Evidence

Task: `VOC-127-T00` — Advance develop to the exact promotion merge SHA and add
exceptional main-only reconciliation.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1035)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1035 |
| Pre-repair `develop` | `883c4a544f24fb9840694a834ebe9c665d4160b5` |
| Pre-repair `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` (PR #1033 merge) |
| Ahead/behind | `git rev-list --left-right --count origin/develop...origin/main` → `0 46` |
| Ancestor | `git merge-base --is-ancestor origin/develop origin/main` succeeded |
| Tree diff | exactly the two VOC-113-T01 evidence files from main-targeting PR #955 |
| Active work at incident | no PR, issue, or workflow |
| Root cause | `release.yml` preserves/recreates the integration ref at `CHECKED_HEAD_SHA` and never binds `mergeCommit.oid` |
| Secondary close path | `ahead_by == 0` closes the audit as "Already promoted" while SHAs still differ |
| Already-merged no-op | terminal promotion PR is treated as successful without sync |
| Policy test that encodes the bug | `tests/test_release_policy.py::test_promotion_preserves_long_lived_integration_branch` required `-f sha="$CHECKED_HEAD_SHA"` |
| 2026-08-27 repair | bounded audited settings operation recreated `develop` at exact `main`; both refs now `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`; compare zero; trees identical |
| Repair audit | release issue #1032 |
| Repair reusable as implementation? | **no** — historical settings evidence, not this task's authority |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`); refs are already equal |
| Bootstrap required? | **no** — VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Ordinary sync | serialized `release.yml` converge job after the single `--merge`, before audit close |
| Merge commit | preserved; do not fast-forward `main` instead |
| Bind | checked head, checked production base, merged PR identity, live tips |
| Fail closed | moved `main`, unexpected `develop`, malformed merge, unique develop commits; never erase |
| Recovery | `reconcile-release` including already-merged PR + open audit |
| Fully promoted | live `develop` SHA equals live `main` SHA, not merely `ahead_by == 0` |
| Release loop | no second promotion PR; still exactly one `gh pr merge` |
| Missing develop | recreate at merge SHA, not `CHECKED_HEAD_SHA` |
| Staging | tree-equivalent sync must not deploy; retain VOC-111 path selection |
| Exceptional main-only | adopted package/task + merged main-targeting PR number via `reconcile-production-change`; not operator SHAs; not scheduled; not ordinary |
| `existing_pr_number` | remains implement-only |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Pin target | this package's exact reviewed infra merge, not `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces

**Infrastructure (`KARSIFT/karsift-ai-infra`, merged PR #163):**

- `.github/workflows/release.yml` — bind-and-sync `develop` to `mergeCommit.oid` before audit close; already-merged recovery; stop `ahead_by == 0` unequal-SHA short-circuit; recreate missing ref at merge SHA
- `.github/workflows/reconcile-production-change.yml` — exceptional governed main-only reconciliation
- `config/branch_sync.py`, `config/branch-sync-runner.py` — testable exact-SHA policy and lease-protected apply
- `tests/test_branch_sync.py`, `tests/test_branch_sync_runner.py`, `tests/test_release_policy.py`
- `templates/project-repo/.github/workflows/pipeline.yml` — `reconcile-production-change` dispatch and release/auto-advance ordering
- `README.md` — current-state exact-merge-SHA sync and exceptional path

**Caller (`vocanova-platform-sandbox`):**

- `.github/workflows/pipeline.yml` — `reconcile-production-change` action, job, and release/auto-advance ordering
- `tooling/governance/fixtures/karsift-ai-infra/**` — synced to reviewed infra merge
- `tooling/governance/tests/test_voc127_implement_policy.py` — VOC-127 fixture and caller regressions
- `tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py` — updated mutating action list and pin
- `scripts/foundation/voc111-deploy-staging-paths.test.mjs` — tree-equivalent empty-diff case
- `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` — pin literals
- `AGENTS.md`, `docs/operations/10-development-workflow.md`, `docs/operations/11-devops-and-ci-cd.md`, `docs/operations/15-ai-native-product-and-engineering-operating-model.md`, `docs/governance/16-autonomous-development-operating-model.md`

**Staging selector:** GitHub `on.push.paths` already skips pushes with no changed files. A tree-equivalent post-promotion develop fast-forward produces an empty diff, so no job-level skip was added to `deploy-staging.yml`.

## Shared-infra carrier and fixture pin

| Item | Value |
|------|-------|
| Coordinated infra PR | KARSIFT/karsift-ai-infra#163 (`Relates to KARSIFT/vocanova-platform-sandbox#1039`) |
| Exact infra merge SHA | `a9df74a63976d5239b84151fd01310835c999e7c` |
| Pin applicable? | **yes** — `release.yml`, helpers, tests, `reconcile-production-change.yml`, template pipeline, and current-state comments |
| Pin must not equal | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| `PINNED_SHA.txt` | `a9df74a63976d5239b84151fd01310835c999e7c` |
| Bootstrap used? | **no** |

## Validation commands

```bash
# Infrastructure (nested checkout at merged SHA)
cd karsift-ai-infra && python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
node scripts/foundation/voc111-deploy-staging-paths.test.mjs
node scripts/foundation/voc097-fixture-matrix.test.mjs
git diff --check
```

### Validation results (attempt 1)

Recorded after implementation; exact pass/fail output is in the CI run for this carrier.

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tests -p 'test_*.py'` (infra checkout `a9df74a…`) | **pass** — 421 tests OK |
| `bash scripts/governance/validate-governance.sh` | **pass** |
| `bash scripts/governance/classify-change-risk.sh` | **pass** (path floor R4) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **pass** |
| `node scripts/foundation/voc111-deploy-staging-paths.test.mjs` | **pass** — 10 tests |
| `node scripts/foundation/voc097-fixture-matrix.test.mjs` | **pass** — 5 tests |
| `git diff --check` | **pass** |

## Acceptance mapping

- `VOC-127-AC-00` / `VOC-127-EV-00` — `release.yml` + `branch_sync` helper advance integration to merge SHA before audit close
- `VOC-127-AC-01` / `VOC-127-EV-00` — fail-closed races and unique commits in `branch_sync.py` / tests
- `VOC-127-AC-02` / `VOC-127-EV-00` — already-merged recovery and single `gh pr merge` in `release.yml`
- `VOC-127-AC-03` / `VOC-127-EV-00` — missing ref create path uses merge SHA in runner
- `VOC-127-AC-04` / `VOC-127-EV-00` — VOC-111 path filter + empty-diff test; no deploy-staging.yml broadening
- `VOC-127-AC-05` / `VOC-127-EV-00` — `reconcile-production-change` on caller/template pipeline; no operator SHAs
- `VOC-127-AC-06` / `VOC-127-EV-00` — pin `a9df74a…`, docs updated, prior controls preserved
