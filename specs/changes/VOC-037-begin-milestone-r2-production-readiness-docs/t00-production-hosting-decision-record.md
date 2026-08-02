---
decision_id: VOC-037-D00
task_id: VOC-037-T00
status: accepted
decision_owner: founder
approved_option: A-modified (same host as staging, logically isolated, portable)
risk: R4
date: 2026-08-01
accepted_date: 2026-08-01
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

- Decision: **Approved, with a modification to Option A** — production runs on
  the **same physical host as staging** (not a separate server), rather than
  either recommended Option A (separate host, same shape) or Option B
  (managed/multi-instance platform).
- Approved option: **Option A-modified ("co-located, logically isolated")**:
  - Production runs as its own Docker Compose project on the existing
    staging server, not a second machine.
  - Full logical isolation from staging is still required: separate env
    files/secrets, separate database (own Postgres instance/container, not a
    shared instance or schema), separate ports, separate domains/hostnames,
    no shared volumes.
  - Portability is an explicit goal: the compose/env layout must not assume
    single-host colocation is permanent, so a future move to a dedicated
    production host is a clean cut (repoint DNS + redeploy), not a
    rearchitecture.
- Conditions/boundaries (founder-acknowledged risk, recorded rather than
  silently accepted): the host is a 2-vCPU/4GB server already running
  staging's full 4-service stack. Adding a second full stack (Postgres, API,
  web, nginx) means staging and production **share CPU/RAM and fault
  domain** on this host — resource contention or a bad staging deploy can
  degrade production, and vice versa, until/unless production is moved to
  its own host. T01 (secrets design) and T04 (monitoring) must account for
  this shared-host risk explicitly (e.g. resource limits per compose
  project, alerting that can distinguish which environment degraded).
- Founder GitHub identity: `m-e-h-r-d-a-a-d`
- Approval link and reviewed revision: recorded via founder-gate delegate
  conversation, 2026-08-01 (this package's T00 PR, #266, at its merged SHA)
- Approval date: 2026-08-01

`status: accepted`. Approved option: **Option A-modified (same host,
logically isolated, portable)**. Conditions: shared-host resource/fault
risk explicitly acknowledged and must be addressed by T01/T04, not treated
as resolved by this decision alone; no fixed migration deadline set, but the
compose/env layout must not create migration lock-in.

## Supersession (2026-08-01) — moved to a genuinely separate host

The founder made a second production server available (`130.185.123.152`)
and asked for a "better way" than manual same-host isolation work. The
decision above is superseded, not deleted, to keep an honest record of why
T06's implementation looks the way it does (it was originally built for
same-host colocation, then verified to work unchanged on a real separate
host — see T06's evidence).

- **New approved option: plain Option A** — production runs on its own
  dedicated host, fully separate from staging's. No shared CPU/RAM/fault
  domain; the "same physical host" conditions/boundaries clause above no
  longer applies.
- The portability goal is satisfied by construction now, not as a future
  intent — production already IS the dedicated host D00 originally
  recommended before the founder's same-host modification.
- T01's corrected 4A mechanism (separate directory tree, separate deploy
  user, separate Compose project) remains valid and in effect even though
  it's no longer strictly required for isolation on a genuinely separate
  host — kept as defense in depth, not removed, since it cost nothing to
  keep and a future consolidation (e.g. moving production back to a shared
  host for cost reasons) would need it again.
- Production host: `130.185.123.152` (Ubuntu 24.04, 2 vCPU/4GB, provisioned
  2026-08-01). Deploy user: `vocanova-production` (dedicated OS user, `docker`
  group membership, narrowly-scoped passwordless sudo for
  `mkdir`/`tar`/`chown`/`touch`/`curl`/`chmod` only — not blanket `ALL`).
- DNS: `production.vocanova.site` / `api-production.vocanova.site`, both
  Cloudflare-proxied A records to the host above.
