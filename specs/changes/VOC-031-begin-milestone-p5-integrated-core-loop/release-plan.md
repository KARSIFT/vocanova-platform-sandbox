# VOC-031 — Release Plan

## Release and deployment authorization

No release or deployment is authorized by this draft. Merge into `develop`,
staging validation, production activation, and the DOC-12 P5 milestone gate
are distinct and none is granted here. Production and autonomous production
release remain disabled. `T01` additionally cannot be accepted until
`VOC-031-D03` is resolved (`VOC-031-DEP-05`); `T04`'s production enablement
additionally requires the founder go/no-go and DOC-05 §16 legal review noted
in `VOC-031-DEP-03`, separate from and in addition to this package's own
merge/staging authorization. This draft is not adopted.

## Preconditions, monitoring, and outcome

Before any staging candidate: an adopted package with its open decisions
resolved, the exact base/revision, required PR checks, the three-new-table
migration plan, non-production identities, exact-SHA independent review, and
a named rollback owner/procedure. Monitor: onboarding completion rate and
any onboarding-gate lockout signal (a completed learner unable to reach
`/home`); Settings write success/error rate; email-change request/confirm
rates and — critically — a mismatch-notification-delivery-failure signal
(the old-email security notification failing silently would be a defect);
account-deletion request rate, sweep completion latency against the 30-day
target, and any row stuck past `purge_after` without transitioning to
`completed`; and the new accessibility/performance CI gates' pass/fail trend
over time. Never include learner identity beyond aggregate counts. The
founder owns the open `VOC-031-D02`/`D03`/`D05`–`D07`/`D09` decisions and the
separate account-deletion production-enablement go/no-go; the future release
authority records the accountable technical/operational owner. Live staging
gate evidence is blocked until the F3 staging environment exists
(`VOC-031-DEP-02`).

## Rollback

Trigger on: a wrong-account or unauthorized account deletion or email
change; an account-deletion sweep leaving inconsistent state (deactivated
without purge progress, or purge without prior deactivation); an onboarding
gate that locks a learner out of the loop with no recovery path; a
regression in any A1–P4 screen; migration/schema failure; or a new-tooling
CI gate that was silently bypassed rather than honestly reported as failing.
Use the approved deploy/migration recovery procedure: preserve
`user_settings`/onboarding/session state, ensure no account-deletion request
is left in a state neither the pre-rollback nor post-rollback code can
resolve, validate with non-production identities, and preserve
incident/rollback evidence. The last-known-good revision is recorded at the
future release decision, not guessed here.

## Independent verification, human approvals, and closure

Claude Code must report the final SHA, evidence, limitations, findings, the
active A-003 authority, and the remaining R3/R4/adoption/activation gates —
including, explicitly, whether `VOC-031-D02`/`D03`/`D05`–`D07`/`D09` were
resolved as this draft proposed or overridden, and whether the
account-deletion production-enablement gate (`VOC-031-DEP-03`) remains
correctly unactivated. Routine R3 needs strengthened controls and
independent verification; the open decisions need founder confirmation
before their affected tasks proceed. Closure requires all P5 task evidence,
the reliability/accessibility/performance automation evidence, the
mock-decommission inventory, and the DOC-12 P5 gate evidence; neither a
package merge nor a staging deploy alone closes the milestone, live staging
gate evidence is blocked until F3 exists (`VOC-031-DEP-02`), and the P5 gate
additionally requires the DOC-12 §5 P5 wording itself — the full loop works
coherently in staging across supported layouts with no critical
product/security/data/accessibility/reliability defect — demonstrated with
evidence, not merely asserted.
