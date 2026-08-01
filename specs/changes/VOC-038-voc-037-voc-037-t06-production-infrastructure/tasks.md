# VOC-038 — Tasks

Each task below is independently implementable and reviewable in one pull request,
matching this repository's existing package convention (e.g. `VOC-037`'s
`T00`–`T06`). Ordering reflects dependency, not priority. None is
implementation-authorized by this package; adoption and each task's own
implementation-authorization are separate.

## VOC-038-T00 — Production directory tree, deploy user, and Compose project

- Requirement source: `VOC-037-D00`, `VOC-037-D01` section 4A
- Acceptance criteria: `VOC-038-AC-00`
- Tests: `VOC-038-TEST-00`
- Evidence: `VOC-038-EV-00`
- Status: pending
- Summary: Create `/opt/vocanova/production/` on the shared host (mode `0750` or
  stricter), fully separate from staging's `/opt/vocanova/infra/`. Create a
  dedicated least-privilege production deploy OS user, distinct from staging's
  deploy user, owning that tree. Add a production Compose file (e.g.
  `infra/docker-compose.production.yml`) defining the `vocanova-production`
  Compose project name and the four-service shape (`postgres`, `api`, `web`,
  `nginx`), with explicit per-service resource limits sized against the host's
  actual current staging usage plus a safety margin (see specification.md open
  question 2), and their rationale recorded in `infra/README.md`. Does not start
  or deploy the stack against real production credentials — this task provisions
  the tree, user, and compose definition only.

## VOC-038-T01 — `production` GitHub Actions environment

- Requirement source: `VOC-037-D01` INV-2
- Acceptance criteria: `VOC-038-AC-01`
- Tests: `VOC-038-TEST-01`
- Evidence: `VOC-038-EV-01`
- Status: pending
- Depends on: none (independent of `VOC-038-T00`; may be done in parallel)
- Summary: Configure a GitHub Actions environment named `production` with a
  required-reviewers protection rule naming the founder as the sole approver. Does
  not add any secret value to the environment yet — this task configures the
  control-plane scope and approval gate only; real production credential values
  remain a separate, non-repository action outside this package's authority.

## VOC-038-T02 — `deploy-production.yml` workflow

- Requirement source: `VOC-037-D00` ("Deployment automation"), `VOC-037-D01`
  section 4A
- Acceptance criteria: `VOC-038-AC-02`
- Tests: `VOC-038-TEST-02`
- Evidence: `VOC-038-EV-02`
- Status: pending
- Depends on: `VOC-038-T00`, `VOC-038-T01`
- Summary: Add `.github/workflows/deploy-production.yml`, triggered only by
  `push`/`workflow_dispatch` (never `pull_request` or `pull_request_target`),
  declaring `environment: production`, writing secret files only under
  `/opt/vocanova/production/secrets/`, and running deploy/update commands scoped
  only to the `vocanova-production` Compose project — mirroring
  `deploy-staging.yml`'s existing health-check-gated, rollback-by-redeploy-digest
  pattern (per `VOC-037-D00`'s "Preserve rollback-by-redeploy-digest behavior in
  production"), but never referencing staging's `STAGING_SSH_*` secrets, host
  path, deploy user, or Compose project. Uses the placeholder production hostname
  from specification.md open question 1 for health-check target configuration,
  flagged for founder confirmation before real DNS/Cloudflare configuration.

## VOC-038-T03 — Negative-access rehearsal (`VOC-037-D01` `INS-9`–`INS-11`)

- Requirement source: `VOC-037-D01` section 5.2 (`INS-9` through `INS-11`), section
  7 follow-up scope
- Acceptance criteria: `VOC-038-AC-03`
- Tests: `VOC-038-TEST-03`
- Evidence: `VOC-038-EV-03` (recorded alongside `VOC-037-EV-01`)
- Status: pending
- Depends on: `VOC-038-T00`, `VOC-038-T01`, `VOC-038-T02`
- Summary: Execute the rehearsal `VOC-037-D01` deferred: with staging's existing
  deploy path/user, attempt to read a file under
  `/opt/vocanova/production/secrets/` and confirm the attempt fails (permission
  denied); confirm production containers start from production-only host env
  files; confirm no shared volume, symlink, or bind mount crosses between the two
  trees. Record redacted evidence (commands and their pass/fail result, no secret
  values) as `VOC-038-EV-03`, and update
  `specs/changes/VOC-037-begin-milestone-r2-production-readiness-docs/t01-production-secrets-decision-record.md`'s
  section 5.2 to reflect that `INS-9`–`INS-11` are now executed, per
  `VOC-037-T06`'s own summary ("records `VOC-037-EV-01` alongside its own
  `EV-06`").

## VOC-038-T04 — Documentation updates

- Requirement source: `VOC-037-D00`, `VOC-037-D01`
- Acceptance criteria: `VOC-038-AC-04`
- Tests: `VOC-038-TEST-04`
- Evidence: `VOC-038-EV-04`
- Status: pending
- Depends on: `VOC-038-T00`, `VOC-038-T01`, `VOC-038-T02`, `VOC-038-T03`
- Summary: Update `infra/README.md` with the production directory tree, Compose
  project name, deploy user, and resource-limit conventions actually implemented
  by `T00`–`T03`. Add an amendment note to `docs/operations/11-devops-and-ci-cd.md`
  recording the now-provisioned production target, matching that document's
  existing amendment convention. Does not restate or duplicate `VOC-037`'s own
  decision records — links to them instead.
