# VOC-087 — Acceptance Criteria

## VOC-087-AC-00 — Live production monitors are adopted or updated, not collided or duplicate-created

- Requirement source: issue #728; `VOC-087-D00`, `VOC-087-D01`
- Tasks: `VOC-087-T00`
- Tests: `VOC-087-TEST-00`, `VOC-087-TEST-01`, `VOC-087-TEST-03`
- Evidence: `VOC-087-EV-00`
- Result: pending

Observable outcome:

1. Canonical inventory `name`, `url`, `adoption.match_name`, and
   `adoption.match_url` for the two production availability monitors equal
   the 2026-08-19 live identity in `VOC-087-D00`.
2. Planning/sync against a remote fixture of exactly those two live records
   (plus an unrelated manual monitor) produces adopt/update operations for
   the two production IDs, zero unmanaged URL collisions, and zero creates
   of those two URLs.
3. Exact live names remain required; a name mismatch still fails closed.

## VOC-087-AC-01 — Shared URL normalization does not weaken collision safety

- Requirement source: issue #728; `VOC-087-D01`
- Tasks: `VOC-087-T00`
- Tests: `VOC-087-TEST-01`, `VOC-087-TEST-02`
- Evidence: `VOC-087-EV-00`
- Result: pending

Observable outcome:

1. Adoption matching and unmanaged-URL collision use the same URL comparison,
   including HTTP(S) trailing-slash equivalence.
2. A trailing-slash-only difference on the live production web URL produces
   adoption/update, not collision and not duplicate create.
3. Distinct URLs after that normalization still fail closed as collisions
   when unmanaged and not explicitly adopted. Name matching stays exact.

## VOC-087-AC-02 — Remote notification bindings survive adopt and update unless inventory owns them

- Requirement source: issue #728; `VOC-087-D02`
- Tasks: `VOC-087-T01`
- Tests: `VOC-087-TEST-06`, `VOC-087-TEST-07`, `VOC-087-TEST-08`
- Evidence: `VOC-087-EV-01`
- Result: pending

Observable outcome:

1. Adopt and later managed update payloads retain the remote monitor's
   existing `notificationIDList` when inventory does not take ownership.
2. Desired-monitor construction does not send `notificationIDList: {}` as a
   default that would clear bindings.
3. When inventory explicitly takes ownership, the inventory value is sent.
4. Tests use synthetic numeric notification IDs only; no live notification
   configuration appears in logs or evidence.

## VOC-087-AC-03 — Second sync against an already-applied inventory is a no-op

- Requirement source: issue #728; `VOC-087-D07`
- Tasks: `VOC-087-T00`, `VOC-087-T01`
- Tests: `VOC-087-TEST-04`
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`
- Result: pending

Observable outcome:

1. After a successful adopt/update of the two live production monitors
   (including preserved notification bindings and ownership marker), a
   second sync emits no mutating Socket.IO calls and exits successfully.

## VOC-087-AC-04 — Shell harnesses and Node tests run in deterministic CI

- Requirement source: issue #728; `VOC-087-D05`
- Tasks: `VOC-087-T02`
- Tests: `VOC-087-TEST-09`
- Evidence: `VOC-087-EV-02`
- Result: pending

Observable outcome:

1. `infra/scripts/kuma-rotate-credentials.selftest.sh` and
   `infra/scripts/sync-kuma-inventory.selftest.sh` execute (not merely exist)
   from the normal deterministic entry point used by CI (`pnpm test` and/or
   `node --test scripts/foundation/*.test.mjs`).
2. Existing Node foundation tests for kuma-sync and sync-monitoring continue
   to run on that same entry point.
3. Harnesses do not require live Kuma, live secrets, or production data.

## VOC-087-AC-05 — Reset-success / proof-transfer-failure remains recoverable without exposing credentials

- Requirement source: issue #728; `VOC-087-D04`
- Tasks: `VOC-087-T02`
- Tests: `VOC-087-TEST-10`, `VOC-087-TEST-12`
- Evidence: `VOC-087-EV-02`
- Result: pending

Observable outcome:

1. If the host reset for this attempt succeeded and proof/metadata transfer
   then fails, the workflow fails closed and does not invoke
   `reset-password.js` a second time.
2. The generated password remains recoverable until stored in GitHub
   `monitoring` `KUMA_PASSWORD`, or an explicit recover-store path exists
   that never resets again.
3. Scrub of the last remaining password copy is gated on successful secret
   store or on proof that this attempt did not reset Kuma.
4. No credential or notification configuration appears in logs, fixtures, or
   evidence.

## VOC-087-AC-06 — Stale VOC-086 T01/T02 evidence describes merged implementation and live preconditions

- Requirement source: issue #728; `VOC-087-D06`
- Tasks: `VOC-087-T00`, `VOC-087-T02`
- Tests: `VOC-087-TEST-13`
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-02`
- Result: pending

Observable outcome:

1. VOC-086 `t01-evidence.md` no longer presents `Production Web` /
   `Production API /healthz` as the live adoption identity. It records the
   2026-08-19 live names/URLs and the shared URL normalizer.
2. VOC-086 `t02-evidence.md` no longer describes storing the password
   *before* proof fetch as the merged ordering. It describes the independently
   reviewed proof-gated store plus this package's recovery boundary, and
   that both shell harnesses execute in CI.
3. Evidence contains no secrets or notification destinations.

## VOC-087-AC-07 — Fail closed, no SQLite, manuals preserved, secrets redacted

- Requirement source: issue #728 constraints; `VOC-087-D07`
- Tasks: `VOC-087-T00`, `VOC-087-T01`, `VOC-087-T02`
- Tests: `VOC-087-TEST-03`, `VOC-087-TEST-05`, `VOC-087-TEST-11`,
  `VOC-087-TEST-12`
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`, `VOC-087-EV-02`
- Result: pending

Observable outcome:

1. Synchronizer and sync/rotation scripts do not open, copy, or mutate
   `kuma.db` / SQLite.
2. Unrelated manually owned monitors are not mutated.
3. Ambiguous adoption (multiple candidates) and distinct unmanaged URL
   collisions still fail closed with explicit errors and no secrets.

## VOC-087-AC-08 — First live inventory apply stays deferred until this package is merged and deployed

- Requirement source: issue #728; `VOC-087-D03`
- Tasks: `VOC-087-T00`, `VOC-087-T01`, `VOC-087-T02`
- Tests: `VOC-087-TEST-14`
- Evidence: `VOC-087-EV-00`, `VOC-087-EV-01`, `VOC-087-EV-02`
- Result: pending

Observable outcome:

1. No task in this package dispatches `sync-monitoring` with
   `sync_inventory=true` against live Kuma or claims live inventory apply.
2. Task evidence records that VOC-086-T05 / first live apply remains blocked
   until this package is merged and deployed.
