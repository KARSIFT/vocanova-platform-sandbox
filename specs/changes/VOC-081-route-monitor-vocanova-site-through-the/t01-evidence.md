---
evidence_id: VOC-081-EV-01
task_id: VOC-081-T01
acceptance_criteria:
  - VOC-081-AC-01
  - VOC-081-AC-06
tests:
  - VOC-081-TEST-01
  - VOC-081-TEST-02
date: 2026-08-15
related_change: VOC-081
accountable_owner: unassigned
gate_status: repository-complete-live-network-create-pending
live_converge_claimed: false
---

# VOC-081-T01 — Dedicated monitoring network and topology invariants

## Scope of this evidence

This task delivers the **repository-side** least-privilege Docker network
topology connecting Uptime Kuma and the shared edge. It does **not** claim
live `docker network create` on the host, monitor vhost loading, deploy
workflow convergence, or public `https://monitor.vocanova.site/` restoration —
those are T02–T04.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Monitoring network membership | `infra/docker-compose.monitoring.yml` |
| Shared-edge network membership | `infra/docker-compose.shared-edge.yml` |
| Topology foundation tests | `scripts/foundation/voc081-monitoring-topology.test.mjs` |
| Operator docs | `infra/README.md` (service table + monitoring network section) |
| This evidence | `specs/changes/VOC-081-route-monitor-vocanova-site-through-the/t01-evidence.md` |

## VOC-081-TEST-01 — Monitoring network topology declared for edge and Kuma

| Assertion | Result |
| --- | --- |
| External network `vocanova-monitoring-net` in monitoring compose | PASS |
| External network `vocanova-monitoring-net` in shared-edge compose | PASS |
| `uptime-kuma` attached to `vocanova-monitoring-net` | PASS |
| `vocanova-shared-edge-nginx` attached to `vocanova-monitoring-net` | PASS |
| Kuma not on `vocanova-net` / `vocanova-production-net` | PASS |
| No tier secret mounts in monitoring compose | PASS |
| App deploy workflows do not own monitoring project | PASS |

## VOC-081-TEST-02 — Exclusive 80/443 ownership; no public Kuma port

| Assertion | Result |
| --- | --- |
| Only shared-edge publishes host `80`/`443` | PASS (monitoring included in check) |
| Monitoring has loopback-only `127.0.0.1:3001` | PASS |
| No `0.0.0.0:3001` / `[::]:3001` / bare `3001:3001` publish | PASS |
| VOC-079 production nginx / `8081`/`8443` invariant retained | PASS (via foundation tests) |

## Compose config validation

Commands (repo root):

```bash
docker compose -f infra/docker-compose.monitoring.yml config
VOCANOVA_PRODUCTION_NGINX_CONF=$PWD/infra/nginx-production/conf.d \
VOCANOVA_PRODUCTION_NGINX_CERT=$PWD/infra/secrets/nginx/cert.pem \
VOCANOVA_PRODUCTION_NGINX_KEY=$PWD/infra/secrets/nginx/key.pem \
  docker compose -f infra/docker-compose.shared-edge.yml config
node --test scripts/foundation/voc081-monitoring-topology.test.mjs
```

Result: **PASS** (all commands exit 0).

## Live converge status

**Not performed in T01.** The external network `vocanova-monitoring-net` is
declared in repository Compose but is not created or attached on the live host
until T03 deploy convergence. Pre-T01 live baseline (issue #665): Kuma on
`monitoring_default` only; shared edge on `vocanova-net` +
`vocanova-production-net` only.

## Limitations

- No SSH or production host access in this task run.
- Network create / `docker network inspect` live evidence deferred to T03/T04.
- Monitor vhost upstream to `vocanova-uptime-kuma:3001` is T02 scope.
