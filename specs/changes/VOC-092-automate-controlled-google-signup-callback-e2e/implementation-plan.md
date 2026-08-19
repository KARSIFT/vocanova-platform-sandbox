# VOC-092 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `apps/api/business/auth/`, `apps/api/app/api/auth.go`,
  `.github/workflows/` CI paths, `docs/operations/staging-controlled-signup.md`.
- Prerequisites: VOC-088 merged (controlled staging signup, readiness synthetic,
  operational-failure observer).
- Prerequisites: Existing disposable Postgres integration-test infrastructure in
  `apps/api` (Docker available in CI).
- The harness must not modify production OAuth provider selection or staging
  deploy configuration.

## File reconciliation and implementation sequence

### T00 — Ephemeral OAuth controlled-signup callback E2E harness

| File | Action | Notes |
|------|--------|-------|
| `apps/api/app/api/controlled_signup_oauth_e2e_test.go` (or chosen path) | create | Integration harness with fake Google server and disposable Postgres |
| `apps/api/business/auth/` | inspect/reuse | Use real `GoogleOAuthProvider` with injectable URLs |
| `infra/scripts/run-controlled-signup-oauth-e2e.sh` | create (optional) | Localhost-only wrapper around `go test` |
| `specs/changes/VOC-092-.../t00-evidence.md` | create | Scrubbed test run output |

Ordered steps:

1. Implement fake Google OAuth HTTP server (token + userinfo) compatible with
   `GoogleOAuthProvider.exchangeCodeForToken` and `fetchUserInfo`.
2. Stand up disposable Postgres and auth repository wiring mirroring production
   auth HTTP registration (`RegisterAuth`).
3. Implement allowlisted success and unlisted 503 cases with synthetic fixtures.
4. Add redaction discipline to test logging.
5. Run `go test` locally and capture scrubbed evidence.

### T01 — Wire CI and deterministic foundation tests

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/` (CI caller or API job) | modify | Run harness `go test` on PR and develop |
| `scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs` | create | Structure, CI wiring, bypass denylist |
| `scripts/foundation/voc092-operator-docs.test.mjs` | create (or defer merge with T02) | Doc/evidence checks if split |
| `specs/changes/VOC-092-.../t01-evidence.md` | create | Green CI run URL |

Ordered steps:

1. Wire CI job with Docker service if required.
2. Add foundation tests and hook into existing test aggregation.
3. Verify fail-closed on intentional test failure in task PR validation.
4. Record CI evidence.

### T02 — Document harness boundary and remediate VOC-088-T03 evidence

| File | Action | Notes |
|------|--------|-------|
| `docs/operations/staging-controlled-signup.md` | modify | Harness boundary + human audit |
| `docs/operations/README.md` | modify (if index link needed) | Link updated section |
| `specs/changes/VOC-088-.../t03-evidence.md` | modify | Remove placeholder; expand allowlisted row |
| `scripts/foundation/voc092-operator-docs.test.mjs` | create/extend | Doc and evidence structure tests |
| `specs/changes/VOC-092-.../t02-evidence.md` | create | Remediation record |

Ordered steps:

1. Update operations documentation.
2. Remediate VOC-088 evidence per `VOC-092-D08`.
3. Run foundation doc tests.

### T03 — Record live CI and staging synthetic verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-092-.../t03-evidence.md` | create | CI + staging synthetic proof |

Ordered steps:

1. Record green CI run on `develop` with both harness cases.
2. Confirm `synthetic.staging.oauth-expected-state` green on live staging.
3. Run `pnpm validate` (or documented subset) and governance checks if paths require.

## Validation and independent verification

Deterministic (T00–T02):

```bash
go test ./apps/api/... -run ControlledSignupOAuth -count=1
node --test scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs
node --test scripts/foundation/voc092-operator-docs.test.mjs
node --test scripts/foundation/voc088-staging-readiness.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
pnpm validate
```

Live (T03):

- CI run URL on `develop` showing harness job success
- Latest `scheduled-synthetics` staging OAuth job success URL

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004. Verifier must confirm no production auth bypass and
no secrets in evidence.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` adds CI coverage; optional
  docs/evidence changes deploy with normal promotion. No standalone production
  activation step.
- **Rollout:** Immediate CI effect on next PR/develop run; staging synthetic
  unchanged.
- **Rollback trigger:** Harness flakes block CI; false positives; suspected test
  data leak; accidental production wiring change.
- **Rollback mechanism:** Revert task commits; remove CI job step; restore prior
  docs/evidence if needed.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge.
