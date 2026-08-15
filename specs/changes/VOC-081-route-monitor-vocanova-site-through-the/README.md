# VOC-081 — Route monitor.vocanova.site through the repository-managed shared edge

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #665](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/665),
prepared for plan review and adoption under **active A-004** (proposed **R4**).

## Identity and lifecycle

- Package ID: VOC-081
- Title: Route monitor.vocanova.site through the repository-managed shared edge
- Canonical path:
  `specs/changes/VOC-081-route-monitor-vocanova-site-through-the`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R4` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R3** for `infra/*` and deploy workflows)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#665](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/665)
    (this package's requirement source)
  - [#624](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624) /
    [VOC-079](../VOC-079-retire-production-nginx-8443-bridge-and-complete)
    (single shared-edge cutover; recorded monitor 520 as out of scope)
  - [#485](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/485) /
    [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying)
    (shared-edge introduction)
  - [#301](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/301) /
    related governance context only as needed — not a requirement source

## Why this exists

VOC-079 completed the single-nginx cutover:
`vocanova-shared-edge-nginx` alone owns host `80`/`443`, and the production
`:8081`/`:8443` bridge is gone. VOC-079-EV-03 recorded that
`https://monitor.vocanova.site` still returns Cloudflare **520** while
Uptime Kuma remains healthy on host loopback `127.0.0.1:3001`.

Issue #665 verified the root cause end-to-end (2026-08-15):

- Cloudflare has a proxied A record for `monitor.vocanova.site` to
  `130.185.123.152`.
- Shared edge deliberately includes only production `10-*.conf` /
  `20-*.conf`, so the live stale
  `/opt/vocanova/production/nginx/conf.d/30-monitor.conf` is never loaded.
- Even if that fragment were loaded, shared edge joins only `vocanova-net`
  and `vocanova-production-net`, while Kuma joins only `monitoring_default`
  — no container DNS/network path.
- The live Compose project `monitoring` / container `vocanova-uptime-kuma`
  and data under `/opt/vocanova/monitoring/kuma-data` are not represented
  in this repository.

This package restores the monitor hostname **through the single
repository-managed shared edge**, without reintroducing a second nginx or
accepting manual SSH / ad hoc Docker mutations as the acceptance path.

## What this package does

1. **Repository source of truth for Kuma** (`VOC-081-T00`): add a
   repository-managed monitoring Compose definition that preserves the
   existing Compose project/container identity and
   `/opt/vocanova/monitoring/kuma-data`, with an explicit backup/migration
   plan before any converge.
2. **Least-privilege network path** (`VOC-081-T01`): introduce a dedicated
   external monitoring Docker network between shared edge and Kuma; keep
   Kuma off public host interfaces; do not weaken staging/production secret
   isolation.
3. **Shared-edge monitor vhost + access control** (`VOC-081-T02`): add a
   repository-owned `monitor.vocanova.site` vhost loaded by shared edge
   (TLS, Cloudflare real-IP inheritance, WebSocket and reverse-proxy
   headers) and implement the adopted access-exposure decision
   (`VOC-081-DEP-00`).
4. **Deploy convergence** (`VOC-081-T03`): extend the normal repository
   deploy path so it creates/converges the monitoring network and Kuma
   service and safely applies required shared-edge network/config changes
   with fail-closed `nginx -t` before reload/recreate — no manual
   `docker network connect`, SSH file edit, or ad hoc container removal.
5. **Live verification and rollback** (`VOC-081-T04`): deploy via the
   normal workflow; prove public HTTPS/WebSocket through Cloudflare;
   confirm single-nginx / no public Kuma port / isolation invariants;
   record monitoring window and rollback evidence; confirm Sentry /
   `error-monitoring` remain unchanged and healthy.

## What this package deliberately does NOT do

- Not reintroducing `vocanova-production-nginx` or ports `8081`/`8443`.
- Not replacing Sentry or Uptime Kuma.
- Not unrelated application, cookie, or staging-product changes.
- Not treating a Cloudflare proxied DNS record as authorization for the
  admin UI (`VOC-081-DEP-00`).
- Not manual SSH edits, `docker network connect`, or ad hoc
  `docker stop/rm` as the acceptance path.
- Not changing the hourly `.github/workflows/error-monitoring.yml`
  behavior except to record that it remains healthy.
- Not adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Access policy for the Kuma UI (`VOC-081-DEP-00`) — private via
   Cloudflare Access (recommended) vs explicit public-exposure acceptance.
2. Monitor vhost path / deploy ownership (`VOC-081-DEP-01`).
3. Controlled shared-edge network convergence shape (`VOC-081-DEP-02`).
4. Accept proposed **R4** (path floor R3; semantic elevation for live
   public edge + administrative UI exposure).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R4 remains a strengthened evidence class.
`automatic_merge_allowed: true` is set per AGENTS.md. This draft still
carries no adoption or implementation authority — plan review, adoption,
independent verification of each exact revision, and production outcome
evidence remain distinct gates.
