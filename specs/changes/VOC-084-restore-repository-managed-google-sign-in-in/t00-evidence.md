# VOC-084-T00 — Staging Google OAuth sync + canonical config evidence

**Task:** VOC-084-T00  
**Package:** VOC-084  
**Evidence ID:** VOC-084-EV-00  
**Date:** 2026-08-15  
**Reviewed revision:** working tree on branch `agent/voc-084-voc-084-t00`

## Summary

Added repository-managed Google OAuth credential synchronization and canonical
staging OAuth/signup configuration to `.github/workflows/deploy-staging.yml`,
mirroring the safe production pattern in `deploy-production.yml`:

- both repository secrets unset → skip credential write; `GOOGLE_OAUTH_ENABLED`
  derives false
- exactly one secret set → fail closed before SSH application convergence
  (runner validation + SSH sync step)
- both secrets present → write `GOOGLE_OAUTH_CLIENT_ID` /
  `GOOGLE_OAUTH_CLIENT_SECRET` to `/opt/vocanova/infra/secrets/api.env` with
  mode `0600` (values never logged)
- every deploy overwrites canonical
  `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`,
  staging `OAUTH_REDIRECT_ALLOWLIST`, `GOOGLE_OAUTH_ENABLED`,
  `NEW_USER_SIGNUP_ENABLED=false`, and `NEW_USER_SIGNUP_ALLOWLIST` from the new
  `workflow_dispatch` input (default empty)

Production OAuth sync behavior and shared-edge isolation are untouched.

## Validation commands and results

```bash
bash scripts/governance/validate-governance.sh
# Repository foundation validation passed.
# Governance structure validation passed.

bash scripts/governance/classify-change-risk.sh
# Detected path-based risk floor: R3

git diff --check
# (no whitespace errors)

node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
# 6 tests PASS (VOC-084-TEST-00 through TEST-04 + partial TEST-07 isolation)
```

## Test coverage mapping

| Test ID | Test | Result |
|---------|------|--------|
| VOC-084-TEST-00 | `partial Google credential pair fails before convergence` | PASS |
| VOC-084-TEST-01 | `both credentials present sync safely to staging api.env` | PASS |
| VOC-084-TEST-02 | `both credentials absent converges to coherent disabled OAuth` | PASS |
| VOC-084-TEST-03 | `canonical staging OAuth callback URI is exact` | PASS |
| VOC-084-TEST-04 | `signup stays false; allowlist defaults empty and is workflow-controlled` | PASS |
| VOC-084-TEST-07 | `production OAuth sync semantics are unchanged` (partial, config isolation) | PASS |

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-084-AC-00 | PASS | Runner + SSH partial-pair guards; no credential logging |
| VOC-084-AC-01 | PASS | Pair sync to staging `api.env` chmod 600; absent pair → `GOOGLE_OAUTH_ENABLED` false |
| VOC-084-AC-02 (config) | PASS | Exact canonical `OAUTH_REDIRECT_URI` written on every deploy |
| VOC-084-AC-03 | PASS | `NEW_USER_SIGNUP_ENABLED=false`; allowlist input defaults empty |
| VOC-084-AC-05 (workflow tests) | PASS | Deterministic foundation tests cover listed branches |
| VOC-084-AC-06 (isolation) | PASS | No production path writes; production workflow unchanged |

## Explicitly deferred to later tasks

| Item | Task |
|------|------|
| Live OAuth-start POST check on staging | VOC-084-T02 |
| Sign-in UI capability gating | VOC-084-T01 |
| Google Cloud Console callback authorization evidence | VOC-084-T02 (`VOC-084-DEP-01`) |

## Files changed

- `.github/workflows/deploy-staging.yml` — OAuth sync, config write, allowlist input, header comment
- `scripts/foundation/voc084-deploy-staging-oauth.test.mjs` — deterministic workflow/config tests
- `specs/changes/VOC-084-restore-repository-managed-google-sign-in-in/t00-evidence.md` — this file
