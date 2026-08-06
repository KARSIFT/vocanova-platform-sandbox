# VOC-043 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention.
Deployment requires whichever human/workflow authorization this repo's
existing `apps/web` release process already requires; this package changes
none of that authorization boundary.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-043-T00` through `VOC-043-T03` implemented, independently
verified against the exact revision to be released, and `VOC-043-AC-00`
through `AC-03` each satisfied with real evidence (not asserted) per
`acceptance-criteria.md`. Because `VOC-043-T01`'s exact diff depends on
`T00`'s real-request finding (per `specification.md`'s open question 1), the
reviewing human should confirm at release time that the merged fix actually
matches a finding recorded in `T00`'s evidence, not an unconfirmed guess.

Monitoring: this fix directly touches the server-side boundary that determines
whether a real authenticated user can actually reach a server-rendered
protected page (not merely pass `middleware.ts`). Post-deploy, monitor:
- nginx access logs for the same pattern the issue used to diagnose this (a
  successful OAuth callback and two successful middleware `/api/v1/me` checks
  no longer followed by a third `401` and a bounce to
  `/signin?returnTo=...`).
- A real successful end-to-end Google OAuth sign-in by an allowlisted account,
  actually reaching and rendering `/onboarding` or `/home` (not merely
  receiving a `200` from `/api/v1/me` at the middleware layer) — the most
  direct outcome signal, and the exact scenario this issue and its
  predecessors (VOC-038-T03, VOC-039, VOC-041, VOC-042) have progressively
  narrowed toward.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  `VOC-037-T04`'s monitoring/alerting evidence) already tracks for the
  affected auth/onboarding endpoints.
- If temporary diagnostic logging from `VOC-043-T00` is retained per
  `specification.md`'s open question 2, confirm it does not silently persist
  raw cookie/session-token values in production logs.

Outcome owner: named explicitly in the implementation PR at deploy time, not
left implicit here.

## Rollback

Trigger: the fix does not resolve the reported `401`, introduces an
unintended cookie-forwarding change elsewhere, or retained diagnostic logging
is found to over-log.

Mechanism: revert the small, self-contained application-code diff (no
migration, no infrastructure change) — expected to be low-risk and fast,
consistent with this repo's prior small-diff rollback precedent (VOC-039,
VOC-041, VOC-042).

Accountable owner: named explicitly in the implementation PR, not left
implicit.

Validation: confirm, after rollback, that `apps/web/src/lib/api-server.ts` and
`apps/web/src/middleware.ts` both return to their pre-`VOC-043` state (i.e.
the same broken state issue #333 reports today), not some new intermediate
state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time,
`a48f4d928bbd796c99f49825bb80eb1eab8ac7db`).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or
EHR trigger is for the reviewing human and independent verifier to determine
at review time, not assumed here.
Closure evidence: not yet produced. Repository merge, release, activation,
and closure are distinct and must not be conflated — this package documents
no closure yet, since no task has been implemented. Closure additionally
requires resolving `specification.md`'s open questions 1-3 (the confirmed root
cause, the diagnostic-logging retention decision, and the full
`createServerApiClient()` call-site inventory) before this issue can be
considered fully resolved in production, not merely fixed in the repository.
