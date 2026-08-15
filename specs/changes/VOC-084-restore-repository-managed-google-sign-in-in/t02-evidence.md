# VOC-084-T02 — Post-deploy OAuth-start check + Google callback disposition evidence

**Task:** VOC-084-T02  
**Package:** VOC-084  
**Evidence ID:** VOC-084-EV-02  
**Date:** 2026-08-15  
**Reviewed revision:** working tree on branch `agent/voc-084-voc-084-t02`

## Summary

Added a post-deploy staging OAuth-start initiation gate to
`.github/workflows/deploy-staging.yml` and a dedicated verification script:

- `infra/scripts/verify-staging-oauth-start.sh` POSTs
  `https://api-staging.vocanova.site/api/v1/auth/oauth/google/start` with
  `redirectUri=https://staging.vocanova.site/home` after both public health
  polls pass.
- When repository Google credentials are both present
  (`EXPECT_OAUTH_ENABLED=true`), the check requires HTTP 200, an
  `accounts.google.com` authorization URL, and `redirect_uri` exactly
  `https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`.
- When both credentials are absent, the check requires coherent HTTP 503
  disabled behavior (not a fake 200).
- The script does not follow Google or complete OAuth (`curl` has no `-L`).
- `infra/scripts/verify-staging-oauth-start.selftest.sh` exercises both
  branches against a local fake server on `127.0.0.1`.

Production OAuth sync behavior and shared-edge isolation are unchanged.

## Validation commands and results

```bash
bash scripts/governance/validate-governance.sh
# (run during workflow)

bash scripts/governance/classify-change-risk.sh
# (run during workflow)

git diff --check
# (run during workflow)

chmod +x infra/scripts/verify-staging-oauth-start.sh \
  infra/scripts/verify-staging-oauth-start.selftest.sh

bash infra/scripts/verify-staging-oauth-start.selftest.sh
# ALL SELFTEST CASES PASSED

node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
# VOC-084-TEST-00 through TEST-08 PASS
```

## Test coverage mapping

| Test ID | Test | Result |
|---------|------|--------|
| VOC-084-TEST-06 | `deploy-staging wires live OAuth-start verification without following Google` | PASS |
| VOC-084-TEST-07 | `public health polls remain before OAuth-start gate` | PASS |
| VOC-084-TEST-07 | `production OAuth sync semantics are unchanged` (partial) | PASS |
| VOC-084-TEST-08 | `Google client callback disposition is precisely recorded` | PASS |
| VOC-084-TEST-06 | `verify-staging-oauth-start.selftest.sh` disposable harness | PASS |

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-084-AC-02 (live) | implemented | Post-deploy check asserts 200 + Google URL + exact callback when enabled |
| VOC-084-AC-05 (live check) | implemented | Wired in `deploy-staging.yml`; no OAuth callback/login in CI |
| VOC-084-AC-06 | implemented | Existing `/healthz` and web `/` polls remain; production paths untouched |
| VOC-084-AC-07 | external action recorded | Console authorization not verified from this environment (see below) |

## Google client callback authorization (`VOC-084-DEP-01`)

**Status: external action required — not verified complete from this implementer run.**

This implementer environment has no Google Cloud Console or Google Cloud API
access to inspect whether the existing repository-managed OAuth 2.0 client
already authorizes the staging callback. Repository work is complete; the
remaining one-time operator action is:

1. Open **Google Cloud Console → APIs & Services → Credentials**.
2. Select the **same OAuth 2.0 Client ID** already used for production
   (`GOOGLE_OAUTH_CLIENT_ID` repository secret).
3. Under **Authorized redirect URIs**, add this URI **exactly** (if not
   already present):

   `https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`

4. Save the client configuration.

**Why this is separate from the deploy gate:** `POST /oauth/google/start`
builds the Google authorization URL locally and does not contact Google.
The deploy gate can pass once staging credentials and config converge, even
if the callback URI is not yet registered. A real user login will fail at
Google's consent/redirect step until the URI above is authorized.

**Do not** invent Console mutations or commit client secrets in the repository.

## Live deploy evidence

A qualifying `deploy-staging` run URL and redacted live OAuth-start output
will be recorded when this task merges to `develop` and the workflow executes
against staging with repository credentials present. This evidence file
documents repository-side implementation and deterministic harness results;
live closure binds to that deploy run.

## Limitations

- No Google Cloud Console/API access in the implementer environment.
- No live staging deploy executed from this sandbox run (SSH/staging secrets
  not available here).
- Public health and OAuth-start live proof depends on the next staging deploy
  after merge.

## Files changed

- `.github/workflows/deploy-staging.yml` — post-deploy OAuth-start verification step
- `infra/scripts/verify-staging-oauth-start.sh` — live initiation checker
- `infra/scripts/verify-staging-oauth-start.selftest.sh` — disposable harness
- `scripts/foundation/voc084-deploy-staging-oauth.test.mjs` — TEST-06/07/08 coverage
- `specs/changes/VOC-084-restore-repository-managed-google-sign-in-in/t02-evidence.md` — this file
