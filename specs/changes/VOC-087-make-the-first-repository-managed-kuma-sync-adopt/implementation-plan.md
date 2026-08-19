# VOC-087 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status:
adopted`, `approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a founder
`approved` comment; exact-revision plan review and deterministic gates still
apply.

Additionally:

- Treat issue #728's 2026-08-19 live monitor identity as the starting point
  (`VOC-087-DEP-00`). Do not invent additional live monitors.
- Record rotation recovery shape (`VOC-087-DEP-02`) and notification
  ownership schema (`VOC-087-DEP-03`) in the relevant task evidence.
- Path floor **R3** for monitoring infra and `sync-monitoring.yml`. Proposed
  package risk **R3** (draft). Do not lower it. Verifier may raise it.
- Preserve VOC-081 monitoring topology and VOC-067 isolation invariants.
- Never commit secrets, mint tokens, session values, Kuma passwords, or
  live notification destinations into evidence files.
- Never introduce SQLite-based deploy/sync.
- Do not dispatch live `sync-monitoring` inventory apply from these tasks.

## File reconciliation and implementation sequence

Existing targets to preserve and correct, not replace:

- `infra/monitoring/monitors.yaml` — keep schema, five availability IDs,
  staging/monitor-host entries; correct the two production identity/adoption
  fields.
- `infra/monitoring/kuma-sync/plan.mjs` — keep fail-closed collision and
  duplicate-candidate errors; replace exact URL equality with the shared
  normalizer for adoption *and* unmanaged URL collision.
- `infra/monitoring/kuma-sync/monitor-payload.mjs` — keep editable field
  list and ownership description builder; stop defaulting
  `notificationIDList` to `{}` on desired monitors that do not own
  notifications.
- `scripts/foundation/voc086-kuma-sync.test.mjs` — keep existing
  create/update/collision/compensation coverage; add or replace adoption
  fixtures with the exact live records.
- `.github/workflows/sync-monitoring.yml` and
  `infra/scripts/kuma-rotate-credentials.sh` — keep official reset-password
  stdin path, proof marker, OpenSSH scp download, and environment-secret
  store; close the reset-success/proof-transfer-failure scrub hole.
- `infra/scripts/kuma-rotate-credentials.selftest.sh` and
  `infra/scripts/sync-kuma-inventory.selftest.sh` — keep disposable
  harnesses; execute them from CI.
- VOC-086 `t01-evidence.md` / `t02-evidence.md` — correct stale statements;
  do not rewrite unrelated VOC-086 history.

Ordered reversible steps:

1. **`VOC-087-T00` — Live adoption identity + shared URL normalization**
   - Set production inventory `name`/`url`/`match_name`/`match_url` to
     `VOC-087-D00`.
   - Implement one URL comparison for adoption and unmanaged-URL collision
     (trailing-slash equivalence required).
   - Add deterministic tests with the exact two live monitor records, a
     trailing-slash web variant, a distinct-URL collision, and a name
     mismatch fail-closed case.
   - Correct VOC-086 `t01-evidence.md`.
   - Do not mutate live Kuma. Do not change notification payloads here
     except as needed for compile/test wiring.

2. **`VOC-087-T01` — Preserve notification bindings**
   - Stop sending empty `notificationIDList` on adopt/update by default.
   - Copy remote bindings into edit payloads unless inventory explicitly
     owns them (`VOC-087-DEP-03`).
   - Prove adopt, update, and second-sync no-op retain synthetic
     notification IDs; prove explicit ownership override; prove create does
     not invent live bindings.
   - Do not dispatch live Kuma.

3. **`VOC-087-T02` — Rotation recovery + CI harness execution + T02 evidence**
   - Close reset-success/proof-transfer-failure per `VOC-087-D04`.
   - Execute both shell harnesses from `pnpm test` / foundation Node tests.
   - Correct VOC-086 `t02-evidence.md`.
   - Do not rotate live credentials and do not apply live inventory.

## Validation and independent verification

Deterministic (as applicable per task):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
pnpm test
# at minimum the monitoring-related subset, including:
# node --test scripts/foundation/voc086-kuma-sync.test.mjs
# node --test scripts/foundation/voc086-sync-monitoring-workflow.test.mjs
# node --test scripts/foundation/voc087-*.test.mjs   # if added
# plus actual execution of:
# bash infra/scripts/kuma-rotate-credentials.selftest.sh
# bash infra/scripts/sync-kuma-inventory.selftest.sh
git diff --check
```

Independent verifier (per `CLAUDE.md`) binds each task report to the exact
SHA, confirms the implementer did not approve/merge its own work, identifies
**A-004** as active, and re-checks after material remediation.

## Deployment and rollback

- Authorization: adoption + each task's independent verification; this
  package does not self-authorize production actions or the first live
  inventory apply.
- Rollout: T00→T01→T02 in order; live credential bootstrap/inventory apply
  occur only through `sync-monitoring` after this package is merged and
  deployed, as a later authorized operator/VOC-086-T05 step.
- Rollback trigger: duplicate create, collision false-negative, notification
  wipe, secret leakage, unrecoverable rotation, SQLite usage, or live apply
  from this package.
- Rollback mechanism: revert the responsible task commit(s); do not SQLite
  wipe Kuma; preserve manuals; do not blindly re-reset passwords.
- Accountable owner: task evidence authors. Last-known-good: merged VOC-086
  tree before these corrections, with live monitors still unmanaged under
  the 2026-08-19 names/URLs.
