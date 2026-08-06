# VOC-042 — Tasks

None of the tasks below is implementation-authorized by this package. Adoption and
each task's own implementation-authorization are separate, mirroring
VOC-039/VOC-040/VOC-041's convention. `T00` is the actual fix and should land
first; `T01`'s regression check depends on `T00`'s fixed value existing to assert
against (or may be written failing-first against the pre-fix value, at the
implementer's discretion).

## VOC-042-T00 — Port-qualify web's API_BASE_URL

- Requirement source: issue #319's confirmed-by-reproduction root cause and
  suggested fix
- Acceptance criteria: `VOC-042-AC-00`, `VOC-042-AC-01`
- Status: pending
- Summary: in `infra/docker-compose.production.yml`'s `web` service
  `environment:` block, change `API_BASE_URL: https://api-production.vocanova.site`
  to `API_BASE_URL: https://api-production.vocanova.site:8443`, exactly as
  specified in `implementation-plan.md` step 1. Leave the existing explanatory
  comment as-is unless found inaccurate. Verify by reading the post-fix file
  directly and by exercising `apps/web/src/lib/env.ts`'s `getApiBaseURL()` against
  the post-fix value.

## VOC-042-T01 — Deterministic regression check for the port-qualification defect class

- Requirement source: issue #319's implicit ask (this defect class, now recurring
  a third time, should be catchable without a live production reproduction)
- Acceptance criteria: `VOC-042-AC-02`
- Status: pending
- Depends on: `VOC-042-T00` (or may be written first as a failing-first check
  against the pre-fix value, at the implementer's discretion)
- Summary: add a deterministic check (automated test or documented script, per
  `implementation-plan.md` step 2) that fails if the `web` service's
  `API_BASE_URL` in `infra/docker-compose.production.yml` is ever written without
  `:8443`, and passes against the post-fix file. Runs via `pnpm validate` or a
  narrower documented script.
