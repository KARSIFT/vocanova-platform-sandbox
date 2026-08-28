# VOC-139 — Acceptance Criteria

## VOC-139-AC-00 — Accumulated promotion PRs pass VOC-112 under head-bound pr-validation

- Requirement source: `VOC-139-D01`, `VOC-139-D02`, `VOC-139-D04`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-00`, `VOC-139-TEST-01`
- Evidence: `VOC-139-EV-00`
- Result: pending

A same-repository `main` ← `develop` promotion PR that supplies valid exact
base/head SHAs keeps `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` and keeps
`PR_BASE_SHA` / `PR_HEAD_SHA` set. When `AGENTS.md` and/or the navigator skill
differ between the PR merge-base (`main`) and `PR_HEAD_SHA`, stored capture
hashes that equal the files at `PR_HEAD_SHA` and the working tree cause
`VOC-112-TEST-12` and `VOC-112-TEST-13` to pass. Those tests do not require
stored hashes to equal the historical `main` merge-base files. Missing or
non-ancestor capture subjects still do not fail promotion `pr-validation`.

## VOC-139-AC-01 — Ordinary pr-validation remains merge-base hash-anchored

- Requirement source: `VOC-139-D03`, `VOC-139-D04`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-02`, `VOC-139-TEST-03`
- Evidence: `VOC-139-EV-00`
- Result: pending

When the promotion signal is absent, `pr-validation` still requires stored
hashes to equal the PR merge-base files. An ordinary checkout whose
`AGENTS.md` hash matches HEAD but not the merge-base still fails closed with
`AGENTS.md hash must be anchored in the PR merge base` (or the navigator
equivalent). Ordinary fixture add/modify/delete still selects `pr-ancestry`.
Forks and other base/head pairs cannot select the promotion hash exception.

## VOC-139-AC-02 — Malformed, unrelated, and wrong-identity cases fail closed

- Requirement source: `VOC-139-D04`, `VOC-139-D07`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-04`, `VOC-139-TEST-05`, `VOC-139-TEST-06`
- Evidence: `VOC-139-EV-00`
- Result: pending

Missing or malformed `PR_BASE_SHA` / `PR_HEAD_SHA`, unrelated base/head
commits, tampered current hashes, unrelated repository or PR number, and
wrong base/head refs still fail closed. Unchanged fixtures with valid SHAs
and no promotion signal still select merge-base-anchored `pr-validation`.

## VOC-139-AC-03 — Recovery metadata resolves without a checkout

- Requirement source: `VOC-139-D07`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-07`
- Evidence: `VOC-139-EV-00`
- Result: pending

Job `promotion-pr-metadata` invokes `gh pr view` with `-R "$GITHUB_REPOSITORY"`
(or equivalent `--repo`) and has no checkout step. A deterministic test
executes the real metadata command body in a directory that is not a git
repository and succeeds only when the explicit repository flag is present.
Live caller `pipeline.yml`, the infrastructure template, and the mirrored
fixture carry the same repository-explicit invocation. Wrong refs or
repository still fail with `promotion pair mismatch` (or the live equivalent).

## VOC-139-AC-04 — Recovery attestation stays exact-identity pr-validation

- Requirement source: `VOC-139-D05`, `VOC-139-D08`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-08`
- Evidence: `VOC-139-EV-00`
- Result: pending

Recovery still requires genuine Actions success bound to the promotion PR
number, immutable base/head SHAs, repository, configured branch pair,
expected workflow/path, and `pr-validation`. Same-head `--squash-safe-push`
is still rejected. Doomed `pull_request` `ci / ci` jobs are still not rerun
as a strategy. Required `ci / ci` remains a required check.

## VOC-139-AC-05 — No hydration, recapture, or seven-path VOC-112 rewrite

- Requirement source: `VOC-139-D06`, `VOC-139-D10`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-09`
- Evidence: `VOC-139-EV-00`
- Result: pending

The implementation diff adds no capture-commit fetch helper, hydrate helper,
skip of `VOC-112-TEST-12` / `VOC-112-TEST-13`, or rewritten VOC-112 JSON
fixtures. The seven remaining no-change paths stay byte-identical to
`b9e74fc2db4691c48c637639b265d527de9f4505`. JSON `subject_revision` remains
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. VOC-138 package records are not
rewritten. Live VOC-138 `NO_CHANGE_PATHS` no longer freezes the provenance
test.

## VOC-139-AC-06 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-139-D11`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-10`
- Evidence: `VOC-139-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `123735c80fec813a5b46a004f3e1122bd425cde2`
if that merge still exhibits the #1096 failure class. Every changed
authoritative fixture file is byte-identical to that merge.

## VOC-139-AC-07 — Current-state docs match the live hash and metadata contract

- Requirement source: `VOC-139-D12`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-11`
- Evidence: `VOC-139-EV-00`
- Result: pending

Fixture README, `docs/operations/11-devops-and-ci-cd.md`, and
`docs/development/agent-skills.md` describe authenticated promotion PR
`pr-validation` with hashes bound to the reviewed head/source revision,
ordinary unchanged-fixture merge-base anchoring, ordinary fixture-changing
`pr-ancestry`, no-checkout repository-explicit recovery metadata, and
non-PR dispatch `squash-safe-push`. They do not claim that accumulated
promotion PRs must match historical `main` hashes. VOC-138 package records
are not rewritten.

## VOC-139-AC-08 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-139-D13`, `VOC-139-D14`, `VOC-139-D15`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-12`
- Evidence: `VOC-139-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite, VOC-112
navigation-benchmark tests, `git diff --check`, and independent exact-revision
review that binds the live head all pass. `roles.yml` is unchanged. Evidence
does not require a commit to contain its own SHA.

## VOC-139-AC-09 — reconcile-release can merge the promotion and converge develop

- Requirement source: `VOC-139-D09`, `VOC-139-D17`
- Tasks: `VOC-139-T00`
- Tests: `VOC-139-TEST-13`
- Evidence: `VOC-139-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. After exact-SHA
review and merge into `develop`, `reconcile-release` for #1089 can merge
#1090 (or the live same-repository promotion at the then-current `develop`
head) once required checks pass under the repaired hash and metadata
contract, synchronize `develop` to that exact merge SHA, and allow the
normal automatic production deployment on the resulting `main` push to run
and be verified. Root issue #1096 closes only after allowlisted metadata
from the successful recovery/release run exists. Closed state alone is not
completion proof. Named incident run/job IDs remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
