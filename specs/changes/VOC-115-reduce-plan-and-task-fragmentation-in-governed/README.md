# VOC-115 — Reduce plan and task fragmentation in governed engineering workflow

| Field | Value |
|-------|-------|
| Package | `VOC-115` |
| Title | Reduce plan and task fragmentation in governed engineering workflow |
| Path | `specs/changes/VOC-115-reduce-plan-and-task-fragmentation-in-governed` |
| Status | `draft` |
| Risk | `R4` (draft proposal; governance/workflow prompts, validation, and task-release sequencing) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#962](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/962) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed lifecycle currently fragments coherent work twice: the planner is told
to produce "small, ordered" independently reviewable tasks, and adoption opens one
issue per task with strictly serial advancement. That combination turns one feature
or bug into multiple full implementation/review/merge cycles even when the work is
still one coherent objective.

Issue #962 records concrete cost:

| Observation | Effect |
|-------------|--------|
| Planner prompt defaults to small ordered tasks | Over-splits a single objective into multiple PR carriers |
| Adopt opens one issue per task and later tasks wait for predecessor merge | Every extra task pays a full CI, review, remediation, merge, and completion cycle |
| Bugs found outside an already adopted task require a fresh issue → plan → adopt loop | In-scope causal fixes cannot stay with the active package/carrier |
| VOC-113 → VOC-114 recovery sequence | A post-merge in-scope failure required a new plan, new roster, and extra serial tasks before the original promotion could continue |

The result is higher elapsed time, GitHub Actions usage, model/token usage, issue
count, PR count, and operator attention without proportional safety benefit.

## Required outcome (summary)

1. Default one coherent feature, user objective, or bug to one change package and
   one end-to-end implementation task.
2. Permit in-scope causal remediation discovered during implementation, review,
   merge, promotion, or reconciliation to remain under the active package/carrier
   when it stays within the original objective, acceptance criteria, risk ceiling,
   and protected-area scope.
3. Require every task after the first to record a concrete split reason tied to a
   real merge-order, rollback, authority, environment, or review-size boundary.
4. Update canonical docs, planner/plan-review prompts, templates, validation, and
   mirrored fixtures/tests together while preserving exact-SHA review, path-based
   risk floors, protected-branch rules, fail-closed checks, and justified
   sequential advancement for genuine multi-task packages.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Consolidate package/task defaults, causal-remediation policy, validation, and regression coverage | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because it changes governance and workflow behavior
across prompts, adoption/release orchestration expectations, validation, and
protected documentation. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
