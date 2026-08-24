# VOC-114 — Fix VOC-113 recovery runner metadata-read failure before dispatch: Specification

## Objective and requirement source

Repair the VOC-113 recovery runner so it can read exact-SHA gate metadata with
the KARSIFT App installation token, emit actionable sanitized diagnostics when
reads fail, and only then proceed to bounded wait and genuine workflow dispatch.

**Requirement source:** [GitHub issue #956](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/956).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

### Confirmed problem evidence (issue #956)

| Item | Value |
|------|-------|
| Date | 2026-08-24 |
| Merged integration SHA | `b97e9575fd30671c336a2e92ca00db6e29b86416` (PR #954) |
| Integration recovery failure | Pipeline run 32696249484, job 97339047384 — `integration_push` mode, immediate `github_metadata_read_failed` |
| Release recovery failure | Pipeline run 32696549963, job 97339655608 — `reconcile-release` for release issue #946, immediate `github_metadata_read_failed` |
| Blocked downstream | VOC-113-T01; promotion PR #947 lacks genuine exact-head required checks |
| App mutation posture | Merge, task-completion publication, and issue closure succeeded before recovery read phase |

## Scope and non-goals

### In scope

1. Restore effective App token read capability for the recovery runner's exact-SHA
   metadata phase (check-runs, commit status, Actions workflow runs, commit file
   list as already invoked by `actions-check-recovery-runner.py`).
2. Request and document the minimum read permissions on every App mint path that
   feeds recovery (merge-gate post-merge recovery, release converge recovery,
   `recover-actions-checks.yml` reusable workflow).
3. Localize sanitized runner errors to endpoint classes without response bodies,
   tokens, logs, or user data.
4. Preserve existing narrow mutation permissions (contents/issues/PR write,
   actions write only where dispatch is already authorized).
5. Add deterministic tests for positive read contract, absent-permission fail-closed,
   and no-dispatch-after-read-failure for both `integration_push` and
   `promotion_pr` modes.
6. Operator-owned live evidence re-running both recovery paths and unblocking
   genuine exact-head checks on promotion PR #947 (VOC-113-T01 outcome).

### Non-goals / explicitly excluded

- Weakening or removing required ruleset checks.
- Synthesizing successful check runs or commit statuses.
- Manually merging promotion PR #947 or bypassing release converge.
- Broader VOC-113 redesign beyond the metadata-read / token-contract defect.
- Expanding App mutation permissions beyond what recovery already requires.
- Product runtime, credentials, signup policy, database, or monitoring inventory
  changes.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4** (CI/CD lifecycle orchestration and App token
  read contract for exact-SHA gate metadata).
- Protected areas: App-token merge/release mutation paths, VOC-108 authoritative
  exact-head selection, VOC-113 no-fabrication invariants, branch ruleset required
  contexts, `recover-actions-checks` recursion guards.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-114-D00`: Recovery metadata reads are **fail-closed prerequisites** for
bounded wait and workflow dispatch. A metadata-read failure MUST abort before any
recovery dispatch is attempted.

`VOC-114-D01`: The KARSIFT App installation token used by recovery MUST include
effective read access for every endpoint the runner calls in its metadata phase:
at minimum **Checks read** for commit check-runs/status aggregation and
**Actions read** for workflow-run discovery, plus **Contents read** for commit
file metadata already used by integration_push staging selection. Mutation scopes
remain the existing narrow set required for merge, PR create, and allowlisted
dispatch.

`VOC-114-D02`: When a metadata read fails, the runner MUST emit a sanitized
endpoint-class error — one of `check_runs_read_failed`, `workflow_runs_read_failed`,
or `commit_metadata_read_failed` — instead of collapsing all failures to
`github_metadata_read_failed`. Response bodies, tokens, logs, and user data MUST
NOT appear in stderr or raised messages.

`VOC-114-D03`: Deterministic tests MUST prove: (a) the declared token contract
allows metadata reads in both recovery modes; (b) absent read capability fails
closed with the appropriate endpoint class; (c) no dispatch plan is emitted or
executed after a metadata-read failure.

`VOC-114-D04`: Cross-repo execution follows the VOC-113 pattern. Primary behavior
lands in `KARSIFT/karsift-ai-infra`; caller evidence/docs land here when claims
would become false. Do not treat an untracked local `karsift-ai-infra/` checkout
as this repo's tracked tree.

`VOC-114-D05`: GitHub App **installation-level** permission grants (repository
settings outside git) may be required if the App is not already granted Checks
and Actions read at installation scope. When needed, record the required
installation permission change in T00 evidence and treat live proof (T01) as
operator-owned; do not weaken the fail-closed contract when installation grants
are missing.

`VOC-114-D06`: Live T01 reuses the existing promotion fixture (PR #947, release
issue #946) and integration SHA `b97e9575…` only as metadata anchors. Completing
#947 remains subject to VOC-108 authoritative exact-head success and VOC-113
no-fabrication rules. After recovery, the operator MUST dispatch the existing
read-only `verify-promotion-check-recovery` action on the exact T01 evidence-carrier
branch. The machine contract observes job
`verify-promotion-check-recovery / verify` with `exact_pr_head` lineage; it MUST
NOT use `integration_contains_pr_head` for a run on `develop`, because the
unmerged carrier head cannot be an ancestor of that integration run.

## Data, migrations, analytics, and accessibility

None. Governance-automation recovery fix only.

## Security, privacy, and authorization

No new long-lived secrets. App installation tokens remain short-lived and scoped.
Evidence and diagnostics remain metadata-only (SHAs, run/job IDs, check names,
conclusions, sanitized error classes). Forbidden: logs, credentials, OAuth/session
material, tokens, user identifiers, personal data.

Expanding read permissions is lower risk than dispatching without authoritative
metadata, but over-broad mutation grants remain forbidden.

## Open questions

1. **Installation vs mint scope:** T00 must confirm whether failure is resolved
   solely by `permission-checks: read` / `permission-actions: read` on
   `create-github-app-token`, by installation-level App permission grants, or
   both. Do not guess in this draft beyond the issue's leading hypothesis.
2. **Statuses read necessity:** If commit status aggregation still fails after
   Checks read is restored, T00 may add `permission-statuses: read` or document
   why Checks read alone suffices via REST — subject to fail-closed tests.
3. **Historical integration SHA:** If `develop` has advanced beyond `b97e9575…`,
   live integration_push proof may target the still-relevant missing-run SHA
   recorded in evidence rather than forcing a stale SHA re-dispatch.
