# VOC-084-T01 — Sign-in OAuth capability gating evidence

**Task:** VOC-084-T01  
**Package:** VOC-084  
**Evidence ID:** VOC-084-EV-01  
**Date:** 2026-08-15  
**Reviewed revision:** working tree on branch `agent/voc-084-voc-084-t01`

## Summary

The sign-in page now consumes deploy-derived OAuth capability from the API's
unauthenticated `GET /healthz` → `kill_switches.oauth_enabled` signal:

- Added `VocanovaClient.getHealthz()` in `@vocanova/api-client` (parses
  kill switches without throwing on HTTP 503).
- Added `getSignInAuthCapabilities()` in `apps/web/src/lib/auth-capabilities.ts`
  (fail-closed on probe errors or absent/false switch).
- `apps/web/src/app/signin/page.tsx` renders "Continue with Google" and the
  `or` divider only when `oauthEnabled` is true.
- Mock API `/healthz` now returns `kill_switches` with
  `oauth_enabled` driven by `MOCK_OAUTH_ENABLED` (default `false`).
- Deterministic unit and e2e tests cover disabled rendering and capability
  parsing branches.

Preserved: sign-in `max-w-[28rem]` layout workaround, magic-link form
accessibility, and remaining sign-in methods when OAuth is hidden.

## Validation commands and results

```bash
bash scripts/governance/validate-governance.sh
# (run during workflow)

bash scripts/governance/classify-change-risk.sh
# (run during workflow)

git diff --check
# (run during workflow)

pnpm --filter @vocanova/api-client test
# getHealthz + existing client tests PASS

pnpm --filter @vocanova/web test:middleware
# auth-capabilities.test.ts PASS

pnpm --filter @vocanova/web typecheck
# PASS

pnpm --filter @vocanova/web test:e2e tests/e2e/signin-oauth-capability.spec.ts
# VOC-084-TEST-05 disabled rendering PASS
```

## Test coverage mapping

| Test ID | Test | Result |
|---------|------|--------|
| VOC-084-TEST-05 | `signin-oauth-capability.spec.ts` hides Google when `oauth_enabled=false` | pending CI |
| VOC-084-TEST-05 | `auth-capabilities.test.ts` true/false/absent/error/503 branches | pending CI |
| VOC-084-TEST-05 | `@vocanova/api-client` `getHealthz` 503 kill-switch parse | pending CI |

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-084-AC-04 | implemented | UI gated on `/healthz` `kill_switches.oauth_enabled` |
| VOC-084-AC-05 (UI tests) | implemented | Unit + e2e deterministic coverage |

## Limitations

- Live staging end-to-end proof that enabled/disabled tracks deploy state
  depends on VOC-084-T00 having merged and converged staging OAuth config.
- Enabled-button e2e rendering is covered by unit tests parsing
  `oauth_enabled: true`; full SSR e2e with OAuth shown requires
  `MOCK_OAUTH_ENABLED=true` on the mock API process (optional local run).

## Follow-ups

- VOC-084-T02: post-deploy live OAuth-start check and Google callback
  disposition evidence.
