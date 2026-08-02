# VOC-037-EV-04 — Production monitoring/alerting evidence (T04)

## Standing at this revision

**`VOC-037-AC-04` is SATISFIED (2026-08-02).** Both halves of the
observable outcome — a deliberately triggered Sentry test error observed
for the production environment, and a deliberate uptime-check failure
producing a founder-reaching alert — were driven live against real
production infrastructure and real founder-confirmed receipt, not
simulated. See "Live verification (2026-08-02)" below for the actual
event IDs, commands, and outputs.

**Disclosed deviation from `VOC-037-TEST-04`'s literal wording:** the test
plan names "Better Stack/UptimeRobot" specifically. UptimeRobot's free
tier defaults HTTP monitors to `HEAD` requests and does not expose the
HTTP-method setting to non-paying accounts; the production API's
`/healthz` route only implements `GET` (confirmed live: `HEAD` returns
`405 Method Not Allowed`), so a UptimeRobot free-tier monitor against the
real endpoint cannot be made to report accurately without either a paid
plan or weakening the endpoint. The founder, once informed of this
constraint live, directed the switch to a self-hosted alternative
(Uptime Kuma) instead of paying or working around it. This satisfies the
AC's actual observable outcome (a real uptime-check failure produces a
real founder-reaching alert) via a different, founder-approved tool than
the one `VOC-037-TEST-04` names by example. Recorded here as a disclosed
substitution, not silently treated as the literally-named tool.

**Correction to this section's earlier text (superseded by the above):**
it previously claimed the in-repo/external split below "matches
`VOC-037-TEST-04`," then was corrected to record `AC-04`/`TEST-04` as not
satisfied at all pending external verification. That external
verification has now actually happened (below), so this section is
updated again to SATISFIED.

- In-repo deliverables for production monitoring are implemented and tested:
  - API-side Sentry wiring (env-driven, no-op when unset).
  - `sentryhttp` middleware wrapping the real HTTP handler, so real
    request errors/panics reach Sentry - not just the deliberate test
    endpoint below (fixes a reviewed Medium finding: the original
    revision only ever called `sentry.CaptureMessage` from the manual
    test route, so "error monitoring active" wasn't actually true for
    real traffic).
  - Authenticated, production-only Sentry test-event endpoint, gated on
    both a non-empty token AND `environment == "production"` (fixes a
    reviewed Medium finding: the original gate was token-only, so a
    non-production tier with the token set would also expose the route).
    Returns `502` rather than a false `200` if Sentry doesn't return an
    event ID (fixes a reviewed Low finding).
  - Production deploy workflow secret sync for monitoring credentials,
    which now skips cleanly (rather than hard-failing every future
    deploy, including unrelated ones like T03) while
    `PRODUCTION_SENTRY_DSN`/`PRODUCTION_MONITORING_TEST_TOKEN` don't exist
    yet, but still fails closed on a half-configured pair (fixes a
    reviewed Medium finding).
- External, founder-facing proof still requires one live production rehearsal:
  - deliberate Sentry test event observed in the production Sentry project;
  - deliberate uptime-check failure observed in Better Stack/UptimeRobot.
- **No uptime-monitor configuration exists yet** (flagged as a reviewed
  High finding) because it requires an external Better Stack/UptimeRobot
  account this task has no access to create. Recorded as an outstanding
  founder step below, not silently treated as done.
- **Known, accepted limitation:** `deploy-production.yml`'s monitoring
  secret sync skips cleanly while `PRODUCTION_SENTRY_DSN`/
  `PRODUCTION_MONITORING_TEST_TOKEN` are unset, so deploys stay green while
  monitoring remains unconfigured (flagged as a reviewed Medium finding).
  This is intentional - hard-failing every deploy (including unrelated
  ones) until external accounts exist would just trade one blocking
  problem for another - but it means a green deploy is explicitly **not**
  evidence that AC-04 is satisfied. `VOC-037-T05` (the R2 go/no-go gate)
  must check AC-04's real status directly, not infer it from deploy
  success.

## Repository deliverables implemented in T04

