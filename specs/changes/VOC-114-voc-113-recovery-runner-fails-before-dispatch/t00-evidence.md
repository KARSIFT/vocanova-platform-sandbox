# VOC-114-T00 — Evidence

Task: `VOC-114-T00` — Restore recovery metadata-read token contract, localize
errors, add tests.

Evidence date: 2026-08-24

Authority issue: `KARSIFT/vocanova-platform-sandbox#958`

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
`main` (`PINNED_SHA.txt` → `d3108dfdef34e2f98c028916e95c36130d329132`) shows
every App token feeding `actions-check-recovery-runner.py`
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

Deterministic contract changes land in this caller repository under
`tooling/governance/fixtures/karsift-ai-infra/` and must be mirrored byte-for-byte
into `KARSIFT/karsift-ai-infra` before live recovery can succeed on `@main`.
The companion shared-infra PR must use `Relates to KARSIFT/vocanova-platform-sandbox#958`
(non-closing). `PINNED_SHA.txt` advances only after that shared PR merges.

| Target | Change |
|--------|--------|
| `config/actions-check-recovery-runner.py` | Localized metadata-read failures to `check_runs_read_failed`, `workflow_runs_read_failed`, and `commit_metadata_read_failed`; extracted `run_metadata_phase()` so read failures abort before dispatch planning |
| `.github/workflows/merge-gate.yml` | App mint requests mutation scopes (`permission-contents/issues/pull-requests: write`, `permission-actions: write`) plus `permission-checks: read` |
| `.github/workflows/release.yml` | Same read/write mint contract on converge recovery token |
| `.github/workflows/recover-actions-checks.yml` | App mint requests `permission-actions: write`, `permission-checks: read`, `permission-actions: read`, `permission-contents: read`, and `permission-pull-requests: read` |
| `tests/test_voc114_actions_check_recovery.py` | Deterministic positive read contract, endpoint-class negatives, and no-dispatch-after-read-failure coverage for both modes |
| `README.md` (shared + fixture) | Documents recovery metadata read contract and sanitized endpoint classes |
| `docs/operations/11-devops-and-ci-cd.md` | Documents caller-facing recovery App read contract and endpoint classes |

Mutation posture is unchanged except for the declared read scopes required by
the metadata phase. Recovery dispatch still uses `permission-actions: write`
only on allowlisted workflows; no broadened mutation grants were introduced.

## Validation results

Commands run at implementation time (2026-08-24):

```text
# karsift-ai-infra checkout (local, mirrors fixture content)
PYTHONPATH=config python3 -m unittest discover -s tests -p 'test_*voc113*'
  → Ran 30 tests — OK
PYTHONPATH=config python3 -m unittest discover -s tests -p 'test_*voc114*'
  → Ran 10 tests — OK

# caller fixture mirror (authoritative for this PR)
PYTHONPATH=config python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*voc114*'
  → Ran 10 tests — OK
node --test scripts/foundation/voc114-actions-check-recovery.test.mjs
  → 3 tests — OK
node --test scripts/foundation/voc113-actions-check-recovery.test.mjs
  → existing VOC-113 caller tests — OK
bash scripts/governance/validate-governance.sh
  → passed
bash scripts/governance/classify-change-risk.sh
  → path floor R4 (package + infra fixture paths)
git diff --check
  → no conflicts
```

## Shared-infra adoption dependency

Live recovery on `@main` remains blocked until the byte-identical remediation
merges in `KARSIFT/karsift-ai-infra`. Record the shared PR URL and exact reviewed
head SHA in this section once that PR is opened and reviewed. T01 operator proof
depends on both this caller task PR and the shared-infra merge being live on the
branch the caller pipeline executes from.

## T01 dependency

Live integration_push and `reconcile-release` recovery proof for promotion PR
#947 remains operator-owned under `VOC-114-T01` after the shared infra changes
land on the branch the caller pipeline executes from.
