# VOC-124 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1013 is one
source-publisher permission outcome. Coordinated caller and infrastructure
pull requests remain one task; repository count, workflow-versus-tests-versus-docs,
and fixture/pin work are not split reasons. Retrying the existing VOC-122
carrier is evidence of this outcome, not a second VOC-124 task.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for the
`publish-source` App-token mint. Because the broken running workflow cannot
publish its own workflow-file repair, the initial infra PR uses the bounded
supervised bootstrap in `VOC-124-D04`; this package is the authorizing change
package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree. Caller
fixture/pin, tests, and evidence land in this repository under the same task.
Infra PRs must say `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and
MUST NOT use a closing keyword.

## VOC-124-T00 — Request workflow-write on the coordinated source publisher token

- Requirement source: issue #1013; `VOC-124-D00` through `VOC-124-D07`
- Acceptance criteria: `VOC-124-AC-00` through `VOC-124-AC-06`
- Tests: `VOC-124-TEST-00` through `VOC-124-TEST-07`
- Evidence: `VOC-124-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1013 live failure in `t00-evidence.md` (task #1003,
   pipeline run `32958526215` / job `98147443377`, nested head
   `f90eb630743c8c523e2e6e8dff017acbb31a7f43`, infrastructure base
   `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`, GitHub App workflows-permission
   rejection after a successful bundle verify and token mint, draft caller PR
   #1012, defect in `implement.yml` `publish-source` mint omitting
   `permission-workflows: write`).
2. In `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml`
   `publish-source`, add `permission-workflows: write` to the
   `Mint least-privilege App token for infrastructure repository` step. Keep
   `repositories: karsift-ai-infra` and the existing contents, issues, and
   pull-requests writes. Do not add `permission-workflows` to the caller
   `publish` mint. Do not add a caller-token fallback. Do not print token
   values.
3. Preserve VOC-121/VOC-123 fail-closed contracts: exact base/head binding,
   nested-repository isolation, no gitlink, named-ref source bundle,
   credential-free source bundle, clean `publish-source` App-token
   separation, `bundle verify` plus authorized-SHA fetch, force-with-lease,
   two-attempt bound, non-closing source PR, caller `Closes #N`, caller
   rejection of `.github/workflows/**`. Never bundle secrets.
4. Add deterministic tests that prove:
   - `publish-source` mint includes `permission-workflows: write`;
   - caller `publish` mint does not, and still rejects workflow-file
     changes;
   - an authorized source bundle that changes `.github/workflows/**` is not
     rejected by the source publisher's own script checks and is covered by
     that token permission;
   - missing App credentials, invalid bundles, stale bases, and stale
     leases still fail closed;
   - `test_live_evidence_reconcile.py` (and any equivalent fixture test)
     inspects the caller `publish` job in isolation so adding the
     source-publisher permission does not make a whole-file `NotIn` assertion
     fail for the wrong reason.
   Do not mint real App tokens or use secrets or production data.
5. Update current-state comments/docs (`implement.yml` current-state /
   publish-source comments, caller `publish` PR body A-004 wording,
   `karsift-ai-infra/README.md` source-carrier paragraph) so they describe the
   infrastructure token's `workflows: write` request and no longer present
   "required human approval" as an engineering-workflow merge gate. Do not
   rewrite historical CHANGELOG or A-003/VOC-075 audit records.
6. Bootstrap the initial infra change through one clean isolated branch from
   current infra `main` under `VOC-124-D04`, because the already-compiled
   `implement.yml@main` cannot consume its own nested edit and would reject
   the workflow-file push. Open one reviewed non-closing infra PR, run full
   source validation, bind independent review to its exact final SHA, and
   require a different actor to merge. Do not mutate PATH/Git, push to
   `main`, rotate secrets, change the App installation, or publish VOC-122
   nested head `f90eb630743c8c523e2e6e8dff017acbb31a7f43`. After that exact
   merge is live, use the normal coordinated carrier for remaining caller
   work. Pin `tooling/governance/fixtures/karsift-ai-infra/` to that exact
   merge SHA when the fixture consumes the change. Update caller fixture
   regressions and any `scripts/foundation/*` pin literals that still assert
   `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`, in the same task.
7. After the exact reviewed infra merge is live on `implement.yml@main`,
   record in `t00-evidence.md` that existing `VOC-122-T00` / #1003 is
   re-dispatched or reconciled against that revision, that #1012 remains this
   package's out-of-scope draft caller PR until VOC-122's authoritative
   source merge exists, and that the bootstrap exception is exhausted. Do not
   implement VOC-122 promotion-recovery replan here and do not reconstruct
   nested head `f90eb630743c8c523e2e6e8dff017acbb31a7f43` by hand.
8. Run applicable validation and record results in `t00-evidence.md`:
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - targeted foundation pin tests if those files change;
   - `git diff --check`;
   - exact reviewed infra SHA and pin applicability;
   - any narrower targeted commands added by the implementation.
9. Preserve independent exact-SHA review for each carrier, risk
   classification, protected checks, and App-token isolation. Do not weaken
   review, risk classification, required checks, or automatic-merge gates.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  installation-permission, or monitor-inventory changes.
- Implementing VOC-122 promotion-recovery replan inside this package, creating
  a replacement VOC-122 task or PR, merging #1012, or hand-publishing nested
  head `f90eb630743c8c523e2e6e8dff017acbb31a7f43`.
- Adding `permission-workflows` to any token other than the `publish-source`
  infrastructure mint.
- OpenAI credentials or execution routes.
- Rewriting historical CHANGELOG / A-003 / VOC-075 records, or sweeping
  planner/plan-review founder-comment comments that are not in this carrier.
- Weakening exact-SHA review, risk floors, protected checks, retry caps, or
  App-token isolation.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review, recorded bootstrap exhaustion, and the existing
  VOC-122 carrier retry through the repaired path.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the live defect, the token-permission contract, the tests, the
  docs, the bootstrap, the caller pin, and the existing VOC-122 retry are one
  publisher-permission outcome.
- Infra should merge first when the caller fixture/pin consumes that change;
  otherwise the two reviewed PRs may complete under the same task without a
  pin bump.
- The initial infra repair cannot consume itself: GitHub resolves the reusable
  workflow from old infra `main` before the job starts, and the repair itself
  changes a workflow file. Use only the bounded `VOC-124-D04` supervised
  bootstrap for that first infra PR. Once merged, normal `implement.yml@main`
  publication is mandatory and the bootstrap authority is exhausted.
- Re-dispatch or reconcile existing #1003 only after that exact infra merge is
  live. Do not treat VOC-124 caller-pin merge as VOC-122 completion.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
