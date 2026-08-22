# VOC-107 — Test Plan

## VOC-107-TEST-00 — Bundle range is integration-anchor..HEAD, not pre-model tip alone

- Covers: `VOC-107-AC-00`
- Preconditions: Fixture or workflow-policy assertion against the post-fix
  implement.yml / helper
- Procedure: Inspect bundle-create command or extracted helper inputs; assert the
  recorded integration-anchor SHA is the exclusive lower bound of the published
  bundle range and that soft-reset still targets the pre-model tip.
- Expected result: Bundle lineage and soft-reset tip remain distinct; soft-reset
  tip is not used as the sole bundle prerequisite.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-01 — Positive: attempt-2 bundle verifies in clean integration-only bare repo

- Covers: `VOC-107-AC-00`, `VOC-107-AC-01`, `VOC-107-AC-04`
- Preconditions: Disposable Git fixture with integration commit I, attempt-1
  commit(s), and attempt-2 commit(s) on the task branch
- Procedure: Build the integration-anchored bundle at attempt-2 HEAD. Initialize
  a clean bare repository; fetch only the integration ref at I; run
  `git bundle verify` and import the declared head.
- Expected result: Verify and import succeed; imported head equals attempt-2
  HEAD; no fetch of the prior task-PR tip is required.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-02 — Positive: rebase-derived remediation lineage still publishes

- Covers: `VOC-107-AC-01`, `VOC-107-AC-04`
- Preconditions: Fixture that rebases attempt-1 commits onto a newer integration
  tip (new SHAs), then adds attempt-2 commits
- Procedure: Build the bundle from the new integration-anchor tip through
  attempt-2 HEAD, including the locally rebased attempt-1 commits; verify/import
  in a clean bare repo containing only that new integration tip.
- Expected result: Verify and import succeed despite locally recreated bases;
  publication is not stranded by rebase-derived prerequisites.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-03 — Publisher guards preserved on successful import

- Covers: `VOC-107-AC-02`
- Preconditions: Successful import fixture from TEST-01 or TEST-02
- Procedure: Assert exact head SHA match; integration-anchor ancestry of the
  published head; workflow-path deny scan covers the full
  integration-anchor..HEAD range; force-with-lease still uses the expected SHA
  lease for attempt 2.
- Expected result: All listed publisher controls remain enforced.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-04 — Regression: attempt-1 fresh-from-integration still publishes

- Covers: `VOC-107-AC-00`, `VOC-107-AC-01`, `VOC-107-AC-04`
- Preconditions: Fixture where the task branch is created from integration and
  only attempt-1 commits exist
- Procedure: Build and verify/import the bundle in a clean integration-only bare
  repo using the same publisher rules.
- Expected result: Attempt-1 path continues to succeed (no remediation
  regression).
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-05 — Negative: incomplete / stale thin lineage fails closed

- Covers: `VOC-107-AC-02`, `VOC-107-AC-04`
- Preconditions: Bundle whose prerequisite is a task-PR head (or other object)
  absent from the clean integration-only bare repository — representing today’s
  failing thin `pre_model_tip..HEAD` class
- Procedure: Attempt `bundle verify` / import in the clean bare repo.
- Expected result: Verify/import fails closed; no forged publication.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-06 — Two-attempt policy unchanged; no model rerun for prerequisite miss

- Covers: `VOC-107-AC-03`
- Preconditions: Post-fix implement.yml / remediation wiring
- Procedure: Policy/fixture assertion that attempt inputs remain capped at 2 and
  that no new path dispatches a third implementer run solely because a committed
  bundle lacked a publisher prerequisite.
- Expected result: Cap preserved; lineage fix is the remediation for that class,
  not another model attempt.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-07 — Docs describe integration-anchored bundles

- Covers: `VOC-107-AC-05`
- Preconditions: T00 edits infra README; calling-repo docs only if needed
- Procedure: Assert README (and any touched calling-repo docs) describe
  integration-anchored complete task lineages for the clean publisher and do not
  claim remediating thin bundles are always publisher-sufficient.
- Expected result: Operator docs match behavior.
- Evidence: `VOC-107-EV-00`

## VOC-107-TEST-08 — Evidence privacy boundary

- Covers: `VOC-107-AC-06`
- Preconditions: `t00-evidence.md` written for this task
- Procedure: Review evidence contents against the allowlist in `VOC-107-D06`.
- Expected result: Only allowlisted metadata; no logs, artifacts, secrets, or
  user identifiers.
- Evidence: `VOC-107-EV-00`

Include positive, negative, rebase-derived, attempt-1 regression, publisher-guard,
attempt-cap, documentation, and privacy coverage as above. Tests must not use
secrets or production data.
