# VOC-142 — Acceptance Criteria

## VOC-142-AC-00 — Roster wait does not complete while required `ci / ci` is missing or IN_PROGRESS

- Requirement source: `VOC-142-D01`, `VOC-142-D03`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-00`, `VOC-142-TEST-01`
- Evidence: `VOC-142-EV-00`
- Result: pending

Roster adoption does not attempt merge until the complete
ruleset-required logical-context set for the exact roster head is registered
and the newest logical attempt of every required row, including `ci / ci`,
is SUCCESS. At issue creation that set is `ci / ci`, `governance-policy`,
and `validate`. The class of run `33343125733` / job `99342230038` cannot
recur.

## VOC-142-AC-01 — A partial green snapshot is not complete merely because its count is stable

- Requirement source: `VOC-142-D02`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-01`
- Evidence: `VOC-142-EV-00`
- Result: pending

Two consecutive authoritative snapshots with `pending == 0`, `failed == 0`,
and an unchanged `total_count` plus head SHA do not complete wait when a
required logical context is still unregistered. Fail closed if the required
set cannot be uniquely resolved.

## VOC-142-AC-02 — In-progress required rows are not-ready, not attestable SUCCESS

- Requirement source: `VOC-142-D03`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-00`, `VOC-142-TEST-08`
- Evidence: `VOC-142-EV-00`
- Result: pending

An IN_PROGRESS or queued required `ci / ci` row keeps wait incomplete. The
wait does not use `statusCheckRollup` or `gh pr checks`. Merge-gate and
release still do not treat an in-progress parent as attestable completed
`ci / ci`.

## VOC-142-AC-03 — Reconcile reuses an existing exact open roster PR

- Requirement source: `VOC-142-D04`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-02`
- Evidence: `VOC-142-EV-00`
- Result: pending

When a new roster commit is pushed (`changed == 'true'`) and exactly one
OPEN PR already exists for the deterministic `karsift/roster-*` head, the
integration-branch base, this repository, and the exact pushed head SHA,
adopt reuses that PR number and resumes wait/merge/cleanup from it. It does
not call `gh pr create`. The class of reconcile run `33343250178` / job
`99342577393` cannot recur.

## VOC-142-AC-04 — Create only when no match; reject mismatched or ambiguous carriers

- Requirement source: `VOC-142-D05`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-03`
- Evidence: `VOC-142-EV-00`
- Result: pending

A roster PR is created only when no matching OPEN or already-merged carrier
exists. Two or more matches, or a single carrier whose repository, base,
head ref, or head SHA does not match the pushed identity, fail closed. No
unrelated PR is closed, recreated, or retargeted.

## VOC-142-AC-05 — An already-merged exact roster carrier is reused, not duplicated

- Requirement source: `VOC-142-D05`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-04`
- Evidence: `VOC-142-EV-00`
- Result: pending

Exactly one already-merged PR with the same repository, head ref, exact head
SHA, and base ref is treated as the matching carrier. Adopt does not create
another roster PR. Cleanup/dispatch follow the existing merged-roster path
as applicable.

## VOC-142-AC-06 — Existing tasks are reused and root dispatch happens exactly once

- Requirement source: `VOC-142-D06`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-05`
- Evidence: `VOC-142-EV-00`
- Result: pending

Reconcile reuses existing task issues. After the checked roster merges, the
root task is dispatched only when that issue is OPEN and no implementation
PR exists for the deterministic implementation head. An interrupted adoption
that later succeeds through the reused roster PR does not open a second task
or dispatch a second root implementation. Task #1111 remains the VOC-141
task.

## VOC-142-AC-07 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-142-D08`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-06`
- Evidence: `VOC-142-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `67bdfd13ef875dead23ce4be01d7d0e8b976e289`
if that merge still completes wait without required `ci / ci` or still
creates a roster PR when the exact open carrier exists. Every changed
authoritative fixture file is byte-identical to that merge.

## VOC-142-AC-08 — Current-state docs match the live wait and reuse contract

- Requirement source: `VOC-142-D10`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-07`
- Evidence: `VOC-142-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim that two
stable zero-pending snapshots complete roster wait or that reconcile always
opens a new roster PR. `AGENTS.md`, fixture README, `adopt.yml` header
comments, and every other current-state match state that wait requires the
complete required set including `ci / ci` and that reconcile reuses a
matching open or already-merged carrier. Clearly marked historical VOC-141
records remain historical. VOC-141, VOC-140, and VOC-139 package directories
are not rewritten.

## VOC-142-AC-09 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-142-D11`, `VOC-142-D12`, `VOC-142-D13`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-08`
- Evidence: `VOC-142-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite,
`git diff --check`, and independent exact-revision review that binds the live
head all pass. `roles.yml` is unchanged. Evidence does not require a commit
to contain its own SHA.

## VOC-142-AC-10 — Documented reconcile can resume #1112; no snapshot-gap

- Requirement source: `VOC-142-D09`, `VOC-142-D15`, `VOC-142-D16`
- Tasks: `VOC-142-T00`
- Tests: `VOC-142-TEST-09`
- Evidence: `VOC-142-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No duplicate
VOC-141 task, roster PR, promotion PR, or release audit is created. #1112 is
not manually merged, closed, recreated, or bypassed by this package. After
exact-SHA review and merge into `develop`, ordinary `reconcile` for plan PR
#1110 can reuse #1112 when it still matches, wait for complete required
checks, merge, clean up, and dispatch the first not-yet-dispatched task
exactly once. Root issue #1113 closes only after allowlisted metadata from a
successful adopt/reconcile run exists. Closed state alone is not completion
proof. Named incident run/job/PR/SHA IDs remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
