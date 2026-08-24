# VOC-115 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because the required work is one
coherent governance objective and issue #962 explicitly requires one-task default
planning unless a real split boundary exists.

## VOC-115-T00 — Consolidate package/task defaults, causal-remediation policy, validation, and regression coverage

- Requirement source: issue #962; `VOC-115-D00` through `VOC-115-D11`
- Acceptance criteria: `VOC-115-AC-00` through `VOC-115-AC-06`
- Tests: `VOC-115-TEST-00` through `VOC-115-TEST-09`
- Evidence: `VOC-115-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update canonical policy/docs so they state the new default clearly:
   - one coherent objective normally maps to one package;
   - one new package normally maps to one end-to-end task;
   - in-scope causal remediation may stay with the active package/carrier only when
     it stays within the original objective, acceptance criteria, risk ceiling, and
     protected-area scope;
   - unrelated or authority-expanding work still requires a new issue/plan.
   - one coherent outcome normally maps to one outcome-sized implementation PR;
   - task IDs are minimum-sufficient outcome traceability groupings;
   - `L`, 800 changed lines, file/component/skill count, and code-vs-tests-vs-docs
     are review signals, not automatic split rules.
2. Replace DOC-15's seven-task authentication example with one end-to-end task and
   reconcile `docs/operations/10-development-workflow.md` so fixed size thresholds
   no longer command task/PR splitting.
3. Update the primary shared-infra planner prompt/validation/tests and the pinned
   caller fixture so normal coherent work drafts one
   task by default rather than splitting code, tests, docs, or same-carrier evidence
   into separate tasks for convenience.
4. Update plan-review guidance and any deterministic validation needed so every task
   after the first records an explicit split reason from the allowed boundary set:
   hard merge-order dependency, independently releasable/rollbackable unit, distinct
   owner/authority/risk boundary, mutually exclusive execution environment,
   post-merge evidence that cannot be produced by the implementation carrier, or
   a change too large for reliable single-PR review.
5. Ensure packages with more than three tasks are treated as exceptional and require
   explicit package-level justification for why consolidation is unsafe.
6. Preserve compatibility with justified multi-task packages: adoption still opens
   ordered task issues / dependency edges correctly and later tasks continue to
   advance sequentially only after predecessor completion proof.
7. Add deterministic regression coverage for:
   - ordinary coherent feature/bug -> one task;
   - adding several related skills + configuration + docs + tests -> one task;
   - code + tests + docs remain in the same task by default;
   - missing or invalid split reason fails closed;
   - justified multi-task package retains correct sequencing;
   - in-scope causal remediation remains under the active package while unrelated or
     authority-expanding work requires a new plan.
8. Update template wording and any mirrored fixture/tests in the same task so no
   canonical source still implies the old fragmentation default after merge.
9. Record changed files, commands, and results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - any narrower targeted governance/fixture regression command added by the
     implementation;
   - `git diff --check`.

### Explicitly out of scope for this task

- Weakening exact-SHA review, risk floors, protected-branch checks, or fail-closed
  behavior.
- Letting unrelated bugs ride along under an active package.
- Allowing in-scope remediation to expand the original risk ceiling or protected
  scope without a new plan.
- Retroactively re-rostering historical packages only to fit the new default.
- Product runtime, credential, deploy, monitor, or database changes.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary from
  `VOC-115-D03` is required at draft time.
- If implementation discovers that a real split boundary is necessary, that change
  must be justified in the package itself under the new split-reason rules rather
  than assumed from file type or review convenience.

Tasks preserve scope, separation of duties, and rollback safety.
