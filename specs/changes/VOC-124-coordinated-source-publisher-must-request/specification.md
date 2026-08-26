# VOC-124 — Coordinated source publisher must request workflow-write permission: Specification

## Objective and requirement source

Make the governed coordinated source-carrier publisher able to publish an
authorized nested `KARSIFT/karsift-ai-infra` commit that changes
`.github/workflows/**` by requesting `permission-workflows: write` on that
publisher token only, without weakening VOC-121 isolation, VOC-123 named-ref
bundle tips, caller workflow-file refusal, lease, App-token, or fail-closed
contracts.

**Requirement source:** [GitHub issue #1013](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1013).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1013)

| Item | Value |
|------|-------|
| Adopted task blocked | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Pipeline run / job | `32958526215` / `98147443377` |
| Nested bundle head | `f90eb630743c8c523e2e6e8dff017acbb31a7f43` |
| Infrastructure base | `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd` |
| Bundle / token mint | succeeded |
| Error | `refusing to allow a GitHub App to create or update workflow ... without workflows permission` |
| Changed workflow file | `.github/workflows/recover-actions-checks.yml` |
| Caller PR | [#1012](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1012) — draft; must remain unmerged until the authoritative source PR merges |
| Defect locus | `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` `publish-source` App-token mint omits `permission-workflows: write` |
| Installation | `karsift-ai-infra-bot` `148001476` already has repository `workflows: write` |

## Scope and non-goals

### In scope

1. On the clean `publish-source` job, add `permission-workflows: write` to the
   `actions/create-github-app-token` mint that already requests
   `permission-contents: write`, `permission-issues: write`, and
   `permission-pull-requests: write` for `repositories: karsift-ai-infra`.
2. Keep that permission off the caller `publish` job token mint. Keep the
   caller publisher's pre-push rejection of `.github/workflows/**`.
3. Keep publication fail-closed:
   - exact 40-character base and head binding;
   - nested-repository isolation (no caller gitlink);
   - credential-free source bundle consumed by a clean `publish-source` job;
   - infrastructure-scoped App token with no caller-token fallback;
   - `bundle verify` then fetch of the authorized head only;
   - force-with-lease against `EXPECTED_SOURCE_HEAD_SHA`;
   - two-attempt implementer bound;
   - independent exact-SHA review and protected checks;
   - missing App credentials, invalid bundles, stale bases, and stale leases
     fail closed;
   - never print credential values.
4. Add deterministic tests proving an authorized source bundle that changes
   `.github/workflows/**` is covered by the required token permission, while
   the named fail-closed cases still fail closed and the caller publisher
   still refuses workflow-file changes without `permission-workflows`.
5. Correct current-state comments and PR body text in the same
   `implement.yml` carrier path that inaccurately say human approval is still
   required under active A-004. Do not rewrite historical CHANGELOG or other
   audit records.
6. Land the infrastructure change through one reviewed infra PR using the
   bounded bootstrap in `VOC-124-D04`, then pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA when
   the fixture consumes the change, and update caller fixture/pin tests in the
   same task.
7. After that exact infra merge is live on `implement.yml@main`, retry the
   existing `VOC-122-T00` carrier. Do not create a replacement VOC-122 task
   or PR, and do not hand-publish nested head
   `f90eb630743c8c523e2e6e8dff017acbb31a7f43`.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Implementing VOC-122 promotion-recovery replan (`VOC-122-T00` / #1003)
  inside this package. That remains a distinct already-authorized outcome
  whose existing carrier is retried after this repair is live.
- Hand-pushing, force-pushing, or reconstructing nested head
  `f90eb630743c8c523e2e6e8dff017acbb31a7f43` outside the governed implement
  path.
- Merging or treating caller PR #1012 as this package's implementation PR.
- Adding `permission-workflows` to the caller `publish` token, planner
  `publish-plan` token, live-evidence reconcile token, or any other mint.
- Changing GitHub App installation permissions, rotating
  `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`, or changing repository
  settings. Installation `148001476` already has `workflows: write`.
- Reopening VOC-121 source-publication semantics or VOC-123 named-ref bundle
  tips.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, or fail-closed missing-bundle
  behavior.
- Rewriting historical CHANGELOG, A-003, or VOC-075 audit records.
- OpenAI credentials or execution routes.
- Splitting workflow logic, tests, docs, infrastructure, caller pin, or
  evidence into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests,
  exact-SHA review, recorded bootstrap exhaustion, and the existing VOC-122
  carrier retry through the repaired `implement.yml@main` path.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: reusable CI/CD implement workflow, coordinated
  second-repository publication of infrastructure workflow files, GitHub App
  token mint, and caller `tooling/governance/` fixtures and tests.
- Protected technical effect: whether an authorized nested infrastructure
  commit that changes `.github/workflows/**` can be pushed by the
  infrastructure-scoped App token. No application runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-124-D00`: This is one outcome-sized publisher-permission change. Use one
end-to-end implementation task covering infrastructure source, tests,
current-state docs/comments, caller fixture/pin, bootstrap, and evidence.
Coordinated pull requests in `KARSIFT/karsift-ai-infra` and this caller remain
one task. Repository count, file count, and workflow-versus-tests-versus-docs
are not split reasons. Retrying the existing VOC-122 carrier is evidence of
this outcome, not a second VOC-124 task and not a replacement VOC-122 roster
entry.

`VOC-124-D01`: The clean `publish-source` job must request
`permission-workflows: write` on its `actions/create-github-app-token` mint,
in addition to the existing contents, issues, and pull-requests writes, still
scoped to `owner: ${{ github.repository_owner }}` and
`repositories: karsift-ai-infra`. The permission is always requested on that
token (the mint happens before the push, and authorized infrastructure
workflow changes must be publishable). It is not added to the caller
`publish` mint or to other workflows' App-token mints.

`VOC-124-D02`: Preserve VOC-121/VOC-123 fail-closed publication contracts:

- exact base/head SHA binding and integration-ancestor checks;
- nested checkout isolation and refusal to stage `karsift-ai-infra` as a
  gitlink;
- VOC-123 named-ref source-bundle creation (`refs/karsift/source-bundle-head`
  or the live equivalent);
- credential-free source bundle consumed by a clean `publish-source` job;
- infrastructure-scoped App token with no caller-token fallback;
- `bundle verify` then fetch of the authorized head only;
- force-with-lease against `EXPECTED_SOURCE_HEAD_SHA`;
- two-attempt implementer bound;
- source PR `Relates to OWNER/CALLER#N` with no closing keyword;
- caller PR keeps local `Closes #N`;
- caller `publish` continues to reject `.github/workflows/**` before push and
  continues to omit `permission-workflows`;
- no secrets in bundles, logs, or fixtures;
- no credential values printed.

`VOC-124-D03`: Deterministic tests must prove:

1. the `publish-source` mint `with:` block contains
   `permission-workflows: write`;
2. the caller `publish` mint `with:` block does not;
3. an authorized source bundle whose commit changes `.github/workflows/**`
   is not rejected by the source publisher's own publication script (that
   script has no caller-style workflow-file grep) and is covered by the
   required token permission;
4. missing App credentials still fail closed with no caller-token fallback;
5. missing/invalid bundles, malformed SHA/branch metadata, stale
   integration bases, and stale/racing leases still fail closed;
6. existing VOC-121 isolation, lease, retry, and non-closing source-PR
   contracts remain;
7. the live-evidence permission assertion inspects the caller `publish` job
   in isolation and does not treat `publish-source` as that job.

Positive cases must prove the corrected token request. Tests must not mint
real App tokens, use secrets, or use production data. Existing VOC-121
publisher tests that push a non-workflow file do not by themselves satisfy
(1) or (3).

`VOC-124-D04`: T00 cannot publish its own initial infrastructure repair with
the broken token request. The caller resolves
`KARSIFT/karsift-ai-infra/.github/workflows/implement.yml@main` before the
model job starts; the nested repair changes a workflow file; the live mint
omits `permission-workflows`. As in `VOC-121-D10` and `VOC-123-D08`,
supervised bootstrap recovery of this task's own source PR is therefore
in-scope under T00, not a second package or task.

The bootstrap is limited to one clean isolated infrastructure branch based on
the current protected `main`, the exact token-permission repair, its
tests/docs/A-004 current-state text, and no caller pin. It must open one
non-closing infra PR linked to the T00 authority issue, pass source self-CI,
receive independent review bound to its exact final SHA, and be merged by
someone other than the implementer. No model-runner credential, PATH/Git
interception, ruleset bypass, direct push to `main`, secret/installation
change, or publication of VOC-122 nested head
`f90eb630743c8c523e2e6e8dff017acbb31a7f43` is allowed. After that exact infra
merge is live, all remaining caller fixture/pin/evidence work and the
VOC-122-T00 retry must use the normal governed `implement.yml@main` path.
This exception expires with the bootstrap infra merge and cannot be reused.

`VOC-124-D05`: Current-state comments in `implement.yml` (header
current-state block and the `publish-source` mint/PR steps) and the
`karsift-ai-infra/README.md` source-carrier paragraph must state that the
infrastructure publisher token requests `workflows: write` so authorized
nested workflow-file commits can be pushed, while the caller publisher still
refuses caller `.github/workflows/**` and still omits that permission. The
caller `publish` PR body must stop saying "required human approval are still
pending" as if a founder `approved` comment were an engineering-workflow
merge gate under active A-004. Replacement text must still say independent
exact-revision review is pending and the PR is not authorized to merge on
its own. Historical CHANGELOG entries that describe the original
no-workflows-permission caller-publisher design stay unchanged.

`VOC-124-D06`: After the exact reviewed infrastructure merge SHA is known,
pin `tooling/governance/fixtures/karsift-ai-infra/` when the mirrored fixture
consumes the changed workflow, tests, or comments. Advance matching caller
pin assertions, including
`tooling/governance/tests/test_voc121_implement_policy.py` and the
`scripts/foundation/voc097-fixture-matrix.test.mjs`,
`voc104-ready-for-review-reuse.test.mjs`, and
`voc108-authoritative-lifecycle.test.mjs` pin literals when those tests still
assert the previous infra merge (`7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`
at drafting time).

`VOC-124-D07`: This package is a hard dependency for #1003 / #1012 and a
distinct publisher-permission outcome. Do not implement VOC-122
promotion-recovery replan behavior here. After the exact reviewed infra merge
is live on `implement.yml@main`, re-dispatch or reconcile the existing
`VOC-122-T00` carrier against that revision so its authoritative
infrastructure PR is published from a newly verified bundle, independently
reviewed, and merged. Then update caller PR #1012 to that exact
infrastructure merge SHA with truthful evidence. Do not reconstruct or
hand-push runner-produced head
`f90eb630743c8c523e2e6e8dff017acbb31a7f43`. Do not merge #1012 from this
package. Record bootstrap exhaustion in `t00-evidence.md` so the exception
cannot be reused.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Source publication
continues to use a credential-free bundle on the model-controlled runner and
an infrastructure-scoped GitHub App token only on the clean `publish-source`
job. The model-controlled implementer runner still never receives the GitHub
App token. Caller publication remains a separate credential-free bundle whose
token still omits `workflows: write`.

The added `permission-workflows: write` request is least-privilege for the
already-installed `karsift-ai-infra-bot` repository permission on
`KARSIFT/karsift-ai-infra`. It does not grant the caller publisher the ability
to push workflow files, and it does not change installation permissions.

Abuse/process risks:

1. Authorized nested workflow-file commits remaining unpublishable — mitigated
   by `VOC-124-D01` and token-permission tests.
2. Broadening `permission-workflows` onto the caller `publish` token, which
   would undermine the caller refusal that prevents unreviewed
   same-repository workflow execution — forbidden by `VOC-124-D01` /
   `VOC-124-D02`.
3. Using the bootstrap exception to hand-publish VOC-122's existing bundle —
   forbidden by `VOC-124-D04` / `VOC-124-D07`.
4. Printing App tokens, private keys, or secret values in logs, tests, or
   evidence — forbidden.

## Contradictions and open questions

1. **Test file layout (`VOC-124-DEP-08`):** the required assertions are
   settled; whether they live in new `tests/test_voc124_*.py` files or extend
   VOC-121/VOC-123 publisher/policy tests is an implementation choice.
2. **A-004 PR-body wording:** the required behavior is to stop claiming a
   standing human-approval merge gate. Exact replacement sentences are an
   implementation choice provided they still require independent
   exact-revision review and still say the PR cannot merge on its own.
3. **plan.yml / plan-review.yml founder-comment comments:** drafting-time
   reading found header comments on the planner/plan-review carrier that
   still mention a founder `approved` comment. Those are a different carrier
   than `publish-source`. T00 changes them only if they appear in the
   `implement.yml` publish / publish-source path actually consumed by this
   repair. Do not expand into a wholesale A-004 comment sweep.
4. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   workflow, tests, or comments. `implement.yml` is in the policy fixture
   subset, so consumption is expected. If some files are not in that subset,
   do not copy them merely to force a pin; record non-consumption.
