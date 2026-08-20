---
evidence_id: VOC-097-EV-00
task_id: VOC-097-T00
acceptance_criteria:
  - VOC-097-AC-00
tests:
  - VOC-097-TEST-00
  - VOC-097-TEST-01
date: 2026-08-20
related_change: VOC-097
accountable_owner: unassigned
gate_status: repository-complete-automation-deferred
live_reconcile_claimed: false
---

# VOC-097-T00 — Live-evidence contract declaration and author-facing docs

## Scope of this evidence

This task documents how change-package authors and operators declare
operator-owned live GitHub Actions evidence and the allowlisted metadata
contract. It does **not** implement waiting lifecycle behavior (T01),
reconcile/observe automation (T02), deterministic infra fixtures beyond doc
assertions (T03), stranded task migration (T04), or live proof (T05).

## VOC-097-DEP-03 resolution (contract path and shape)

Adoption did not record an alternate karsift convention. T00 confirms the
**VOC-097-D03 default**:

| Decision | Value |
| --- | --- |
| On-disk path | `<package-canonical-path>/.karsift/live-evidence/<task_id>.yaml` |
| Ownership values | `operator` or `live-actions` |
| Workflow identity | At least one of `workflow_file`, `workflow_name`, `workflow_id` |
| Job allowlist | `job_names` when the workflow has multiple jobs |
| Trigger allowlist | `events` (for example `push`, `workflow_dispatch`, `schedule`) |
| Branch | `branch` ref name |
| SHA lineage | `sha_lineage`: `exact_pr_head`, `integration_ancestor`, or `exact_sha` |
| Conclusion | `conclusion` (default `success`) |
| Staleness | optional `max_age` |
| Dispatch | optional `dispatch` block; absent means observe-only |

Roster-metadata-only declaration remains possible for future infra support, but
this repository's author-facing docs and templates standardize on the
`.karsift/live-evidence/<task_id>.yaml` path above so planners and operators
can locate contracts without reading karsift-ai-infra source.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Operator and author guide | `docs/operations/live-evidence.md` |
| Operations index entry | `docs/operations/README.md` |
| Change-package template guidance | `specs/templates/change-package/README.md` |
| Task template notes | `specs/templates/change-package/tasks.md` |
| Doc assertion tests | `scripts/foundation/voc097-operator-docs.test.mjs` |
| This evidence | `specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self/t00-evidence.md` |

## Acceptance mapping

| Acceptance criterion | Repository result |
| --- | --- |
| AC-00 | Operator guide names ownership, allowlisted metadata, fail-closed rules, and the `.karsift/live-evidence/<task_id>.yaml` contract path; template cross-links avoid contradictory author guidance. |

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc097-operator-docs.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
```

Secrets: none committed; no tokens, logs, or personal data in this evidence.

## Deferred to later tasks

- Waiting marker and remediate skip (`VOC-097-T01`)
- Reconcile workflow, timeout default enforcement, deduplication (`VOC-097-T02`)
- Stranded #779 / #785 contracts and migration (`VOC-097-T04`)
