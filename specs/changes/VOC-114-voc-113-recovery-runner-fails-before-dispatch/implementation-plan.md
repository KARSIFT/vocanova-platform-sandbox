# VOC-114 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` merge-gate / release / recovery
  mutation paths, App installation token minting, job-token recovery dispatch,
  VOC-108 authoritative check
  selection, VOC-113 no-fabrication invariants, branch ruleset required contexts.
- Prerequisites: VOC-113-T00 recovery mechanism merged; App credentials
  (`KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`) configured; issue #956 live
  failures reproduced as metadata-only evidence.
- Cross-repo: open infra PRs against `KARSIFT/karsift-ai-infra` with
  `Relates to KARSIFT/vocanova-platform-sandbox#<task>` (non-closing). Do not
  treat an untracked local `karsift-ai-infra/` checkout as this repository's
  tracked tree.

## File reconciliation and implementation sequence

### T00 — Restore metadata-read token contract, localize errors, add tests

| Target | Action | Notes |
|--------|--------|-------|
| `specs/changes/VOC-114-.../t00-evidence.md` | create/update | Diagnosis + validation results (metadata-only) |
| `karsift-ai-infra/config/actions-check-recovery-runner.py` | modify | Endpoint-class errors; abort before dispatch on read failure |
| `karsift-ai-infra/.github/workflows/merge-gate.yml` | modify | Job-token recovery scopes; mutation-only App mint |
| `karsift-ai-infra/.github/workflows/release.yml` | modify | Job-token converge recovery; mutation-only App mint |
| `karsift-ai-infra/.github/workflows/recover-actions-checks.yml` | modify | Job-token metadata/dispatch scopes |
| `karsift-ai-infra/tests/test_voc113_actions_check_recovery.py` | modify/extend | VOC-114 read-contract and no-dispatch negatives |
| Additional focused VOC-114 policy tests | create if cleaner | Absent-permission endpoint-class fixtures |
| karsift-ai-infra README; package/ops docs | modify when claims become false | Document separated token contract |

Ordered steps:

1. Record issue #956 plus live run `32724415871` and the verified App/job-token
   boundary (`VOC-114-D05`).
2. Put Actions write and required metadata reads on each dedicated recovery job;
   remove Actions/Checks/Statuses requests from mutation App mints.
3. Localize `gh` adapter failures to `check_runs_read_failed`,
   `workflow_runs_read_failed`, `commit_metadata_read_failed`.
4. Guard dispatch planning/execution so metadata-read failures exit before wait
   or dispatch.
5. Extend deterministic tests for both recovery modes and absent-permission
   negatives.
6. Update docs whose recovery credential or App permission claims would become false.
7. Run applicable validation; record results in `t00-evidence.md`.

### T01 — Live recovery proof and unblock promotion PR #947

| Target | Action | Notes |
|--------|--------|-------|
| `t01-evidence.md` | update | Allowlisted live recovery metadata for both modes |
| `.karsift/live-evidence/VOC-114-T01.yaml` | observe/dispatch | Operator-owned contract |

Ordered steps:

1. Confirm the explicit job-token recovery permissions are active; do not expand
   the App installation for Actions access.
2. Re-run integration_push recovery for documented SHA `b97e9575…` (or T00
   recorded still-blocking SHA).
3. Dispatch `reconcile-release` for release issue #946; confirm promotion PR #947
   exact-head recovery progresses past metadata read.
4. Confirm genuine required checks on #947's exact head; allow release converge
   to merge only under VOC-108 success.
5. Dispatch the read-only `verify-promotion-check-recovery` action on the exact
   T01 carrier head and require job `verify-promotion-check-recovery / verify`.
6. Record metadata-only evidence; do not manually merge #947.

## Validation and independent verification

Deterministic (T00), as applicable to the real file list:

```bash
# In KARSIFT/karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*voc113*'
python3 -m unittest discover -s tests -p 'test_*voc114*'   # if separate suite added
python3 -m unittest tests.test_release_policy tests.test_merge_gate_policy

# In this calling repository when caller paths change
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Independent verifier (exact reviewed task PR SHA):

- Confirm job-token read/dispatch scopes are minimum necessary and App mutation
  posture is not broadened.
- Confirm endpoint-class errors contain no secrets and generic
  `github_metadata_read_failed` is not used for localized endpoint failures.
- Confirm no dispatch after metadata-read failure in deterministic fixtures.
- For T01, confirm operator-owned metadata is allowlisted and bound to declared
  SHA/PR lineage.

## Deployment and rollback

- **Staging effect:** None intentional beyond existing develop push selection once
  recovery reads succeed.
- **Production effect:** Unblocking PR #947 may promote reviewed history and
  trigger the existing automatic `main` deploy path — recovery of stranded release
  handoff, not new deploy policy.
- **Rollback trigger:** Recovery dispatches without readable metadata; broadened
  mutation scopes; false-negative blocks on legitimate green SHAs; reintroduced
  generic errors hiding endpoint failures.
- **Rollback mechanism:** Revert T00 infra PR(s); recovery returns to immediate
  fail-closed metadata error (known broken state).
- **Last-known-good reference:** Pre-T00 recovery runner and workflow mint blocks
  on consumed karsift-ai-infra `@main` immediately before T00 lands.
