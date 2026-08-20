# VOC-096 — Persist controlled production Google signup cohort

| Field | Value |
|-------|-------|
| Package | `VOC-096` |
| Title | Persist controlled production Google signup cohort |
| Path | `specs/changes/VOC-096-persist-controlled-production-google-signup-cohort` |
| Status | `draft` |
| Risk | `R4` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue #809 |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Production Google OAuth is enabled and blanket new signup remains disabled, but
`deploy-production.yml` sources `NEW_USER_SIGNUP_ALLOWLIST` only from the ephemeral
`workflow_dispatch` input `new_user_signup_allowlist` whose default is empty. Every
automatic push deployment (the normal post-promotion path) rewrites the runtime
cohort to empty. The encrypted repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`
already exists but is not consumed. Live public readiness reports
`controlled_signup_ready=false`, consistent with the empty runtime cohort.

## Required outcome (summary)

1. Persistent, production-scoped secret for the controlled-cohort allowlist; automatic
   and manual production deploys synchronize it without printing email addresses.
2. `NEW_USER_SIGNUP_ENABLED` stays `false`; only explicitly admitted emails sign up.
3. Ephemeral dispatch input removed so automatic deploys cannot silently erase the cohort.
4. Fail closed on missing, empty, multiline, or malformed cohort configuration when
   production OAuth is enabled; expose only a non-sensitive readiness boolean in logs
   and `/healthz`.
5. Production OAuth/readiness synthetic extended to fail when the controlled cohort
   is empty while controlled first-time signup is expected.
6. Deterministic workflow/config tests covering push/manual persistence, malformed/empty
   failure, boolean-only logging, and staging/production secret isolation.
7. Operator procedure and deploy-and-verify evidence for production.
8. Preserve production/staging tier isolation; do not automate real Google interactive
   login in CI.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Persist production allowlist secret and remove ephemeral dispatch input | — |
| T01 | Extend production OAuth/readiness synthetic for controlled signup readiness | T00 |
| T02 | Document operator procedure and production deploy-and-verify evidence | T01 |

See `tasks.md` for full task definitions.
