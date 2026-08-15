# VOC-079 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status: adopted`,
`approval_status` approved per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`) and R4 founder acceptance is recorded for the proposed risk.

Additionally:

- T01–T02 require `VOC-079-T00` evidence that
  `cloudflare_remap_api_status: absent` (or an adoption-amended exception —
  default: no exception).
- Protected areas: `infra/`, `.github/workflows/deploy-production.yml` /
  `deploy-staging.yml`, secrets-boundary behavior, Cloudflare production
  Origin Rules (ops only via existing scripts/workflow).
- Path floor R3; proposed package risk **R4** — founder authority required.
- No manual SSH or ad hoc `docker rm` as the acceptance path
  (`VOC-079-D01`).

Any change to deploy sequencing or host layout must update affected workflow
header comments and `infra/README.md` in the same PR so docs do not claim
the `:8443` bridge is still required steady state.

## File reconciliation and implementation sequence

1. **`VOC-079-T00`** — Ops/evidence: production `verify-only` dispatch;
   write `t00-evidence.md`; update VOC-067-EV-05 frontmatter to `absent`
   only on `OK`. If `FOUND`, authorize and run `apply`, then re-verify —
   do not retire the bridge.
2. **`VOC-079-T01`** — URL normalization in production compose +
   `deploy-production.yml` + current operator docs; invert VOC-041/VOC-042
   foundation tests to assert ordinary `:443`. Leave bridge service in
   place for this PR unless adoption explicitly combines T01+T02.
3. **`VOC-079-T02`** — Remove production compose `nginx`; add
   project-scoped orphan removal on production deploy; remove
   production-nginx validate/reload; keep shared-edge fail-closed reload;
   replace bridge-retention gate with single-edge invariant tests; update
   compose/README lifecycle comments.
4. **`VOC-079-T03`** — After normal release/deploy, record live host
   convergence, four-hostname checks, port/container absence, rollback
   rehearsal notes, monitoring window, and named rollback owner.

Preserve compatible VOC-067 shared-edge work. Do not weaken staging write
isolation. Do not commit TLS private keys.

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
docker compose -f infra/docker-compose.production.yml config
# shared-edge with documented local production path overrides:
VOCANOVA_PRODUCTION_NGINX_CONF=$PWD/infra/nginx-production/conf.d \
VOCANOVA_PRODUCTION_NGINX_CERT=$PWD/infra/secrets/nginx/cert.pem \
VOCANOVA_PRODUCTION_NGINX_KEY=$PWD/infra/secrets/nginx/key.pem \
  docker compose -f infra/docker-compose.shared-edge.yml config
```

Live T00/T03 clauses require production environment access. Missing
credentials or SSH is a limitation, not a pass.

Independent verification (per `CLAUDE.md`) must bind the exact commit SHA
per task, confirm the implementer-role occupant did not approve/merge its
own work, confirm authority model `a003-active`, escalate if semantic risk
exceeds R4 proposal needs, and report every still-required R4 / adoption /
activation gate. Codex (or current implementer) must not self-approve.

## Deployment and rollback

Authorization: this package does **not** itself authorize production
deployment by draft or by adoption alone. After task merges, existing
auto-promotion / `deploy-production.yml` on `main` push behavior applies
per AGENTS.md (2026-08-08 founder delegation), unless the adopting human
records a temporary cutover hold for the T02→T03 window (recommended for
R4 edge changes).

Rollout sequence (matches issue #624):

1. Adopt/authorize package; record exact revision.
2. T00 verify Cloudflare remap absence; update EV gate.
3. Merge repository cleanup (T01 then T02).
4. Deploy through the normal production workflow.
5. T03 verify canonical endpoints; confirm ports/container absence.
6. Observe monitoring window; retain explicit rollback owner.

Rollback trigger: either tier 5xx/unreachable on `:443`; OAuth/CORS
failure after URL normalization; shared-edge crash loop; orphan-removal
accident.

Rollback mechanism:

1. Redeploy last-known-good prior revision (recreates
   `vocanova-production-nginx` bridge definition and publishes).
2. If needed, Cloudflare
   `cloudflare-remove-production-origin-port-remap.sh --restore` (or
   workflow `restore`) to re-establish edge → origin `:8443`.
3. Re-run `infra/scripts/verify-voc067-cutover.sh` (and bridge probes only
   if the bridge revision is restored).

Accountable owner: named in T03 evidence (unassigned at drafting).
Last-known-good reference: dual-publish + remap-capable configuration that
served production on origin `:8443` during VOC-067 cutover (pre-T02 tip of
this package).
