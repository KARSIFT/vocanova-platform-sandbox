# VOC-086 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → T03 → T04 → T05**.

## VOC-086-T00 — Canonical inventory schema, initial IDs, and validation tests

- Requirement source: issue #716; `VOC-086-D00`–`D02`; `VOC-086-D08`
- Acceptance criteria: `VOC-086-AC-00`, `VOC-086-AC-09` (inventory/docs stubs
  as needed)
- Tests: `VOC-086-TEST-00`, `VOC-086-TEST-01`
- Evidence: `VOC-086-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Add `infra/monitoring/monitors.yaml` with stable IDs and required metadata
   for:
   - staging web
   - staging API `/healthz`
   - production web (marked for adoption of the existing monitor)
   - production API `/healthz` (marked for adoption of the existing monitor)
   - independent `https://monitor.vocanova.site` reachability
2. Add synthetics registry entries for the five stable synthetic IDs from
   `VOC-086-D02` (same file or adjacent registry; record choice).
3. Include interval/timeout/retries/severity/expected status/body and
   feature/page/API coverage references on each entry.
4. Add deterministic schema and duplicate-ID validation tests.
5. Do not call live Kuma; do not add credentials; do not change Sentry.

### Explicitly out of scope for this task

- Socket.IO synchronizer (T01).
- Sync workflow / credential bootstrap (T02).
- Scheduled synthetics execution (T03).
- Governance field CI (T04).
- Live proof (T05).

## VOC-086-T01 — Idempotent Kuma Socket.IO synchronizer

- Requirement source: issue #716; `VOC-086-D03`, `VOC-086-D04`;
  `VOC-086-DEP-01`
- Acceptance criteria: `VOC-086-AC-01`, `VOC-086-AC-02`, `VOC-086-AC-03`
- Tests: `VOC-086-TEST-02`–`VOC-086-TEST-07`
- Evidence: `VOC-086-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-086-T00`

### Required work

1. Implement an idempotent synchronizer using authenticated Kuma Socket.IO
   events `add`, `editMonitor`, and monitor-list (exact client shape per
   `VOC-086-DEP-01`).
2. Validate the full inventory before mutation.
3. Create missing managed monitors; update changed managed monitors; adopt
   existing production monitors when inventory requests adoption.
4. Identify ownership via a stable repository marker/tag; never overwrite
   unrelated manual monitors unless explicitly adopted.
5. Detect collisions/duplicates and fail explicitly.
6. On partial failure, compensate/rollback applied operations; exit non-zero
   if apply or rollback is incomplete; never log credentials/tokens.
7. Never read/write SQLite.
8. Add protocol mocks/tests for create/update/adopt/manual-preservation,
   collisions, compensation, and credential redaction.

### Explicitly out of scope for this task

- GitHub workflow credential bootstrap (T02).
- Scheduled synthetics (T03).
- Governance CI (T04).
- Claiming live production inventory green without T05 evidence.

## VOC-086-T02 — Sync/deploy workflow and one-time credential bootstrap

- Requirement source: issue #716; `VOC-086-D05`; `VOC-086-DEP-02`
- Acceptance criteria: `VOC-086-AC-03`, `VOC-086-AC-04`
- Tests: `VOC-086-TEST-07`–`VOC-086-TEST-09`
- Evidence: `VOC-086-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-086-T01`

### Required work

1. Add an explicit repository workflow for monitoring inventory sync/deploy.
2. Read Kuma credentials only from GitHub secrets / workflow environment
   (names recorded per `VOC-086-DEP-02`).
3. Provide explicit `rotate_credentials` / bootstrap input that:
   - invokes Kuma's official `/app/extra/reset-password.js` through
     SSH/container stdin
   - generates/stores a strong credential without printing it
   - preserves the existing username
   - documents operator impact (existing sessions invalidated once)
4. Ensure normal sync runs never reset credentials.
5. Add workflow wiring and redaction tests; fail if secrets appear in fixture
   logs.
6. Do not commit secrets to git.

### Explicitly out of scope for this task

- Scheduled synthetics registry (T03).
- Governance field (T04).
- Full live AC-05/AC-10 closure proof (T05), though T02 may perform the first
  bootstrap when authorized and record redacted success.

## VOC-086-T03 — Canonical scheduled synthetics workflow/registry

- Requirement source: issue #716; `VOC-086-D02`, `VOC-086-D06`;
  `VOC-086-DEP-03`
