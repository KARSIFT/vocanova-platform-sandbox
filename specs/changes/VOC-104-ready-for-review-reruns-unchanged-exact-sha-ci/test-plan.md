# VOC-104 — Test Plan

## VOC-104-TEST-00 — Draft PR with green evidence still does not auto-merge

- Covers: `VOC-104-AC-00`
- Preconditions: Fixture PR `isDraft=true` with successful required checks and
  trusted App PASS on exact base/head
- Procedure: Run merge-gate / reuse decision fixtures as draft.
- Expected result: Auto-merge path does not fire; decision remains draft-blocked.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-01 — ready_for_review is still subscribed so non-draft re-entry exists

- Covers: `VOC-104-AC-00`, `VOC-104-AC-05`
- Preconditions: Caller pipeline template and calling-repo `pipeline.yml`
- Procedure: Assert PR types include `ready_for_review` and merge-gate still
  rechecks `isDraft` immediately before merge.
- Expected result: Draft-aware re-entry remains; draft blocking remains.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-02 — Positive: unchanged exact SHA reuses evidence and skips CI/model review

- Covers: `VOC-104-AC-01`
- Preconditions: Fixture ready_for_review event; live base/head equal expected
  pair; required checks SUCCESS for that head; trusted App PASS bound to exact
  base/head and package/task identity; attestation not required or true
- Procedure: Run reuse-eligibility helper / pipeline condition fixture.
- Expected result: outcome `reuse-evidence`; CI and model review skipped;
  merge-gate still selected to run.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-03 — Negative: base or head drift forces full path

- Covers: `VOC-104-AC-02`, `VOC-104-AC-04`
- Preconditions: ready_for_review fixture where live head or base differs from
  expected pair
- Procedure: Run reuse decision fixture.
- Expected result: Outcome `full-path`; full CI + review path selected.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-04 — Negative: missing or non-successful required checks force full path

- Covers: `VOC-104-AC-02`
- Preconditions: Unchanged base/head but a required check is missing, pending,
  failed, or otherwise non-successful
- Procedure: Run reuse decision fixture for each case.
- Expected result: Outcome `full-path`; full CI + review path selected.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-05 — Negative: missing / WAITING / FAIL / PENDING / malformed / untrusted verdict

- Covers: `VOC-104-AC-02`
- Preconditions: Fixtures for each bad verdict class, including wrong App
  identity and wrong base/head binding
- Procedure: Run reuse decision fixture for each case.
- Expected result: Outcome `full-path`; full CI + review path selected.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-06 — Negative: required live-evidence attestation absent

- Covers: `VOC-104-AC-02`
- Preconditions: Fixture where attestation is required and absent/false
- Procedure: Run reuse decision fixture.
- Expected result: Outcome `full-path`; full CI + review path selected.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-07 — Human and implementer comments are rejected as reusable authority

- Covers: `VOC-104-AC-03`
- Preconditions: Fixture PR with a human or implementer PASS-looking comment and
  no qualifying App-signed publisher comment (or with both)
- Procedure: Run reuse decision fixture.
- Expected result: Human/implementer text never authorizes reuse; only App
  publisher shape qualifies.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-08 — Controlled workflow: ready_for_review skips CI/review on unchanged SHA

- Covers: `VOC-104-AC-01`, `VOC-104-AC-06`
- Preconditions: T00 live; controlled draft PR with prior green CI + App PASS on
  unchanged base/head
- Procedure: Mark ready; observe ready_for_review run metadata; after metadata is
  recorded, dispatch the read-only proof action on the exact PR head and
  reconcile under the T01 contract.
- Expected result: CI and model review jobs skipped; merge-gate runs; evidence
  metadata-only. Verifier succeeds at `exact_pr_head`.
- Evidence: `VOC-104-EV-01` (metadata only)

## VOC-104-TEST-09 — Controlled or fixture proof: unsafe ready_for_review still takes full path

- Covers: `VOC-104-AC-02`, `VOC-104-AC-06`
- Preconditions: T00 live; fixture or sanitized observation where a D02
  precondition fails at ready_for_review
- Procedure: Confirm full CI + review path is selected (deterministic fixture
  preferred).
- Expected result: No false reuse claim.
- Evidence: `VOC-104-EV-01` and/or `VOC-104-EV-00` if fully covered by TEST-03–06
  — record which

## VOC-104-TEST-10 — Non-ready_for_review activities still run full CI and review

- Covers: `VOC-104-AC-04`, `VOC-104-AC-05`
- Preconditions: Fixtures for `opened`, `synchronize`, and `reopened` with
  otherwise reuse-eligible evidence
- Procedure: Run reuse / pipeline condition fixtures.
- Expected result: Outcome `full-path`; full CI + review path selected because
  reuse applies only to `ready_for_review`.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-11 — Doc consistency and caller wiring fixture

- Covers: `VOC-104-AC-05`, `VOC-104-AC-07`
- Preconditions: T00 must edit the infra README and calling-repo DOC-15 §17.3;
  caller `pipeline.yml` is in scope.
- Procedure: Assert caller wires reuse decision into CI/review/merge-gate
  conditions. Assert both mandatory documentation files are in the task diff and
  distinguish `reuse-evidence` from the normal and fail-closed full paths.
- Expected result: Both mandatory docs are updated consistently; caller wiring
  is present.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-11A — Eligibility evaluation uncertainty fails closed

- Covers: `VOC-104-AC-02`, `VOC-104-AC-05`
- Preconditions: Fixtures for API failure, malformed metadata response, helper
  execution failure, and an unknown/missing machine outcome.
- Procedure: Run the decision/helper and caller-condition fixtures for each
  uncertain evaluation case.
- Expected result: The helper emits `fail-closed-to-full-path` when it can; the
  caller treats that value and any unknown/missing outcome as full CI + review,
  never as reusable evidence.
- Evidence: `VOC-104-EV-00`

## VOC-104-TEST-12 — Post-transition verifier is exact-head and fail-closed

- Covers: `VOC-104-AC-05`, `VOC-104-AC-06`
- Preconditions: Source-run fixtures including wrong event/action, executed CI or
  model-review job on the ready_for_review run, failed merge-gate, drifted SHA,
  and exact-head success cases
- Procedure: Exercise the verifier metadata adapter without logs or artifacts;
  assert the caller job is read-only and the T01 contract names the deterministic
  proof branch, `workflow_dispatch`, and `exact_pr_head`.
- Expected result: Only a matching ready_for_review reuse run plus current proof
  head succeeds. Every mismatch fails closed with a sanitized reason; the
  verifier has no mutation, model, deploy, or application-secret capability.
- Evidence: `VOC-104-EV-00`

Include positive, negative, authorization, attestation, regression, and verifier
lineage coverage as above. Tests must not use secrets or production data beyond
public Actions metadata and issue/PR fields already governed as sanitized.
