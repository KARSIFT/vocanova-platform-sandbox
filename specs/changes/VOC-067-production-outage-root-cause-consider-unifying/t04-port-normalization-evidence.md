---
evidence_id: VOC-067-EV-04
task_id: VOC-067-T04
acceptance_criteria: VOC-067-AC-05
tests: VOC-067-TEST-05
date: 2026-08-11
related_change: VOC-067
---

# VOC-067-T04 — Production :8443 port-normalization evidence

## Summary

Removed production `:8443` port-qualification workarounds introduced by
VOC-041/VOC-042 for the dual-nginx / Cloudflare origin-port remap cutover.
Steady-state production HTTPS URLs now use ordinary hostnames on `:443`.

| Area | Before (cutover) | After (T04) |
| --- | --- | --- |
| `docker-compose.production.yml` `API_BASE_URL` | `https://api-production.vocanova.site:8443` | `https://api-production.vocanova.site` |
| `deploy-production.yml` `BASE_URL` / OAuth values | `:8443`-qualified hosts | Unqualified `https://${PRODUCTION_*_HOST}` |
| Health polls / smoke suite | `:8443` URLs | Ordinary `:443` URLs |
| Production cutover bridge nginx | `vocanova-production-nginx` on `8081`/`8443` | Service removed from production compose |
| `infra/README.md` | Documented remap + dual-publish as interim | Documents `edge :443 → origin :443` steady state |

## Repository verification (VOC-067-TEST-05)

Deterministic checks run locally against this revision:

```bash
# Foundation guards (VOC-067-TEST-05)
node --test scripts/foundation/docker-compose-production-api-base-url-port.test.mjs
node --test scripts/foundation/deploy-production-oauth-port.test.mjs

# No steady-state :8443 in touched infra/workflow paths
rg ':8443' infra/docker-compose.production.yml .github/workflows/deploy-production.yml infra/README.md
# Expected: no matches (historical package docs under specs/changes/ excluded)

# Compose still parses
docker compose -f infra/docker-compose.production.yml config >/dev/null
```

## Preconditions (T00 cutover order)

T04 is gated on origin `:443` serving production and T05 Cloudflare remap
removal per `t00-edge-architecture-decision-record.md`. This evidence
documents the repository-side normalization; live external `:443` proof is
recorded in `VOC-067-EV-05` (T05).

## Rollback note

Re-introducing `:8443` qualifications or the cutover bridge requires reverting
this task's compose/workflow diff and restoring the Cloudflare origin-port
override documented in T00/T05 rollback steps.
