# VOC-038 — Test Plan

## VOC-038-TEST-00 — Directory tree, deploy user, and Compose project isolation

- Covers: `VOC-038-AC-00`
- Preconditions: `VOC-038-T00` implemented; host access available to the
  implementer/verifier.
- Procedure:
  1. Confirm `/opt/vocanova/production/` exists with mode `0750` or stricter and no
     symlink into `/opt/vocanova/infra/`.
  2. Confirm the production deploy OS user is distinct from staging's deploy user
     (`id <production-user>` vs. `id <staging-user>` return different UIDs/groups).
  3. Run `docker compose -p vocanova-production config` against the new compose
     file and confirm it resolves the four-service shape with resource limits set
     per service.
  4. Confirm no bind mount, volume, or path in the new compose file references
     `/opt/vocanova/infra/`.
- Expected result: all four checks pass; any failure blocks `VOC-038-AC-00`.
- Evidence: `VOC-038-EV-00`.

## VOC-038-TEST-01 — `production` environment protection rule

- Covers: `VOC-038-AC-01`
- Preconditions: `VOC-038-T01` implemented; repository admin/API access available
  to the verifier.
- Procedure: inspect the `production` environment's configuration via the GitHub
  API or UI; confirm required reviewers is non-empty and names only the founder;
  confirm no workflow other than `deploy-production.yml` declares
  `environment: production`.
- Expected result: required reviewers present and correctly scoped; no other
  workflow references the environment.
- Evidence: `VOC-038-EV-01`.

## VOC-038-TEST-02 — Workflow trigger and scope isolation

- Covers: `VOC-038-AC-02`
- Preconditions: `VOC-038-T02` implemented.
- Procedure:
  1. `rg -n "^on:" -A5 .github/workflows/deploy-production.yml` and confirm only
     `push`/`workflow_dispatch` triggers are present.
  2. `rg -n "environment:" .github/workflows/deploy-production.yml` and confirm it
     declares `environment: production`.
  3. `rg -n "STAGING_SSH|infra/secrets|vocanova-production" .github/workflows/deploy-production.yml`
     and confirm no `STAGING_SSH_*` or staging path reference appears, and the
     `vocanova-production` project name is used for every compose invocation.
- Expected result: all three checks pass.
- Evidence: `VOC-038-EV-02`.

## VOC-038-TEST-03 — Negative-access rehearsal

- Covers: `VOC-038-AC-03`
- Preconditions: `VOC-038-T00`–`VOC-038-T02` implemented; a disposable or
  staging-equivalent rehearsal of the production shape available, per
  `VOC-037-D01`'s `INS-9`–`INS-11` preconditions.
- Procedure:
  1. As staging's deploy user, attempt `cat /opt/vocanova/production/secrets/*.env`
     and confirm permission is denied.
  2. Confirm production containers, once started, report the expected non-secret
     health-check response using only production-tier host env files.
  3. Confirm `find /opt/vocanova -type l` (or equivalent) reports no symlink
     crossing between `/opt/vocanova/infra/` and `/opt/vocanova/production/`.
- Expected result: permission denial in step 1; healthy response in step 2; no
  crossing symlink in step 3. This test uses no production secret value and no
  production data.
- Evidence: `VOC-038-EV-03`.

## VOC-038-TEST-04 — Documentation consistency

- Covers: `VOC-038-AC-04`
- Preconditions: `VOC-038-T04` implemented.
- Procedure: read `infra/README.md` and `docs/operations/11-devops-and-ci-cd.md`
  and confirm each documented path/project/user name matches the actual
  implementation from `VOC-038-T00`–`VOC-038-T02` exactly (no stale placeholder
  text).
- Expected result: documented values match implemented values; no leftover `TBD`
  or placeholder text remains in the amended sections.
- Evidence: `VOC-038-EV-04`.

## Rollback coverage

Every task above must remain reversible by construction (per AGENTS.md's "keep
changes reversible"): `VOC-038-T00`'s directory tree and deploy user can be removed
without affecting staging; `VOC-038-T01`'s environment can be deleted from GitHub
settings; `VOC-038-T02`'s workflow can be reverted by a normal git revert with no
production traffic depending on it yet (no production DNS points at the target
until a separate, later, founder-confirmed cutover); `VOC-038-T03` is read-only
rehearsal with no destructive step. No task in this package requires a database
migration or an irreversible action.
