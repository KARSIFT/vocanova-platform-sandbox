# VOC-106-T01 — Controlled remediation evidence

gate_status: source-proof-complete

Package: `specs/changes/VOC-106-operator-owned-live-evidence-failures-can`
Change: `VOC-106`
Task: `VOC-106-T01`

## Controlled operator-owned failure

source_run_id: `32541413458`
source_run_url: `https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32541413458`
source_head_sha: `4e414cb90484d8fe5ddb22e6b1536d5a4a0ae741`
source_pipeline_conclusion: `success`
source_review_verdict: `FAIL`
merge_gate: `fail-closed`
remediation_decision_job: `success`
should_retry: `false`
implementer_job: `skipped`
implementation_attempt_consumed: `false`
operator_escalation_marker: `present`

The controlled source head changed only this evidence file to a mismatched
task/package identity. Independent review correctly failed that head. The
remediation ownership classifier read the unchanged operator contract from the
same head, emitted one sanitized escalation marker, and did not start the
general implementer.

## Ordinary-task regression

ordinary_retry_fixture: `passed`

VOC-106-TEST-04 remains the deterministic positive control: ordinary tasks
without an ownership contract retain `RETRY` for review FAIL and CI failure,
within the existing attempt cap. The exact-SHA T00 CI and independent review
passed that fixture before merge.

## Privacy boundary

This evidence contains only allowlisted workflow, PR, conclusion, and SHA
metadata. It contains no logs, artifacts, secrets, OAuth/session/cookie/token
material, environment values, user identifiers, or evidence payloads.