- Acceptance criteria: `VOC-086-AC-06`, `VOC-086-AC-07` (non-regression half)
- Tests: `VOC-086-TEST-11`, `VOC-086-TEST-12`, `VOC-086-TEST-13`
- Evidence: `VOC-086-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-086-T00` (IDs); should follow T02 so
  operational monitoring split is documented alongside live sync path

### Required work

1. Add a canonical scheduled synthetic workflow/registry (not naive
   unauthenticated Kuma page checks) covering:
   - staging OAuth expected state / exact callback
   - production OAuth expected state / exact callback
   - non-empty production journey content and real details
   - critical authenticated staging journey
   - non-mutating production authenticated route/content sweep
2. Bind each check to its stable synthetic ID.
3. Reuse reserved synthetic accounts and existing mint secrets; mask session
   values; do not affect real users.
4. Keep Sentry/`error-monitoring.yml` separate; do not replace it.
5. Add deterministic mapping/wiring tests.
6. Prefer reusing existing deploy-time harness pieces where they already
   implement the scenarios; schedule them under stable IDs rather than
   inventing divergent duplicate logic without cause.

### Explicitly out of scope for this task

- Kuma availability monitor sync (T01/T02).
- Governance CI (T04).
- Live green proof for all synthetics (T05), except local/CI deterministic
  wiring.

## VOC-086-T04 — monitoring_impact governance and CI validation

- Requirement source: issue #716; `VOC-086-D07`, `VOC-086-D08`
- Acceptance criteria: `VOC-086-AC-08`, `VOC-086-AC-09` (governance docs/tests)
- Tests: `VOC-086-TEST-14`, `VOC-086-TEST-15`, `VOC-086-TEST-16`
- Evidence: `VOC-086-EV-04` (`t04-evidence.md`)
- Status: pending — depends on `VOC-086-T00` for canonical ID set

### Required work

1. Require `monitoring_impact` on newly created or newly modified change
   packages with states `none|existing|add|update`.
2. `none` requires rationale; other states require valid stable
   monitor/synthetic IDs from the canonical inventory/registry.
3. Wire CI/governance validation (expected under `scripts/governance/`, which
   establishes the R4 path floor).
4. Fail additions/changes to page routes or critical endpoints when monitoring
   impact is missing or invalid.
5. Grandfather historical unmodified packages without rewriting history.
6. Update the change-package template and any docs that describe package
   fields in the same PR (AGENTS.md doc-sync rule).
7. Add positive/negative deterministic tests.
8. Confirm this package's own `monitoring_impact: add` declaration remains
   valid against the landed schema.

### Explicitly out of scope for this task

- Live Kuma mutation.
- Changing historical package contents solely to satisfy the new field.

## VOC-086-T05 — Operator documentation and live verification proof

- Requirement source: issue #716 required design items 10–11; AC live half
- Acceptance criteria: `VOC-086-AC-05`, `VOC-086-AC-06` (live),
  `VOC-086-AC-07`, `VOC-086-AC-09` (docs), `VOC-086-AC-10`
- Tests: `VOC-086-TEST-10`, `VOC-086-TEST-12`, `VOC-086-TEST-17`
- Evidence: `VOC-086-EV-05` (`t05-evidence.md`)
- Status: pending — depends on T00–T04 merging and promoting through the
  normal release/deploy path as needed for live sync/synthetics

### Required work

1. Document adding monitoring for a page/API/feature, credential
   bootstrap/rotation, rollback, ownership, and alert/check proof.
2. Deploy/sync inventory through the normal workflow.
3. Confirm intended Kuma list/status via supported Socket.IO (redacted).
4. Manually dispatch scheduled synthetics and prove all checks green.
5. Independently verify `https://monitor.vocanova.site` reachability.
6. Confirm one shared-edge nginx, Kuma isolation, environment isolation, and
   absence of 8081/8443.
7. Confirm Sentry error-monitoring remains separate/healthy.
8. Record run URLs and redacted evidence in `t05-evidence.md`.

### Explicitly out of scope for this task

- Manual host SQLite edits.
- Cloudflare configuration changes.
- Real-user test activity.

## Task ordering notes

- T00 blocks ID-stable work for all later tasks.
- T01 blocks meaningful live sync; T02 blocks credentialed workflow apply.
- T03 can be developed against mocks after T00, but live synthetic proof is T05.
- T04 can proceed after T00; should not claim route fail-closed coverage until
  the validator lands.
- T05 is the package's live closure proof task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #716 is gated on AC results with evidence, not on task issue
  closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
