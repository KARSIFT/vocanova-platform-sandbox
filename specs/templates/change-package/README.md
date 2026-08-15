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
