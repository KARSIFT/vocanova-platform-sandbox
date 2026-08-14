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
trigger, and whether human authority arises from R4 or another explicit rule rather
than assuming ownership routing proves approval.

## Verification, approvals, release, and closure

Link deterministic evidence, exact-SHA independent verification, required human
approvals, deployment/rollback controls, hosted activation, and closure evidence.

## `automatic_merge_allowed` drafting

Set `automatic_merge_allowed` in `change.yaml` per the risk-class rule in
`AGENTS.md` (subsection "Drafting `automatic_merge_allowed` in `change.yaml`")
before the plan PR is reviewed. **R0–R3** packages must set `true`; **R4**
packages must set `false` (self-describing; redundant with merge-gate's R4 hard
block). The template literal (`true`) is correct for the default R0 risk class.
Do not set `false` on non-R4 packages to require founder approval on merge.
