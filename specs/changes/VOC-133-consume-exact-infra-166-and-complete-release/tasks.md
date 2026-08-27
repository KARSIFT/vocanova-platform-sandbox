# VOC-133 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1061 is one
caller replacement outcome: consume exact infrastructure #166 from current
`develop` with a complete VOC-112 no-change boundary, restore shared policy
after caller checkout in both release jobs, land the #166 nested-checkout
helper-lifetime contract, and let ordinary release complete VOC-129,
exhausted VOC-130, exhausted VOC-131, superseded VOC-132, and this blocker.
Repository count, workflow-versus-tests-versus-docs, fixture/pin work, and
the VOC-112 no-change boundary are not split reasons. VOC-129 / VOC-130 /
VOC-131 / VOC-132 promotion and closure are evidence of this outcome, not
additional VOC-133 tasks and not a further attempt on those exhausted or
unpublished carriers.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#166 is already
merged as `f3d79177bf8a9abe0dae550f39502165d494c576`. T00 does not open a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repo's tracked tree. Caller fixture/pin, tests,
docs, complete VOC-112 no-change boundary, hash-based #166 byte identity, and
evidence land in this repository under the same task. No bootstrap exception:
VOC-124 already enables nested workflow-file publication; T00's first run is
attempt `1` on a new VOC-133 carrier from current `develop`. Do not reuse PR
#1051 or PR #1056. Do not redispatch VOC-132-T00 (#1059).

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-133-T00 — Pin exact infra #166 from current develop with a complete VOC-112 no-change boundary

