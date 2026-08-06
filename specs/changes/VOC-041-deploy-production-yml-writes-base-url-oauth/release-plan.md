# VOC-041 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention. Deployment
requires whichever human/workflow authorization this repo's existing
`deploy-production.yml` already requires; this package changes none of that
authorization boundary — it only changes three values and a comment inside one
existing step of that workflow.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-041-T00` and `VOC-041-T01` implemented, independently verified
against the exact revision to be released, and `VOC-041-AC-00`/`01`/`02` each
satisfied with real evidence (not asserted) per `acceptance-criteria.md`. Because
this fix only takes effect on the *next* dispatch of `deploy-production.yml`, the
reviewing human's decision on `specification.md`'s open question 2 (whether the
currently-live host needs an interim manual correction) should be resolved before
or alongside merge, not left implicit.

Monitoring: this fix directly touches the production OAuth/CORS boundary for
Google sign-in. Post-deploy (i.e. after the next real dispatch that includes this
fix), monitor:
- nginx access logs for the same pattern the issue used to diagnose this
  (`OPTIONS .../oauth/google/start` returning `204` followed by a real
  `POST .../oauth/google/start` — confirming the browser no longer silently blocks
  the follow-up request).
- The API's own CORS-allowed-origins boot log line
  (`api: cors allowed origins: ...`), confirming it now includes the `:8443` port.
- A real successful end-to-end Google OAuth sign-in by an allowlisted account, as
  the most direct outcome signal — this is the exact scenario (VOC-038-T03) whose
  repeated failure motivated both this issue and VOC-039.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  `VOC-037-T04`'s monitoring/alerting evidence) already tracks for the affected
  auth endpoints.

Outcome owner: named explicitly in the implementation PR at deploy time, not left
implicit here.

## Rollback

Trigger: a future dispatch writes a value that breaks sign-in in a new way (a
malformed port, or an allow-list change that admits an origin beyond production's
own host).

Mechanism: revert the three-line change (and, if bundled in the same PR, the
comment correction) or redeploy the last known-good artifact — whichever this
repo's existing rollback tooling already supports (per `VOC-037`'s rollback
evidence and `VOC-038-T06`'s rehearsal convention). This is a small, independently
reversible change with no migration, so rollback is expected to be low-risk and
fast.

Accountable owner: named explicitly in the implementation PR, not left implicit.

Validation: confirm, after rollback, that the workflow writes the same values it
wrote before this package's fix (i.e. the pre-`VOC-041` unqualified state — the
same broken state issue #312 reports today), not some new intermediate state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time, `6e6a44c` — the merge of
`VOC-039-T01`, PR #308).

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
`specification.md`'s open question 2 (whether the live host needed an interim
manual correction, and if so, that it was actually performed and recorded) before
this issue can be considered fully resolved in production, not merely fixed in the
repository.
