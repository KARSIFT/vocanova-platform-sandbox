# VOC-088 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-staging.yml`,
  `.github/workflows/deploy-production.yml`, `infra/monitoring/synthetics.yaml`,
  staging secrets, `apps/api/` healthz handler.
- Prerequisite: VOC-084 merged (staging Google OAuth credentials live).
- Prerequisite: VOC-086-T03 merged (scheduled synthetics infrastructure exists).
- Release prerequisite: repository secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST`
  must be populated before T00 merges, because that merge triggers an automatic
  staging deployment that intentionally fails closed when the secret is empty.

## File reconciliation and implementation sequence

### T00 — Persist staging allowlist secret and constrain ephemeral dispatch input

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/deploy-staging.yml` | modify | Read secret, constrain/remove dispatch input, add fail-closed guard, redact emails |
| `scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` | create | Deterministic tests for secret precedence, dispatch constraint, fail-closed, redaction |

Ordered steps:
1. Remove the `new_user_signup_allowlist` dispatch input.
2. Add `STAGING_NEW_USER_SIGNUP_ALLOWLIST: ${{ secrets.STAGING_NEW_USER_SIGNUP_ALLOWLIST }}`
   to the "Write staging application configuration" step's env block.
3. Replace `${STAGING_NEW_USER_SIGNUP_ALLOWLIST}` source from inputs to the secret.
4. Add fail-closed guard: if `STAGING_GOOGLE_OAUTH_ENABLED=true` and resolved
   allowlist is empty, exit 1 with non-sensitive diagnostic.
5. Add boolean-only log line.
6. Write deterministic tests.

### T01 — Extend staging OAuth/readiness synthetic for controlled signup readiness

| File | Action | Notes |
|------|--------|-------|
| `apps/api/app/api/production.go` (or healthz handler) | modify | Add `controlled_signup_ready` boolean to healthz response |
| `apps/api/app/api/production_test.go` | modify | Test readiness field |
| `infra/scripts/verify-staging-oauth-start.sh` (or new harness) | modify/create | Assert controlled_signup_ready is true |
| `infra/monitoring/synthetics.yaml` | modify | Update coverage metadata |
| `.github/workflows/scheduled-synthetics.yml` | modify | Wire readiness assertion |
| `scripts/foundation/voc088-staging-readiness.test.mjs` | create | Deterministic synthetic readiness tests |

Ordered steps:
1. Extend healthz response with `controlled_signup_ready`.
2. Add Go unit tests for the new field.
3. Extend or create the synthetic harness script.
4. Update synthetics.yaml coverage.
5. Wire into scheduled-synthetics.yml.
6. Write Node deterministic tests for the synthetic logic.

### T02 — Add governed failure-to-issue agent

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/operational-failure-monitoring.yml` | create | Standalone workflow_run observer with App-token issue creation |
| `infra/scripts/open-failure-issue.sh` (or equivalent) | create | Deduplicated, sanitized issue opener |
| `.github/workflows/scheduled-synthetics.yml` | inspect/test | Covered by observer workflow-name allowlist; no inline handler |
| `.github/workflows/deploy-staging.yml` | inspect/test | Covered by observer workflow-name allowlist; no inline handler |
| `.github/workflows/deploy-production.yml` | inspect/test | Covered by observer workflow-name allowlist; no inline handler |
| `scripts/foundation/voc088-failure-to-issue.test.mjs` | create | Deterministic deduplication, sanitization, identity tests |

Ordered steps:
1. Create a standalone `workflow_run` observer restricted to the three named
   workflows and completed non-success conclusions.
2. Mint an installation token with `KARSIFT_BOT_APP_ID` and
   `KARSIFT_BOT_PRIVATE_KEY`; do not create issues with the default token.
3. Create the failure-to-issue helper with stable workflow + conclusion
   deduplication and fixed sanitized content.
4. Add recursion guards so observer/test failures cannot create an uncontrolled
   issue cascade.
5. Write deterministic tests for early failure, cancellation, timeout,
   deduplication, sanitization, App identity, and environment mapping.

### T03 — Document operator procedure and deploy-and-verify evidence

| File | Action | Notes |
|------|--------|-------|
| `docs/operations/` (new or updated file) | create/modify | Operator procedure for cohort management |
| `specs/changes/VOC-088-.../t03-evidence.md` | create | Deploy-and-verify evidence |

Ordered steps:
1. Write operator procedure documentation.
2. Execute deploy-and-verify checklist after T00–T02 merged and deployed.
3. Record evidence.

## Validation and independent verification

For each task PR:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
pnpm validate
git diff --check
```

Independent verification: exact-SHA review of each task PR by the independent
reviewer role, confirming scope, test coverage, redaction, and acceptance criteria.

## Deployment and rollback

- **Authorization:** this package does not authorize production deployment.
  Deployment follows the normal merge → develop → main promotion path.
- **Rollout:** populate the staging allowlist repository secret first. T00–T02
  then merge to develop; staging deploy picks up changes automatically. T03
  records evidence after deployment.
- **Rollback trigger:** allowlist sync fails; readiness synthetic false-positives;
  failure-to-issue creates noisy/duplicate issues.
- **Rollback mechanism:** preserve the secret-backed allowlist in `api.env` while
  reverting unrelated tasks. Do not revert T00 while staging OAuth remains
  enabled; if T00 itself must be reverted, disable staging OAuth through the
  repository workflow first, then revert. No rollback may silently erase the
  controlled cohort.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** the develop HEAD before each task PR merge.
