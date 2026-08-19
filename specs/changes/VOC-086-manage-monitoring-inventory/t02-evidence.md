---
evidence_id: VOC-086-EV-02
task_id: VOC-086-T02
acceptance_criteria:
  - VOC-086-AC-03
  - VOC-086-AC-04
tests:
  - VOC-086-TEST-07
  - VOC-086-TEST-08
  - VOC-086-TEST-09
date: 2026-08-16
related_change: VOC-086
accountable_owner: unassigned
gate_status: repository-complete-live-bootstrap-pending
live_sync_claimed: false
remediation_of: 29238096f3f605a8fde5bd65a6103d660059a0b2
voc087_remediation: VOC-087-T02
---

# VOC-086-T02 — Sync/deploy workflow and credential bootstrap

## Scope of this evidence

This task delivers the repository `sync-monitoring` workflow, host-side
rotation/sync scripts, and deterministic wiring/redaction tests. It does **not**
claim live credential bootstrap, inventory apply, or AC-05 closure (`VOC-086-T05`).

## Remediation (post independent FAIL on `29238096…`)

| Finding                                          | Fix                                                                                                                                     |
| ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| High — inverted `appleboy/scp-action` "download" | Replaced with OpenSSH `scp` host→runner fetch of reset-applied proof and `/tmp/kuma-rotate-metadata.env`                                |
| Medium — username store soft-exit                | Fail-closed `exit 1` when metadata missing; rotate script also fails if username cannot be extracted                                    |
| Medium — password material left on paths         | Host + runner scrub steps remove rotation files when the recovery gate says SCRUB (store completed, or this attempt did not reset Kuma) |

## VOC-087 rotation recovery boundary (`VOC-087-D04`, corrected by VOC-087-T02)

Merged behavior is **proof-gated store**, not "store password before fetch":

1. Host rotation runs once via `kuma-rotate-credentials.sh` → official
   `reset-password.js` (stdin only). Proof is written after reset success and
   before username extraction, so a later username failure still leaves
   reset-applied proof on the host.
2. Runner downloads reset-applied proof and username metadata via OpenSSH `scp`,
   recording host reachability and whether the proof file was present.
3. `KUMA_PASSWORD` is stored with `gh secret set --body-file` when the
   reset-applied proof matches this attempt **or** host rotation succeeded but
   proof transfer failed (recovery from the runner-local file).
4. If reset may have been applied (`rotate_host` failed after proof write) and
   proof transfer also failed, store and scrub **RETAIN**: the job fails closed,
   does not store an unproven password over the existing secret, and does not
   delete the last remaining copies. `recover_store_only` later stores the
   host-retained file and never invokes `reset-password.js`.
5. Scrub after credential storage requires both `password_stored=true` and
   `username_stored=true`. Unused copies may be scrubbed only when the reserved
   remote probe explicitly confirms the reset marker is absent, or when rotate
   was skipped after a successful clean preflight. A soft skip of store is not
   treated as store success.
6. `reset-password.js` is never invoked a second time on a recovery retry.

Both shell harnesses (`kuma-rotate-credentials.selftest.sh`,
`sync-kuma-inventory.selftest.sh`) execute in CI via
`scripts/foundation/voc087-sync-monitoring-workflow.test.mjs` (`pnpm test`).

## Decision records for this task

| Decision                          | Recorded choice                                                                                                                                                                          |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `VOC-086-DEP-02` secret/env names | `KUMA_USERNAME`, `KUMA_PASSWORD` in GitHub `monitoring` environment; workflow reads `STAGING_SSH_*` repository secrets for host access                                                   |
| `rotate_credentials` input        | `workflow_dispatch.inputs.rotate_credentials` (boolean, default `false`)                                                                                                                 |
| Password reset path               | Host `infra/scripts/kuma-rotate-credentials.sh` → `docker exec -i vocanova-uptime-kuma node /app/extra/reset-password.js` with stdin (no argv password)                                  |
| Secret storage                    | `gh secret set KUMA_PASSWORD --env monitoring --body-file` (never echoed); username preserved via host metadata file, OpenSSH scp download to runner, then `gh secret set KUMA_USERNAME` |
| Host→runner transfer              | OpenSSH `scp` only for proof/metadata download; `appleboy/scp-action` remains upload-only (bundle + password file)                                                                       |
| Inventory apply path              | Host `infra/scripts/sync-kuma-inventory.sh` → disposable `node:24-bookworm-slim` on `vocanova-monitoring-net` running `sync-kuma.mjs`                                                    |
| Session invalidation              | Documented in workflow header and rotation script output (one-time per rotation)                                                                                                         |

## Repository deliverables

| Artifact                   | Path                                                                                                                         |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Sync workflow              | `.github/workflows/sync-monitoring.yml`                                                                                      |
| Credential rotation script | `infra/scripts/kuma-rotate-credentials.sh`                                                                                   |
| Inventory sync script      | `infra/scripts/sync-kuma-inventory.sh`                                                                                       |
| Disposable harnesses       | `infra/scripts/kuma-rotate-credentials.selftest.sh`, `infra/scripts/sync-kuma-inventory.selftest.sh`                         |
| Deterministic tests        | `scripts/foundation/voc086-sync-monitoring-workflow.test.mjs`, `scripts/foundation/voc087-sync-monitoring-workflow.test.mjs` |
| This evidence              | `specs/changes/VOC-086-manage-monitoring-inventory/t02-evidence.md`                                                          |

## Operator prerequisites (not in git)

1. Create GitHub environment `monitoring` on this repository.
2. Ensure repository `STAGING_SSH_*` secrets remain valid (shared host).
3. First run: `workflow_dispatch` with `rotate_credentials=true` and
   `sync_inventory=true` to bootstrap secrets and apply inventory.
4. Subsequent runs: `sync_inventory=true` only (`rotate_credentials` stays false).
5. If a rotate attempt reset Kuma but did not store the secret, dispatch
   `recover_store_only=true` (never combined with `rotate_credentials`) to store
   the host-retained password without a second reset.

## Validation commands (deterministic)

```bash
bash infra/scripts/kuma-rotate-credentials.selftest.sh
bash infra/scripts/sync-kuma-inventory.selftest.sh
node --test scripts/foundation/voc086-sync-monitoring-workflow.test.mjs
node --test scripts/foundation/voc087-sync-monitoring-workflow.test.mjs
node --test scripts/foundation/voc086-kuma-sync.test.mjs
```

## Live bootstrap / sync

Pending operator dispatch of `.github/workflows/sync-monitoring.yml` after merge
and deploy of VOC-087. No live Kuma mutation is claimed in this evidence revision.
First live inventory apply remains deferred (`VOC-086-T05` / `VOC-087-D03`).
