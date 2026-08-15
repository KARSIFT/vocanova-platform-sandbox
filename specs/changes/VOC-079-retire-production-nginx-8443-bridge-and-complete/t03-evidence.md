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
attempt: 1
related_change: VOC-079
prior_tasks:
  - VOC-079-T00 (commit 5ecbdd5, PR #657; verify-only run #39)
  - VOC-079-T01 (commit 58f803b, PR #659)
  - VOC-079-T02 (commit da8dd5d, PR #660)
develop_tip_at_evidence: da8dd5d3b3cabc744f78d9c405c614a3666447b8
main_tip_at_evidence: b5998472b9fd315cb25b51eb2468a3613abd3575
rollback_owner: m-e-h-r-d-a-a-d
rollback_owner_role: founder gate operator (production workflow dispatch and cutover authority per VOC-079-T00)
last_known_good_sha: 58f803bd1fed05a5e0efae09aff017ee1195412a
gate_status: blocked-pending-release-and-deploy
---

# VOC-079-T03 — Live deploy verification, monitoring window, and rollback evidence

Evidence for `VOC-079-AC-02`, `VOC-079-AC-03`, and `VOC-079-AC-07`
(tests `VOC-079-TEST-05`, `VOC-079-TEST-06`).

Attempt 1 records the **pre-cutover baseline**, repository revision state,
rollback credibility, and monitoring expectations. **Live host convergence**
(bridge retirement, port absence, single-nginx container proof) remains
**blocked** until `develop` promotes `VOC-079-T02` to `main` and
`deploy-production.yml` runs on that revision — see §2.

## Summary

| Check | Status | Notes |
| --- | --- | --- |
| T02 merged on `develop` | **Pass** | `da8dd5d` (PR #660) |
| T02 on `main` / production deploy path | **Blocked** | `main` @ `b599847` — T01/T02 not promoted |
| Converging production deploy run | **Blocked** | Latest deploy #38 is pre-T02 (`b599847`) |
| External four-hostname `:443` (TEST-05) | **Pass** | `verify-voc067-cutover.sh` 2026-08-15 |
| `:8443` bridge still live (pre-cutover baseline) | **Observed** | `:8443` returns HTTP 200 — expected until T02 deploy |
| Rollback path documented (TEST-06) | **Pass** | §5; offline selftest PASS |
| Write isolation post-design | **Pass (repository)** | §6; live rehearsal deferred to post-deploy |
| Monitoring window named | **Pass** | §7 |

**Result (attempt 1):** `VOC-079-AC-07` rollback documentation and owner are
recorded. `VOC-079-AC-02` and `VOC-079-AC-03` live clauses remain **pending**
until a qualifying `deploy-production.yml` run on a bridge-free `main` revision
converges the host (attempt 2).

## 1. Preconditions and revision state

| Item | Value |
| --- | --- |
| `VOC-079-T00` gate | `cloudflare_remap_api_status: absent` (run #39) |
| `develop` tip | `da8dd5d` — includes T01 URL normalization + T02 bridge removal |
| `main` tip | `b599847` — T00 verifier correction only; **no T01/T02** |
| Latest production deploy | Run **#38** @ `b599847` (push to `main`, 2026-08-15T09:10:39Z) |
| Production host bridge | **Still present** — `:8443` probes return HTTP 200 (§3) |

T03's declared dependency (`VOC-079-T02` merged **and** released/deployed) is
only half satisfied: T02 merged to `develop` but not yet promoted to `main`.
Independent review should treat AC-02/AC-03 as **not closed** on this attempt.

## 2. Qualifying production deploy run (blocked)

A converging deploy must:

1. Run `deploy-production.yml` on a `main` revision that includes
   `da8dd5d` (or descendant) — production compose without `nginx` service.
2. Execute project-scoped
   `docker compose -f docker-compose.production.yml -p vocanova-production up -d --remove-orphans`
   (wired in `deploy-production.yml` since T02).
3. Retire `vocanova-production-nginx` declaratively — no manual `docker rm`.

**Latest deploy (not qualifying):**

| Field | Value |
| --- | --- |
| Run | <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31876327248> (#38) |
| Ref / SHA | `main` / `b5998472b9fd315cb25b51eb2468a3613abd3575` |
| Event | `push` (promotion merge #658) |
| Job | `deploy to production` → **success** |
| Bridge-free compose | **No** — host still serves `:8443` (§3) |

Attempt 2 must cite the first successful deploy run whose head SHA includes
T02 and whose deploy job logs show orphan removal converging without manual
Docker commands.

## 3. External verification (VOC-079-TEST-05 partial)

Captured **2026-08-15T09:35:56Z** from the implementer environment.

### 3a. Canonical `:443` checks (required happy path)

```bash
infra/scripts/verify-voc067-cutover.sh
```

```text
VOC-067 cutover verification — external :443 (via Cloudflare)
PASS: staging web (https://staging.vocanova.site/) -> HTTP 200
PASS: staging api healthz (https://api-staging.vocanova.site/healthz) -> HTTP 200
PASS: production web (https://production.vocanova.site/) -> HTTP 200
PASS: production api healthz (https://api-production.vocanova.site/healthz) -> HTTP 200

All required :443 checks passed.
```

This satisfies the four-hostname external `:443` portion of TEST-05. Shared
edge continues Host/SNI routing for all four production/staging hostnames on
ordinary HTTPS.

### 3b. Pre-cutover `:8443` baseline (bridge not yet retired)

```bash
curl -sS --max-time 15 -o /dev/null -w "production:8443 web -> HTTP %{http_code}\n" \
  "https://production.vocanova.site:8443/"
curl -sS --max-time 15 -o /dev/null -w "production:8443 api -> HTTP %{http_code}\n" \
  "https://api-production.vocanova.site:8443/healthz"
```

```text
production:8443 web -> HTTP 200
production:8443 api -> HTTP 200
```

**Interpretation:** `vocanova-production-nginx` (or equivalent bridge publish)
is still live on the production host. This is expected while `main` lacks T02.
Post-cutover acceptance requires these endpoints to **stop** returning 2xx and
host ports `8081`/`8443` to be unpublished — verified on attempt 2 via deploy
logs and/or authorized on-host inspection.

### 3c. Post-cutover command (attempt 2)

After the qualifying deploy, re-run:

```bash
infra/scripts/verify-voc067-cutover.sh
```

Do **not** pass `--include-8443-bridge` for the VOC-079 steady-state happy
path. Optionally confirm bridge absence with failed/non-2xx `:8443` curls or
deploy-log excerpts showing orphan removal of `vocanova-production-nginx`.

## 4. Container and port absence (blocked — attempt 2)

Required observable outcomes (AC-02):

1. Only VocaNova nginx container: `vocanova-shared-edge-nginx`.
2. Host ports `8081` and `8443` unpublished by the production stack.
3. Convergence via repository deploy + scoped orphan removal only.

SSH-gated `docker ps` / `ss -tlnp` transcripts are not available in this
environment. Attempt 2 must attach deploy-log excerpts or authorized ops
evidence from the converging run cited in §2.

## 5. Rollback path (VOC-079-TEST-06 / AC-07)

### 5a. Application rollback (primary)

Redeploy the **last-known-good** revision that still defined the production
nginx bridge:

| Field | Value |
| --- | --- |
| SHA | `58f803b` (`VOC-079-T01` merge — bridge service + canonical `:443` URLs) |
| PR | #659 |
| Compose | `vocanova-production-nginx` on `8081:80` / `8443:443` |

Normal `deploy-production.yml` on that revision recreates the bridge from
compose — no manual `docker stop/rm` required (`VOC-079-D01`).

### 5b. Cloudflare rollback (secondary)

If edge routing must again target origin `:8443` after remap removal:

```bash
# Repository script (production token required)
PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN=… \
  infra/scripts/cloudflare-remove-production-origin-port-remap.sh --restore

# Or workflow_dispatch on deploy-production.yml:
#   voc067_cloudflare_origin_cutover=restore
```

Offline selftest coverage (deterministic, no live API):

```bash
infra/scripts/cloudflare-remove-production-origin-port-remap.selftest.sh
```

Result on 2026-08-15: **All cloudflare cutover selftests passed** (includes
`restore re-adds port 8443` case).

After bridge restoration, optional bridge probes:

```bash
infra/scripts/verify-voc067-cutover.sh --include-8443-bridge
```

### 5c. Rollback triggers

Per `release-plan.md`:

- Either tier unreachable or 5xx on canonical `:443` after bridge retirement.
- OAuth/CORS failure after URL normalization.
- Shared-edge nginx crash loop or failed `nginx -t` reload.
- Accidental orphan removal outside the production compose project.

### 5d. No manual server surgery

Happy path, rollback path, and bridge retirement all use repository
`deploy-production.yml` + project-scoped compose only. Manual SSH edits or
ad hoc `docker rm` are not acceptance paths.

## 6. Write isolation (repository confirmation)

Post-T02 design preserves VOC-067-T03 / VOC-037-D01 isolation:

| Control | Evidence |
| --- | --- |
| Production deploy writes only `/opt/vocanova/production/nginx/` | `deploy-production.yml` deploy steps |
| Staging secrets not copied to production tree | `rehearse-production-secrets-boundary.sh` in deploy path |
| Shared-edge reload fail-closed | `docker exec vocanova-shared-edge-nginx nginx -t` before reload |
| No production-nginx reload steps | Removed in T02 (`voc079-single-edge-invariants.test.mjs`) |
| Routine deploy does not recreate shared-edge | Header + workflow comments; staging owns rare bring-up |
| Orphan removal scoped to production project | `compose … -p vocanova-production up -d --remove-orphans` only |

Live post-deploy rehearsal output belongs in attempt 2's converging deploy
transcript.

## 7. Monitoring window and accountable owner

| Item | Value |
| --- | --- |
| **Rollback owner** | `m-e-h-r-d-a-a-d` (founder gate operator) |
| **Monitoring start** | First successful bridge-free production deploy (attempt 2) |
| **Duration** | 24 hours minimum observation per R4 edge-change practice |
| **Signals** | External HTTPS on four canonical URLs; shared-edge container health; OAuth start/callback on production; existing Sentry/uptime tier signals |
| **Escalation** | Rollback owner dispatches prior-revision redeploy (§5a) and, if needed, Cloudflare `--restore` (§5b) |

## 8. Attempt 2 checklist (independent reviewer / operator)

1. Confirm `develop` → `main` promotion includes `da8dd5d` (T02).
2. Record the first successful `deploy-production.yml` run on that `main` SHA.
3. Re-run `infra/scripts/verify-voc067-cutover.sh` — all four `:443` checks pass.
4. Confirm `:8443` no longer returns 2xx; `vocanova-production-nginx` absent;
   only `vocanova-shared-edge-nginx` remains.
5. Update this file `gate_status` to `resolved` and set AC-02/AC-03 results in
   `acceptance-criteria.md`.

## Limitations

- No production SSH or `docker` access in the implementer environment.
- GitHub Actions job logs require authenticated access; this evidence uses
  public run metadata, live `curl` transcripts, and offline selftests.
- `:443` external success alone does not prove bridge retirement while the
  bridge container may still be running (pre-cutover state documented in §3b).
