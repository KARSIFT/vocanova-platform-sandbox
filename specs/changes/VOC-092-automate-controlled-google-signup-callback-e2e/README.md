# VOC-092 — Automate controlled Google signup callback E2E safely

| Field | Value |
|-------|-------|
| Package | `VOC-092` |
| Title | Automate controlled Google signup callback E2E safely |
| Path | `specs/changes/VOC-092-automate-controlled-google-signup-callback-e2e` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#769](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/769) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-088 closed the operational gaps for persistent staging controlled signup,
readiness synthetics, and failure-to-issue coverage. It still required two manual
Google sign-in checks because repository automation verifies only OAuth start,
the canonical callback URL, and `controlled_signup_ready` on `/healthz` — not the
full application-owned OAuth callback path with controlled-signup policy enforcement.

That gap means regressions like an empty runtime allowlist or callback-policy
failures can slip past CI until a human repeats interactive Google login.

## Required outcome (summary)

1. Add an ephemeral, privacy-preserving E2E harness that exercises real HTTP OAuth
   start and callback handlers, OAuth state cookie validation, the real
   `GoogleOAuthProvider` token/userinfo HTTP boundary via a local fake provider,
   database-backed user/external-identity/session behavior, and the final redirect.
2. Cover a never-before-seen synthetic allowlisted identity reaching the
   authenticated onboarding/home path while `NEW_USER_SIGNUP_ENABLED=false`.
3. Cover a never-before-seen synthetic unlisted identity receiving HTTP 503 with
   the stable `ErrSignupsDisabled` response.
4. Use only `@synthetic.vocanova.invalid` identities and ephemeral test data.
5. Do not add a public staging/production test-auth route, provider switch,
   backdoor, or runtime authentication bypass.
6. Retain the deployed staging synthetic
   `synthetic.staging.oauth-expected-state` unchanged.
7. Wire deterministic tests and CI coverage; failures surface through the existing
   operational-failure observer where applicable.
8. Document honestly that the harness is E2E for the application/provider HTTP
   contract, not Google's interactive consent/login UI; keep a periodic human
   real-provider audit procedure.
9. Remediate VOC-088-T03 non-blocking review findings (self-referential
   `reviewed_sha` placeholder; scrubbed allowlisted onboarding/home outcome).
10. Preserve staging/production database, secret, directory, Docker-network,
    deploy-user, and shared-edge isolation.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Ephemeral OAuth controlled-signup callback E2E harness | — |
| T01 | Wire CI and deterministic foundation tests | T00 |
| T02 | Document harness boundary and remediate VOC-088-T03 evidence | T01 |
| T03 | Record live CI and staging synthetic verification | T02 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
