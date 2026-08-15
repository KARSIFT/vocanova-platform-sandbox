# VOC-081 — Route monitor.vocanova.site through the repository-managed shared edge: Specification

## Objective and requirement source

Restore `https://monitor.vocanova.site` through the single repository-managed
`vocanova-shared-edge-nginx`, without reintroducing a second nginx or making
an ad hoc server change — per
[GitHub issue #665](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/665).

Primary context (issue #665 and drafting-time repo read):

| Item | Value |
|------|--------|
| Predecessor cutover | [VOC-079](../VOC-079-retire-production-nginx-8443-bridge-and-complete) — single shared edge; bridge gone |
| Shared-edge intro | [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying) |
| Live public symptom (2026-08-15) | Cloudflare HTTP **520** for `https://monitor.vocanova.site/` |
| DNS | Cloudflare proxied A → `130.185.123.152` |
| Kuma health | `vocanova-uptime-kuma` healthy; bound only to `127.0.0.1:3001` |
| Live Compose | Project `monitoring` from `/opt/vocanova/monitoring/docker-compose.yml` — **not in this repository** |
| Data path to preserve | `/opt/vocanova/monitoring/kuma-data` |
| Stale vhost | `/opt/vocanova/production/nginx/conf.d/30-monitor.conf` → `vocanova-uptime-kuma:3001` |
| Shared include gap | `infra/nginx-shared/nginx.conf` loads only staging/production `10-*.conf` / `20-*.conf` |
| Network gap | Shared edge: `vocanova-net` + `vocanova-production-net`; Kuma: `monitoring_default` only |
| Single-nginx invariant | Exactly one VocaNova nginx; ports `8081`/`8443` absent (VOC-079) |

**Objective:** after this package's implementation and normal repository
deploy, (a) `https://monitor.vocanova.site/` serves the Uptime Kuma UI through
Cloudflare on canonical HTTPS with working WebSockets, (b) exactly one
VocaNova nginx remains (`vocanova-shared-edge-nginx`), (c) Kuma stays
non-public at the host level (no `0.0.0.0` / `[::]` bind for port 3001),
(d) shared edge reaches Kuma over an explicit repository-managed
least-privilege Docker network and service DNS name, (e) deployment
converges existing Kuma data/service without data loss or manual server
mutation, (f) staging/production write and secret isolation remains intact,
and (g) access exposure is controlled per the adopted `VOC-081-DEP-00`
decision rather than assuming proxied DNS is authorization.

## Confirmed findings (drafting-time)

- `infra/nginx-shared/nginx.conf` includes shared `*.conf` plus staging and
  production `10-*.conf` / `20-*.conf` only — a production
  `30-monitor.conf` on the host is invisible to the running edge.
- `infra/docker-compose.shared-edge.yml` attaches only to `vocanova-net` and
  `vocanova-production-net`; there is no monitoring network and no monitoring
  Compose file in the repository.
- No `infra/**` references to Uptime Kuma / `monitor.vocanova.site` exist at
  drafting (beyond historical VOC-079 evidence noting the 520).
- VOC-079 preserved isolated Compose projects including `monitoring` on the
  live host; this package must converge that project from repository source
  without orphan-removing shared-edge, staging, or production.
- Sentry / hourly `error-monitoring` (VOC-051) is a separate mechanism and
  must remain unchanged and healthy.

## Scope and non-goals

In scope:

1. **Repository-managed Kuma Compose + data lifecycle (T00):** Compose
   definition preserving project name `monitoring`, container name
   `vocanova-uptime-kuma`, bind of `/opt/vocanova/monitoring/kuma-data`,
   loopback-only (or unpublished) port semantics, and an explicit
   backup/migration plan before first converge.
2. **Least-privilege network (T01):** dedicated external Docker network
   (recommended name `vocanova-monitoring-net`) joined by Kuma and shared
   edge; deterministic topology tests; no public Kuma publish.
3. **Monitor vhost + access control (T02):** repository-owned
   `monitor.vocanova.site` vhost loaded by shared edge with production TLS,
   Cloudflare client-IP handling inherited from shared config, WebSocket
   upgrade headers, and required reverse-proxy headers; implement adopted
   `VOC-081-DEP-00` access control.
4. **Deploy convergence (T03):** normal repository workflow creates/converges
   monitoring network + Kuma and applies shared-edge network/config changes
   with fail-closed `nginx -t` before reload/recreate; preserve single-nginx
   and per-project ownership boundaries.
5. **Live verification (T04):** public HTTPS + WebSocket evidence through
   Cloudflare; host/container/listener invariants; rollback rehearsal;
   confirm Sentry / `error-monitoring` unchanged and healthy.

Non-goals / explicitly excluded:

- Reintroducing `vocanova-production-nginx` or ports `8081`/`8443`.
- Replacing Sentry or Uptime Kuma.
- Unrelated application feature work.
- Manual SSH file edits, `docker network connect`, or ad hoc container
  removal as the acceptance path.
- Snapshot-then-recheck-drift promotion tasks (not applicable; this package
  introduces new infra/workflow content, then live evidence).
- Adopting or authorizing this package from within the draft.

## Risk and protected areas

Builder assessment: expected paths include `infra/*` and
`.github/workflows/deploy-*.yml`, which the path classifier floors at
**R3**. Drafting-time
`scripts/governance/classify-change-risk.sh --files-from` against the
expected list reported **Detected path-based risk floor: R3**.

This package **proposes R4** for the change as a whole because it changes
the live public edge (shared-edge networks and vhosts), restores a
Cloudflare-fronted administrative monitoring UI, and alters monitoring
Compose lifecycle on the production host. This is a **draft proposal for
the reviewing human at adoption time, not a determination**. Each task's
real file list must be re-measured; the independent verifier may raise or
lower.

Protected areas: production infrastructure (`infra/`), deployment workflows,
secrets-boundary / per-tier write isolation (must not regress VOC-037-D01 /
VOC-067-AC-03 / VOC-079-AC-05), shared-edge single fault domain, monitoring
persistent data (`kuma-data`), Cloudflare DNS/Access (ops), and TLS private
keys (never commit).

Under **active A-004**, routine R3 does not require standing
technical-steward or founder approval merely for being R3; R4 engineering-
workflow gates require no founder `approved` comment — only stronger
evidence, validation, verification, rollout, monitoring, and rollback. EHR
is not triggered by this drafting pass.

## Decisions, contradictions, security, and privacy

`VOC-081-D00` (recorded here for traceability; formal acceptance at
adoption): `monitor.vocanova.site` is restored **only** through
`vocanova-shared-edge-nginx`. No second VocaNova nginx and no host
`8081`/`8443` reintroduction. Acceptance paths are repository deploy
convergence only.

`VOC-081-D01`: Uptime Kuma remains unreachable on public host interfaces.
Port `3001` must not bind to `0.0.0.0` or `[::]`. Shared edge reaches Kuma
via repository-managed Docker network DNS (service/container name), not via
published host ports.

`VOC-081-D02`: A Cloudflare proxied DNS record is **not** authorization for
the administrative UI. Access exposure must be settled via
`VOC-081-DEP-00` before claiming AC satisfaction for public restoration.

Open questions for the reviewing human:

1. **`VOC-081-DEP-00` — Access exposure.** Recommended default: treat the
   Kuma UI as a private administrative surface; require Cloudflare Access
   (Zero Trust) for `monitor.vocanova.site`, verified with redacted
   evidence, in addition to keeping Kuma's own authentication enabled. If
   the adopting authority instead accepts public exposure, record that
   decision explicitly in adoption evidence and still forbid relying on DNS
   proxying alone as the access control.
2. **`VOC-081-DEP-01` — Vhost ownership path.** Recommended default: place
   the monitor vhost under `infra/nginx-shared/` (loaded via existing
   `include /etc/nginx/conf.d/shared/*.conf`) or a dedicated
   `infra/nginx-monitor/` tree mounted read-only into shared-edge — not as
   an unloaded `30-*.conf` under the production write tree. Stale host
   `production/nginx/conf.d/30-monitor.conf` should be retired or rendered
   inert by the repository converge path without manual SSH edits.
3. **`VOC-081-DEP-02` — Shared-edge network apply.** Recommended default:
   declare `vocanova-monitoring-net` external on monitoring + shared-edge
   compose; create it from the deploy path; apply shared-edge membership via
   controlled shared-edge project convergence with fail-closed `nginx -t`,
   never via undocumented `docker network connect`, and never as a side
   effect of ordinary app `compose up` for staging/production.
4. **Risk.** Accept proposed R4 (path floor R3).

Security / privacy: administrative monitoring UI may expose operational
status and (depending on Kuma config) monitor targets. No application
user PII migration is in scope, but uptime check metadata may be sensitive.
Do not commit Kuma credentials, Cloudflare tokens, or TLS keys. Staging and
production secret trees, deploy users, and Docker networks stay isolated;
the monitoring network must not become a shortcut that mounts or shares
tier secrets.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** Persistent Uptime Kuma data at
  `/opt/vocanova/monitoring/kuma-data` must be preserved. T00 requires an
  explicit backup/migration plan and a non-destructive first converge
  (reuse existing volume/bind; do not recreate empty data). No application
  database schema change.
- **Analytics:** None expected — evidence-backed non-applicability for
  product analytics.
- **Accessibility:** None for product UI. Kuma's upstream UI accessibility
  is out of scope; this package only reverse-proxies it.
