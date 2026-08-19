# VOC-087 — Test Plan

## VOC-087-TEST-00 — Exact live monitor records produce adopt/update, not collision or create

- Covers: `VOC-087-AC-00`
- Preconditions: `VOC-087-T00` tree; Socket.IO protocol mock; no live Kuma
- Procedure:
  1. Seed remote monitors with exactly:
     - `{name: "VocaNova Production API", url: "https://api-production.vocanova.site/healthz"}`
     - `{name: "VocaNova Production Web", url: "https://production.vocanova.site"}`
     plus one unrelated manual monitor at a distinct URL.
  2. Run plan/sync against canonical inventory.
  3. Assert operations for the two production IDs are adopt/update (edit of
     those existing IDs), not collision errors and not `add` creates for
     those URLs. Assert zero mutate calls against the manual monitor.
- Expected result: live identity is adoptable; manuals preserved.
- Evidence: `VOC-087-EV-00`

## VOC-087-TEST-01 — Trailing-slash-only web URL still adopts rather than duplicate-creating

- Covers: `VOC-087-AC-00`, `VOC-087-AC-01`
- Preconditions: `VOC-087-T00` tree; protocol mock
- Procedure:
  1. Seed the live web monitor URL as `https://production.vocanova.site/`
     (trailing slash) with live name `VocaNova Production Web`, and the live
     API record without slash.
  2. Run plan/sync.
  3. Assert web is adopted/updated, not collided, and not created.
- Expected result: shared URL normalizer treats trailing slash as equivalent.
- Evidence: `VOC-087-EV-00`

## VOC-087-TEST-02 — Distinct URLs after normalization still fail closed

- Covers: `VOC-087-AC-01`, `VOC-087-AC-07`
- Preconditions: `VOC-087-T00` tree
- Procedure:
  1. Seed an unmanaged monitor at a different host or path than the
     production web inventory URL (not a trailing-slash variant).
  2. Inventory does not adopt it.
  3. Assert unmanaged URL collision (or equivalent fail-closed error), no
     create onto that distinct URL, no secret in the error.
- Expected result: collision safety is not weakened to fuzzy URL matching.
- Evidence: `VOC-087-EV-00`

## VOC-087-TEST-03 — Name mismatch still fails closed

- Covers: `VOC-087-AC-00`, `VOC-087-AC-07`
- Preconditions: `VOC-087-T00` tree
- Procedure:
  1. Seed a remote monitor whose URL matches a production inventory URL
     after normalization but whose name is not the live/inventory name.
  2. Run plan/sync.
  3. Assert fail-closed (no silent adopt, no duplicate create of that URL).
- Expected result: exact name match remains required.
- Evidence: `VOC-087-EV-00`

## VOC-087-TEST-04 — Second sync is a no-op

- Covers: `VOC-087-AC-03`
- Preconditions: `VOC-087-T00` / `VOC-087-T01` tree; protocol mock
- Procedure:
  1. Adopt/update the two live production fixtures until they carry the
     ownership marker and desired managed metadata (T01: including preserved
     notification bindings).
  2. Re-run sync unchanged.
  3. Assert zero `add` / `editMonitor` / `deleteMonitor` calls and success
     exit.
