# VOC-039 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention. Deployment
requires whichever human/workflow authorization this repo's existing
`deploy-staging.yml`/`deploy-production.yml` already requires; this package changes
none of that authorization boundary.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-039-T00` through `T02` implemented, independently verified
against the exact revision to be released, and `VOC-039-AC-00`/`01`/`02` each
satisfied with real evidence (not asserted) per `acceptance-criteria.md`.

Monitoring: this fix directly touches the production authentication gate for every
protected route. Post-deploy, monitor:
- The new structured log lines (`VOC-039-T02`) themselves, as the primary
  observability this package adds — a spike in "got 401" or "fetch threw" log
  volume immediately after deploy would indicate the fix did not fully resolve the
  issue, and is the concrete signal this package exists to make visible without
  needing SSH access.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  `VOC-037-T04`'s monitoring/alerting evidence) already tracks for latency and
  error rate on the protected routes, given `specification.md`'s open question 2
  about Node-runtime middleware's documented heavier resource profile relative to
  Edge runtime.
- A real successful end-to-end OAuth sign-in by an allowlisted account reaching a
  protected route, immediately after deploy, as the most direct outcome signal.

Outcome owner: named explicitly in the implementation PR at deploy time, not left
implicit here.

## Rollback

Trigger: the auth check begins failing for real users in a new way after this
change (fail-open observed, a Node-runtime-specific build/runtime error, or a
latency/resource regression severe enough to be release-blocking).

Mechanism: revert the single-line `runtime = "nodejs"` export (and, if bundled in
the same PR, the accompanying logging change) or redeploy the last known-good
artifact — whichever this repo's existing rollback tooling already supports (per
`VOC-037`'s rollback evidence and `VOC-038-T06`'s rehearsal convention). This is a
one-line, independently reversible change with no migration, so rollback is
expected to be low-risk and fast relative to the schema-touching changes those
prior packages' rehearsals covered.

Accountable owner: named explicitly in the implementation PR, not left implicit.

Validation: confirm, after rollback, that the auth check returns to its
pre-`VOC-039` state (i.e. Edge runtime, still broken for real users — the same
state issue #297 reports today) rather than some new intermediate broken state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time, `d45f002` — the merge of PR #296,
the last of the two already-confirmed prior fixes in this investigation).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or
EHR trigger is for the reviewing human and independent verifier to determine at
review time, not assumed here.
Closure evidence: not yet produced. Repository merge, release, activation, and
closure are distinct and must not be conflated — this package documents no
closure yet, since no task has been implemented.
