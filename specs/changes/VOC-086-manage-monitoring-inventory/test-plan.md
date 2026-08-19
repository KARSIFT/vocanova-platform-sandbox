# VOC-086 — Test Plan

## VOC-086-TEST-00 — Inventory schema accepts complete entries

- Covers: `VOC-086-AC-00`
- Preconditions: `VOC-086-T00` tree; deterministic schema harness; no live Kuma
- Procedure:
  1. Load `infra/monitoring/monitors.yaml` (and synthetics registry).
  2. Assert all five availability IDs and five synthetic IDs are present.
  3. Assert required metadata fields exist on each entry.
- Expected result: schema validation passes for the canonical inventory.
- Evidence: `VOC-086-EV-00`

## VOC-086-TEST-01 — Inventory schema rejects missing fields and duplicate IDs

- Covers: `VOC-086-AC-00`
- Preconditions: `VOC-086-T00` tree
- Procedure:
  1. Fixture inventories missing required fields.
  2. Fixture inventories with duplicate stable IDs.
  3. Assert deterministic non-zero failure with explicit messages (no secrets).
- Expected result: invalid inventories fail closed.
- Evidence: `VOC-086-EV-00`

## VOC-086-TEST-02 — Synchronizer creates missing managed monitors

- Covers: `VOC-086-AC-01`
- Preconditions: `VOC-086-T01` tree; Socket.IO protocol mock
- Procedure:
  1. Start with empty managed set in the mock.
  2. Run sync against inventory containing managed monitors.
  3. Assert `add` (or equivalent create) calls for each missing managed entry.
- Expected result: creates exactly the missing managed monitors.
- Evidence: `VOC-086-EV-01`

## VOC-086-TEST-03 — Synchronizer updates changed managed monitors and is idempotent

- Covers: `VOC-086-AC-01`
- Preconditions: `VOC-086-T01` tree; protocol mock
- Procedure:
  1. Seed mock with managed monitors matching inventory, then change one
     inventory field.
  2. Run sync; assert update/`editMonitor` for the changed monitor only.
  3. Re-run sync unchanged; assert no mutating calls (idempotent success).
- Expected result: update-on-change; no-op when already matched.
- Evidence: `VOC-086-EV-01`

## VOC-086-TEST-04 — Adopt existing + preserve unrelated manual monitors

- Covers: `VOC-086-AC-01`
- Preconditions: `VOC-086-T01` tree; protocol mock
- Procedure:
  1. Seed mock with two existing production monitors (web + API `/healthz`)
     lacking ownership marker, plus one unrelated manual monitor.
  2. Inventory marks the two production monitors for adoption and does not
     list the manual monitor.
  3. Run sync; assert adoption updates ownership/metadata for the two, and
     zero mutate calls against the unrelated manual monitor.
- Expected result: adopt when requested; preserve manuals otherwise.
- Evidence: `VOC-086-EV-01`

## VOC-086-TEST-05 — Full prevalidation blocks mutation on invalid inventory

- Covers: `VOC-086-AC-02`
- Preconditions: `VOC-086-T01` tree
- Procedure:
  1. Fixture an invalid inventory (schema/collision).
  2. Run sync.
  3. Assert no Socket.IO mutate calls occurred and exit is non-zero.
- Expected result: validate-before-mutate.
- Evidence: `VOC-086-EV-01`

## VOC-086-TEST-06 — Partial failure compensation and incomplete rollback fail non-zero

- Covers: `VOC-086-AC-02`
- Preconditions: `VOC-086-T01` tree; protocol mock that fails mid-apply and/or
  mid-compensation
- Procedure:
  1. Drive a mid-apply failure after at least one successful mutate.
  2. Assert compensation/rollback attempts for applied operations.
  3. Drive a compensation failure; assert non-zero exit and explicit error
     without secrets.
- Expected result: no silent partial apply; incomplete rollback is failure.
- Evidence: `VOC-086-EV-01`

## VOC-086-TEST-07 — No SQLite read/write path in sync tooling

- Covers: `VOC-086-AC-03`
- Preconditions: `VOC-086-T01` / `VOC-086-T02` tree
- Procedure:
  1. Static or harness check that synchronizer/workflow helpers do not
     reference `kuma.db`, `sqlite`, or direct `/app/data` DB mutation APIs
     for sync/deploy (password-reset tool invocation via official script is
     allowed for bootstrap only).
  2. Fail if a SQLite deployment mechanism is introduced.
- Expected result: Socket.IO (+ official reset tool for bootstrap) only.
- Evidence: `VOC-086-EV-01`, `VOC-086-EV-02`

## VOC-086-TEST-08 — rotate_credentials is opt-in; normal sync never resets

- Covers: `VOC-086-AC-04`
- Preconditions: `VOC-086-T02` tree; workflow fixture
- Procedure:
  1. Assert default/normal sync path does not invoke password reset.
  2. Assert only the explicit bootstrap/rotation input triggers
     `/app/extra/reset-password.js`.
