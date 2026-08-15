---
evidence_id: VOC-081-EV-04
task_id: VOC-081-T04
acceptance_criteria:
  - VOC-081-AC-03
  - VOC-081-AC-05
  - VOC-081-AC-06
  - VOC-081-AC-07
tests:
  - VOC-081-TEST-07
  - VOC-081-TEST-08
  - VOC-081-TEST-09
date: 2026-08-15
attempt: 2
related_change: VOC-081
develop_sha: 6d7a3ec663cc6e0cab013ce6fa90018cbeb7d42c
production_sha: 25fca3fbaa79
qualifying_deploy_run: https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31889058108
qualifying_deploy_status: success
rollback_owner: autonomous staging/production deploy workflows (active A-004)
last_known_good_sha: 25fca3fbaa79
gate_status: live-closure-complete
---

# VOC-081-T04 — Live deploy verification, WebSocket/HTTPS, rollback evidence

Attempt 2 remediates the High findings from attempt 1 (commit
`8431d4de`): a successful repository Deploy step converged monitoring +
shared-edge, and external probes now return the Uptime Kuma UI through
Cloudflare.

## Summary

| Requirement                                                     | Result                                                                                                                                                                                                                                                                                             |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Qualifying deploy run recorded                                  | **PASS** — [run 31889058108](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31889058108) (`VOC-082: verify capped mission completion in staging` #684 on `6d7a3ec`); the full workflow, including monitoring convergence, app health probes, and the real core-loop, succeeded. |
| `https://monitor.vocanova.site/` serves Kuma through Cloudflare | **PASS** — HTTP 200 after `/` → `/dashboard` redirect; body includes Uptime Kuma                                                                                                                                                                                                                   |
| WebSocket path for Kuma UI                                      | **PASS** — Engine.IO polling handshake HTTP 200 with `sid` and `upgrades:["websocket"]`                                                                                                                                                                                                            |
| Access control matches VOC-081-DEP-00                           | **PASS** — unauthenticated SPA + `/api/entry-page` without session token; proxied DNS alone not cited as auth                                                                                                                                                                                      |
| Four app hostnames healthy on canonical :443                    | **PASS** — all HTTP 200                                                                                                                                                                                                                                                                            |
| Single-edge / no public Kuma port (live)                        | **PASS** — deploy logs show single `vocanova-shared-edge-nginx` recreate/start; compose keeps `127.0.0.1:3001`; external `nc` to host `:3001` times out                                                                                                                                            |
| Sentry / `error-monitoring` unchanged                           | **PASS** — zero VOC-081 diff to `error-monitoring.yml`; recent scheduled runs successful                                                                                                                                                                                                           |
| Rollback documented                                             | **PASS** — see §Rollback                                                                                                                                                                                                                                                                           |

## Root cause closed between attempt 1 and attempt 2

Attempt 1 recorded Cloudflare/origin **502** because staging deploy
[31888512579](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31888512579)
started Kuma but never marked it healthy: compose used bare
`CMD extra/healthcheck.js` (no shebang → Docker exec failure). The Deploy
step aborted before shared-edge `compose up` joined `vocanova-monitoring-net`.

#686 fixed the healthcheck to `CMD node extra/healthcheck.js`. Run
31888841507 then:

1. Created/validated first-converge backup
   `/opt/vocanova/monitoring/backups/kuma-data-pre-voc081-20260815T140504Z.tar.gz`
2. Recreated/started `vocanova-uptime-kuma` and recorded
   `.repository-converged-voc081`
3. Recreated/started `vocanova-shared-edge-nginx` and reloaded
   (“Shared edge converged and reloaded… VOC-081-T03 routine path”)

The deploy remains fail-closed on Docker health, so a future healthcheck
regression cannot be masked by a shallower HTTP-only readiness signal.

## Repository deliverables (this task)

| Artifact                    | Path                                                                            |
| --------------------------- | ------------------------------------------------------------------------------- |
| Live verification script    | `infra/scripts/verify-voc081-monitor.sh`                                        |
| Disposable selftest harness | `infra/scripts/verify-voc081-monitor.selftest.sh`                               |
| Foundation tests            | `scripts/foundation/voc081-live-verification.test.mjs` (+ topology assertions)  |
| Operator docs               | `infra/README.md`, `infra/monitoring/access-policy.md`                          |
| This evidence               | `specs/changes/VOC-081-route-monitor-vocanova-site-through-the/t04-evidence.md` |

## External verification (2026-08-15, implementer CI runner)

```bash
infra/scripts/verify-voc081-monitor.sh
```

```
VOC-081 monitor verification — external checks (via Cloudflare unless MONITOR_BASE_URL is set)

App tier :443 regression guard (verify-voc067-cutover.sh)
PASS: staging web (https://staging.vocanova.site/) -> HTTP 200
PASS: staging api healthz (https://api-staging.vocanova.site/healthz) -> HTTP 200
PASS: production web (https://production.vocanova.site/) -> HTTP 200
PASS: production api healthz (https://api-production.vocanova.site/healthz) -> HTTP 200

Monitor hostname checks (https://monitor.vocanova.site)
PASS: monitor web (https://monitor.vocanova.site/ → final) -> HTTP 200
PASS: monitor body includes Uptime Kuma marker
PASS: unauthenticated SPA/entry-page boundary present (VOC-081-DEP-00; Kuma auth retained)
PASS: monitor socket.io polling handshake (.../socket.io/?EIO=4&transport=polling) -> HTTP 200 (sid present; websocket upgrade advertised)

All VOC-081 monitor verification checks passed.
```

Disposable harness:

```bash
infra/scripts/verify-voc081-monitor.selftest.sh
# SELFTEST PASS: healthy monitor login + websocket
# SELFTEST PASS: origin 502 is rejected
# SELFTEST PASS: anonymous dashboard is rejected
```

Direct origin (Cloudflare bypass via `--resolve` to `130.185.123.152`) also
returns HTTP 302 to `/dashboard` (nginx serving Kuma, not CF error pages).

## VOC-081-AC-03 — Access exposure (live)

Adopted policy (`VOC-081-DEP-00`): public Kuma login with Kuma authentication
as the authorization boundary (`infra/monitoring/access-policy.md`). Proxied
DNS alone is not authorization.

Live checks:

- Unauthenticated HTML is the Kuma SPA shell (title/description “Uptime Kuma”),
  not an authenticated admin document.
- `GET /api/entry-page` returns `{"type":"entryPage","entryPage":null}` without
  session token / username fields.
- FAIL condition “DNS is proxied” was not used as the access control.

## VOC-081-AC-06 — Topology, exclusivity, and rollback

### Live / deploy-log invariants

From successful run 31889058108, external probes, and authorized read-only
host inspection:

| Check                                                          | Evidence                                                                                                                            |
| -------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Monitoring converge                                            | `Recorded successful first repository-managed Kuma convergence`                                                                     |
| Shared-edge converge                                           | `Container vocanova-shared-edge-nginx Recreated/Started` + reload line                                                              |
| Single VocaNova nginx                                          | Exactly one nginx container: `vocanova-shared-edge-nginx`, healthy, publishing host `80`/`443`                                      |
| Dedicated network                                              | `vocanova-monitoring-net` contains exactly `vocanova-shared-edge-nginx` and `vocanova-uptime-kuma`; app containers are not attached |
| No public Kuma `:3001`                                         | Host listener is `127.0.0.1:3001` only; no `8081`/`8443` listeners exist                                                            |
| App tiers healthy                                              | Four canonical hostnames HTTP 200 via `verify-voc067-cutover.sh`                                                                    |
| Data preserved                                                 | Validated backup                                                                                                                    |
| `kuma-data-pre-voc081-20260815T140504Z.tar.gz` before recreate |

Kuma and shared-edge both reported healthy. The first-converge marker exists;
the two guarded backup attempts are non-empty and mode `0600`. Kuma's public
Socket.IO path negotiated WebSocket successfully, and an unauthenticated
`getMonitorList` callback returned `ok:false` with `You are not logged in.`

### Rollback

| Field                        | Value                                                                                                                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Accountable owner**        | Autonomous staging/production deploy workflows under active A-004 (no founder-comment gate)                                                                               |
| **Last-known-good SHA**      | `25fca3fbaa79` (post-VOC-079 `main` tip; monitor returned Cloudflare 520, Kuma on loopback only — VOC-079-EV-03 baseline)                                                 |
| **Primary rollback**         | Promote/redeploy prior revision that removes monitor vhost / monitoring-network membership from shared-edge compose while preserving `/opt/vocanova/monitoring/kuma-data` |
| **Manual SSH**               | Not the primary acceptance path (`VOC-081-D00`)                                                                                                                           |
| **Re-verify after rollback** | Four app hostnames HTTP 200; monitor hostname 520 or login per rolled-back policy; single nginx; no `8081`/`8443`                                                         |

## VOC-081-AC-07 — Sentry / error-monitoring

| Check                                                        | Result                                                                                                                                             |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.github/workflows/error-monitoring.yml` modified by VOC-081 | **PASS (zero behavioral change; not in this task diff)**                                                                                           |
| Recent scheduled `error-monitoring` runs                     | **PASS** — e.g. [run 31887093187](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31887093187) (`success`, 2026-08-15T13:23:42Z) |

## Monitoring window

Observation opened 2026-08-15 with attempt 2 evidence and closed after fully
successful run 31889058108:

- **Green:** four app-tier `:443` checks; monitor HTTPS + Engine.IO; hourly
  `error-monitoring` success; first-converge marker recorded.
- **Green:** the staging core-loop also passed after its target-cap assertion
  was corrected, so the qualifying workflow has no failed step.

**Closure criteria for this task:** met — successful Deploy convergence +
`infra/scripts/verify-voc081-monitor.sh` exit 0 + rollback fields + AC-07.

## Validation commands (repository)

```bash
node --test scripts/foundation/voc081-live-verification.test.mjs
node --test scripts/foundation/voc081-*.test.mjs
bash infra/scripts/verify-voc081-monitor.selftest.sh
bash infra/scripts/verify-voc081-monitor.sh
bash scripts/governance/validate-governance.sh
git diff --check
```

## Acceptance-criteria status (this attempt)

| AC                   | Status                                                          |
| -------------------- | --------------------------------------------------------------- |
| VOC-081-AC-03 (live) | **PASS**                                                        |
| VOC-081-AC-05        | **PASS** — monitor hostname restored; WebSocket/Engine.IO works |
| VOC-081-AC-06 (live) | **PASS** — deploy-log topology + external exclusivity probes    |
| VOC-081-AC-07        | **PASS**                                                        |

Issue #665 package closure still requires the remaining ACs from T00–T03
evidence plus independent verification of this exact revision; this task’s
live clauses are recorded as satisfied.
