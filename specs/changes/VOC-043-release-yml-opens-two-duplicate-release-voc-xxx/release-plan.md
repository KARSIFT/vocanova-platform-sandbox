# VOC-043 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention. This
fix's "release" is landing in `karsift-ai-infra`'s reusable `release.yml`;
whatever mechanism this repository (and any other consumer of that reusable
workflow) uses to pick up a new revision determines when it actually takes
effect — see `specification.md`'s open question 2. This package does not
authorize or perform that pickup.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-043-T00` implemented, independently verified against the
exact revision to be released, and `VOC-043-AC-00`/`01`/`02` each satisfied
with real evidence (not asserted) per `acceptance-criteria.md` — in
particular, real evidence that the concurrent-events race
(`VOC-043-TEST-00`) was actually reproduced and actually closed, since a
race condition fix that is only reasoned about on paper without an actual
concurrent-run reproduction is the kind of unverified claim `CLAUDE.md`
requires an independent reviewer to reject.

Monitoring: after this fix is live (i.e. after whatever consumes this
reusable workflow picks up the new revision), monitor the next few package
completions with multi-task rosters (especially ones likely to have
near-simultaneous auto-merges, per `merge-gate.yml`'s auto-merge path) for:
- Exactly one "Release: `<change_id>`" issue per completed package, with no
  recurrence of the duplicate-issue pattern this issue reports.
- No new failure mode introduced by the job split (e.g. `identify`'s outputs
  not correctly reaching `check-and-open`, or the `concurrency:` group
  silently swallowing a legitimate run instead of queuing it).

Outcome owner: named explicitly in the implementation PR at merge time, not
left implicit.

## Rollback

Trigger: the two-job split (or whichever mechanism is chosen) introduces a
regression in the common single-event case, or fails to actually close the
race in a real concurrent-events scenario after landing.

Mechanism: revert the job-split commit, restoring the single
`check-completion` job exactly as it exists today — the same known,
if racy, behavior that is correct in the common case. This is a small,
independently reversible workflow-logic change with no migration, so
rollback is expected to be low-risk and fast.

Accountable owner: named explicitly in the implementation PR, not left
implicit.

Validation: confirm, after rollback, that `check-completion` behaves
identically to its pre-`VOC-043` state (i.e. the same known race issue #328
reports today), not some new intermediate state.

Last-known-good reference: the commit immediately preceding this package's
implementation PR, on whichever branch `karsift-ai-infra`'s own history
tracks (unconfirmed at drafting time — see `specification.md`'s open
question 2 for why this repository's own consuming reference is not yet
known).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R3 approvals: under active A-003, routine R3 does not require standing
technical-steward or founder approval solely because it is R3; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or
EHR trigger (given its cross-package, shared-infrastructure blast radius) is
for the reviewing human and independent verifier to determine at review
time, not assumed here.
Closure evidence: not yet produced. Repository merge, release, activation,
and closure are distinct and must not be conflated — this package documents
no closure yet, since no task has been implemented. Closure additionally
requires resolving `specification.md`'s open questions 1-3 (confirming the
concrete concurrency mechanism, confirming how this repository actually
consumes the fixed reusable workflow, and deciding on manual cleanup of the
two already-open duplicate issues #310/#311) before this issue can be
considered fully resolved, not merely fixed in `karsift-ai-infra`'s source.
