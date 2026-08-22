# VOC-108 — Acceptance Criteria

## VOC-108-AC-00 — Latest authoritative exact-SHA checks

Adoption, merge/reuse, and release consumers select one newest authoritative
attempt per logical gate, bound to repository, workflow, PR, base SHA, and head
SHA. Earlier failures do not poison a later pass, and earlier passes do not hide
a later failure or pending result. Missing, ambiguous, incomplete, truncated, or
untrusted evidence fails closed.

## VOC-108-AC-01 — Repository-safe task references

Generated cross-repository PR/evidence text contains no `close`, `closes`,
`closed`, `fix`, `fixes`, `fixed`, `resolve`, `resolves`, or `resolved` keyword
that targets a caller issue. The caller implementation PR retains its expected
local task binding.

## VOC-108-AC-02 — Caller-merge-bound completion

A roster issue's CLOSED state alone cannot advance the package. Completion
requires exactly one valid App-authored marker matching the roster task and a
merged caller PR at the recorded reviewed head. Forged, mismatched, duplicate,
foreign-repository, closed-unmerged, or stale-head records fail closed.

## VOC-108-AC-03 — Safe task and release advancement

Auto-advance and release use the same completion validator. Premature closure
does not dispatch a successor or open a release audit. After valid caller merge
evidence exists, reconciliation advances once and remains idempotent.

## VOC-108-AC-04 — One promotion merge decision

All automatic and reconcile release triggers converge on one serialized,
exact-head decision and at most one effective merge. A late/stale invocation
observing an already merged PR exits successfully without adding a contradictory
pending comment or retrying the merge.

## VOC-108-AC-05 — Event-driven cheap re-evaluation

Terminal completion of a required external check causes release eligibility to
be re-evaluated using prior exact-SHA CI/review evidence; no full unchanged-SHA
CI or reviewer model run is required merely to observe that completion.

## VOC-108-AC-06 — Controls and documentation preserved

Exact-SHA/base binding, App identity, branch protection, fail-closed gates,
task ordering, publisher isolation, attempt caps, and no founder-comment approval
remain intact. Documentation describes the authoritative-state and recovery
contract accurately.

## VOC-108-AC-07 — Sanitized evidence

Task evidence records only repository/PR/issue/run identifiers, commit SHAs,
dates, test commands/results, and scrubbed outcomes. It contains no logs,
credentials, tokens, sessions, OAuth material, secrets, or user identifiers.
