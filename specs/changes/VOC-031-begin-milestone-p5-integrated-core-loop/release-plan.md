# VOC-031 — Release Plan

## Release and deployment authorization

No release or deployment is authorized by this draft. Merge into `develop`, staging validation,
production activation, and the DOC-12 P5 milestone gate are distinct and none is granted here.
Production and autonomous production release remain disabled. `T03` and `T04` additionally cannot
finalize their evidence quality until `D02` and `D03` (respectively) are resolved, and `T05`'s
gate-readiness interpretation depends on `D01`. This draft is not adopted.

## Preconditions, monitoring, and outcome

Before any staging candidate: an adopted package with `D01`–`D04` resolved (or explicitly
deferred with a recorded rationale), the exact base/revision, required PR checks, non-production
identities, exact-SHA independent review, and a named rollback owner/procedure. Monitor: the rate
of the `T01` "backend unreachable, please retry" state versus genuine `401` redirects (a
sustained high rate of the former would itself be an incident, not just a UX detail); sentence-
practice and review-submission retry/duplicate-attempt counts (should remain zero duplicates
post-`T01`); and, if `D02`/`D03` activate them, accessibility-scan and Lighthouse-score trends over
time. The founder owns the open `D01`–`D04` product/scope/tooling decisions; the future release
authority records the accountable technical/operational owner. Live staging gate evidence is
blocked until the F3 staging environment exists (`VOC-031-DEP-02`).

## Rollback

Trigger on: any authorization regression traced to the `T01` middleware change; a duplicate reward
or lost learner-input incident traced to the `T01` retry-safety changes; a confirmed critical
accessibility regression reaching learners; or a regression in the underlying A1/P1/P2/P3/P4 loop
this package integrates but does not itself own. Use the approved redeploy-previous-artifact
procedure: this package touches no backend write path, migration, or schema, so no learner data
state is at risk from a rollback — redeploy the previous known-good frontend/middleware artifact,
confirm the pre-P5 loop still functions, validate with non-production identities, and preserve
incident/rollback evidence. The last-known-good revision is recorded at the future release
decision, not guessed here.

## Independent verification, human approvals, and closure

Claude Code must report the final SHA, evidence, limitations, findings, the active A-003 authority,
and the remaining R3/R4/adoption/activation gates. Routine R3 needs strengthened controls and
independent verification; the open `D01`–`D04` product/scope/tooling decisions need founder
approval before their affected tasks' evidence is treated as final. Closure requires all P5 task
evidence, the no-new-backend-scope evidence (`VOC-031-AC-07`), the P5-specific mock-inventory
extension, and the DOC-12 P5 gate evidence; neither a package merge nor a staging deploy alone
closes the milestone, live staging gate evidence is blocked until F3 exists
(`VOC-031-DEP-02`), and the P5 gate additionally requires the DOC-12 §5 P5 wording itself — the
full loop works coherently across supported layouts with no critical product/security/data/
accessibility/reliability defect — demonstrated with evidence, not merely asserted. Given P5's
position immediately before R1 (Staging Readiness) in the DOC-12 §6 dependency chain, any
limitation recorded here (F3 absence, open `D01`–`D04`, deferred automated tooling) is also a
direct precondition R1 must resolve before its own gate can pass.
