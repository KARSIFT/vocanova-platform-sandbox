# Remaining UI/UX findings — 12 November 2026

This delivery follows the [initial audit](ui-ux-audit-2026-11-12.md), released as
0.4.30. The owner requested the remaining actionable findings be implemented and
confirmed that the learner study should be prepared while they arrange learners.

## Implemented behavior

- Password signup and password-reset check-email screens retain resend and
  change-address controls. A 60-second wall-clock cooldown and immediate request
  lock prevent duplicate clicks; resend errors keep recovery available. Generic
  acknowledgements do not reveal account existence or claim guaranteed delivery.
  Signup passwords remain only in React memory and clear when changing address.
  The UI cooldown applies to the mounted recovery view; existing server-side
  per-IP and hashed-recipient budgets remain enforced across clients. A focused
  regression checks the production 10-request/minute window, normalization and
  renewal without sending mail. The UI timer is not a security boundary.
- Profile and unsent email-change inputs use the existing validated same-user,
  two-hour tab drafts. Confirmed actions clear resolved drafts; newer edits survive
  earlier responses and uncertain commits. User/account boundaries, successful
  logout and deactivation clear private recovery. Successful exit also invalidates
  the departing document before storage cleanup, so late responses cannot recreate
  cleared drafts or apply stale save results if replacement navigation fails. No password, confirmation proof,
  pending sent state or authentication cookie is persisted as a draft.
- Signup, Profile and Settings use the API's 80 Unicode code-point name limit,
  including supplementary characters, instead of a native UTF-16 unit limit.
  Visible, screen-reader-linked hints state the limit.
- Existing lists and their meanings precede optional create/rename/delete controls.
  Native keyboard-accessible disclosures keep management available; uncertain
  mutation errors reopen the relevant disclosure and focus recovery. The existing
  exact-body/idempotency/revision and confirmed-result behavior stays intact.
- Story cards have one primary action: start, continue the actual saved session,
  or read/practise again. An ongoing story retains a secondary overview link.
- A successful empty writing-topic response explains that no topics are available
  and offers saved-word practice. Outages remain errors with recovery rather than
  being presented as an empty catalog.
- The [catalog audit and editorial review](catalog-readability-2026-11-12.md)
  cover 10,051 canonical meanings. All 320 heuristic candidates received an
  individual decision: 257 original explanations and 63 retained editorial
  definitions. Exact-revision optional display explanations preserve canonical
  short definitions, sense IDs, media, grading and learner-state payloads.
  Discovery's lazy details preserve distinct existing authored explanations even
  when the current map card does not contain that field.

## Learner study

The [moderator protocol](learner-usability-study-2026-11-12.md) and
[blank private observation sheet](learner-usability-observation-template.md)
are ready. They specify four core tasks plus rotating optional tasks, preflight
state verification, consent, safe test accounts, non-leading prompts, observation
criteria and honest fixture/provider limitations. The owner will arrange learners.
**Zero sessions have been observed; no comprehension or usability score exists.**

## Verification and release boundary

Root verification is recorded with the delivery PR and release-readiness entry.
The implemented checks exercise recovery, account boundaries, response races,
Unicode limits, list/story action hierarchy, API empty success, outage recovery,
canonical identity, source fallback and exact meaning display. Local browser tests
use synthetic APIs; they do not prove email delivery, fresh real-user OAuth,
production database durability or A2–B1 comprehension. Existing authored exercise
snapshots and canonical answer options remain unchanged.

The whole-catalog audit is a reproducible source inventory, not a certification
that every definition meets a reading level or that every English sense is taught.
The study observations and human editorial judgments remain the next evidence
needed to evaluate the design with actual learners.
