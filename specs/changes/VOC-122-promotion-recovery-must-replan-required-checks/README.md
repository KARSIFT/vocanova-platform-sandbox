# VOC-122 — Promotion recovery must replan required checks that appear after its initial snapshot

| Field | Value |
|-------|-------|
| Package | `VOC-122` |
| Title | Promotion recovery must replan required checks that appear after its initial snapshot |
| Path | `specs/changes/VOC-122-promotion-recovery-must-replan-required-checks` |
| Status | `draft` |
| Risk | `R4` (draft proposal; promotion-recovery orchestration and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1001](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1001) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-121 made promotion recovery follow GitHub's required pull-request view and
rerun the exact cancelled or failed Actions run that view selects. That
selection contract is not enough when the selected row does not exist yet at
plan time.

Promotion recovery still plans reruns and absent-context dispatches **once**,
from `initial_gate_summary` / `pr_required_checks`, before its polling loop.
The loop then only refreshes metadata and asks `recovery_complete(...)`. If a
required row appears afterwards as failed or cancelled, that invocation sees
the unsatisfied row and waits until the 1800-second timeout instead of
recovering it.

### Live reproduction (2026-08-26)

| Item | Value |
|------|-------|
| Caller merge that opened the promotion | VOC-121 at `develop` `bb531fc211e91e40cec1847e1e28d98beaedacff` |
| Promotion PR | [#1000](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1000) |
| Recovery invocation that missed the row | release run `32912818046`, job `98010275057`, `2026-08-25T23:54:50Z` |
| Initial snapshot | required pull-request rows had not all appeared; absent-context dispatch of `governance-policy` as run `32912851134` (passed) |
| Later-appearing selected row | pull-request `governance-policy` run `32912850066` at `2026-08-25T23:54:52Z`, cancelled by concurrency |
| GitHub required view | `gh pr checks 1000 --required` continued to select that row as `CANCELLED` |
| Defect | `plan_required_check_recovery`, exact-run validation, rerun, and dispatch planning occur only before the `while` loop in `config/actions-check-recovery-runner.py` |
| Later recovery that worked | queued convergence run `32912851044`, job `98010820950`, started with the cancelled row already present, reran `32912850066` as attempt 2 |
| Outcome | App merged #1000 at `70db9d7dbc4789264732e4fe4f347914dd6f764e` without a ruleset bypass or fabricated status |

Root cause: a required row that transitions from absent to failed/cancelled
after the initial plan is visible but never actionable in that same recovery
run.

## Required outcome (summary)

Make promotion recovery safely re-evaluate newly appearing required-check rows
during polling while preserving the existing fail-closed contracts:

1. Validate the same open promotion PR, exact head SHA, branch, workflow path,
   event, attempt identity, and required context before every rerun.
2. Rerun a selected failed/cancelled exact pull-request run at most once per
   recovery invocation (deduplicate run IDs).
3. Dispatch a genuinely absent context at most once; do not let an alternate
   run or same-named status override GitHub's selected required row.
4. Do not duplicate active or successful work.
5. Keep the 1800-second timeout, App-token separation, exact-head merge
   decision, and retry limits intact.
6. Add deterministic time-evolving tests: absent in snapshot 1, cancelled in
   snapshot 2, rerun once, then success; repeated snapshots stay idempotent;
   ambiguous, foreign, and mismatched rows fail closed.
7. Update current-state recovery docs/comments and the caller fixture/pin
   through the coordinated exact-SHA flow.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Replan newly appearing required-check rows during promotion-recovery polling, with tests, docs, and caller pin | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because promotion recovery
mutates required Actions runs on an open develop→main PR. The path-based
classifier and independent verifier remain authoritative; this draft proposal
is not a determination.
