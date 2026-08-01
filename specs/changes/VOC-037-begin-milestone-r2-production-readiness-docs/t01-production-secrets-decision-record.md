---
decision_id: VOC-037-D01
task_id: VOC-037-T01
status: proposed
decision_owner: founder
risk: R3
date: 2026-08-01
related_change: VOC-037
depends_on:
  - VOC-037-D00
---

# VOC-037-D01 — Production Secrets Management Decision Record

## Decision requested

Approve the production secrets mechanism for the production target selected in `VOC-037-D00`:

- Storage location and ownership model for production-only secrets.
- Injection path from CI/deploy workflow into the production host.
- Rotation and revocation process.
- Isolation controls proving production secrets are unreachable from preview, staging, and CI runtime.

This task defines the mechanism and verification criteria only. It does not create real credentials, write secret values, or deploy production.

## Recommendation

Recommend a hardened variant of the existing staging shape:

1. Keep secret material in a dedicated **GitHub Actions Environment** named `production`, with founder-controlled approvals and no sharing with preview/staging jobs.
2. Inject secrets only during a production deploy job into production-host files under `/opt/vocanova/infra/secrets/production/`.
3. Run production containers from those host files only (`env_file`/runtime environment), never from repository files, image build args, or committed workflow literals.
4. Use separate production credentials for every provider and service (database, AI provider, email provider, OAuth, monitoring).

This preserves operational simplicity from the current stack while adding explicit boundary controls for production data and credential handling.

## Decision details

## 1) Secret inventory and naming boundary

Production values exist only as production-tier credentials for the variable names already documented in:

- `apps/api/.env.example`
- `apps/web/.env.example`

No variable-name changes are required for `VOC-037-T01`; therefore those files are unchanged in this task.

Provider accounts and credentials must be distinct per tier:

- Production database credentials are not reused by staging.
- Production AI keys are not reused by staging.
- Production email-provider keys are not reused by staging.
- Production OAuth client credentials are not reused by staging.
- Production monitoring DSN/tokens are not reused by staging.

## 2) Storage and access model

### Control plane (CI/deploy)

- Create and use a GitHub Actions `production` environment for all production deploy jobs.
- Store production secrets only in that environment scope, never at repository-global scope.
- Restrict who can approve/use the `production` environment to founder-controlled approvers.
- Ensure preview and staging workflows do not reference `environment: production`.

### Runtime plane (production host)

- Store decrypted runtime secret files only on the production host at:
  - `/opt/vocanova/infra/secrets/production/api.env`
  - `/opt/vocanova/infra/secrets/production/web.env` (if needed)
  - `/opt/vocanova/infra/secrets/production/postgres.env`
  - `/opt/vocanova/infra/secrets/production/nginx/*` (TLS keypair if hosted similarly)
- File ownership and permission baseline:
  - owner: root (or dedicated deploy user with least privilege)
  - mode: `0600` for `*.env` and private keys
  - mode: `0700` for containing directories that hold private key material
- Never place production secrets in:
  - repository files
  - docker image layers
  - build logs
  - preview/staging host paths

## 3) Injection path

Approved injection flow:

1. Production deploy workflow obtains production-scoped secrets from GitHub Actions `production` environment after required approval.
2. Workflow connects only to the production host and writes/updates the production secrets files.
3. Workflow executes deploy/update commands that read those host files at runtime.
4. Workflow verifies health checks without printing secret values.

Disallowed flow:

- Passing production secrets as Docker build args for `apps/web` or `apps/api`.
- Writing production secrets into `docker-compose.yml`.
- Reusing staging SSH secrets or staging host paths for production.

## 4) Rotation and revocation

Minimum rotation policy:

- Rotate all production provider credentials on a fixed cadence (every 90 days) and immediately after suspected exposure.
- Rotation execution order:
  1. issue new provider credential
  2. update GitHub `production` environment secret
  3. redeploy production
  4. verify service health and critical flows
  5. revoke old credential
- Keep the previous credential active only for the shortest overlap window required to avoid outage.

Emergency revocation policy:

- If exposure is suspected, disable affected feature via existing kill switches where possible (`AI_FEATURES_ENABLED`, `EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`, `NEW_USER_SIGNUP_ENABLED`) and rotate/revoke immediately.
- Record incident and follow-up under DOC-11 incident process.

## 5) Isolation proof (inspection-based acceptance)

`VOC-037-AC-01` requires confirmation by inspection of the chosen mechanism, not assertion alone. The following checks satisfy that requirement:

1. CI/workflow inspection
   - production deploy workflow references `environment: production`.
   - preview/staging workflows do not reference `environment: production`.
   - no production secret names appear in preview/staging secret scopes.
2. Host-path inspection
   - production secret files exist only under production host path.
   - staging host path contains only staging secrets.
   - permissions on production secret files are restricted (`0600` or stricter).
3. Runtime inspection
   - production containers start from production host env files.
   - preview/staging containers cannot read production host files or credentials.
4. Negative-access rehearsal
   - preview deployment attempt to access production secret fails.
   - staging deployment attempt to access production secret fails.
   - non-production CI job attempt to read production secret fails.

## 6) Follow-up implementation scope after founder approval

This decision implies follow-up implementation tasks (outside `T01` itself):

- Add/adjust production deploy workflow to use `production` environment protection.
- Add production secret-file path conventions in `infra/` docs.
- Add an operator runbook for rotation and emergency revocation.
- Produce `VOC-037-EV-01` evidence with redacted screenshots/log excerpts proving boundary enforcement.

## Founder approval record (required to move from proposed to accepted)

- Decision: Pending founder approval
- Approved mechanism: Pending
- Required conditions: Pending
- Founder GitHub identity: Pending
- Approval link and reviewed revision: Pending
- Approval date: Pending

When founder approval is recorded, set:

- `status: accepted`
- "Approved mechanism" to the exact selected model
- "Required conditions" to any gating constraints (reviewers, cadence, runbook requirements)
