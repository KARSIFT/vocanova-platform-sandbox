---
id: DOC-11
title: VocaNova DevOps and CI/CD Plan
version: 1.1
document_type: operations-plan
status: approved
owner: founder
canonical_path: docs/operations/11-devops-and-ci-cd.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-30
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-10
  - DOC-16
  - DOC-19
related_decisions: []
adoption_change: VOC-008
amendments:
  - id: VOC-032-§1-amendment
    title: "§1 target-infrastructure baseline amended to self-hosted Docker Compose + nginx on vocanova.site"
    adopted_in: VOC-032
    adopted_at: 2026-07-30
    approving_owner: founder
    resolution_recorded_in: specs/changes/VOC-032-begin-milestone-r1-staging-readiness-docs-product/change.yaml
    notes: "Supersedes the prior Render Web Service + Cloudflare Workers + Render PostgreSQL rows in §1's target-infrastructure table and the vocanova.com domain set, per VOC-032-D02 (resolved at adoption 2026-07-28, founder-gate delegation). The superseded rows are annotated, not silently deleted, consistent with this repository's existing convention for amending an approved document (see DOC-15 §17.0 and the A-003 active-authority notice in DOC-16). Detailed in §1's amendment note below."
source_files:
  - path: 10-development-workflow.md
    sha256: 7fdd38cb7f877051907cc68e0930ece507fe3466dab3e008795c2827eeb21aaf
---
# 11 — VocaNova DevOps and CI/CD Plan

## 1. Environments and infrastructure

Canonical environments: Local, Preview (per-PR, temporary, isolated, no production data/secrets),
Staging (from `develop`), Production (from `main`).

> **Amendment note (`VOC-032-§1-amendment`, adopted 2026-07-30 via VOC-032; founder as approving
> owner).** The Frontend/Backend/Database rows of the target-infrastructure table below and the
> `vocanova.com` domain set in the next paragraph were the **original (v1.0) baseline** as of
> 2026-07-21. They are **superseded as of 2026-07-30** by the real, working infrastructure the
> VOC-032 change package actually built (`T00`–`T09`): self-hosted Docker Compose + nginx on the
> founder's own 2 vCPU / 4 GB staging server, with Cloudflare used **only** for DNS, TLS, WAF,
> and CDN — not for compute. The new shape replaces Render Web Service, Cloudflare Workers via
> OpenNext, and Render PostgreSQL with the founder's own server running the four-service stack
> `postgres` + `api` (Go, Docker image) + `web` (Next.js, Docker image, `output: 'standalone'`)
> + `nginx`, and replaces the `vocanova.com` apex with `vocanova.site` (with
> `staging.vocanova.site` and `api-staging.vocanova.site` as the staging subdomains). The
> superseded rows are retained in place and marked **~~strikethrough~~** below so this section
> preserves the v1.0 historical record of what DOC-11 originally targeted, exactly as DOC-15 §17.0
> retains the A-001 prose that A-003 actually supersedes and DOC-16 retains its A-003
> active-authority notice. The amended (v1.1) baseline immediately follows.

**Original (v1.0) target infrastructure baseline** (2026-07-21 — 2026-07-29, **superseded as of
2026-07-30 by `VOC-032-§1-amendment`**; retained in place as historical record):

| Area | Decision (v1.0) | Status |
|---|---|---|
| Frontend | ~~Next.js App Router on Cloudflare Workers via OpenNext~~ | **Superseded 2026-07-30 by `VOC-032-§1-amendment`** |
| Backend | ~~Go modular monolith, Docker image, Render Web Service~~ | **Superseded 2026-07-30 by `VOC-032-§1-amendment`** |
| Database | ~~Render PostgreSQL, Frankfurt region~~ | **Superseded 2026-07-30 by `VOC-032-§1-amendment`** |
| CI/CD | GitHub Actions | Unchanged |
| Container registry | GitHub Container Registry | Unchanged |
| DNS/TLS/WAF/CDN | Cloudflare | **Narrowed 2026-07-30 by `VOC-032-§1-amendment`** to DNS/TLS/WAF/CDN only — Cloudflare no longer hosts compute (see amended baseline below) |
| Error monitoring | Sentry | Unchanged |
| Uptime monitoring | Better Stack / UptimeRobot | Unchanged |
| Harness, Terraform/OpenTofu, Cloudflare D1/KV/Durable Objects/Queues/R2 | Deferred post-MVP | Unchanged |

**Amended (v1.1) target infrastructure baseline** (2026-07-30, adopted via VOC-032; founder as
approving owner). This is the shape `T00`–`T09` actually built and that the staging environment
is currently deployed against — not a plan:

