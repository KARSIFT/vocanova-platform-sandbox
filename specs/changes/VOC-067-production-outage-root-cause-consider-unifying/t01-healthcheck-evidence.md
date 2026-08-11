---
evidence_id: VOC-067-EV-01
task_id: VOC-067-T01
acceptance_criteria: VOC-067-AC-01
date: 2026-08-11
related_change: VOC-067
prior_implementation: VOC-066
prior_commit: 545a7efa415a566c5095d4555e9ca53cb726ddb1
prior_pr: 518
---

# VOC-067-T01 — Nginx HEALTHCHECK Fix Evidence

## Summary

`VOC-067-T01` requires the same nginx Docker `HEALTHCHECK` repair as
`VOC-066-T00`. That repair is **already on `develop`** (commit `545a7ef`,
merged via PR #518 on 2026-08-11). This task records satisfaction of
`VOC-067-AC-01` against the current tree; no further compose or nginx conf
edits were required for T01.

## Repository changes (already landed via VOC-066)

| File | Change |
|---|---|
| `infra/docker-compose.yml` | nginx `healthcheck.test` probes `http://127.0.0.1/healthz` instead of bare `http://127.0.0.1/` |
| `infra/docker-compose.production.yml` | same probe update |
| `infra/nginx/conf.d/05-default.conf` | exact-match `location = /healthz` returning `200 'ok'` on both `:80` and `:443` default servers |
| `infra/nginx-production/conf.d/05-default.conf` | same `/healthz` exception on the default server |

Catch-all hardening is unchanged: every other path on unrecognized `Host`
still `return 444`.

## Before state (pre-VOC-066)

Both compose files used:

```yaml
test: ["CMD-SHELL", "wget --quiet --tries=1 -O /dev/null http://127.0.0.1/ || exit 1"]
```

That probe has no `Host` header, so it always hit the `default_server` catch-all
and received `444`. Docker reported the container permanently `unhealthy` even
while real vhosts served traffic (`VOC-067-DEP-04`).

## After state (current `develop`)

Both compose files use:

```yaml
test: ["CMD-SHELL", "wget --quiet --tries=1 -O /dev/null http://127.0.0.1/healthz || exit 1"]
```

The default server exposes only an exact-match `/healthz` that returns `200`;
all other unrecognized-Host traffic still gets `444`.

## Deterministic checks run for this evidence

```bash
docker compose -f infra/docker-compose.yml config --quiet
docker compose -f infra/docker-compose.production.yml config --quiet
```

Both exited `0` on 2026-08-11.

## Disposable local verification (2026-08-11)

A throwaway `nginx:1.27-alpine` compose stack was started with the same
`/healthz` default-server pattern and the post-fix `HEALTHCHECK` command:

1. Container reached Docker health status **`healthy`** within four probe
   intervals (`start_period` 2s, `interval` 2s).
2. `wget http://127.0.0.1/healthz` inside the container returned **`HTTP/1.1
   200 OK`**.
3. A bare `wget http://127.0.0.1/` probe failed as expected (connection closed
   / no normal 2xx site response), confirming catch-all reject behavior is
   preserved.

## Live host verification (limitation)

`docker inspect … State.Health` on the real staging and production nginx
containers was **not** re-run in this CI implementer session (no SSH/host
credentials). Per `VOC-066-DEP-02`, the next normal
`deploy-staging.yml` / `deploy-production.yml` recreate that picks up the
fixed `HEALTHCHECK` is the expected live confirmation path. Until then, the
repository-side fix and disposable local probe above satisfy T01's
implementation scope; live `healthy` status is a follow-up for
`VOC-066-T02` / post-deploy verification.

## Acceptance mapping

| Criterion | Result |
|---|---|
| `VOC-067-AC-01` — HEALTHCHECK healthy when nginx is serving | **Satisfied in-repo** (probe targets `/healthz`; local disposable run reached `healthy`) |
| Catch-all `return 444` for unmatched Host preserved | **Satisfied** (`05-default.conf` unchanged except exact `/healthz`; local bare `/` probe still rejected) |