| Deliverable | Location | Verification |
| --- | --- | --- |
| Sentry runtime config fields | `apps/api/app/api/production.go` (`SENTRY_DSN`, `SENTRY_ENVIRONMENT`, `SENTRY_RELEASE`) | `go test ./...` in `apps/api` |
| Deliberate test-event endpoint | `POST /ops/monitoring/sentry-test` in `apps/api/app/api/production.go` | `TestRegisterMonitoringSentryTest_AuthBehavior` |
| Monitoring test token gating | `MONITORING_TEST_TOKEN` in `apps/api/app/api/production.go` | `TestRegisterMonitoringSentryTest_AuthBehavior` |
| Production deploy secret sync | `.github/workflows/deploy-production.yml` (`PRODUCTION_SENTRY_DSN`, `PRODUCTION_MONITORING_TEST_TOKEN`) | workflow file inspection |
| API env schema docs | `apps/api/.env.example` monitoring section | file inspection |

## Deterministic checks run

```bash
cd apps/api
go test ./...
```

Observed result: PASS for all packages, including `app/api` and `cmd/api`.

## Live verification (2026-08-02)

**Sentry (error monitoring):**

- `PRODUCTION_SENTRY_DSN`/`PRODUCTION_MONITORING_TEST_TOKEN` populated in the
  `production` GitHub environment; `deploy-production` run wrote them into
  `api.env` as `SENTRY_DSN`/`SENTRY_ENVIRONMENT=production`/
  `SENTRY_RELEASE=sha-<deployed-sha>`/`MONITORING_TEST_TOKEN`.
- Deliberate test event triggered against the real production API:
  `POST https://api-production.vocanova.site:8443/ops/monitoring/sentry-test`
  with the real bearer token.
- Real event confirmed delivered and visible in the production Sentry
  project: **event ID `60c282e455a843ff9151a235ebb71dda`**.
- `sentryhttp` middleware (wired in `apps/api/cmd/api/main.go`) confirmed
  live, so real request errors/panics reach Sentry, not just this
  deliberate test route.

**Uptime monitoring (self-hosted Uptime Kuma, substituted for
Better Stack/UptimeRobot — see disclosed deviation above):**

- Deployed at `/opt/vocanova/monitoring/` on the shared production host,
  as its own `docker compose` project (`vocanova-uptime-kuma`), bound to
  `127.0.0.1:3001` only (no new DNS/Cloudflare/firewall changes — the
  dashboard is reached via SSH tunnel; background checks and alerting run
  independent of whether the dashboard is exposed).
- Two monitors created against real production URLs, `GET` method
  (avoiding the `HEAD`-only free-tier limitation that ruled out
  UptimeRobot): `https://api-production.vocanova.site:8443/healthz` and
  `https://production.vocanova.site:8443/`, both confirmed initially `UP`
  with real `200` responses.
- Alert channel: Telegram bot (`@vocanova_alerts_bot`), configured with
  the founder's real chat ID, test notification sent and founder-confirmed
  received.
- **Live down/up rehearsal, not simulated:** the API monitor's URL was
  temporarily repointed to a real nonexistent path
  (`/this-path-does-not-exist-live-test`) against the real host. Real
  heartbeat observed: `status=0 msg="Request failed with status code 404"`.
  Telegram alert for the down event founder-confirmed received. Monitor
  then restored to the real `/healthz` URL and confirmed `UP` again -
  the founder was not asked to simply trust the config; the alert path
  was proven to fire end-to-end.
- Dashboard admin credentials created (`mehrdad` / real password handed to
  founder directly, not recorded in this document or any repository
  file).

## Notes

- This task adds monitoring instrumentation and safe test hooks only; it does not change launch authority or go/no-go gates (`VOC-037-T05` remains the founder gate).
- The monitoring test endpoint is not registered unless `MONITORING_TEST_TOKEN` is set.
- Uptime Kuma's admin credentials, the Telegram bot token, and the chat ID
  are operational secrets for a founder-facing tool, not application
  secrets; they are not stored in this repository. If the bot token needs
  rotation, regenerate it via `@BotFather` and update Kuma's notification
  config directly (SSH tunnel + dashboard, or the same socket.io scripting
  approach used to set it up).
