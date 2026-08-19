---
evidence_id: VOC-086-EV-01
task_id: VOC-086-T01
acceptance_criteria:
  - VOC-086-AC-01
  - VOC-086-AC-02
  - VOC-086-AC-03
tests:
  - VOC-086-TEST-02
  - VOC-086-TEST-03
  - VOC-086-TEST-04
  - VOC-086-TEST-05
  - VOC-086-TEST-06
  - VOC-086-TEST-07
  - VOC-086-TEST-09 (T01 redaction subset; full bootstrap print coverage remains T02)
date: 2026-08-15
related_change: VOC-086
accountable_owner: unassigned
gate_status: repository-complete-live-sync-pending
live_sync_claimed: false
remediation_of: ad01442e36495970e4447d3f3d3d3b9d633836a3
---

# VOC-086-T01 — Idempotent Kuma Socket.IO synchronizer

## Scope of this evidence

This task delivers the repository Socket.IO synchronizer, protocol mocks, and
deterministic tests. It does **not** claim live credentialed sync against
production Kuma, workflow bootstrap (`VOC-086-T02`), scheduled synthetics
(`VOC-086-T03`), governance CI (`VOC-086-T04`), or live AC-05/AC-10 closure
(`VOC-086-T05`).

## Decision records for this task

| Decision | Recorded choice |
| --- | --- |
| `VOC-086-DEP-01` synchronizer runtime | Node.js ESM under `infra/monitoring/kuma-sync/` + CLI `infra/monitoring/sync-kuma.mjs`; `socket.io-client` dependency |
| Adoption match keys (existing production monitors) | Both `adoption.match_name` **and** `adoption.match_url` must match (`plan.mjs`) using exact name equality and shared HTTP(S) trailing-slash URL comparison (`infra/monitoring/kuma-sync/url-compare.mjs`). **2026-08-19 live identity (VOC-087-D00):** `VocaNova Production Web` / `https://production.vocanova.site` and `VocaNova Production API` / `https://api-production.vocanova.site/healthz`. Supersedes the pre-VOC-087 inventory labels `Production Web` / `Production API /healthz`. |
| Ownership marker | `vocanova:repo-managed` embedded in managed monitor description metadata |
| Secret env names (`VOC-086-DEP-02`) | CLI reads `KUMA_URL` (default `http://127.0.0.1:3001`), `KUMA_USERNAME`, `KUMA_PASSWORD` — storage/bootstrap remains T02 |

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Sync CLI | `infra/monitoring/sync-kuma.mjs` |
| Socket.IO client (login + disconnect-on-failure) | `infra/monitoring/kuma-sync/socket-client.mjs` |
| Auth handshake (monitorList armed before login) | `infra/monitoring/kuma-sync/auth-handshake.mjs` |
| Plan / apply / sync / redact / mock | `infra/monitoring/kuma-sync/*.mjs` |
| Deterministic protocol tests | `scripts/foundation/voc086-kuma-sync.test.mjs` |
| This evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t01-evidence.md` |

## Remediation notes (attempt 2)

Independent verification of `ad01442e36495970e4447d3f3d3d3b9d633836a3` failed on:

1. **High** — `monitorList` listener was armed only after the login ack; Uptime Kuma often pushes `monitorList` on successful auth, so the live client could time out after a successful login. Fixed in `auth-handshake.mjs` (`createMonitorListWaiter` / `authenticateAndLoadMonitors`): listener is registered **before** `login` emit; tests cover the push-before-ack race.
2. **Medium** — Missing `t01-evidence.md` (this file).
3. **Medium** — Connect/login/`monitorList` failures now `socket.disconnect()` inside `createSocketKumaClient` before rethrow; `formatSyncFailure` always runs `redactSecrets` for generic `Error`s (CLI path).
4. **Low** — Redaction coverage renamed from a duplicate `VOC-086-TEST-07` title to `VOC-086-TEST-09 (T01 subset)` so TEST-07 remains the SQLite static ban.

## Acceptance mapping

| AC | Repository outcome |
| --- | --- |
| AC-01 | Create / update / adopt (name+URL) / preserve manuals / idempotent no-op via plan+apply; TEST-02–04 |
| AC-02 | Full inventory validation before mutation; create/update compensation; incomplete rollback → `SyncApplyError`; redacted failure formatting; TEST-05–06 + redaction subset |
| AC-03 | No `kuma.db` / `sqlite` / `/app/data` references in sync sources; TEST-07 |

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc086-kuma-sync.test.mjs
git diff --check
```

Secrets: none committed; no live Kuma password, mint token, or session value appears in this evidence.

## Live sync status

Not claimed. Live inventory apply requires T02 credential bootstrap/workflow and T05 proof.
