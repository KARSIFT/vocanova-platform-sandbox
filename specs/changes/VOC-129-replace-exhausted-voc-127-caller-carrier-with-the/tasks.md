# VOC-129 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1042 is one
caller replacement outcome: consume exact infrastructure #164, recreate the
intended VOC-127 caller contract from current `develop`, and supersede
unpublishable PR #1041. Repository count, workflow-versus-tests-versus-docs,
and fixture/pin work are not split reasons. Closing #1041 / #1039 / #1035 is
evidence of this outcome, not an additional VOC-129 task and not a third
VOC-127 attempt.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#164 is already
merged as `863fc1f35b1d35e4981a59166b0e939be1a2b681`. T00 does not open a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repo's tracked tree. Caller fixture/pin, live
pipeline, tests, docs, and evidence land in this repository under the same
task. No bootstrap exception: VOC-124 already enables nested workflow-file
publication; T00's first run is attempt `1` on a new VOC-129 carrier.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-129-T00 — Recreate the VOC-127 caller contract from current develop and pin exact infra #164

- Requirement source: issue #1042; `VOC-129-D00` through `VOC-129-D09`
- Acceptance criteria: `VOC-129-AC-00` through `VOC-129-AC-06`
- Tests: `VOC-129-TEST-00` through `VOC-129-TEST-09`
- Evidence: `VOC-129-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1042 evidence in `t00-evidence.md` (VOC-127 task #1039,
   origin #1035, caller PR #1041 at
   `9d459813a733f9c6d58ad3352df0db27d33ee7f4`, pipeline run `33058603158`,
   artifact `implement-VOC-127-VOC-127-T00-attempt2` id `9640920826`, bundle
   SHA-256 `724e5547e29283b2701b70b72e000253a5100edf77545807610fce17151f7906`,
   artifact head `bbdb93aadec461830435771490f5c79ba524fed9`, stale pin
   `a9df74a63976d5239b84151fd01310835c999e7c`, authoritative merge
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`, consumed VOC-127 retry, current
   develop pin `60afda3a44fd06b8c00b219771de7112f1aded6e`).
2. Recreate the complete intended VOC-127 caller diff from current `develop`.
   Do not compose or publish #1041, the attempt-2 artifact, or unpublished
   head `9d459813a733f9c6d58ad3352df0db27d33ee7f4`.
3. Mirror every in-scope infrastructure file needed by the caller from exact
   merge `863fc1f35b1d35e4981a59166b0e939be1a2b681` into
   `tooling/governance/fixtures/karsift-ai-infra/`, including `release.yml`,
   `reconcile-production-change.yml`, `release-checkout-ref-runner.py`,
   branch-sync helpers, the project-repo `pipeline.yml` template, and the
   #164 tests those files require. Set `PINNED_SHA.txt` to that exact merge.
4. Update live `.github/workflows/pipeline.yml` to the #164 dispatch
   identity: add `reconcile-production-change`, forward `issue_number` as
   task authority for that action, make `release` need
   `[merge-gate, reconcile-production-change]`, and make `auto-advance` wait
   on a successful production-change reconcile for issue-closed wakes. Do not
   add operator SHA inputs. Do not overload `existing_pr_number`. Keep every
   `workflow_dispatch` block at most 25 inputs. Preserve live
   `live_evidence_mode` unless the new action itself requires touching it.
5. Prove that a tree-equivalent `develop` synchronization does not schedule
   an unnecessary staging deployment. Retain VOC-111 path selection for
   actual runtime/deploy tree changes. If `on.push.paths` is insufficient for
   merge-commit fast-forwards, add a fail-closed job-level skip in
   `deploy-staging.yml` without broadening the allowlist. Extend
   `scripts/foundation/voc111-deploy-staging-paths.test.mjs` (or an adjacent
   foundation test in the same task) accordingly.
6. Add deterministic caller regressions for exact pin consistency with
   `863fc1f35b1d35e4981a59166b0e939be1a2b681` and inequality with
   `a9df74a6…` / `60afda3a…`, and for the #164 checkout-ref
   ordering/missing-`develop` path. Update caller tests that currently
   hard-code the `pipeline.yml` action-options list (including
   `test_voc126_workflow_dispatch_input_limit.py`) so they expect
   `reconcile-production-change` without dropping VOC-126's 25-input bound.
   Advance matching `scripts/foundation/*` pin literals in the same task.
7. Preserve VOC-113 through VOC-126 contracts: roster markers, exact-head
   checks, recovery, App-token isolation, match-head-commit, two-attempt
   bound, Cursor Composer implementer, Cursor Grok review, no OpenAI,
   unchanged `config/roles.yml`. Never bundle secrets. Never print credential
   values. Never allow VOC-127 attempt `3`.
8. Update current-state comments/docs: live `pipeline.yml`, fixture
   README/comments as needed, `AGENTS.md` reconcile-release and
   release-authority text, `docs/operations/11-devops-and-ci-cd.md`,
   `docs/operations/10-development-workflow.md`, and current-state
   branch/release paragraphs in
   `docs/operations/15-ai-native-product-and-engineering-operating-model.md`
   and `docs/governance/16-autonomous-development-operating-model.md` so they
   describe exact-merge-SHA develop sync, `reconcile-release` retry of that
   sync, and exceptional `reconcile-production-change`. Do not rewrite
   historical CHANGELOG or A-003/VOC-075/VOC-127 audit records.
9. This package's caller PR `Closes` only its own VOC-129 task issue. Do not
   merge #1041. Do not dispatch VOC-127-T00 as attempt `3`.
10. After the exact reviewed caller replacement is merged and promoted, record
    in `t00-evidence.md` that #1041 is closed as superseded (never merged),
    that #1039 and #1035 are closed with audit comments naming the VOC-129
    merge, and that no VOC-127 completion marker was manufactured from #1041.
11. Run applicable validation and record results in `t00-evidence.md`:
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`
      if the mirrored fixture suite is runnable in the caller tree;
    - targeted foundation tests if those files change (including VOC-111 path
      selection and pin literals);
    - `git diff --check`;
    - exact pin SHA `863fc1f35b1d35e4981a59166b0e939be1a2b681`;
    - any narrower targeted commands added by the implementation.
12. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. Do not weaken review, risk
    classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes, except
  the staging skip for tree-equivalent develop-sync.
- Re-implementing infrastructure #164 or opening a replacement infra PR.
- Publishing, merging, or composing #1041 / the attempt-2 artifact.
- Dispatching VOC-127-T00 as attempt `3`.
- Snapshotting the current develop/main gap, or promoting "current develop"
  to `main` as this package's work.
- Fast-forwarding `main` instead of preserving `--merge` commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`; reviving
  `reconcile-main-to-develop`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Implementing VOC-122, merging unrelated carriers, or rewriting historical
  CHANGELOG / A-003 / VOC-075 / VOC-127 records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, fixture/pin, or evidence into
  separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded #1041 / #1039 / #1035 handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the exhausted VOC-127 caller, the #164 pin, live dispatch,
  tests, docs, and the #1041 / #1039 / #1035 handoff are one replacement
  outcome.
- Infrastructure #164 is already merged; consume it. Do not wait on a new
  infra PR.
- Close #1041 / #1039 / #1035 only after this package's exact reviewed caller
  merge is promoted. Do not treat VOC-129 caller-pin merge as a VOC-127
  completion marker bound to #1041.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
