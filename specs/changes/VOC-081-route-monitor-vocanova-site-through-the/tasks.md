# VOC-081 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → T03 → T04**.

## VOC-081-T00 — Repository-managed Kuma Compose, data ownership, and backup plan

- Requirement source: issue #665 required change 1; `VOC-081-D00`
- Acceptance criteria: `VOC-081-AC-00`
- Tests: `VOC-081-TEST-00`
- Evidence: `VOC-081-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Add a repository Compose file for Uptime Kuma (recommended:
   `infra/docker-compose.monitoring.yml`) that preserves:
   - Compose project name `monitoring`
   - container name `vocanova-uptime-kuma`
   - persistent data bind `/opt/vocanova/monitoring/kuma-data`
   - non-public host publish semantics (loopback-only or no host publish —
     must satisfy `VOC-081-D01` once T01 lands; T00 must not introduce a
     public `0.0.0.0:3001` bind)
2. Document backup/migration steps for first converge (copy/snapshot of
   `kuma-data`, verification that the bind path is reused, rollback if
   empty data appears).
3. Update `infra/README.md` with the monitoring project layout and
   identity-preservation notes.
4. Run `docker compose … config` validation for the new file.
5. Record evidence in `t00-evidence.md`. Do not claim live host converge
   until T03/T04.

### Explicitly out of scope for this task

- Shared-edge network membership (T01).
- Monitor vhost (T02).
- Deploy workflow wiring (T03).
- Live public HTTPS claims (T04).

## VOC-081-T01 — Dedicated monitoring network and topology invariants

- Requirement source: issue #665 required change 2; `VOC-081-D01`;
  `VOC-081-DEP-02` (network declaration portion)
- Acceptance criteria: `VOC-081-AC-01`, `VOC-081-AC-06` (repository-side
  topology tests)
- Tests: `VOC-081-TEST-01`, `VOC-081-TEST-02`
- Evidence: `VOC-081-EV-01` (`t01-evidence.md`)
- Status: pending

### Required work

1. Declare a dedicated external Docker network (recommended:
   `vocanova-monitoring-net`) in monitoring Compose and in
   `infra/docker-compose.shared-edge.yml`.
2. Attach Kuma and shared-edge to that network only for the monitoring
   path — do not attach monitoring to staging/production app networks in a
   way that weakens secret isolation.
3. Add deterministic foundation tests asserting:
   - shared-edge and Kuma both reference the monitoring network
   - only shared-edge publishes host `80`/`443`
   - Kuma has no public `0.0.0.0` / `[::]` port `3001` publish
   - staging/production compose projects are not given write ownership of
     the monitoring project
4. Keep routine app-deploy recreate semantics unchanged in this task
   (workflow apply lands in T03).

### Explicitly out of scope for this task

- Vhost content and Cloudflare Access (T02).
- Live `docker network create` on the server without repository deploy
  wiring (T03).
- Manual `docker network connect`.

## VOC-081-T02 — monitor.vocanova.site vhost + adopted access control

- Requirement source: issue #665 required change 3; `VOC-081-D02`;
  `VOC-081-DEP-00`; `VOC-081-DEP-01`
- Acceptance criteria: `VOC-081-AC-02`, `VOC-081-AC-03` (repository side)
- Tests: `VOC-081-TEST-03`, `VOC-081-TEST-04`
- Evidence: `VOC-081-EV-02` (`t02-evidence.md`)
- Status: pending

### Required work

1. Add a repository-owned nginx server block for `monitor.vocanova.site`
   loaded by shared edge per the adopted `VOC-081-DEP-01` path
   (recommended: under `infra/nginx-shared/conf.d/` so existing
   `include /etc/nginx/conf.d/shared/*.conf` loads it, or a dedicated
   mounted tree + nginx.conf include).
2. Configure production TLS certificate paths consistent with shared-edge
   mounts; inherit Cloudflare real-IP from shared config; set WebSocket
   upgrade headers and reverse-proxy headers; upstream to Kuma over the
   monitoring network DNS name.
3. Implement the adopted `VOC-081-DEP-00` access decision (Cloudflare
   Access / equivalent, or recorded public-exposure acceptance with Kuma
   auth retained). Repository docs/tests must encode the chosen policy's
   verifiable artifacts.
4. Ensure the stale unloaded host
   `production/nginx/conf.d/30-monitor.conf` cannot silently become a
   second source of truth — retire, replace, or document inertness via
   repository-owned converge (no SSH edit as acceptance).
5. Add/extend deterministic checks that the shared-edge config set actually
   loads the monitor server_name (parse includes / disposable `nginx -t`
   with documented local overrides).

### Explicitly out of scope for this task

- Claiming live Cloudflare Access provisioning complete without evidence
  (live proof is T04; T02 must still land the repository-side control or
  the explicit public-acceptance record).
- Deploy workflow convergence (T03).

## VOC-081-T03 — Deploy workflow convergence for monitoring + shared-edge

- Requirement source: issue #665 required changes 4–5; `VOC-081-DEP-02`
- Acceptance criteria: `VOC-081-AC-04`, `VOC-081-AC-06` (deploy-side)
- Tests: `VOC-081-TEST-05`, `VOC-081-TEST-06`
- Evidence: `VOC-081-EV-03` (`t03-evidence.md`)
- Status: pending

### Required work

1. Extend the normal repository deploy path so it:
   - ensures the monitoring external network exists
   - converges the monitoring Compose project from repository source
   - applies shared-edge network/config changes needed for monitor routing
   - runs fail-closed `nginx -t` before reload or controlled shared-edge
     recreate
2. Preserve VOC-079 invariants: exactly one VocaNova nginx; no routine
   app-deploy ownership of shared-edge; no cross-project orphan removal;
   staging/production secret write isolation intact.
3. Update workflow header comments and `infra/README.md` for the new
   ownership boundaries.
4. Add/extend deterministic tests that inspect the workflow/compose for
   the above safeguards.

### Explicitly out of scope for this task

- Manual SSH / `docker network connect` / ad hoc `docker rm`.
- Live public HTTPS closure (T04).
- Changing `.github/workflows/error-monitoring.yml`.

## VOC-081-T04 — Live deploy verification, WebSocket/HTTPS, rollback evidence

- Requirement source: issue #665 acceptance criteria; risk sequencing
- Acceptance criteria: `VOC-081-AC-03` (live), `VOC-081-AC-05`,
  `VOC-081-AC-06` (live), `VOC-081-AC-07`
- Tests: `VOC-081-TEST-07`, `VOC-081-TEST-08`, `VOC-081-TEST-09`
- Evidence: `VOC-081-EV-04` (`t04-evidence.md`)
- Status: pending

### Required work

1. After repository cleanup is on the production deploy path, record the
   deploy run that converges monitoring + shared-edge (no manual Docker).
2. Verify externally: `https://monitor.vocanova.site/` serves Kuma through
   Cloudflare; WebSocket works; access control matches DEP-00.
3. Verify on-host (via deploy logs or authorized ops evidence): single
   `vocanova-shared-edge-nginx`; no public Kuma port; monitoring network
   attachment; no `8081`/`8443`; four app hostnames still healthy.
4. Document rollback owner, last-known-good SHA, and monitoring window.
5. Confirm Sentry / hourly `error-monitoring` remain unchanged and healthy.

### Explicitly out of scope for this task

- Replacing Sentry or Kuma.
- Reintroducing production nginx bridge.

## Task ordering notes

- T00 establishes source of truth and backup before any converge.
- T01 provides the network path before the vhost can usefully upstream.
- T02 adds the vhost and access control once DEP-00/01 are settled.
- T03 makes the normal workflow the convergence path.
- T04 proves live outcome and rollback credibility.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #665 is gated on AC results with evidence, not on task
  issue closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
