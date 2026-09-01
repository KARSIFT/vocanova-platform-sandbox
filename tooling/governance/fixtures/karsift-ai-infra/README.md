# Pinned karsift-ai-infra contract fixtures (VOC-080-T05, VOC-097-T03, VOC-102-T00, VOC-104, VOC-106, VOC-108, VOC-115, VOC-117, VOC-121, VOC-122, VOC-123, VOC-126, VOC-129, VOC-136, VOC-140)

These copies are deterministic fixtures for caller-repo policy regressions.
This README is adapted caller-local provenance documentation, not a canonical
byte mirror of the infrastructure repository's README. The workflow, config,
and test paths locked by the governance hash tables mirror
`KARSIFT/karsift-ai-infra` at the SHA in `PINNED_SHA.txt` so
`tooling/governance/tests/test_voc080_*.py`,
`tooling/governance/tests/test_voc097_*.py`,
`tooling/governance/tests/test_voc115_package_task_policy.py`,
`tooling/governance/tests/test_auto_advance_ownership.py`,
`tooling/governance/tests/test_remediate_ownership.py`, and related suites can
assert merge/adopt/release/remediate/plan-review/live-evidence/auto-advance/role
contracts without cloning the infra repository in CI.

They are not a second runtime source of truth. Callers still `uses:`
`KARSIFT/karsift-ai-infra/...@main`. Update the fixtures when VOC-080-,
VOC-097-, VOC-102-, VOC-104-, VOC-106-, VOC-108-, VOC-115-, VOC-121-, VOC-122-, VOC-123-, VOC-126-, VOC-129-, VOC-136-, or VOC-140-related infra contracts change and
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
The final VOC-140 reuse repair identifies the canonical pipeline from immutable
workflow path, pull-request event, exact head/base/refs, policy revision, and job
set; a PR-title-derived run display name is not identity. Every non-empty
`pull_requests` payload must be one exact association. Null, scalar, partial,
mixed, duplicate, or unrelated-extra entries fail closed, and every supplied
repository `full_name`, name, owner, API URL, or web URL must agree. GitHub's
real compact name-plus-API-URL repository shape remains eligible. Only an exact
empty association array may use the existing unique App-authored attestation
fallback. For an ordinary non-production agent/plan PR, already-terminal required
checks may be admitted while their exact pipeline parent waits on merge-gate;
promotion/release carriers and any run with an executed non-skipped release job
retain completed-success parent attestation.
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

