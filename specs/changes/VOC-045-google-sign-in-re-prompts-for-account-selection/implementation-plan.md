# VOC-045 — Implementation Plan

## Preconditions and protected areas

Do not begin either task until this package is adopted (`change.yaml`'s
`status: adopted` and `implementation_authorized: true`), and until each
task is individually authorized/dispatched, per this repository's own
convention for prior packages (VOC-039 through VOC-044).

Protected areas touched:
- `apps/api/business/auth/google_oauth.go` and its test file - the Google
  OAuth authorization-URL construction, an authentication protected area.
- `apps/api/business/auth/service.go` and its test file - the
  `OAuthStart`/`OAuthCallback` session-issuing flow, an authentication
  protected area.

No other file may be touched by either task without a package amendment.

## File reconciliation and implementation sequence

Existing state, confirmed by reading the files during drafting:
- `apps/api/business/auth/google_oauth.go`'s `AuthURL` (lines 204-223) sets
  `prompt=consent` unconditionally on every call. This is the confirmed
  target of `VOC-045-T01`'s change.
- `apps/api/business/auth/service.go`'s `OAuthStart` (lines 205-236) calls
  `AuthURL` with no additional session/prior-consent check. `VOC-045-T01`
  may need to add such a check here, depending on which replacement
  condition (specification.md's open question 2) is chosen.
- No existing test in `apps/api/business/auth/google_oauth_test.go` or
  `apps/api/business/auth/service_test.go` currently asserts on the
  presence or absence of the `prompt` parameter (confirmed by the grep
  performed during drafting, which found no `prompt=` assertions in either
  file) - `VOC-045-T01` must add this coverage, it is not already present
  to preserve.

Ordered, reversible steps:

1. `VOC-045-T00` (investigation, no production code change): manually or
   via a disposable test harness, confirm whether the app's own session
   persists as intended across page loads/browser restarts, independent of
   Google's prompt behavior. Document findings in the task's pull request
   description. This step makes no change to `apps/api` production code and
   is trivially revertible (a documentation-only PR, if any file changes at
   all - it may result in no code change and only a written finding).
2. `VOC-045-T01` (fix): based on `T00`'s findings and the reviewing human's
   resolution of open question 2, modify `AuthURL` and/or `OAuthStart` so
   `prompt=consent` (or `select_account`) is requested only when actually
   needed, per whichever replacement condition was chosen. Update or add
   tests in both `google_oauth_test.go` and `service_test.go` accordingly.
   This step is reversible by reverting the single commit/PR; it does not
   alter any stored data.

Each step is independently reviewable: `T00` produces a finding with no (or
minimal) code change; `T01` produces the actual behavior change, gated on
`T00`'s output and the open questions' resolution.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section, for
changes touching `apps/api`):

```bash
pnpm validate   # or the narrower pnpm lint / typecheck / test / build, scoped to apps/api
```

Additionally, for `VOC-045-T01` specifically:

```bash
go test ./apps/api/business/auth/...
```

Independent verification: Claude Code reviews the exact final revision of
each task's pull request against this package's specification, acceptance
criteria, and the applicable R3 floor, per `CLAUDE.md`. Neither task may be
self-approved by its own implementer (Codex or otherwise).

## Deployment and rollback

- No task in this package is deployment-authorized. `release-plan.md`
  records that deployment/release remains a separate, later decision.
- Rollback mechanism: revert the merged commit/PR for the affected task.
  Neither task introduces a data migration, so rollback carries no data
  compatibility risk.
- Rollback trigger: if `VOC-045-T01`'s change causes any user to be signed
  in without a genuine, verified Google OAuth round-trip (a regression of
  `acceptance-criteria.md`'s `VOC-045-AC-02`), or if it weakens the CSRF
  state-token check (`VOC-045-AC-03`), revert immediately.
- Rollback owner: whoever holds deployment authority at the time, per this
  repository's standing authority model (A-003); this package does not
  itself grant that authority to anyone.
- Last-known-good reference: the `develop` branch commit immediately prior
  to `VOC-045-T01`'s merge.
