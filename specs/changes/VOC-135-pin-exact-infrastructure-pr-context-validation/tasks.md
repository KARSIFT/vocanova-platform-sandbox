# VOC-135 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1071 is one
caller replacement outcome: consume exact infrastructure #167 from current
`develop` with immutable PR-context validation, a complete eight-path
no-change boundary, no VOC-112 hydration bypasses, feasible exact-revision
evidence, restore shared policy after caller checkout in both release jobs,
land the #166 nested-checkout helper-lifetime contract, and let ordinary
release complete VOC-129 plus exhausted / superseded VOC-127 and VOC-130
through VOC-134. Repository count, workflow-versus-tests-versus-docs,
fixture/pin work, the VOC-112 no-change boundary, hydration-bypass
regression, package.json identity, feasible evidence, and release
reconciliation are not split reasons. VOC-127 / VOC-130 / VOC-131 / VOC-132 /
VOC-133 / VOC-134 promotion and closure are evidence of this outcome, not
additional VOC-135 tasks and not a further attempt on those exhausted or
unpublished carriers.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#167 is already
merged as `b263c0c110591cc798b89277dfc35542abb1597b`. T00 does not open a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repo's tracked tree. Caller fixture/pin, tests,
docs, complete eight-path no-change boundary, complete-diff hydration/bypass
scan, package.json identity, hash-based #167 byte identity, feasible
exact-revision evidence, and evidence land in this repository under the same
task. No bootstrap exception: VOC-124 already enables nested workflow-file
publication; T00's first run is attempt `1` on a new VOC-135 carrier from
current `develop`. Do not reuse PR #1051, PR #1056, PR #1065, or PR #1070.
Do not redispatch VOC-132-T00 (#1059), VOC-133-T00 (#1063), or VOC-134-T00
(#1068).

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-135-T00 — Pin exact infra #167 from current develop with immutable PR-context validation and no VOC-112 hydration bypasses

- Requirement source: issue #1071; `VOC-135-D00` through `VOC-135-D16`
- Acceptance criteria: `VOC-135-AC-00` through `VOC-135-AC-10`
- Tests: `VOC-135-TEST-00` through `VOC-135-TEST-14`
- Evidence: `VOC-135-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1071 evidence in `t00-evidence.md` (exhausted VOC-134-T00
   #1068 / PR #1070, attempt-1 head `f373d9a61f90659e982173395a96ff54776fc945`,
   attempt-2 head `718bfdcdaec65e5df6e918fbc25f0fe64f6a2cad`, truthful
   missing-subject local fail-closed message, prohibition on redispatched
   VOC-134-T00 #1068 and reuse of PR #1070, exhausted VOC-130 / VOC-131 /
   VOC-132 / VOC-133 carriers, VOC-129 caller PR #1046 at
   `429d8c6d49303148ca1cc14dba5f6768a7863346`, authoritative merge
   `b263c0c110591cc798b89277dfc35542abb1597b`, PASS infra head
   `eb11c4fc6841ec73816e2e064dcd449d98c1e933`, current develop pin
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-134 unmerged target
   `f3d79177bf8a9abe0dae550f39502165d494c576`, expected immutable carrier
   base `b9e74fc2db4691c48c637639b265d527de9f4505`, VOC-112
   `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`). Do not
   copy raw provider responses, credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Expected SHA is `b9e74fc2db4691c48c637639b265d527de9f4505`. Record
   that SHA as the immutable carrier base at PR creation. If live `develop`
   differs, fail closed: do not implement against a moved tip. Do not treat
   a moving branch name as authority after that constant is selected.
3. From current `develop`, create a new VOC-135 implementation branch. Do not
   reuse, merge, cherry-pick, modify, or compose PR #1051, PR #1056, PR #1065,
   PR #1070, or their branches. Do not redispatch VOC-132-T00 (#1059),
   VOC-133-T00 (#1063), or VOC-134-T00 (#1068). Do not re-check out or rewrite
   VOC-129 PR #1046.
4. Fetch KARSIFT/karsift-ai-infra at exact
   `b263c0c110591cc798b89277dfc35542abb1597b` into an untracked
   workspace-relative nested checkout. Mirror every necessary changed
   authoritative file from that SHA into
   `tooling/governance/fixtures/karsift-ai-infra/`, byte-for-byte:
   `.github/workflows/ci.yml`, `.github/workflows/implement.yml`,
   `.github/workflows/release.yml`, `config/run-app-checks.sh`,
   `config/implementer_nested_checkout.py`, `tests/test_app_check_context.py`,
   `tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`,
   `tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md`. Set
   `PINNED_SHA.txt` to that exact merge. Compute SHA-256 of the mirrored files
   and commit those hashes into the caller regression. Remove the nested
   checkout before commit so it cannot stage as a gitlink. Do not leave a
   `/tmp` checkout as a test-time dependency. Do not overwrite the
   caller-owned fixture `README.md` with the infrastructure repository README.
5. Add deterministic caller regressions proving both release jobs restore the
   exact shared-policy revision after caller checkout and before
   `task-completion-runner.py`; #166 helper copy before the unrestricted
   model and nested-checkout classifier behavior; and #167 immutable
   PR-context validation (`--pr-base-sha` / `--pr-head-sha`, `pr-validation`
   vs `pr-ancestry`, fail-closed comparison errors, no evidence fetch, CI
   `fetch-depth: 0`, implementer integration-anchor plus live HEAD including
   after self-correction). Advance matching `scripts/foundation/*` pin
   literals and any governance tests that currently hard-code `863fc1f…` in
   the same task.
6. Add a deterministic regression bound to the immutable carrier-base SHA
   that fails if any of the eight no-change paths differs from that SHA or
   appears in the implementation diff. The test must require the exact commit
   object and all eight paths to resolve and must fail—not skip—if proof
   cannot run. Confirm JSON `subject_revision` remains
   `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Confirm the provenance test
   still fail-closes `local` mode on a missing capture commit in a full
   checkout. Do not retarget, recapture, or weaken those files.
7. Add a complete caller-diff scan of changed executable paths that fails if
   any new or changed caller script invokes Git fetch for capture subjects,
   sets VOC-112 provenance variables around tests, hydrates evidence objects,
   or makes local fail-closed unreachable. Do not rely on banning only
   `hydrate-voc112-git-objects.mjs`. Assert `package.json`,
   `validate-workspace.mjs`, and `voc112-navigation-benchmark-run.mjs` remain
   byte-identical to the carrier base.
8. Preserve #164 contracts: missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, review/implementer separation, retry bounds,
   `reconcile-production-change`, App-token isolation, sanitized raw-error
   controls, Cursor Composer implementer, Cursor Grok review, no OpenAI,
   unchanged `config/roles.yml`. Never bundle secrets. Never print credential
   values.
9. Update current-state fixture README/comments so they name pin `b263c0c…`,
   the post-caller-checkout restore, the post-implementer helper-lifetime
   contract, and the immutable PR-context contract. Do not rewrite historical
   caller CHANGELOG or A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 /
   VOC-131 / VOC-132 / VOC-133 / VOC-134 audit records.
10. Live `.github/workflows/pipeline.yml` changes only if comparing the #167
    merge with the live caller dispatch surface proves they are required.
    Expected: no live pipeline edit. Keep every `workflow_dispatch` block at
    most 25 inputs. Do not add operator SHA inputs.
11. This package's caller PR `Closes` only its own VOC-135 task issue. Do not
    re-implement VOC-127, VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, or
    VOC-134. Do not manufacture a completion marker for those packages. Do
    not snapshot the current develop/main gap.
12. Record in `t00-evidence.md` the immutable carrier-base SHA, the final
    infra merge, the confirmed SHA-256 hashes, the protected-path base/head
    equality results, the complete-diff scan result, the feasible exact-head
    binding contract, and — after merge — the ordinary
    release/promotion/sync/closure evidence for VOC-127 through VOC-134 and
    this package. Evidence must not require a commit to contain its own SHA.
13. Run applicable validation and record results in `t00-evidence.md`:
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`
      if the mirrored fixture suite is runnable in the caller tree;
    - targeted foundation tests if those files change (`voc097`, `voc104`,
      `voc108`, exact #167 pin/hash tests, immutable carrier-base eight-path
      boundary tests, the complete-diff hydration/bypass scan, and the
      `package.json` carrier-base-identity assertion);
    - exact pin SHA `b263c0c110591cc798b89277dfc35542abb1597b`;
    - recorded SHA-256 identity of the mirrored #167 files;
    - eight-path equality against the immutable carrier-base SHA and diff
      absence;
    - `git diff --check`;
    - any narrower targeted commands added by the implementation.
14. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. Do not weaken review, risk
    classification, required checks, or automatic-merge gates. The
    independent review comment must bind the live PR head exactly.
    Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Re-implementing infrastructure #165, #166, or #167, or opening a
  replacement infra PR.
- Silently retargeting VOC-132, VOC-133, or VOC-134 pins in place.
- Redispatching VOC-132-T00 (#1059), VOC-133-T00 (#1063), or VOC-134-T00
  (#1068).
- Reusing, merging, cherry-picking, modifying, or composing PR #1051,
  PR #1056, PR #1065, or PR #1070.
- Re-implementing VOC-127, VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, or
  VOC-134, or manufacturing a completion marker for those packages.
- Changing any of the eight no-change paths, including weakening provenance
  `local` mode or editing `AGENTS.md` / the navigator skill / the runner /
  `validate-workspace.mjs` / `package.json` for this pin.
- Adding a capture-commit fetch helper, hydrate/materialize helper,
  provenance-mode wrapper, import side effect, environment override,
  evidence-stamping helper, skip, or test-time evidence mutation under any
  filename.
- Requiring a commit to contain its own SHA.
- Using a moving branch ref as the no-change authority after the immutable
  carrier-base SHA is selected, or implementing against a moved `develop` tip.
- Snapshotting the current develop/main gap, or promoting "current develop"
  to `main` as this package's work.
- Fast-forwarding `main` instead of preserving `--merge` commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Implementing VOC-122, merging unrelated carriers, or rewriting historical
  caller CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 / VOC-131 /
  VOC-132 / VOC-133 / VOC-134 records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 no-change
  boundary, hydration-bypass regression, package.json identity, feasible
  evidence, or release reconciliation into separate tasks.
- Depending on a machine-specific `/tmp` checkout as the test-time source of
  #167 bytes.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded VOC-127 through VOC-135 promotion
  handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the #167 pin, PR-context coverage, restore coverage,
  nested-checkout coverage, complete eight-path no-change boundary,
  complete-diff hydration/bypass scan, package.json identity, hash-based byte
  identity, feasible exact-revision evidence, preserved #164 contracts, docs,
  and VOC-127 through VOC-135 promotion handoff are one repair outcome.
- Infrastructure #167 is already merged; consume it. Do not wait on a new
  infra PR.
- VOC-129 promotion may proceed through hosted `release.yml@main` as soon as
  that reusable workflow contains the #165 restore. Do not add a second
  VOC-135 task to wait on that event.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
