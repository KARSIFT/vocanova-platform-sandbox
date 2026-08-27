# VOC-138 — Acceptance Criteria

## VOC-138-AC-00 — Promotion PRs with an unreachable capture subject use pr-validation

- Requirement source: `VOC-138-D01`, `VOC-138-D02`, `VOC-138-D05`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-00`, `VOC-138-TEST-01`, `VOC-138-TEST-02`
- Evidence: `VOC-138-EV-00`
- Result: pending

A same-repository `main` <- `develop` promotion PR that supplies valid exact
base/head SHAs, whose capture fixture differs between those SHAs, and whose
recorded `subject_revision` commit object cannot be resolved in the existing
checkout, selects `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` and keeps
`PR_BASE_SHA` / `PR_HEAD_SHA` set. It does not select `pr-ancestry` and does
not switch to `--squash-safe-push`.

## VOC-138-AC-01 — Named VOC-112 tests pass under that exact-base/head mode

- Requirement source: `VOC-138-D02`, `VOC-138-D08`, `VOC-138-D09`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-01`, `VOC-138-TEST-02`
- Evidence: `VOC-138-EV-00`
- Result: pending

`VOC-112-TEST-12` and `VOC-112-TEST-13` pass when provenance mode is
`pr-validation` for a promotion checkout whose recorded subject
`f9d11e232a07c7d7a9c433d02c9267912543ba10` is absent. The eight VOC-112
no-change paths remain byte-identical to
`b9e74fc2db4691c48c637639b265d527de9f4505`. The provenance test itself is
not edited to weaken `pr-ancestry`.

## VOC-138-AC-02 — Ordinary fixture-changing PRs remain pr-ancestry fail-closed

- Requirement source: `VOC-138-D03`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-03`, `VOC-138-TEST-04`
- Evidence: `VOC-138-EV-00`
- Result: pending

When the promotion signal is absent, adding, modifying, or deleting the
capture fixture still selects `pr-ancestry`. An ordinary fixture-changing
checkout whose subject commit object is missing still fails closed with
`PR ancestry mode requires every captured commit object`. Forks and other
base/head pairs cannot select the promotion exception.

## VOC-138-AC-03 — Hash and SHA negatives still fail closed

- Requirement source: `VOC-138-D04`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-05`, `VOC-138-TEST-06`
- Evidence: `VOC-138-EV-00`
- Result: pending

Missing or malformed `PR_BASE_SHA` / `PR_HEAD_SHA`, unrelated base/head
pairs, tampered merge-base, and tampered current hashes still fail closed
under `pr-validation`. Capture-fixture comparison errors still fail closed.
Unchanged fixtures with valid SHAs still select `pr-validation`.

## VOC-138-AC-04 — No evidence hydration or weakened required check

- Requirement source: `VOC-138-D06`, `VOC-138-D07`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-07`, `VOC-138-TEST-08`
- Evidence: `VOC-138-EV-00`
- Result: pending

The implementation diff adds no capture-commit fetch helper, hydrate helper,
provenance-mode wrapper, skip, or test-time evidence mutation. `run-app-checks.sh`
and reusable `ci.yml` still do not `git fetch` evidence commits. Required
`ci / ci` remains a required check. Status attestations, if published, are
backed by genuine exact-head Actions success.

## VOC-138-AC-05 — Recovery does not rerun a structurally doomed PR job

- Requirement source: `VOC-138-D07`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-08`, `VOC-138-TEST-09`
- Evidence: `VOC-138-EV-00`
- Result: pending

When GitHub's selected required `ci / ci` row is a failed `pull_request` run
whose provenance classification cannot change on rerun, and a successful
exact-head application-check of the same workflow already exists (including
`workflow_dispatch` class `33122158425`), recovery selects or publishes that
unambiguous success instead of rerunning the doomed job. Identity mismatches
outside this class still raise `selected_required_run_mismatch`.

## VOC-138-AC-06 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-138-D10`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-10`
- Evidence: `VOC-138-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `b263c0c110591cc798b89277dfc35542abb1597b`
if that merge still exhibits the #1090 failure class. Every changed
authoritative fixture file is byte-identical to that merge.

## VOC-138-AC-07 — Current-state docs match the live contract

- Requirement source: `VOC-138-D11`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-11`
- Evidence: `VOC-138-EV-00`
- Result: pending

Fixture README, `docs/operations/11-devops-and-ci-cd.md`, and
`docs/development/agent-skills.md` describe promotion PR `pr-validation` when
the subject is unreachable, ordinary fixture-changing `pr-ancestry`, and
non-PR dispatch `squash-safe-push`. They do not claim that the promotion PR
application check uses `--squash-safe-push`. VOC-135/VOC-136/VOC-137 package
records are not rewritten.

## VOC-138-AC-08 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-138-D12`, `VOC-138-D13`, `VOC-138-D14`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-12`
- Evidence: `VOC-138-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite, VOC-112
navigation-benchmark tests, `git diff --check`, and independent exact-revision
review that binds the live head all pass. `roles.yml` is unchanged. Evidence
does not require a commit to contain its own SHA.

## VOC-138-AC-09 — reconcile-release can merge the promotion and converge develop

- Requirement source: `VOC-138-D16`
- Tasks: `VOC-138-T00`
- Tests: `VOC-138-TEST-13`
- Evidence: `VOC-138-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. After exact-SHA
review and merge into `develop`, `reconcile-release` for #1089 can merge
#1090 (or the live same-repository promotion at the then-current `develop`
head), synchronize `develop` to that exact merge SHA, and allow the normal
production deployment gate to run. Closed state alone is not completion
proof. Named incident run/job IDs remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
