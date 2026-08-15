---
evidence_id: VOC-081-EV-02
task_id: VOC-081-T02
acceptance_criteria:
  - VOC-081-AC-02
  - VOC-081-AC-03
tests:
  - VOC-081-TEST-03
  - VOC-081-TEST-04
date: 2026-08-15
related_change: VOC-081
accountable_owner: unassigned
gate_status: repository-complete-live-access-and-https-pending
live_converge_claimed: false
---

# VOC-081-T02 — monitor.vocanova.site vhost + adopted access control

## Scope of this evidence

This task delivers the **repository-side** shared-edge vhost for
`monitor.vocanova.site` and encodes the adopted `VOC-081-DEP-00` access
policy. It does **not** claim deploy workflow convergence or public
`https://monitor.vocanova.site/` restoration — those are T03–T04.

## VOC-081-DEP-00 / DEP-01 decisions implemented

| Decision         | Resolution                                                                                                                                                                                                                                                                                              |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VOC-081-DEP-00` | Restore the existing **public Kuma login** through the proxied Cloudflare hostname, with Kuma authentication retained as the authorization boundary. No unverified Cloudflare Access application is assumed. Documented in `infra/monitoring/access-policy.md`. Proxied DNS alone is not authorization. |
| `VOC-081-DEP-01` | Vhost owned under `infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf`, loaded via existing `include /etc/nginx/conf.d/shared/*.conf`. Stale production `30-monitor.conf` superseded (marker + T03 removal).                                                                                       |

## Repository deliverables

| Artifact                      | Path                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------- |
| Monitor vhost                 | `infra/nginx-shared/conf.d/30-monitor.vocanova.site.conf`                       |
| Access policy                 | `infra/monitoring/access-policy.md`                                             |
| Stale vhost retirement marker | `infra/nginx-production/conf.d/30-monitor.conf.superseded`                      |
| Foundation tests              | `scripts/foundation/voc081-monitor-vhost.test.mjs`                              |
| Operator docs                 | `infra/README.md` (monitor vhost + access sections)                             |
| Shared-edge main comment      | `infra/nginx-shared/nginx.conf`                                                 |
| This evidence                 | `specs/changes/VOC-081-route-monitor-vocanova-site-through-the/t02-evidence.md` |

## VOC-081-TEST-03 — Shared edge loads monitor.vocanova.site vhost

| Assertion                                                                           | Result |
| ----------------------------------------------------------------------------------- | ------ |
| `server_name monitor.vocanova.site` in repository vhost                             | PASS   |
| Production TLS cert paths (`/etc/nginx/certs/production/…`)                         | PASS   |
| Upstream `vocanova-uptime-kuma:3001` (monitoring network DNS)                       | PASS   |
| Reverse-proxy headers (`Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`) | PASS   |
| WebSocket upgrade (`Upgrade`, `Connection` map)                                     | PASS   |
| Loaded via `nginx-shared/nginx.conf` → `shared/*.conf` include                      | PASS   |
| Not dependent on unloaded production `30-*.conf` glob                               | PASS   |
| No active `30-monitor.conf` in `infra/nginx-production/conf.d/`                     | PASS   |

## VOC-081-TEST-04 — Access exposure control is explicit

| Assertion                                                         | Result |
| ----------------------------------------------------------------- | ------ |
| Policy selects public Kuma login for `monitor.vocanova.site`      | PASS   |
| Policy states proxied DNS is not authorization                    | PASS   |
| Kuma authentication must remain enabled                           | PASS   |
| Policy does not invent an unverified Cloudflare Access dependency | PASS   |
| T04 live verification defined (HTTPS probe + login/auth boundary) | PASS   |
| Policy does not cite DNS proxying alone as the control            | PASS   |

## Validation commands

```bash
node --test scripts/foundation/voc081-monitor-vhost.test.mjs
node --test scripts/foundation/voc081-monitoring-topology.test.mjs
node --test scripts/foundation/*.test.mjs

# Disposable nginx -t (requires dev certs — see infra/README.md):
sh infra/nginx/generate-dev-cert.sh
# … production placeholder sed steps …
docker run --rm \
  -v $(pwd)/infra/nginx-shared/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v $(pwd)/infra/nginx-shared/conf.d:/etc/nginx/conf.d/shared:ro \
  -v $(pwd)/infra/nginx/conf.d:/etc/nginx/conf.d/staging:ro \
  -v /tmp/vocanova-nginx-prod-conf.d:/etc/nginx/conf.d/production:ro \
  -v $(pwd)/infra/secrets/nginx/cert.pem:/etc/nginx/certs/staging/cert.pem:ro \
  -v $(pwd)/infra/secrets/nginx/key.pem:/etc/nginx/certs/staging/key.pem:ro \
  -v $(pwd)/infra/secrets/nginx/cert.pem:/etc/nginx/certs/production/cert.pem:ro \
  -v $(pwd)/infra/secrets/nginx/key.pem:/etc/nginx/certs/production/key.pem:ro \
  nginx:1.27-alpine nginx -t
```

Foundation tests: **PASS** (this run). Disposable `nginx -t` with the
documented shared-edge mount set (including `30-monitor.vocanova.site.conf`):
**PASS** (`nginx: configuration file /etc/nginx/nginx.conf test is successful`).

## Limitations

- Live Kuma login/authentication boundary not verified here (T04).
- Host removal of stale `/opt/vocanova/production/nginx/conf.d/30-monitor.conf`
  deferred to T03 deploy convergence.
- Public HTTPS / WebSocket through Cloudflare deferred to T04.
