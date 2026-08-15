# VOC-086 — Acceptance Criteria

## VOC-086-AC-00 — Canonical inventory and stable IDs validate deterministically

- Requirement source: issue #716; `VOC-086-D00`–`D02`
- Tasks: `VOC-086-T00`
- Tests: `VOC-086-TEST-00`, `VOC-086-TEST-01`
- Evidence: `VOC-086-EV-00`
- Result: pending

Observable outcome:

1. `infra/monitoring/monitors.yaml` (and synthetics registry) exists with the
   five availability monitor IDs and five synthetic IDs from `VOC-086-D01` /
   `VOC-086-D02` (or consistently renamed equivalents recorded in evidence).
2. Each entry includes stable ID, name, environment, owner, type,
   URL/synthetic workflow reference, expected status/body, interval,
   timeout/retries, severity, and feature/page/API coverage references.
3. A deterministic validator rejects missing/invalid required fields and
   duplicate IDs.

## VOC-086-AC-01 — Socket.IO sync is idempotent; managed create/update/adopt works; manuals preserved

- Requirement source: issue #716; `VOC-086-D03`
- Tasks: `VOC-086-T01`
- Tests: `VOC-086-TEST-02`, `VOC-086-TEST-03`, `VOC-086-TEST-04`
- Evidence: `VOC-086-EV-01`
- Result: pending

Observable outcome:

1. Synchronizer creates missing managed monitors and updates changed managed
   monitors via authenticated Socket.IO.
2. Existing production web and production API `/healthz` monitors can be
   adopted under stable IDs without duplicate unmanaged copies.
3. Unrelated manually owned monitors without the repository ownership marker
   are left untouched unless an inventory entry explicitly adopts them.
4. Re-running sync against an already-applied inventory is a no-op success
   (idempotent).

## VOC-086-AC-02 — Prevalidation and compensation prevent unreported partial apply

- Requirement source: issue #716; `VOC-086-D04`
- Tasks: `VOC-086-T01`
- Tests: `VOC-086-TEST-05`, `VOC-086-TEST-06`
- Evidence: `VOC-086-EV-01`
- Result: pending

Observable outcome:

1. Full inventory is validated before any mutation.
2. On partial failure after mutation begins, previously applied operations in
   that run are compensated/rolled back via supported protocol operations.
3. Process exits non-zero when apply or rollback is incomplete.
4. Failure output is explicit and contains no secrets.

## VOC-086-AC-03 — No SQLite read/write in deployment/sync

- Requirement source: issue #716; `VOC-086-D03`
- Tasks: `VOC-086-T01`, `VOC-086-T02`
- Tests: `VOC-086-TEST-07`
- Evidence: `VOC-086-EV-01`, `VOC-086-EV-02`
- Result: pending

Observable outcome:

1. Synchronizer and sync workflow paths do not open, copy, or mutate
   `kuma.db` / SQLite files.
2. Deterministic tests or static checks fail if a SQLite deployment path is
   introduced.

## VOC-086-AC-04 — Credential bootstrap is explicit one-time/rotation-only

- Requirement source: issue #716; `VOC-086-D05`
- Tasks: `VOC-086-T02`
- Tests: `VOC-086-TEST-08`, `VOC-086-TEST-09`
- Evidence: `VOC-086-EV-02`
- Result: pending

Observable outcome:

1. Workflow exposes an explicit `rotate_credentials` (or equivalent) input.
2. When set, workflow invokes Kuma's official password-reset tool via
   SSH/container stdin, generates/stores a strong credential without printing
   it, preserves the existing username, and documents one-time session
   invalidation.
3. Normal sync runs never reset credentials.
4. Credentials exist only as GitHub secrets / workflow environment values and
   are masked; redaction tests fail if plaintext secrets appear in logged
   output fixtures.

## VOC-086-AC-05 — Kuma contains the five intended availability monitors with metadata

- Requirement source: issue #716; `VOC-086-D01`
- Tasks: `VOC-086-T00`, `VOC-086-T01`, `VOC-086-T02`, `VOC-086-T05`
- Tests: `VOC-086-TEST-10`
- Evidence: `VOC-086-EV-05`
- Result: pending

