# VOC-096 — Release Plan

## Release and deployment authorization

This package does not by itself authorize production deployment. After T01 merges, a
separately reviewed controlled activation promotion uses the repository-controlled
develop → main → automatic `deploy-production.yml` path so T02 can collect live
evidence without a circular roster-completion dependency. After roster completion, the
normal release reconciliation promotes any remaining documentation/evidence commits.
Under active A-004, no founder `approved` comment is a merge/adopt/release gate.

## Preconditions, monitoring, and outcome

- **Preconditions:** Confirm `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` is populated before
  T00 merges; record confirmation only, never the secret value. T00 and T01 merge to
  develop with CI, governance validation, and independent verification passing on each
  task PR. Promotion to `main` triggers automatic production deploy.
- **Exact revision:** recorded at task completion and promotion, not at drafting time.
- **Monitoring:** after deployment, `synthetic.production.oauth-expected-state` must pass,
  asserting OAuth start/callback and `controlled_signup_ready: true` on production
  `/healthz`. Production route/content smoke and journey synthetics should remain green.
- **Outcome owner:** unassigned (set at adoption).

## Rollback

- **Trigger:** production deploy fail-closed loop with OAuth enabled; live
  `controlled_signup_ready=false` after successful deploy; production OAuth synthetic
  regression unrelated to intentional empty cohort.
- **Mechanism:** revert offending task on develop and promote forward, or manually dispatch
  deploy-production after restoring last-known-good workflow on `main`. Preserve GitHub
  secret value through approved credential-management paths. Do not revert T00 alone while
  OAuth stays enabled — that restores empty-cohort convergence. If T00 must be reverted,
  first deploy with OAuth disabled or restore secret-backed workflow on a hotfix branch.
  No database rollback required.
- **Owner:** unassigned (set at adoption).
- **Validation:** after rollback or fix-forward, confirm `/healthz` boolean readiness,
  production OAuth synthetic, and deploy logs show boolean-only cohort diagnostics.
- **Last-known-good:** `main` HEAD before promotion of the failing package revision, or
  last green production deploy run URL recorded in T02 evidence.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow merge
  gate. R4 evidence obligations: deterministic tests pass, live production verification
  in T02, redaction verified, staging/production isolation confirmed, rollback path
  documented.
- Closure: all AC results with evidence in `t00-evidence.md`, `t01-evidence.md`, and
  `t02-evidence.md`. Package closure follows roster completion and release
  reconciliation after the controlled activation has supplied T02 live evidence.
- EHR: not triggered.

Do not conflate repository merge, release, activation, or closure. Historically under
A-003, R4 merge required founder approval. **Under active A-004,** engineering-workflow
gates require no founder `approved` comment; preserve triggered EHR evidence if product
ambiguity arises separately from merge gates.
