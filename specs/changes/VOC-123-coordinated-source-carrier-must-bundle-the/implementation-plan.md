# VOC-123 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` implement/plan workflows and
  any helper that creates or verifies Git bundles for publication; caller
  `tooling/governance/` fixtures and tests.
- Prerequisites: confirm `Commit implementer's work` still creates
  `/tmp/implementer-source.bundle` from
  `"${{ steps.infra-checkout.outputs.base_sha }}..$SOURCE_HEAD_SHA"`. Confirm
  caller recovery still uses `integration_sha..HEAD` and planner recovery
  still uses `base_sha..HEAD`. Confirm VOC-121 publisher contracts
  (`publish-source` App token, `bundle verify`, SHA fetch, force-with-lease)
  remain the baseline this change must preserve.
- VOC-121's isolated source publisher is live but its raw-SHA bundle command
  cannot publish this task's own repair. Use the bounded supervised bootstrap
  in `VOC-123-D08` for the first infra PR only. Do not treat an untracked local
  `karsift-ai-infra/` checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- The already-resolved `implement.yml@main` cannot be recompiled by nested
  edits in its running job. Do not dispatch a predictably failing self-carrier
  or mutate its runner environment to intercept Git.

## File reconciliation and implementation sequence

### T00 — Bundle the coordinated source-carrier committed head through a named ref

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` | modify | Bind exact `SOURCE_HEAD_SHA` to an isolated named ref; `bundle create` from `base_sha..that-ref`; verify advertised heads; delete the temp ref; keep `source_head_sha` output |
| `KARSIFT/karsift-ai-infra/.github/workflows/plan.yml` | modify only if tests reproduce the defect | Drafting-time path is `base_sha..HEAD`; change only if empty-bundle or unsafe advertised-head is proven |
| `KARSIFT/karsift-ai-infra/config/implementer_source_carrier.py` | modify only if a helper is extracted | Optional; publisher metadata/gitlink helpers may stay unchanged |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state source-carrier paragraph must describe named-ref bundle tips |
| `KARSIFT/karsift-ai-infra/tests/test_voc123_*.py` and/or extend `test_voc121_implement_policy.py` / `test_voc121_source_carrier_publisher.py` / `test_implementer_bundle.py` | create/extend | Real-repository raw-SHA failure, named-ref success, fail-closed advertised-head cases, caller/planner `..HEAD` proof |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed (`implement.yml` is expected) |
| `tooling/governance/tests/test_voc121_implement_policy.py` and/or new VOC-123 fixture test | modify/extend | Fixture regressions for named-ref bundle create; advance pin literal when consumed |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify if they still assert the previous pin | Same-task pin literals (`99476c2a1018e42d4bd442657b5257885ac9f1c9` at drafting time) |
| `specs/changes/VOC-123-.../t00-evidence.md` | update | Record mechanism, advertised-head proof, caller/planner result, commands, infra SHA, pin applicability, #1003 re-dispatch note |

Ordered steps:

1. In a clean isolated `KARSIFT/karsift-ai-infra` worktree based on current
   `main`, use the one-time `VOC-123-D08` bootstrap to change nested
   source-bundle creation so
   the exact committed `SOURCE_HEAD_SHA` is bound to an isolated temporary
   named ref, the bundle is created from `base_sha..that-ref`, advertised
   heads are verified, and the temp ref is deleted. Do not reuse the publish
   branch or `source-carrier`.
2. Prove caller `integration_sha..HEAD` and planner `base_sha..HEAD` with
   real-repository tests. Change those paths only if they reproduce the
   defect.
3. Add deterministic tests for raw-SHA empty-bundle, named-ref success,
   fail-closed advertised-head/base/SHA/unrelated-ref cases, and
   caller/planner proof. Update any VOC-121 test that currently bundles
   `..HEAD` for the nested source path so the production command shape is
   what is executed.
4. Update current-state comments/docs so they describe named-ref source
   bundle tips.
5. Run the infra unit/policy suite. Open one reviewed bootstrap infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first when the caller fixture consumes the
   change.
6. After a different actor merges that exact reviewed infra head and the
   fixed workflow is live on `main`, resume the normal governed carrier. Sync
   and pin the caller fixture to that exact merge SHA when consumed;
   update caller governance and foundation pin tests; record evidence in
   `t00-evidence.md`, including that #1003 is a distinct VOC-122
   re-dispatch against that SHA.

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
`python3 -m unittest tests.test_voc123_source_carrier_bundle` or the three
foundation pin tests), record the exact commands in `t00-evidence.md` and run
them in addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- raw-SHA positive tip still reproduces empty-bundle in tests;
- the fixed source-carrier path advertises exactly the expected committed
  head and then removes the temporary ref;
- wrong/missing/multiple advertised heads, wrong base, malformed SHA, and
  unrelated refs fail closed;
- caller and planner `..HEAD` paths are proven safe or repaired with
  equivalent coverage;
- VOC-121 isolation, App-token split, lease, retry limits, and non-closing
  source PR remain;
- the caller fixture pin equals the exact reviewed infra merge when the
  fixture consumes the change, or evidence records why the pin was not
  applicable;
- VOC-122 / #1003 behavior was not implemented in this package;
- the implementer did not approve or merge its own work on either carrier.
- the `VOC-123-D08` bootstrap was limited to the first infra PR, used no
  runner-environment interception or direct `main` push, and was exhausted by
  the exact infra merge before normal caller work resumed.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future implement jobs that commit nested
  `karsift-ai-infra` changes create a named-ref source bundle and can upload
  it for `publish-source`, instead of failing with empty-bundle after the
  nested commit.
- **Rollback trigger:** Source bundle is empty; advertises the wrong head;
  publishes unrelated refs; or caller/planner recovery bundles that were
  previously working become empty or unsafe.
- **Rollback mechanism:** Revert the infra and caller fixture/test/doc
  changes to the prior reviewed VOC-121 source-carrier behavior.
- **Last-known-good reference:** Current source-carrier workflows on
  `main`/`develop` after VOC-121 (infra merge
  `99476c2a1018e42d4bd442657b5257885ac9f1c9`) and before VOC-123
  implementation lands. That last-known-good still has the empty-bundle
  defect; rollback restores a known reviewed state, not a working nested
  publisher. Re-introducing the raw-SHA tip re-breaks #1003-class delivery.