- GitHub `production` environment created with `m-e-h-r-d-a-a-d` as required
  reviewer.
- Approval date: 2026-08-01 (founder-gate delegate, executing the founder's
  direct instruction in conversation, not a new independent R4 judgment call
  — the founder explicitly chose "second server" over continued same-host
  troubleshooting).

## Second supersession (2026-08-01) — the "second server" was staging's own host

The address given for the "genuinely separate host" above (`130.185.123.152`)
was verified live to already be running staging's real containers
(`vocanova-nginx`/`vocanova-api`/`vocanova-web`/`vocanova-postgres`,
deployed by the `ubuntu` OS user (the real value of the `STAGING_SSH_USER`
secret — an earlier revision of this document incorrectly said `deploy`,
an unused account on this host; see `VOC-037-EV-06`'s "Confirmed residual
risk" section), serving the real `staging.vocanova.site` on port 443
with real staging data). It is not a second machine — it is staging's
existing host. The founder, once informed, explicitly chose to keep
production co-located here rather than wait for an actual second machine
(see conversation record, 2026-08-01).

- **Approved option reverts to Option A-modified (same host, logically
  isolated, portable)** — the original founder decision earlier in this
  document, not the "plain Option A" text in the first supersession above.
  The shared-host CPU/RAM/fault-domain risk that decision's conditions
  section describes is real and back in effect.
- T01's corrected 4A mechanism (separate directory tree, separate deploy
  user, separate Compose project, resource limits) is not "defense in
  depth" as the first supersession claimed — it is load-bearing again,
  exactly as originally designed. **Correction (2026-08-02):** this line
  previously claimed T06's `INS-9`-`INS-11` rehearsal fully PASSed — stale.
  `INS-9`/`INS-10` pass, but `INS-11` correctly FAILS: `ubuntu` (staging's
  real deploy identity) has independent pre-existing blanket sudo, so
  directory-based isolation cannot be proven against it. Accepted as a
  founder-waived residual risk, not a pass — see `VOC-037-EV-06`'s
  "Confirmed residual risk" section for the full finding.
- **New, previously unaddressed constraint found live: port collision.**
  Staging's nginx already owns the host's ports 80/443. Production's nginx
  publishes on 8081/8443 instead (already built into
  `infra/docker-compose.production.yml`). Cloudflare proxies port 8443
  automatically for any proxied hostname without any dashboard/API change
  (one of Cloudflare's built-in alternate HTTPS ports) - so
  `https://production.vocanova.site:8443/` and
  `https://api-production.vocanova.site:8443/` work with zero Cloudflare
  configuration, but every client (browsers, the web app's own API calls,
  health checks) must include `:8443` explicitly. This is now baked into
  `deploy-production.yml`'s defaults and into
  `NEXT_PUBLIC_API_BASE_URL`/`BASE_URL`/`OAUTH_REDIRECT_URI` at deploy time.
- TLS: a real Cloudflare Origin CA certificate (`*.vocanova.site` +
  `vocanova.site`, 15-year validity) was installed at
  `/opt/vocanova/production/secrets/nginx/{cert,key}.pem` — not the
  throwaway self-signed cert T06's code originally generated for
  verification, which Cloudflare's strict SSL mode correctly rejected
  (`526` at the edge) before the real cert was installed.
- Real end-to-end verification (2026-08-01): `https://production.vocanova.site:8443/`
  → `200`; `https://api-production.vocanova.site:8443/healthz` → `200`,
  `{"status":"ok","database":"ok"}`. Full stack (postgres, api, web, nginx)
  healthy; all 13 migrations applied cleanly (the Atlas directive/duplicate-
  index issues that blocked R1's first attempts were already fixed
  upstream by `VOC-033`, confirmed live here).
- This second supersession does not reopen `D01`'s approval — its
  invariants and mechanism were correct for this exact scenario and needed
  no substantive change, only the port/TLS/config-value fixes recorded in
  T06's own evidence.
