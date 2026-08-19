# VOC-096 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-production.yml`, production repository
  secrets, `infra/scripts/verify-production-oauth-start.sh`,
  `infra/monitoring/synthetics.yaml`, production host `api.env`.
- Prerequisite: VOC-088-T01 merged (`controlled_signup_ready` on `/healthz`).
- Prerequisite: VOC-086-T03 merged (scheduled production OAuth synthetic exists).
- Prerequisite: repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` populated
  before T00 merges, because promotion to `main` triggers automatic production deploy
  that must fail closed if OAuth is enabled and the secret is empty.

## File reconciliation and implementation sequence

### T00 — Persist production allowlist secret and remove ephemeral dispatch input

| File | Action | Notes |
|------|--------|-------|
| `infra/scripts/validate-production-signup-allowlist.sh` | create | Mirror staging validator with production env var names |
| `.github/workflows/deploy-production.yml` | modify | Remove dispatch input; secret-backed allowlist; validation step; fail-closed SSH guard |
| `scripts/foundation/voc096-deploy-production-allowlist.test.mjs` | create | Secret precedence, isolation, fail-closed, redaction |
| `scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` | modify (optional) | Extend isolation assertion if not fully covered in voc096 test |

Ordered steps:

1. Create production validator script (copy/adapt staging script; production diagnostics text).
2. Remove `new_user_signup_allowlist` from workflow_dispatch inputs.
3. Add runner validation step with `secrets.PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`.
4. Update config step env to use secret instead of `inputs.new_user_signup_allowlist`.
5. Add SSH-side empty-cohort guard and boolean-only readiness log lines.
6. Write foundation tests; run `node --test scripts/foundation/voc096-deploy-production-allowlist.test.mjs`.

### T01 — Extend production OAuth/readiness synthetic

| File | Action | Notes |
|------|--------|-------|
| `infra/scripts/verify-production-oauth-start.sh` | modify | Add /healthz controlled_signup_ready assertion (mirror staging) |
| `infra/monitoring/synthetics.yaml` | modify | Update synthetic.production.oauth-expected-state coverage |
| `.github/workflows/scheduled-synthetics.yml` | modify (comment only, if needed) | Document readiness assertion |
| `scripts/foundation/voc096-production-readiness.test.mjs` | create | Harness + inventory tests |
| `infra/monitoring/scheduled-synthetics.mjs` or validate-inventory | inspect | Ensure inventory validators accept updated coverage |

Ordered steps:

1. Port staging healthz readiness block into production harness.
2. Update synthetics.yaml coverage metadata.
3. Add foundation tests for harness and inventory.
4. Run applicable monitoring inventory / scheduled-synthetics foundation tests.

### T02 — Operator procedure and production deploy-and-verify evidence

| File | Action | Notes |
|------|--------|-------|
| `docs/operations/production-controlled-signup.md` | create | Operator guide |
| `scripts/foundation/voc096-operator-docs.test.mjs` (or extend voc092) | create/modify | Doc existence test |
| `specs/changes/VOC-096-.../t02-evidence.md` | create | Live production verification |

Ordered steps:

1. Write production operator documentation.
2. After T01 merges, open and independently review a controlled activation promotion
   from `develop` to `main`; merge it only after exact-SHA checks pass.
3. Monitor the automatic production deployment and capture scrubbed evidence per AC-10.
4. Record evidence in `t02-evidence.md`.

## Validation and independent verification

For each task PR:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
pnpm validate
git diff --check
node --test scripts/foundation/voc096-deploy-production-allowlist.test.mjs
node --test scripts/foundation/voc096-production-readiness.test.mjs
```

Independent verification: exact-SHA review of each task PR by the independent
reviewer role, confirming scope, test coverage, redaction, production tier isolation,
and acceptance criteria. Under active A-004, R4 carries strengthened evidence
obligations but no founder `approved` comment merge gate.

## Deployment and rollback

- **Authorization:** this package does not authorize production deployment by itself.
  The reviewed task changes use the repository-controlled `develop` → `main`
  promotion path and automatic production workflow.
- **Rollout:** confirm `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` populated. Merge T00 → T01
  to `develop`; then use a separately reviewed controlled activation promotion to
  `main`. The automatic production deploy applies the secret-backed cohort, and T02
  records live evidence. After roster completion, the normal release reconciliation
  promotes any remaining documentation/evidence commits.
- **Rollback trigger:** production deploy fails closed unexpectedly; synthetic false
  positives; cohort accidentally erased by reverted workflow.
- **Rollback mechanism:** revert task PRs on develop and promote a fix forward. Do not
  revert T00 while production OAuth remains enabled without an alternate cohort source —
  that restores the erasing behavior. If T00 must be reverted, disable production OAuth
  through repository-driven means first, then revert. GitHub secret remains for recovery.
  No database rollback required.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** develop HEAD before each task PR merge.
