# VOC-042 — Test Plan

## VOC-042-TEST-00 — Post-fix file content matches the expected port-qualified value

- Covers: `VOC-042-AC-00`
- Preconditions: `VOC-042-T00`'s fix applied to
  `infra/docker-compose.production.yml`.
- Procedure:
  1. Read the `web` service's `environment:` block in the post-fix file.
  2. Confirm the `API_BASE_URL` line reads exactly
     `API_BASE_URL: https://api-production.vocanova.site:8443`.
  3. Confirm no other line in the `web` service's `environment:` block, and no
     other service's `environment:` block in this file, was changed.
- Expected result: exactly one line changed, matching the expected string.
- Evidence: `VOC-042-EV-00`

## VOC-042-TEST-01 — getApiBaseURL() resolves a reachable, port-qualified URL server-side

- Covers: `VOC-042-AC-01`
- Preconditions: `VOC-042-T00`'s fix applied; `apps/web/src/lib/env.ts` unchanged
  by this package (only the input env value changes).
- Procedure:
  1. With `API_BASE_URL=https://api-production.vocanova.site:8443` set in the
     environment (directly, or via an existing/new unit test harness for
     `apps/web/src/lib/env.ts`), call `getApiBaseURL()` server-side.
  2. Confirm it returns `https://api-production.vocanova.site:8443` (i.e.
     confirms `API_BASE_URL` is still preferred over `NEXT_PUBLIC_API_BASE_URL`
     server-side, and that the returned value carries the port through
     unmodified).
  3. Confirm this host:port matches what `docker-compose.production.yml`'s own
     `nginx` service publishes (`8443:443`, line 131) for the same production
     stack, i.e. the resolved URL is actually reachable from the `web` container
     given the compose network topology.
- Expected result: `getApiBaseURL()` returns the port-qualified URL; the port
  matches nginx's published production port.
- Evidence: `VOC-042-EV-01`

## VOC-042-TEST-02 — Regression check fails pre-fix, passes post-fix, runs in CI

- Covers: `VOC-042-AC-02`
- Preconditions: `VOC-042-T01`'s check implemented.
- Procedure:
  1. Run the new check against a copy of the file with `VOC-042-T00`'s fix
     temporarily reverted (unqualified `API_BASE_URL`) and confirm it fails.
  2. Run the new check against the post-fix file and confirm it passes.
  3. Run the check as part of this repo's normal CI invocation (`pnpm validate` or
     the narrower relevant script) and confirm it is picked up.
- Expected result: fails pre-fix, passes post-fix, runs in normal CI.
- Evidence: `VOC-042-EV-02`

This package introduces no migration, so no migration-rollback test is applicable.
No accessibility surface is affected, so no accessibility test is applicable. This
package's fix narrows a currently-fail-safe-but-fully-blocking defect (denying
access to everyone) rather than widening authorization, so no separate
negative/authorization-failure test beyond confirming the resolved URL targets only
production's own API host (not a wildcard or additional host) is applicable —
`VOC-042-TEST-01` step 2's confirmation of the exact resolved value is this
package's authorization-boundary coverage.
