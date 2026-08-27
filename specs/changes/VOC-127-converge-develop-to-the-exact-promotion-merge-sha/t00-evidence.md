# VOC-127-T00 — Evidence

Task: `VOC-127-T00` — Advance develop to the exact promotion merge SHA and add
exceptional main-only reconciliation.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

This file is a planning-time evidence contract. Implementation results are
recorded here only after the adopted task runs. Nothing below adopts or
authorizes this package.

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
| Policy test that encodes the bug | `tests/test_release_policy.py::test_promotion_preserves_long_lived_integration_branch` requires `-f sha="$CHECKED_HEAD_SHA"` |
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
| Exceptional main-only | adopted package/task + merged main-targeting PR number; not operator SHAs; not scheduled; not ordinary |
| `existing_pr_number` | remains implement-only |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Pin target | this package's exact reviewed infra merge, not `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces (to be filled at implementation)

**Infrastructure (`KARSIFT/karsift-ai-infra`):** pending T00.

**Caller (`vocanova-platform-sandbox`):** pending T00.

## Shared-infra carrier and fixture pin (to be filled at implementation)

| Item | Value |
|------|-------|
| Coordinated infra PR | pending |
| Exact infra merge SHA | pending |
| Pin applicable? | expected yes — `release.yml`, helpers, tests, and current-state comments |
| Pin must not equal | `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| `PINNED_SHA.txt` | pending |
| Bootstrap used? | **no** |

## Validation commands

```bash
# Infrastructure (nested checkout at merged SHA)
python3 -m unittest discover -s tests -p 'test_*.py'

# Caller
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
node scripts/foundation/voc111-deploy-staging-paths.test.mjs
git diff --check
```

Record exact additional targeted commands and pass/fail results here after
implementation. Do not treat a missing suite as a pass.

## Acceptance mapping

- `VOC-127-AC-00` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-01` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-02` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-03` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-04` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-05` / `VOC-127-EV-00` — pending implementation
- `VOC-127-AC-06` / `VOC-127-EV-00` — pending implementation
