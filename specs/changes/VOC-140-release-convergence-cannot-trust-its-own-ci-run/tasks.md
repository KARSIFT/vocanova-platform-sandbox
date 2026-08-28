# VOC-140 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1102 is one
release-convergence repair outcome: never trust a still-running release
carrier as recovered `ci / ci`, make the merge App identity able to prove
the production merge guard, pin the new infrastructure merge, update
current-state docs, and let `reconcile-release` for #1089 merge #1090.
Repository count, CI-identity-versus-token-contract, workflow-versus-tests-
versus-docs, and pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`599436835371f27fac52ec6b47a18b36257366ac` is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-140 carrier from current `develop`. Do not
reuse PR #1090 as this package's implementation PR.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Promotion of already-completed
work through #1090 is evidence of this outcome after the repair lands. A
VOC-097 live-evidence second task is not used: an evidence-carrier PR cannot
satisfy sha_lineage against promotion PR #1090's recovery HEAD.

## VOC-140-T00 — Unblock release-convergence CI identity and production-merge-guard App-token visibility

- Requirement source: issue #1102; `VOC-140-D00` through `VOC-140-D16`
- Acceptance criteria: `VOC-140-AC-00` through `VOC-140-AC-09`
- Tests: `VOC-140-TEST-00` through `VOC-140-TEST-12`
- Evidence: `VOC-140-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1102 evidence in `t00-evidence.md` (PR #1090, head
   `21eef75549226766fc4f78f62f232ee5fbdb8d6d`, base
   `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`, pin
   `599436835371f27fac52ec6b47a18b36257366ac`, circular-CI run
   `33136633666`, job `98738317266`, `untrusted_ci_recovery_identity`,
   dedicated recovery `33136865709`, release runs `33136984634` and
   `33137091931`, jobs `98739074178` and `98739420310`,
   `production-merge-guard: production_merge_guard_missing` versus
   administrator-visible `ok ruleset_id=20575146`). Do not copy raw provider
   responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits after issue-creation develop `21eef755…` do
   not count as protected-file drift.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Implement D01–D08 there:
   never select an in-progress/failed release carrier as attestable `ci / ci`;
   require dedicated completed `promotion-pr-validation PR #<n>` when no
   completed non-carrier run exists; request the least-privilege mint
   permission that actually returns `bypass_actors` as an array on the
   release merge identity and the production-branch merge-gate path; fail
   distinctly when that field is omitted. Capture live App-token versus
   administrator-visible ruleset JSON at this revision before choosing the
   mint permission. Do not add bypass actors, fabricate statuses, or rerun
   doomed `pull_request` `ci / ci` jobs.
4. Add deterministic tests for the #1102 circular-CI class, dedicated
   promotion-pr-validation dispatch/selection, omitted-`bypass_actors`
   payload, empty-bypass acceptance, non-empty-bypass rejection, and a
   subprocess of the real verifier with a mock `gh`. Do not treat helper-only
   full-fixture tests as covering the token-visible shape.
5. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA. From current caller `develop`, create a new
   VOC-140 implementation branch. Pin `PINNED_SHA.txt` to that infra merge
   and mirror every changed authoritative fixture file. Update fixture
   README and `docs/operations/11-devops-and-ci-cd.md`. Do not rewrite
   VOC-139 or VOC-138 package records. Reconcile every live caller
   pin-lock/current-pin assertion and mirrored-file hash table with the new
   merge. Preserve issue-era `AUTHORITATIVE_PIN` values and historical
   package evidence; split current versus authoritative constants where a
   test currently conflates them.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no fabricated-status helper, bypass-actor addition, or test-time
   evidence mutation was added.
7. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   recovery-identity change, token/API contract including the diagnosed mint
   permission and any required operator App-setting action, negative-case
   results, pin advance, validation commands after commit, and the feasible
   exact-head binding contract. Evidence must not require a commit to contain
   its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - targeted VOC-140 circular-CI and token-visible payload cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates the circular-CI identity case, dedicated
     promotion-pr-validation requirement, omitted-`bypass_actors` token
     shape, empty-bypass acceptance, non-empty-bypass rejection,
     no-fabricated-status constraint, and pin advance.
9. This package's caller PR `Closes` only its own VOC-140 task issue. Do
   not snapshot the current develop/main gap. After the exact reviewed
   caller merge, rerun dedicated promotion recovery if necessary, then
   `reconcile-release` for #1089 may merge #1090 (or the live promotion at
   the then-current `develop` head) and converge `develop` to that exact
   merge SHA. Root issue #1102 closes only after allowlisted metadata from
   the successful recovery/release run exists.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging #1090.
- Creating a duplicate promotion PR or release audit issue.
- Switching the promotion PR application check to `--squash-safe-push`.
- Rerunning doomed `pull_request` `ci / ci` jobs as a strategy.
- Changing GitHub App installation permissions inside the implementation PR
  except as the documented D07 operator action if the mint requests a
  permission the installation lacks.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting CI-identity, token contract, pin, tests, docs, or release
  reconciliation into separate tasks.
- Rewriting VOC-139 or VOC-138 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, recovery identity, guard
  token/API, tests, docs, evidence, and release handoff are one repair
  outcome.
- Infrastructure merge `59943683…` is already merged and is the defective
  live contract; consume a new independently reviewed infra merge inside this
  same task.
- Merging #1090 is ordinary release after this correction, not a second
  VOC-140 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
