---
evidence_id: VOC-081-EV-03
task_id: VOC-081-T03
acceptance_criteria:
  - VOC-081-AC-04
  - VOC-081-AC-06
tests:
  - VOC-081-TEST-05
  - VOC-081-TEST-06
date: 2026-08-15
attempt: 1
related_change: VOC-081
gate_status: repository-complete
live_closure: VOC-081-T04
---

# VOC-081-T03 — Deploy workflow convergence evidence

Attempt 1 records **repository-side** deploy-path convergence for monitoring and
shared-edge. Live public HTTPS / WebSocket closure remains `VOC-081-T04`.

## Deliverables

| Artifact | Path |
| -------- | ---- |
| Staging monitoring + shared-edge convergence | `.github/workflows/deploy-staging.yml` |
| Production stale vhost retirement | `.github/workflows/deploy-production.yml` |
| Operator ownership docs | `infra/README.md` |
| Deploy convergence tests | `scripts/foundation/voc081-deploy-convergence.test.mjs` |
| Topology test update (production isolation) | `scripts/foundation/voc081-monitoring-topology.test.mjs` |

## VOC-081-AC-04 (repository side)

| Check | Result |
| ----- | ------ |
| Normal deploy creates/converges monitoring network + Kuma from repository Compose | PASS — `deploy-staging.yml` idempotently creates `vocanova-monitoring-net`, copies `infra/docker-compose.monitoring.yml` to `/opt/vocanova/monitoring/`, runs `docker compose … -p monitoring up -d` |
| Shared-edge network/config applied with fail-closed `nginx -t` before reload/convergence | PASS — disposable `run_shared_edge_nginx_t` + `docker exec … nginx -t` before `nginx -s reload`; controlled `compose up -d` without `--force-recreate` on routine path |
| Acceptance path does not require manual SSH, `docker network connect`, or ad hoc `docker rm` | PASS — workflow uses repository SCP + compose only |
| Routine staging/production app deploys do not own monitoring/shared-edge orphan removal | PASS — production scoped `--remove-orphans` on `vocanova-production` only; monitoring `up -d` has no `--remove-orphans`; staging app `up -d` omits `--remove-orphans` |
| Exactly one VocaNova nginx invariant preserved | PASS — no second nginx introduced; production still reloads `vocanova-shared-edge-nginx` only |

## VOC-081-AC-06 (deploy-side repository checks)

| Check | Result |
| ----- | ------ |
| Deterministic tests cover deploy convergence safeguards | PASS — `voc081-deploy-convergence.test.mjs` |
| Rollback credibility documented | PASS — release plan + `infra/README.md` backup section; live rollback owner/SHA in T04 |

## Commands inspected

```bash
node --test scripts/foundation/voc081-deploy-convergence.test.mjs
node --test scripts/foundation/voc081-monitoring-topology.test.mjs
node --test scripts/foundation/voc081-monitor-vhost.test.mjs
docker compose -f infra/docker-compose.monitoring.yml config
VOCANOVA_PRODUCTION_NGINX_CONF=$PWD/infra/nginx-production/conf.d \
VOCANOVA_PRODUCTION_NGINX_CERT=$PWD/infra/secrets/nginx/cert.pem \
VOCANOVA_PRODUCTION_NGINX_KEY=$PWD/infra/secrets/nginx/key.pem \
  docker compose -f infra/docker-compose.shared-edge.yml config
```

## Limitations (not a pass)

- No qualifying production/staging deploy run URL yet — live convergence evidence
  is `VOC-081-T04`.
- `error-monitoring.yml` unchanged (expected; AC-07 live check deferred to T04).
