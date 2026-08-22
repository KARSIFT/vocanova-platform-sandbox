# VOC-107 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `karsift-ai-infra` `implement.yml` isolated publish job;
  exact published SHA; integration ancestry; `.github/workflows/**` deny;
  SHA-valued force-with-lease; two-attempt implementer/remediation cap;
  App token only on the clean publish runner.
- Prerequisites: issue #891 records the thin-bundle reject; drafting-time
  diagnosis of `base_sha..HEAD` vs integration-only publisher fetch; this draft
  is adopted under A-004.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package authorizes the
required outcome. Do not treat an untracked local `karsift-ai-infra/` checkout
as this repository's tracked tree. Calling-repo foundation-test or doc changes
land here under the same package.

## File reconciliation and implementation sequence

### T00 — Integration-anchored bundle lineage, publisher guards, docs, fixtures

| File / area                                              | Action        | Notes                                                                 |
| -------------------------------------------------------- | ------------- | --------------------------------------------------------------------- |
| `karsift-ai-infra/.github/workflows/implement.yml`       | modify        | Record integration-anchor; bundle `anchor..HEAD`; publish uses anchor |
| karsift-ai-infra helper (optional extract)               | create/modify | Only if needed for deterministic fixture reuse                        |
| karsift-ai-infra README                                  | modify        | Document integration-anchored implementer bundles                     |
| karsift-ai-infra / calling-repo tests                    | create/modify | End-to-end Git fixture + policy assertions (`voc107-*` / self-ci)     |
| calling-repo docs                                        | modify if needed | Only if current wording would become false                         |
| `specs/changes/VOC-107-.../t00-evidence.md`              | create/update | Commands + results (metadata only)                                    |

Ordered steps:

1. Confirm drafting-time diagnosis against current `implement.yml` HEAD: bundle
   create uses `steps.branch.outputs.base_sha..HEAD`; attempt 2+ sets `base_sha`
   after agent-branch checkout/rebase; publish fetches only
   `integration_branch` before `bundle verify` (no secrets).
2. When creating/preparing the implementation branch, record an explicit
   integration-anchor SHA equal to the integration tip the branch is based on
   after fresh checkout, successful rebase, or fresh-after-rebase-conflict
   restart. Keep outputting the pre-model tip for soft-reset commit semantics.
3. Change bundle creation to `integration_anchor..HEAD` (complete task-branch-only
   lineage). Do not change soft-reset to the integration tip.
4. Update the publish job environment so ancestry and workflow-path deny use the
   integration-anchor as `PUBLISH_BASE_SHA` (or equivalent), while exact head SHA
   check and SHA-valued force-with-lease remain as today.
5. Preserve publisher isolation: fresh bare repo; no model code execution before
   publish; App token only on the clean runner; reject `.github/workflows/**` in
   the full published lineage.
6. Do not alter the two-attempt cap or add a model rerun for publisher
   prerequisite misses.
7. Land deterministic Git fixtures per `VOC-107-D05` / TEST-00–05 and TEST-06/07.
8. Align infra README (and calling-repo docs only if needed).
9. Record current `@main` reusable-workflow consumption (no pin bump expected);
   reconcile explicitly if repository state differs.
10. Run applicable tests and governance validation; write `t00-evidence.md`
    without logs, artifacts, or secrets.
11. If an identical planner (`plan.yml`) failure class is discovered, open a
    separate unlabeled issue; do not expand this task.

## Validation and independent verification

Deterministic (T00):

```bash
# Exact commands depend on where tests land; record actual invocations in evidence.
# Example shapes:
cd karsift-ai-infra
python3 -m unittest discover -s tests -p 'test_*bundle*.py' -v
# and/or
python3 -m unittest discover -s tests -p 'test_implement*.py' -v

cd ..
node --test scripts/foundation/voc107-*.test.mjs

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Independent verifier binds the exact task PR SHA and confirms fixture coverage
for positive attempt-2 / rebase-derived import, negative incomplete lineage, and
preserved publisher guards. Missing live Actions proof is not required for this
package’s acceptance (deterministic Git fixture is the specified proof).

## Deployment and rollback

- **Authorization:** Package adoption + task implementation authorization only;
  this package does not itself authorize production application deployment.
- **Rollout:** Infra implement.yml change merges and is consumed by calling
  `pipeline.yml` at `@main`; calling-repo foundation tests/docs land on
  `develop` and promote normally.
- **Rollback trigger:** Attempt-2 remediation publications again fail
  `bundle verify` in an integration-only bare repo; or soft-reset begins
  squashing prior task commits; or publisher guards weaken.
- **Rollback mechanism:** Revert the implement.yml / fixture commit(s) on the
  infra default branch and any calling-repo pin/tests through normal PR paths.
- **Last-known-good:** commit before T00 merge on each affected repo.