- Expected result: rotation is explicit-only.
- Evidence: `VOC-086-EV-02`

## VOC-086-TEST-09 — Credential redaction and no plaintext bootstrap print

- Covers: `VOC-086-AC-04`
- Preconditions: `VOC-086-T02` tree
- Procedure:
  1. Fixture bootstrap/sync log streams containing password/token-like values.
  2. Assert redaction masks them.
  3. Assert bootstrap path stores secret via GitHub secret APIs/mechanisms
     without echoing the generated password to stdout fixtures.
- Expected result: secrets never appear plaintext in logged fixtures.
- Evidence: `VOC-086-EV-02`

## VOC-086-TEST-10 — Live Socket.IO monitor-list matches five intended monitors

- Covers: `VOC-086-AC-05`, `VOC-086-AC-10`
- Preconditions: `VOC-086-T05` after successful sync; live credentials in
  workflow environment only
- Procedure:
  1. Authenticate via Socket.IO using workflow secrets (not printed).
  2. Fetch monitor-list; assert the five availability monitors and metadata.
  3. Record redacted evidence.
- Expected result: live managed set matches inventory intent.
- Evidence: `VOC-086-EV-05`

## VOC-086-TEST-11 — Scheduled synthetics map to required stable IDs

- Covers: `VOC-086-AC-06`
- Preconditions: `VOC-086-T03` tree
- Procedure:
  1. Inspect scheduled synthetics workflow/registry.
  2. Assert jobs/entries exist for both OAuth checks, production journey
     content/details, authenticated staging journey, and non-mutating
     production route/content sweep under the stable IDs.
- Expected result: complete stable-ID mapping.
- Evidence: `VOC-086-EV-03`

## VOC-086-TEST-12 — Synthetic checks reuse mint secrets and mask sessions

- Covers: `VOC-086-AC-06`
- Preconditions: `VOC-086-T03` tree; workflow/script fixtures
- Procedure:
  1. Assert reuse of reserved synthetic accounts / existing mint secret names.
  2. Assert session cookie/token values are masked in logs.
  3. Assert production sweep uses non-mutating methods only.
- Expected result: synthetic-only, masked, non-mutating production sweep.
- Evidence: `VOC-086-EV-03`, `VOC-086-EV-05`

## VOC-086-TEST-13 — error-monitoring.yml remains separate

- Covers: `VOC-086-AC-07`
- Preconditions: `VOC-086-T03` / `VOC-086-T05` tree
- Procedure:
  1. Diff or assert `.github/workflows/error-monitoring.yml` is not replaced
     by Kuma page-check logic in this package's changes.
  2. Record healthy/separate posture in evidence.
- Expected result: Sentry path unchanged and still the error channel.
- Evidence: `VOC-086-EV-05`

## VOC-086-TEST-14 — monitoring_impact positive cases

- Covers: `VOC-086-AC-08`
- Preconditions: `VOC-086-T04` tree; disposable package fixtures
- Procedure:
  1. Fixture packages with `none` + rationale; `existing`/`add`/`update` +
     valid IDs.
  2. Run governance validator; assert pass.
- Expected result: valid declarations accepted.
- Evidence: `VOC-086-EV-04`

## VOC-086-TEST-15 — monitoring_impact negative cases

- Covers: `VOC-086-AC-08`
- Preconditions: `VOC-086-T04` tree
- Procedure:
  1. Fixture missing field, `none` without rationale, invalid IDs, unknown
     state.
  2. Assert deterministic failure.
- Expected result: invalid declarations rejected.
- Evidence: `VOC-086-EV-04`

## VOC-086-TEST-16 — Route/critical-endpoint changes fail without valid impact

- Covers: `VOC-086-AC-08`
- Preconditions: `VOC-086-T04` tree
- Procedure:
  1. Fixture a package touching page routes or critical API endpoints with
     missing/invalid `monitoring_impact`.
  2. Assert CI/governance check fails.
  3. Fixture the same with valid `add`/`update`/`existing` IDs; assert pass.
  4. Assert an unmodified historical package fixture remains grandfathered.
- Expected result: fail closed for new route/endpoint impact gaps;
  grandfather historical.
- Evidence: `VOC-086-EV-04`

## VOC-086-TEST-17 — Live monitor-host reachability and isolation invariants

- Covers: `VOC-086-AC-10`
- Preconditions: `VOC-086-T05`; reuse VOC-081 verifier where applicable
- Procedure:
  1. Run independent `monitor.vocanova.site` reachability verification.
  2. Confirm single shared-edge nginx, Kuma isolation, environment isolation,
     and absence of 8081/8443.
  3. Record redacted evidence.
- Expected result: hostname reachable; topology/isolation intact.
- Evidence: `VOC-086-EV-05`

Include positive, negative, authorization, failure, and rollback coverage as
above. Tests must not embed real secrets or production personal data.
