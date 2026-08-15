# VOC-081 — Acceptance Criteria

## VOC-081-AC-00 — Repository owns Kuma Compose and preserves existing data identity

- Requirement source: issue #665 required change 1; `VOC-081-D00`
- Tasks: `VOC-081-T00`
- Tests: `VOC-081-TEST-00`
- Evidence: `VOC-081-EV-00`
- Result: pending

Observable outcome:

1. A repository Compose definition exists for the monitoring stack (expected
   path: `infra/docker-compose.monitoring.yml` or equivalent documented in
   `infra/README.md`).
2. It preserves Compose project identity `monitoring`, container name
   `vocanova-uptime-kuma`, and bind/ownership of
   `/opt/vocanova/monitoring/kuma-data`.
3. A written backup/migration plan is recorded before first live converge
   (evidence may live in `t00-evidence.md`).
4. First converge is non-destructive: existing Kuma data is reused; empty
   re-initialization is a FAIL for this criterion.

## VOC-081-AC-01 — Least-privilege Docker path from shared edge to Kuma

- Requirement source: issue #665 required change 2; `VOC-081-D01`
- Tasks: `VOC-081-T01`
- Tests: `VOC-081-TEST-01`, `VOC-081-TEST-02`
- Evidence: `VOC-081-EV-01`
- Result: pending

Observable outcome:

1. A dedicated repository-managed external Docker network (recommended:
   `vocanova-monitoring-net`) connects shared edge and Kuma.
2. Shared edge resolves and reaches Kuma by Compose/service DNS name over
   that network.
3. Kuma is not published on a public host interface; port `3001` is not
   bound to `0.0.0.0` or `[::]`.
4. Staging and production secret / network write isolation is not weakened
   (no cross-tier secret mounts via the monitoring network).

## VOC-081-AC-02 — Shared edge loads a repository-owned monitor.vocanova.site vhost

- Requirement source: issue #665 required change 3; `VOC-081-DEP-01`
- Tasks: `VOC-081-T02`
- Tests: `VOC-081-TEST-03`
- Evidence: `VOC-081-EV-02`
- Result: pending

Observable outcome:

1. A repository-owned nginx server block for `monitor.vocanova.site` is
   loaded by `vocanova-shared-edge-nginx` (not an unloaded host-only
   `30-monitor.conf`).
2. The vhost uses production TLS material paths consistent with shared-edge
   mounts, inherits Cloudflare client-IP handling from shared config, and
   sets WebSocket upgrade headers plus required reverse-proxy headers
   (`Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`, and
   Connection upgrade semantics needed by Kuma).
3. Upstream targets the Kuma service over the monitoring network DNS name.

## VOC-081-AC-03 — Access exposure is controlled (proxied DNS is not authorization)

- Requirement source: issue #665 risk section; `VOC-081-D02`; `VOC-081-DEP-00`
- Tasks: `VOC-081-T02`, `VOC-081-T04`
- Tests: `VOC-081-TEST-04`
- Evidence: `VOC-081-EV-02`, `VOC-081-EV-04`
- Result: pending

Observable outcome:

1. Adoption records the chosen access policy (Cloudflare Access / equivalent
   repository-managed control, **or** explicit written acceptance of public
   exposure).
2. If private: unauthenticated requests from the public internet are denied
   or challenged by the adopted control before Kuma UI content is served;
   redacted verification evidence is attached.
3. If public exposure is accepted: evidence cites the adoption decision and
   confirms Kuma's own authentication remains enabled.
4. Treating “Cloudflare orange-cloud / proxied A record exists” alone as the
   access control is a FAIL.

## VOC-081-AC-04 — Normal deploy converges monitoring + shared-edge without manual mutation

- Requirement source: issue #665 required changes 4–5; `VOC-081-DEP-02`
- Tasks: `VOC-081-T03`
- Tests: `VOC-081-TEST-05`, `VOC-081-TEST-06`
- Evidence: `VOC-081-EV-03`
- Result: pending

Observable outcome:

1. The normal repository deploy workflow creates/converges the monitoring
   network and Kuma service from repository Compose.
2. Required shared-edge network/config changes are applied with fail-closed
   `nginx -t` before reload or controlled recreate.
3. Acceptance does **not** depend on manual SSH file edit,
   `docker network connect`, or ad hoc container removal.
4. Routine staging/production app deploys do not take ownership of the
   monitoring project, do not orphan-remove shared-edge, and do not weaken
   per-tier write isolation.
5. Exactly one VocaNova nginx remains: `vocanova-shared-edge-nginx`.

## VOC-081-AC-05 — Public HTTPS and WebSocket work through Cloudflare

- Requirement source: issue #665 acceptance criteria
- Tasks: `VOC-081-T04`
- Tests: `VOC-081-TEST-07`
- Evidence: `VOC-081-EV-04`
- Result: pending

Observable outcome after normal deploy:

1. `https://monitor.vocanova.site/` returns the Uptime Kuma UI through
   Cloudflare on canonical HTTPS (no origin-port qualification).
2. WebSocket functionality required by the Kuma UI works (evidence may be a
   scripted upgrade check, authenticated browser note, or equivalent
   deterministic probe recorded in `t04-evidence.md`).
3. Evidence includes the qualifying deploy run URL and redacted external
   verification.

## VOC-081-AC-06 — Topology, exclusivity, and rollback evidence pass

- Requirement source: issue #665 required changes 6–7
- Tasks: `VOC-081-T01`, `VOC-081-T03`, `VOC-081-T04`
- Tests: `VOC-081-TEST-01`, `VOC-081-TEST-02`, `VOC-081-TEST-05`,
  `VOC-081-TEST-06`, `VOC-081-TEST-08`
- Evidence: `VOC-081-EV-01`, `VOC-081-EV-03`, `VOC-081-EV-04`
- Result: pending

Observable outcome:

1. Deterministic tests cover: vhost loading, network reachability topology,
   exclusive host `80`/`443` ownership by shared-edge, no public Kuma port,
   scoped Compose ownership, and rollback credibility.
2. Live evidence confirms single-edge / listener state after deploy.
3. Rollback is documented with named owner and last-known-good reference;
   primary rollback does not require undocumented manual SSH.

## VOC-081-AC-07 — Sentry and hourly error-monitoring remain unchanged and healthy

- Requirement source: issue #665 acceptance criteria; VOC-051
- Tasks: `VOC-081-T04`
- Tests: `VOC-081-TEST-09`
- Evidence: `VOC-081-EV-04`
- Result: pending

Observable outcome:

1. `.github/workflows/error-monitoring.yml` is not modified by this package
   (or any diff is explicitly justified as non-behavioral and approved in
   evidence — default expectation: **zero change**).
2. Post-deploy evidence notes that scheduled/error-monitoring health and
   Sentry error monitoring remain successful (or records a limitation if
   credentials/schedule window prevent a live check — never invent a pass).
