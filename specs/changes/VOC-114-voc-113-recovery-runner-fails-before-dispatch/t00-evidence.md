# VOC-114-T00 — Evidence

Task: `VOC-114-T00` — Restore recovery metadata-read token contract, localize
errors, add tests.

Evidence date: 2026-08-24

## Issue #956 observations (metadata-only)

| Item | Value |
|------|-------|
| Merged integration SHA | `b97e9575fd30671c336a2e92ca00db6e29b86416` (PR #954) |
| Integration recovery failure | Pipeline run `32696249484`, job `97339047384` — `integration_push` mode, immediate `github_metadata_read_failed` before bounded wait or dispatch |
| Release recovery failure | Pipeline run `32696549963`, job `97339655608` — `reconcile-release` for release issue #946, same immediate metadata failure in `promotion_pr` mode |
| Blocked downstream | VOC-113-T01; promotion PR #947 lacks genuine exact-head required checks |
| Prior mutation posture | Merge, task-completion publication, and issue closure succeeded before the recovery metadata-read phase failed |

## Verified root cause (`VOC-114-D00`, `VOC-114-D01`, `VOC-114-D05`)

Code review of the shared recovery mint paths on pre-T00 `karsift-ai-infra`
`main` shows every App token feeding `actions-check-recovery-runner.py`
(`merge-gate.yml` post-merge recovery, `release.yml` converge recovery,
`recover-actions-checks.yml`) invoked `actions/create-github-app-token` without
any `permission-*` inputs. The runner's metadata phase calls GitHub REST
endpoints for:

- commit check-runs and commit status aggregation (`/commits/{sha}/check-runs`,
  `/commits/{sha}/status`) — requires effective **Checks read**
- workflow-run discovery (`/actions/runs?head_sha=…`) — requires effective
  **Actions read**
- commit file metadata (`/commits/{sha}`) — requires effective **Contents read**
- promotion target validation (`/pulls/{number}`) — requires effective **Pull
  requests read**

When `create-github-app-token` is called without explicit `permission-*`
inputs, the minted installation token does not reliably carry the read scopes
needed for those metadata calls even though the hosting workflow job declares
`checks: read` / `actions: read` on `GITHUB_TOKEN`. The `gh` adapter used by
the recovery runner authenticates with the **App installation token**, not the
workflow `GITHUB_TOKEN`, so job-level read permissions do not satisfy the
metadata phase.

The live failure class therefore matches an App **mint-scope** defect: recovery
attempted metadata reads with a token that lacked explicitly requested Checks,
Actions, and Contents read. Pre-T00 code also collapsed every `gh` failure to the
generic `github_metadata_read_failed`, hiding which endpoint class failed.

**Installation-scope note (operator-owned before T01):** If live recovery still
fails after the mint contract lands, verify the KARSIFT GitHub App installation
on the caller repository grants repository-level **Checks: Read**, **Actions:
Read and write**, and **Contents: Read and write** (write is already required for
merge/release mutation). T01 operator proof owns that confirmation; T00 does not
weaken fail-closed behavior when installation grants are absent.

`permission-statuses: read` was not added: commit status aggregation uses the
same metadata phase as check-runs and is covered by explicit `permission-checks:
read` on every recovery mint path.

## Remediation applied (T00)

| Target | Change |
|--------|--------|
| `actions-check-recovery-runner.py` | Localized metadata-read failures to `check_runs_read_failed`, `workflow_runs_read_failed`, and `commit_metadata_read_failed`; metadata-read exceptions abort before bounded wait or dispatch planning |
| `merge-gate.yml` | App mint now requests `permission-contents/issues/pull-requests: write`, `permission-actions: write`, and `permission-checks: read` |
| `release.yml` | Same read/write mint contract on converge recovery token |
| `recover-actions-checks.yml` | App mint now requests `permission-actions: write`, `permission-checks: read`, `permission-contents: read`, `permission-pull-requests: read` |
| `tests/test_voc114_actions_check_recovery.py` | Deterministic positive read contract, endpoint-class negatives, and no-dispatch-after-read-failure coverage for both modes |
| `README.md` | Documents recovery metadata read contract and sanitized endpoint classes |

Mutation posture is unchanged except for the declared read scopes required by
the metadata phase. Recovery dispatch still uses `permission-actions: write`
only on allowlisted workflows; no broadened mutation grants were introduced.

## Validation results

Commands run from `karsift-ai-infra/` at implementation time (2026-08-24):

```text
python3 -m unittest discover -s tests -p 'test_*voc113*'  → Ran 30 tests — OK
python3 -m unittest discover -s tests -p 'test_*voc114*'  → Ran 9 tests — OK
python3 -m unittest tests.test_release_policy tests.test_merge_gate_policy
  → Ran 14 tests — OK
```

Caller governance validation (no caller workflow paths changed in T00):

```text
bash scripts/governance/validate-governance.sh  → passed
bash scripts/governance/classify-change-risk.sh → path floor R1 (karsift-ai-infra/)
git diff --check → no conflicts
```

## T01 dependency

Live integration_push and `reconcile-release` recovery proof for promotion PR
#947 remains operator-owned under `VOC-114-T01` after the shared infra changes
land on the branch the caller pipeline executes from.
