# VOC-043 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention.
Deployment requires whichever human/workflow authorization this repo's existing
production-deploy procedure already requires; this package changes none of that
authorization boundary — it only changes one server-side cookie-forwarding
function in one file, plus its test coverage.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-043-T00` and `VOC-043-T01` implemented, independently
verified against the exact revision to be released, and
`VOC-043-AC-00`/`01`/`02` each satisfied with real evidence (not asserted) per
`acceptance-criteria.md`. Because the issue's own reproduction is against real
production Google sign-in, and this package's fix is offered as the "likely"
correction rather than an empirically pre-confirmed one (see
`specification.md`'s open question 1), the reviewing human and independent
verifier should explicitly confirm the fix's own regression test
(`VOC-043-TEST-00`) demonstrates the divergence and its correction before
release, not merely that the file compiles.

Monitoring: this fix directly touches the server-side boundary every
authenticated server-rendered page in `apps/web` depends on. Post-deploy (i.e.
after the next real redeploy of the `web` service that includes this fix),
monitor:
- nginx access logs for the same pattern the issue used to diagnose this (the
  third `/api/v1/me` call in a real sign-in sequence returning `200` instead of
  `401`, and `/onboarding` no longer redirecting to `/signin`).
- A real successful end-to-end Google OAuth sign-in by an allowlisted account,
  reaching `/onboarding` (or `/home`, if onboarding is already complete) and
  staying there — the most direct outcome signal, and the exact scenario
  (VOC-038-T03) whose repeated failure has now motivated VOC-039, VOC-041,
  VOC-042, and this package in sequence.
- Whatever this repo's existing production monitoring (Sentry/Kuma, per
  VOC-037-T04's monitoring/alerting evidence) already tracks for the affected
  onboarding/settings-account routes.

Outcome owner: named explicitly in the implementation PR at deploy time, not
left implicit here.

## Rollback

Trigger: a future edit reintroduces a `cookies()`-based reconstruction (guarded
pre-merge by `VOC-043-T01`'s regression test, but also a valid post-deploy
rollback trigger if it somehow reaches production), or the fix is confirmed not
to resolve the reported `401` per `specification.md`'s open question 1.

Mechanism: revert this package's diff (a small, self-contained change to
`apps/web/src/lib/api-server.ts` plus one new test file) or redeploy from the
last known-good artifact — whichever this repo's existing rollback tooling
already supports (per VOC-037's rollback evidence). This is a small,
independently reversible change with no migration, so rollback is expected to
be low-risk and fast.

Accountable owner: named explicitly in the implementation PR, not left
implicit.

Validation: confirm, after rollback, that `apps/web/src/lib/api-server.ts`
matches the same content it had before this package's fix (i.e. the pre-VOC-043
`cookies()`-based state — the same broken state issue #333 reports today), not
some new intermediate state.

Last-known-good reference: the commit at the tip of `develop` at drafting time
(`a48f4d9`).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or EHR
trigger is for the reviewing human and independent verifier to determine at
review time, not assumed here.

Closure evidence: not yet produced. Repository merge, release, activation, and
closure are distinct and must not be conflated — this package documents no
closure yet, since no task has been implemented. Closure additionally requires
resolving `specification.md`'s open questions 1-3 (whether the fix actually
resolves the reported 401 in a real reproduction, whether a broader
cookies()-reconstruction audit is warranted, and whether any real user is
currently stuck mid-onboarding and needs separate remediation) before this
issue can be considered fully resolved in production, not merely fixed in the
repository.
