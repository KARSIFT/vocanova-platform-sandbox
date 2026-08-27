# VOC-130 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1047 is one
caller pin outcome: consume exact infrastructure #165, restore shared policy
after caller checkout in both release jobs, and let ordinary release complete
VOC-129 plus this blocker. Repository count, workflow-versus-tests-versus-docs,
and fixture/pin work are not split reasons. VOC-129 promotion is evidence of
this outcome, not an additional VOC-130 task and not a second VOC-129 attempt.

Cross-repo note: infrastructure PR KARSIFT/karsift-ai-infra#165 is already
merged as `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. T00 does not open a
replacement infra PR. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repo's tracked tree. Caller fixture/pin, tests,
docs, and evidence land in this repository under the same task. No bootstrap
exception: VOC-124 already enables nested workflow-file publication; T00's
first run is attempt `1` on a new VOC-130 carrier.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`).

## VOC-130-T00 — Pin exact infra #165 and restore shared policy after caller checkout

- Requirement source: issue #1047; `VOC-130-D00` through `VOC-130-D09`
- Acceptance criteria: `VOC-130-AC-00` through `VOC-130-AC-05`
- Tests: `VOC-130-TEST-00` through `VOC-130-TEST-08`
- Evidence: `VOC-130-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1047 evidence in `t00-evidence.md` (VOC-129 caller PR
   #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346`, release run
   `33066533397`, missing `task-completion-runner.py` after caller root
   checkout, safe no-op, skipped converge, authoritative merge
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, reviewed head
   `e33931d02f7bdbb094ae8177fd88324cd19ac5ce`, current develop pin
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`).
2. Mirror every in-scope infrastructure file needed by the caller from exact
   merge `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` into
   `tooling/governance/fixtures/karsift-ai-infra/`, including `release.yml`
   with `Restore shared lifecycle policy after caller checkout` in both
   `identify` and `converge`, and the #165 tests those steps require. Set
   `PINNED_SHA.txt` to that exact merge.
3. Add deterministic caller regressions proving both jobs restore the exact
   shared-policy revision after caller checkout and before
   `task-completion-runner.py` (`validate-task` / `validate-roster`), using
   `job.workflow_repository`, `job.workflow_sha`, and `path: karsift-ai-infra`.
   Advance matching `scripts/foundation/*` pin literals and any governance
   tests that currently hard-code `863fc1f…` in the same task.
4. Preserve #164 contracts: missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, review/implementer separation, retry bounds,
   `reconcile-production-change`, App-token isolation, sanitized raw-error
   controls, Cursor Composer implementer, Cursor Grok review, no OpenAI,
   unchanged `config/roles.yml`. Never bundle secrets. Never print credential
   values.
5. Update current-state fixture README/comments so they name pin `8ce2b77…`
   and the post-caller-checkout restore. Do not rewrite historical CHANGELOG
   or A-003/VOC-075/VOC-127/VOC-129 audit records.
6. Live `.github/workflows/pipeline.yml` changes only if comparing the #165
   merge with the live caller dispatch surface proves they are required.
   Expected: no live pipeline edit. Keep every `workflow_dispatch` block at
   most 25 inputs. Do not add operator SHA inputs.
7. This package's caller PR `Closes` only its own VOC-130 task issue. Do not
   re-implement VOC-129. Do not snapshot the current develop/main gap.
8. After the exact reviewed caller pin is merged, record in `t00-evidence.md`
   the ordinary release/promotion/sync/closure evidence for VOC-129 and this
   package. Hosted `release.yml@main` may already contain the #165 restore
   before this pin merges; `reconcile-release` for VOC-129 may succeed
   independently. That does not remove this pin/test/doc obligation.
9. Run applicable validation and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`
     if the mirrored fixture suite is runnable in the caller tree;
   - targeted foundation tests if those files change;
   - `git diff --check`;
   - exact pin SHA `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`;
   - any narrower targeted commands added by the implementation.
10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. Do not weaken review, risk
    classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, `roles.yml`, or monitor-inventory changes.
- Re-implementing infrastructure #165 or opening a replacement infra PR.
- Re-implementing VOC-129 or manufacturing a second VOC-129 completion
  marker.
- Snapshotting the current develop/main gap, or promoting "current develop"
  to `main` as this package's work.
- Fast-forwarding `main` instead of preserving `--merge` commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Implementing VOC-122, merging unrelated carriers, or rewriting historical
  CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 records.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, fixture/pin, or evidence into
  separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded VOC-129 / VOC-130 promotion handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the #165 pin, restore coverage, preserved #164 contracts,
  docs, and VOC-129 / VOC-130 promotion handoff are one repair outcome.
- Infrastructure #165 is already merged; consume it. Do not wait on a new
  infra PR.
- VOC-129 promotion may proceed through hosted `release.yml@main` as soon as
  that reusable workflow contains the #165 restore. Do not add a second
  VOC-130 task to wait on that event.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
