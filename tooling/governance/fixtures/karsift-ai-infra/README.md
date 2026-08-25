# Pinned karsift-ai-infra contract fixtures (VOC-080-T05, VOC-097-T03, VOC-102-T00, VOC-104, VOC-106, VOC-108, VOC-115, VOC-117)

These copies are deterministic fixtures for caller-repo policy regressions.
They mirror `KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py`,
`tooling/governance/tests/test_voc097_*.py`,
`tooling/governance/tests/test_voc115_package_task_policy.py`,
`tooling/governance/tests/test_auto_advance_ownership.py`,
`tooling/governance/tests/test_remediate_ownership.py`, and related suites can
assert merge/adopt/release/remediate/plan-review/live-evidence/auto-advance/role
contracts without cloning the infra repository in CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080-,
VOC-097-, VOC-102-, VOC-104-, VOC-106-, VOC-108-, or VOC-115-related infra contracts change and
record the new pin in evidence.

## package/task defaults (VOC-115)

Planning now chooses the largest safe coherent package for the complete user or
business outcome. A plan may be broad or massive and contain several tasks, but it
must use the minimum sufficient number of maximal tasks; one end-to-end task and
implementation PR is the default whenever technically possible. Backend, frontend,
contracts, tests, docs, configuration, skills, and rollback evidence stay together
when they serve the same outcome. `config/package_task_policy.py` and
`package-task-policy-runner.py` fail closed when a second or later task lacks an
allowed split reason with a concrete explanation, or when more than three tasks lack
package-level justification. `plan.yml` enforces this in the planner retry loop for
newly drafted packages. Line, file, component, skill, repository, or layer counts and
convenience are never sufficient split reasons. The caller fixture must pin the exact
reviewed shared-infra merge that provides this contract.

VOC-104 adds the pinned `ready-for-review-reuse.yml` and
`verify-ready-for-review-reuse.yml` contracts. The decision vocabulary is
`reuse-evidence`, `full-path`, and `fail-closed-to-full-path`; only the first may
skip duplicate exact-SHA CI/model work, and merge-gate still re-evaluates the
validated prior run identity. Reusable review records bind the exact pipeline run
ID in addition to base/head and package/task identity. Controlled proof keeps the
source transition PR/base/head/ref distinct from the later evidence-carrier SHA,
requires the workflow-controlled `decide (ready_for_review)` job marker, and
recomputes the exact latest eligible prior run instead of trusting a supplied ID.
Both source runs must resolve the complete reuse-policy workflow set to one
identical authenticated shared-infra SHA. If GitHub clears a closed run's PR
association, a prior run is admitted only through its exact App-authored review
record on that PR. Optimized merge publishes one App-authored pre-merge record
binding repository, PR, branch, base/head, ready/prior run IDs, and policy SHA;
the post-transition verifier requires that unique attestation.
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

## auto-advance ownership (VOC-102)

Adoption starts the first task automatically. The adopted roster records an explicit
`depends_on` edge from every later task to its predecessor. `auto-advance.yml` releases the next
task only after one valid App-authored completion marker proves the preceding task's exact reviewed
caller PR merged. The issue-close event is only a wake-up hint; closed state alone cannot advance.
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

## remediation ownership gate (VOC-106)

`remediate.yml` resolves task ownership from the exact reviewed PR head before any
implementer retry. Valid `operator` or `live-actions` contracts suppress `implement.yml`
dispatch for review `FAIL` and CI failure, posting a sanitized operator/reconcile
escalation instead. Malformed or contradictory ownership metadata uses a separate
fail-closed escalation. Ordinary implementer-owned tasks retain the bounded retry.
WAITING / STALE / review-infra failures remain non-retrying and do not consume an
implementation attempt. A separate read-only
`verify-remediate-operator-ownership.yml` workflow validates controlled remediate
proof from Actions, issue, and PR metadata only—never logs or artifacts. Its
runner extracts the source run's associated PR base SHA from the
`pull_requests` list object (never by stringifying that list).

The implementer job deliberately has no `actions` permission and receives no
general Actions inspection/dispatch credential. Operator reconciliation is a
separate repository-controlled responsibility; adding it must not broaden the
implementer's permissions or secrets.

The VOC-106 workflow, policy, verifier, and regression-test copies correspond to
shared-infra merge `db164eb3905a96b74b039ab6aa36944408bf0a44`, including the
hosted verifier base-SHA adapter fix recorded in this package's T00 remediation.

## authoritative lifecycle state (VOC-108)

