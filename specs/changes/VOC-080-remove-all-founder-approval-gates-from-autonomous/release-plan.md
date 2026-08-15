# VOC-080 — Release Plan

## Release and deployment authorization

This adopted package authorizes implementation PRs; it does **not** authorize
production application deployment merely by adoption.

Risk is **R4**: governance authority replacement for autonomous merge,
adoption, R4 progression, release, and deploy control.
`automatic_merge_allowed: false` remains on this transition package while
A-003 is active. T02/T07 implement and activate the adopted decision that the
post-transition R4 path uses deterministic and independent-verification gates
without a later founder-comment gate.

Founder adoption was recorded on PR #628 under pre-transition authority. The
one final founder approval must bind the exact T07 activation revision. After
activation, no later founder approval comment is required. T07 also remains
fail-closed on deterministic validation, exact-revision independent
verification, rehearsal evidence, and rollback readiness.

Existing 2026-08-08 auto-promote / push-to-main deploy remains the
production path; T03 removes residual founder-comment / environment-
reviewer gates on that path.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted; `VOC-080-DEP-00`–`DEP-04` resolved or explicitly
  deferred.
- T00–T05 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on exact SHAs.
- T06 rehearsal evidence recorded on the settled venue.
- T07 exact-revision founder transition approval and independent verification
  pass before activation markers flip.

Monitoring after activation:

- Plan → adopt → task dispatch completes without founder comments.
- R4 PRs auto-merge only when CI + independent verification pass.
- Unparseable risk fails closed (no merge).
- Reconcile dispatch repairs a deliberate merged-but-unadopted fixture
  idempotently.
- Release promotion and production deploy proceed without founder
  interaction; failed deploys stay fail-closed.
- Docs and live settings agree (spot-check pipeline inputs and
  environment reviewers).

Outcome owner: named in `VOC-080-EV-07` (unassigned at drafting).
Success = `VOC-080-AC-00` through `VOC-080-AC-10` with linked evidence.

## Rollback

Trigger: verification bypass; silent draft merges; authority markers
wrong; doc/settings disagreement; unsafe production progression.

Mechanism: restore known-good caller and infra workflow revisions and
pre-transition authority markers as described in
`implementation-plan.md`; do not rewrite audit history.

Validation: governance scripts pass; merge-gate again matches the
restored model; adopt reconcile behavior matches restored revision;
docs describe the restored model.

Accountable owner: T07 evidence. Last-known-good: pre-VOC-080-T01 infra
SHAs + pre-T04 caller tip + A-003-active transition-state.

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind the exact reviewed commit SHA for each task.
- Confirm the change matches this specification and acceptance criteria.
- Run/inspect applicable deterministic checks; never treat missing
  rehearsal or infra access as a pass for T06/T07.
- Escalate if semantic risk exceeds the declared class.
- Verify the implementer-role occupant did not approve or merge its own
  implementation.
- Identify active authority model (**a003-active** until T07 activation;
  successor thereafter) and report every still-required R4 / EHR /
  adoption / activation gate.
- On T07: confirm the founder transition approval and independent verdict are
  bound to the exact activation revision, and that post-activation no workflow waits on
  founder `approved` while non-founder controls remain.

Closure requires acceptance-criteria results with evidence, not merely
merged PRs. Repository merge, release to `main`, production deploy,
transition activation, and package closure are distinct events.

Issue #627 closes only when AC-00–AC-10 are satisfied with evidence.
VOC-075 / A-003 remain historical; cite supersession by #627.
