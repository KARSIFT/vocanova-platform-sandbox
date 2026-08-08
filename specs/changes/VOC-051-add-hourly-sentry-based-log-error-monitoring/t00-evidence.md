# VOC-051-EV-04 — T00 Sentry organization/plan precondition check

Evidence for `VOC-051-T00` and `VOC-051-TEST-09`.

## Outcome

`VOC-051-T00` is **blocked pending human confirmation** of the founder's actual
Sentry organization/plan capacity and token-scope availability.

This repository workspace has no Sentry API/UI credentials and no direct access
to the founder's Sentry organization settings. Under `AGENTS.md` safety and
scope rules, this task cannot infer or fabricate that confirmation from source
code alone.

Because `tasks.md` explicitly requires confirmation "against the founder's
actual Sentry organization" before `T01`/`T02`, this run records the blocker
rather than guessing.

## Repository-grounded precheck findings (non-authoritative)

These checks establish current repository wiring and assumptions, but do **not**
prove Sentry plan/tier capacity:

- `apps/api` already includes Sentry integration and deployment-time DSN wiring:
  - `apps/api/go.mod` depends on `github.com/getsentry/sentry-go`.
  - `apps/api/app/api/production.go` reads `SENTRY_DSN`,
    `SENTRY_ENVIRONMENT`, and `SENTRY_RELEASE`.
  - `.github/workflows/deploy-production.yml` references
    `PRODUCTION_SENTRY_DSN` and writes it into runtime env as `SENTRY_DSN`
    (with the existing both-or-neither partial-configuration guard).
- No equivalent `apps/web` Sentry DSN deployment wiring exists yet in this
  package state.
- `change.yaml` for `VOC-051` is adopted and implementation-authorized, but
  dependency `VOC-051-DEP-00` remains explicitly unresolved at drafting time:
  the real Sentry org/plan capacity must be confirmed externally.

## Blocking constraint to resolve before T01/T02

A human with access to the real Sentry organization must confirm and record:

1. Whether the current Sentry plan supports the required `apps/web` project
   layout (or the approved alternative layout if project limits apply).
2. Whether a read-only token scope sufficient for `VOC-051-T02`'s API calls is
   available, and the exact scope set to use.
3. The exact per-environment DSN/project mapping for both `apps/api` and
   `apps/web` (staging and production), to unblock downstream implementation.

Until this is recorded by an authorized human, `VOC-051-T01` and `VOC-051-T02`
must remain blocked per `implementation-plan.md` and `tasks.md`.
