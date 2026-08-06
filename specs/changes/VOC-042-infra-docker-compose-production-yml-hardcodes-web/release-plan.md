# VOC-042 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention. Deployment
requires whichever human/workflow authorization this repo's existing
`deploy-production.yml`/production-recreate procedure already requires; this
package changes none of that authorization boundary — it only changes one value
inside one service's `environment:` block in one compose file.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-042-T00` and `VOC-042-T01` implemented, independently verified
against the exact revision to be released, and `VOC-042-AC-00`/`01`/`02` each
satisfied with real evidence (not asserted) per `acceptance-criteria.md`. Because
this fix only takes effect the next time production's `web` service is recreated,
the reviewing human's decision on `specification.md`'s open question 2 (whether the
currently-running host needs an interim manual recreate) should be resolved before
or alongside merge, not left implicit.

Monitoring: this fix directly touches the server-side boundary
`apps/web/src/middleware.ts` uses for every protected route's auth check.
Post-deploy (i.e. after the next real recreation of the `web` service that
includes this fix), monitor:
- nginx access logs for the same pattern the issue used to diagnose this
  (a successful OAuth callback `302` no longer followed by
  `/home` -> `/onboarding` -> `/signin?returnTo=...` bounces).
- `apps/web/src/middleware.ts`'s server-side `/api/v1/me` fetch outcome, if this
  repo's existing logging/monitoring surfaces it (per `VOC-037-T04`'s
  monitoring/alerting evidence) — confirming the fetch now succeeds rather than
  silently failing.
- A real successful end-to-end Google OAuth sign-in by an allowlisted account,
  landing on `/home` or `/onboarding` and staying there — the most direct outcome
  signal, and the exact scenario (VOC-038-T03) whose repeated failure has now
  motivated VOC-039, VOC-041, and this package in sequence.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  `VOC-037-T04`'s monitoring/alerting evidence) already tracks for the affected
  auth endpoints.

Outcome owner: named explicitly in the implementation PR at deploy time, not left
implicit here.

## Rollback

Trigger: a future edit reintroduces an unqualified value, or the appended port
turns out to be wrong for a reason not anticipated at drafting time.

Mechanism: revert the single-line change (and, if bundled in the same PR, the
regression check) or redeploy/recreate from the last known-good artifact —
whichever this repo's existing rollback tooling already supports (per
`VOC-037`'s rollback evidence and `VOC-038-T06`'s rehearsal convention). This is a
small, independently reversible change with no migration, so rollback is expected
to be low-risk and fast.

Accountable owner: named explicitly in the implementation PR, not left implicit.

Validation: confirm, after rollback, that the compose file specifies the same
value it specified before this package's fix (i.e. the pre-`VOC-042` unqualified
state — the same broken state issue #319 reports today), not some new intermediate
state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time, `a8ff2e3` — the merge of
`VOC-041-T01`, PR #316).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or EHR
trigger is for the reviewing human and independent verifier to determine at review
time, not assumed here.
Closure evidence: not yet produced. Repository merge, release, activation, and
closure are distinct and must not be conflated — this package documents no closure
yet, since no task has been implemented. Closure additionally requires resolving
`specification.md`'s open questions 1-3 (broader-audit follow-up decision, whether
the live host needed an interim manual recreate and, if so, that it was actually
performed and recorded, and `infra/docker-compose.yml`'s unaudited status) before
this issue can be considered fully resolved in production, not merely fixed in the
repository.
