# VOC-125 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1020 is one
existing-carrier recovery-identity outcome. Coordinated caller and
infrastructure pull requests remain one task; repository count,
workflow-versus-tests-versus-docs, and fixture/pin work are not split reasons.
Resuming the existing VOC-122 carrier is evidence of this outcome, not a
second VOC-125 task.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for implement recovery
identity and the project-repo pipeline template, and changes this caller's
`.github/workflows/pipeline.yml`. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree. Caller
fixture/pin, tests, and evidence land in this repository under the same task.
Infra PRs must say `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and
MUST NOT use a closing keyword. No bootstrap exception: VOC-124 already
enables nested workflow-file publication; T00's first run is attempt `1`.

## VOC-125-T00 — Bind existing-carrier resume identity for operator implement dispatch

- Requirement source: issue #1020; `VOC-125-D00` through `VOC-125-D09`
- Acceptance criteria: `VOC-125-AC-00` through `VOC-125-AC-06`
- Tests: `VOC-125-TEST-00` through `VOC-125-TEST-07`
- Evidence: `VOC-125-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1020 live failure in `t00-evidence.md` (task #1003,
   pipeline run `32966618512` / job `98170418081`, existing branch
   `agent/voc-122-voc-122-t00@0b7be8c531be8300d5a1d5534acc83bf4d6a1791`,
   prior-review base `e910eb4a21d48bbb5b3e0c30b8ee647d64683dbe`, caller
   `develop@f0f7a54283c3527f40544fd236516dc3a5f4dc82`, infrastructure
   `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`, dispatch `attempt=2` with no
   recovery identity, stop at `Create implementation branch` before model
   resolution).
2. In `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml`, add
   `existing_pr_number` and a fail-closed bind step that runs before
   `Create implementation branch` and before model resolution, per
   `VOC-125-D02` / `VOC-125-D03`. Reuse or extend
   `config/verify-expected-head.py` rather than inventing a second SHA
   dialect. Do not print credential values.
3. In caller `.github/workflows/pipeline.yml` and
   `karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml`,
   add implement-only `existing_pr_number` and forward it on the `implement`
   job. Do not add operator-typed SHA inputs.
4. In `remediate.yml` retry, keep forwarding event-derived
   `expected_head_sha` / `expected_base_sha` and also forward `pr_number` as
   `existing_pr_number`.
5. Preserve VOC-121/VOC-123/VOC-124 fail-closed contracts: exact base/head
   binding, nested isolation, named-ref source bundle, credential-free
   bundles, `publish-source` App-token isolation with workflow-write only on
   that mint, caller workflow-file refusal, force-with-lease, two-attempt
   bound, Cursor Composer implementer, Cursor Grok review. Never bundle
   secrets. Never allow attempt `3` or attempt-`1` rewrite of an existing
   open carrier.
6. Add deterministic source and caller tests that prove:
   - valid attempt-2 resume with `existing_pr_number` binds exact head/base
     and reuses the existing carrier;
   - the #1020 empty-binding class fails closed;
   - every `VOC-125-D03` mismatch class fails closed;
   - `remediate.yml` still forwards SHAs and now forwards `existing_pr_number`;
   - caller/template pipeline expose and forward `existing_pr_number` only;
   - attempt `1` with an existing open carrier fails closed;
   - publication/isolation/role contracts remain.
   Do not mint real App tokens or use secrets or production data.
7. Update current-state comments/docs (`implement.yml`, `remediate.yml`,
   pipeline comments, `karsift-ai-infra/README.md`) so they describe operator
   resume as attempt `2` plus `existing_pr_number`. Do not rewrite historical
   CHANGELOG or A-003/VOC-075 audit records. Update `AGENTS.md` only if an
   existing sentence would become false.
8. Land the infrastructure change through the normal coordinated source
   carrier. Merge the independently reviewed infra PR first so
   `implement.yml@main` declares `existing_pr_number` before the caller
   forwards it. Pin `tooling/governance/fixtures/karsift-ai-infra/` to that
   exact merge SHA when the fixture consumes the change. Update caller
   fixture regressions and any `scripts/foundation/*` pin literals that still
   assert `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`, in the same task.
9. After the exact reviewed infra merge is live on `implement.yml@main` and
   the caller dispatch contract is merged, record in `t00-evidence.md` that
   existing `VOC-122-T00` / #1003 / #1012 is resumed through the repaired
   route at attempt `2` with `existing_pr_number=1012`, not replaced. Do not
   implement VOC-122 promotion-recovery replan here and do not merge #1012
   from this package.
10. Run applicable validation and record results in `t00-evidence.md`:
    - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
      `KARSIFT/karsift-ai-infra` checkout;
    - `bash scripts/governance/validate-governance.sh`;
    - `bash scripts/governance/classify-change-risk.sh`;
    - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
    - targeted foundation pin tests if those files change;
    - `git diff --check`;
    - exact reviewed infra SHA and pin applicability;
    - any narrower targeted commands added by the implementation.
11. Preserve independent exact-SHA review for each carrier, risk
    classification, protected checks, and App-token isolation. Do not weaken
    review, risk classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, or monitor-inventory changes.
- Implementing VOC-122 promotion-recovery replan inside this package, creating
  a replacement VOC-122 task or PR, or merging #1012.
- Adding operator-typed SHA inputs to caller `workflow_dispatch`.
- Allowing attempt `3` or rewriting an existing carrier as attempt `1`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG / A-003 / VOC-075 records, or expanding
  AGENTS.md into a new implement-dispatch runbook.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded #1003 / #1012 resume handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the live defect, the recovery-identity contract, the tests, the
  docs, the caller pin, and the existing VOC-122 resume handoff are one
  operator-resume outcome.
- Infra should merge first so `implement.yml@main` declares
  `existing_pr_number` before the caller forwards that input.
- Resume existing #1003 / #1012 only after that exact infra merge and the
  caller dispatch contract are live. Do not treat VOC-125 caller-pin merge as
  VOC-122 completion.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
