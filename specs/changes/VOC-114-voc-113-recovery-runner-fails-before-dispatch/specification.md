# VOC-114 — Fix VOC-113 recovery runner metadata-read failure before dispatch: Specification

## Objective and requirement source

Repair the VOC-113 recovery runner so it can read exact-SHA gate metadata with
a repository-scoped short-lived credential, emit actionable sanitized diagnostics
when reads fail, and only then proceed to bounded wait and genuine workflow
dispatch. Live causal evidence authorizes separating the job Actions credential
from the App mutation identity when the installed App lacks Actions permission.

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

1. Restore effective repository-scoped read capability for the recovery runner's
   exact-SHA metadata phase (check-runs, commit status, Actions workflow runs,
   commit file list as already invoked by `actions-check-recovery-runner.py`).
2. Grant the minimum job-token permissions on every recovery path (merge-gate
   post-merge recovery, release converge recovery, and
   `recover-actions-checks.yml`) while keeping App tokens mutation-only.
3. Localize sanitized runner errors to endpoint classes without response bodies,
   tokens, logs, or user data.
4. Preserve existing narrow App mutation permissions (contents/issues/PR write)
   and grant Actions write only to the job tokens that perform allowlisted
   recovery dispatch.
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

- **Draft package proposal:** **R4** (CI/CD lifecycle orchestration and recovery
  credential contract for exact-SHA gate metadata).
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

`VOC-114-D01` (amended by live run `32724415871`): Recovery metadata reads and
allowlisted workflow dispatch MUST use the short-lived job `GITHUB_TOKEN`, with
explicit **Actions write**, **Checks read**, **Commit statuses read**, **Contents
read**, and **Pull requests read** permissions on each applicable caller/called
job. The App token remains limited to the existing Contents/Issues/Pull requests
mutation scopes required for App-identity merge, PR, marker, and issue operations;
it MUST NOT be a dependency for Actions metadata or dispatch.

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

`VOC-114-D05` (resolved by live run `32724415871`): The installed App has the
required mutation permissions but no Actions grant. Do not expand its installation
permissions merely to combine unrelated capabilities in one token. Keep Actions
metadata/dispatch on the explicit job-token boundary from D01 and record the
observed installation contract in T00/T01 evidence. Missing job permissions remain
fail-closed.

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

No new long-lived secrets. App installation tokens and job tokens remain short-lived
and scoped to distinct mutation and recovery responsibilities.
Evidence and diagnostics remain metadata-only (SHAs, run/job IDs, check names,
conclusions, sanitized error classes). Forbidden: logs, credentials, OAuth/session
material, tokens, user identifiers, personal data.

Granting recovery reads/Actions write only on dedicated jobs is lower risk than
dispatching without authoritative metadata; over-broad mutation grants remain
forbidden.

## Open questions

1. **Resolved — credential boundary:** live run `32724415871` and installation
   metadata proved the App has no Actions permission. D01 uses the job token for
   Actions metadata/dispatch and leaves the App mutation-only.
2. **Resolved — Statuses read:** combined commit-status metadata requires an
   explicit job-level `statuses: read` grant alongside `checks: read`.
3. **Historical integration SHA:** If `develop` has advanced beyond `b97e9575…`,
   live integration_push proof may target the still-relevant missing-run SHA
   recorded in evidence rather than forcing a stale SHA re-dispatch.
