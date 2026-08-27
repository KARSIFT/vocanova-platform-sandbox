# VOC-132 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1057 is one
caller replacement outcome: consume exact infrastructure #165 from current
`develop` with a complete VOC-112 no-change boundary, restore shared policy
after caller checkout in both release jobs, and let ordinary release complete
VOC-129, exhausted VOC-130, exhausted VOC-131, and this blocker. Repository
count, workflow-versus-tests-versus-docs, fixture/pin work, and the VOC-112
no-change boundary are not split reasons. VOC-129 / VOC-130 / VOC-131
promotion and closure are evidence of this outcome, not additional VOC-132
tasks and not a further attempt on those exhausted carriers.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#165 is already
merged as `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. T00 does not open a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repo's tracked tree. Caller fixture/pin, tests,
docs, complete VOC-112 no-change boundary, hash-based #165 byte identity, and
evidence land in this repository under the same task. No bootstrap exception:
VOC-124 already enables nested workflow-file publication; T00's first run is
attempt `1` on a new VOC-132 carrier from current `develop`. Do not reuse PR
#1051 or PR #1056.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-132-T00 — Pin exact infra #165 from current develop with a complete VOC-112 no-change boundary

- Requirement source: issue #1057; `VOC-132-D00` through `VOC-132-D12`
- Acceptance criteria: `VOC-132-AC-00` through `VOC-132-AC-06`
- Tests: `VOC-132-TEST-00` through `VOC-132-TEST-10`
- Evidence: `VOC-132-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1057 evidence in `t00-evidence.md` (exhausted VOC-130
   task #1049 / PR #1051, exhausted VOC-131 PR #1056 at
   `c11454e717a6d778143de1f2023acc4480305845`, exact-review comment
   5439802080, supervisor comment 5439768013, reproduction merge-base
   `790ea1b66b096414fe6709e6cd4b8342ff5ac587`, VOC-129 caller PR #1046 at
   `429d8c6d49303148ca1cc14dba5f6768a7863346`, release run `33066533397`,
   missing `task-completion-runner.py` after caller root checkout, safe
   no-op, skipped converge, authoritative merge
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, reviewed infra head
   `e33931d02f7bdbb094ae8177fd88324cd19ac5ce`, current develop pin
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112 `subject_revision`
   `f9d11e232a07c7d7a9c433d02c9267912543ba10`).
2. From current `develop`, create a new VOC-132 implementation branch. Do not
   reuse, merge, modify, or compose PR #1051, PR #1056, or their branches.
   Do not re-check out or rewrite VOC-129 PR #1046. Inspecting #1056 as a
   reference for in-scope pin/fixture/restore files is allowed; copying the
   provenance-test edit or false-revert evidence is not.
3. Fetch KARSIFT/karsift-ai-infra at exact
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` into an untracked workspace-relative
   nested checkout. Mirror only the necessary changed authoritative files
   from that SHA into `tooling/governance/fixtures/karsift-ai-infra/`.
   Expected pair: `.github/workflows/release.yml` and
   `tests/test_release_policy.py`, byte-for-byte. Set `PINNED_SHA.txt` to that
   exact merge. Compute SHA-256 of the mirrored files and commit those hashes
   into the caller regression. Remove the nested checkout before commit so it
   cannot stage as a gitlink. Do not leave a `/tmp` checkout as a test-time
   dependency.
4. Add deterministic caller regressions proving both jobs restore the exact
   shared-policy revision after caller checkout and before
   `task-completion-runner.py` (`validate-task` / `validate-roster`), using
   `job.workflow_repository`, `job.workflow_sha`, `path: karsift-ai-infra`,
   and `persist-credentials: false`. Advance matching `scripts/foundation/*`
   pin literals and any governance tests that currently hard-code `863fc1f…`
   in the same task.
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
7. Update current-state fixture README/comments so they name pin `8ce2b77…`
   and the post-caller-checkout restore. Do not rewrite historical CHANGELOG
   or A-003/VOC-075/VOC-127/VOC-129/VOC-130/VOC-131 audit records.
8. Live `.github/workflows/pipeline.yml` changes only if comparing the #165
   merge with the live caller dispatch surface proves they are required.
   Expected: no live pipeline edit. Keep every `workflow_dispatch` block at
   most 25 inputs. Do not add operator SHA inputs.
9. This package's caller PR `Closes` only its own VOC-132 task issue. Do not
   re-implement VOC-129, VOC-130, or VOC-131. Do not manufacture a VOC-129,
   VOC-130, or VOC-131 completion marker. Do not snapshot the current
   develop/main gap.
10. Record in `t00-evidence.md` the exact reviewed implementation head, the
    confirmed SHA-256 hashes, the protected-path base/head equality results,
    and — after merge — the ordinary release/promotion/sync/closure evidence
    for VOC-129, VOC-130, VOC-131, and this package. Evidence must not claim
    a protected-path revert unless that path is absent from the diff. Hosted
    `release.yml@main` may already contain the #165 restore before this pin
    merges; `reconcile-release` for VOC-129 may succeed independently. That
    does not remove this pin/test/doc obligation.
11. Run applicable validation and record results in `t00-evidence.md`:
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`
      if the mirrored fixture suite is runnable in the caller tree;
    - targeted foundation tests if those files change (`voc097`, `voc104`,
      `voc108`, and the VOC-112 no-change-boundary test added in this task);
    - exact pin SHA `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`;
    - recorded SHA-256 identity of the mirrored #165 files;
    - byte-identity of all five VOC-112 no-change paths against the carrier
      `develop` base;
    - `git diff --check`;
    - any narrower targeted commands added by the implementation.
12. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. Do not weaken review, risk
    classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Re-implementing infrastructure #165 or opening a replacement infra PR.
- Reusing, merging, modifying, or composing PR #1051 or PR #1056.
- Re-implementing VOC-129, VOC-130, or VOC-131, or manufacturing a
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
  CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 / VOC-131
  records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 no-change
  boundary, or evidence into separate tasks.
- Depending on a machine-specific `/tmp` checkout as the test-time source of
  #165 bytes.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded VOC-129 / VOC-130 / VOC-131 /
  VOC-132 promotion handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the #165 pin, restore coverage, complete VOC-112 no-change
  boundary, hash-based byte identity, preserved #164 contracts, docs, and
  VOC-129 / VOC-130 / VOC-131 / VOC-132 promotion handoff are one repair
  outcome.
- Infrastructure #165 is already merged; consume it. Do not wait on a new
  infra PR.
- VOC-129 promotion may proceed through hosted `release.yml@main` as soon as
  that reusable workflow contains the #165 restore. Do not add a second
  VOC-132 task to wait on that event.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
