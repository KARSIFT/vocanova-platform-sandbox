# VOC-087 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged as
a draft or even by being adopted alone. Adoption authorizes implementation
PRs only. Each task PR still requires independent verification against the
exact revision.

This package also does **not** authorize the first live `sync-monitoring`
inventory apply. That apply stays deferred until this package is merged and
deployed; VOC-086-T05 / operator dispatch remains the live-proof path after
that.

Proposed risk is **R3** (draft): measured path floor R3 for
`infra/monitoring`, `infra/scripts`, and
`.github/workflows/sync-monitoring.yml`. Issue #728 did not declare a class.
Semantic first-live-apply and credential-recovery consequences may still be
raised to R4 by the independent verifier or reviewing human. This proposal
is not a determination.

Under **active A-004**, engineering-workflow gates (plan adoption, merge,
release promotion, repository-controlled deploy) require **no** founder
`approved` comment. `automatic_merge_allowed: true` is set per AGENTS.md
(`VOC-080-DEP-02`); setting true does not bypass path floors, CI,
independent verification, unparseable-risk fail-closed, or EHR.

Production/monitoring rollout uses the normal path: task merges to `develop`
→ package roster completion → develop→main promotion via `release.yml` /
`pipeline.yml` → repository deploy. Interrupted promotion retries via
`reconcile-release`; failed gates remain fail-closed.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R3 proposal accepted or raised in writing.
- Stance on `VOC-087-DEP-02` and `VOC-087-DEP-03` recorded at adoption or in
  task evidence as accepted implementer choices.
- T00–T02 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on each exact SHA.
- Deterministic Node tests and both shell harnesses green in CI.
- No live inventory apply claimed by this package.

Monitoring after T00–T02 merge/deploy, before first live apply:

- Do not treat current live Kuma names/URLs as already managed.
- Sync workflow must remain fail-closed; watch for secret leakage.
- Hourly Sentry `error-monitoring` remains separate.
- Isolation invariants (no 8081/8443; loopback-only Kuma 3001; single
  shared-edge) remain intact.

Outcome owner: named in `VOC-087-EV-02` (unassigned at drafting).
Success for this package = `VOC-087-AC-00` through `VOC-087-AC-08` with
linked evidence. Success for the later first live apply is **not** this
package's closure condition.

## Rollback

Trigger: duplicate create or failed adoption against live-shaped fixtures;
notification wipe in payload tests; unrecoverable rotation; secret leakage;
SQLite usage; live apply from this package; topology/isolation breakage.

Mechanism:

1. Revert the responsible task commit(s) (primary for repository state).
2. Do not SQLite-edit Kuma to "undo" a later live apply; if a later apply
   already happened, compensate via supported Socket.IO from the reverted
   inventory without touching unrelated manuals.
3. If credential compromise is suspected, perform an explicit
   `rotate_credentials` run. Proof-transfer failure recovery must store the
   already-generated password, not reset again.
4. Confirm gates match the rolled-back revision intent.

Validation: tests and evidence match the reverted tree; no secrets in git
or workflow transcripts; isolation intact; manually owned monitors
preserved.

Accountable owner: T00–T02 evidence authors. Last-known-good: tree
immediately preceding the first merged VOC-087 task (VOC-086 synchronizer
with unsafe live identity/notification defaults; live monitors still the
2026-08-19 unmanaged pair).

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind each task report to the exact reviewed commit SHA.
- Confirm implementer did not approve/merge its own work.
- Identify active authority model **A-004** (`a004-active`).
- Confirm AC/test/evidence traceability; live-identity tests; notification
  preservation; harness execution; rotation recovery; no SQLite; no live
  apply from this package; VOC-086 T01/T02 evidence corrections.
- Report remaining R3 evidence obligations and whether semantic escalation
  to R4 was applied; EHR not expected.
- Confirm this package did not dispatch live inventory apply.

Do not conflate repository merge, release promotion, activation, or closure.
Closing issue #728 requires AC results with evidence **and** that the first
live apply remains deferred until after merge and deploy — not task-issue
closure alone, and not a live apply performed by this package.
