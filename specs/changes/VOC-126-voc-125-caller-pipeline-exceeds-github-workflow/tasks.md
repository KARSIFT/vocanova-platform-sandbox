# VOC-126 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1025 is one
caller-dispatch validity outcome. Coordinated caller and infrastructure pull
requests remain one task; repository count, workflow-versus-tests-versus-docs,
and fixture/pin work are not split reasons. Superseding unusable PR #1024 and
handing off VOC-125 / VOC-122 are evidence of this outcome, not additional
VOC-126 tasks.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for the project-repo
pipeline template, the dedicated verifier workflow template, and source tests,
and changes this caller's `.github/workflows/pipeline.yml` plus the matching
dedicated verifier workflow. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree. Caller
fixture/pin, tests, and evidence land in this repository under the same task.
Infra PRs must say `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and
MUST NOT use a closing keyword. No bootstrap exception: VOC-124 already
enables nested workflow-file publication; T00's first run is attempt `1`.

## VOC-126-T00 — Relocate read-only verifier dispatch so VOC-125 existing_pr_number can land under GitHub's 25-input limit

- Requirement source: issue #1025; `VOC-126-D00` through `VOC-126-D09`
- Acceptance criteria: `VOC-126-AC-00` through `VOC-126-AC-06`
- Tests: `VOC-126-TEST-00` through `VOC-126-TEST-07`
- Evidence: `VOC-126-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1025 live failure in `t00-evidence.md` (VOC-125 task
   #1022, origin #1020, caller PR #1024 at
   `8621f12dd466edab37fddb86d4e5e0a348ed3609`, infrastructure merge
   `1f1705dbad41729563b0ad1e878e4154e5511e93`, Actions run `32977045898`,
   GitHub 25-input annotation, 26 template inputs, consumed VOC-125 retry,
   downstream VOC-122 #1003 / #1012 still waiting).
2. In `KARSIFT/karsift-ai-infra/templates/project-repo/.github/workflows/`,
   keep `existing_pr_number` on `pipeline.yml` `workflow_dispatch` and the
   `implement` job. Move the five read-only verifier jobs and their dedicated
   inputs onto a dedicated workflow (preferred name `pipeline-verify.yml`).
   Keep mutating operator-loop actions on `pipeline.yml`. Do not drop an
   input. Do not pack verifier scalars into JSON. Do not add operator SHA
   inputs.
3. Add an explicit source regression that every project-repo template
   workflow with `workflow_dispatch` has at most 25 input keys. Update source
   tests that currently hard-code the `pipeline.yml` action-options list so
   they assert the relocated identities instead of requiring the five verify
   actions to remain on `pipeline.yml`.
4. Preserve VOC-121 through VOC-125 fail-closed contracts: exact base/head
   binding, nested isolation, named-ref source bundle, credential-free
   bundles, `publish-source` App-token isolation with workflow-write only on
   that mint, caller workflow-file refusal, force-with-lease, two-attempt
   bound, Cursor Composer implementer, Cursor Grok review. Never bundle
   secrets. Never allow attempt `3`. The dedicated verifier workflow stays
   read-only (no `secrets: inherit`, no `actions: write`, no App-token mint).
5. Update current-state comments/docs (`pipeline.yml` template, dedicated
   verifier workflow, `karsift-ai-infra/README.md`) so they describe operator
   resume as `pipeline.yml` attempt `2` plus `existing_pr_number`, and
   describe the five verifiers as dispatching through the dedicated workflow.
   Do not rewrite historical CHANGELOG or A-003/VOC-075 audit records. Update
   `AGENTS.md` only if an existing sentence would become false.
6. Land the infrastructure change through the normal coordinated source
   carrier. Merge the independently reviewed infra PR first so the
   project-repo template is GitHub-valid under 25 inputs and still declares
   `existing_pr_number` before the caller consumes it. Pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA
   when the fixture consumes the change — not to `1f1705d…`. Update caller
   fixture regressions and any `scripts/foundation/*` pin literals that still
   assert the previous pin, in the same task.
7. After that exact reviewed infra merge is live, update live caller
   `.github/workflows/pipeline.yml` the same way: add `existing_pr_number`,
   relocate the five verifier jobs, keep mutating recover/reconcile/implement
   actions, and add the dedicated verifier workflow. Preserve the live
   caller's existing `live_evidence_mode` reconcile shape unless the
   relocation itself requires touching it. Do not merge PR #1024; this
   package's caller PR is the governed replacement.
8. Add caller tests proving the 25-input maximum, `existing_pr_number`
   forward, relocated verifier jobs, preserved recover/reconcile actions, and
   read-only verifier permissions. Update caller tests that currently
   hard-code the `pipeline.yml` action-options list.
9. After the exact reviewed caller dispatch contract is merged and promoted,
   record in `t00-evidence.md` that #1024 is closed as superseded, that
   VOC-125 #1022 / #1020 are closed only once the live route is valid, and
   that existing `VOC-122-T00` / #1003 / #1012 is resumed through the
   repaired route at attempt `2` with `existing_pr_number=1012`, not
   replaced. This package's implementation PR `Closes` only its own VOC-126
   task issue. Do not implement VOC-122 here, do not merge #1012, and do not
   dispatch VOC-125 as attempt `3`.
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
  a replacement VOC-122 task or PR, or merging #1012 or #1024.
- Dispatching VOC-125-T00 as attempt `3`.
- Silently dropping a recovery/verifier input, or packing verifier scalars
  into JSON.
- Adding operator-typed SHA inputs to caller `workflow_dispatch`.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG / A-003 / VOC-075 records, or expanding
  AGENTS.md into a new verifier-dispatch runbook.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review and the recorded #1024 / #1022 / #1003 / #1012
  handoff.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the live definition defect, the relocated verifier contract,
  the preserved `existing_pr_number` interface, the tests, the docs, the
  caller pin, and the VOC-125 / VOC-122 handoff are one caller-dispatch
  validity outcome.
- Infra should merge first so the project-repo template is GitHub-valid
  under 25 inputs and still declares `existing_pr_number` before the caller
  consumes that template.
- Close #1024 / #1022 / #1020 and resume existing #1003 / #1012 only after
  that exact infra merge and the caller dispatch contract are live. Do not
  treat VOC-126 caller-pin merge as VOC-122 completion.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
