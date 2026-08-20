# VOC-097 — Release Plan

## Release and deployment authorization

This package does not by itself authorize application production deployment.
Governance-automation changes take effect when karsift-ai-infra reusable workflows
merge and this repository consumes them (pin bump or `@main` per open question 4),
and when calling-repo doc/wiring task PRs merge to `develop` and promote through the
normal release path. Under active A-004, no founder `approved` comment is a
merge/adopt/release gate.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00→T01→T02
  land before T03/T04; T05 after both; CI, governance validation, and independent
  verification pass on each task PR (and on infra PRs under that repo's process).
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new or updated Kuma monitors/synthetics. Outcome signals are
  governance lifecycle boolean/state/count outputs and sanitized run identifiers
  only. Operational-failure observer and Sentry remain separate and must stay
  healthy (`VOC-097-AC-10`).
- **Outcome owner:** unassigned (set at adoption).
- **Stranded tasks:** #779 and #785 closed or migrated per `VOC-097-AC-08` before
  package closure.

## Rollback

- **Trigger:** Waiting incorrectly suppresses remediation for real defects; reconciler
  accepts non-qualifying evidence; sensitive metadata appears in comments; timeout
  fails to escalate; caller never receives infra fix.
- **Mechanism:** Revert T01/T02 infra commits and any calling-repo `pipeline.yml`
  wiring; re-promote via normal develop→main path if already released. Preserve
  evidence files as audit history. Do not "fix" by granting implementer Actions
  credentials.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm remediate again retries genuine FAIL;
  confirm no live-evidence wake path remains active for this caller if wiring
  reverted.
- **Last-known-good:** last known good infra release + calling-repo `develop`/`main`
  SHAs before T01/T02.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification. Post-reconcile heads
  for live-evidence tasks require a **new** exact-SHA review (`VOC-097-D04`).
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic matrix green, sanitization
  verified, least-privilege verified, stranded tasks reconciled, live proof and
  observer separation recorded.
- Closure: all AC results with evidence in `t00-evidence.md` … `t05-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.

Historically under A-003, R4 merge required founder approval. **Under active
A-004,** engineering-workflow gates require no founder `approved` comment;
preserve triggered EHR evidence if product ambiguity arises separately from merge
gates.
