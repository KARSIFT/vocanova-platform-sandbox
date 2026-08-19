---
evidence_id: VOC-087-EV-00
task_id: VOC-087-T00
acceptance_criteria:
  - VOC-087-AC-00
  - VOC-087-AC-01
  - VOC-087-AC-03
  - VOC-087-AC-06
  - VOC-087-AC-07
  - VOC-087-AC-08
tests:
  - VOC-087-TEST-00
  - VOC-087-TEST-01
  - VOC-087-TEST-02
  - VOC-087-TEST-03
  - VOC-087-TEST-04
  - VOC-087-TEST-05
  - VOC-087-TEST-13
  - VOC-087-TEST-14
date: 2026-08-19
related_change: VOC-087
accountable_owner: unassigned
gate_status: repository-complete-live-inventory-apply-deferred
live_inventory_apply_claimed: false
---

# VOC-087-T00 — Live adoption identity and shared URL normalization

## Scope of this evidence

This task aligns canonical production monitor identity to the 2026-08-19 live
Kuma records, adds one shared HTTP(S) trailing-slash URL comparator for
adoption matching and unmanaged URL collision, and extends deterministic
protocol tests. It does **not** claim live inventory apply, credential
rotation recovery (`VOC-087-T02`), notification-binding preservation
(`VOC-087-T01`), or first live `sync-monitoring` dispatch (`VOC-086-T05`).

## VOC-087-D00 live production identity

| Inventory id | name / match_name | url / match_url |
| --- | --- | --- |
| `kuma.availability.production.web` | `VocaNova Production Web` | `https://production.vocanova.site` |
| `kuma.availability.production.api-healthz` | `VocaNova Production API` | `https://api-production.vocanova.site/healthz` |

Canonical source: `infra/monitoring/monitors.yaml`.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Production adoption identity | `infra/monitoring/monitors.yaml` |
| Shared URL comparison | `infra/monitoring/kuma-sync/url-compare.mjs` |
| Planner adoption/collision wiring | `infra/monitoring/kuma-sync/plan.mjs` |
| Deterministic protocol tests | `scripts/foundation/voc087-kuma-sync.test.mjs` |
| Corrected VOC-086 T01 evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t01-evidence.md` |
| This evidence | `specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t00-evidence.md` |

## URL normalization (`VOC-087-D01`)

`monitorUrlsEqual` / `normalizeMonitorUrl` in `url-compare.mjs` are used by
both `adoptionMatches` and unmanaged URL collision in `plan.mjs`. Required
equivalence: HTTP(S) URLs that differ only by trailing slashes. Name matching
remains exact string equality. No host rewrite, scheme dropping, or fuzzy
name matching was observed in Kuma storage during this task.

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc087-kuma-sync.test.mjs
node --test scripts/foundation/voc086-kuma-sync.test.mjs
git diff --check
```

Secrets: none committed; no live Kuma password, notification destination, or
session value appears in this evidence.

## Live inventory apply status

**Not claimed.** First live `sync-monitoring` inventory apply remains
**deferred** until VOC-087 is merged and deployed (`VOC-087-D03`;
`VOC-086-T05`). This task did not dispatch `sync-monitoring` with
`sync_inventory=true` against live Kuma.