| Area | Decision (v1.1, as actually built by VOC-032 `T00`–`T09`) |
|---|---|
| Frontend | Next.js App Router, Docker image (`apps/web/Dockerfile`), `output: 'standalone'` build, served from the `web` service on the founder's own server, reachable at `https://staging.vocanova.site` in staging |
| Backend | Go modular monolith, Docker image (`apps/api/Dockerfile`), running as the `api` service on the founder's own server, reachable at `https://api-staging.vocanova.site` in staging; env-driven `LoadProductionConfig` refuses to start without a reachable database and the four DOC-11 §3 kill switches |
| Database | PostgreSQL 16 (`postgres:16-alpine`), running as the `postgres` service in the same Docker Compose stack on the founder's own server, named volume for persistence, `pg_isready` healthcheck; reachable only on the internal `vocanova-net` Docker network, never on a host port |
| CI/CD | GitHub Actions (unchanged) |
| Container registry | GitHub Container Registry (unchanged) — `ghcr.io/karsift/vocanova-api:sha-<sha>` and `ghcr.io/karsift/vocanova-web:sha-<sha>` per DOC-11 §2's existing build-once / promote-by-digest model |
| Orchestration | Docker Compose (`infra/docker-compose.yml`) — four services (`postgres` + `api` + `web` + `nginx`) on a single internal network (`vocanova-net`); only `nginx` publishes host ports `80`/`443` |
| Reverse proxy / TLS termination | `nginx` service (the only host-published service) terminates TLS using a Cloudflare-issued origin certificate (`VOC-032-DEP-01`); real client IP restored from `CF-Connecting-IP` only when the connection genuinely originates from Cloudflare's published IP ranges (never `0.0.0.0/0`); `staging.vocanova.site` → `web`, `api-staging.vocanova.site` → `api`; plain-`80` redirects to `443`; standard security headers set on every response |
| DNS/TLS/WAF/CDN (edge) | Cloudflare (narrowed role) — DNS, TLS, WAF, and CDN only; **no Cloudflare compute** (no Workers, no OpenNext, no Cloudflare D1/KV/Durable Objects/Queues/R2 at the edge) |
| Atlas migration tooling | `apps/api/atlas.hcl` + `apps/api/scripts/migrate.sh` apply the existing `apps/api/migrations/*.sql` set; `.down.sql.example` files remain non-executable by Atlas per the existing `apps/api/migrations/README.md` rule |
| Deploy automation | `.github/workflows/deploy-staging.yml` — build + tag by commit SHA + push to `ghcr.io` on every merge to `develop`, then SSH to the staging host, pull, apply migrations via the `T06` wrapper, `docker compose up -d`, poll `/healthz`; fails closed without tearing down currently-running containers if the health check does not pass within a bounded timeout |
| Error monitoring | Sentry (unchanged) |
| Uptime monitoring | Better Stack / UptimeRobot (unchanged) |
| Harness, Terraform/OpenTofu, Cloudflare D1/KV/Durable Objects/Queues/R2 | Deferred post-MVP (unchanged) |

This table is an implementation target, not authority to procure vendors, incur spend, create
infrastructure, deploy, or release. Each such action requires its own approved change package and
the authority applicable at execution time. The v1.1 rows above describe the staging tier that
already exists as a result of VOC-032 `T00`–`T09`; production-tier deployment, RL1/RL2
technical activation, and autonomous production release remain disabled per
`docs/governance/a003-transition-state.yaml` and are not authorized by this amendment.

**Domains (v1.1, post-`VOC-032-§1-amendment`):** `vocanova.site` apex (reserved, not currently
used by the staging tier), `staging.vocanova.site` (web app, browser-facing — the staging tier's
apex), `api-staging.vocanova.site` (Go API, browser and server-side fetch target). The
`vocanova.com` / `app.vocanova.com` / `api.vocanova.com` / `staging.vocanova.com` /
`api-staging.vocanova.com` domain set that v1.0 named is **superseded** by this paragraph; if
the founder later wants to migrate to `vocanova.com` as the production domain, that is a
separate, founder-approved DOC-11 amendment, not an implicit consequence of this one. Separate
Google OAuth clients, AI-provider keys, and Sentry environments per environment tier; no
production secrets ever reachable from preview/staging/CI (unchanged from v1.0).

## 2. Release artifacts and deployment ordering

Every deployable candidate produces three immutable artifacts: frontend OpenNext bundle, Go API OCI
image (`ghcr.io/karsift/vocanova-api:sha-<sha>`), Atlas migration OCI image. Production deploys by
digest, never by rebuilding from source: **build once → test in staging → promote exactly to
production.** Deployment order: resolve release manifest → validate artifacts → acquire environment
lock → confirm backup readiness → migration preflight → run migration → verify → deploy API by
digest → wait for readiness → deploy frontend → verify → smoke tests → record → notify. Production
deployments are never automatically cancelled once migration work has begun.

Release authority and technical activation are separate. The required human authority
depends on the effective R0–R4 risk, RL1–RL3 release class, any predefined founder-controlled launch
event, and any actual EHR trigger. The deployment sequence may run only after the live governance
and technically enabled gates permit it; failed migrations, health checks, or smoke tests stop the
deployment and invoke the governed rollback path. See the
[canonical governance index](../governance/README.md) and [DOC-19](19-governance-reconciliation-notes.md).

## 3. Rollback

Roll forward first; rollback application code only when safe; never automatically reverse production
migrations. Frontend/backend rollback = redeploy previous known-good artifact by digest. Database
rollback is not automatic — prefer a corrective forward migration; restore from backup only when
data integrity is at risk and roll-forward is unsafe. Required kill switches:
`AI_FEATURES_ENABLED`, `EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`,
`NEW_USER_SIGNUP_ENABLED`.

## 4. Backups, monitoring, incidents

Paid managed PostgreSQL with automated backups and point-in-time recovery where available; restore
tested before launch, then monthly for the first three months, then quarterly. RPO ≤24h, RTO same
business day. Severity levels SEV1 (outage/data-integrity risk) through SEV4 (minor); every SEV1/2
records date, environment, impact, detection, root cause, actions, rollback/roll-forward decision,
follow-up. Founder owns incident decision-making during MVP; GitHub issues track technical follow-up.

## 5. Production and launch readiness (checklists, condensed)

**Production-ready** requires: DNS/TLS configured, both services deployable, migrations tested,
secrets configured, OAuth/email verified, Sentry + uptime monitoring active, smoke tests passing,
backup/restore tested, rollback and incident runbooks documented, branch protection + Dependabot +
secret scanning enabled, AI budget cap configured, no production secrets/data reachable from
non-production tiers.

**Launch-ready** additionally requires: the full core MVP journey verified end to end on mobile,
privacy policy and terms published, a support/contact path, founder alerting confirmed, and an
accepted cost budget.
