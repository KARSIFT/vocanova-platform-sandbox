# VOC-042 — Acceptance Criteria

## VOC-042-AC-00 — web's API_BASE_URL is port-qualified with :8443

- Requirement source: issue #319's confirmed root cause and suggested fix
- Tasks: `VOC-042-T00`
- Tests: `VOC-042-TEST-00`
- Evidence: `VOC-042-EV-00`
- Result: pending
- Observable outcome: `infra/docker-compose.production.yml`'s `web` service
  `environment:` block sets
  `API_BASE_URL: https://api-production.vocanova.site:8443`. No other line in the
  `web` service's `environment:` block, and no other service's `environment:`
  block in this file, is changed. Verified by reading the post-fix file directly
  (a YAML config value, not a rendered script — unlike VOC-041's shell-templated
  values, no rendering step is needed to confirm this).

## VOC-042-AC-01 — apps/web/src/lib/env.ts's getApiBaseURL() now resolves a reachable, port-qualified URL server-side

- Requirement source: issue #319's root-cause chain (`env.ts`'s `getApiBaseURL()`
  preferring `API_BASE_URL` server-side, consumed by `middleware.ts`'s `/api/v1/me`
  auth check)
- Tasks: `VOC-042-T00`
- Tests: `VOC-042-TEST-01`
- Evidence: `VOC-042-EV-01`
- Result: pending
- Observable outcome: a server-side call to `getApiBaseURL()` (or a test exercising
  it with `API_BASE_URL` set to the post-fix value) returns
  `https://api-production.vocanova.site:8443`, and this value, when used as the
  base for a fetch to `/api/v1/me`, targets a host:port that actually resolves to
  the running API service, consistent with `docker-compose.production.yml`'s own
  `nginx` service already publishing `8443` (line 131) to the same API/web stack.

## VOC-042-AC-02 — A deterministic regression check catches an unqualified (portless) API_BASE_URL in this file before the next real deploy

- Requirement source: issue #319's implicit ask (the same defect class recurring a
  third time should not require a live production reproduction to catch again)
- Tasks: `VOC-042-T01`
- Tests: `VOC-042-TEST-02`
- Evidence: `VOC-042-EV-02`
- Result: pending
- Observable outcome: a deterministic check exists (as an automated test if this
  repository's existing test stack can parse this file's YAML content, or as a
  documented script otherwise) that fails if the `web` service's `API_BASE_URL` in
  `infra/docker-compose.production.yml` is ever written without `:8443`, and passes
  against the post-fix file. This check runs in this repo's normal CI (`pnpm
  validate` or a narrower documented script), not only manually.
