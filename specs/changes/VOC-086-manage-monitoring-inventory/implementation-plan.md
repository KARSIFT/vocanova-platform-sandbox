# VOC-086 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status:
adopted`, `approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a founder
`approved` comment; exact-revision plan review and deterministic gates still
apply.

Additionally:

- Treat issue #716's verified evidence as the starting point
  (`VOC-086-DEP-00`). If implementer evidence shows Socket.IO events or the
  password-reset tool differ from the read-only confirmation, stop and record
  before expanding scope.
- Record synchronizer runtime (`VOC-086-DEP-01`), secret/environment names
  (`VOC-086-DEP-02`), and scheduled-synthetics packaging (`VOC-086-DEP-03`) in
  the relevant task evidence.
- Path floor **R4** when governance scripts are touched; proposed package risk
  **R4**. Strengthened evidence and rollback credibility required; no
  founder-comment merge gate under A-004.
- Preserve VOC-081 monitoring topology and VOC-067 isolation invariants.
- Never commit secrets, mint tokens, session values, or Kuma passwords into
  evidence files.
- Never introduce SQLite-based deploy/sync.

## File reconciliation and implementation sequence

1. **`VOC-086-T00` — Canonical inventory + schema validation**
   - Add `infra/monitoring/monitors.yaml` with the five availability monitors
     and required metadata fields.
   - Add synthetics registry entries for the five stable synthetic IDs
     (same file or adjacent registry — record choice in evidence).
   - Add deterministic schema/duplicate-ID validation tests.
   - Do not mutate live Kuma in this task.

2. **`VOC-086-T01` — Socket.IO synchronizer + protocol mocks**
   - Implement idempotent sync using authenticated Socket.IO
     (`add`, `editMonitor`, monitor-list).
   - Ownership marker/tag; create/update/adopt; preserve manuals; collisions;
     full prevalidation; compensating rollback; non-zero incomplete apply.
   - Ban SQLite access; redact credentials in logs.
   - Protocol mocks/tests for the behaviors above.
   - May run against disposable mocks only until T02/T05 authorize live sync.

3. **`VOC-086-T02` — Sync workflow + credential bootstrap/rotation**
   - Add explicit repository workflow for inventory sync/deploy.
   - Wire GitHub secrets / environment (names per `VOC-086-DEP-02`).
   - Implement `rotate_credentials` (or equivalent) input that invokes
     `/app/extra/reset-password.js` via SSH/container stdin, stores the
     generated secret without printing it, preserves username, documents
     session invalidation.
   - Ensure normal sync never resets credentials.
   - Workflow wiring + redaction tests.

4. **`VOC-086-T03` — Scheduled synthetics registry/workflow**
   - Add canonical scheduled synthetic workflow covering the five scenarios.
   - Reuse reserved synthetic accounts and existing mint secrets; mask
     session values; production checks remain non-mutating.
   - Map each job to stable synthetic IDs; document relationship to deploy-time
     checks without deleting necessary deploy gates prematurely unless
     explicitly replaced with equivalent coverage.
   - Deterministic mapping/wiring tests.

5. **`VOC-086-T04` — monitoring_impact governance**
   - Require `monitoring_impact` on new/changed packages
     (`none|existing|add|update`).
   - Validate rationale/IDs; fail route/critical-endpoint changes without valid
     coverage; grandfather historical unmodified packages.
   - Wire into governance/CI validation under `scripts/governance/` (R4 path).
   - Update change-package template / operator docs that describe package
     fields in the same PR as behavior changes (AGENTS.md doc-sync rule).
   - Positive/negative deterministic tests.

6. **`VOC-086-T05` — Docs + live proof**
   - Document adding monitoring for a page/API/feature, credential
     bootstrap/rotation, rollback, ownership, and alert/check proof.
   - Deploy/sync inventory through the normal workflow; confirm Kuma
     list/status via Socket.IO; manually dispatch synthetics and prove green;
     independently verify `monitor.vocanova.site`; confirm single-edge,
     isolation, and absence of 8081/8443.
   - Record redacted evidence in `t05-evidence.md`.

## Validation and independent verification

Deterministic (as applicable per task):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
# plus package-introduced monitoring schema / sync / governance tests, e.g.
# node --test scripts/foundation/voc086-*.test.mjs
# bash scripts/governance/validate-monitoring-impact.sh
git diff --check
```

Independent verifier (per `CLAUDE.md`) binds each task report to the exact
SHA, confirms the implementer did not approve/merge its own work, identifies
**A-004** as active, and re-checks after material remediation.

## Deployment and rollback

- Authorization: adoption + each task's independent verification; this package
  does not self-authorize production actions.
- Rollout: T00→T05 in order; live credential bootstrap and inventory apply occur
  only through the repository sync workflow after T02 exists; live closure
  proof is T05.
- Rollback trigger: partial apply without compensation, secret leakage,
  unrelated monitor overwrite, SQLite usage, topology regression, or false-green
  synthetics/governance.
- Rollback mechanism: revert repository inventory/workflow commits; re-run
  compensating supported-protocol sync; preserve manually owned monitors;
  rotate credentials again only if compromise is suspected (explicit input).
- Accountable owner: task evidence authors. Last-known-good: pre-VOC-086
  tree with two unmanaged production monitors and deploy-only synthetics per
  issue #716.
