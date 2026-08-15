# VOC-084 — Restore repository-managed Google sign-in in staging

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #691](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/691),
prepared for plan review and adoption under **active A-004** (proposed **R3**).

## Identity and lifecycle

- Package ID: VOC-084
- Title: Restore repository-managed Google sign-in in staging
- Canonical path:
  `specs/changes/VOC-084-restore-repository-managed-google-sign-in-in`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R3** for `.github/workflows/deploy-staging.yml` and related auth paths)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`, `implementation.authorized: false`,
  `repository_adoption_status: not-adopted`
- Target branch: `develop`
- Linked GitHub issues:
  - [#691](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/691)
    (this package's requirement source)
- Related but distinct packages:
  - [VOC-038](../VOC-038-begin-milestone-l1-controlled-launch)
    — allowlist + production OAuth adoption patterns (predecessor)
  - [VOC-041](../VOC-041-deploy-production-yml-writes-base-url-oauth)
    — production OAuth redirect/CORS deploy wiring (predecessor lessons)
  - [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying)
    — staging/production isolation + shared-edge (must be preserved)

## Why this exists

Staging Google sign-in is advertised in the UI but fails at OAuth start.
Verified 2026-08-16 evidence in issue #691:

1. `POST https://api-staging.vocanova.site/api/v1/auth/oauth/google/start`
   with `redirectUri=https://staging.vocanova.site/home` returns **HTTP 503**.
2. The production equivalent returns **HTTP 200**, targets
   `accounts.google.com`, and embeds
   `redirect_uri=https://api-production.vocanova.site/api/v1/auth/oauth/google/callback`.
3. Staging API is healthy but has `GOOGLE_OAUTH_ENABLED=false` and absent
   `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET`.
4. `NEW_USER_SIGNUP_ENABLED=false` on staging (must remain false).
5. `deploy-production.yml` already synchronizes the repository Google OAuth
   pair, rejects a partial pair, derives the canonical `/api/v1` callback,
   and enables OAuth from actual pair availability.
6. `deploy-staging.yml` does **not** synchronize that pair or canonicalize
   staging OAuth configuration.
7. The staging sign-in page shows "Continue with Google" even while the API
   reports OAuth disabled.
8. Shared-edge / public health topology is healthy; Cloudflare is not
   implicated.

Root cause: staging deployment owns neither the Google credential pair nor the
related canonical runtime flags/URI, leaving a stale disabled host
configuration while the web UI advertises the method.

## What this package does

1. **Staging OAuth sync + config + allowlist + workflow tests** (`VOC-084-T00`):
   mirror the safe production pattern in `deploy-staging.yml`; fail closed on
   a partial credential pair; write credentials only to the staging secret
   file (mode `0600`, never log values); set
   `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`;
   set `GOOGLE_OAUTH_ENABLED` from actual complete credential availability; keep
   `NEW_USER_SIGNUP_ENABLED=false` and add repository-workflow-controlled
   staging `NEW_USER_SIGNUP_ALLOWLIST` (default empty).
2. **Capability-gated sign-in UI** (`VOC-084-T01`): make the web sign-in UI
   consume a deploy-derived / auth-capability signal (prefer existing
   `/healthz` `kill_switches.oauth_enabled`) so it does not advertise Google
   when OAuth is disabled.
3. **Post-deploy OAuth-start check + Google callback evidence** (`VOC-084-T02`):
   add a live staging check that requires HTTP 200, an `accounts.google.com`
   authorization URL, and the exact staging callback `redirect_uri`, without
   following Google or completing OAuth; determine or precisely record the
   Google client callback authorization requirement.

## What this package deliberately does NOT do

- Manual server edits, database edits, or Cloudflare rule changes.
- Real OAuth login automation in CI (no callback completion).
- Production OAuth behavior changes or unrelated auth redesign.
- Google Cloud Console mutation unless access already exists (external
  callback registration is evidence/record-only when access is missing).
- Adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed **R3** (path floor R3), or raise to R4 given auth/secret
   deploy sensitivity.
2. Staging allowlist control surface (default: mirror production's
   `workflow_dispatch` input, empty by default) — `VOC-084-DEP-02`.
3. Disposition of Google client staging-callback authorization when Console
   access is unavailable — `VOC-084-DEP-01`.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R3 still requires strengthened evidence, independent
verification, and rollback credibility. `automatic_merge_allowed: true` is
set per AGENTS.md. This draft still carries no adoption or implementation
authority.
