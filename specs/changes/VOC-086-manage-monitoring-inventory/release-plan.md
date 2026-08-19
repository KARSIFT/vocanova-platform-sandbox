# VOC-086 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being merged as
a draft or even by being adopted alone. Adoption authorizes implementation
PRs only. Each task PR still requires independent verification against the
exact revision.

Proposed risk is **R4** (draft): measured path floor R4 for
`scripts/governance/*` when monitoring-impact validation is wired there, plus
semantic R3+ consequences for protected workflows, monitoring credentials, and
live monitor mutation. Issue #716 proposed R3; the draft raises to R4 from the
measured floor. Under **active A-004**, engineering-workflow gates (plan
adoption, merge, release promotion, repository-controlled deploy) require
**no** founder `approved` comment. R4 still requires strengthened evidence,
independent verification, monitoring, named rollback owner, and tested
recovery. `automatic_merge_allowed: true` is set per AGENTS.md
(`VOC-080-DEP-02`); setting true does not bypass path floors, CI, independent
verification, unparseable-risk fail-closed, or EHR.

Production/monitoring rollout uses the normal path: task merges to `develop`
→ package roster completion → develop→main promotion via `release.yml` /
`pipeline.yml` → repository deploy/sync workflows (including the new
monitoring sync workflow and existing deploy paths as needed). Interrupted
promotion retries via `reconcile-release`; failed gates remain fail-closed.

Issue #716 records a founder remediation directive for the governed loop;
that does not let this draft self-adopt or bypass independent verification.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R4 proposal accepted or amended in writing.
- Stance on `VOC-086-DEP-01`–`DEP-03` recorded at adoption or in task
  evidence as accepted implementer choices.
- T00–T05 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on each exact SHA.
- Kuma credentials bootstrapped via explicit rotation input and stored only
  in GitHub secrets before claiming live sync success.
- Live Kuma Socket.IO inventory/status, manual synthetic dispatch, and
  monitor-host reachability evidence recorded before claiming package
  closure against issue #716.

Monitoring after T00–T05:

- Kuma managed availability monitors (five intended) via Socket.IO status.
- Scheduled synthetics under stable IDs (manual + schedule).
- Hourly Sentry `error-monitoring` remains separate and healthy.
- Sync workflow success/failure (fail closed; no secret leakage).
- Isolation invariants (no 8081/8443; loopback-only Kuma 3001; single
  shared-edge; staging/production boundaries).

Outcome owner: named in `VOC-086-EV-05` (unassigned at drafting).
Success = `VOC-086-AC-00` through `VOC-086-AC-10` with linked evidence.

## Rollback

Trigger: uncompensated partial apply; secret leakage; unrelated monitor
overwrite; SQLite usage; unexpected password reset on normal sync;
topology/isolation/health breakage; false-green synthetics; governance
bypass for new route/endpoint packages.

Mechanism:

1. Revert the responsible task commit(s) (primary for repository state).
2. Re-run compensating supported-protocol sync from the rolled-back inventory
   (or last-known-good inventory) without touching unrelated manual monitors.
3. If credential compromise is suspected, perform an explicit
   `rotate_credentials` run (not a silent normal sync side effect).
4. Confirm gates and live monitor list match the rolled-back revision intent.

Validation: sync/synthetics/governance expectations match the reverted tree;
no secrets in git or workflow transcripts from the failed revision; isolation
intact; manually owned monitors preserved.

Accountable owner: T00–T05 evidence authors. Last-known-good: tree
immediately preceding the first merged VOC-086 task (two unmanaged production
monitors; deploy-only synthetics; no Kuma GitHub secrets) per issue #716.

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind each task report to the exact reviewed commit SHA.
- Confirm implementer did not approve/merge its own work.
- Identify active authority model **A-004** (`a004-active`).
- Confirm AC/test/evidence traceability; no SQLite sync path; credential
  bootstrap is explicit-only; manuals preserved; governance grandfathering
  works without history rewrite.
- Report remaining R4 evidence obligations; EHR not expected for this
  package.
- Confirm live Socket.IO inventory/status, synthetic dispatch greens, and
  monitor-host reachability before closure.

Do not conflate repository merge, release promotion, activation, or closure.
Closing issue #716 requires AC results with evidence, not task-issue closure
alone.
