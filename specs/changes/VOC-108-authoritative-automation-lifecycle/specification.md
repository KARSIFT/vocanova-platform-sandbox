# VOC-108 — Specification

## Objective

Make lifecycle decisions depend on one authoritative, exact-revision state
model. A historical check, foreign-repository close event, concurrent observer,
or late retry must not advance or contradict the caller repository's current
task and release state.

This draft is grounded in issue #903 and current `karsift-ai-infra` behavior:
`adopt.yml` and `release.yml` aggregate raw check-run history; package completion
uses roster issue state; `merge-gate.yml` closes the caller task after merge;
and release promotion duplicates auto/retry polling and merge sequences.

## Decisions

`VOC-108-D00`: For each required logical check name, automation MUST select one
latest authoritative run using deterministic ordering (terminal attempt/run
identity and GitHub ID as a tie-breaker) after binding candidates to repository,
workflow, PR, base SHA, and head SHA. Superseded runs MUST be ignored. Missing,
ambiguous, pending, cancelled, timed-out, or latest-failed evidence MUST fail
closed. Pagination and more-than-100 check histories MUST be handled.

`VOC-108-D01`: The selection rule MUST be shared by adoption, merge/reuse, and
release consumers where they answer the same question. A later success supersedes
an earlier failure; a later failure or pending result supersedes an earlier
success. Consumers MUST NOT use an undifferentiated `all(check_runs)` result.

`VOC-108-D02`: Cross-repository PRs and evidence MUST use non-closing references
such as `Relates to owner/repository#N`. A closing keyword for a task is permitted
only on the caller repository's own task implementation PR and MUST be rejected
or normalized when the target repository differs from the authority repository.

`VOC-108-D03`: Task completion MUST have an immutable App-authored marker bound
to caller repository, authority issue, package path, task ID, merged caller PR,
and exact reviewed head SHA. The marker is written only after the caller PR is
observed merged at that exact head. Closure alone, issue title alone, foreign PR
timeline events, or a free-form comment is insufficient.

`VOC-108-D04`: Auto-advance and release MUST validate the same task-completion
marker for every roster entry. A premature close becomes a fail-closed no-op; it
MUST NOT dispatch the next task or create a release audit. Reconciliation after
the legitimate caller merge MUST be idempotent and able to recover.

`VOC-108-D05`: Promotion MUST expose a single idempotent final merge operation.
Automatic and reconcile triggers may request evaluation, but they MUST converge
on the same serialized decision keyed by repository and promotion PR/release
identity. Before commenting or merging, the evaluator MUST re-read PR state and
exact head; a merged/closed PR is terminal success/no-op and receives no later
pending comment.

`VOC-108-D06`: Completion of an external prerequisite such as staging deployment
MUST trigger or wake a cheap release re-evaluation. It MUST reuse exact-SHA CI and
independent-review evidence under the authoritative selection rule and MUST NOT
dispatch a full unchanged-SHA pipeline solely to make release notice the result.
Duplicate completion events MUST be harmless.

`VOC-108-D07`: Deterministic fixtures MUST cover: historical failure then later
success; historical success then later failure/pending; duplicate named runs;
pagination; wrong repo/PR/base/head; cross-repo closing text; premature issue
closure; legitimate caller merge proof; external check completion race;
concurrent auto/reconcile requests; merged-before-comment; and exactly-once
effective merge.

`VOC-108-D08`: Preserve App authentication, branch protection, exact-SHA review,
base/head rechecks, fail-closed unknown risk, SHA-valued merge protection,
two-attempt implementation policy, task ordering, and no founder-comment gate.
Evidence is metadata-only and MUST exclude logs, credentials, sessions, OAuth
material, user identifiers, and repository secrets.

## Scope boundaries

This package is R3 governance automation. It does not alter application code,
runtime deployment configuration, databases, Google OAuth policy, Kuma/Sentry
inventory, or secret values. Scheduled-synthetic branch selection, operational
failure marker cleanup, Node action upgrades, and historical stranded issue
cleanup require separate roots.
