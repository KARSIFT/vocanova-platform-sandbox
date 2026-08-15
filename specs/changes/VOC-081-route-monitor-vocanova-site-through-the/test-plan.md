# VOC-081 — Test Plan

## VOC-081-TEST-00 — Monitoring Compose preserves identity and data path

- Covers: `VOC-081-AC-00`
- Preconditions: `VOC-081-T00` tree
- Procedure:
  1. Parse the repository monitoring Compose file.
  2. Assert project/container naming and
     `/opt/vocanova/monitoring/kuma-data` (or documented equivalent bind)
     are present.
  3. Assert no public `0.0.0.0:3001` / `[::]:3001` publish is introduced.
  4. `docker compose -f <monitoring-compose> config` succeeds.
  5. Confirm backup/migration notes exist in package evidence or
     `infra/README.md`.
- Expected result: identity and data-path assertions pass; compose config
  succeeds.
- Evidence: `VOC-081-EV-00`

## VOC-081-TEST-01 — Monitoring network topology declared for edge and Kuma

- Covers: `VOC-081-AC-01`, `VOC-081-AC-06`
- Preconditions: `VOC-081-T01` tree
- Procedure:
  1. Assert monitoring Compose and shared-edge Compose both reference the
     dedicated external monitoring network by name.
  2. Assert Kuma is not attached to staging/production secret-bearing
     networks in a way that mounts tier secrets.
  3. `docker compose … config` for both files succeeds with documented
     local overrides.
- Expected result: least-privilege network topology is explicit in
  repository Compose.
- Evidence: `VOC-081-EV-01`

## VOC-081-TEST-02 — Exclusive 80/443 ownership; no public Kuma port

- Covers: `VOC-081-AC-01`, `VOC-081-AC-06`
- Preconditions: `VOC-081-T01` tree (retain through later tasks)
- Procedure:
  1. Assert only shared-edge Compose publishes host `80`/`443`.
  2. Assert monitoring Compose does not publish public `3001`.
  3. Assert production Compose still has no nginx / no `8081`/`8443`
     (VOC-079 invariant retained).
- Expected result: port exclusivity and non-public Kuma invariants hold.
- Evidence: `VOC-081-EV-01`, `VOC-081-EV-03`

## VOC-081-TEST-03 — Shared edge loads monitor.vocanova.site vhost

- Covers: `VOC-081-AC-02`
- Preconditions: `VOC-081-T02` tree
- Procedure:
  1. Assert repository vhost content includes `server_name` for
     `monitor.vocanova.site`, TLS paths, WebSocket upgrade headers, and
     required proxy headers.
  2. Assert shared-edge nginx main/includes actually load that fragment
     (include glob or explicit mount — not an unloaded `30-*.conf` under
     production-only globs).
  3. Run disposable `nginx -t` with the documented shared-edge mount set
     (follow existing infra/README or staging deploy pattern).
- Expected result: config test passes; server_name is part of the loaded
  set.
- Evidence: `VOC-081-EV-02`

## VOC-081-TEST-04 — Access exposure control is explicit

- Covers: `VOC-081-AC-03`
- Preconditions: adopted `VOC-081-DEP-00` decision; `VOC-081-T02` tree
- Procedure:
  1. If private Access policy: assert repository docs/workflow/evidence
     reference the Cloudflare Access (or equivalent) application for
     `monitor.vocanova.site`, and define the live verification probe used
     in T04.
  2. If public exposure accepted: assert adoption evidence records that
     decision and that Kuma authentication remains required.
  3. FAIL if the only cited control is “DNS is proxied.”
- Expected result: access policy is explicit and testable.
- Evidence: `VOC-081-EV-02`, `VOC-081-EV-04`

## VOC-081-TEST-05 — Deploy path converges monitoring without manual Docker

- Covers: `VOC-081-AC-04`
- Preconditions: `VOC-081-T03` workflow/compose tree
- Procedure:
  1. Inspect deploy workflow(s) for steps that create/converge the
     monitoring network and monitoring Compose from repository files.
  2. Assert acceptance path does not instruct operators to run
     `docker network connect`, manual SSH conf edits, or ad hoc
     `docker rm`.
  3. Assert fail-closed `nginx -t` remains before shared-edge
     reload/recreate.
- Expected result: convergence is declarative via repository workflow.
- Evidence: `VOC-081-EV-03`

## VOC-081-TEST-06 — Scoped ownership and isolation preserved

- Covers: `VOC-081-AC-04`, `VOC-081-AC-06`
- Preconditions: `VOC-081-T03` tree
- Procedure:
  1. Assert routine staging/production app deploys do not
     `--force-recreate` shared-edge as ordinary behavior.
  2. Assert orphan-removal / compose project flags remain scoped so
     monitoring and shared-edge are not accidentally removed by app
     deploys.
  3. Assert no new cross-tier secret copy steps are introduced.
- Expected result: VOC-067/VOC-079 isolation semantics preserved.
- Evidence: `VOC-081-EV-03`

## VOC-081-TEST-07 — Live HTTPS and WebSocket through Cloudflare

- Covers: `VOC-081-AC-05`, `VOC-081-AC-03` (live)
- Preconditions: T03 released/deployed via normal production workflow
- Procedure:
  1. Record qualifying deploy run URL.
  2. External check `https://monitor.vocanova.site/` (expect Kuma UI or
     Access challenge per DEP-00 — not Cloudflare 520).
  3. Verify WebSocket upgrade path used by Kuma (scripted probe or
     recorded authenticated check).
  4. Verify access control behavior matches DEP-00.
- Expected result: hostname restored; WebSocket works; access policy
  holds. Missing credentials/network is a limitation, not a pass.
- Evidence: `VOC-081-EV-04`

## VOC-081-TEST-08 — Live single-edge / listener / rollback credibility

- Covers: `VOC-081-AC-06`
- Preconditions: T04 evidence in progress
- Procedure:
  1. Confirm only `vocanova-shared-edge-nginx` among VocaNova nginx
     containers; host `8081`/`8443` absent; Kuma not on public `3001`.
  2. Confirm four app hostnames still return healthy canonical HTTPS.
  3. Document rollback to last-known-good revision without manual SSH as
     the primary path; name rollback owner and monitoring window.
- Expected result: live invariants hold; rollback is specific and
  reversible.
- Evidence: `VOC-081-EV-04`

## VOC-081-TEST-09 — Sentry / error-monitoring unchanged and healthy

- Covers: `VOC-081-AC-07`
- Preconditions: T04 window
- Procedure:
  1. Diff/check that `.github/workflows/error-monitoring.yml` was not
     modified by this package (default).
  2. Record recent successful `error-monitoring` / Sentry health evidence,
     or explicitly record why a live check could not be performed.
- Expected result: no behavioral change to Sentry hourly monitoring;
  health noted without inventing a pass.
- Evidence: `VOC-081-EV-04`

## Cross-cutting validation (every repository task PR)

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
node --test scripts/foundation/*.test.mjs
# plus compose config dry-runs / disposable nginx -t for touched infra
```

Missing production credentials, SSH, or live Cloudflare access is a
recorded limitation — **never** a pass for T04 live clauses.
