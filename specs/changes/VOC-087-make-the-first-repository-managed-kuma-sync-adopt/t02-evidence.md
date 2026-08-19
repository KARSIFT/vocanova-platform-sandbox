---
evidence_id: VOC-087-EV-02
task_id: VOC-087-T02
acceptance_criteria:
  - VOC-087-AC-04
  - VOC-087-AC-05
  - VOC-087-AC-06
  - VOC-087-AC-07
  - VOC-087-AC-08
tests:
  - VOC-087-TEST-09
  - VOC-087-TEST-10
  - VOC-087-TEST-11
  - VOC-087-TEST-12
  - VOC-087-TEST-13
  - VOC-087-TEST-14
date: 2026-08-19
related_change: VOC-087
accountable_owner: unassigned
gate_status: repository-complete-live-inventory-apply-deferred
live_inventory_apply_claimed: false
remediation_of: fb50f1b9597af1f34fe01adaaf73516bd9e3d467
---

# VOC-087-T02 — Rotation recovery, CI harness execution, and T02 evidence

## Scope of this evidence

This task closes the reset-success / proof-transfer-failure recovery hole in
`sync-monitoring`, executes both Kuma shell harnesses from the normal
deterministic CI entry point, and corrects stale VOC-086 T02 evidence. It
does **not** claim live credential rotation, live inventory apply, or first
live `sync-monitoring` dispatch (`VOC-086-T05`).

## Rotation recovery (`VOC-087-D04` / `VOC-087-DEP-02`)

Decision helper: `infra/scripts/kuma-rotation-recovery-gate.sh` (no credentials).
The workflow store and scrub steps invoke the same helper that TEST-10 executes.

| Scenario | Store | Scrub |
| --- | --- | --- |
| Reset-applied proof matches this attempt | STORE from runner-local file | SCRUB only after both password and username secrets are stored |
| `rotate_host` succeeded, proof/metadata fetch failed | STORE password from runner-local file | RETAIN until both credential stores succeed |
| Host reachable and proof file absent (`rotate_host` not success) | SKIP_UNUSED (do not overwrite secret) | SCRUB unused copies |
| `rotate_host` failed after reset **and** proof fetch failed (host unreachable / no local proof) | RETAIN (fail closed, no store) | RETAIN last copies |
| `recover_store_only` | require password, proof, and username metadata bound to one attempt; never call `reset-password.js` | SCRUB only after both credential stores succeed |

Attempt 1 (`fb50f1b9597af1f34fe01adaaf73516bd9e3d467`) stored only on proof match
or `rotate_host == success`, then scrubbed when `store_password.outcome == success`
(including a soft skip) or `rotate_host != success`. That deleted the last copy
when reset applied, username extraction failed, and scp also failed.

This revision:

1. Tracks explicit password and username store completion (not step success).
2. Scrubs only after both secret writes complete, or when a successful host
   fetch proves this attempt did not reset Kuma.
3. Adds `recover_store_only` as a later store-only retry that requires a
   complete, internally consistent host bundle and never resets.
4. Refuses a new rotation whenever retained recovery material exists, and a
   failed preflight cannot enter cleanup and delete the bundle it found.
5. Invokes host rotation once; recover-store cannot be combined with rotate.

## CI harness execution (`VOC-087-D05`)

Both disposable harnesses execute from `pnpm test` via
`scripts/foundation/voc087-sync-monitoring-workflow.test.mjs` (`spawnSync` of
the `.selftest.sh` scripts). The rotate selftest also runs the recovery-gate
matrix, including the post-reset rotate-failure row. No live Kuma, live
secrets, or production data required.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Rotation recovery workflow | `.github/workflows/sync-monitoring.yml` |
| Recovery decision helper | `infra/scripts/kuma-rotation-recovery-gate.sh` |
| Rotation script (unchanged reset path) | `infra/scripts/kuma-rotate-credentials.sh` |
| Extended rotation harness | `infra/scripts/kuma-rotate-credentials.selftest.sh` |
| Inventory sync harness | `infra/scripts/sync-kuma-inventory.selftest.sh` |
| Deterministic tests | `scripts/foundation/voc087-sync-monitoring-workflow.test.mjs` |
| Corrected VOC-086 T02 evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t02-evidence.md` |
| This evidence | `specs/changes/VOC-087-make-the-first-repository-managed-kuma-sync-adopt/t02-evidence.md` |

## Deterministic validation

Commands (repo root):

```bash
bash infra/scripts/kuma-rotate-credentials.selftest.sh
bash infra/scripts/sync-kuma-inventory.selftest.sh
node --test scripts/foundation/voc087-sync-monitoring-workflow.test.mjs
node --test scripts/foundation/voc086-sync-monitoring-workflow.test.mjs
git diff --check
```

Secrets: none committed; no live Kuma password, notification destination, or
session value appears in this evidence.

## Live inventory apply status

**Not claimed.** First live `sync-monitoring` inventory apply remains
**deferred** until VOC-087 is merged and deployed (`VOC-087-D03`;
`VOC-086-T05`). This task did not dispatch `sync-monitoring` with
`sync_inventory=true` against live Kuma.
