# Pinned karsift-ai-infra contract fixtures (VOC-080-T05, VOC-097-T03, VOC-102-T00)

These copies are deterministic fixtures for caller-repo policy regressions.
They mirror `KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py`,
`tooling/governance/tests/test_voc097_*.py`, and
`tooling/governance/tests/test_auto_advance_ownership.py` can assert
merge/adopt/release/remediate/plan-review/live-evidence/auto-advance/role
contracts without cloning the infra repository in CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080-,
VOC-097-, or VOC-102-related infra contracts change and record the new pin in
evidence.

VOC-104 adds the pinned `ready-for-review-reuse.yml` and
`verify-ready-for-review-reuse.yml` contracts. The decision vocabulary is
`reuse-evidence`, `full-path`, and `fail-closed-to-full-path`; only the first may
skip duplicate exact-SHA CI/model work, and merge-gate still re-evaluates the
validated prior run identity. Reusable review records bind the exact pipeline run
ID in addition to base/head and package/task identity. Controlled proof keeps the
source transition PR/base/head/ref distinct from the later evidence-carrier SHA,
and requires the workflow-controlled `decide (ready_for_review)` job marker.
Task-owned live evidence is also bound to the exact task identifier at production,
duplicate-suppression, review-classification, and reuse-decision boundaries. The
post-transition verifier requires GitHub's authenticated source PR object to record
the PR as merged; a successful auto-merge job alone is not accepted as merge proof.
On an eligible unchanged draft-to-ready transition, `ci.yml` still emits the
repository ruleset's required `ci / ci` check context, but runs only a named
exact-SHA reuse marker. Checkout and application validation remain skipped. This
small compatibility context is necessary because GitHub evaluates the newest
required check context; omitting the reusable CI caller changes its visible name
to `ci` and leaves `ci / ci` unsatisfied even when prior evidence is valid. The
post-transition verifier requires the marker to succeed and both checkout steps
plus the full validation step to be skipped. Merge-gate also requires exactly one current `ci / ci`
context in `SUCCESS`; prior evidence can never replace a skipped or ambiguous
current compatibility context.

The VOC-104 workflow, policy, normalizer, and regression-test copies correspond
to independently reviewed shared-infra merge `115363806bae4f3b3b8c90670f4dbbb0b361a00f`.

## auto-advance ownership (VOC-102)

Adoption starts the first task automatically. The adopted roster records an explicit
`depends_on` edge from every later task to its predecessor, and `auto-advance.yml` releases the
next task only after the preceding task's implementation PR merges and its tracking issue closes.
For ordinary implementation tasks it dispatches `implement.yml` attempt 1. When the next roster
task has a valid `operator` or `live-actions` contract at
`<package>/.karsift/live-evidence/<task_id>.yaml`, auto-advance instead prepares a deterministic
draft evidence-carrier PR and posts one sanitized waiting marker without executing the general
implementer. A repeat event repairs a partial carrier publication but preserves any existing
operator evidence. Missing-required, malformed, invalid, duplicate, or conflicting ownership
metadata fails closed. “No marker” means a readable task stanza has no ownership marker; a missing
or unreadable `tasks.md` cannot establish that condition and fails closed instead of guessing that
the task is ordinary. The classifier and proof verifier stay read-only; only a clean publisher
receives carrier writes, and the fail-closed notifier receives issue-write only.
Carrier PRs inherit their `Risk classification: R#` line from exactly one valid
root `risk: R0` through `risk: R4` declaration in the adopted package's
`change.yaml`. Missing, duplicate, or malformed risk metadata fails closed; the
publisher does not over-classify every future carrier as R4.
If that deterministic branch already belongs to a closed or merged carrier PR,
publication fails closed with `conflicting_existing_pr`; it never silently
reopens or replaces historical state. An operator must confirm why the carrier
was closed before a governed restoration or cleanup/retry.
