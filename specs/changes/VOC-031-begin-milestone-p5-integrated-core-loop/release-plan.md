# VOC-031 — Release Plan

## Release and deployment authorization

No release or deployment is authorized by this draft. Merge into `develop`,
staging validation, production activation, and the DOC-12 P5 milestone gate
are distinct and none is granted here. Production and autonomous production
release remain disabled. `T00` cannot be accepted until `D07` (onboarding→
`user_settings` seeding) is resolved; `T02` additionally depends on `D06`
(settings/account write-field boundary); `T08` depends on `D08`'s "do not
rename" default holding. This draft is not adopted.

## Preconditions, monitoring, and outcome

Before any staging candidate: an adopted package with `D06`–`D08` resolved,
the exact base/revision, required PR checks (including the two new
accessibility/performance CI jobs, confirmed blocking), the one-new-table
migration plan, non-production identities, exact-SHA independent review,
and a named rollback owner/procedure. Monitor onboarding-completion rate,
settings/account update success and error rate, the two new write
endpoints' CSRF/idempotency-rejection rate (a nonzero baseline is expected
and healthy; a spike may indicate a client-side retry bug from `T05`),
Lighthouse CI threshold trend per route, and the automated accessibility
sweep's pass/fail trend per route — never including learner identity beyond
the aggregate count. The founder owns the open `D06`–`D08` product/scope
decisions; the future release authority records the accountable
technical/operational owner. Live staging gate evidence is blocked until
the F3 staging environment exists (`VOC-031-DEP-02`) — and, distinctly from
every prior milestone, the DOC-12 §5 P5 gate's own wording is centrally
about staging behavior, so this is this milestone's largest completion risk
(`VOC-031-R11`), not one evidence item among several.

## Rollback

Trigger on false onboarding/settings/account state reaching a learner,
suspected cross-user exposure of onboarding/settings/account data, a
confirmed CSRF or duplicate-write defect on either new write surface, a
regression in any underlying P1–P4 write path or screen this package's
reliability/UX work touches, an accessibility or performance threshold
regression reaching production, or migration/schema failure. Use the
approved deploy/migration recovery procedure: preserve
`user_onboarding_profiles`/`user_settings`/`users.display_name` state,
restore the pre-P5 P1–P4 behavior cleanly, validate with non-production
identities, and preserve incident/rollback evidence. The last-known-good
revision is recorded at the future release decision, not guessed here.

## Independent verification, human approvals, and closure

Claude Code must report the final SHA, evidence, limitations, findings, the
active A-003 authority, and the remaining R3/R4/adoption/activation gates.
Routine R3 needs strengthened controls and independent verification; the
open `D06`–`D08` product/scope decisions need founder approval before their
affected tasks proceed. Closure requires all P5 task evidence, the new
accessibility/performance CI jobs confirmed blocking (not advisory), the
mock-decommission/extension inventory, and the DOC-12 P5 gate evidence;
neither a package merge nor a staging deploy alone closes the milestone,
live staging gate evidence is blocked until F3 exists (`VOC-031-DEP-02`),
and the P5 gate additionally requires the DOC-12 §5 P5 wording itself — the
full loop works coherently in staging across supported layouts with no
critical product/security/data/accessibility/reliability defect —
demonstrated with evidence, not merely asserted. Because the P1→P2→P3→P4→P5
acceptance chain has not itself closed any prior milestone's live-staging
gate (`VOC-031-DEP-01`), the P5 gate cannot be declared complete before
those upstream gates are, even once this package's own evidence is
otherwise ready.
