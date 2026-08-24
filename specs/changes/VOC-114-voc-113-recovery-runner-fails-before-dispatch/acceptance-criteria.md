# VOC-114 — Acceptance Criteria

## VOC-114-AC-00 — Verified root cause and token contract documented

- Requirement source: issue #956; `VOC-114-D00`, `VOC-114-D01`, `VOC-114-D05`
- Tasks: `VOC-114-T00`
- Tests: `VOC-114-TEST-00`
- Evidence: `VOC-114-EV-00`
- Result: pending

`t00-evidence.md` records issue #956 observations and documents the verified
metadata-read failure class (including whether installation grants, mint
`permission-*` inputs, or both were required). Evidence contains no secrets,
full logs, or personal data.

## VOC-114-AC-01 — Recovery metadata reads succeed under declared App token contract

- Requirement source: `VOC-114-D01`, `VOC-114-D03`
- Tasks: `VOC-114-T00`
- Tests: `VOC-114-TEST-01`, `VOC-114-TEST-02`
- Evidence: `VOC-114-EV-00`
- Result: pending

For both `integration_push` and `promotion_pr` modes, the recovery runner can
read exact-SHA check-runs, commit status, and workflow-run metadata using the
minted App token from merge-gate, release converge, and
`recover-actions-checks.yml`. Existing mutation permissions remain unchanged
except for explicitly declared read scopes required by the metadata phase.

## VOC-114-AC-02 — Sanitized endpoint-class diagnostics replace generic metadata failure

- Requirement source: `VOC-114-D02`
- Tasks: `VOC-114-T00`
- Tests: `VOC-114-TEST-03`
- Evidence: `VOC-114-EV-00`
- Result: pending

When a metadata endpoint fails, the runner emits exactly one sanitized class:
`check_runs_read_failed`, `workflow_runs_read_failed`, or
`commit_metadata_read_failed`. Generic `github_metadata_read_failed` is not used
for localized endpoint failures. No response bodies, tokens, logs, or user data
are emitted.

## VOC-114-AC-03 — Fail closed with no dispatch after metadata-read failure

- Requirement source: `VOC-114-D00`, `VOC-114-D03`
- Tasks: `VOC-114-T00`
- Tests: `VOC-114-TEST-04`, `VOC-114-TEST-05`
- Evidence: `VOC-114-EV-00`
- Result: pending

Deterministic tests prove absent read capability fails closed with the correct
endpoint class and that recovery does not plan or execute workflow dispatch after
any metadata-read failure. VOC-113 no-fabrication and VOC-108 authoritative
selection invariants remain intact.

## VOC-114-AC-04 — Live both recovery modes produce genuine exact-SHA validation

- Requirement source: issue #956 required outcome; `VOC-114-D06`
- Tasks: `VOC-114-T01`
- Tests: `VOC-114-TEST-06`, `VOC-114-TEST-07`
- Evidence: `VOC-114-EV-01`
- Result: pending

Operator-owned live evidence shows: (a) integration_push recovery for the
documented merged SHA progresses past metadata read and creates or observes
genuine push/validation runs; (b) `reconcile-release` for release issue #946
progresses past metadata read and creates or observes genuine pull-request checks
for promotion PR #947's exact head; and (c) the existing read-only
`verify-promotion-check-recovery / verify` job succeeds on the exact T01 carrier
PR head. No fabricated statuses.

## VOC-114-AC-05 — Promotion PR #947 unblocked for VOC-113-T01 completion

- Requirement source: issue #956; `VOC-114-D06`; VOC-113 `VOC-113-AC-05`
- Tasks: `VOC-114-T01`
- Tests: `VOC-114-TEST-07`
- Evidence: `VOC-114-EV-01`
- Result: pending

After live recovery, promotion PR #947's exact head receives genuine required
checks (`governance-policy`, `validate`, `ci / ci` or current ruleset equivalents)
and becomes eligible for release converge merge under VOC-108 authoritative
selection. The exact-carrier verifier independently confirms those contexts.
Metadata-only evidence; merge occurs only via release converge after genuine
success, not by manual intervention or status fabrication.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
