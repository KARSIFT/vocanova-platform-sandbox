# VOC-108 — Release Plan

## Activation

Shared lifecycle behavior activates when the independently reviewed
`KARSIFT/karsift-ai-infra` PR merges to its default branch; this caller consumes
the reusable workflows at `@main`. Caller contract/evidence changes then merge
to `develop` and promote through the repository-controlled release path. No
manual server mutation or application deployment is authorized by this package.

## Preconditions

- Package adopted and T00 implementation authorized.
- Shared self-CI and deterministic race fixtures pass at the exact infra SHA.
- Any caller changes pass governance/foundation validation at the exact caller
  SHA.
- Independent reviews are bound to each final SHA with no unresolved blocking
  finding.
- Cross-repository PR text contains no caller-closing keyword.

## Live verification

Observe one normal caller task/release lifecycle after activation: authoritative
latest checks selected; caller task closes only after caller PR merge evidence;
release opens only after proven roster completion; duplicate/reconcile events are
no-ops; and no unchanged-SHA full CI rerun is needed merely to observe a terminal
external check. Record only repository/PR/issue/run IDs, SHAs, dates, and scrubbed
outcomes.

## Rollback

Revert shared workflow/helper changes via a reviewed infra PR and caller contract
changes via the normal governed caller path. Trigger rollback on false-positive
completion, failure to fail closed on ambiguous/latest-failed evidence, duplicate
effective merge, lost recovery path, or weakened auth/SHA/branch protection.
Re-run the negative fixtures after rollback and document that the issue #903
race class is restored until a corrected replacement lands.

## Closure

Close the task only after shared infra and caller evidence PRs merge and exact-SHA
verification is complete. Close issue #903 only after the resulting caller
release finishes and the live lifecycle evidence is recorded. Under A-004, no
founder `approved` comment is required; non-human gates remain mandatory.
