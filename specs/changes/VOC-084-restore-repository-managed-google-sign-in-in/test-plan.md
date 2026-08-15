# VOC-084 — Test Plan

## VOC-084-TEST-00 — Partial credential pair rejected before convergence

- Covers: `VOC-084-AC-00`
- Preconditions: `VOC-084-T00` tree; deterministic workflow/config harness
- Procedure:
  1. Drive the staging deploy OAuth sync logic (unit/selftest or extracted
     script) with exactly one of client id/secret present.
  2. Assert fail-closed exit before application convergence.
  3. Assert no credential values appear in captured stdout/stderr.
- Expected result: rejection; no half-configured OAuth write.
- Evidence: `VOC-084-EV-00`

## VOC-084-TEST-01 — Both credentials present sync safely

- Covers: `VOC-084-AC-01`
- Preconditions: `VOC-084-T00` tree
- Procedure:
  1. Provide both credentials to the sync logic under a disposable temp
     secret-file fixture (never real production secrets in the repo).
  2. Assert both keys are written, file mode is `0600`, and logs do not
     contain secret values.
  3. Assert `GOOGLE_OAUTH_ENABLED` derives true from pair availability.
- Expected result: safe sync + enabled derivation.
- Evidence: `VOC-084-EV-00`

## VOC-084-TEST-02 — Both credentials absent → coherent disabled state

- Covers: `VOC-084-AC-01`
- Preconditions: `VOC-084-T00` tree
- Procedure:
  1. Run sync/config with both secrets unset.
  2. Assert coherent disabled posture (`GOOGLE_OAUTH_ENABLED=false` or
     equivalent) and no partial credential leftovers required for
     acceptance.
- Expected result: clean disabled convergence (mirrors production
  skip/disable semantics adapted to staging).
- Evidence: `VOC-084-EV-00`

## VOC-084-TEST-03 — Canonical staging callback URI is exact

- Covers: `VOC-084-AC-02`
- Preconditions: `VOC-084-T00` tree
- Procedure:
  1. Inspect deploy-staging config-write logic / deterministic fixture
     output.
  2. Assert
     `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`
     exactly (no port qualification, no production host).
- Expected result: exact URI match.
- Evidence: `VOC-084-EV-00`

## VOC-084-TEST-04 — Signup remains false; allowlist defaults empty and controllable

- Covers: `VOC-084-AC-03`
- Preconditions: `VOC-084-T00` tree
- Procedure:
  1. Assert deploy-staging writes `NEW_USER_SIGNUP_ENABLED=false`.
  2. Assert allowlist defaults empty.
  3. Assert a non-empty workflow-controlled allowlist value is written when
     supplied (fixture), without enabling the global signup kill switch.
- Expected result: controlled cohort mechanism; empty means no first-time
  signup.
- Evidence: `VOC-084-EV-00`

## VOC-084-TEST-05 — UI hides Google when OAuth capability is disabled

- Covers: `VOC-084-AC-04`
- Preconditions: `VOC-084-T01` tree
- Procedure:
  1. Render sign-in with capability signal `oauth_enabled=false` (mock
     `/healthz` or equivalent harness).
  2. Assert "Continue with Google" is not displayed.
  3. Optionally assert it appears when capability is true (still without
     completing Google login).
- Expected result: disabled-method rendering is honest.
- Evidence: `VOC-084-EV-01`

## VOC-084-TEST-06 — Live staging OAuth-start check (no Google follow)

- Covers: `VOC-084-AC-02`, `VOC-084-AC-05`
- Preconditions: `VOC-084-T00` (and T02 wiring) on staging deploy path;
  repository credentials present for the positive case
- Procedure:
  1. From `deploy-staging.yml` post-deploy step, POST OAuth start on
     staging.
  2. Assert HTTP 200, `accounts.google.com` authorization URL, and
     `redirect_uri` exactly
     `https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`.
  3. Do not follow the redirect / complete OAuth.
- Expected result: live start proves initiation only.
- Evidence: `VOC-084-EV-02`

## VOC-084-TEST-07 — Public health and isolation remain intact

- Covers: `VOC-084-AC-06`
- Preconditions: staging deploy of this package's revision
- Procedure:
  1. Confirm staging/public health endpoints remain healthy after deploy.
  2. Confirm diff/scope did not alter production OAuth sync semantics or
     shared-edge isolation invariants.
- Expected result: healthy topology; isolation preserved.
- Evidence: `VOC-084-EV-02`

## VOC-084-TEST-08 — Google client callback authorization verified or precisely recorded

- Covers: `VOC-084-AC-07`
- Preconditions: `VOC-084-T02`
- Procedure:
  1. Using available access/evidence, determine whether the existing Google
     OAuth client authorizes the staging callback URI.
  2. If verified, record redacted evidence.
  3. If access is unavailable, record the exact external configuration
     action required and explicitly do **not** claim it complete.
- Expected result: verified **or** precise external remaining action —
  never a guessed Console mutation in-repo.
- Evidence: `VOC-084-EV-02`

Include positive, negative, authorization, failure, and rollback coverage
as applicable. Tests must not use production secrets or production data in
repository fixtures.