VOC-108 originally advanced the fixture to shared-infra merge
`d3108dfdef34e2f98c028916e95c36130d329132`; VOC-115 then advanced it to
`3fd40f52aba602fab8399482bc5b772731675d1a`, and VOC-114 now advances the
consolidated fixture pin through `30cc0a6f443b95e45527b03094767b8357b0a2dc`
and `bdc6736568827103b48255521f4bc83d5103bd3b` to
`9d7e334f917643c42bb4b7a062c8fcddecc7927f`, then to
`6999e2beda5bbf00028fae04ca0e65324fc59afa`, and finally to
`c5d8bccfa8676bd367b53ad5f6f9a51a40c99405`.
Adoption, merge/reuse, and release
select the newest authoritative attempt per logical exact-SHA gate from complete
paginated histories and bind the selected evidence to the authenticated pull
request's repository, number, base, and head. The merge gate writes one App-authored
caller-merge marker only for task branches before emitting the task close event;
auto-advance and release validate that same marker against live PR state. A
premature close is a safe wake-up no-op. Foreign qualified closing references are
rejected at merge time, while cross-repository links remain non-closing.
Automatic, reconcile, promotion-PR, check-provider, and external-workflow wake-ups
share one serialized promotion evaluator and one exact-head merge command. Shared
instructions explicitly define issue closure as a wake-up hint rather than task
completion authority.
The pinned caller template also retains the read-only
`verify-remediate-operator-ownership` dispatch surface used by live callers.
External `check_run` wake-ups invoke release evaluation only when the check is
attached to the `develop` → `main` promotion pull request.
Auto-advance comments and diagnostics use the current serialized-convergence
name, preventing the retired `check-completion` job name from becoming false
operator guidance.
The generated adoption roster PR is evaluated through the same paginated,
exact-identity newest-attempt selector as the merged plan head. Its complete
green logical set must stabilize on one unchanged head before exact-SHA merge.
The merge-gate App credential contract names both authoritative completion
actions: publish the immutable task marker, then close the linked task so the
release observer receives its authenticated wake-up events.
Close/reopen or draft/ready transitions on a promotion pull request do not
recover missing required checks; VOC-113 recovery dispatches genuine allowlisted
workflows for the exact SHA with a bounded 1800-second fail-closed wait. Recovery
calls are serialized by mode and target SHA. The caller template's hourly
schedule also resolves the current integration head and performs a secondary
integration recovery wake; it is a mutation-free no-op after both required
integration workflows have completed successfully.
Immediate post-merge recovery is limited to governed `agent/` task branches;
other integration advances rely on that hourly exact-tip wake.

VOC-114 (VOC-113 recovery metadata-read fix) pins shared-infra merge
`c5d8bccfa8676bd367b53ad5f6f9a51a40c99405`, including the live-proof
corrections from PRs #137 through #145 and the project-template correction that
keeps Statuses write on release alone.

VOC-117 advances the six active role bindings to Cursor Composer 2.5
(implementer + escalation) and Grok 4.6 Standard with explicit bracket
parameters (planner, reviewer, reviewer_fast_retry, plan_reviewer). Workflow
routing uses `config/prepare_cursor_model.py` so stored values like
`cursor/grok-4.6[fast=false]` reach the Cursor CLI without silent vendor/model
fallback. Reviewer and plan-review terminal failures pass the structured Cursor
response through `config/extract-cursor-result.py`, which emits only bounded
reason codes and withholds raw provider output. `PINNED_SHA.txt` records exact
reviewed karsift-ai-infra merge
`42aa66757a521b1187193fba17b74e440964c27f`; fixture file content in this
directory is synchronized to that merge in the same task. Recovery metadata reads and allowlisted
dispatches use narrowly job-scoped `GITHUB_TOKEN` permissions: Actions write plus
Checks, Commit statuses, Contents, and Pull requests read. App tokens remain
limited to App-identity release mutations and no longer depend on an installation
Actions grant. The runner uses valid
`gh api` invocation, reports sanitized endpoint-class failures, binds promotion
dispatch suppression, completion, and promotion verification to required
contexts, ignoring unrelated checks when the three required contexts have
succeeded. Both hosted verification adapters
also use valid `gh api` repository context. The contract is covered by
`tests/test_voc114_actions_check_recovery.py`.
Release additionally grants its job token Commit statuses write. After genuine
exact-head checks from the three expected workflows pass, it publishes only the
three ruleset-required same-SHA status attestations. The attestations link to the
release run and are excluded from authoritative selection, so they cannot replace
the underlying Actions evidence; the App token remains mutation-only.
The caller template exposes the existing integration resolver/recovery pair to
operator dispatch without accepting a free-form target SHA, closing the
default-branch schedule bootstrap gap encountered during the live proof.
Its paginated commit query pipes slurped pages to standalone `jq`, avoiding the
GitHub CLI's invalid `--slurp` plus `--jq` combination.
