# VOC-053 — Release Plan

## Release and deployment authorization

This package requests no new deployment authority. Once adopted and
implemented, each task's pull request follows the existing governed path:
independent review, then automatic merge into `develop` (per
`karsift-ai-infra`'s `merge-gate.yml`, already live), then — per AGENTS.md's
"Release and deployment authority" section — automatic promotion to `main`
and automatic production deployment once the package's task roster closes,
with the founder-approval comment path remaining available as a manual retry
mechanism only. This package does not alter any of that mechanism.

## Preconditions, monitoring, and outcome

Precondition: `VOC-053-T00` must confirm a real root cause before
`VOC-053-T01` implements a fix, and `VOC-053-T01` must be merged to `develop`
before a real staging deploy can exercise it for `VOC-053-T02`'s
verification. Monitoring: the staging E2E workflow run's own log
(`tests/staging-e2e/core-loop.staging.spec.ts`'s pass/fail output and step 7's
specific `reviewedBefore`/`reviewedCards`/`reviewedAfter` values) is the
direct evidence source; no new dashboard or alerting is introduced by this
package. Outcome owner: whoever implements `VOC-053-T02` is responsible for
confirming and recording, in that task's own evidence file, that a real
staging deploy run — specifically one where the synthetic account already
carries `reviewedBefore >= 1` residue — passed step 7.

This package's completion directly unblocks VOC-052-T01's own staging E2E
evidence requirement (per issue #450's "Impact" section), which in turn
unblocks VOC-052's completion and the pending PR #435 `develop` → `main`
release — this package does not itself modify VOC-052's or PR #435's scope,
only removes the blocking failure they depend on not recurring.

## Rollback

Trigger: a staging deploy failure directly attributable to this package's
fix, or a subsequent real run showing the fixed behavior is still wrong.
Mechanism: revert the affected task's diff. Because no candidate fix under
consideration is expected to require a destructive data change (see
`impact-analysis.md`'s `VOC-053-R02` for the one conditional exception, which
would itself need to be additive/corrective if adopted), reverting the code
change is expected to be sufficient. Accountable owner: the implementer of
the affected task. Validation: confirm a subsequent staging deploy succeeds
again with the revert applied, and that
`tests/staging-e2e/core-loop.staging.spec.ts` returns to its pre-fix
(known) failure mode rather than a new one, so the revert is confirmed clean.
Last-known-good reference: the affected file's revision immediately
preceding this package's merge.

## Independent verification, human approvals, and closure

Independent verification (per `CLAUDE.md`) must confirm, against the exact
implemented revision's commit SHA: `VOC-053-T00`'s evidence genuinely
supports the claimed root cause with direct, reproducible evidence (not
merely a plausible-sounding narrative); the fix in `VOC-053-T01` is scoped to
that confirmed cause and does not weaken
`tests/staging-e2e/core-loop.staging.spec.ts`'s step 7 assertion or any other
existing check; `VOC-053-T02`'s staging verification evidence actually
exercises the originally-observed failure condition
(`reviewedBefore >= 1` from prior-run residue), not only a fresh-state pass
that would not have caught the original bug; no unrelated change was
introduced; and (per this repository's active authority model,
`a003-active`) that no standing technical-steward or founder approval is
being silently assumed beyond what A-003 already delegates for routine R3
work, while any R4-level consequence this drafting pass did not anticipate
(e.g. if `VOC-053-R02`'s conditional migration path is taken and it affects
real user data at a scale or nature this package did not anticipate) is
escalated rather than resolved by the implementer or reviewer alone.
Repository merge into `develop` and production release/deployment are not
the same event as closure — closure requires this package's acceptance
criteria being recorded as passing with their linked evidence, not merely a
merged PR.
