# VOC-038 — Acceptance Criteria

## VOC-038-AC-00 — Production directory tree, deploy user, and Compose project exist and are isolated from staging

- Requirement source: `VOC-037-D00`, `VOC-037-D01`
- Tasks: `VOC-038-T00`
- Tests: `VOC-038-TEST-00`
- Evidence: `VOC-038-EV-00`
- Result: pending
- Observable outcome: `/opt/vocanova/production/` exists on the shared host, fully
  separate from `/opt/vocanova/infra/` (staging) — no shared path, symlink, or bind
  mount between the two trees. A dedicated production OS deploy user exists,
  distinct from staging's deploy user. A `vocanova-production` Docker Compose
  project runs the four-service stack (`postgres`, `api`, `web`, `nginx`) from its
  own compose file, with explicit per-service resource limits set and their
  rationale recorded in `infra/README.md`.

## VOC-038-AC-01 — `production` GitHub Actions environment exists with founder-controlled required reviewers

- Requirement source: `VOC-037-D01` INV-2
- Tasks: `VOC-038-T01`
- Tests: `VOC-038-TEST-01`
- Evidence: `VOC-038-EV-01`
- Result: pending
- Observable outcome: A GitHub Actions environment named `production` exists,
  scoped to hold production secrets at environment level (never
  repository-global), with a required-reviewers protection rule naming the founder
  and no other approver, confirmed by inspection of the environment's protection
  settings (not assertion alone).

## VOC-038-AC-02 — `deploy-production.yml` deploys only to the production tree/project and never touches staging

- Requirement source: `VOC-037-D01` section 4A ("Disallowed" list)
- Tasks: `VOC-038-T02`
- Tests: `VOC-038-TEST-02`
- Evidence: `VOC-038-EV-02`
- Result: pending
- Observable outcome: `.github/workflows/deploy-production.yml` exists, declares
  `environment: production`, is triggered only by `push`/`workflow_dispatch` (never
  `pull_request` or `pull_request_target`), writes secret files only under
  `/opt/vocanova/production/secrets/`, operates only on the `vocanova-production`
  Compose project, and contains no reference to staging's `STAGING_SSH_*` secrets,
  staging's directory tree, or staging's Compose project.

## VOC-038-AC-03 — Negative-access rehearsal proves staging cannot read production's secrets

- Requirement source: `VOC-037-D01` `INS-9` through `INS-11`
- Tasks: `VOC-038-T03`
- Tests: `VOC-038-TEST-03`
- Evidence: `VOC-038-EV-03` (recorded alongside `VOC-037-EV-01`)
- Result: pending
- Observable outcome: an executed rehearsal, with redacted evidence, showing that
  staging's deploy OS user and deploy path each fail to read any file under
  `/opt/vocanova/production/secrets/`, and that production containers start
  correctly from production-only host env files (Option A per `VOC-037-D01`
  section 4A) rather than any staging-tier value.

## VOC-038-AC-04 — Documentation reflects the actual provisioned target

- Requirement source: `VOC-037-D00`, `VOC-037-D01`
- Tasks: `VOC-038-T04`
- Tests: `VOC-038-TEST-04`
- Evidence: `VOC-038-EV-04`
- Result: pending
- Observable outcome: `infra/README.md` documents the production directory tree,
  Compose project name, deploy user, and resource-limit conventions; and
  `docs/operations/11-devops-and-ci-cd.md` carries an amendment note recording the
  now-provisioned production target, consistent with its existing amendment
  convention (e.g. the referenced `VOC-032-section-1-amendment`).
