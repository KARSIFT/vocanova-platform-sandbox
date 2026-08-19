# VOC-088 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment. Deployment follows the
normal merge → develop → main promotion path under A-004. No founder `approved`
comment is a merge/adopt/release gate.

## Preconditions, monitoring, and outcome

- **Preconditions:** Before T00 merges, an operator confirms the repository secret
  `STAGING_NEW_USER_SIGNUP_ALLOWLIST` is populated; its value is never recorded.
  All four tasks (T00–T03) then merge to develop. CI, governance validation, and
  independent verification pass on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** staging OAuth/readiness synthetic
  (`synthetic.staging.oauth-expected-state`) must pass after deployment. The
  failure-to-issue agent must not create spurious issues.
- **Outcome owner:** unassigned (set at adoption).

## Rollback

- **Trigger:** staging allowlist sync fails; readiness synthetic
  produces false positives; failure-to-issue agent creates noisy or duplicate
  issues.
- **Mechanism:** preserve the secret-backed staging `api.env` value while
  reverting T01/T02. Do not revert T00 while OAuth stays enabled because doing so
  would restore the cohort-erasing behavior. If T00 must be reverted, first
  disable staging OAuth through a repository-driven deployment, then revert T00.
  The GitHub secret remains available for recovery. No database rollback required.
- **Owner:** unassigned (set at adoption).
- **Validation:** after rollback, confirm the controlled cohort is preserved (or
  OAuth is explicitly disabled), staging deploy succeeds, and scheduled
  synthetics reflect the selected safe state.
- **Last-known-good:** develop HEAD before each task PR merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic tests pass, scope confirmed,
  redaction verified, protected areas unchanged outside approved scope.
- Closure: all AC results with evidence recorded in t00–t03-evidence.md files.
  Package closure follows roster completion and develop → main promotion.
- EHR: not triggered.
