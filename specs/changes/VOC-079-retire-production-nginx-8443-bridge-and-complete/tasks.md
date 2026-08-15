# VOC-079 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → T03**.

## VOC-079-T00 — Cloudflare remap absence evidence and EV gate update

- Requirement source: issue #624; `VOC-079-D00`; `VOC-079-DEP-02`
- Acceptance criteria: `VOC-079-AC-00`
- Tests: `VOC-079-TEST-00`
- Evidence: `VOC-079-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending — blocked on package adoption; founder/ops production
  environment authority for `verify-only`

### Required work

1. Confirm `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` is present in the
   GitHub `production` environment (do not print the value).
2. Dispatch `deploy-production.yml` with
   `voc067_cloudflare_origin_cutover=verify-only` on a revision that already
   wires that secret (VOC-072-T01 tip or later). Prefer `develop` if `main`
   still lacks the wiring.
3. Record redacted run URL, job conclusion, and log excerpt in
   `t00-evidence.md`.
4. On `OK: no origin rules remap production hosts to port 8443`: update
   VOC-067-EV-05 frontmatter `cloudflare_remap_api_status: absent` and note
   the verifying run. On `FOUND:…`: do **not** set `absent`; either execute
   authorized `apply` / `--apply` then re-verify, or stop and escalate —
   never invent absence.
5. Independent review must FAIL any claim of AC-00 without a repository
   transcript (dashboard-only is insufficient).

### Explicitly out of scope for this task

- Removing `vocanova-production-nginx` from compose.
- Stripping `:8443` from application URLs.
- Manual Cloudflare dashboard edits as the sole proof.

## VOC-079-T01 — Canonical production URLs on ordinary :443

- Requirement source: issue #624 required change 3; VOC-067-AC-05
- Acceptance criteria: `VOC-079-AC-04`
- Tests: `VOC-079-TEST-03`
- Evidence: `VOC-079-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-079-T00` (AC-00 / `absent` gate)

### Required work

1. Remove operational `:8443` qualifications from at least:
   - `infra/docker-compose.production.yml` `API_BASE_URL`
   - `.github/workflows/deploy-production.yml` emitted `BASE_URL` /
     `OAUTH_REDIRECT_URI` / `OAUTH_REDIRECT_ALLOWLIST`
   - readiness probes, session-mint calls, and smoke URL construction that
     still target production `:8443`
   - current operator documentation under `infra/README.md` and
     `docs/operations/` that presents `:8443` / remap as steady state
2. Invert / replace foundation tests that currently **require** `:8443`
   (`docker-compose-production-api-base-url-port.test.mjs`,
   `deploy-production-oauth-port.test.mjs`) so they assert ordinary `:443`
   / no port qualification for the completed end state.
3. Do **not** rewrite historical VOC-041/VOC-042/VOC-067 evidence as if the
   bridge never existed — add forward-looking notes only.
4. Keep the production nginx bridge service in compose for this task unless
   adoption explicitly merges T01+T02; default is URLs first while bridge
   still exists so dual-publish remains available until T02.

### Explicitly out of scope for this task

- Bridge container removal (T02).
- Live cutover apply beyond what T00 already required.

## VOC-079-T02 — Retire production nginx bridge, workflow steps, and gate tests

- Requirement source: issue #624 required changes 2, 4, 5; `VOC-079-D01`;
  `VOC-079-DEP-01`
- Acceptance criteria: `VOC-079-AC-01`, `VOC-079-AC-02` (repository side),
  `VOC-079-AC-05`, `VOC-079-AC-06`
- Tests: `VOC-079-TEST-01`, `VOC-079-TEST-02`, `VOC-079-TEST-04`
- Evidence: `VOC-079-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-079-T00` and `VOC-079-T01`

### Required work

1. Remove the `nginx` service from
   `infra/docker-compose.production.yml`. Preserve production conf.d and
   TLS paths consumed read-only by shared edge. Update header comments on
   production and shared-edge compose so they no longer instruct operators
   to keep the bridge.
2. Add a **scoped** declarative orphan-removal mechanism to the production
   deploy path so `vocanova-production-nginx` disappears on normal deploy
   without manual Docker commands, without orphan-removing
   `vocanova-shared-edge-nginx` or staging services (`VOC-079-DEP-01`).
3. In `.github/workflows/deploy-production.yml`: remove validation/reload
   steps targeting `vocanova-production-nginx`; retain fail-closed
   `nginx -t` + reload for `vocanova-shared-edge-nginx`; keep routine
   deploys from recreating shared-edge; keep per-tier write isolation.
4. Replace `scripts/foundation/voc067-cutover-bridge-gate.test.mjs` (or
   supersede it with a clearly named successor included in `pnpm test`)
   with single-edge invariants:
   - production Compose has no nginx service / no `8443:443` publish
   - only shared-edge Compose publishes host `80`/`443`
   - shared edge attaches to both `vocanova-net` and
     `vocanova-production-net`
   - current operational config has no production `:8443` URLs
5. Run applicable deterministic checks listed in `implementation-plan.md`.

### Explicitly out of scope for this task

- Manual SSH cleanup.
- Staging cookie / Sentry / monitor hostname work.
- Claiming live host convergence without T03 evidence.

## VOC-079-T03 — Live deploy verification, monitoring window, rollback evidence

- Requirement source: issue #624 acceptance criteria; risk sequencing steps
  5–7
- Acceptance criteria: `VOC-079-AC-02`, `VOC-079-AC-03`, `VOC-079-AC-07`
- Tests: `VOC-079-TEST-05`, `VOC-079-TEST-06`
- Evidence: `VOC-079-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-079-T02` merged and released/deployed
  through the normal production workflow

### Required work

1. After the repository cleanup is on the production deploy path, record
   the production deploy run that converges the host (no manual Docker).
2. Verify: exactly one VocaNova nginx (`vocanova-shared-edge-nginx`); host
   ports `8081`/`8443` absent; all four canonical HTTPS checks pass
   (reuse/extend `infra/scripts/verify-voc067-cutover.sh` without requiring
   `:8443` bridge probes for the happy path).
3. Document rollback: redeploy prior revision recreates the bridge;
   Cloudflare `--restore` remains available; name the accountable rollback
   owner; record monitoring window expectations.
4. Confirm write isolation expectations still hold post-deploy (no
   cross-tier secret copy).

### Explicitly out of scope for this task

- New monitoring products.
- Changing the 2026-08-08 auto-release/auto-deploy delegation (unless
  adoption records a temporary cutover hold — see `release-plan.md`).

## Task ordering notes

- T00 unlocks the bridge-retention gate via `cloudflare_remap_api_status:
  absent`.
- T01 normalizes URLs while the bridge may still exist (safer dual-publish
  window).
- T02 removes the bridge and hardens invariants; must not merge before T00.
- T03 proves live convergence and rollback credibility.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #624 is gated on AC results with evidence, not on task
  issue closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
