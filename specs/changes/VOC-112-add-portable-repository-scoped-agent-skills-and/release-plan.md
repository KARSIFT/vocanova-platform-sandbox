# VOC-112 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment by itself. Agent skills and
documentation take effect when task PRs merge to `develop`. No application artifact,
deploy bundle, or monitoring inventory change is in scope. Production receives nothing
from this package except whatever docs/skills are already on `develop` if promotion
later includes them incidentally through normal monorepo promotion — there is no
runtime dependency. No founder `approved` comment is a merge/adopt/release gate under
active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; tasks merge in order
  T00 → (T01 ∥ T02 ∥ T03) → T04; CI, governance validation, and independent verification
  pass on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No monitor or synthetic changes (`monitoring_impact.state: none`).
  Outcome evidence is deterministic foundation tests plus T04 benchmark/discovery docs.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #933:** closes with package roster completion after T04 satisfies the acceptance
  principle.

## Rollback

- **Trigger:** Unsafe skill content discovered post-merge; validation bypass; navigator
  benchmark regression beyond declared threshold; Graphify pilot commits unintended artifacts.
- **Mechanism:** Revert the task PR that introduced the defect; if adapter/canonical pair
  must be removed, remove both in the same revert. Disable Graphify skill by removing or
  marking opt-in-only with validation failure until remediated.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, `pnpm test` foundation suites and governance validation
  pass; no orphaned adapters remain.
- **Last-known-good:** `develop` HEAD before the reverted task merge.

## Independent verification, authority, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow merge gate.
  R3 evidence obligations: AGENTS.md changes reviewed for governance precedence; skill safety
  validation passes; provenance complete for adapted skills; benchmark/discovery evidence in
  `t04-evidence.md`.
- Closure: all AC results with evidence in `t00-evidence.md` through `t04-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion if other
  work triggers promotion — this package alone does not require production activation.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
