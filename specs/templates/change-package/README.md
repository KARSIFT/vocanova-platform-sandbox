# Unapproved Change-Package Template

This directory is a reusable placeholder, not an approved package and not
implementation authority. Replace every `VOC-000`, `REPLACE`, `TBD`, and `unknown`
value only from an approved requirement source.

## Identity and lifecycle

Record the package ID, title, canonical path, lifecycle state, risk, owner, approval
evidence, target branch, and linked GitHub issue.

## Objective and requirement source

State the approved objective and canonical documents or decisions that authorize it.

## Scope, non-goals, risk, and protected areas

Define bounded scope, explicit exclusions, protected paths, production effect, and
the highest applicable R0–R4 class. Record the active governance model, any EHR
trigger, and whether human authority arises from requirement clarification or another
explicit rule rather than assuming ownership routing proves approval.

## Verification, approvals, release, and closure

Link deterministic evidence, exact-SHA independent verification, required controls,
deployment/rollback controls, hosted activation, and closure evidence.

## `automatic_merge_allowed` drafting

Set `automatic_merge_allowed` in `change.yaml` per `AGENTS.md` (subsection
"Drafting `automatic_merge_allowed` in `change.yaml`") before the plan PR is
reviewed. **Under active A-004,** set `true` for **R0–R4**. The field is
audit-compatible; merge-gate does not treat `false` as a founder-attention gate.
Do not set `false` to require founder approval on merge.

**Historical (pre-A-004):** Under A-003 / VOC-075, R4 packages set `false` and R4
merge required founder approval. That policy is preserved in historical records only.

## `monitoring_impact` drafting

When a package is newly created or its `change.yaml` is modified, declare
`monitoring_impact` in `change.yaml` per VOC-086 and `AGENTS.md`:

- `state: none|existing|add|update`
- `none` requires a non-empty `rationale` and must not list IDs
- `existing`, `add`, and `update` require at least one valid `monitor_ids` or
  `synthetic_ids` entry from `infra/monitoring/monitors.yaml` and
  `infra/monitoring/synthetics.yaml`
- Page-route or critical API-endpoint changes fail CI when the change package
  lacks a valid `monitoring_impact` declaration

Historical packages whose `change.yaml` is untouched remain grandfathered until
modified.

## Operator-owned live evidence (VOC-097)

Some tasks need acceptance proof from a **live GitHub Actions run** after deploy or
schedule — for example production route sweeps or deploy-verification jobs. The
implementation agent intentionally lacks general Actions dispatch/inspect
credentials; missing operator-owned live evidence is **not** a code defect and
must not be "fixed" with unrelated pipeline edits.

When a task requires operator-owned live evidence:

1. Mark the task in `tasks.md` (see template notes in `tasks.md`).
2. Add a machine-readable **evidence contract** at
   `<package-canonical-path>/.karsift/live-evidence/<task_id>.yaml`.
3. Follow the field allowlist and metadata-only rules in
   [`docs/operations/live-evidence.md`](../../../docs/operations/live-evidence.md).

Each contract declares `ownership: operator` (or `live-actions`), allowlisted
workflow identity (`workflow_file`, `workflow_name`, and/or `workflow_id`),
optional `job_names`, required `events`, `branch`, `sha_lineage`, and
`conclusion`, optional `max_age`, and an optional `dispatch` block when repo
automation may trigger the run (otherwise observe-only). Automation fails closed
on ambiguous identity, wrong branch/SHA lineage, missing jobs, or non-success
conclusions.

Waiting, reconcile, and timeout behavior are implemented in karsift-ai-infra
(VOC-097-T01 onward); this template and operator guide define the author-facing
contract shape only.
