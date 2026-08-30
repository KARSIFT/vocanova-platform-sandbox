# VOC-141 — Acceptance Criteria

## VOC-141-AC-00 — SUCCESS `ci / ci` with an unattestable parent immediately dispatches dedicated validation

- Requirement source: `VOC-141-D01`, `VOC-141-D02`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-00`, `VOC-141-TEST-01`
- Evidence: `VOC-141-EV-00`
- Result: pending

When GitHub's required PR view reports `ci / ci` SUCCESS but the
authoritative composed evidence has no uniquely attestable completed
non-carrier parent, promotion recovery immediately plans exactly one
`pipeline.yml` dispatch with `action=recover-promotion-pr-checks` and
`promotion_pr_number` bound to the promotion PR. It does not rerun a doomed
`pull_request` `ci / ci` carrier and does not poll until the 1,800-second
timeout with no dispatch. The class of run `33340381776` / job `99334840338`
cannot recur.

## VOC-141-AC-01 — A valid completed dedicated parent completes without redispatch

- Requirement source: `VOC-141-D03`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-02`
- Evidence: `VOC-141-EV-00`
- Result: pending

When composed evidence already contains a unique exact-head / PR-bound
completed successful `promotion-pr-validation PR #<n>` parent,
`recovery_complete` is true and dedicated `recover-promotion-pr-checks` is
not dispatched again. A skipped `release` job merely defined in the shared
workflow does not disqualify that dedicated parent.

## VOC-141-AC-02 — Active or successful exact dedicated recovery suppresses duplicates

- Requirement source: `VOC-141-D04`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-03`
- Evidence: `VOC-141-EV-00`
- Result: pending

A queued, in-progress, or completed-successful exact-head dedicated
`promotion-pr-validation PR #<n>` run suppresses a duplicate dedicated
dispatch. An in-progress, failed, cancelled, or successful release carrier,
another native `pipeline.yml` carrier, and a GitHub-required SUCCESS row
whose parent is unattestable do not suppress that dispatch.

## VOC-141-AC-03 — Timeout diagnostics identify unattestable CI evidence

- Requirement source: `VOC-141-D05`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-04`
- Evidence: `VOC-141-EV-00`
- Result: pending

When the recovery poll deadline expires and composed `ci / ci` is still
unattestable, timeout diagnostics include a stable sanitized token such as
`unattestable_ci_evidence` (or the live equivalent) rather than
`missing_checks: none` alone. Existing mode, target SHA, promotion PR number,
timeout, pending, failed, and successful fields remain. Diagnostics do not
print tokens, secrets, or raw provider payloads.

## VOC-141-AC-04 — Production and release-carrier fail-closed boundaries remain unchanged

- Requirement source: `VOC-141-D06`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-05`
- Evidence: `VOC-141-EV-00`
- Result: pending

In-progress or failed release carriers are never trusted recovered `ci / ci`.
Exact-identity `pr-validation` binding, doomed-job rerun refusal, required
`ci / ci`, and the VOC-140 two-token production merge guard remain. The
implementation does not add bypass actors, fabricate statuses, change App
token mints, or manually merge.

## VOC-141-AC-05 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-141-D08`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-06`
- Evidence: `VOC-141-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `67bdfd13ef875dead23ce4be01d7d0e8b976e289`
if that merge still omits dedicated dispatch for unattestable SUCCESS CI.
Every changed authoritative fixture file is byte-identical to that merge.

## VOC-141-AC-06 — Current-state docs match the live unattestable-SUCCESS dispatch contract

- Requirement source: `VOC-141-D10`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-07`
- Evidence: `VOC-141-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim that
required-check SUCCESS completes promotion recovery or that dedicated
validation is dispatched whenever no completed non-carrier run exists.
Fixture README, `docs/operations/11-devops-and-ci-cd.md`, and every other
current-state match state that GitHub-required SUCCESS is not sufficient when
composed CI is unattestable and that recovery dispatches dedicated validation
immediately rather than polling to timeout. Clearly marked historical VOC-140
records remain historical. VOC-140, VOC-139, and VOC-138 package directories
are not rewritten.

## VOC-141-AC-07 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-141-D11`, `VOC-141-D12`, `VOC-141-D13`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-08`
- Evidence: `VOC-141-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite,
`git diff --check`, and independent exact-revision review that binds the live
head all pass. `roles.yml` is unchanged. Evidence does not require a commit
to contain its own SHA.

## VOC-141-AC-08 — reconcile-release can proceed without the 1,800-second hang

- Requirement source: `VOC-141-D09`, `VOC-141-D15`
- Tasks: `VOC-141-T00`
- Tests: `VOC-141-TEST-09`
- Evidence: `VOC-141-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No duplicate
promotion PR or release audit is created. After exact-SHA review and merge
into `develop`, ordinary `reconcile-release` can recover the live
same-repository promotion without waiting 1,800 seconds on unattestable
SUCCESS `ci / ci`, merge that promotion, synchronize `develop` to that exact
merge SHA, and allow the normal automatic production deployment on the
resulting `main` push to run and be verified. Root issue #1109 closes only
after allowlisted metadata from a successful recovery/release run exists.
Closed state alone is not completion proof. Named incident run/job IDs remain
audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
