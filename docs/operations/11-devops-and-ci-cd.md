---
id: DOC-11
title: VocaNova DevOps and CI/CD Plan
version: 1.1
document_type: operations-plan
status: approved
owner: founder
canonical_path: docs/operations/11-devops-and-ci-cd.md
approved_at: 2026-07-21
last_reviewed_at: 2026-08-19
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
  - id: VOC-051-§1-amendment
    title: "§1 Error monitoring row extended to cover apps/web browser-side reporting and the hourly Sentry-to-GitHub-issue monitoring workflow"
    adopted_in: VOC-051
    adopted_at: 2026-08-08
    approving_owner: founder
    resolution_recorded_in: specs/changes/VOC-051-add-hourly-sentry-based-log-error-monitoring/change.yaml
    notes: "Sentry remains the error-monitoring tool §1 already named; nothing is superseded. This amendment records what VOC-051 added around it: apps/web now reports browser-side and server-side errors to Sentry (T01), Layout B gives each application a separate Sentry project per environment tier (four projects, per the package's t00-evidence.md §3), and .github/workflows/error-monitoring.yml queries all four hourly and opens one unlabeled GitHub issue per genuinely new problem so plan-from-issue drafts a governed change package from it (T02). The existing row text is annotated rather than rewritten, consistent with VOC-032-§1-amendment's convention. Detailed in §1's amendment note below."
  - id: VOC-086-§1-amendment
    title: "§1 Uptime monitoring row amended from unimplemented Better Stack / UptimeRobot to repository-managed Kuma + scheduled synthetics"
    adopted_in: VOC-086
    adopted_at: 2026-08-19
    approving_owner: autonomous-engineering (A-004)
    resolution_recorded_in: specs/changes/VOC-086-manage-monitoring-inventory/change.yaml
    notes: "Supersedes the unimplemented Better Stack / UptimeRobot uptime-monitoring choice. Availability/TLS/basic API health is Uptime Kuma (infra/monitoring/monitors.yaml, sync-monitoring.yml). Authenticated behavior is scheduled synthetics (infra/monitoring/synthetics.yaml, scheduled-synthetics.yml). Sentry remains the error-monitoring channel (VOC-051). Operator runbook: docs/operations/monitoring.md."
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
| Uptime monitoring | ~~Better Stack / UptimeRobot~~ | **Superseded 2026-08-19 by `VOC-086-§1-amendment`** |
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
| Deploy automation | `.github/workflows/deploy-staging.yml` — on `develop` **push**, runs only when changed files match the fail-closed runtime/deploy allowlist (repository-root files, `apps/**`, `packages/**`, `infra/**`, `tests/staging-e2e/**`, and the workflow/selector-test paths); documentation, governed `specs/**` package material, and other evidence-only carriers **do not** schedule the workflow. When selected (or via manual `workflow_dispatch` retry/redeploy), build + tag by commit SHA + push to `ghcr.io`, then SSH to the staging host, pull, apply migrations via the `T06` wrapper, `docker compose up -d`, poll `/healthz`; fails closed without tearing down currently-running containers if the health check does not pass within a bounded timeout |
| Error monitoring | Sentry — **extended 2026-08-08 by `VOC-051-§1-amendment`** to cover `apps/web` browser-side/server-side error reporting and an hourly Sentry-to-GitHub-issue monitoring workflow (see the amendment note below); the tool choice itself is unchanged |
| Uptime monitoring | Uptime Kuma (availability/TLS/basic API health) + scheduled synthetics (authenticated behavior) — **amended 2026-08-19 by `VOC-086-§1-amendment`**; Sentry remains the separate error-monitoring channel (VOC-051) |
| Harness, Terraform/OpenTofu, Cloudflare D1/KV/Durable Objects/Queues/R2 | Deferred post-MVP (unchanged) |

### Missing Actions activation recovery

