# VOC-124 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` implement workflow and App-token
  mint; caller `tooling/governance/` fixtures and tests.
- Prerequisites: confirm `publish-source` still mints with
  `permission-contents: write`, `permission-issues: write`, and
  `permission-pull-requests: write` only. Confirm the caller `publish` job
  still omits `permission-workflows` and still rejects `.github/workflows/**`.
  Confirm VOC-121/VOC-123 publisher contracts (`publish-source` App token, no
  caller-token fallback, named-ref bundle, `bundle verify`, SHA fetch,
  force-with-lease) remain the baseline this change must preserve.
- VOC-123's isolated named-ref source publisher is live but its token mint
  cannot publish this task's own workflow-file repair. Use the bounded
  supervised bootstrap in `VOC-124-D04` for the first infra PR only. Do not
  treat an untracked local `karsift-ai-infra/` checkout as this repository's
  tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- The already-resolved `implement.yml@main` cannot be recompiled by nested
  edits in its running job. Do not dispatch a predictably failing self-carrier
  or mutate its runner environment to intercept Git.
- Do not print credential values. Do not change App installation permissions
  or rotate `KARSIFT_BOT_*` secrets. Installation `148001476` already has
  repository `workflows: write`.

## File reconciliation and implementation sequence

### T00 — Request workflow-write on the coordinated source publisher token

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` | modify | Add `permission-workflows: write` to the `publish-source` mint only; correct A-004 caller `publish` PR-body text; update current-state comments |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state source-carrier paragraph must describe infrastructure `workflows: write` without giving that permission to the caller publisher |
| `KARSIFT/karsift-ai-infra/CHANGELOG.md` | do not rewrite historical isolated-publication entry | Historical "token has no workflows permission" text is caller-publisher audit evidence |
| `KARSIFT/karsift-ai-infra/tests/test_voc124_*.py` and/or extend `test_voc121_implement_policy.py` / `test_voc121_source_carrier_publisher.py` / `test_live_evidence_reconcile.py` / `test_implementer_bundle.py` | create/extend | Token-permission assertions, workflow-file source bundle coverage, isolated caller `publish_job` inspection, preserved fail-closed cases |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed (`implement.yml` is expected) |
| `tooling/governance/tests/` | modify/extend | Fixture regressions for the source-publisher mint and isolated caller assertions; advance pin literal when consumed |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify if they still assert the previous pin | Same-task pin literals (`7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` at drafting time) |
| `specs/changes/VOC-124-.../t00-evidence.md` | update | Record mechanism, validation commands, infra SHA, pin applicability, bootstrap exhaustion, #1003 / #1012 retry note |

Ordered steps:

1. In a clean isolated `KARSIFT/karsift-ai-infra` worktree based on current
   `main`, use the one-time `VOC-124-D04` bootstrap to add
   `permission-workflows: write` to the `publish-source` mint only. Do not
   add it to the caller `publish` mint. Do not publish VOC-122 nested head
   `f90eb630743c8c523e2e6e8dff017acbb31a7f43`.
2. Correct the caller `publish` PR body so it no longer claims required
   human approval is a pending merge gate under active A-004. Update
   current-state comments/README for the infrastructure token permission.
3. Add deterministic tests for the mint isolation, workflow-file source
   bundle coverage, missing-credential / invalid-bundle / stale-base /
   stale-lease fail-closed cases, and the live-evidence `publish_job` split.
4. Run the infra unit/policy suite. Open one reviewed bootstrap infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first when the caller fixture consumes the
   change.
5. After a different actor merges that exact reviewed infra head and the
   fixed workflow is live on `main`, resume the normal governed carrier. Sync
   and pin the caller fixture to that exact merge SHA when consumed;
   update caller governance and foundation pin tests; record evidence in
   `t00-evidence.md`, including bootstrap exhaustion and that #1003 / #1012
   is a distinct VOC-122 retry against that SHA.
6. Re-dispatch or reconcile existing `VOC-122-T00` through the repaired
   `implement.yml@main` path. Do not create a replacement VOC-122 task or PR.
   After that source PR merges, #1012 is updated to the exact infrastructure
   merge SHA with truthful evidence by the existing VOC-122 carrier, not by
   merging #1012 from this package.

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
`python3 -m unittest tests.test_voc124_source_publisher_workflows` or the
three foundation pin tests), record the exact commands in `t00-evidence.md`
and run them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- `publish-source` mint requests `permission-workflows: write` and the caller
  `publish` mint does not;
- an authorized `.github/workflows/**` source bundle is covered by that
  permission and is not rejected by the source publisher script;
- missing App credentials, invalid bundles, stale bases, and stale leases
  still fail closed;
- VOC-121/VOC-123 isolation, named-ref bundle, App-token split, lease, retry
  limits, and non-closing source PR remain;
- carrier current-state text matches active A-004 and historical CHANGELOG
  is not rewritten;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable;
- VOC-122 / #1003 / #1012 behavior was not implemented or merged in this
  package, and nested head `f90eb630743c8c523e2e6e8dff017acbb31a7f43` was not
  hand-published;
- the implementer did not approve or merge its own work on either carrier;
- the `VOC-124-D04` bootstrap was limited to the first infra PR, used no
  runner-environment interception or direct `main` push, did not publish the
  VOC-122 bundle, and was exhausted by the exact infra merge before normal
  caller work resumed.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future implement jobs that commit nested
  `karsift-ai-infra` workflow-file changes can be pushed by `publish-source`
  instead of failing GitHub's workflows-permission check after a successful
  bundle verify.
- **Rollback trigger:** Source publisher cannot push authorized workflow-file
  commits; caller publisher gains workflows permission or stops refusing
  caller workflow files; missing credentials, invalid bundles, or stale
  leases start succeeding; or App tokens appear on the model runner.
- **Rollback mechanism:** Revert the infra and caller fixture/test/doc
  changes to the prior reviewed VOC-123 source-carrier behavior.
- **Last-known-good reference:** Current source-carrier workflows on
  `main`/`develop` after VOC-123 (infra merge
  `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`) and before VOC-124
  implementation lands. That last-known-good still has the missing
  `permission-workflows` defect; rollback restores a known reviewed state,
  not a working nested workflow-file publisher. Re-introducing the omitted
  permission re-breaks #1013-class delivery.
