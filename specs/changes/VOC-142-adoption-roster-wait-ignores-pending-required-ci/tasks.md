# VOC-142 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1113 is one
adoption-recovery repair outcome: wait until the complete required roster-check
set including `ci / ci` is SUCCESS, reuse a matching open or already-merged
roster PR on reconcile, pin the new infrastructure merge, update current-state
docs, and let ordinary `reconcile` for plan PR #1110 resume #1112. Repository
count, wait-versus-reuse, workflow-versus-tests-versus-docs, and pin-versus-
release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-142 carrier from current `develop`. Do not
reuse PR #1110, #1112, or task #1111 as this package's implementation
artifacts.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Interrupted VOC-141 adoption
via #1112 is incident evidence that this repair must unblock; it is not a
second VOC-142 task. A VOC-097 live-evidence second task is not used.

## VOC-142-T00 — Wait for the complete required roster-check set and reuse the exact open roster PR on reconcile

- Requirement source: issue #1113; `VOC-142-D00` through `VOC-142-D16`
- Acceptance criteria: `VOC-142-AC-00` through `VOC-142-AC-10`
- Tests: `VOC-142-TEST-00` through `VOC-142-TEST-08`
- Post-T00 release/closure verification: `VOC-142-TEST-09` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-142-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1113 evidence in `t00-evidence.md` (plan PR #1110, plan
   head `178ae81ed4f0224fbb12359c90ddb67c6687ca9a`, plan merge
   `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2`, pin
   `67bdfd13ef875dead23ce4be01d7d0e8b976e289`, adoption run `33343125733`,
   job `99342230038`, task #1111, roster PR #1112, roster head
   `98dd0936a73b64a6b548da6cf2000a6d000917ac`, roster pipeline
   `33343147453`, `ci / ci` job `99342299218`, wait SUCCESS at
   `2026-08-30T23:58:46Z` while `ci / ci` remained IN_PROGRESS until
   `2026-08-30T23:59:16Z`, reconcile run `33343250178` / job `99342577393`
   failed at `Open roster PR`). Do not copy raw provider responses,
   credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits after issue-creation plan merge `bb4ffdf…`
   do not count as protected-file drift. Do not edit
   `specs/changes/VOC-141-…/`.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Implement D01–D07 there:
   roster wait must require the complete ruleset-required set including
   `ci / ci`; a partial stable green snapshot must not complete wait;
   IN_PROGRESS required rows are not-ready; `Open roster PR` must reuse an
   exact matching OPEN carrier; create only when no matching OPEN or
   already-merged carrier exists; reject mismatched/ambiguous carriers;
   keep existing-task reuse and single root dispatch. Do not use
   `statusCheckRollup` or `gh pr checks`. Do not treat in-progress parents
   as attestable SUCCESS for merge-gate or release. Do not manually merge
   #1112.
4. Add deterministic tests that call the live wait and open-PR paths.
   Cover late registration of required `ci / ci`, continued IN_PROGRESS
   `ci / ci` with other required rows already SUCCESS, exact open-PR reuse,
   already-merged reuse, mismatched/ambiguous rejection, and duplicate
   task/dispatch suppression. Do not treat YAML-only `stable_green_count`
   assertions as covering #1113.
5. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA. From current caller `develop`, create a new
   VOC-142 implementation branch. Pin `PINNED_SHA.txt` to that infra merge
   and mirror every changed authoritative fixture file. Update `AGENTS.md`,
   fixture README, `adopt.yml` header comments, and all current-state docs
   found by exhaustive tracked-source search. Do not rewrite VOC-141,
   VOC-140, or VOC-139 package records. Reconcile every live caller
   pin-lock/current-pin assertion and mirrored-file hash table with the new
   merge. Preserve issue-era `AUTHORITATIVE_PIN` values and historical
   package evidence; split current versus authoritative constants where a
   test currently conflates them.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no fabricated-status helper, bypass-actor addition, App-token mint
   change, or test-time evidence mutation was added. Confirm no helper was
   added that merges, closes, or recreates #1112.
7. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   wait-completeness change, reuse change, negative-case results,
   source-search disposition, pin advance, validation commands after
   commit, and the feasible exact-head binding contract. Evidence must not
   require a commit to contain its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - targeted VOC-142 wait-completeness, open-PR reuse, already-merged
     reuse, mismatch/ambiguous rejection, and duplicate-suppression cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates complete-required-set wait including `ci / ci`,
     fail-closed partial green snapshots, exact open-PR reuse,
     already-merged reuse, mismatched/ambiguous rejection, duplicate
     task/dispatch suppression, current-state docs, and pin advance.
9. This package's caller PR `Closes` only its own VOC-142 task issue. Do
   not snapshot the current develop/main gap. After the exact reviewed
   caller merge, ordinary `reconcile` for plan PR #1110 may resume #1112
   when that PR still matches. Root issue #1113 closes only after
   allowlisted metadata from a successful adopt/reconcile run exists. Do
   not create a duplicate VOC-141 task, roster PR, promotion PR, or release
   audit.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging a roster PR including #1112.
- Switching the roster wait to `statusCheckRollup` or `gh pr checks`.
- Treating an in-progress parent as attestable SUCCESS for merge-gate or
  release.
- Creating a duplicate VOC-141 task, roster PR, promotion PR, or release
  audit.
- Repairing VOC-141's promotion-recovery hang (issue #1109) as this
  package's implementation work.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting wait, reuse, pin, tests, docs, or release reconciliation into
  separate tasks.
- Rewriting VOC-141, VOC-140, or VOC-139 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, complete-required-set wait,
  carrier reuse, tests, docs, evidence, and release handoff are one repair
  outcome.
- Infrastructure merge `67bdfd13…` is already merged and is the defective
  live contract; consume a new independently reviewed infra merge inside this
  same task.
- Interrupted VOC-141 adoption (#1111 / #1112 / reconcile `33343250178`) is
  incident proof, not a second VOC-142 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