If GitHub does not activate required workflows after an App-driven task merge or
promotion-PR creation, do not toggle PR state or create unbacked check/status
records. The
repository-managed recovery path dispatches the genuine governance, validation,
and path-required deployment workflows for an immutable SHA, waits only for
successful terminal evidence, and times out fail-closed. Its hourly wake repairs
a stranded current `develop` tip without duplicating already-successful runs;
`reconcile-release` re-enters the same promotion recovery path idempotently.
After a successful promotion merge, `develop` is advanced to that exact merge SHA
before the release audit closes; tree-equivalent integration pushes therefore do
not schedule staging when no allowlisted runtime/deploy path changed. Exceptional
governed production-target work uses `reconcile-production-change` to bind an
already merged `main` tree onto `develop` before ordinary release evaluation.
Immediate post-merge recovery is scoped to governed `agent/` task branches.
Other ways of advancing `develop` are covered by the hourly exact-tip wake.
For bounded operator recovery of the current integration tip, dispatch
`pipeline.yml` with `action=recover-integration-push`. The workflow resolves the
current `develop` SHA internally and accepts no caller-supplied target SHA, then
reuses the same genuine-workflow, exact-SHA, idempotent recovery contract.

Recovery metadata reads are fail-closed prerequisites. The merge-gate,
release-converge, and standalone recovery jobs use their short-lived
`GITHUB_TOKEN` with explicit `actions: write`, `checks: read`, `statuses: read`,
and the required Contents/Pull requests access. That job token discovers workflow
runs and dispatches only the runner's allowlisted genuine workflows. The mutation
App token remains limited to exactly Contents, Issues, and Pull requests write for
PR, issue, content, and `gh pr merge` mutations that require App identity. A
separate ephemeral guard-only App token, scoped to the current caller repository
with Administration write only, is minted immediately before production merge-guard
verification and is never passed to merge, status, issue, or content mutation
steps. Recovery does not treat a still-running release carrier as attestable
`ci / ci`; when no completed non-carrier run exists it dispatches or waits for
`promotion-pr-validation PR #<n>` to finish before attestation. When a metadata
endpoint fails, the runner emits one sanitized endpoint class
(`check_runs_read_failed`, `workflow_runs_read_failed`, or
`commit_metadata_read_failed`) and aborts before dispatch planning.

GitHub does not associate successful manual recovery runs with the promotion
PR's required ruleset contexts. The release job therefore receives the additional
`statuses: write` capability. Only after the trusted selector proves genuine
successful exact-head runs from the expected governance-policy,
repository-governance, and pipeline workflows does it revalidate the open PR and
publish same-SHA success attestations for `governance-policy`, `validate`, and
`ci / ci`. Those derived statuses link to the release run and are excluded from
future evidence selection, so they satisfy the repository ruleset but can never
replace the underlying Actions evidence. The mutation App token does not receive
Administration permission; production merge-guard verification uses the separate
guard-only token described above.

The canonical same-repository `develop` → `main` promotion PR validates capture
provenance with head/source-revision-bound `pr-validation` using the immutable PR
base/head SHAs, independent of whether the recorded capture subject commit
object is reachable in the synthetic checkout. Ordinary pull requests retain
merge-base-anchored `pr-validation` when the capture fixture is unchanged; PRs
that change the capture fixture retain strict `pr-ancestry` unless the authenticated
promotion signal applies. Non-PR dispatch recovery uses `squash-safe-push`; a weaker
same-head squash-safe dispatch is not sufficient promotion-check proof. A fork
or any other base/head branch pair cannot select the promotion exception.

This table is an implementation target, not authority to procure vendors, incur spend, create
infrastructure, deploy, or release. Each such action requires its own approved change package and
the authority applicable at execution time. The v1.1 rows above describe the staging tier that
already exists as a result of VOC-032 `T00`–`T09`; RL1/RL2 technical activation remains
disabled per `docs/governance/a003-transition-state.yaml` and is not authorized by this
amendment. **Updated 2026-08-08**: production-tier deployment and autonomous production
release, previously also listed here as disabled, are no longer disabled -
`docs/governance/a003-transition-state.yaml` itself now records both as enabled,
following the founder's explicit authorization (see `AGENTS.md`'s "Release and
deployment authority"). This amendment still does not itself authorize them; that
authorization came from a separate, later decision.

