# VOC-105 — Test Plan

## VOC-105-TEST-00 — Ready-for-review reuse classifier invoked before full CI/review

- Covers: `VOC-105-AC-00`
- Preconditions: T00 reuse gate present
- Procedure: Fixture `ready_for_review` event with complete prior exact-SHA
  evidence; invoke classifier / pipeline decision helper; assert the decision
  runs before any full CI or model-review invocation.
- Expected result: Deterministic reuse decision is evaluated first.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-01 — Positive: unchanged base/head with trusted PASS reuses evidence

- Covers: `VOC-105-AC-01`
- Preconditions: Fixture where base/head unchanged, required checks successful
  for that exact head, trusted App-authored PASS (or PASS WITH NON-BLOCKING
  FINDINGS) bound to package/task authority; live-evidence attestation present
  if required by the prior review lifecycle
- Procedure: Run reuse decision fixture.
- Expected result: reuse permitted; full CI and model review not requested;
  merge-gate path still selected for non-draft PR.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-02 — Negative: changed head or base refuses reuse

- Covers: `VOC-105-AC-02`
- Preconditions: Fixture where live head or base differs from prior evidence
  pair
- Procedure: Run decision fixture for head change and base change.
- Expected result: reuse refused; normal full CI + review path selected.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-03 — Negative: missing or non-successful required checks refuse reuse

- Covers: `VOC-105-AC-02`
- Preconditions: Fixture with missing check, failing check, or pending check
  for the exact head
- Procedure: Run decision fixture for each case.
- Expected result: reuse refused; normal full path selected.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-04 — Negative: missing, waiting, failing, malformed, or untrusted verdict refuses reuse

- Covers: `VOC-105-AC-02`
- Preconditions: Fixtures covering absent verdict, WAITING, FAIL, malformed
  comment shape, and non-App author
- Procedure: Run decision fixture for each case.
- Expected result: reuse refused in every case.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-05 — Negative: human/implementer comment is not reusable authority

- Covers: `VOC-105-AC-03`
- Preconditions: Fixture with a human or implementer comment that mimics PASS
  verdict shape, and no trusted App-signed PASS
- Procedure: Run decision fixture.
- Expected result: reuse refused.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-06 — Negative: live-evidence attestation required but absent refuses reuse

- Covers: `VOC-105-AC-02`
- Preconditions: Fixture where prior review lifecycle required live-evidence
  attestation and attestation is absent or invalid for the exact head
- Procedure: Run decision fixture.
- Expected result: reuse refused; normal full path selected.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-07 — Draft PRs never auto-merge even when reuse would otherwise be safe

- Covers: `VOC-105-AC-04`
- Preconditions: Draft PR fixture with otherwise-complete exact-SHA evidence
- Procedure: Run merge-gate / reuse interaction fixture.
- Expected result: draft remains non-mergeable; marking ready is still required
  before merge.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-08 — Controlled live proof: ready_for_review skips full CI and model review

- Covers: `VOC-105-AC-01`, `VOC-105-AC-06`
- Preconditions: T00 live; controlled draft PR with green exact-SHA CI and
  trusted App PASS; base/head unchanged at ready
- Procedure: Observe sanitized ready_for_review pipeline metadata; confirm full
  CI and model review jobs did not execute; confirm merge-gate re-evaluated.
  After carrier metadata is recorded, dispatch the read-only proof action on the
  exact carrier ref and reconcile under the T01 contract.
- Expected result: Optimized path observed; verifier succeeds at
  `exact_pr_head`; allowlisted run/job metadata only.
- Evidence: `VOC-105-EV-01` (metadata only)

## VOC-105-TEST-09 — Synchronize / opened / reopened still take the full path

- Covers: `VOC-105-AC-04`
- Preconditions: Fixtures for `synchronize`, `opened`, and `reopened` with
  otherwise-reusable prior evidence on the same SHA if applicable
- Procedure: Run decision fixture for each action.
- Expected result: full CI + review path selected; reuse path is
  ready_for_review-only.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-10 — Doc and caller-wiring consistency

- Covers: `VOC-105-AC-05`, `VOC-105-AC-07`
- Preconditions: T00 may edit infra README / pipeline.yml / foundation tests
- Procedure: Assert docs describe safe reuse and fail-closed normal path; assert
  foundation fixtures lock caller wiring for the optimized ready_for_review
  path; assert out-of-scope roots from issue #872 are not claimed fixed.
- Expected result: No false doc claim; caller wiring covered.
- Evidence: `VOC-105-EV-00`

## VOC-105-TEST-11 — Read-only verifier rejects stale head, wrong event, or executed full CI/review

- Covers: `VOC-105-AC-06`
- Preconditions: T00 verifier job present
- Procedure: Fixture verifier inputs with stale carrier head, wrong source
  event, or a ready_for_review run that executed full CI/review contrary to the
  reuse claim
- Expected result: verifier fails closed; does not read logs/artifacts; no write
  or model credentials required.
- Evidence: `VOC-105-EV-00` (and live confirmation in `VOC-105-EV-01`)

Include positive, negative, authorization, failure, and rollback-adjacent
coverage as above. Tests must not use secrets or production data.
