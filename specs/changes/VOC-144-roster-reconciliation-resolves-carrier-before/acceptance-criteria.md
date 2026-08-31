# VOC-144 — Acceptance Criteria

## VOC-144-AC-00 — Post-push REST lag is not a durable mismatched carrier

- Requirement source: `VOC-144-D01`, `VOC-144-D02`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-00`
- Evidence: `VOC-144-EV-00`
- Result: pending

After `Push roster branch`, when exactly one OPEN PR exists for the
deterministic `karsift/roster-*` head, this repository, the integration-branch
base, and a listed head SHA that later equals the locally pushed exact SHA,
adopt reuses that PR (`reuse_open`) and does not call `gh pr create`. The
class of reconcile runs `33437239322` and `33437514152` cannot recur.

## VOC-144-AC-01 — Wait is scoped to transient SHA lag only

- Requirement source: `VOC-144-D02`, `VOC-144-D03`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-00`, `VOC-144-TEST-02`
- Evidence: `VOC-144-EV-00`
- Result: pending

The bounded wait lives in the GitHub adapter. `resolve_roster_carrier`
remains a single-snapshot identity predicate. Wait runs only when exactly
one OPEN PR matches repository, head ref, and base ref and its listed SHA
differs from local HEAD. Zero same-ref OPEN PRs may create. Ambiguous,
wrong-repository, and wrong-base carriers do not enter the SHA-lag wait.

## VOC-144-AC-02 — Finite timeout; stable SHA mismatch remains fail-closed

- Requirement source: `VOC-144-D04`, `VOC-144-D06`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-01`
- Evidence: `VOC-144-EV-00`
- Result: pending

Timeout and poll interval are named constants. The timeout ceiling is 60
seconds. After the bound, a still-different SHA on an otherwise matching
OPEN PR is `MISMATCHED_OPEN_CARRIER` and does not create a second PR. The
wait does not use VOC-141's 1,800-second recovery timeout. Tests inject
sequenced snapshots or a fake clock.

## VOC-144-AC-03 — Exact-SHA identity is not weakened; no duplicate PR

- Requirement source: `VOC-144-D05`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-00`, `VOC-144-TEST-01`, `VOC-144-TEST-03`
- Evidence: `VOC-144-EV-00`
- Result: pending

Reuse requires the listed PR head SHA to equal local `git rev-parse HEAD`
as a 40-character hex SHA. A still-stale listed head is not reused. No
second roster PR is created for the same deterministic branch. #1112 is not
closed, recreated, or retargeted.

## VOC-144-AC-04 — Durable mismatches and API failure fail closed without SHA-lag wait

- Requirement source: `VOC-144-D06`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-02`, `VOC-144-TEST-04`
- Evidence: `VOC-144-EV-00`
- Result: pending

Two or more OPEN same-ref PRs, a single OPEN PR whose repository or base
ref does not match, invalid identity inputs, and GitHub API or parse
failure fail closed without waiting for SHA convergence. Unrelated PRs are
not closed, recreated, or retargeted.

## VOC-144-AC-05 — VOC-142 wait-completeness and non-lag carrier paths remain

- Requirement source: `VOC-144-D06`, `VOC-144-D17`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-03`
- Evidence: `VOC-144-EV-00`
- Result: pending

An exact first-snapshot OPEN match still reuses immediately. An
already-merged exact carrier is not duplicated. Zero matches may create
exactly one new PR. Existing tasks are reused and root dispatch happens
exactly once. Roster wait still requires the complete required set
including `ci / ci`. Merge-gate and release still do not treat an
in-progress parent as attestable completed `ci / ci`.

## VOC-144-AC-06 — Caller pin and mirrored fixture bytes match the new infra merge

- Requirement source: `VOC-144-D08`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-05`
- Evidence: `VOC-144-EV-00`
- Result: pending

`PINNED_SHA.txt` equals the independently reviewed infrastructure merge that
contains this repair, not leftover `8993e867640dfb604dec0466c4e0787e68d8e258`
if that merge still fails `MISMATCHED_OPEN_CARRIER` on the first post-push
snapshot. Every changed authoritative fixture file is byte-identical to that
merge.

## VOC-144-AC-07 — Current-state docs match the live convergence contract

- Requirement source: `VOC-144-D10`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-06`
- Evidence: `VOC-144-EV-00`
- Result: pending

An exhaustive tracked-source search identifies every current claim that
carrier resolution is a single immediate REST snapshot or that a same-ref
SHA difference is always a durable mismatch with no wait. `AGENTS.md`,
fixture README, `adopt.yml` header comments, and every other current-state
match state that reconcile waits boundedly for the existing same-ref OPEN
PR to expose the exact pushed SHA, then reuses it, and still fails closed
on a stable mismatch. Clearly marked historical VOC-142/VOC-141 records
remain historical. Those package directories are not rewritten.

## VOC-144-AC-08 — Deterministic suites and exact-SHA review pass

- Requirement source: `VOC-144-D11`, `VOC-144-D12`, `VOC-144-D13`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-07`
- Evidence: `VOC-144-EV-00`
- Result: pending

After the repair is tracked and committed, governance validation, R4
classification, caller governance suite, mirrored infra suite,
`git diff --check`, and independent exact-revision review that binds the live
head all pass. `roles.yml` is unchanged. Evidence does not require a commit
to contain its own SHA.

## VOC-144-AC-09 — Documented reconcile can resume #1112; no snapshot-gap

- Requirement source: `VOC-144-D09`, `VOC-144-D15`, `VOC-144-D16`
- Tasks: `VOC-144-T00`
- Tests: `VOC-144-TEST-08`
- Evidence: `VOC-144-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. No duplicate
VOC-141 task, roster PR, promotion PR, or release audit is created. #1112 is
not manually merged, closed, recreated, or bypassed by this package. After
exact-SHA review and merge into `develop`, ordinary `reconcile` for plan PR
#1110 can reuse #1112 when it still matches, including across post-push REST
lag, wait for complete required checks, merge, clean up, and dispatch the
first not-yet-dispatched task exactly once. Root issue #1122 closes only
after allowlisted metadata from a successful adopt/reconcile run exists.
Closed state alone is not completion proof. Named incident run/PR/SHA IDs
remain audit evidence.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
