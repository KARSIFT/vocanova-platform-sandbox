# Live-evidence tasks (VOC-097)

Operator and change-package author guide for tasks whose acceptance depends on a
**live GitHub Actions run** that the implementation agent cannot dispatch or
inspect. Repository-controlled automation (not the implementer) observes or
dispatches only workflows declared in each task's evidence contract.

## When to declare operator-owned live evidence

Use this pattern when a task's acceptance criteria require proof from a real
Actions run after merge or deploy — for example production route sweeps,
deploy-verification jobs, or scheduled synthetics — and the implementer role
must remain least-privileged (no general `actions: write` or run-inspection
credentials).

Do **not** use live-evidence waiting for:

- Code or documentation defects fixable in the task PR.
- Operational-failure observer paths (`operational-failure-monitoring.yml`,
  Sentry ingestion) — those remain separate mechanisms.
- Snapshot-then-drift patterns that invalidate themselves when the repo moves
  on.

Missing operator-owned live evidence is **not** an implementation defect. It
enters an explicit **waiting** lifecycle state (implemented in karsift-ai-infra
from VOC-097-T01 onward) that does not consume remediation retries and must not
prompt unrelated pipeline or workflow edits solely to manufacture evidence.

## Machine-readable evidence contract

Each live-evidence task declares an allowlisted contract at:

```text
<package-canonical-path>/.karsift/live-evidence/<task_id>.yaml
```

Example for task `VOC-093-T01` in package
`specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure`:

```text
specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-093-T01.yaml
```

Replace `<task_id>` with the adopted roster task ID (for example `VOC-097-T04`).
One file per live-evidence task. Planners add the contract in the plan PR or
the first task that needs it; do not commit contracts without an approved
package.

### Required and optional fields

| Field | Required | Purpose |
| --- | --- | --- |
| `ownership` | yes | Must be `operator` or `live-actions`. Marks the task as operator-owned live evidence. |
| `workflow_file` | one of workflow identity fields | Allowlisted workflow path relative to `.github/workflows/` (for example `deploy-production.yml`). |
| `workflow_name` | one of workflow identity fields | Stable workflow `name:` when file alone is ambiguous. |
| `workflow_id` | one of workflow identity fields | Numeric workflow ID when file and name are unstable across repos. |
| `job_names` | when the workflow has multiple jobs | Allowlisted job `name:` values that must succeed. Omit only for single-job workflows. |
| `events` | yes | Allowlisted trigger set (`push`, `workflow_dispatch`, `schedule`, …). |
| `branch` | yes | Required ref name for qualifying runs (for example `main`, `develop`). |
| `sha_lineage` | yes | How the run's HEAD SHA relates to the task PR or integration tip (see below). |
| `conclusion` | yes | Required success conclusion; default and recommended value is `success`. |
| `max_age` | no | Staleness bound (for example `72h`) after which a run no longer qualifies. |
| `dispatch` | no | When present, repo automation **may** dispatch this workflow with exactly these inputs; when absent, **observe-only**. |

At least one of `workflow_file`, `workflow_name`, or `workflow_id` is required.
Automation fails closed when identity is ambiguous or multiple candidates match.

### SHA lineage rules

Declare exactly one lineage rule under `sha_lineage`:

| Rule | Meaning |
| --- | --- |
| `exact_pr_head` | Run HEAD SHA must equal the task PR head SHA at reconcile time. |
| `integration_ancestor` | Run HEAD SHA must be an ancestor of (or equal to) the integration branch tip that contains the task PR. |
| `exact_sha` | Run HEAD SHA must equal a pinned full SHA recorded in the contract (rare; prefer PR-head rules). |

Wrong branch, wrong SHA lineage, missing required jobs, non-success conclusions,
or stale runs (past `max_age`) are rejected without waking the task and without
starting remediation-as-fix.

### Example contract (observe-only)

```yaml
schema_version: 1
task_id: VOC-094-T01
ownership: operator

workflow_file: deploy-production.yml
job_names:
  - deploy-production
events:
  - push
branch: main
sha_lineage: integration_ancestor
conclusion: success
max_age: 72h
# No dispatch block — repository automation observes qualifying runs only.
```

### Example contract (automation may dispatch)

```yaml
schema_version: 1
task_id: VOC-093-T01
ownership: live-actions

workflow_file: smoke-test-production.yml
job_names:
  - production-route-sweep
events:
  - workflow_dispatch
branch: main
sha_lineage: exact_pr_head
conclusion: success
max_age: 24h
dispatch:
  workflow_file: smoke-test-production.yml
  inputs:
    reason: voc097-live-evidence
```

The `dispatch` block must mirror an allowlisted workflow and inputs. Automation
must not dispatch workflows undeclared in the contract.

## Allowlisted evidence metadata only

Reconcile outputs, task-issue comments, PR comments, and task evidence files may
include **only** these metadata fields:

- Workflow identity (`workflow_file`, `workflow_name`, and/or `workflow_id`)
- Trigger `event`
- Branch ref and exact run HEAD SHA
- Run ID and job ID(s)
- Job/run `conclusion`
- Timestamps and bounded duration

**Forbidden** in governed paths: workflow logs, artifacts, secrets, OAuth or
session material, cookies, tokens, user identifiers, email addresses, and
arbitrary job output. Paste run URLs only when they contain no credentials;
prefer run/job IDs in evidence tables.

## Lifecycle (after VOC-097-T01 / T02 land)

| State | Meaning |
| --- | --- |
| **Waiting** | Declared contract exists; code/docs meet review except live evidence. Remediation does **not** retry. |
| **Ready for re-review** | Qualifying evidence arrived; exactly one reconcile wake per run identity. |
| **Merged** | Fresh independent verification bound to the **exact** post-reconcile PR head SHA (prior PASS on an earlier SHA does not suffice). |

Default waiting timeout (package proposal): **72 hours** wall clock from
entering waiting, then a single sanitized escalation comment on the task issue
and no automatic retry loops. Exact timeout wiring is implemented in
VOC-097-T02.

Live-evidence lifecycle signals stay separate from
`operational-failure-monitoring.yml` and Sentry. Expected waiting outcomes must
not open operational-failure issues or application alerts.

## Author checklist

1. In the package `tasks.md`, state which task(s) require operator-owned live
   evidence and link to each contract path.
2. Add `<package>/.karsift/live-evidence/<task_id>.yaml` before or during the
   task that needs it.
3. Document in the task evidence file (`tNN-evidence.md`) which contract fields
   apply and what the operator must trigger or wait for (no secrets).
4. Do not ask the implementer to dispatch Actions or paste logs — use waiting +
   reconcile after T01/T02, or operator-triggered runs that match the contract.
5. Cross-link this guide from the change-package template notes when adding live
   evidence (see `specs/templates/change-package/README.md`).

## Operator checklist

1. Confirm the task PR shows the waiting marker (after T01) and the contract
   file on the branch.
2. Trigger or wait for the declared workflow/event on the declared branch so the
   run matches `sha_lineage` and job set.
3. Do not paste logs, tokens, or cohort values into issues or PRs — only
   allowlisted metadata or scrubbed run URLs.
4. After reconcile wakes the task, ensure independent review runs again on the
   **current** PR head SHA before merge.

## Related documentation

- Change-package template live-evidence notes:
  `specs/templates/change-package/README.md`
- VOC-097 package specification and decisions:
  `specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/specification.md`
- Monitoring and operational-failure paths (separate from live evidence):
  `docs/operations/monitoring.md`
