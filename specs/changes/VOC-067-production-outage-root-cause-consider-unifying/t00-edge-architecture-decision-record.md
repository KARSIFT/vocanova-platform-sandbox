---
decision_id: VOC-067-DEP-00
task_id: VOC-067-T00
status: accepted
decision_owner: founder
approved_path: shared-nginx
risk: R4
date: 2026-08-11
accepted_date: 2026-08-11
related_change: VOC-067
related_issue: 485
supersedes_partially:
  - VOC-037-D00
  - VOC-037-D01
evidence: VOC-067-EV-00
---

# VOC-067 — Edge Architecture and Lifecycle Decision Record

## Decision requested

Resolve the open dependencies `VOC-067-DEP-00` through `VOC-067-DEP-03` for
package VOC-067 (issue
[#485](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/485)):

1. **Edge design** — shared nginx on host `80`/`443` with Host/SNI routing, or
   dual nginx with hardened Cloudflare origin-port remap.
2. **Lifecycle ownership** — who may recreate vs reload the edge process; how
   per-tier deploys fail closed on bad config.
3. **Isolation supersession** — whether sharing the nginx *process* supersedes
   `VOC-037-D00`/`D01`'s separate-ports / separate-nginx clause while keeping
   config, certs, secrets, and deploy-user write scopes isolated.
4. **Cloudflare cutover** — accountable operator, ordered steps, and rollback
   for removing the production `:443 → origin :8443` remap.

This task records the decision only. It does not change compose files, deploy
workflows, or live Cloudflare settings.

## Why founder authority is required

This decision revises an accepted R4 production-hosting shape
(`VOC-037-D00`/`D01`): production today publishes `8081`/`8443` because staging
nginx already owns host `80`/`443`, and public production HTTPS depends on a
Cloudflare origin-port remap to `:8443`. The 2026-08-11 outage showed that
remap is a single point of failure even when the origin stack is healthy.
Adopting shared nginx couples both tiers' public edge into one process fault
domain during reload/cutover. Under active A-003, this is a founder R4
decision, not routine R3 delegate authority.

## Incident context (requirement source)

On 2026-08-11, `production.vocanova.site` and `api-production.vocanova.site`
returned persistent Cloudflare-edge `502 Bad Gateway` while all four production
containers were healthy and curling the host public IP on `:8443` with the
correct `Host:` header returned `200` from both on-box and external networks.
Root cause: the Cloudflare `:443 → origin :8443` remap broke; the origin on
`:8443` did not.

The same investigation confirmed both nginx containers' Docker `HEALTHCHECK`
is permanently broken (`wget http://127.0.0.1/` with no `Host:` hits the
catch-all `return 444` default server). That defect is tracked as
`VOC-067-DEP-04` / `VOC-067-T01` and is independent of the edge-architecture
choice.

## VOC-067-DEP-00 — Edge design: **shared nginx**

**Chosen path: shared nginx** on host `80`/`443`, routing by `server_name` /
SNI to each tier's upstream containers on their existing Docker networks.

**Rejected alternate: dual nginx + Cloudflare harden.** Keeping two independent
nginx containers and treating the origin-port remap as a hardened, monitored
control does not remove the outage class that caused the 2026-08-11 incident.

**Rationale:**

1. Removes the non-standard Cloudflare origin-port remap that failed while the
   origin remained healthy.
2. Aligns production with ordinary `edge :443 → origin :443`, matching staging
   and eliminating client-facing `:8443` qualifications introduced by
   `VOC-041` / `VOC-042`.
3. Preserves per-tier isolation for config fragments, TLS certs, secrets trees,
   directory trees, compose projects (except the shared edge attachment), and
   deploy-user write scopes — see DEP-02 below.

**Task impact:** `VOC-067-T02` through `VOC-067-T05` are **authorized** as
scoped in this package. The dual-nginx alternate path (`VOC-067-AC-07`) is
**not pursued**; no follow-up hardening tasks replace T02–T05 in this package.

## VOC-067-DEP-01 — Lifecycle ownership: **accepted as proposed**

The specification's shared-nginx lifecycle defaults are **accepted without
amendment**:

| Control | Decision |
| --- | --- |
| **Compose ownership** | Introduce a dedicated shared-edge compose file (name chosen in T02, proposed `infra/docker-compose.shared-edge.yml`) whose **recreate is rare and documented** — not part of ordinary app deploys. |
| **Process** | One `nginx:1.27-alpine` (or current pinned image) container binds host `80`/`443`, attached to both `vocanova-net` and `vocanova-production-net`. |
| **Config isolation** | Staging vhost fragments remain under `/opt/vocanova/infra/nginx/` (repo: `infra/nginx/`); production under `/opt/vocanova/production/nginx/` (repo: `infra/nginx-production/`). The shared main config `include`s both trees at distinct in-container paths (e.g. `/etc/nginx/conf.d/staging/*.conf` and `/etc/nginx/conf.d/production/*.conf`). **Only one `default_server` catch-all** may exist across the combined set. |
| **Cert isolation** | Each tier's `ssl_certificate` / key paths mount from that tier's secrets tree only (`infra/secrets/nginx/` vs `production/secrets/nginx/`). No cross-tier cert mount. |
| **Ordinary deploys** | `deploy-staging.yml` and `deploy-production.yml` each: (a) write only that tier's conf/certs, (b) run `nginx -t`, (c) `nginx -s reload` on success. Failed `nginx -t` **fails the deploy closed** and leaves the previous generation running (both tiers keep serving). |
| **Forbidden on routine deploy** | Neither pipeline may `compose down`, recreate, or otherwise replace the shared edge container as part of a normal deploy. |

Reload signaling must not become a backdoor into the other tier's filesystem —
prefer `docker exec` reload against a known shared-edge container name with no
cross-tree writes (T03 implements and verifies).

## VOC-067-DEP-02 — Supersession of `VOC-037-D00`/`D01` edge isolation

The founder **explicitly accepts** this package superseding
`VOC-037-D00`/`D01`'s **"separate nginx process per tier"** and **"separate
host ports for production public HTTPS (`8081`/`8443`)"** isolation clause.

**Unchanged invariants (mandatory):**

- Separate directory trees (`/opt/vocanova/infra` vs `/opt/vocanova/production`).
- Separate deploy users and SSH identities (`ubuntu` / staging vs
  `vocanova-production`).
- Separate compose projects for application stacks (postgres, api, web).
- Separate secrets trees; neither deploy user may read or write the other
  tier's secrets (`VOC-037-D01`, `rehearse-production-secrets-boundary.sh`).
- Separate upstream containers and databases — no shared postgres, api, or web
  across tiers.
- Config, TLS certs, and vhost fragments remain per-tier with per-tier
  ownership.

**What changes:** one shared nginx *process* terminates TLS on host `80`/`443`
for both tiers. Production no longer publishes `8081`/`8443` as the primary
public HTTPS path once cutover completes (T02/T04/T05).

This supersession is recorded here and in `change.yaml`
`requirement_approval_status`; it does not delete or rewrite
`t00-production-hosting-decision-record.md` — that file remains historical
evidence of why `:8443` and the Cloudflare remap existed before VOC-067.

## VOC-067-DEP-03 — Cloudflare cutover owner and ordered steps

**Operator:** repository-driven **Cloudflare API execution** (not a manual
dashboard-only change). Production GitHub Actions already holds
`PRODUCTION_CLOUDFLARE_API_TOKEN` and `PRODUCTION_CLOUDFLARE_ACCOUNT_ID` for
deploy-time secret injection; T05 implements the origin-port-remap removal via
the Cloudflare API and records evidence in `VOC-067-EV-05`.

**Accountable owner:** founder (`m-e-h-r-d-a-a-d`), with implementation and
evidence captured in `VOC-067-T05` / `VOC-067-EV-05`.

**Ordered cutover (mandatory sequence):**

1. **T02** — Shared edge live in repository and on origin; Host/SNI routing
   verified on origin `:443` for all four hostnames.
2. **T03** — Deploy workflows updated: per-tier write isolation, `nginx -t` +
   reload only, fail closed on bad config.
3. **Origin verification** — Confirm production serves correctly on origin
   `:443` (not only `:8443`) before any Cloudflare or URL-normalization change.
4. **T05 (Cloudflare API)** — Remove the production hostname origin-port
   override so edge `:443` maps to origin `:443`. Record external checks for
   staging and production web/API on ordinary HTTPS `:443`.
5. **T04** — Remove `:8443` port-qualification workarounds from compose,
   deploy-emitted URLs, and operator-facing docs **after** step 4 succeeds —
   must not race ahead of live origin `:443` or remap removal.

**Rollback (must be credible before cutover completes):**

- **Primary:** restore the Cloudflare origin-port override to `:8443` if edge
  checks fail after remap removal.
- **Secondary:** redeploy last-known-good compose/workflow digests; re-enable
  prior per-tier nginx port publishes if needed.
- **Last-known-good reference:** pre-cutover dual-nginx + remap configuration
  that served origin `:8443` successfully (as proven during the outage
  investigation).

T01 (HEALTHCHECK fix) may proceed in parallel with or after this record; it
does not depend on cutover.

## Scoped implementation impact (follow-up tasks)

The following is authorized scope implied by this decision. None of it is
implemented in T00 itself.

| Task | Scope |
| --- | --- |
| **T01** | Fix nginx `HEALTHCHECK` in both compose files (and shared-edge compose when it exists) without weakening `05-default.conf` `return 444` for unmatched Host. |
| **T02** | Shared-edge compose + nginx include layout; dual-network attachment; single `default_server`; drain production `8081`/`8443` as primary public path; update `infra/README.md`. |
| **T03** | Per-tier deploy write isolation + safe shared reload in `deploy-staging.yml` / `deploy-production.yml`. |
| **T04** | Remove production `:8443` workarounds from compose, deploy workflows, and steady-state docs — gated on origin `:443` live per cutover order above. |
| **T05** | Cloudflare API remap removal + external `:443` verification + rollback evidence (`VOC-067-EV-05`). |

## Non-goals (unchanged from package specification)

- No sharing of Postgres, API, web containers, secrets files, or env trees.
- No cross-tier deploy write into the other tier's nginx or secrets paths.
- No general Cloudflare redesign beyond removing the production origin-port remap.
- No commit of cert/key material to git.

## Founder approval record

- **Decision:** Approved — shared nginx path with lifecycle defaults, VOC-037
  edge supersession, and repository-driven Cloudflare API cutover as recorded
  above.
- **Founder GitHub identity:** `m-e-h-r-d-a-a-d`
- **Approval date:** 2026-08-11
- **Approval evidence:** `change.yaml` `requirement_approval_status`; adoption
  PR #487 / package adoption recorded 2026-08-11 (PR #493)
- **R4 classification:** Accepted as proposed for the package as a whole.

`status: accepted`. Approved path: **shared nginx**. Tasks T02–T05 authorized;
dual-nginx hardening not pursued. Conditions: config, TLS certs, secrets, and
deploy-user write access remain fully isolated per tier even though the nginx
process is shared.
