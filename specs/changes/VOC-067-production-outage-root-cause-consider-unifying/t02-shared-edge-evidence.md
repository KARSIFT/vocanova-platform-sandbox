---
evidence_id: VOC-067-EV-02
task_id: VOC-067-T02
acceptance_criteria: VOC-067-AC-02
tests: VOC-067-TEST-02
date: 2026-08-11
related_change: VOC-067
---

# VOC-067-T02 — Shared-edge repository layout evidence

## Summary

Repository side of the shared edge is in place per T00 (shared nginx path):

| Requirement | Implementation |
| --- | --- |
| One process on host `80`/`443` | `infra/docker-compose.shared-edge.yml` → `vocanova-shared-edge-nginx` |
| Both Docker networks | External `vocanova-net` + `vocanova-production-net` |
| Isolated conf/cert mounts | Staging under `./nginx` + `./secrets/nginx`; production via `VOCANOVA_PRODUCTION_*` / `/opt/vocanova/production/...` |
| Single `default_server` in shared process | `infra/nginx-shared/conf.d/05-default.conf`; shared `nginx.conf` includes only staging/production `10-*` / `20-*` |
| Unambiguous upstreams | `vocanova-web` / `vocanova-api` and `vocanova-production-web` / `vocanova-production-api` |
| Primary path not `8081`/`8443` | Shared edge owns `80`/`443`; production keeps a **temporary** cutover bridge on `8081`/`8443` until T05 (T00 ordered cutover — Cloudflare still remaps to `:8443`) |
| Controlled bring-up | `deploy-staging.yml` bundles shared-edge files, preflights production conf readiness + disposable `nginx -t`, then hands off host `80`/`443` from legacy `vocanova-nginx` |

## Cutover dual-publish (remediation of attempt-1 High finding)

T00 keeps Cloudflare remapping production to origin `:8443` until **T05**.
Attempt 1 removed production's `:8443` publish with no dual-publish and no
deploy bring-up path — unsafe.

This revision:

1. Restores `vocanova-production-nginx` on `8081`/`8443` as an explicitly
   temporary bridge (cert mounts aligned to `/etc/nginx/certs/production/`).
2. Adds staging deploy bring-up for `vocanova-shared-edge-nginx` with
   fail-closed preflight (production conf must already use container_name
   upstreams and production cert paths) and rollback attempt to legacy
   staging nginx if shared edge fails after port handoff.
3. Documents dual-publish + removal at T05 in `infra/README.md` and compose
   headers. Remap is **not** claimed as steady-state design.

## Single `default_server` (Low finding)

- Shared process: only `infra/nginx-shared/conf.d/05-default.conf`.
- Staging `infra/nginx/conf.d/05-default.conf`: emptied of server blocks
  (comments only) so widening includes cannot reintroduce dual defaults.
- Production `05-default.conf`: restored for the temporary bridge container
  only; **not** included by shared `nginx.conf`.

## Deterministic checks (2026-08-11, this working tree)

```bash
docker compose -f infra/docker-compose.yml config --quiet
# exit 0

docker compose -f infra/docker-compose.production.yml config --quiet
# exit 0

VOCANOVA_PRODUCTION_NGINX_CONF=$PWD/infra/nginx-production/conf.d \
VOCANOVA_PRODUCTION_NGINX_CERT=$PWD/infra/secrets/nginx/cert.pem \
VOCANOVA_PRODUCTION_NGINX_KEY=$PWD/infra/secrets/nginx/key.pem \
  docker compose -f infra/docker-compose.shared-edge.yml config --quiet
# exit 0

# Placeholder-substituted production vhosts + disposable certs:
docker run --rm \
  -v $PWD/infra/nginx-shared/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v $PWD/infra/nginx-shared/conf.d:/etc/nginx/conf.d/shared:ro \
  -v $PWD/infra/nginx/conf.d:/etc/nginx/conf.d/staging:ro \
  -v /tmp/vocanova-nginx-prod-conf.d:/etc/nginx/conf.d/production:ro \
  -v $PWD/infra/secrets/nginx/cert.pem:/etc/nginx/certs/staging/cert.pem:ro \
  -v $PWD/infra/secrets/nginx/key.pem:/etc/nginx/certs/staging/key.pem:ro \
  -v $PWD/infra/secrets/nginx/cert.pem:/etc/nginx/certs/production/cert.pem:ro \
  -v $PWD/infra/secrets/nginx/key.pem:/etc/nginx/certs/production/key.pem:ro \
  nginx:1.27-alpine nginx -t
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful

git diff --check
# exit 0

bash scripts/governance/validate-governance.sh
# Repository foundation validation passed.
# Governance structure validation passed.

bash scripts/governance/classify-change-risk.sh
# Detected path-based risk floor: R3 (package remains R4 for VOC-037 supersession)
```

## Live origin Host-routing (TEST-02 limitation)

Full Host/SNI proof against origin `:443` for all four hostnames requires the
shared host with both app stacks healthy, production conf already deployed
(shared-edge-ready), and staging deploy completing the handoff. That live
access is founder/ops-held and was **not** available in this implementer run.

Recorded as a limitation, not a pass. Origin `:443` verification remains a
gate before T05 Cloudflare remap removal (T00 step 3 / T05).

## Rollback notes (repository)

- Keep Cloudflare remap + production bridge on `:8443` until T05 evidence.
- If shared-edge bring-up fails, staging deploy attempts to restart legacy
  `vocanova-nginx` after stopping the shared container.
- Last-known-good public production path during cutover: origin `:8443` via
  `vocanova-production-nginx` (same shape proven during the 2026-08-11
  outage investigation).