Observable outcome:

1. After sync, Kuma (via supported Socket.IO monitor-list) shows the five
   intended availability monitors with configured status/body/interval/
   timeout/retries/severity metadata matching inventory.
2. Live proof is recorded without secrets.

## VOC-086-AC-06 — Scheduled stable-ID synthetics cover required scenarios

- Requirement source: issue #716; `VOC-086-D02`, `VOC-086-D06`
- Tasks: `VOC-086-T03`, `VOC-086-T05`
- Tests: `VOC-086-TEST-11`, `VOC-086-TEST-12`
- Evidence: `VOC-086-EV-03`, `VOC-086-EV-05`
- Result: pending

Observable outcome:

1. Scheduled synthetic registry/workflow covers both OAuth expected-state/
   exact-callback checks, production journey content/details, authenticated
   staging journey, and non-mutating production route/content sweep under the
   stable IDs.
2. Checks reuse reserved synthetic accounts and mint secrets; session values
   are masked; real users are not affected.
3. Manual dispatch can prove all checks green.

## VOC-086-AC-07 — Sentry error-monitoring remains separate and healthy

- Requirement source: issue #716; `VOC-086-D06`
- Tasks: `VOC-086-T03`, `VOC-086-T05`
- Tests: `VOC-086-TEST-13`
- Evidence: `VOC-086-EV-05`
- Result: pending

Observable outcome:

1. `.github/workflows/error-monitoring.yml` remains the Sentry path and is not
   replaced by Kuma page checks.
2. Evidence notes the latest healthy error-monitoring posture (or an explicit
   non-regression check) without requiring this package to change Sentry
   logic.

## VOC-086-AC-08 — monitoring_impact governance validates and fails closed for routes/endpoints

- Requirement source: issue #716; `VOC-086-D07`, `VOC-086-D08`
- Tasks: `VOC-086-T04`
- Tests: `VOC-086-TEST-14`, `VOC-086-TEST-15`, `VOC-086-TEST-16`
- Evidence: `VOC-086-EV-04`
- Result: pending

Observable outcome:

1. New/changed packages must declare `monitoring_impact` with
   `none|existing|add|update`.
2. `none` requires rationale; other states require valid stable IDs from the
   canonical inventory/registry.
3. CI rejects missing/invalid declarations.
4. Additions/changes to page routes or critical endpoints fail when monitoring
   impact is missing or invalid.
5. Historical unmodified packages are grandfathered (no history rewrite).

## VOC-086-AC-09 — Documentation and deterministic tests cover developer workflow and failure paths

- Requirement source: issue #716 required design items 9–10
- Tasks: `VOC-086-T00`–`VOC-086-T05`
- Tests: `VOC-086-TEST-00`–`VOC-086-TEST-16` (as applicable per task)
- Evidence: `VOC-086-EV-00`–`VOC-086-EV-05`
- Result: pending

Observable outcome:

1. Docs explain adding monitoring for a page/API/feature, credential
   bootstrap/rotation, rollback, ownership, and alert/check proof.
2. Deterministic tests cover schema, sync behaviors, collisions, compensation,
   redaction, workflow wiring, synthetic mappings, and governance
   positive/negative cases.

## VOC-086-AC-10 — Live inventory/status, synthetics, monitor-host reachability, and isolation

- Requirement source: issue #716 acceptance + required design item 11
- Tasks: `VOC-086-T05`
- Tests: `VOC-086-TEST-10`, `VOC-086-TEST-12`, `VOC-086-TEST-17`
- Evidence: `VOC-086-EV-05`
- Result: pending

Observable outcome:

1. Live Kuma inventory/status via Socket.IO matches the intended managed set.
2. Manually dispatched scheduled synthetics pass.
3. `https://monitor.vocanova.site` is independently reachable per existing
   verifier semantics.
4. One shared-edge nginx, Kuma isolation, environment isolation, and absence
   of 8081/8443 remain intact.
