# VOC-081 — Impact Analysis

## Security and privacy

- **Administrative UI exposure:** Restoring `monitor.vocanova.site` without
  an explicit access control would put an operational monitoring console
  behind Cloudflare proxy alone. Proxied DNS is not authorization
  (`VOC-081-D02` / `VOC-081-DEP-00`). Default recommendation: Cloudflare
  Access (or equivalent repository-managed control) plus Kuma's own auth.
- **TLS / secrets:** Shared edge continues read-only mounts of staging and
  production cert/conf trees. Monitor vhost must use production TLS paths
  already mounted (or a dedicated documented mount) without committing
  private keys. No monitoring network may mount staging/production secret
  trees.
- **Isolation:** Must not regress VOC-037-D01 / VOC-067-AC-03 /
  VOC-079-AC-05 write isolation or copy secrets across
  `/opt/vocanova/infra`, `/opt/vocanova/production`, and
  `/opt/vocanova/monitoring`.
- **Shared-edge fault domain:** Bad monitor vhost or failed reload can
  affect both app tiers. Mitigation: fail-closed `nginx -t` before
  reload/recreate; keep routine app deploys from recreating shared-edge.
- **Host publish:** Publishing Kuma on `0.0.0.0:3001` would bypass the
  edge and Cloudflare Access. Deterministic tests must forbid that.

No application user-data schema change. Kuma persistent data may contain
monitor targets and operational metadata — treat backup/evidence as
sensitive; redact as needed.

## Data and migrations

- **Preserve** `/opt/vocanova/monitoring/kuma-data` across repository
  adoption of Compose.
- T00 backup before first converge; non-destructive bind reuse.
- Rollback must not require wiping Kuma data to restore the prior edge
  behavior (rollback may detach routing while leaving data intact).
- No Postgres/Atlas application migrations.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** No product UI change. Upstream Kuma UI a11y is out of
  scope.

## Risks, dependencies, and evidence

- `VOC-081-R00`: **Public admin UI exposure** if DEP-00 is skipped or
  treated as “DNS is enough.” Mitigation: AC-03 blocks closure without an
  explicit access decision and verification.
- `VOC-081-R01`: **Kuma data loss** on naive Compose recreate or wrong
  volume path. Mitigation: T00 backup plan; preserve paths/names; TEST-00.
- `VOC-081-R02`: **Shared-edge outage** from bad vhost or network
  recreate. Mitigation: `nginx -t` fail-closed; controlled shared-edge
  convergence (`VOC-081-DEP-02`); monitoring window in T04.
- `VOC-081-R03`: **Compose ownership conflict** if staging/production
  orphan-removal or bundle sync touches monitoring/shared-edge incorrectly.
  Mitigation: scoped project commands; AC-04; TEST-05/06.
- `VOC-081-R04`: **Stale 30-monitor.conf confusion** left under production
  conf.d but never loaded (or accidentally loaded later). Mitigation:
  DEP-01 recommended shared/dedicated path; retire or inert the stale
  fragment via repository converge.
- `VOC-081-R05`: **WebSocket breakage** if upgrade headers missing —
  Kuma UI appears “up” on HTTP but realtime features fail. Mitigation:
  AC-02/AC-05; TEST-07.
- `VOC-081-R06`: **Second nginx / port regression** accidentally
  reintroduced while debugging. Mitigation: preserve VOC-079 single-edge
  invariants in tests; AC-04/AC-06.
- `VOC-081-DEP-00`: Access exposure policy.
- `VOC-081-DEP-01`: Vhost path / deploy ownership.
- `VOC-081-DEP-02`: Shared-edge network apply shape.
- `VOC-081-EV-00`: Kuma Compose + backup/migration evidence.
- `VOC-081-EV-01`: Network topology + non-public port evidence.
- `VOC-081-EV-02`: Vhost + access-control repository evidence.
- `VOC-081-EV-03`: Deploy workflow convergence evidence.
- `VOC-081-EV-04`: Live HTTPS/WebSocket, invariants, rollback, Sentry /
  error-monitoring health evidence.
