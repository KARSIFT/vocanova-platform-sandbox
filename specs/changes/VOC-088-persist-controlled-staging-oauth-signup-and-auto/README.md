# VOC-088 — Persist controlled staging OAuth signup and auto-file operational failures

| Field | Value |
|-------|-------|
| Package | `VOC-088` |
| Title | Persist controlled staging OAuth signup and auto-file operational failures |
| Path | `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto` |
| Status | `draft` |
| Risk | `R3` (draft proposal; path-based floor and independent verification govern) |
| Authority model | A-004 active |
| Requirement source | GitHub issue #746 |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

After the staging Google redirect URI was authorized, a real first-time Google
login reached the canonical callback but returned HTTP 503 with "new sign-ups are
disabled". Root cause: `deploy-staging.yml` sources `NEW_USER_SIGNUP_ALLOWLIST`
only from the optional `workflow_dispatch` input whose default is empty. Every
automatic push deployment rewrites the runtime allowlist to empty, silently
erasing any previously dispatched controlled cohort. No persistent GitHub secret
backs the allowlist. The staging OAuth/readiness synthetic does not assert that
controlled first-time signup is operational, and failed scheduled/deploy workflows
do not open governed GitHub issues.

## Required outcome (summary)

1. Persistent, staging-scoped secret for the controlled-cohort allowlist; automatic
   and manual deploys synchronize it without printing email addresses.
2. `NEW_USER_SIGNUP_ENABLED` stays `false`; only explicitly admitted emails sign up.
3. Ephemeral dispatch-input removed or safely constrained so automatic deploys
   cannot silently erase the cohort.
4. Fail closed on malformed config; expose only a non-sensitive readiness boolean.
5. Staging OAuth/readiness synthetic extended to fail when the controlled cohort is
   empty and controlled first-time signup is an expected capability.
6. Repository-managed standalone observer for failed synthetics and
   deploy/operational gates (staging and production), using the existing
   automation GitHub App so created issues enter the issue-to-plan loop.
7. Responsibility separation preserved (Sentry, Kuma, synthetics, failure-to-issue).
8. Deterministic tests for secret precedence, redaction, readiness, deduplication,
   App-token triggering, and environment mapping.
9. Operator procedure documented for adding/removing staging cohort members.
10. Deploy-and-verify evidence.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Persist staging allowlist secret and constrain ephemeral dispatch input | — |
| T01 | Extend staging OAuth/readiness synthetic for controlled signup readiness | T00 |
| T02 | Add governed failure-to-issue agent for scheduled synthetics and deploys | T01 |
| T03 | Document operator procedure and deploy-and-verify evidence | T02 |

See `tasks.md` for full task definitions.
