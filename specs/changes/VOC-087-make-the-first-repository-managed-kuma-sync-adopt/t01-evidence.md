---
evidence_id: VOC-087-EV-01
task_id: VOC-087-T01
acceptance_criteria:
  - VOC-087-AC-02
  - VOC-087-AC-03
  - VOC-087-AC-07
  - VOC-087-AC-08
tests:
  - VOC-087-TEST-04
  - VOC-087-TEST-06
  - VOC-087-TEST-07
  - VOC-087-TEST-08
  - VOC-087-TEST-11
  - VOC-087-TEST-14
date: 2026-08-19
related_change: VOC-087
accountable_owner: unassigned
gate_status: repository-complete-live-inventory-apply-deferred
live_inventory_apply_claimed: false
---

# VOC-087-T01 — Preserve remote notification bindings on adopt and update

## Scope of this evidence

This task stops the Kuma synchronizer from sending an empty
`notificationIDList` that would clear remote alert bindings on adopt or
managed update. Canonical `monitors.yaml` remains preserve-by-default. It
does **not** claim live inventory apply, credential rotation recovery
(`VOC-087-T02`), or first live `sync-monitoring` dispatch (`VOC-086-T05`).

## Notification ownership (`VOC-087-D02` / `VOC-087-DEP-03`)

| Mode | Inventory shape | Edit payload behavior |
| --- | --- | --- |
| Preserve (default) | `notification_id_list` omitted | Copy remote `notificationIDList` into adopt/update payloads |
| Explicit ownership | `notification_id_list` present on the availability monitor entry | Send the inventory-owned object (including `{}` when explicitly set) |

Implementation: `resolveNotificationIDList` and
`inventoryOwnsNotificationBindings` in
`infra/monitoring/kuma-sync/monitor-payload.mjs`; planner passes the
remote monitor into `inventoryEntryToDesiredMonitor` for adopt/update paths
in `infra/monitoring/kuma-sync/plan.mjs`.

`monitorsMatch` does not compare `notificationIDList`, so preserved bindings
do not cause drift on a second sync.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Notification preserve/ownership helpers | `infra/monitoring/kuma-sync/monitor-payload.mjs` |
| Planner remote-binding merge | `infra/monitoring/kuma-sync/plan.mjs` |
| Deterministic protocol tests | `scripts/foundation/voc087-kuma-sync.test.mjs` |
| This evidence | `specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t01-evidence.md` |

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc087-kuma-sync.test.mjs
node --test scripts/foundation/voc086-kuma-sync.test.mjs
git diff --check
```

Tests use synthetic numeric notification IDs only (for example `{ "42": true }`).
No live Kuma password, notification destination, or session value appears in
this evidence.

## Live inventory apply status

**Not claimed.** First live `sync-monitoring` inventory apply remains
**deferred** until VOC-087 is merged and deployed (`VOC-087-D03`;
`VOC-086-T05`). This task did not dispatch `sync-monitoring` with
`sync_inventory=true` against live Kuma.
