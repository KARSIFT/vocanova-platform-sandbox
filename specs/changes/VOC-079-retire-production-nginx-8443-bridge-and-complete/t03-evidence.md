---
evidence_id: VOC-079-EV-03
task_id: VOC-079-T03
acceptance_criteria:
  - VOC-079-AC-02
  - VOC-079-AC-03
  - VOC-079-AC-07
tests:
  - VOC-079-TEST-05
  - VOC-079-TEST-06
date: 2026-08-15
attempt: 2
related_change: VOC-079
production_sha: be8de870b26547b93407b4444f5c01234a77251c
production_run: https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31884987715
rollback_owner: autonomous production workflow
last_known_good_sha: 58f803bd1fed05a5e0efae09aff017ee1195412a
gate_status: resolved
---

# VOC-079-T03 — Live single-edge convergence evidence

Attempt 2 records successful production convergence after T02 reached `main`.
It closes the live portions of AC-02, AC-03, and AC-07 without manual server
surgery.

## Qualifying deployment

PR [#662](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/662)
promoted T01/T02 to `main` as `be8de870`. Production run
[31884987715](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31884987715)
completed successfully. Its deployment logs show the repository-controlled,
project-scoped command and declarative cleanup:

```text
docker compose -f docker-compose.production.yml -p vocanova-production up -d --remove-orphans
Container vocanova-production-nginx Removing
Container vocanova-production-nginx Removed
```

The same run passed production Sentry configuration validation, image build,
canonical API/web readiness on the first attempts, and the authenticated smoke
suite, including OAuth start and authenticated endpoints.

## Post-deploy host evidence

Authenticated read-only inspection of `ubuntu@130.185.123.152` after the run
showed:

- exactly one VocaNova nginx container: `vocanova-shared-edge-nginx`, healthy,
  publishing host ports 80/443;
- no `vocanova-production-nginx` container, including stopped containers;
- no listeners on host ports 8081 or 8443;
- healthy production and staging API, web, and PostgreSQL containers;
- isolated Compose projects: `production`, `staging`, `shared-edge`,
  `monitoring`, and `hermes-agent`.

No server file, container, or listener was manually changed during inspection.

## External verification

After deployment, canonical Cloudflare-fronted checks returned:

```text
https://staging.vocanova.site/                 HTTP 200
https://api-staging.vocanova.site/healthz      HTTP 200
https://production.vocanova.site/              HTTP 200
https://api-production.vocanova.site/healthz   HTTP 200
```

Both `https://production.vocanova.site:8443/` and
`https://api-production.vocanova.site:8443/healthz` timed out with no HTTP
response, corroborating that the bridge port is no longer published.

`monitor.vocanova.site` returned Cloudflare 520. That hostname is explicitly
outside VOC-079/T03's four-host routing scope; Uptime Kuma itself was healthy on
host loopback port 3001. Its missing shared-edge route is tracked separately so
this evidence does not misrepresent it as part of the completed bridge cutover.

## Isolation and rollback

The deploy retained per-tier writable nginx/TLS/secret trees and used
fail-closed `nginx -t` before shared-edge reload. Scoped production orphan
removal did not affect staging or shared edge.

The primary rollback is to promote/redeploy last-known-good T01 revision
`58f803b`, whose production Compose still defines the bridge. Normal production
deployment recreates it; no manual Docker command or SSH edit is required. The
accountable rollback owner is the autonomous production workflow under the
repository's no-approval operating policy.

The Cloudflare cutover script's `--restore` path remains deterministic and
self-tested. Because the zone currently has no `http_request_origin`
entrypoint ruleset, live restore intentionally fails closed rather than
inventing a ruleset. If rollback requires the old origin-port remap, restore is
performed only after the bridge revision is deployed and an existing eligible
entrypoint is present; otherwise canonical 443 remains the safe route.

## Monitoring window

Monitoring began with successful run 31884987715. Initial signals are green:
four canonical public checks, container health, authenticated production smoke,
OAuth start, and production Sentry configuration validation. Scheduled
`error-monitoring` runs also remain successful. The repository continues to
observe these signals after package closure; a failure triggers the normal
last-known-good promotion/deployment path above.
