---
evidence_id: VOC-067-EV-03
task_id: VOC-067-T03
acceptance_criteria: VOC-067-AC-03, VOC-067-AC-04
tests: VOC-067-TEST-03, VOC-067-TEST-04
date: 2026-08-11
related_change: VOC-067
---

# VOC-067-T03 — Per-tier deploy write isolation and safe shared reload evidence

## Summary

Both deploy workflows now follow T00 / DEP-01 lifecycle defaults:

| Requirement | Implementation |
| --- | --- |
| Staging writes only staging nginx tree | `deploy-staging.yml` bundle + extract confined to `infra/` and `apps/`; routine reload after staging conf update |
| Production writes only production nginx tree | `deploy-production.yml` bundle extracts only under `/opt/vocanova/production/` |
| `nginx -t` before reload | Both workflows `docker exec vocanova-shared-edge-nginx nginx -t` before `nginx -s reload` |
| Fail closed on bad config | Non-zero exit on `nginx -t` failure; no reload; deploy aborts (AC-04) |
| No routine shared-edge recreate | Staging routine path reloads only; T02 bring-up runs only when container absent |
| No cross-tier filesystem writes | Production reload is docker-exec only; staging rare bring-up uses Docker read-only mounts for production preflight only |
| Production bridge reload | `deploy-production.yml` reloads `vocanova-production-nginx` after compose up (cutover `:8443` path until T05) |

## Workflow review (VOC-067-TEST-03)

### `deploy-staging.yml`

- **Writes:** `/opt/vocanova/infra/nginx/`, `nginx-shared/`, `docker-compose.shared-edge.yml` — staging tree only. Bundle guard rejects paths outside `infra/` and `apps/`.
- **Does not write:** `/opt/vocanova/production` (no extract, chown, or secret sync into production tree).
- **Shared-edge routine path:** When `vocanova-shared-edge-nginx` exists → `docker exec nginx -t` → `nginx -s reload` only.
- **Shared-edge rare path:** When container absent → T02 controlled bring-up (preflight, disposable `nginx -t`, `compose up -d` once, health wait). No `compose down` / `--force-recreate` on routine deploys.

### `deploy-production.yml`

- **Writes:** `/opt/vocanova/production/nginx/conf.d/` (and rest of production bundle) — production tree only.
- **Does not write:** `/opt/vocanova/infra` (pre-flight compose scan still rejects stray staging paths).
- **Shared-edge signal:** After host substitution on production vhosts, `docker exec vocanova-shared-edge-nginx nginx -t` + reload if running; skip with log line if staging has not brought shared edge up yet.
- **Bridge nginx:** After `compose up -d`, `docker exec vocanova-production-nginx nginx -t` + reload for the temporary `:8443` cutover container.

## Fail-closed semantics (VOC-067-TEST-04)

Both workflows use the same pattern:

```bash
if ! docker exec vocanova-shared-edge-nginx nginx -t; then
  echo "ERROR: ... deploy aborted without reload (previous config still active ...)" >&2
  exit 1
fi
docker exec vocanova-shared-edge-nginx nginx -s reload
```

Because `nginx -s reload` is never reached when `-t` fails, the running process keeps the prior in-memory generation — the other tier continues to serve.

**Limitation:** Live injection of a deliberately broken vhost on production was not executed in this implementer run (would risk production traffic). The fail-closed shape is enforced by `set -euo pipefail` and ordering reload after `-t` in both workflow scripts. Disposable rehearsal of the extracted commands on a mirror host remains operator-held.

## Secrets-boundary rehearsal

`rehearse-production-secrets-boundary.sh` is unchanged by T03. Production deploy still runs it after compose up. T03 adds only `docker exec` reload against containers the deploy user already manages — no new cross-tree read/write paths.

## Deterministic checks (2026-08-11, this working tree)

```bash
git diff --check
# exit 0

bash scripts/governance/validate-governance.sh
# (run by independent verifier on final revision)

bash scripts/governance/classify-change-risk.sh
# Expected path floor: R3 (.github/workflows/*)
```

## Documentation

`infra/README.md` — new "Per-tier deploy reload (VOC-067-T03)" subsection under port mapping / shared edge.
