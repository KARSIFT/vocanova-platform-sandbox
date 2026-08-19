# VOC-086 — Impact Analysis

## Security and privacy

- **Monitoring credentials:** Kuma username/password are introduced into
  GitHub Actions secrets and the sync workflow environment. Bootstrap must
  generate a strong password without printing it, preserve the existing
  username, and mask secrets in logs. Existing Kuma sessions are invalidated
  once on bootstrap/rotation.
- **No SQLite path:** synchronizer and deploy/sync tooling must not read or
  write Kuma SQLite. Supported Socket.IO and the official password-reset tool
  are the only mutation paths.
- **Synthetic-only authenticated checks:** scheduled synthetics reuse reserved
  synthetic accounts and mint secrets; mask session cookies/tokens; do not
  mutate real users or perform state-changing learning actions in production
  sweeps.
- **Manual monitor preservation:** ownership marker/tag prevents accidental
  overwrite of unrelated operator-created monitors.
- **Isolation:** preserve staging/production secret files, directories, deploy
  users, databases, Docker networks, loopback-only Kuma `3001`, single
  shared-edge nginx, and absence of 8081/8443.
- **Sentry separation:** do not fold error monitoring into naive Kuma page
  checks; leave `error-monitoring.yml` healthy and separate.

## Data and migrations

- No application schema migration.
- Kuma monitor definitions change via Socket.IO create/update/adopt only.
- Credential bootstrap changes the Kuma admin password when explicitly
  requested; not a database migration.
- Rollback: revert repository inventory/workflow revisions and re-run
  compensating supported-protocol sync; preserve manually owned monitors;
  do not claim destructive wipe of Kuma data as the primary rollback.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability for product
  analytics instrumentation.
- **Accessibility:** No product UI redesign. Operator documentation only.

## Risks, dependencies, and evidence

- `VOC-086-R00`: **Partial apply leaves Kuma half-updated.** Mitigation:
  AC-02/TEST-05/TEST-06 prevalidation + compensation; non-zero exit.
- `VOC-086-R01`: **Credential exposure in logs/evidence.** Mitigation:
  AC-04/TEST-09 redaction; never print bootstrap password; mask secrets.
- `VOC-086-R02`: **Overwrite of unrelated manual monitors.** Mitigation:
  AC-01/TEST-04 ownership marker; adopt only when inventory explicitly
  requests adoption.
- `VOC-086-R03`: **SQLite shortcut regresses into deploy tooling.** Mitigation:
  AC-03/TEST-07 explicit ban and static/deterministic checks.
- `VOC-086-R04`: **Normal sync resets password unexpectedly.** Mitigation:
  AC-04/TEST-08; rotate only on explicit input.
- `VOC-086-R05`: **False confidence from deploy-only synthetics without
  scheduled stable IDs.** Mitigation: AC-06 scheduled registry + manual proof.
- `VOC-086-R06`: **Route/endpoint changes ship without monitoring coverage.**
  Mitigation: AC-08 governance fail-closed for new/changed packages.
- `VOC-086-R07`: **Topology/isolation regression while syncing monitors.**
  Mitigation: AC-10; reuse VOC-081 verifiers; no Cloudflare/manual host fixes.
- `VOC-086-R08`: **Risk under-declaration** if package stays at issue's R3
  while touching `scripts/governance/*`. Mitigation: draft proposes R4 from
  measured path floor; human confirms at adoption.

Dependencies: see `change.yaml` `VOC-086-DEP-00`–`DEP-03`, VOC-081, VOC-051,
VOC-050, VOC-067.

Evidence IDs:

- `VOC-086-EV-00` — T00 inventory/schema (`t00-evidence.md`)
- `VOC-086-EV-01` — T01 synchronizer (`t01-evidence.md`)
- `VOC-086-EV-02` — T02 workflow/credentials (`t02-evidence.md`)
- `VOC-086-EV-03` — T03 scheduled synthetics (`t03-evidence.md`)
- `VOC-086-EV-04` — T04 governance (`t04-evidence.md`)
- `VOC-086-EV-05` — T05 docs + live proof (`t05-evidence.md`)

## Monitoring impact (this package)

`add`: establishes canonical monitor/synthetic IDs and the
`monitoring_impact` governance mechanism. Declared in `change.yaml` under
`monitoring_impact` with the initial stable ID lists from `VOC-086-D01` /
`VOC-086-D02`.
