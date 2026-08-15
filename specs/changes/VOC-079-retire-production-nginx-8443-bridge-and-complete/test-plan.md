# VOC-079 — Test Plan

## VOC-079-TEST-00 — Repository Cloudflare verify-only reports remap absent

- Covers: `VOC-079-AC-00`
- Preconditions: package adopted; `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`
  present in GitHub `production` environment; VOC-072-T01 wiring on the
  dispatched ref
- Procedure:
  1. Dispatch `deploy-production.yml` with
     `voc067_cloudflare_origin_cutover=verify-only`.
  2. Inspect the cutover job log for zone resolution and remap status.
  3. If `FOUND`, run authorized `apply` then repeat verify-only; do not
     treat missing credentials as pass.
- Expected result: exit 0 and `OK: no origin rules remap production hosts to
  port 8443`; VOC-067-EV-05 status becomes `absent` with cited run URL.
- Evidence: `VOC-079-EV-00`

## VOC-079-TEST-01 — Production Compose has no nginx service

- Covers: `VOC-079-AC-01`, `VOC-079-AC-06`
- Preconditions: `VOC-079-T02` tree
- Procedure:
  1. Parse `infra/docker-compose.production.yml` (and/or run a foundation
     node test).
  2. Assert no `nginx` service, no `vocanova-production-nginx`, no
     `8081:80` / `8443:443` publish.
  3. `docker compose -f infra/docker-compose.production.yml config` succeeds
     with production-root overrides suitable for CI/local dry-run.
- Expected result: assertions pass; compose config quiet/success.
- Evidence: `VOC-079-EV-02`

## VOC-079-TEST-02 — Only shared-edge publishes 80/443; both networks attached

- Covers: `VOC-079-AC-01`, `VOC-079-AC-02`, `VOC-079-AC-06`
- Preconditions: `VOC-079-T02` tree
- Procedure:
  1. Assert `infra/docker-compose.shared-edge.yml` publishes `"80:80"` and
     `"443:443"` and attaches to `vocanova-net` and
     `vocanova-production-net`.
  2. Assert staging `infra/docker-compose.yml` and production
     `infra/docker-compose.production.yml` do **not** publish host `80`/`443`
     for nginx (staging legacy nginx may still exist in file historically —
     follow VOC-067 end-state: shared edge is the sole public 80/443
     publisher; do not reintroduce production 8081/8443).
  3. `docker compose -f infra/docker-compose.shared-edge.yml config` with
     documented local overrides succeeds.
- Expected result: single-edge port/network invariants hold.
- Evidence: `VOC-079-EV-02`

## VOC-079-TEST-03 — No operational production :8443 URLs

- Covers: `VOC-079-AC-04`, `VOC-079-AC-06`
- Preconditions: `VOC-079-T01` (and T02 invariants) tree
- Procedure:
  1. Run updated foundation tests formerly requiring `:8443`
     (`docker-compose-production-api-base-url-port.test.mjs`,
     `deploy-production-oauth-port.test.mjs`) — they must assert ordinary
     HTTPS without `:8443`.
  2. Grep/scan current operational paths (`infra/docker-compose.production.yml`,
     `.github/workflows/deploy-production.yml` readiness/smoke/config-write
     steps, current `infra/README.md` steady-state sections) for production
     `:8443` qualifications presented as current config.
  3. Allow historical package evidence and Cloudflare rollback script
     defaults that intentionally mention 8443 as the **restore** target.
- Expected result: operational config clean; historical mentions remain only
  where historically accurate or as rollback tooling.
- Evidence: `VOC-079-EV-01`, `VOC-079-EV-02`

## VOC-079-TEST-04 — Deploy isolation and shared-edge reload safeguards

- Covers: `VOC-079-AC-05`
- Preconditions: `VOC-079-T02` workflow/compose tree
- Procedure:
  1. Inspect `deploy-production.yml` / `deploy-staging.yml` for absence of
     production-nginx reload and presence of shared-edge `nginx -t` fail-closed
     reload.
  2. Confirm routine deploys do not `compose up --force-recreate` /
     recreate shared-edge as part of ordinary app deploys.
  3. Confirm production orphan-removal is project-scoped (does not target
     shared-edge project).
  4. Re-run or cite applicable secrets-boundary rehearsal expectations from
     VOC-037 / VOC-067-T03 where still installed.
- Expected result: isolation and reload semantics preserved; no cross-tier
  secret copy steps introduced.
- Evidence: `VOC-079-EV-02`

## VOC-079-TEST-05 — Live four-hostname canonical HTTPS verification

- Covers: `VOC-079-AC-02`, `VOC-079-AC-03`
- Preconditions: T02 released/deployed via normal production workflow; T00
  absent confirmed
- Procedure:
  1. Record production deploy run URL that applied the bridge-free revision.
  2. Run `infra/scripts/verify-voc067-cutover.sh` (canonical `:443` checks).
  3. Confirm on-host (via deploy logs or authorized ops evidence) that
     `vocanova-production-nginx` is gone and `8081`/`8443` are unpublished;
     `vocanova-shared-edge-nginx` remains.
- Expected result: all four external checks pass; single nginx; ports gone;
  no manual Docker cleanup claimed as the method.
- Evidence: `VOC-079-EV-03`

## VOC-079-TEST-06 — Rollback path credible

- Covers: `VOC-079-AC-07`
- Preconditions: T02/T03 evidence in progress
- Procedure:
  1. Document that redeploying the last-known-good prior revision restores
     the production nginx bridge service definition and publish ports.
  2. Confirm Cloudflare
     `cloudflare-remove-production-origin-port-remap.sh --restore` (or
     workflow `restore`) remains available and run the repository's offline
     `--restore` selftest. A redacted live restore rehearsal may be cited if
     incident response required one, but is not required solely for this test.
  3. Name rollback owner and monitoring window.
- Expected result: rollback steps are specific, reversible, and do not
  require undocumented manual SSH as the primary path.
- Evidence: `VOC-079-EV-03`

## Cross-cutting validation (every repository task PR)

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
node --test scripts/foundation/*.test.mjs
# plus compose config dry-runs for touched compose files
```

Missing production credentials, SSH, or live Cloudflare access is a recorded
limitation — **never** a pass for T00/T03 live clauses.
