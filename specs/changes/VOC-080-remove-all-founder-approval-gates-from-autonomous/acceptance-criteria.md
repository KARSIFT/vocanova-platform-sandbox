# VOC-080 — Acceptance Criteria

## VOC-080-AC-00 — End-to-end planning through task dispatch without founder comment

- Requirement source: issue #627 AC; `VOC-080-D00`, `VOC-080-D01`
- Tasks: `VOC-080-T02`, `VOC-080-T04`, `VOC-080-T06`, `VOC-080-T07`
- Tests: `VOC-080-TEST-00`, `VOC-080-TEST-06`
- Evidence: `VOC-080-EV-02`, `VOC-080-EV-06`, `VOC-080-EV-07`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome after activation: opening a valid change issue can lead
from planning through an adopted package and task dispatch without any
founder `approved` comment. Adoption records exact revision, review
evidence, resolved/deferred decisions, risk, and authority provenance.

## VOC-080-AC-01 — R4 plan and implementation PRs do not wait on founder approval

- Requirement source: issue #627 AC; `VOC-080-D00`
- Tasks: `VOC-080-T01`, `VOC-080-T04`, `VOC-080-T06`, `VOC-080-T07`
- Tests: `VOC-080-TEST-01`, `VOC-080-TEST-06`
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-06`, `VOC-080-EV-07`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome after activation: when CI, governance, scope, and
independent verification (plan_reviewer or review) pass, R4 plan and
implementation PRs merge without waiting for a founder `approved` comment.
R4 still requires stronger evidence/validation/rollback as specified by
risk policy — but not founder-comment gating.

## VOC-080-AC-02 — No approval override of failed or missing gates

- Requirement source: issue #627 AC; `VOC-080-D02`
- Tasks: `VOC-080-T01`, `VOC-080-T03`, `VOC-080-T05`
- Tests: `VOC-080-TEST-02`
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-05`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: no merge, release, or deploy path uses a founder
`approved` comment (or equivalent human click) to override a failed or
missing deterministic check or independent-verification verdict. Failed
gates stay fail-closed until remediation succeeds.

## VOC-080-AC-03 — Unparseable risk fails for correction

- Requirement source: issue #627 AC; `VOC-080-D02`
- Tasks: `VOC-080-T01`, `VOC-080-T05`
- Tests: `VOC-080-TEST-03`
- Evidence: `VOC-080-EV-01`, `VOC-080-EV-05`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: unparseable or inconsistent risk classification fails
closed with an actionable correction signal. It does not wait for founder
override and does not auto-merge.

## VOC-080-AC-04 — Merged plan packages are adopted or reconciled; no silent draft

- Requirement source: issue #627 AC; `VOC-080-D01`; VOC-040 gap
- Tasks: `VOC-080-T02`, `VOC-080-T06`
- Tests: `VOC-080-TEST-00`, `VOC-080-TEST-04`
- Evidence: `VOC-080-EV-02`, `VOC-080-EV-06`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: a merged plan package is atomically adopted as part of
the autonomous adoption path, or is reliably detected and reconciled. No
package can remain silently merged as `draft` / unauthorized after the
happy path or after running the reconciliation dispatch.

## VOC-080-AC-05 — Idempotent reconciliation / dispatch path exists

- Requirement source: issue #627 AC; VOC-040
- Tasks: `VOC-080-T02`, `VOC-080-T05`, `VOC-080-T06`
- Tests: `VOC-080-TEST-04`
- Evidence: `VOC-080-EV-02`, `VOC-080-EV-06`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: a documented, dispatchable, observable reconciliation
workflow recovers missing adoption events and missing task rosters without
replaying an old GitHub event and without depending solely on
`gh run rerun --failed` retention. Re-running is idempotent (no duplicate
task roster / duplicate issues).

## VOC-080-AC-06 — Release and production deploy require no founder interaction

- Requirement source: issue #627 AC; preserve 2026-08-08 auto-release/deploy
- Tasks: `VOC-080-T03`, `VOC-080-T04`, `VOC-080-T06`, `VOC-080-T07`
- Tests: `VOC-080-TEST-05`, `VOC-080-TEST-06`
- Evidence: `VOC-080-EV-03`, `VOC-080-EV-06`, `VOC-080-EV-07`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome after activation: develop→main promotion and
push-to-main production deployment require no founder `approved` comment
and no residual founder environment-reviewer requirement on the
repository-controlled path. Failed deployments remain fail-closed until
successful remediation checks.

## VOC-080-AC-07 — Non-founder controls remain enforced

- Requirement source: issue #627 “does not remove…”; `VOC-080-D00`
- Tasks: `VOC-080-T01`–`VOC-080-T05`, `VOC-080-T07`
- Tests: `VOC-080-TEST-02`, `VOC-080-TEST-07`
- Evidence: `VOC-080-EV-01`–`VOC-080-EV-05`, `VOC-080-EV-07`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: builder/verifier separation, protected-area checks,
secrets isolation, rollback requirements, monitoring expectations, and
audit evidence for automatic adoption, merge, release, deployment, and
rollback remain enforced. No agent may self-review its own exact revision.

## VOC-080-AC-08 — Tests cover R0–R4 and full loop surfaces

- Requirement source: issue #627 AC
- Tasks: `VOC-080-T05`
- Tests: `VOC-080-TEST-01`–`VOC-080-TEST-05`, `VOC-080-TEST-07`
- Evidence: `VOC-080-EV-05`
- Result: pass — `VOC-080-EV-07` (`t07-evidence.md`); live proof `t06-evidence.md`

Observable outcome: repository and/or reusable-workflow tests cover R0–R4
merge behavior, plan PRs, task PRs, remediation, recovery/reconcile,
release, and deployment fail-closed/retry semantics relevant to this
change.

## VOC-080-AC-09 — Canonical docs and repository settings agree with behavior

- Requirement source: issue #627 AC; AGENTS.md doc-reconciliation rule
- Tasks: `VOC-080-T00`, `VOC-080-T04`, `VOC-080-T07`
- Tests: `VOC-080-TEST-08`
- Evidence: `VOC-080-EV-00`, `VOC-080-EV-04`, `VOC-080-EV-07`
- Result: activation candidate — deterministic doc/settings checks pass; final result
  binds to canonical merge and post-merge validation of the exact T07 revision

Observable outcome: AGENTS.md, CLAUDE.md, DOC-15/DOC-16 as applicable,
A-004 (or settled amendment), governance matrices, templates, workflow
comments, and repository-settings documentation agree with implemented
no-founder-gate behavior. Historical evidence remains historical.
Documented settings match live repository/environment configuration for
paths this package claims to clear.

## VOC-080-AC-10 — VOC-079 / issue #624 can resume on the new path

- Requirement source: issue #627 AC
- Tasks: `VOC-080-T07`
- Tests: `VOC-080-TEST-06`
- Evidence: `VOC-080-EV-07`
- Result: pending canonical merge of the exact independently reviewed T07 revision

Observable outcome: after transition activation, VOC-079 (issue #624) and
related recovery work can progress through adopt/merge/release without a
founder `approved` comment, subject only to remaining non-founder gates.
This criterion does not require VOC-079 implementation to complete inside
VOC-080.
