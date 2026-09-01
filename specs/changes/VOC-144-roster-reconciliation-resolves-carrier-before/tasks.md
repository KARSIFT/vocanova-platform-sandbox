# VOC-144 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1122 is one
adoption-recovery repair outcome: after a roster-branch push, boundedly wait
for the existing same-ref OPEN PR to expose the locally pushed exact SHA,
reuse that carrier, pin the new infrastructure merge, update current-state
docs, and let ordinary `reconcile` for plan PR #1110 resume #1112.
Repository count, adapter-versus-resolver, workflow-versus-tests-versus-docs,
and pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`8993e867640dfb604dec0466c4e0787e68d8e258` is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-144 carrier from current `develop`. Do not
reuse PR #1110, #1112, or task #1111 as this package's implementation
artifacts.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Interrupted VOC-141 adoption
via #1112 is incident evidence that this repair must unblock; it is not a
second VOC-144 task. A VOC-097 live-evidence second task is not used.

## VOC-144-T00 — Boundedly wait for existing roster-PR head metadata to expose the pushed SHA, then reuse that carrier

- Requirement source: issue #1122; `VOC-144-D00` through `VOC-144-D17`
- Acceptance criteria: `VOC-144-AC-00` through `VOC-144-AC-09`
- Tests: `VOC-144-TEST-00` through `VOC-144-TEST-07`
- Post-T00 release/closure verification: `VOC-144-TEST-08` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-144-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1122 evidence in `t00-evidence.md` (plan PR #1110,
   roster PR #1112, branch `karsift/roster-voc-141`, pin
   `8993e867640dfb604dec0466c4e0787e68d8e258`, run `33437239322` pushed
   `958e0fedf742173320bb89cfe690ec7070b49e93` then `MISMATCHED_OPEN_CARRIER`
   with later `reuse_open` on that head, run `33437514152` pushed
   `0206cb70437cb751a19a2e715d8202b672060b50` then the same class). Do not
   copy raw provider responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits do not count as protected-file drift. Do not
   edit `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, or
   `VOC-139-…/`.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Implement D01–D07 and D17
   there: the GitHub adapter boundedly re-fetches the unique same-repository /
   same-head-ref / same-base OPEN PR until listed head equals local HEAD or
   the named bound is exhausted; `resolve_roster_carrier` stays a
   single-snapshot identity predicate; durable mismatches and API failure do
   not wait; timeout still-stale remains `MISMATCHED_OPEN_CARRIER`; no
   duplicate PR; VOC-142 complete-required-set wait is unchanged. Named
   timeout ceiling 60 seconds. Do not manually merge #1112.
4. Add deterministic tests that call the live adapter path with injected
   snapshots or a fake clock. Cover stale-then-converge → `reuse_open`,
   timeout-still-stale → `MISMATCHED_OPEN_CARRIER` without `gh pr create`,
   durable mismatch / API failure without wait, and preserved VOC-142
   first-snapshot reuse / already-merged / create-when-zero. Do not treat
   YAML-only or single-snapshot-only assertions as covering #1122.
5. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA. From current caller `develop`, create a new
   VOC-144 implementation branch. Pin `PINNED_SHA.txt` to that infra merge
   and mirror every changed authoritative fixture file. Update `AGENTS.md`,
   fixture README, `adopt.yml` header comments, and all current-state docs
   found by exhaustive tracked-source search. Do not rewrite VOC-142,
   VOC-141, VOC-140, or VOC-139 package records. Reconcile every live caller
   pin-lock/current-pin assertion and mirrored-file hash table with the new
   merge. Preserve issue-era `AUTHORITATIVE_PIN` values and historical
   package evidence; split current versus authoritative constants where a
   test currently conflates them.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no fabricated-status helper, bypass-actor addition, App-token mint
   change, or test-time evidence mutation was added. Confirm no helper was
   added that merges, closes, or recreates #1112. Confirm roster wait still
   requires the complete required set including `ci / ci`.
7. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   named timeout and poll-interval constants, convergence-wait change,
   negative-case results, source-search disposition, pin advance, validation
   commands after commit, and the feasible exact-head binding contract.
   Evidence must not require a commit to contain its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - targeted VOC-144 stale-then-converge, timeout-still-stale,
     durable-mismatch-does-not-wait, and preserved VOC-142 carrier cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates bounded post-push SHA-lag wait, stale-then-converge
     reuse, timeout-still-stale fail-closed, durable mismatch without wait,
     unchanged exact-SHA identity, no duplicate PR, current-state docs, and
     pin advance.
9. This package's caller PR `Closes` only its own VOC-144 task issue. Do
   not snapshot the current develop/main gap. After the exact reviewed
   caller merge, ordinary `reconcile` for plan PR #1110 may resume #1112
   when that PR still matches. Root issue #1122 closes only after
   allowlisted metadata from a successful adopt/reconcile run exists. Do
   not create a duplicate VOC-141 task, roster PR, promotion PR, or release
   audit.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Weakening exact-SHA identity, adding bypass actors, fabricating statuses,
  or manually merging a roster PR including #1112.
- Weakening VOC-142 complete-required-set wait, switching it to
  `statusCheckRollup` or `gh pr checks`, or treating an in-progress parent
  as attestable SUCCESS for merge-gate or release.
- Creating a duplicate VOC-141 task, roster PR, promotion PR, or release
  audit.
- Repairing VOC-141's promotion-recovery hang (issue #1109) as this
  package's implementation work.
- Unbounded sleep, a 1,800-second recovery-style timeout, or wall-clock
  GitHub-lag tests.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting adapter wait, resolver, pin, tests, docs, or release
  reconciliation into separate tasks.
- Rewriting VOC-142, VOC-141, VOC-140, or VOC-139 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, bounded SHA-lag wait,
  unchanged identity predicate, tests, docs, evidence, and release handoff
  are one repair outcome.
- Infrastructure merge `8993e867…` is already merged and is the defective
  live contract; consume a new independently reviewed infra merge inside this
  same task.
- Interrupted VOC-141 adoption (#1112 / reconcile `33437239322` /
  `33437514152`) is incident proof, not a second VOC-144 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
