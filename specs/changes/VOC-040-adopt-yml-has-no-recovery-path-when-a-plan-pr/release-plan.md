# VOC-040 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment, per this repository's own template convention. This
package's own change has no deployment surface at all — it edits `AGENTS.md`
only — so no `deploy-staging.yml`/`deploy-production.yml` run is expected as a
consequence of merging it.

## Preconditions, monitoring, and outcome

Preconditions: `VOC-040-T00` implemented and independently verified against the
exact revision to be released, and `VOC-040-AC-00` satisfied with real evidence
(the merged `AGENTS.md` diff itself) per `acceptance-criteria.md`.

Monitoring: none applicable in the usual production-monitoring sense — this
change has no runtime component. The only real "outcome" to observe is whether a
future plan-adoption incident of the same shape (a `plan/`-branch PR merged
without the adoption edit) is actually resolved faster/without re-derivation by
whoever hits it next, using this documentation. That is not something this
package can measure at merge time; it is left as a qualitative observation for
whenever (if ever) the procedure is next used.

Outcome owner: named explicitly in the implementation PR, not left implicit.

## Rollback

Trigger: the documented procedure is found to be inaccurate (e.g. `adopt.yml`'s
actual current behavior no longer matches what is described, or the exact `gh run
rerun --failed` invocation shape documented does not work as written).

Mechanism: `git revert` of the implementation commit, or a follow-up documentation
correction PR — either is low-risk given this is an additive documentation-only
change with no runtime or migration surface.

Accountable owner: named explicitly in the implementation PR, not left implicit.

Validation: confirm, after rollback or correction, that `AGENTS.md` either returns
to its pre-`VOC-040` state or contains a corrected version of the procedure — not a
silently half-reverted section.

Last-known-good reference: the commit immediately preceding this package's
implementation PR on `develop` (at drafting time, `0a6de36` — see `change.yaml`'s
`base_sha`).

## Independent verification, human approvals, and closure

Independent verifier result: not yet produced — pending implementation.
R1/R3 approvals: under active A-003, routine work does not require standing
technical-steward or founder approval solely because of its risk class; strengthened
applicable controls and independent verification remain required (see
`CLAUDE.md`). Whether this specific change also implicates a separate R4 or EHR
trigger is for the reviewing human and independent verifier to determine at review
time, not assumed here — though nothing about a documentation-only addition to
`AGENTS.md` obviously does.
Closure evidence: not yet produced. Repository merge, release, activation, and
closure are distinct and must not be conflated — this package documents no closure
yet, since `VOC-040-T00` has not been implemented.