**Domains (v1.1, post-`VOC-032-§1-amendment`):** `vocanova.site` apex (reserved, not currently
used by the staging tier), `staging.vocanova.site` (web app, browser-facing — the staging tier's
apex), `api-staging.vocanova.site` (Go API, browser and server-side fetch target). The
`vocanova.com` / `app.vocanova.com` / `api.vocanova.com` / `staging.vocanova.com` /
`api-staging.vocanova.com` domain set that v1.0 named is **superseded** by this paragraph; if
the founder later wants to migrate to `vocanova.com` as the production domain, that is a
separate, founder-approved DOC-11 amendment, not an implicit consequence of this one. Separate
Google OAuth clients, AI-provider keys, and Sentry environments per environment tier; no
production secrets ever reachable from preview/staging/CI (unchanged from v1.0).

> **Amendment note (`VOC-051-§1-amendment`, adopted 2026-08-08 via VOC-051; founder as approving
> owner).** Sentry remains the error-monitoring tool both tables above already name — this
> amendment supersedes nothing and retires nothing. It records what VOC-051 built around that
> unchanged choice, so the "Error monitoring" row no longer under-describes the mechanism actually
> in place:
>
> - **`apps/web` now reports errors to Sentry** (`VOC-051-T01`), via `@sentry/nextjs` across the
>   browser, server, and edge runtimes. Previously only `apps/api` did. The SDK is a no-op when its
>   DSN is unset, matching `apps/api`'s existing behaviour, so an unconfigured environment reports
>   nothing rather than failing.
> - **The "separate Sentry environments per environment tier" requirement in the paragraph above is
>   satisfied by separate Sentry *projects*, not only by the `environment` event tag.** VOC-051
>   adopted a four-project layout — `prod-api`, `prod-web`, `stage-api`, `stage-web` under the
>   `vocanova` organization — one per application per tier, each with its own DSN held in its own
>   GitHub Actions secret. The per-tier `SENTRY_ENVIRONMENT` / `NEXT_PUBLIC_SENTRY_ENVIRONMENT` tag
>   is still set on top of that. The full layout record, including which secret feeds which project,
>   is in `specs/changes/VOC-051-add-hourly-sentry-based-log-error-monitoring/t00-evidence.md` §3.
>   This also closed a gap: `apps/api` staging had no Sentry wiring at all before VOC-051.
> - **Sentry data now feeds the governed change loop automatically.**
>   `.github/workflows/error-monitoring.yml` runs hourly, queries all four projects for unresolved
>   issues first seen in the preceding 90 minutes using a read-only (`project:read` + `event:read`)
>   Sentry token held as the `SENTRY_API_TOKEN` Actions secret, and opens one plain unlabeled GitHub
>   issue per genuinely new problem — deduplicated on a stable Sentry issue-ID marker embedded in the
>   issue body — so `pipeline.yml`'s `plan-from-issue` job drafts a change package from it. This is
>   the same route AGENTS.md's "Reporting a bug found outside the normal loop" requires of a human
>   who spots a bug; it is observability feeding that existing loop, and it changes no release,
>   deployment, or approval gate. The workflow holds `issues: write` and nothing else, and
>   deliberately has no SSH access to either host.
> - **Uptime/liveness monitoring was untouched by VOC-051** — the "Uptime monitoring" row's
>   Better Stack / UptimeRobot choice remained as stated and unimplemented here until
>   `VOC-086-§1-amendment`.
>
> **Amendment note (`VOC-086-§1-amendment`, adopted 2026-08-19 via VOC-086).** The
> "Uptime monitoring" Better Stack / UptimeRobot rows above were the **original
> unimplemented choice**. They are **superseded as of 2026-08-19** by the
> repository-managed split VOC-086 actually built:
>
> - **Availability / TLS / basic API health** is Uptime Kuma
>   (`infra/monitoring/monitors.yaml`, applied by `.github/workflows/sync-monitoring.yml`
>   over authenticated Socket.IO; never SQLite).
> - **Authenticated behavior** (OAuth expected-state, journey content, core-loop,
>   non-mutating production route sweep) is `.github/workflows/scheduled-synthetics.yml`
>   under stable IDs in `infra/monitoring/synthetics.yaml`.
> - **Application errors remain Sentry** via `.github/workflows/error-monitoring.yml`
>   (VOC-051). Do not replace that path with Kuma page checks.
>
> Operator runbook: [monitoring.md](monitoring.md).

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
