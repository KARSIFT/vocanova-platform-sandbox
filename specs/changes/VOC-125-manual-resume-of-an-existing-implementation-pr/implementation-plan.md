# VOC-125 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` implement and remediate
  workflows; caller `.github/workflows/pipeline.yml`; exact-head remediation
  bindings; attempt caps; caller `tooling/governance/` fixtures and tests.
- Prerequisites: confirm `implement.yml` still requires
  `expected_head_sha` on `attempt != 1` via `verify-expected-head.py`. Confirm
  caller `pipeline.yml` still forwards only `change_id`, `package_path`,
  `task_id`, `issue_number`, `attempt`, and `integration_branch`. Confirm
  `remediate.yml` retry still passes event-derived SHAs. Confirm VOC-121/
  VOC-123/VOC-124 publisher contracts remain the baseline this change must
  preserve.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source` (infra merge
  `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`). T00's first run is attempt `1`.
  Do not treat an untracked local `karsift-ai-infra/` checkout as this
  repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions
  or rotate `KARSIFT_BOT_*` secrets.

## File reconciliation and implementation sequence

### T00 — Bind existing-carrier resume identity

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` | modify | Add `existing_pr_number`; bind identity before `Create implementation branch`; fail closed per `VOC-125-D02` / `VOC-125-D03`; update current-state comments |
| `KARSIFT/karsift-ai-infra/.github/workflows/remediate.yml` | modify | Forward `pr_number` as `existing_pr_number` on retry; keep SHA forwards |
| `KARSIFT/karsift-ai-infra/config/verify-expected-head.py` and/or new `config/bind-existing-carrier.py` | extend/create | Deterministic derivation and mismatch classes; layout is `VOC-125-DEP-07` |
| `KARSIFT/karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml` | modify | Expose and forward `existing_pr_number`; no operator SHA inputs |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state implement-resume paragraph |
| `KARSIFT/karsift-ai-infra/tests/test_voc125_*.py` and/or extend `test_remediate_policy.py` / `test_voc121_implement_policy.py` | create/extend | Valid resume, #1020 empty-binding, every mismatch class, remediate forward, attempt cap |
| `.github/workflows/pipeline.yml` | modify | Same dispatch contract as the template |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed (`implement.yml` is expected) |
| `tooling/governance/tests/` | modify/extend | Fixture regressions; advance pin literal when consumed |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify if they still assert the previous pin | Same-task pin literals (`f406cc95a3f853e8aef5bf8bcf22d37a29d64547` at drafting time) |
| `AGENTS.md` | modify only if an existing sentence would become false | Do not add a new implement-dispatch runbook |
| `specs/changes/VOC-125-.../t00-evidence.md` | update | Record mechanism, validation commands, infra SHA, pin applicability, #1003 / #1012 resume handoff |

Ordered steps:

1. In a clean isolated `KARSIFT/karsift-ai-infra` worktree based on current
   `main`, add `existing_pr_number` and the bind step to `implement.yml`.
   Fail closed for the `VOC-125-D03` classes before `Create implementation
   branch`. Do not add operator SHA inputs to the pipeline template.
2. Forward `existing_pr_number` from `remediate.yml` retry while keeping
   event-derived SHA forwards.
3. Add deterministic tests for valid resume, the #1020 empty-binding class,
   every mismatch class, attempt-cap preservation, and remediate forwarding.
4. Update current-state comments/README. Run the infra unit/policy suite.
   Open one reviewed infra PR that `Relates to
   KARSIFT/vocanova-platform-sandbox#<task>` and does not use a closing
   keyword. Merge it first so `implement.yml@main` declares the new input.
5. After a different actor merges that exact reviewed infra head, update the
   live caller `pipeline.yml` to expose and forward `existing_pr_number`.
   Sync and pin the caller fixture to that exact merge SHA when consumed;
   update caller governance and foundation pin tests; record evidence in
   `t00-evidence.md`, including that #1003 / #1012 is a distinct VOC-122
   resume against that SHA at attempt `2` with `existing_pr_number=1012`.
6. Record the repaired dispatch command for existing `VOC-122-T00`. Do not
   create a replacement VOC-122 task or PR. Do not merge #1012 from this
   package.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
git diff --check
```

If implementation adds narrower targeted commands (for example
`python3 -m unittest tests.test_voc125_existing_carrier` or the three
foundation pin tests), record the exact commands in `t00-evidence.md` and run
them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- caller/template pipeline expose and forward `existing_pr_number` and do not
  expose operator SHA inputs;
- a valid attempt-2 resume derives and binds exact head/base before model
  resolution and reuses the existing carrier;
- the #1020 empty-binding class and every `VOC-125-D03` mismatch class fail
  closed before model or mutation;
- attempt `1` with an existing open carrier and attempt `3` fail closed;
- `remediate.yml` still forwards event-derived SHAs and now forwards
  `existing_pr_number`;
- VOC-121/VOC-123/VOC-124 isolation, named-ref bundle, App-token split,
  lease, retry limits, Cursor Composer/Grok roles, and non-closing source PR
  remain;
- carrier current-state text describes attempt `2` plus `existing_pr_number`;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable;
- VOC-122 / #1003 / #1012 behavior was not implemented or merged in this
  package;
- the implementer did not approve or merge its own work on either carrier;
- no bootstrap exception was used.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** After automatic remediation stops, a governed
  operator can resume the same implementation PR by dispatching
  `action=implement` with `attempt=2` and `existing_pr_number=<PR>`. The
  already-authorized VOC-122-T00 carrier (#1003 / #1012) can use that route.
- **Rollback trigger:** Operator resume still cannot bind exact head/base;
  free-form SHA paste becomes the operator interface; attempt `1` rewrites an
  existing carrier; attempt `3` is accepted; mismatch classes start
  succeeding; automatic remediate retry loses event-derived SHAs; or App
  tokens appear on the model runner.
- **Rollback mechanism:** Revert the infra and caller workflow/fixture/test/
  doc changes to the prior reviewed VOC-124 implement/dispatch behavior.
- **Last-known-good reference:** Current implement/dispatch workflows on
  `main`/`develop` after VOC-124 (infra merge
  `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`) and before VOC-125
  implementation lands. That last-known-good still has the missing
  operator-resume identity defect; rollback restores a known reviewed state,
  not a working existing-carrier resume. Re-introducing empty SHA forwards
  re-breaks the #1020 class.
