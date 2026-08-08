# VOC-049 — Release Plan

## Release and deployment authorization

Not authorized by this package on its own. `change.yaml`'s `release.mode`
and `release.deployment` are both `not-authorized-by-this-package`. However,
this package's entire purpose is to promote content to `main`, and
`AGENTS.md`'s "Release and deployment authority" section (2026-08-08) already
delegates automatic production deployment on every push to `main` via
`deploy-production.yml` — including, once the batch this package promotes
lands, the `auto_release_enabled` path itself for future packages. This
package does not add any new deployment authorization; it exercises the
promotion step that makes already-delegated, already-documented automatic
deployment apply to this specific, already-reviewed batch. A human adopting
this package is explicitly deciding to trigger that already-existing
delegated path for this content, not granting a new one.

## Preconditions, monitoring, and outcome

- Exact revision: to be recorded as `VOC-049-T00`'s re-verified `develop` tip
  SHA, and (if `T01` is dispatched) the resulting `main` tip SHA after
  promotion.
- Checks: `bash scripts/governance/validate-governance.sh` and
  `bash scripts/governance/classify-change-risk.sh` against the promoted
  diff, per `implementation-plan.md`'s validation section. Since the promoted
  diff should be empty relative to already-reviewed `develop` content, these
  checks confirm no unreviewed content rides along, not that new content
  passes review for the first time.
- Approvals: independent verification by Claude Code against this package's
  specification and acceptance criteria, per `CLAUDE.md`; R2 (as proposed)
  or R3 routine review under active A-003 (no standing technical-steward or
  founder approval required solely for being R3, per `AGENTS.md`) — unless
  the reviewing human decides at adoption time that the `auto_release_enabled`
  content riding along (per `specification.md`'s open question 2) warrants
  elevated review before this package's own promotion task is dispatched.
- Staged evidence: `VOC-049-T00` completed with implementation-time compare
  output captured in `t00-evidence.md` (`VOC-049-EV-00`). `VOC-049-T01` remains
  pending and must re-run T00 if `develop` moves before promotion.
- Monitoring: after promotion, watch `deploy-production.yml`'s run for the
  resulting `main` push, and watch for any regression traceable to the
  promoted governance/infra-cleanup content itself (e.g. a status-sync
  commit that misrepresents a package's real adoption state, surfaced by a
  subsequent `adopt.yml`/`release.yml` run behaving unexpectedly).
- Outcome owner: unassigned; to be recorded at adoption time.

## Rollback

- Trigger: any post-promotion production issue traceable to content in this
  batch, or a promotion that included content beyond the re-verified set
  (regression of `VOC-049-AC-01`).
- Mechanism: revert the promotion merge commit on `main`. No data migration
  is introduced by this batch, so no data-compatibility rollback work is
  required.
- Accountable owner: unassigned; to be recorded at adoption time, per this
  repository's active authority model.
- Validation: re-run `bash scripts/governance/validate-governance.sh` against
  `main`'s post-rollback state to confirm it matches the last-known-good
  reference.
- Last-known-good reference: `main`'s tip immediately prior to this
  package's promotion commit (`0914ea7` at drafting time, subject to
  `VOC-049-T00`'s re-verification at implementation time).

## Independent verification, human approvals, and closure

- Independent verification: Claude Code reviews the exact final promoted
  revision, binds its report to that commit SHA, confirms no implementer
  self-approved the promotion, and confirms which of
  `specification.md`'s open question 1's two mechanisms was actually used
  and was explicitly authorized rather than defaulted to, per `CLAUDE.md`.
- Human approvals: routine review under active A-003 does not require
  standing technical-steward or founder approval merely for being R2/R3;
  strengthened applicable controls and independent verification remain
  required. Whether the `auto_release_enabled` content riding along in this
  batch warrants elevated (e.g. founder) review before this package's own
  promotion task is dispatched is an open question for the adopting human,
  not settled here (per `specification.md`'s open question 2).
- Closure evidence: this package closes once either (a) `VOC-049-T00` finds
  a zero gap and that finding is recorded as `VOC-049-AC-03`'s satisfying
  evidence with no promotion performed, or (b) `VOC-049-T01` completes,
  independent verification confirms the exact promoted SHA, and
  `acceptance-criteria.md`'s criteria are all marked satisfied with linked
  evidence. Repository merge, release, activation, and closure remain
  distinct events and must not be conflated.
