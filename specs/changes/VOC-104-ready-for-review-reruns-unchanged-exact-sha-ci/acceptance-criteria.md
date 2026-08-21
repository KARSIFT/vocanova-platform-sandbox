# VOC-104 — Acceptance Criteria

## VOC-104-AC-00 — Draft PRs remain non-mergeable

- Requirement source: `VOC-104-D00`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-00`, `VOC-104-TEST-01`
- Evidence: `VOC-104-EV-00`
- Result: pending

A draft PR with green required checks and a trusted App PASS still does not
auto-merge. Only after `ready_for_review` (or equivalent non-draft state) may
merge-gate proceed, and only when all other gates pass.

## VOC-104-AC-01 — Safe unchanged ready_for_review reuses prior exact-SHA evidence

- Requirement source: `VOC-104-D01`, `VOC-104-D02`, `VOC-104-D03`
- Tasks: `VOC-104-T00`, `VOC-104-T01`
- Tests: `VOC-104-TEST-02`, `VOC-104-TEST-08`
- Evidence: `VOC-104-EV-00`, `VOC-104-EV-01`
- Result: pending

On `ready_for_review`, when the live base/head pair is unchanged, required checks
for that exact head are successful, and a trusted App-authored PASS (or PASS WITH
NON-BLOCKING FINDINGS) is bound to that exact base/head and package/task
authority (with live-evidence attestation present when required), the pipeline
skips full CI and model review and still runs deterministic merge-gate
re-evaluation.

## VOC-104-AC-02 — Fail closed to the normal path when reuse is unsafe

- Requirement source: `VOC-104-D04`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-03`, `VOC-104-TEST-04`, `VOC-104-TEST-05`,
  `VOC-104-TEST-06`, `VOC-104-TEST-11A`
- Evidence: `VOC-104-EV-00`
- Result: pending

If the head or base changed, required checks are missing/non-successful, the
verdict is missing/WAITING/FAIL/PENDING/malformed/untrusted, live-evidence
attestation is required but absent, or identity/scope metadata does not match,
the run does not claim reuse. It takes the normal full CI and applicable review
path (fail closed toward verification, not toward merge).

## VOC-104-AC-03 — Human and implementer comments are never reusable authority

- Requirement source: `VOC-104-D05`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-07`
- Evidence: `VOC-104-EV-00`
- Result: pending

Only the existing App-signed independent verification publisher comment shape
qualifies for reuse. Human comments, implementer comments, and other bot text do
not.

## VOC-104-AC-04 — Exact-SHA stale-run and role-separation protections preserved

- Requirement source: `VOC-104-D06`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-03`, `VOC-104-TEST-10`
- Evidence: `VOC-104-EV-00`
- Result: pending

Stale expected base/head pairs still refuse reuse. `opened` / `synchronize` /
`reopened` still run full CI and review. Independent implementer/reviewer
separation is unchanged; the implementer cannot mint reusable PASS authority.

## VOC-104-AC-05 — Deterministic shared-infra and calling-repo fixture coverage landed

- Requirement source: `VOC-104-D07`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-00` through `VOC-104-TEST-07`, `VOC-104-TEST-10`,
  `VOC-104-TEST-11`, `VOC-104-TEST-11A`
- Evidence: `VOC-104-EV-00`
- Result: pending

Shared-infra policy tests and calling-repository fixture/foundation coverage exist
and pass for positive reuse, deterministic negative `full-path`, uncertain
`fail-closed-to-full-path`, draft non-merge, authority rejection, attestation
absence, and non-ready_for_review regression cases.

## VOC-104-AC-06 — Controlled draft-to-ready optimized-path proof

- Requirement source: `VOC-104-D08`
- Tasks: `VOC-104-T00`, `VOC-104-T01`
- Tests: `VOC-104-TEST-08`, `VOC-104-TEST-09`, `VOC-104-TEST-12`
- Evidence: `VOC-104-EV-00`, `VOC-104-EV-01`
- Result: pending

After T00 is live, a controlled draft→ready transition on an unchanged exact SHA
shows CI and model review skipped and merge-gate re-evaluated successfully.
Evidence is metadata-only (run IDs, job conclusions, SHAs, reuse boolean). No
logs or secrets. T00 contributes the read-only verifier; T01 closes the live
proof under the operator-owned contract.

## VOC-104-AC-07 — Docs match reuse-vs-full-path behavior

- Requirement source: AGENTS.md doc-consistency rule; `VOC-104-D03`
- Tasks: `VOC-104-T00`
- Tests: `VOC-104-TEST-11`
- Evidence: `VOC-104-EV-00`
- Result: pending

Infra README and calling-repo
`docs/operations/15-ai-native-product-and-engineering-operating-model.md` §17.3
must accurately state that a fresh `ready_for_review` pipeline evaluation may
reuse prior exact-SHA CI and App PASS evidence when D02 holds, otherwise it runs
the full path. Any other touched ops text must use the same distinction.
