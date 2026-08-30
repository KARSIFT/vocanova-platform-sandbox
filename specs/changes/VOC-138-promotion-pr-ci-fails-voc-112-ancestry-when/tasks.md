# VOC-138 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1091 is one
promotion-CI repair outcome: select merge-base/hash-bound `pr-validation` for
a same-repository `main` <- `develop` promotion when the historical VOC-112
subject is unreachable, keep ordinary PR `pr-ancestry` fail-closed, stop
recovery from rerunning a structurally doomed `pull_request` job, pin the
new infrastructure merge, update current-state docs, and let
`reconcile-release` for #1089 merge #1090. Repository count,
provenance-versus-recovery, workflow-versus-tests-versus-docs, and
pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure #167
(`b263c0c110591cc798b89277dfc35542abb1597b`) is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-138 carrier from current `develop`. Do not
reuse PR #1090 as this package's implementation PR.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Promotion of already-completed
work through #1090 is evidence of this outcome after the repair lands.

## VOC-138-T00 — Unblock promotion PR CI and exact-head recovery when the VOC-112 subject is unreachable

- Requirement source: issue #1091; `VOC-138-D00` through `VOC-138-D16`
- Acceptance criteria: `VOC-138-AC-00` through `VOC-138-AC-09`
- Tests: `VOC-138-TEST-00` through `VOC-138-TEST-13`
- Evidence: `VOC-138-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1091 evidence in `t00-evidence.md` (PR #1090, head
   `87f0efcb94a213a0ede9fdbca94a707a22d42b86`, base
   `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`, run `33122154521`, jobs
   `98691441027` / `98692552949`, tests `VOC-112-TEST-12` /
   `VOC-112-TEST-13`, subject `f9d11e232a07c7d7a9c433d02c9267912543ba10`,
   dispatch `33122158425`, reconcile-release `33122099253` /
   `33122436137`, `selected_required_run_mismatch`, pin
   `b263c0c110591cc798b89277dfc35542abb1597b`). Do not copy raw provider
   responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. Retain protected
   comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` for the eight
   VOC-112 no-change paths. This package's own plan/adoption/roster commits
   after issue-creation develop `87f0efcb…` do not count as protected-file
   drift.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Add an explicit promotion signal
   on reusable `ci.yml` only for same-repository `main` <- `develop`. In
   `run-app-checks.sh`, that authenticated signal always selects
   `pr-validation` with exact PR SHAs, independent of subject-object
   availability. Do not fetch. Keep `--squash-safe-push` for generic non-PR
   events. Ordinary fixture add/modify/delete without the signal stays
   `pr-ancestry`.
4. Change exact-head recovery so it does not rerun a structurally doomed PR
   job or attest a weaker same-head dispatch. Resolve and bind immutable PR
   number/base/head/repository/branch-pair/workflow metadata, dispatch or
   select a recovery execution that actually runs `pr-validation`, wait for
   genuine success, and only then publish the equivalent required result.
   Update caller `.github/workflows/pipeline.yml` in the same T00 if required
   to pass this metadata. Treat run `33122158425` as incident evidence only.
   Keep `selected_required_run_mismatch` for all missing or mismatched
   identity or semantic evidence. Do not fabricate unbacked statuses.
5. Add deterministic tests for the #1090 class, ordinary missing-subject
   fail-closed, resolvable-but-non-ancestor promotion, hash/SHA negatives, no
   `git fetch`, doomed-job rerun refusal, and weaker-dispatch rejection.
   Obtain independent exact-revision review of the infra PR and
   merge it. Record the exact merge SHA.
6. From current caller `develop`, create a new VOC-138 implementation branch.
   Pin `PINNED_SHA.txt` to that infra merge and mirror every changed
   authoritative fixture file. Update fixture README,
   `docs/operations/11-devops-and-ci-cd.md`, and
   `docs/development/agent-skills.md`. Do not rewrite VOC-135/VOC-136/VOC-137
   package records. Do not edit the eight VOC-112 no-change paths.
7. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no capture-commit fetch helper, hydrate helper, provenance-mode
   wrapper, or test-time evidence mutation was added.
8. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   mode-selection change, recovery change, negative-case results, pin
   advance, validation commands after commit, and the feasible exact-head
   binding contract. Evidence must not require a commit to contain its own
   SHA.
9. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`;
   - targeted VOC-138 provenance and recovery cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates the promotion missing-subject case, ordinary
     `pr-ancestry` retention, hash/SHA negatives, no-fetch constraint, and
     recovery behavior.
10. This package's caller PR `Closes` only its own VOC-138 task issue. Do
    not snapshot the current develop/main gap. After the exact reviewed
    caller merge, `reconcile-release` for #1089 may merge #1090 (or the live
    promotion at the then-current `develop` head) and converge `develop` to
    that exact merge SHA.
11. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Fetching or hydrating `f9d11e23…`.
- Switching the promotion PR application check to `--squash-safe-push`.
- Weakening ordinary fixture-changing `pr-ancestry`.
- Changing any of the eight VOC-112 no-change paths.
- Adding a capture-commit fetch helper, hydrate helper, provenance-mode
  wrapper, or test-time evidence mutation under any filename.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting provenance, recovery, pin, tests, docs, or release
  reconciliation into separate tasks.
- Rewriting VOC-112 through VOC-137 package records.
- Operator-owned live evidence contracts.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, tests, docs, evidence, and
  release handoff are one repair outcome.
- Infrastructure #167 is already merged and is the defective live contract;
  consume a new independently reviewed infra merge inside this same task.
- Merging #1090 is ordinary release after this correction, not a second
  VOC-138 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