- Expected result: idempotent success.
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`

## VOC-087-TEST-05 — Unrelated manual monitors are not mutated

- Covers: `VOC-087-AC-07`
- Preconditions: `VOC-087-T00` tree
- Procedure:
  1. Include a founder/manual monitor without the ownership marker and
     without an inventory adoption entry.
  2. Run sync.
  3. Assert no edit/delete against that monitor ID.
- Expected result: manuals preserved.
- Evidence: `VOC-087-EV-00`

## VOC-087-TEST-06 — Adopt/update payload retains remote notification bindings

- Covers: `VOC-087-AC-02`
- Preconditions: `VOC-087-T01` tree; protocol mock
- Procedure:
  1. Seed live production monitors with a synthetic `notificationIDList`
     (numeric IDs only, e.g. `{ "1": true }`).
  2. Run adopt/update.
  3. Assert each `editMonitor` payload includes that same binding object,
     not `{}`.
- Expected result: bindings survive adoption and managed update.
- Evidence: `VOC-087-EV-01`

## VOC-087-TEST-07 — Default desired monitor does not clear notifications

- Covers: `VOC-087-AC-02`
- Preconditions: `VOC-087-T01` tree
- Procedure:
  1. Build desired monitors from canonical inventory (no explicit
     notification ownership).
  2. Assert the value sent on edit is the remote list (TEST-06) and that
     `inventoryEntryToDesiredMonitor` does not default a clearing `{}`
     onto update payloads.
  3. For create of a *new* managed monitor with no remote counterpart,
     assert the path does not invent live notification destinations.
- Expected result: preserve-by-default; create does not clone live alerts.
- Evidence: `VOC-087-EV-01`

## VOC-087-TEST-08 — Explicit inventory notification ownership is honored

- Covers: `VOC-087-AC-02`
- Preconditions: `VOC-087-T01` tree; fixture inventory, not necessarily
  canonical `monitors.yaml`
- Procedure:
  1. Fixture an inventory entry that explicitly takes notification
     ownership (`VOC-087-DEP-03` shape).
  2. Seed a remote monitor with a different synthetic binding.
  3. Assert the edit payload uses the inventory-owned value, not the remote
     value and not an accidental wipe unless the inventory-owned value is
     empty *by explicit ownership*.
- Expected result: explicit ownership works; default remains preserve.
- Evidence: `VOC-087-EV-01`

## VOC-087-TEST-09 — Shell harnesses execute from the normal deterministic entry point

- Covers: `VOC-087-AC-04`
- Preconditions: `VOC-087-T02` tree; no live Kuma; no live secrets
- Procedure:
  1. From the same command CI uses (`pnpm test` or
     `node --test scripts/foundation/*.test.mjs`), prove both
     `kuma-rotate-credentials.selftest.sh` and
     `sync-kuma-inventory.selftest.sh` run and pass.
  2. Fail the Node/CI suite if either harness is missing or exits non-zero.
- Expected result: documented harnesses are executed, not only present.
- Evidence: `VOC-087-EV-02`

## VOC-087-TEST-10 — Reset-success / proof-transfer-failure is recoverable and fail-closed

- Covers: `VOC-087-AC-05`
- Preconditions: `VOC-087-T02` tree; workflow and/or script fixtures
- Procedure:
  1. Model a host reset success followed by failed proof/metadata fetch.
  2. Assert `reset-password.js` is not invoked again.
  3. Assert the last remaining generated-password copy is not scrubbed
     before a recoverable store path exists.
  4. Assert the path fails closed (non-zero) and fixture logs contain no
     password material.
- Expected result: rotation remains recoverable without credential exposure.
- Evidence: `VOC-087-EV-02`

## VOC-087-TEST-11 — No SQLite read/write path in sync/rotation tooling

- Covers: `VOC-087-AC-07`
- Preconditions: `VOC-087-T00` / `VOC-087-T01` / `VOC-087-T02` tree
- Procedure:
  1. Static or harness check that synchronizer, host sync script, rotation
     script, and `sync-monitoring.yml` do not reference `kuma.db`, `sqlite`,
     or direct `/app/data` DB mutation APIs for sync/deploy.
  2. Official `/app/extra/reset-password.js` invocation remains allowed for
     explicit rotation only.
- Expected result: Socket.IO plus official reset tool only.
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`, `VOC-087-EV-02`

## VOC-087-TEST-12 — Credential and notification redaction

- Covers: `VOC-087-AC-05`, `VOC-087-AC-07`
- Preconditions: `VOC-087-T02` tree; T01 fixtures as applicable
- Procedure:
  1. Fixture log streams containing password/token-like values.
  2. Assert redaction masks them.
  3. Assert evidence files and test fixtures do not embed live Kuma
     passwords or live notification destinations.
- Expected result: secrets never appear plaintext in logged fixtures.
- Evidence: `VOC-087-EV-02`

## VOC-087-TEST-13 — VOC-086 T01/T02 evidence no longer records stale live identity or obsolete rotation ordering

- Covers: `VOC-087-AC-06`
- Preconditions: `VOC-087-T00` (T01 evidence) and `VOC-087-T02` (T02 evidence)
- Procedure:
  1. Assert VOC-086 `t01-evidence.md` records the live names/URLs from
     `VOC-087-D00` (or explicitly states they superseded the old match
     keys) and does not claim `Production Web` / `Production API /healthz`
     are the current live adoption identity.
  2. Assert VOC-086 `t02-evidence.md` does not describe "store password
     before fetch" as the merged behavior; it describes proof-gated store
     plus the recovery boundary and CI execution of both harnesses.
- Expected result: evidence matches merged implementation and live
  preconditions, with no secrets.
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-02`

## VOC-087-TEST-14 — Package tasks do not dispatch live inventory apply

- Covers: `VOC-087-AC-08`
- Preconditions: each task tree
- Procedure:
  1. Review task changes and evidence for any `workflow_dispatch` of
     `sync-monitoring` with `sync_inventory=true` against live Kuma, or any
     claim of live inventory apply.
  2. Evidence must state live apply remains deferred until this package is
     merged and deployed.
- Expected result: no live apply from VOC-087 tasks.
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`, `VOC-087-EV-02`

Include positive, negative, authorization, failure, and rollback coverage as
above. Tests must not embed real secrets, production personal data, or live
notification destinations.
