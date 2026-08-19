# VOC-087 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.
No task may dispatch live `sync-monitoring` inventory apply.

## VOC-087-T00 — Live adoption identity and shared URL normalization

- Requirement source: issue #728; `VOC-087-D00`, `VOC-087-D01`,
  `VOC-087-D06`, `VOC-087-D07`
- Acceptance criteria: `VOC-087-AC-00`, `VOC-087-AC-01`, `VOC-087-AC-03`
  (identity half), `VOC-087-AC-06` (T01 evidence), `VOC-087-AC-07` (collision
  and manuals subset), `VOC-087-AC-08`
- Tests: `VOC-087-TEST-00`, `VOC-087-TEST-01`, `VOC-087-TEST-02`,
  `VOC-087-TEST-03`, `VOC-087-TEST-04` (identity half), `VOC-087-TEST-05`,
  `VOC-087-TEST-13` (T01 evidence), `VOC-087-TEST-14`
- Evidence: `VOC-087-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `infra/monitoring/monitors.yaml` production entries so `name`,
   `url`, `adoption.match_name`, and `adoption.match_url` equal
   `VOC-087-D00`:
   - `VocaNova Production API` /
     `https://api-production.vocanova.site/healthz`
   - `VocaNova Production Web` /
     `https://production.vocanova.site`
2. Change `plan.mjs` so adoption matching and unmanaged-URL collision use
   the same URL comparison, including HTTP(S) trailing-slash equivalence.
   Keep exact name equality. Keep fail-closed multiple-candidate and
   distinct-URL collisions.
3. Add deterministic protocol tests that seed the exact two live monitor
   records (those names/URLs), plus an unrelated manual monitor, and assert
   adopt/update rather than collision or create. Cover trailing-slash-only
   web URL, distinct-URL collision, name mismatch fail-closed, and a second
   sync no-op after adopt (ownership marker applied).
4. Correct
   `specs/changes/VOC-086-manage-monitoring-inventory/t01-evidence.md` so it
   no longer records `Production Web` / `Production API /healthz` as the
   live adoption identity.
5. Do not call live Kuma; do not rotate credentials; do not send empty
   notification bindings as a new requirement of this task (T01 owns that
   payload change).

### Explicitly out of scope for this task

- Notification payload preservation (T01).
- Rotation recovery and shell-harness CI execution (T02).
- Live inventory apply.

## VOC-087-T01 — Preserve remote notification bindings on adopt and update

- Requirement source: issue #728; `VOC-087-D02`, `VOC-087-DEP-03`
- Acceptance criteria: `VOC-087-AC-02`, `VOC-087-AC-03`, `VOC-087-AC-07`,
  `VOC-087-AC-08`
- Tests: `VOC-087-TEST-04` (notification half), `VOC-087-TEST-06`,
  `VOC-087-TEST-07`, `VOC-087-TEST-08`, `VOC-087-TEST-11`
- Evidence: `VOC-087-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-087-T00`

### Required work

1. Stop `inventoryEntryToDesiredMonitor` (and edit payload construction)
   from sending `notificationIDList: {}` as a default that clears remote
   bindings.
2. On adopt and later managed update, copy the remote monitor's existing
   `notificationIDList` into the Socket.IO edit payload unless inventory
   explicitly takes ownership.
3. Keep canonical `monitors.yaml` preserve-by-default. If an explicit
   ownership field is added (`VOC-087-DEP-03`), document it and test both
   default preserve and override.
4. Add update-payload tests with synthetic numeric notification IDs proving
   bindings survive adopt and update, that a second sync is a no-op, and
   that create does not invent live notification destinations.
5. Do not log notification configuration. Do not call live Kuma.

### Explicitly out of scope for this task

- Live identity/URL matcher (T00).
- Rotation recovery / CI harness execution (T02).
- Live inventory apply.
- Building a notification-provider catalog.

## VOC-087-T02 — Rotation recovery, CI harness execution, and T02 evidence

- Requirement source: issue #728; `VOC-087-D04`, `VOC-087-D05`,
  `VOC-087-D06`, `VOC-087-DEP-01`, `VOC-087-DEP-02`
- Acceptance criteria: `VOC-087-AC-04`, `VOC-087-AC-05`, `VOC-087-AC-06`
  (T02 evidence), `VOC-087-AC-07` (redaction/SQLite subset), `VOC-087-AC-08`
- Tests: `VOC-087-TEST-09`, `VOC-087-TEST-10`, `VOC-087-TEST-11`,
  `VOC-087-TEST-12`, `VOC-087-TEST-13` (T02 evidence), `VOC-087-TEST-14`
- Evidence: `VOC-087-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-087-T01`

### Required work

1. Close the reset-success/proof-transfer-failure hole in
   `.github/workflows/sync-monitoring.yml` and/or
   `infra/scripts/kuma-rotate-credentials.sh` per `VOC-087-D04`: fail closed,
   no second reset, no credential logs, last password copy not scrubbed
   before a recoverable store path exists.
2. Make `infra/scripts/kuma-rotate-credentials.selftest.sh` and
   `infra/scripts/sync-kuma-inventory.selftest.sh` execute from the normal
   deterministic test entry point (`pnpm test` /
   `node --test scripts/foundation/*.test.mjs`), not `existsSync` only.
   Extend the harnesses if needed to cover the recovery/scrub gating
   without live secrets.
3. Correct
   `specs/changes/VOC-086-manage-monitoring-inventory/t02-evidence.md` so it
   describes the merged independently reviewed proof-gated store, this
   recovery boundary, and that the harnesses actually run in CI. Remove the
   obsolete "store password before fetch" claim as a description of current
   merged behavior.
4. Do not rotate live Kuma credentials. Do not apply live inventory.

### Explicitly out of scope for this task

- Inventory name/URL identity (T00).
- Notification payload semantics (T01), except not regressing them.
- Live inventory apply.

## Task ordering notes

- T00 blocks safe first-apply matching.
- T01 blocks safe first-apply updates (notification wipe).
- T02 blocks claiming rotation/CI evidence completeness.
- Closing issue #728 is gated on AC results with evidence, merge, and
  deploy of this package — not on performing the first live apply.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
