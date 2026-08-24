# VOC-113 — Acceptance Criteria

## VOC-113-AC-00 — Problem evidence and trigger/token/event behavior documented

- Requirement source: issue #948; `VOC-113-D00`, `VOC-113-DEP-04`
- Tasks: `VOC-113-T00`
- Tests: `VOC-113-TEST-00`
- Evidence: `VOC-113-EV-00`
- Result: pending

`t00-evidence.md` records the issue #948 observations (VOC-112 develop merge
without push workflows; promotion PR #947 without required checks; failed
close/reopen and draft/ready recovery; reconcile-release reuse without merge)
and documents the verified trigger/token/event behavior that explains missing
downstream runs. Evidence contains no secrets, full logs, or personal data.

## VOC-113-AC-01 — Task-merge-to-integration recovery obtains genuine exact-SHA validation

- Requirement source: `VOC-113-D00`–`D04`, `VOC-113-D07`
- Tasks: `VOC-113-T00`
- Tests: `VOC-113-TEST-01`, `VOC-113-TEST-02`, `VOC-113-TEST-05`, `VOC-113-TEST-06`
- Evidence: `VOC-113-EV-00`
- Result: pending

After an App-driven task merge to `develop`, if required push/validation workflows
are absent for the exact integration SHA, repository-managed recovery starts
genuine runs for that SHA (explicit reusable-workflow / dispatch orchestration
allowed). Merge-gate / adopt consumers still fail closed until authoritative
exact-SHA evidence exists. No fabricated statuses.

## VOC-113-AC-02 — Release-PR recovery obtains genuine exact-head required checks

- Requirement source: `VOC-113-D00`–`D06`, `VOC-113-D08`
- Tasks: `VOC-113-T00`, `VOC-113-T01`
- Tests: `VOC-113-TEST-03`, `VOC-113-TEST-04`, `VOC-113-TEST-05`, `VOC-113-TEST-06`,
  `VOC-113-TEST-08`
- Evidence: `VOC-113-EV-00`, `VOC-113-EV-01`
- Result: pending

After App-created promotion PR creation (or on `reconcile-release`), if required
pull-request checks (`governance-policy`, `validate`, `ci / ci` or their current
ruleset-equivalent contexts) are absent for the exact head SHA, recovery starts
genuine runs for that head. Release converge remains fail-closed until newest
authoritative exact-head checks succeed. PR #947 is completed only under that
condition.

## VOC-113-AC-03 — Fail-closed, no fabrication, no duplicates, App auth preserved

- Requirement source: `VOC-113-D01`, `VOC-113-D02`, `VOC-113-D05`, `VOC-113-D06`
- Tasks: `VOC-113-T00`
- Tests: `VOC-113-TEST-05`, `VOC-113-TEST-06`, `VOC-113-TEST-07`
- Evidence: `VOC-113-EV-00`
- Result: pending

Deterministic tests prove: fabricated/wrong-SHA/stale evidence is rejected;
timeout fails closed with sanitized diagnostics; duplicate promotion PRs / release
audits are not created; recovery does not recurse unboundedly; App-token mutation
posture is preserved and `github.token` merge fallback remains refused when App
credentials are configured.

## VOC-113-AC-04 — Ruleset and risk classification preserved

- Requirement source: issue #948 scope boundary; `VOC-113-D01`
- Tasks: `VOC-113-T00`
- Tests: `VOC-113-TEST-07`
- Evidence: `VOC-113-EV-00`
- Result: pending

The package does not weaken branch ruleset required contexts or governance risk
classification. Recovery never marks a required context successful without a
real workflow run bound to the exact SHA.

## VOC-113-AC-05 — Promotion PR #947 completed only after genuine exact-head checks

- Requirement source: `VOC-113-D08`
- Tasks: `VOC-113-T01`
- Tests: `VOC-113-TEST-08`
- Evidence: `VOC-113-EV-01`
- Result: pending

Operator-owned live evidence shows `verify-promotion-check-recovery / verify`
succeeded for promotion PR #947's exact head (genuine required checks present),
and the promotion merge occurred only afterward via release converge (or remains
correctly fail-closed if checks fail). Metadata-only evidence; no status
fabrication.

## VOC-113-AC-06 — Post-promotion workflows verified before remediation closes

- Requirement source: issue #948 required outcome; `VOC-113-D08`
- Tasks: `VOC-113-T02`
- Tests: `VOC-113-TEST-09`
- Evidence: `VOC-113-EV-02`
- Result: pending

After #947 merges to `main`, operator-owned live evidence from
`verify-post-promotion-workflow / verify` confirms the expected post-promotion
workflow ran for #947's exact merge-result SHA (at minimum the repository's
normal `main` push path such as `deploy-production` when selected). The verifier
run is itself bound to the exact T02 carrier head. Issue #948 / this remediation
closes only after that evidence is recorded.

## VOC-113-AC-07 — Provenance validation accepts later PRs without trusting discarded history

- Requirement source: live plan-PR #949 failure; `VOC-113-D11`
- Tasks: `VOC-113-T00`
- Tests: `VOC-113-TEST-10`
- Evidence: `VOC-113-EV-00`
- Result: pending

The original capture PR still requires subject-commit ancestry and exact hashes.
A later PR based on the accepted squash passes only when the expected source
hashes are already anchored in its merge base and remain unchanged at its exact
head. Missing, tampered, or changed hashes fail closed.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
