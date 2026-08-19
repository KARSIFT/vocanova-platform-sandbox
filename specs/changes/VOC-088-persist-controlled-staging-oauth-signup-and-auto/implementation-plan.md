# VOC-088 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-staging.yml`,
  `.github/workflows/deploy-production.yml`, `infra/monitoring/synthetics.yaml`,
  staging secrets, `apps/api/` healthz handler.
- Prerequisite: VOC-084 merged (staging Google OAuth credentials live).
- Prerequisite: VOC-086-T03 merged (scheduled synthetics infrastructure exists).

## File reconciliation and implementation sequence

### T00 — Persist staging allowlist secret and constrain ephemeral dispatch input

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/deploy-staging.yml` | modify | Read secret, constrain/remove dispatch input, add fail-closed guard, redact emails |
| `scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` | create | Deterministic tests for secret precedence, dispatch constraint, fail-closed, redaction |

Ordered steps:
1. Remove or constrain the `new_user_signup_allowlist` dispatch input.
2. Add `STAGING_NEW_USER_SIGNUP_ALLOWLIST: ${{ secrets.STAGING_NEW_USER_SIGNUP_ALLOWLIST }}`
   to the "Write staging application configuration" step's env block.
3. Replace `${STAGING_NEW_USER_SIGNUP_ALLOWLIST}` source from inputs to the secret.
4. Add fail-closed guard: if `STAGING_GOOGLE_OAUTH_ENABLED=true` and resolved
   allowlist is empty, exit 1 with non-sensitive diagnostic.
5. Add count-only log line.
6. Write deterministic tests.

### T01 — Extend staging OAuth/readiness synthetic for controlled signup readiness

| File | Action | Notes |
|------|--------|-------|
| `apps/api/app/api/production.go` (or healthz handler) | modify | Add `signup_allowlist_count` to healthz response |
| `apps/api/app/api/production_test.go` | modify | Test readiness field |
| `infra/scripts/verify-staging-oauth-start.sh` (or new harness) | modify/create | Assert signup_allowlist_count > 0 |
| `infra/monitoring/synthetics.yaml` | modify | Update coverage metadata |
| `.github/workflows/scheduled-synthetics.yml` | modify | Wire readiness assertion |
| `scripts/foundation/voc088-staging-readiness.test.mjs` | create | Deterministic synthetic readiness tests |

Ordered steps:
1. Extend healthz response with `signup_allowlist_count`.
2. Add Go unit tests for the new field.
3. Extend or create the synthetic harness script.
4. Update synthetics.yaml coverage.
5. Wire into scheduled-synthetics.yml.
6. Write Node deterministic tests for the synthetic logic.

### T02 — Add governed failure-to-issue agent

| File | Action | Notes |
|------|--------|-------|
| `infra/scripts/open-failure-issue.sh` (or equivalent) | create | Deduplicated, sanitized issue opener |
| `.github/workflows/scheduled-synthetics.yml` | modify | Add `if: failure()` step |
| `.github/workflows/deploy-staging.yml` | modify | Add `if: failure()` step |
| `.github/workflows/deploy-production.yml` | modify | Add `if: failure()` step |
| `scripts/foundation/voc088-failure-to-issue.test.mjs` | create | Deterministic deduplication, sanitization, identity tests |

Ordered steps:
1. Create the failure-to-issue helper script.
2. Wire into scheduled-synthetics.yml as an `if: failure()` step.
3. Wire into deploy-staging.yml.
4. Wire into deploy-production.yml.
5. Write deterministic tests.

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
- **Rollout:** T00–T02 merge to develop; staging deploy picks up changes
  automatically. T03 records evidence after deployment.
- **Rollback trigger:** allowlist sync fails; readiness synthetic false-positives;
  failure-to-issue creates noisy/duplicate issues.
- **Rollback mechanism:** revert the task PR. Next push deploy falls back to
  previous behavior. The GitHub secret persists harmlessly until deleted.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** the develop HEAD before each task PR merge.
