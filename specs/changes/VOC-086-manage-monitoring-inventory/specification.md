# VOC-086 — Manage monitoring inventory, synthetics, and monitoring-impact governance: Specification

## Objective and requirement source

Close the gap reported in
[GitHub issue #716](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/716):
make monitoring definitions and monitoring-impact governance
repository-managed, then deploy and prove the intended Kuma and scheduled
synthetic coverage without direct SQLite mutation or credential exposure.

Authority context recorded in the issue: founder/user remediation directive
dated 2026-08-16 authorizes issue/package/PR/workflow/deploy/verification/
remediation/closure actions without routine approval. **This draft package
still does not adopt or authorize itself**; adoption remains a separate
A-004 plan-review / adopt path.

Primary context (issue #716 + drafting-time repo read):

| Item | Value |
|------|--------|
| Promoted revision (evidence) | `96a1fd10155c605ba444e624928552f7b62b93d4` |
| Kuma | 1.23.17, `louislam/uptime-kuma:1`, `127.0.0.1:3001` only |
| Topology | Isolated monitoring network + shared edge (VOC-081) |
| Repository ownership today | Compose, vhost, access policy, reachability verifier |
| Missing today | Monitor inventory, Socket.IO synchronizer, scheduled synthetic registry, `monitoring_impact` governance |
| Live Kuma monitors | Two active 60s HTTP: Production API `/healthz`, Production Web |
| Deploy synthetics | Staging OAuth/core loop; production OAuth/content/authenticated routes (deploy-only) |
| Sentry | Hourly `error-monitoring` healthy (leave separate) |
| Kuma credentials in GitHub | None yet (smoke-mint/Sentry tokens exist) |
| Supported mutation path | Authenticated Socket.IO `add` / `editMonitor` / monitor-list; official `/app/extra/reset-password.js` |

**Objective:** after this package's implementation, the repository owns a
canonical monitor/synthetic inventory with stable IDs; an idempotent Socket.IO
synchronizer applies managed monitors with compensation on partial failure; an
explicit workflow bootstraps/rotates credentials only when requested; scheduled
synthetics cover the required authenticated scenarios; change packages declare
`monitoring_impact` and CI enforces it for new/changed packages (historical
packages grandfathered); and live proof confirms Kuma inventory/status,
synthetics, monitor-host reachability, and isolation invariants.

## Confirmed findings (issue #716 + drafting-time re-read)

- `infra/docker-compose.monitoring.yml` already defines the monitoring
  Compose project, loopback publish, and `vocanova-monitoring-net`.
- `infra/monitoring/access-policy.md` records public Kuma login protected by
  Kuma's own authentication (VOC-081-DEP-00).
- Reachability verifier scripts from VOC-081 exist; they do not manage monitor
  inventory.
- No `infra/monitoring/monitors.yaml` exists yet.
- Deploy workflows already mint synthetic sessions and run deploy-time
  smoke/core-loop checks (VOC-050 / later production content gates); those
  checks are not a scheduled stable-ID synthetic inventory.
- `.github/workflows/error-monitoring.yml` is the Sentry path and must remain
  unchanged by this package's tasks except where docs map coverage.
- Path classification of the expected file set yields **R4** when
  `scripts/governance/*` is included; issue #716 proposed R3.

## Scope and non-goals

In scope:

1. Canonical `infra/monitoring/monitors.yaml` with stable ID, name,
   environment, owner, type, URL/synthetic workflow reference, expected
   status/body, interval, timeout/retries, severity, and feature/page/API
   coverage references.
2. Idempotent Kuma synchronizer using authenticated Socket.IO:
   validate full inventory before mutation; create missing managed monitors;
   update changed managed monitors; identify ownership by a stable repository
   marker/tag; never overwrite unrelated manual monitors unless explicitly
   adopted; detect collisions/duplicates; compensate/rollback applied
   operations on partial failure and fail non-zero if apply or rollback is
   incomplete; never read/write SQLite; never log credentials/tokens.
3. Explicit repository workflow for sync/deploy. Credentials live only in
   GitHub secrets and the workflow environment.
4. Because credentials do not presently exist, an explicit one-time
   `rotate_credentials` / bootstrap input that invokes Kuma's shipped
   official password-reset tool through SSH/container stdin, then
   authenticates via Socket.IO. Normal runs must never reset credentials.
   Generate/store the initial strong credential without printing it; preserve
   the existing username; document that existing sessions are invalidated
   once.
5. Initial Kuma availability monitors:
   - staging web
   - staging API `/healthz`
   - production web (adopt existing)
   - production API `/healthz` (adopt existing)
   - independent `https://monitor.vocanova.site` reachability
6. Canonical scheduled synthetic workflow/registry for:
   - staging OAuth expected state / exact callback
   - production OAuth expected state / exact callback
   - non-empty production journey content and real details
   - critical authenticated staging journey
   - non-mutating production authenticated route/content sweep
   Reuse reserved synthetic accounts and mint secrets; mask session values;
   do not affect real users.
7. Keep Kuma for availability/TLS/basic API health, scheduled workflows for
   authenticated behavior, and Sentry for errors. Every check has a stable ID
   and coverage mapping.
8. Governance: required `monitoring_impact` in new/changed change packages
   with `none|existing|add|update`; `none` requires rationale; other states
   require valid stable monitor/synthetic IDs; CI validates; additions/changes
   to page routes or critical endpoints fail when monitoring impact is missing
   or invalid; grandfather historical packages without rewriting history.
9. Deterministic protocol mocks/tests for inventory schema,
   create/update/adopt/manual-preservation, collisions, partial-failure
   compensation, credential redaction, workflow wiring, synthetic mappings,
   and governance positive/negative cases.
10. Documentation for adding monitoring for a page/API/feature, credential
    bootstrap/rotation, rollback, ownership, and alert/check proof.
11. Deploy inventory through the normal workflow, confirm intended Kuma
    list/status via supported Socket.IO, run scheduled synthetics manually and
    prove all checks green, verify monitor hostname independently, and preserve
    single-edge/network/secret isolation.

Non-goals / explicitly excluded:

- Manual server fixes outside the repository workflow path.
- Direct SQLite read/write for deployment or sync.
- Cloudflare configuration changes.
- Naive authenticated-page HTTP monitors in Kuma.
- Credential logging or printing bootstrap secrets to logs/evidence.
- Overwriting unrelated manually owned monitors without explicit adoption.
- Changing Sentry/`error-monitoring.yml` semantics.
- Application schema migrations or real-user mutation.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Builder/issue proposal:** R3 (issue #716).
- **Measured path floor:** R4 when `scripts/governance/*` is in the task file
  set; R3 for monitoring workflows / `infra/monitoring` alone.
- **Draft package proposal:** **R4** (highest of builder and measured floor).
  This is a draft proposal for the reviewing human at adoption time, never a
  determination.
- Protected areas: `.github/workflows/*`, `scripts/governance/*`, monitoring
  credentials, live Kuma inventory mutation, shared-edge / monitoring topology
  (must not regress VOC-081 / VOC-067 invariants).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate; R4 still requires strengthened evidence.

## Decisions

`VOC-086-D00`: Canonical inventory lives at `infra/monitoring/monitors.yaml`
(plus a co-located or adjacent synthetics registry file if needed). Every
managed check has a stable repository ID.

`VOC-086-D01`: Initial stable availability monitor IDs (proposed; implementer
may adjust naming only if documentation, governance validator, and inventory
stay consistent):

- `kuma.availability.staging.web`
- `kuma.availability.staging.api-healthz`
- `kuma.availability.production.web` (adopt existing Production Web)
- `kuma.availability.production.api-healthz` (adopt existing Production API `/healthz`)
- `kuma.availability.monitor-host`

`VOC-086-D02`: Initial stable synthetic IDs:

- `synthetic.staging.oauth-expected-state`
- `synthetic.production.oauth-expected-state`
- `synthetic.production.journey-content`
- `synthetic.staging.authenticated-core-journey`
- `synthetic.production.authenticated-route-content-sweep`

`VOC-086-D03`: Synchronizer uses authenticated Kuma Socket.IO only
(`add`, `editMonitor`, monitor-list). No SQLite. Ownership is marked with a
stable repository tag/marker embedded in managed monitor metadata. Unrelated
manual monitors are preserved unless an inventory entry explicitly adopts them.

`VOC-086-D04`: Apply algorithm is fail-closed: full inventory prevalidation;
then apply; on partial failure, compensate/rollback applied operations; exit
non-zero if apply or rollback is incomplete; never emit credentials in logs.

`VOC-086-D05`: Credential bootstrap/rotation is an explicit workflow input
(recommended name `rotate_credentials`). It invokes Kuma's official
`/app/extra/reset-password.js` via SSH/container stdin, generates a strong
password without printing it, stores it in GitHub secrets, preserves the
existing username, and documents one-time session invalidation. Normal sync
never resets credentials.

`VOC-086-D06`: Responsibility split — Kuma = availability/TLS/basic API
health; scheduled synthetics = authenticated behavior; Sentry =
error-monitoring. Docs and coverage maps must state this split.

`VOC-086-D07`: `monitoring_impact` is required on newly created or newly
modified change packages with states `none|existing|add|update`. `none`
requires rationale. Other states require valid stable IDs present in the
canonical inventory/registry. Route or critical-endpoint additions/changes
fail CI when impact is missing/invalid. Historical unmodified packages are
grandfathered.

`VOC-086-D08`: This package's own `monitoring_impact.state` is `add` because
it establishes the canonical IDs and the governance mechanism.

## Contradictions / open questions

1. **Risk class:** issue #716 proposes R3; measured path floor is R4 when
   governance validators under `scripts/governance/` are in scope. Draft
   proposal is R4. Reviewing human must accept R4 or relocate the validator
   and record the amended floor in writing — do not silently weaken
   `scripts/governance/*` path rules.
2. **Synchronizer runtime** (`VOC-086-DEP-01`): Node/TypeScript vs Python (or
   equivalent) is an implementer choice if AC/tests hold.
3. **Secret and environment names** (`VOC-086-DEP-02`): recommended
   `KUMA_USERNAME` / `KUMA_PASSWORD` in the monitoring sync environment;
   confirm at adoption or in T02 evidence.
4. **Scheduled synthetics packaging** (`VOC-086-DEP-03`): matrix vs
   workflow_call jobs; must preserve stable IDs and reuse mint secrets.
5. **Adoption matching for existing monitors:** exact match keys for adopting
   the two live production monitors (name, URL, or both) should be recorded in
   T01 evidence if the inventory fields alone are ambiguous.

## Security and privacy

- Credentials only in GitHub secrets / workflow environment; never committed.
- Bootstrap must not print the generated password; mask secrets in logs.
- Synchronizer and synthetics must redact tokens/cookies/session values.
- Synthetics use reserved synthetic accounts only; no real-user mutation.
- No SQLite access from deploy/sync tooling.
- Preserve Kuma authentication as the admin boundary
  (`infra/monitoring/access-policy.md`).

## Data, migrations, analytics, and accessibility

- **Data/migrations:** None for application databases. Kuma monitor
  definitions change only via supported Socket.IO. Credential bootstrap
  changes Kuma admin password once when explicitly requested.
- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** No product UI redesign. Monitoring docs and workflows
  only.
