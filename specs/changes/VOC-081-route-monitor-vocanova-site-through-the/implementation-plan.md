# VOC-081 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status: adopted`,
`approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a founder
`approved` comment; exact-revision plan review and deterministic gates still
apply.

Additionally:

- Resolve or explicitly defer `VOC-081-DEP-00`–`DEP-02` in adoption
  evidence before T02/T03 claim completion against those decisions.
- Protected areas: `infra/`, `.github/workflows/deploy-*.yml`, monitoring
  persistent data, shared-edge single fault domain, secrets-boundary
  behavior, Cloudflare DNS/Access (ops).
- Path floor R3; proposed package risk **R4** — strengthened R4 evidence
  obligations apply; no founder-comment merge gate under A-004.
- No manual SSH, `docker network connect`, or ad hoc `docker rm` as the
  acceptance path (`VOC-081-D00`).

Any change to deploy sequencing or host layout must update
`infra/README.md` (and workflow header comments if they describe ownership)
in the same PR so docs do not claim the monitor hostname is out of scope or
that `30-monitor.conf` under production is loaded by shared edge when it is
not.

## File reconciliation and implementation sequence

1. **`VOC-081-T00`** — Add repository monitoring Compose preserving live
   identity (`monitoring` / `vocanova-uptime-kuma` /
   `/opt/vocanova/monitoring/kuma-data`); document backup/migration;
   validate `docker compose … config`; do not claim live converge yet.
2. **`VOC-081-T01`** — Introduce dedicated external monitoring network;
   attach Kuma and shared-edge compose; add deterministic topology / no-
   public-port / exclusive `80`/`443` tests; keep staging/production
   isolation intact.
3. **`VOC-081-T02`** — Add repository-owned `monitor.vocanova.site` vhost
   loaded by shared edge (per `VOC-081-DEP-01`); implement adopted access
   control (`VOC-081-DEP-00`); include WebSocket and proxy headers; update
   nginx include/mounts as needed; retire or inert stale unloaded
   `30-monitor.conf` via repository path (not SSH).
4. **`VOC-081-T03`** — Wire normal deploy workflow to create/converge
   monitoring network + Kuma and apply shared-edge network/config with
   fail-closed `nginx -t`; enforce scoped Compose ownership; update
   operator docs.
5. **`VOC-081-T04`** — After normal release/deploy, record live HTTPS +
   WebSocket evidence, single-edge/listener/network state, rollback
   rehearsal, monitoring window, and Sentry / `error-monitoring` health.

Preserve VOC-079 single-nginx invariants. Do not weaken staging write
isolation. Do not commit TLS private keys or Cloudflare/Kuma secrets.

## Validation and independent verification

Deterministic commands before claiming repository tasks complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
node --test scripts/foundation/*.test.mjs
```

Infra/workflow changes also need:

```bash
docker compose -f infra/docker-compose.monitoring.yml config
VOCANOVA_PRODUCTION_NGINX_CONF=$PWD/infra/nginx-production/conf.d \
VOCANOVA_PRODUCTION_NGINX_CERT=$PWD/infra/secrets/nginx/cert.pem \
VOCANOVA_PRODUCTION_NGINX_KEY=$PWD/infra/secrets/nginx/key.pem \
  docker compose -f infra/docker-compose.shared-edge.yml config
# Plus disposable nginx -t against the shared-edge mount set when vhosts change
# (follow infra/README.md / existing staging deploy disposable nginx -t pattern).
```

Live T04 clauses require production environment access and external
HTTPS/WebSocket checks. Missing credentials or SSH is a limitation, not a
pass.

Independent verification (per `CLAUDE.md`) must bind the exact commit SHA
per task, confirm the implementer-role occupant did not approve/merge its
own work, confirm authority model **`a004-active`**, escalate if semantic
risk exceeds the R4 proposal, and report every still-required R4 evidence
obligation, EHR, adoption, and activation gate. The implementer must not
self-approve.

## Deployment and rollback

Authorization: this package does **not** itself authorize production
deployment by draft or by adoption alone. After task merges, existing
auto-promotion / `deploy-production.yml` on `main` push behavior applies
per AGENTS.md under active A-004, unless adoption records a temporary
hold for the shared-edge network/vhost window (recommended for R4 edge
changes between T03 merge and T04 verification).

Rollout sequence (matches issue #665):

1. Adopt/authorize package; record exact revision; resolve DEP-00–02.
2. Establish repository source of truth + backup/migration plan (T00).
3. Deterministic topology validation (T01) and vhost/access (T02).
4. Deploy-path convergence (T03).
5. Deploy via normal workflow; verify public HTTPS/WebSocket and
   single-edge/container/listener state (T04).
6. Observe monitoring window; retain explicit rollback owner; record
   evidence.

Rollback trigger: shared-edge 5xx/unreachable for app tiers or monitor;
Kuma data path mismatch; accidental public publish of port 3001;
mis-scoped orphan removal; access-control misconfiguration that either
blocks intended operators indefinitely or accidentally opens the UI
contrary to DEP-00.

Rollback mechanism:

1. Redeploy last-known-good prior revision that removes the monitor vhost
   / monitoring-network membership from shared-edge (or restores prior
   shared-edge compose) while leaving `/opt/vocanova/monitoring/kuma-data`
   intact unless a data-corrupting change forces restore-from-backup.
2. Monitoring Compose can remain running on loopback-only for local ops
   during edge rollback if that matches the last-known-good topology.
3. Re-verify four app hostnames plus monitor (expect monitor 520 or
   Access-deny per rolled-back state); confirm single nginx and no
   `8081`/`8443`.

Accountable owner: named in T04 evidence (unassigned at drafting).
Last-known-good reference: post-VOC-079 single-edge tip with monitor
hostname returning Cloudflare 520 and Kuma healthy on loopback only
(VOC-079-EV-03 baseline).