- Requirement source: issue #1061; `VOC-133-D00` through `VOC-133-D13`
- Acceptance criteria: `VOC-133-AC-00` through `VOC-133-AC-07`
- Tests: `VOC-133-TEST-00` through `VOC-133-TEST-11`
- Evidence: `VOC-133-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1061 evidence in `t00-evidence.md` (VOC-132 run
   `33079499176` / job `98543693163`, bounded `cp: cannot stat
   karsift-ai-infra/config/run-app-checks.sh`, no caller branch/PR/completion
   marker, prohibition on redispatched VOC-132-T00 #1059, exhausted VOC-130
   task #1049 / PR #1051, exhausted VOC-131 PR #1056, VOC-129 caller PR #1046
   at `429d8c6d49303148ca1cc14dba5f6768a7863346`, original release run
   `33066533397`, authoritative merge
   `f3d79177bf8a9abe0dae550f39502165d494c576`, failed infra head
   `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f`, PASS infra head
   `1488619d0d37aaa179d8e739bfe931881d6c51aa`, current develop pin
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112 `subject_revision`
   `f9d11e232a07c7d7a9c433d02c9267912543ba10`). Do not copy raw provider
   responses, credentials, or the full log.
2. From current `develop`, create a new VOC-133 implementation branch. Do not
   reuse, merge, modify, or compose PR #1051, PR #1056, or their branches.
   Do not redispatch VOC-132-T00 (#1059). Do not re-check out or rewrite
   VOC-129 PR #1046.
3. Fetch KARSIFT/karsift-ai-infra at exact
   `f3d79177bf8a9abe0dae550f39502165d494c576` into an untracked
   workspace-relative nested checkout. Mirror every necessary changed
   authoritative file from that SHA into
   `tooling/governance/fixtures/karsift-ai-infra/`, byte-for-byte:
   `.github/workflows/implement.yml`, `.github/workflows/release.yml`,
   `config/implementer_nested_checkout.py`, `tests/test_release_policy.py`,
   `tests/test_voc121_implement_policy.py`,
   `tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md`. Set
   `PINNED_SHA.txt` to that exact merge. Compute SHA-256 of the mirrored files
   and commit those hashes into the caller regression. Remove the nested
   checkout before commit so it cannot stage as a gitlink. Do not leave a
   `/tmp` checkout as a test-time dependency. Do not overwrite the
   caller-owned fixture `README.md` with the infrastructure repository README.
4. Add deterministic caller regressions proving both release jobs restore the
   exact shared-policy revision after caller checkout and before
   `task-completion-runner.py` (`validate-task` / `validate-roster`), using
   `job.workflow_repository`, `job.workflow_sha`, `path: karsift-ai-infra`,
   and `persist-credentials: false`. Add regressions for #166 helper copy
   before the unrestricted model, absent-nested-checkout continuation, and
   fail-closed symlink / non-directory / parent-Git inheritance. Advance
   matching `scripts/foundation/*` pin literals and any governance tests that
   currently hard-code `863fc1f…` in the same task.
5. Add a deterministic regression that fails if any of the five VOC-112
   no-change paths differs from the carrier's `develop` base or appears in
   the implementation diff:
   - `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
   - `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
   - `scripts/foundation/voc112-navigation-benchmark.test.mjs`
   - `AGENTS.md`
   - `.agents/skills/vocanova-repo-navigator/SKILL.md`
   Confirm JSON `subject_revision` remains
   `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Confirm the provenance test
   still fail-closes `local` mode on a missing capture commit in a full
   checkout. Do not retarget, recapture, or weaken those files.
6. Preserve #164 contracts: missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, review/implementer separation, retry bounds,
   `reconcile-production-change`, App-token isolation, sanitized raw-error
   controls, Cursor Composer implementer, Cursor Grok review, no OpenAI,
   unchanged `config/roles.yml`. Never bundle secrets. Never print credential
   values.
7. Update current-state fixture README/comments so they name pin `f3d791…`,
   the post-caller-checkout restore, and the post-implementer helper-lifetime
   contract. Do not rewrite historical caller CHANGELOG or A-003 / VOC-075 /
   VOC-127 / VOC-129 / VOC-130 / VOC-131 / VOC-132 audit records.
8. Live `.github/workflows/pipeline.yml` changes only if comparing the #166
   merge with the live caller dispatch surface proves they are required.
   Expected: no live pipeline edit. Keep every `workflow_dispatch` block at
   most 25 inputs. Do not add operator SHA inputs.
9. This package's caller PR `Closes` only its own VOC-133 task issue. Do not
   re-implement VOC-129, VOC-130, VOC-131, or VOC-132. Do not manufacture a
   completion marker for those packages. Do not snapshot the current
   develop/main gap.
10. Record in `t00-evidence.md` the exact reviewed implementation head, the
    final infra merge, the confirmed SHA-256 hashes, the protected-path
    base/head equality results, and — after merge — the ordinary
    release/promotion/sync/closure evidence for VOC-129, VOC-130, VOC-131,
    VOC-132, and this package. Evidence must not claim a protected-path
    revert unless that path is absent from the diff. Hosted
    `implement.yml@main` / `release.yml@main` may already contain the #166 /
    #165 repairs before this pin merges; `reconcile-release` for VOC-129 may
    succeed independently. That does not remove this pin/test/doc obligation.
11. Run applicable validation and record results in `t00-evidence.md`:
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`
      if the mirrored fixture suite is runnable in the caller tree;
    - targeted foundation tests if those files change (`voc097`, `voc104`,
      `voc108`, and the VOC-112 no-change-boundary test added in this task);
    - exact pin SHA `f3d79177bf8a9abe0dae550f39502165d494c576`;
    - recorded SHA-256 identity of the mirrored #166 files;
    - five-path VOC-112 base/head equality and diff absence;
    - `git diff --check`;
    - any narrower targeted commands added by the implementation.
12. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. Do not weaken review, risk
    classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Re-implementing infrastructure #165 or #166, or opening a replacement
  infra PR.
- Silently retargeting VOC-132 from pin #165 to pin #166.
- Redispatching VOC-132-T00 (#1059).
- Reusing, merging, modifying, or composing PR #1051 or PR #1056.
- Re-implementing VOC-129, VOC-130, VOC-131, or VOC-132, or manufacturing a
  completion marker for those packages.
- Changing any of the five VOC-112 no-change paths, including weakening
  provenance `local` mode or editing `AGENTS.md` / the navigator skill for
  this pin.
- Snapshotting the current develop/main gap, or promoting "current develop"
  to `main` as this package's work.
- Fast-forwarding `main` instead of preserving `--merge` commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Implementing VOC-122, merging unrelated carriers, or rewriting historical
  caller CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 / VOC-131 /
  VOC-132 records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 no-change
  boundary, or evidence into separate tasks.
- Depending on a machine-specific `/tmp` checkout as the test-time source of
  #166 bytes.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded VOC-129 / VOC-130 / VOC-131 /
  VOC-132 / VOC-133 promotion handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the #166 pin, restore coverage, nested-checkout coverage,
  complete VOC-112 no-change boundary, hash-based byte identity, preserved
  #164 contracts, docs, and VOC-129 / VOC-130 / VOC-131 / VOC-132 / VOC-133
  promotion handoff are one repair outcome.
- Infrastructure #166 is already merged; consume it. Do not wait on a new
  infra PR.
- VOC-129 promotion may proceed through hosted `release.yml@main` as soon as
  that reusable workflow contains the #165 restore. Do not add a second
  VOC-133 task to wait on that event.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