VOC-125 (VOC-126 caller pin) documents operator resume through caller
`workflow_dispatch` `action=implement` with `attempt=2` and
`existing_pr_number=<open PR>`. The caller forwards only the PR number—not
free-form SHAs. `implement.yml` derives immutable `expected_head_sha` /
`expected_base_sha` before `Create implementation branch`. Read-only verifier
dispatch uses caller `pipeline-verify.yml` (VOC-126) so `pipeline.yml` stays
within GitHub's 25-input `workflow_dispatch` maximum while preserving
`existing_pr_number`.

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
exact-identity newest-attempt selector as the merged plan head. Roster wait
requires the complete ruleset-required logical-context set for the exact head
(including `ci / ci`, `governance-policy`, and `validate`) to be registered and
SUCCESS; two stable zero-pending subset snapshots are not complete. Reconcile
reuses a matching open or already-merged roster carrier instead of always
calling `gh pr create`. The complete required set must stabilize on one
unchanged head before exact-SHA merge.
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
(implementer + escalation) and explicit-high Grok 4.6 Standard with bracket
parameters (planner, reviewer, reviewer_fast_retry, plan_reviewer). Live Cursor
CLI model discovery proved the effort-omitted identifier unavailable, so
`config/prepare_cursor_model.py` passes stored values like
`cursor/grok-4.6[effort=high,fast=false]` to the Cursor CLI without silent
vendor/model fallback and rejects Grok 4.6 forms that omit effort. Reviewer and
plan-review terminal failures pass the structured Cursor
response through `config/extract-cursor-result.py`, which writes only a strict
schema-v1 failure artifact containing an allowlisted reason and regex-bounded
subtype. The same-run artifact is retained for one day. Dedicated clean
publisher jobs download and strictly validate it, validate the exact live PR
base/head pair, check out their own exact reusable-workflow SHA, mint a narrowly
scoped App token only after validation, and publish a non-verdict infrastructure
failure comment while withholding raw provider output. When Cursor exits with
an empty JSON response, the failed producer may inspect at most 64 KiB of local
stderr and retain only an existing allowlisted reason code; missing, oversized,
or unrecognized text remains `unspecified` and never enters the artifact. The
exact Cursor phrase `API key is invalid` maps to `authentication`, while
negative regressions reject unrelated API-key help prose. Cursor's bounded
`Available models:` diagnostic is classified as an unavailable/invalid model
without publishing the list or other raw output.
`PINNED_SHA.txt` records the exact reviewed karsift-ai-infra merge consumed by
this fixture. VOC-117 originally advanced the fixture to merge
`37b06aa95030e235b7311b3c14ee23977f62ac76`. VOC-121-T00 advances it to
infrastructure PR #157 merge
`99476c2a1018e42d4bd442657b5257885ac9f1c9`. VOC-123-T00 advances it to
infrastructure PR #158 merge
`7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`. VOC-124-T00 advances it to
infrastructure PR #159 merge
`f406cc95a3f853e8aef5bf8bcf22d37a29d64547`. VOC-126-T00 advances it to
infrastructure PR #161 merge
`20dcf340fa73a36ebc6074442fde79530dfa5871`. VOC-122-T00 advances it to
infrastructure PR #162 merge
`60afda3a44fd06b8c00b219771de7112f1aded6e`. VOC-129-T00 advances it to
infrastructure PR #164 merge
`863fc1f35b1d35e4981a59166b0e939be1a2b681`; the mirrored source files are
byte-for-byte copies from that merge's reviewed tree. VOC-136-T00 advances it to
infrastructure PR #167 merge
`b263c0c110591cc798b89277dfc35542abb1597b`; the mirrored source files include
the #165 post-caller-checkout restore in both release jobs, the #166
post-implementer helper-lifetime and nested-checkout contracts, and the #167
immutable PR-context validation contract. VOC-138-T00 advances the pin to
infrastructure merge `123735c80fec813a5b46a004f3e1122bd425cde2` with authenticated
promotion `pr-validation`, recovery dispatch metadata, and doomed `ci / ci` rerun
refusal. `run-app-checks.sh` validates an exact
PR base/head pair without fetching evidence; an unchanged capture fixture
selects `pr-validation`, add/modify/delete selects `pr-ancestry`, reusable CI
checks out full reachable history (`fetch-depth: 0`), and implementer pre-push
uses the integration anchor plus live committed HEAD including after
self-correction. VOC-138-T00 advances the pin to the infrastructure merge that
adds an authenticated same-repository `main` <- `develop` promotion signal:
that pair always selects `pr-validation` with exact PR SHAs regardless of capture
subject availability; ordinary fixture-changing PRs remain `pr-ancestry`
fail-closed; promotion recovery dispatches `recover-promotion-pr-checks` with
immutable PR metadata instead of rerunning doomed `ci / ci` pull-request jobs,
and rejects weaker same-head `squash-safe-push` dispatches as completion proof.
VOC-139-T00 advances the pin to infrastructure merge
`599436835371f27fac52ec6b47a18b36257366ac` with promotion head/source-revision
hash binding under `pr-validation`, repository-explicit no-checkout recovery
metadata via `gh api repos/$GITHUB_REPOSITORY/pulls/...`, and ordinary
merge-base-anchored `pr-validation` for non-promotion PRs. VOC-140-T00 initially
consumed independently reviewed infrastructure merge
`9fdff24cd387cc2cdc468c84a3012b0c34b6c8e8` with
release-carrier CI identity repair, dedicated `promotion-pr-validation` completion
requirements, and a strict two-token production merge-guard contract: the mutation
App token remains exactly Contents/Issues/Pull requests write for merge and
mutations, while a separate caller-repository-scoped Administration-write guard
token is used only for `verify-production-merge-guard.sh` immediately before merge.
An in-progress or otherwise non-successfully-completed parent is never attestable,
and any actually executed non-skipped `release / converge` job marks its run as a
carrier regardless of title. A skipped release definition remains eligible; when no
completed non-carrier result exists, exact-PR/head `promotion-pr-validation` must
complete successfully before attestation. The final independently reviewed
in-scope repair advances the current pin to
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` and adds the ordinary-PR
self-deadlock exception plus the shared strict association and custom-run-name
reuse rules described above. VOC-142-T00 advances the current pin to
`8993e867640dfb604dec0466c4e0787e68d8e258` with roster wait completeness for
the full ruleset-required set including `ci / ci` and exact open or
already-merged roster PR carrier reuse on reconcile.

Recovery metadata reads, exact selected-run reruns, and allowlisted absent-context
dispatches use narrowly job-scoped `GITHUB_TOKEN` permissions: Actions write plus
Checks, Commit statuses, Contents, and Pull requests read. App tokens remain
limited to App-identity release mutations and no longer depend on an installation
Actions grant. The runner uses valid
`gh api` invocation, reports sanitized endpoint-class failures, binds promotion
dispatch suppression, completion, and promotion verification to required
contexts. A failed or cancelled ruleset-selected pull-request check is rebound
to its original run, PR, head, branch, event, workflow name/path, and first
attempt before that exact run is rerun once, except doomed `ci / ci` promotion
rows dispatch `recover-promotion-pr-checks` instead of rerunning. Alternate
successful runs and same-named statuses cannot suppress the rerun; workflow
dispatch is used only for a genuinely absent required context or promotion
`ci / ci` recovery. Both hosted verification adapters
also use valid `gh api` repository context. The contract is covered by
`tests/test_voc114_actions_check_recovery.py`.
Release additionally grants its job token Commit statuses write. After genuine
exact-head checks from the three expected workflows pass, it publishes only the
three ruleset-required same-SHA status attestations. The attestations link to the
release run and are excluded from authoritative selection, so they cannot replace
the underlying Actions evidence; the mutation App token remains exactly
Contents/Issues/Pull requests write for merge and mutations, while a separate
caller-repository-scoped Administration-write guard token is used only for
`verify-production-merge-guard.sh` immediately before merge.
The caller template exposes the existing integration resolver/recovery pair to
operator dispatch without accepting a free-form target SHA, closing the
default-branch schedule bootstrap gap encountered during the live proof.
Its paginated commit query pipes slurped pages to standalone `jq`, avoiding the
GitHub CLI's invalid `--slurp` plus `--jq` combination.

VOC-121 (coordinated infrastructure carrier publication and required-check
recovery) advances implementer helper preservation, isolated `publish-source`
for authorized nested `karsift-ai-infra/` edits, and promotion recovery that
consults `gh pr checks --required`. A cancelled or failed exact-head required
Actions run is rerun by its selected run ID after exact metadata validation;
only an absent context is redispatched. Alternate successful runs and
same-named statuses cannot override the ruleset-selected row. Status
attestation, release evaluation, and the read-only recovery verifier use the
same required-check view and fail closed on ambiguity or read failure.

VOC-122 (promotion-recovery replan during polling) re-evaluates GitHub's
required PR view during the bounded 1800-second wait, not only from the first
snapshot. A context that is absent in snapshot 1 and later appears as a
cancelled or failed ruleset-selected pull-request row is rerun once in the same
invocation after the existing identity checks. Invocation-scoped run-ID and
absent-context dedupe prevent duplicate reruns or dispatches.

VOC-123 (named-ref nested source-carrier bundle tips) binds the exact committed
infrastructure head to one temporary `refs/karsift/source-bundle-head` ref
before `bundle create`, verifies the sole advertised head, and removes the ref
on every exit path. Raw object IDs remain a proven empty-bundle failure class.
Caller and planner `..HEAD` recovery bundles were proven safe with real
repositories and remain unchanged.

VOC-124 (coordinated source publisher workflow-write permission) requests
`workflows: write` on the clean `publish-source` App token so authorized nested
infrastructure commits that change `.github/workflows/**` can be published.
The caller `publish` token still omits `workflows: write` and still rejects every caller `.github/workflows/**` change before push.
