# VOC-114 — Test Plan

## VOC-114-TEST-00 — Evidence documents issue #956 and verified token contract

- Covers: `VOC-114-AC-00`
- Preconditions: T00 evidence drafted at implementation time
- Procedure: Read `t00-evidence.md`; assert it names PR #954 merge SHA, pipeline
  runs 32696249484 and 32696549963, immediate pre-dispatch failure, blocked PR
  #947, and the verified job-token/App-token separation (not a guess).
- Expected result: Diagnosis is metadata-only and bounds remediation to read-capability
  restoration plus localized diagnostics
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-01 — Integration_push metadata read succeeds under declared contract

- Covers: `VOC-114-AC-01`, `VOC-114-D03`
- Preconditions: T00 task branch with runner and job-token permission updates
- Procedure: Deterministic fixture simulates `integration_push` metadata phase with
  token contract including Checks/Actions/Contents read; assert check-runs, status,
  workflow-run, and commit file reads succeed and recovery may proceed to wait/dispatch
  planning.
- Expected result: Metadata phase completes; no generic metadata failure
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-02 — Promotion_pr metadata read succeeds under declared contract

- Covers: `VOC-114-AC-01`, `VOC-114-D03`
- Preconditions: T00 task branch
- Procedure: Deterministic fixture simulates `promotion_pr` metadata phase with the
  same declared read contract; assert gate summary and workflow-run discovery succeed.
- Expected result: Metadata phase completes for promotion recovery path
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-03 — Endpoint-class errors are sanitized and specific

- Covers: `VOC-114-AC-02`, `VOC-114-D02`
- Preconditions: T00 task branch
- Procedure: Negative fixtures force check-runs, workflow-runs, and commit-metadata
  `gh` failures; assert raised/printed classes are
  `check_runs_read_failed`, `workflow_runs_read_failed`, and
  `commit_metadata_read_failed` respectively, with no response bodies or secrets
  in messages.
- Expected result: Localized fail-closed diagnostics; generic
  `github_metadata_read_failed` not used for these cases
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-04 — Absent read permission fails closed with correct endpoint class

- Covers: `VOC-114-AC-03`, `VOC-114-D03`
- Preconditions: T00 task branch
- Procedure: Fixtures/model tests where the job contract omits Checks or Actions access;
  assert the corresponding endpoint class is emitted and recovery exits non-zero
  before wait/dispatch.
- Expected result: Fail-closed refusal; no dispatch side effects
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-05 — No dispatch after metadata-read failure

- Covers: `VOC-114-AC-03`, `VOC-114-D00`
- Preconditions: T00 task branch
- Procedure: Simulate metadata-read failure mid-run; assert
  `plan_recovery_dispatches` / dispatch execution paths are not invoked and
  merge-gate/release policy tests still forbid unbacked fabrication shortcuts.
  D07 tests also require exact successful workflow identities, reject status or
  foreign/neutral/failing evidence, revalidate the live PR, and prove that
  bridge statuses are excluded from future authoritative selection.
- Expected result: Recovery stops at metadata phase; VOC-113 invariants preserved
- Evidence: `VOC-114-EV-00`

## VOC-114-TEST-06 — Live integration_push recovery progresses past metadata read

- Covers: `VOC-114-AC-04`
- Preconditions: T00 live; operator-owned contract
  `.karsift/live-evidence/VOC-114-T01.yaml`
- Procedure: Operator/repository-controlled path re-runs integration_push recovery
  for documented SHA; read `t01-evidence.md` for allowlisted metadata showing
  metadata read success and genuine workflow run creation or observation for that
  exact SHA.
- Expected result: No immediate metadata-read failure; genuine runs bound to exact SHA
- Evidence: `VOC-114-EV-01`

## VOC-114-TEST-07 — Live reconcile-release recovery unblocks promotion PR #947

- Covers: `VOC-114-AC-04`, `VOC-114-AC-05`
- Preconditions: T00 live; operator-owned contract
  `.karsift/live-evidence/VOC-114-T01.yaml`
- Procedure: Operator dispatches `reconcile-release` for release issue #946 (or
  observes equivalent release converge recovery); assert promotion PR #947 exact
  head receives genuine required checks and becomes merge-eligible under VOC-108
  selection. Then dispatch `pipeline.yml` on the exact T01 carrier branch with
  `action=verify-promotion-check-recovery` and `promotion_pr_number=947`; require
  job `verify-promotion-check-recovery / verify` on the carrier's exact PR head.
  Read `t01-evidence.md` for the complementary both-mode metadata.
- Expected result: #947 unblocked for VOC-113-T01; no unbacked statuses; any D07
  ruleset attestation is derived only after genuine success, and merge occurs
  only via release converge
- Evidence: `VOC-114-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
