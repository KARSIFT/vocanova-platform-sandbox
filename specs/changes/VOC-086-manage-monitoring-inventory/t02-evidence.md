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
---

# VOC-086-T02 — Sync/deploy workflow and credential bootstrap

## Scope of this evidence

This task delivers the repository `sync-monitoring` workflow, host-side
rotation/sync scripts, and deterministic wiring/redaction tests. It does **not**
claim live credential bootstrap, inventory apply, or AC-05 closure (`VOC-086-T05`).

## Remediation (post independent FAIL on `29238096…`)

| Finding | Fix |
| --- | --- |
| High — inverted `appleboy/scp-action` "download" | Replaced with OpenSSH `scp` host→runner fetch of `/tmp/kuma-rotate-metadata.env` (`Fetch Kuma username metadata from host`); `KUMA_PASSWORD` is stored from the runner-local file *before* that fetch so a transfer failure cannot strand a reset Kuma password outside GitHub secrets |
| Medium — username store soft-exit | Fail-closed `exit 1` when metadata missing; rotate script also fails if username cannot be extracted |
| Medium — password material left on paths | Store step deletes workspace `kuma-new-password.secret`; `always()` host + runner scrub steps remove rotation files even when `sync_inventory=false` |

## Decision records for this task

| Decision | Recorded choice |
| --- | --- |
| `VOC-086-DEP-02` secret/env names | `KUMA_USERNAME`, `KUMA_PASSWORD` in GitHub `monitoring` environment; workflow reads `STAGING_SSH_*` repository secrets for host access |
| `rotate_credentials` input | `workflow_dispatch.inputs.rotate_credentials` (boolean, default `false`) |
| Password reset path | Host `infra/scripts/kuma-rotate-credentials.sh` → `docker exec -i vocanova-uptime-kuma node /app/extra/reset-password.js` with stdin (no argv password) |
| Secret storage | `gh secret set KUMA_PASSWORD --env monitoring --body-file` (never echoed); username preserved via host metadata file, OpenSSH scp download to runner, then `gh secret set KUMA_USERNAME` |
| Host→runner transfer | OpenSSH `scp` only for metadata download; `appleboy/scp-action` remains upload-only (bundle + password file) |
| Inventory apply path | Host `infra/scripts/sync-kuma-inventory.sh` → disposable `node:24-bookworm-slim` on `vocanova-monitoring-net` running `sync-kuma.mjs` |
| Session invalidation | Documented in workflow header and rotation script output (one-time per rotation) |

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Sync workflow | `.github/workflows/sync-monitoring.yml` |
| Credential rotation script | `infra/scripts/kuma-rotate-credentials.sh` |
| Inventory sync script | `infra/scripts/sync-kuma-inventory.sh` |
| Disposable harnesses | `infra/scripts/kuma-rotate-credentials.selftest.sh`, `infra/scripts/sync-kuma-inventory.selftest.sh` |
| Deterministic tests | `scripts/foundation/voc086-sync-monitoring-workflow.test.mjs` |
| This evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t02-evidence.md` |

## Operator prerequisites (not in git)

1. Create GitHub environment `monitoring` on this repository.
2. Ensure repository `STAGING_SSH_*` secrets remain valid (shared host).
3. First run: `workflow_dispatch` with `rotate_credentials=true` and
   `sync_inventory=true` to bootstrap secrets and apply inventory.
4. Subsequent runs: `sync_inventory=true` only (`rotate_credentials` stays false).

## Validation commands (deterministic)

```bash
bash infra/scripts/kuma-rotate-credentials.selftest.sh
bash infra/scripts/sync-kuma-inventory.selftest.sh
node --test scripts/foundation/voc086-sync-monitoring-workflow.test.mjs
node --test scripts/foundation/voc086-kuma-sync.test.mjs
```

## Live bootstrap / sync

Pending operator dispatch of `.github/workflows/sync-monitoring.yml` after merge.
No live Kuma mutation is claimed in this evidence revision.
