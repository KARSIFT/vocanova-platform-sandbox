# VOC-108 — Implementation Plan

## Phase 1 — Shared state model

1. Extract a deterministic helper that groups check runs/statuses by logical
   gate and selects the newest authoritative exact-SHA attempt with pagination.
2. Use it in adoption and every merge/release path that currently aggregates
   historical check runs. Preserve consumer-specific allowlists and self-check
   exclusions.
3. Add unit fixtures for ordering, duplicates, pagination, mismatch, ambiguity,
   and latest-failure/latest-success behavior.

## Phase 2 — Task completion provenance

1. Define one machine-readable App-authored completion marker bound to caller
   repository, issue, package, task, PR, reviewed head, and merge state.
2. Publish it only after the caller merge gate verifies the exact merge.
3. Make auto-advance and release validate the marker plus live caller PR state;
   closed-only and cross-repository events become no-ops.
4. Normalize or reject cross-repository closing-keyword text and document the
   non-closing `Relates to` contract.

## Phase 3 — Promotion convergence

1. Consolidate automatic and reconcile promotion into one serialized,
   idempotent evaluator/final merge helper.
2. Re-read exact PR/head/state before any decision comment and before merge.
3. Add a lightweight terminal-check event/reconcile path that reuses authoritative
   exact-SHA evidence instead of rerunning full CI/review.
4. Add race fixtures proving duplicate triggers have one effective merge and no
   stale post-merge comment.

## Phase 4 — Evidence and rollout

Run shared self-CI, caller governance/foundation contracts if changed, syntax and
diff checks, and independent exact-SHA review. Merge shared infra first; then land
caller evidence/adoption updates. Observe one real governed caller lifecycle
without creating artificial production or OAuth failures. Record only sanitized
metadata in `t00-evidence.md`.
