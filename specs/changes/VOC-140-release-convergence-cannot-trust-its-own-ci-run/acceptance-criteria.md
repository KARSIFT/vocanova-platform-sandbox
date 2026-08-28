# VOC-140 — Acceptance Criteria

## VOC-140-AC-00 — In-progress or failed release carriers are never trusted recovered `ci / ci`

- Requirement source: `VOC-140-D01`, `VOC-140-D03`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-00`, `VOC-140-TEST-01`
- Evidence: `VOC-140-EV-00`
- Result: pending

Promotion recovery and authoritative selection do not treat a `ci / ci` check
as attestable when its parent workflow run is not `completed`/`success`, when
that run contains `release /` jobs, or when that run is a
`reconcile-release` / release-converge carrier. The circular-CI class of run
`33136633666` / job `98738317266` cannot recur: recovery does not report
complete without dispatch merely because required PR checks show SUCCESS for
that carrier check.

## VOC-140-AC-01 — Dedicated `promotion-pr-validation` is dispatched or selected and must be completed/successful

- Requirement source: `VOC-140-D02`, `VOC-140-D03`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-02`, `VOC-140-TEST-03`
- Evidence: `VOC-140-EV-00`
- Result: pending

When no completed non-carrier `ci / ci` run exists, recovery dispatches
`pipeline.yml` `action=recover-promotion-pr-checks` with
`display_title` `promotion-pr-validation PR #<n>` and does not treat the
in-progress release carrier as covering that dispatch. Attestation accepts
that dedicated run only after it is `completed`/`success` and exact-head /
PR / repository / path bound. An in-progress dedicated dispatch may suppress
a duplicate dispatch; a completed successful dedicated run may suppress
redispatch. Any other in-progress `pipeline.yml` run must not.

## VOC-140-AC-02 — Recovery attestation stays exact-identity and does not rerun doomed pull-request `ci / ci`

- Requirement source: `VOC-140-D04`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-04`
- Evidence: `VOC-140-EV-00`
- Result: pending

Recovery still requires genuine Actions success bound to the promotion PR
number, immutable base/head SHAs, repository, configured branch pair,
expected workflow/path, and `pr-validation`. Same-head `--squash-safe-push`
is still rejected. Doomed `pull_request` `ci / ci` jobs are still not rerun
as a strategy. Required `ci / ci` remains a required check. Completed
exact-bound `pull_request` `ci / ci` may attest only when its parent
workflow run is completed/successful and is not a release carrier.

## VOC-140-AC-03 — Production merge guard still requires an effective non-bypassable ruleset

- Requirement source: `VOC-140-D05`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-05`, `VOC-140-TEST-06`
- Evidence: `VOC-140-EV-00`
- Result: pending

The verifier still requires an effective active repository-owned ruleset on
the production branch, a pull-request rule, strict non-empty required checks,
and `bypass_actors: []`. Non-empty bypass, inactive enforcement, missing
pull-request rule, non-strict checks, empty/malformed required-check lists,
and source mismatches still fail closed. The implementation does not add
bypass actors, fabricate statuses, or manually merge.

## VOC-140-AC-04 — The merge App identity can prove those fields; omitted fields fail distinctly

- Requirement source: `VOC-140-D06`, `VOC-140-D07`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-07`, `VOC-140-TEST-08`
- Evidence: `VOC-140-EV-00`
- Result: pending

The identity that calls `verify-production-merge-guard.sh` immediately before
`gh pr merge` requests the least-privilege mint permission that actually
returns `bypass_actors` as an array. A payload that omits `bypass_actors` or
returns a non-array fails closed with a distinct sanitized class (not
`production_merge_guard_missing`) and names the operator action to grant that
permission on GitHub App `karsift-ai-infra-bot` for this repository. The
production-branch `merge-gate.yml` path that already calls the same verifier
uses the same mint contract. The verifier is not switched to `github.token`
to avoid diagnosing the App contract.

## VOC-140-AC-05 — Regression tests exercise the real token-visible payload and circular-CI identity

- Requirement source: `VOC-140-D08`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-00`, `VOC-140-TEST-07`, `VOC-140-TEST-08`
- Evidence: `VOC-140-EV-00`
- Result: pending

Tests include a fixture in which the newest `ci / ci` check is SUCCESS while
the parent workflow run is `in_progress` and contains `release / converge`.
Tests include omitted-`bypass_actors`, `bypass_actors: []`, and non-empty
bypass payloads, including a subprocess of the real
`verify-production-merge-guard.sh` with a mock `gh`. A helper-only test that
duplicates `validate_production_merge_guard` with complete admin fixtures is
not sufficient coverage.

## VOC-140-AC-06 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-140-D09`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-09`
- Evidence: `VOC-140-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `599436835371f27fac52ec6b47a18b36257366ac`
if that merge still exhibits the #1102 failure class. Every changed
authoritative fixture file is byte-identical to that merge.

## VOC-140-AC-07 — Current-state docs match the live recovery and App-token contract

- Requirement source: `VOC-140-D11`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-10`
- Evidence: `VOC-140-EV-00`
- Result: pending

Fixture README and `docs/operations/11-devops-and-ci-cd.md` describe that
recovery/selection never treats a still-running release carrier as attestable
`ci / ci`, that dedicated `promotion-pr-validation` must be
completed/successful, and that the merge App identity requests the
least-privilege permission required to prove ruleset/no-bypass fields. They
do not claim the App token is contents/issues/pull-requests-only if D06 added
an administration permission. VOC-139 and VOC-138 package records are not
rewritten.

## VOC-140-AC-08 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-140-D12`, `VOC-140-D13`, `VOC-140-D14`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-11`
- Evidence: `VOC-140-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite,
`git diff --check`, and independent exact-revision review that binds the live
head all pass. `roles.yml` is unchanged. Evidence does not require a commit
to contain its own SHA.

## VOC-140-AC-09 — reconcile-release can merge the promotion and converge develop

- Requirement source: `VOC-140-D10`, `VOC-140-D16`
- Tasks: `VOC-140-T00`
- Tests: `VOC-140-TEST-12`
- Evidence: `VOC-140-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No duplicate
promotion PR or release audit is created. After exact-SHA review and merge
into `develop`, dedicated promotion recovery may be rerun if necessary, then
`reconcile-release` for #1089 can merge #1090 (or the live same-repository
promotion at the then-current `develop` head) once required checks pass under
the repaired identity and guard contract, synchronize `develop` to that exact
merge SHA, and allow the normal automatic production deployment on the
resulting `main` push to run and be verified. Root issue #1102 closes only
after allowlisted metadata from the successful recovery/release run exists.
Closed state alone is not completion proof. Named incident run/job IDs remain
audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
