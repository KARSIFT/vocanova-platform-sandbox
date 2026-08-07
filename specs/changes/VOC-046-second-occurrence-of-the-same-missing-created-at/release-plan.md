# VOC-046 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention.
Deployment requires whichever human/workflow authorization this repo's
existing `apps/api` release process already requires; this package changes
none of that authorization boundary. Given the confirmed live production
outage this issue reports (every user with no mission yet today currently
blocked from `/home`), the reviewing human may reasonably prioritize
`VOC-046-T00`'s release once adopted and implemented, ahead of the broader
audit tasks — but that scheduling decision belongs to the human, not this
draft.

## Preconditions, monitoring, and outcome

Preconditions: at minimum `VOC-046-T00` implemented, independently verified
against the exact revision to be released, and `VOC-046-AC-00` satisfied
with real evidence (not asserted) per `acceptance-criteria.md`. The
remaining tasks (`T01`-`T03`) may reasonably release separately, per this
repository's small-PR convention, provided each is independently verified
before its own release.

Monitoring: this fix directly restores the core post-onboarding `/home`
path. Post-deploy, monitor:

- The same production error signature the issue used to diagnose this
  (`GET /api/v1/daily-mission -> 500`) — its absence, together with a
  corresponding drop in generic `/home` load failures, is the most direct
  outcome signal. Because the API's generic `500` handling masked the
  underlying Postgres error this time (unlike VOC-045), confirm whatever
  structured logging or error tracking this repo's existing production
  monitoring (Sentry/Kuma, per `VOC-037-T04`'s monitoring/alerting evidence)
  captures for this endpoint, since the plain HTTP status code alone may not
  distinguish this specific defect from any other `500`.
- A real successful `/home` load by a genuinely new (or realistically
  simulated new) user with no mission yet today, actually persisting a
  `daily_mission_snapshots` row with non-null `created_at`/`updated_at` —
  the exact scenario this issue reports as currently broken, and the same
  scenario VOC-038-T03's own core-loop validation depends on.
- For any call site fixed by `VOC-046-T02`'s audit, whatever equivalent
  error signature or endpoint that call site backs, per that task's own
  evidence.

Outcome owner: named explicitly in the implementation PR at deploy time, not
left implicit here.

## Rollback

Trigger: any fixed call site still raises the underlying `NOT NULL`
violation in production, or a fix introduces an unintended change to an
existing `ON CONFLICT DO UPDATE` branch's behavior for a user who already
has a row.

Mechanism: revert the small, self-contained application-code diff (no
migration, no infrastructure change, for this package's primary fix) —
expected to be low-risk and fast, consistent with this repo's prior
small-diff rollback precedent (VOC-045 and earlier). Any rows successfully
created by real users after deployment and before a rollback remain valid
and require no cleanup, since they are correctly-formed rows, not corrupted
ones.

Accountable owner: named explicitly in the implementation PR, not left
implicit.

Validation: confirm, after rollback, that every file this package changed
returns to its exact pre-`VOC-046` state (the same broken state issue #352
reports today for the confirmed call sites), not some new intermediate
state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time,
`b6ecae02b944b7eb7462971616ae988a29ee9358`).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or
EHR trigger (e.g. if the audit surfaces a call site touching a materially
more sensitive table, or if `VOC-046-T03`'s schema-scanning check is
adopted and found to have a broader effect than expected) is for the
reviewing human and independent verifier to determine at review time, not
assumed here.
Closure evidence: not yet produced. Repository merge, release, activation,
and closure are distinct and must not be conflated — this package documents
no closure yet, since no task has been implemented. Closure additionally
requires resolving `specification.md`'s open questions 1-3 (the audit
scope's exact boundaries, whether the schema-scanning check is in scope,
and whether the generic-500-masking behavior is in scope) before this issue
can be considered fully resolved, not merely fixed at the confirmed call
site.
