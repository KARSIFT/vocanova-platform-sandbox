---
decision_id: VOC-037-D00
task_id: VOC-037-T00
status: proposed
decision_owner: founder
risk: R4
date: 2026-08-01
related_change: VOC-037
---

# VOC-037-D00 — Production Hosting/Deploy Target Decision Record

## Decision requested

Approve one production hosting/deploy target for R2:

- Option A (recommended): reuse the current staging architecture shape on a separate production host (single founder-owned server with Docker Compose + nginx, plus Cloudflare for DNS/TLS/WAF/CDN).
- Option B: switch production to a managed platform or multi-instance architecture now.

This task produces the decision record and scoped follow-up impact only. It does not provision infrastructure, purchase services, or deploy production.

## Why founder authority is required

This decision controls production architecture, spend, operational ownership, release risk, and rollback boundaries. Under the active governance model, this is a consequential R4 founder decision because it defines production launch shape and has user-trust and incident-response impact.

## Recommendation

Recommend **Option A** for R2: run production as a separate host that matches the known staging shape.

Rationale:

1. Lowest implementation risk and shortest path to an evidence-backed R2 gate because the stack already exists and is running.
2. Operational behavior is already understood (compose lifecycle, nginx routing, health checks, rollback-by-redeploy).
3. Keeps the R2 objective focused on production readiness controls (secrets, legal/privacy, kill-switch verification, monitoring, release go/no-go) rather than introducing a simultaneous platform migration.
4. Preserves an easy later migration path once real production load and failure data exist.

## Alternatives

| Option | Benefits | Costs/risks | Reversibility |
| --- | --- | --- | --- |
| **A. Separate production host, same shape as staging (recommended)** | Fastest implementation, minimal cognitive load, direct reuse of proven workflows and runbooks, small scope for T01/T03/T04/T05 | Single-host availability constraints remain, founder-operated host patching burden remains, requires disciplined secrets and monitoring hardening | High reversibility; future migration can be staged with low lock-in |
| **B. New managed or multi-instance target now** | Potentially better availability/security primitives, less host-level ops in some variants, may better fit longer-term scale | Significant added scope during R2, new workflow and rollback unknowns, higher timeline risk before go/no-go | Medium reversibility; platform-specific coupling likely |

## Scoped implementation impact after decision

The following is the exact scope this decision implies for follow-up tasks. These changes are not part of T00 itself.

### 1) Infrastructure and runtime layout

- Provision one production server (separate from staging) with capacity equal to or greater than staging baseline.
- Reuse four-service layout (`postgres`, `api`, `web`, `nginx`) with production-only environment files.
- Keep production and staging fully isolated: no shared database, volumes, env files, or credentials.

Expected repository areas for follow-up tasks:

- `infra/` (production host bootstrap docs and deployment paths)
- root `docker-compose.yml` or a production-specific compose file under `infra/` (final path to be chosen in implementation PR)

### 2) Domain and edge configuration

- Add production DNS records and hostnames (final names to be founder-confirmed during implementation).
- Configure Cloudflare proxy/TLS/WAF settings for production host.
- Keep staging and production zones/records operationally separate.

Expected repository/document areas for follow-up tasks:

- `docs/operations/11-devops-and-ci-cd.md` (amendment note documenting approved production target)
- `infra/README.md` (production topology and ownership notes)

### 3) Deployment automation

- Add a dedicated production deploy workflow separate from staging deploy automation.
- Require explicit production trigger and health-check gating.
- Preserve rollback-by-redeploy-digest behavior in production.

Expected repository area for follow-up tasks:

- `.github/workflows/deploy-production.yml` (new workflow)

### 4) Security and secrets boundary

- T01 must define production secrets storage/injection/rotation for this target.
- Confirm production secrets are unreachable from preview/staging/CI.

Expected repository areas for follow-up tasks:

- `apps/api/.env.example` and `apps/web/.env.example` (variable-name documentation only, never real values)
- deployment workflow/environment plumbing files as required by T01/T04

### 5) Verification and release path

- T03 verifies kill switches and rollback against the approved production target.
- T04 verifies Sentry and uptime alerting against the same target.
- T05 performs final R2 release PR checks plus founder go/no-go recording.

## Non-goals and constraints

- No infrastructure provisioning in this task.
- No production credentials creation in this task.
- No production deployment in this task.
- No legal/privacy policy publication in this task.

## Founder approval record (required to move from proposed to accepted)

- Decision: Pending founder approval
- Approved option: Pending
- Conditions/boundaries: Pending
- Founder GitHub identity: Pending
- Approval link and reviewed revision: Pending
- Approval date: Pending

When founder approval is recorded, set:

- `status: accepted`
- "Approved option" to Option A or Option B
- any required conditions (budget cap, uptime targets, migration deadline, etc.)
