# VOC-045 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment. `change.yaml`'s `release.deployment` and
`release.mode` are both set to `not-authorized-by-this-package`; a separate
release decision, made under the repository's active authority model
(A-003), is required before either task's change reaches production.

## Preconditions, monitoring, and outcome

- Exact revision: to be recorded as the merged commit SHA for each of
  `VOC-045-T00` and `VOC-045-T01` once implemented.
- Checks: `pnpm validate` (or its narrower `apps/api`-scoped equivalents)
  and `go test ./apps/api/business/auth/...`, per
  `implementation-plan.md`'s validation section.
- Approvals: independent verification by Claude Code against this
  package's specification and acceptance criteria, per `CLAUDE.md`; R3
  routine review under active A-003 (no standing technical-steward or
  founder approval required solely for being R3, per `AGENTS.md`), unless
  the reviewing human decides at adoption time that this specific reversal
  of a prior deliberate decision (`specification.md`'s open question 1)
  warrants elevated review.
- Staged evidence: none exists yet; this package is unadopted and
  unimplemented.
- Monitoring: after release, watch for any increase in unexpected sign-in
  failures or unexpectedly silent re-authentication (a user reaching an
  authenticated state without any visible confirmation step), which would
  indicate the chosen replacement condition (`specification.md`'s open
  question 2) was too permissive.
- Outcome owner: unassigned; to be recorded at adoption time.

## Rollback

- Trigger: any user signed in without a genuine, verified Google OAuth
  round-trip (regression of `VOC-045-AC-02`), any weakening of the CSRF
  state-token check (regression of `VOC-045-AC-03`), or a spike in
  authentication failures following release.
- Mechanism: revert the merged commit/PR for the affected task. No data
  migration is introduced by either task, so no data-compatibility rollback
  work is required.
- Accountable owner: unassigned; to be recorded at adoption time, per this
  repository's active authority model.
- Validation: re-run `go test ./apps/api/business/auth/...` and the
  applicable `pnpm validate` scope after any rollback to confirm the
  reverted state is coherent.
- Last-known-good reference: the `develop` branch commit immediately prior
  to the affected task's merge.

## Independent verification, human approvals, and closure

- Independent verification: Claude Code reviews the exact final revision of
  each task's pull request, binds its report to that commit SHA, and
  confirms Codex (or whichever implementer) did not approve or merge its
  own work, per `CLAUDE.md`.
- Human approvals: routine R3 under active A-003 does not require standing
  technical-steward or founder approval merely for being R3; strengthened
  applicable controls and independent verification remain required.
  Whether this specific package's reversal of a prior deliberate decision
  warrants elevated (e.g. founder) review is an open question for the
  adopting human, not settled here.
- Closure evidence: this package closes only once both tasks are
  implemented, independently verified, and `acceptance-criteria.md`'s
  criteria are all marked satisfied with linked evidence - repository
  merge, release, activation, and closure remain distinct events and must
  not be conflated.
