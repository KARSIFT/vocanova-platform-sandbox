# VOC-105 — Acceptance Criteria

## VOC-105-AC-00 — Ready-for-review reuse decision runs before full CI or model review

- Requirement source: `VOC-105-D00`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-00`
- Evidence: `VOC-105-EV-00`
- Result: pending

On `pull_request` action `ready_for_review`, the governed pipeline evaluates
whether prior exact-SHA CI and independent-review evidence may be reused before
starting full CI or model review. The decision is deterministic and fail
closed.

## VOC-105-AC-01 — Safe unchanged case skips full CI and model review

- Requirement source: `VOC-105-D01`, `VOC-105-D02`
- Tasks: `VOC-105-T00`, `VOC-105-T01`
- Tests: `VOC-105-TEST-01`, `VOC-105-TEST-08`
- Evidence: `VOC-105-EV-00`, `VOC-105-EV-01`
- Result: pending

When every `VOC-105-D01` precondition holds for the current base/head pair,
full CI and model review / plan-review do not run. Deterministic merge-gate
still re-evaluates so a non-draft PR can merge. Draft PRs remain non-mergeable.

## VOC-105-AC-02 — Unsafe cases fail closed to the normal full path

- Requirement source: `VOC-105-D01`, `VOC-105-D03`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-02` through `VOC-105-TEST-06`
- Evidence: `VOC-105-EV-00`
- Result: pending

If the head or base changed, required checks are missing or non-successful, the
verdict is missing/waiting/failing/malformed/untrusted, required live-evidence
attestation is absent, or identity/scope metadata does not match, the pipeline
runs the normal full CI + applicable independent review + merge-gate path.

## VOC-105-AC-03 — Human and implementer comments are never reusable authority

- Requirement source: `VOC-105-D01`, `VOC-105-D05`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-05`
- Evidence: `VOC-105-EV-00`
- Result: pending

A human, implementer, or non-App comment shaped like a PASS verdict does not
authorize reuse. Only a trusted App-authored independent PASS (or PASS WITH
NON-BLOCKING FINDINGS) bound to the exact SHA and package/task authority
qualifies.

## VOC-105-AC-04 — Exact-SHA stale-run and draft-never-merge protections preserved

- Requirement source: `VOC-105-D02`, `VOC-105-D04`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-07`, `VOC-105-TEST-09`
- Evidence: `VOC-105-EV-00`
- Result: pending

Mismatched or newer base/head pairs invalidate reuse. Draft PRs never
auto-merge. Synchronize / opened / reopened events still run full CI and
review.

## VOC-105-AC-05 — Deterministic test matrix landed

- Requirement source: `VOC-105-D06`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-00` through `VOC-105-TEST-07`, `VOC-105-TEST-09`,
  `VOC-105-TEST-10`
- Evidence: `VOC-105-EV-00`
- Result: pending

Positive reuse, negative refuse-and-rerun, human-comment rejection,
draft-never-merge, synchronize regression, and caller-wiring coverage exist
and pass in CI or infra self-ci.

## VOC-105-AC-06 — Controlled draft-to-ready live proof

- Requirement source: `VOC-105-D07`
- Tasks: `VOC-105-T00`, `VOC-105-T01`
- Tests: `VOC-105-TEST-08`, `VOC-105-TEST-11`
- Evidence: `VOC-105-EV-00`, `VOC-105-EV-01`
- Result: pending

After T00 is live, a controlled draft → ready transition with unchanged
base/head proves the optimized path: prior green exact-SHA evidence is reused,
full CI and model review are skipped on the ready_for_review run, merge-gate
re-evaluates, and allowlisted run/job metadata (no logs or secrets) is
recorded. The read-only verifier succeeds at `exact_pr_head`.

## VOC-105-AC-07 — Doc consistency when docs touched

- Requirement source: `VOC-105-D00`, `VOC-105-D08`
- Tasks: `VOC-105-T00`
- Tests: `VOC-105-TEST-10`
- Evidence: `VOC-105-EV-00`
- Result: pending

Docs touched by T00 describe the safe reuse path and fail-closed normal path.
Untouched docs do not falsely claim ready_for_review always re-runs full CI
and model review. Out-of-scope items from issue #872 remain excluded.
