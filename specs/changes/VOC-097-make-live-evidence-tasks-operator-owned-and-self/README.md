# VOC-097 — Make live-evidence tasks operator-owned and self-reconciling

| Field | Value |
|-------|-------|
| Package | `VOC-097` |
| Title | Make live-evidence tasks operator-owned and self-reconciling |
| Path | `specs/changes/VOC-097-make-live-evidence-tasks-operator-owned-and-self` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue #823 |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Evidence-only tasks that require a live GitHub Actions run can become permanently
stranded even after the underlying system is healthy. The implementation agent
intentionally lacks Actions API credentials, so it cannot dispatch or inspect the
qualifying run. Independent review correctly rejects pending evidence, but
remediation then treats the missing operator evidence as a code defect, consumes
bounded retries, and can introduce unrelated pipeline or task-specific workflow
changes.

Observed examples from issue #823:

- PR #765 / task #762 — valid later live evidence existed, but the branch
  accumulated out-of-scope pipeline and task-specific workflow changes.
- PR #789 / task #779 (`VOC-093-T01`) — production route-sweep evidence remains
  pending with substantial pipeline scope expansion on the branch.
- PR #791 / task #785 (`VOC-094-T01`) — evidence remains pending despite later
  healthy deployment activity.

## Required outcome (summary)

1. Explicit lifecycle state for operator/live-evidence waiting, distinct from
   implementation or review failure.
2. Keep the implementer least-privileged; do not grant general Actions credentials.
3. Repository-controlled automation dispatches or observes only workflows/jobs
   declared by the governed task/package.
4. Collect only allowlisted metadata (workflow identity, event, branch, exact SHA,
   run/job IDs, conclusion, timestamps, bounded duration). Never copy logs,
   artifacts, secrets, OAuth data, sessions, cookies, tokens, user identifiers, or
   arbitrary output.
5. Reconcile waiting tasks when qualifying evidence arrives; require a new
   exact-SHA independent review before merge.
6. Fail closed on ambiguous workflow identity, wrong branch/SHA lineage, stale
   runs, missing jobs, non-success conclusions, or malformed evidence contracts.
7. Do not spend remediation retries or permit unrelated code/workflow edits while
   the only unmet condition is operator-owned live evidence.
8. Bounded timeout/escalation and deduplication so a missing live run cannot loop
   indefinitely.
9. Deterministic tests for waiting, successful reconciliation, timeout, wrong
   workflow/job/branch/SHA, duplicates, sanitization, and no-credential/no-log
   invariants.
10. Reconcile stranded tasks #779 and #785 via the governed mechanism or an
    explicitly documented safe migration path.
11. Update governance and operator documentation for declaring live-evidence
    ownership and the allowlisted contract.
12. Preserve existing risk classification, exact-SHA review, branch protection,
    App-authenticated mutations, deployment isolation, and release behavior.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Declare live-evidence contract and author-facing docs/templates | — |
| T01 | Waiting lifecycle; stop remediation on operator-owned pending evidence | T00 |
| T02 | Allowlisted observe/dispatch reconciler with timeout, dedup, wake + re-review | T01 |
| T03 | Deterministic fixture matrix (waiting, reconcile, reject, sanitize) | T02 |
| T04 | Reconcile stranded #779 and #785 | T02 |
| T05 | Controlled live proof and observer/Sentry separation health | T03, T04 |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Grant the implementer general GitHub Actions credentials.
- Merge live-evidence waiting into the operational-failure observer or Sentry path.
- Weaken exact-SHA independent review, branch protection, risk floors, or release gates.
- Self-adopt or self-authorize this package.
