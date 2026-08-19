# VOC-088 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment. Deployment follows the
normal merge → develop → main promotion path under A-004. No founder `approved`
comment is a merge/adopt/release gate.

## Preconditions, monitoring, and outcome

- **Preconditions:** All four tasks (T00–T03) merged to develop. CI, governance
  validation, and independent verification pass on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** staging OAuth/readiness synthetic
  (`synthetic.staging.oauth-expected-state`) must pass after deployment. The
  failure-to-issue agent must not create spurious issues.
- **Outcome owner:** unassigned (set at adoption).

## Rollback

- **Trigger:** allowlist sync fails in production deploy; readiness synthetic
  produces false positives; failure-to-issue agent creates noisy or duplicate
  issues.
- **Mechanism:** revert the package's task PRs from develop. Next push deploy
  restores previous workflow behavior. The GitHub secret persists harmlessly.
  No database rollback required.
- **Owner:** unassigned (set at adoption).
- **Validation:** after rollback, confirm staging deploy succeeds with previous
  behavior and scheduled synthetics pass.
- **Last-known-good:** develop HEAD before each task PR merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic tests pass, scope confirmed,
  redaction verified, protected areas unchanged outside approved scope.
- Closure: all AC results with evidence recorded in t00–t03-evidence.md files.
  Package closure follows roster completion and develop → main promotion.
- EHR: not triggered.
