# VOC-045 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention.
Deployment requires whichever human/workflow authorization this repo's
existing `apps/api` release process already requires; this package changes
none of that authorization boundary. Given the confirmed live production
outage this issue reports (every real new user currently blocked), the
reviewing human may reasonably prioritize this package's release once
adopted and implemented — but that scheduling decision belongs to the human,
not this draft.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-045-T00` through `VOC-045-T02` implemented,
independently verified against the exact revision to be released, and
`VOC-045-AC-00` through `AC-03` each satisfied with real evidence (not
asserted) per `acceptance-criteria.md`.

Monitoring: this fix directly restores the core new-user onboarding path.
Post-deploy, monitor:
- The same production error signature the issue used to diagnose this
  (`pq: null value in column "created_at"` / `"updated_at"` of relation
  `"user_settings"`) — its absence, together with a corresponding drop in
  `POST /api/v1/onboarding -> 400` responses, is the most direct outcome
  signal.
- A real successful onboarding completion by a genuinely new (or
  realistically simulated new) user, actually persisting a `user_settings`
  row with non-null `created_at`/`updated_at` — the exact scenario this
  issue reports as currently broken, and the same scenario VOC-038-T03's
  own core-loop validation depends on.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  `VOC-037-T04`'s monitoring/alerting evidence) already tracks for the
  onboarding endpoint and `user_settings`-writing code paths.
- Gamification's own lazy-creation path (`UpsertUserSettings`), for any user
  who reaches it before onboarding (if such a path exists in this repo's
  current flow) — confirming it also no longer raises the same constraint
  violation.

Outcome owner: named explicitly in the implementation PR at deploy time, not
left implicit here.

## Rollback

Trigger: the fix does not resolve the reported `NOT NULL` violation, or
introduces an unintended change to either upsert's existing
`ON CONFLICT DO UPDATE` behavior for a user who already has a
`user_settings` row.

Mechanism: revert the small, self-contained application-code diff (no
migration, no infrastructure change, for this package's primary fix) —
expected to be low-risk and fast, consistent with this repo's prior
small-diff rollback precedent (VOC-039, VOC-041, VOC-042, VOC-043). Any
`user_settings` rows successfully created by real users after deployment and
before a rollback remain valid and require no cleanup, since they are
correctly-formed rows, not corrupted ones.

Accountable owner: named explicitly in the implementation PR, not left
implicit.

Validation: confirm, after rollback, that
`apps/api/business/users/postgres.go` and
`apps/api/business/gamification/repository.go` both return to their
pre-`VOC-045` state (i.e. the same broken state issue #341 reports today),
not some new intermediate state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time,
`fb1e1a33ad6b0cc58ea42c2c573e120d35fd8539`).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or
EHR trigger (e.g. if the reviewing human decides open question 2's
schema-level `DEFAULT now()` migration is also in scope) is for the
reviewing human and independent verifier to determine at review time, not
assumed here.
Closure evidence: not yet produced. Repository merge, release, activation,
and closure are distinct and must not be conflated — this package documents
no closure yet, since no task has been implemented. Closure additionally
requires resolving `specification.md`'s open questions 1-3 (the
`gamification` timestamp-sourcing approach, whether a schema-level `DEFAULT`
is also in scope, and whether `mapOnboardingError`'s messaging is in scope)
before this issue can be considered fully resolved, not merely fixed in the
two named call sites.
