# VOC-124 — Test Plan

## VOC-124-TEST-00 — publish-source mint requests permission-workflows: write

- Covers: `VOC-124-AC-00`
- Preconditions: final `implement.yml` on the implementation branch; no secrets
  or production data
- Procedure: Isolate the `publish-source:` job from the caller `publish:`
  job (do not split on `\n  publish:` and treat the remainder as one job).
  Assert the `Mint least-privilege App token for infrastructure repository`
  `with:` block contains `permission-workflows: write`,
  `permission-contents: write`, `permission-issues: write`,
  `permission-pull-requests: write`, and `repositories: karsift-ai-infra`.
  Assert the caller `publish` mint `with:` block does not contain
  `permission-workflows`. Keep this as a regression of pipeline run
  `32958526215` / job `98147443377`, not only as a comment.
- Expected result: only the infrastructure publisher token requests
  workflow-write. A later successful non-workflow publish does not satisfy
  this test.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-01 — Authorized workflow-file source bundle is covered by the required permission

- Covers: `VOC-124-AC-00`, `VOC-124-AC-02`
- Preconditions: the existing
  `tests/test_voc121_source_carrier_publisher.py` publisher fixture class, or
  an equivalent real-temporary-repository helper; no real App token
- Procedure: Create an authorized nested source bundle whose commit changes
  `.github/workflows/**` (for example a fixture
  `recover-actions-checks.yml` edit of the #1013 class). Drive the
  `Publish exact infrastructure bundle from an isolated bare repository`
  script with fixture remotes as existing VOC-121 publisher tests do. Assert:
  - the source publisher script does not fail with the caller-publisher
    `cannot publish workflow-file changes` rejection;
  - lineage, lease, and bundle-verify checks still apply to that commit;
  - the covering `publish-source` mint requests `permission-workflows: write`.
  Do not call GitHub and do not mint `KARSIFT_BOT_*` credentials. Existing
  VOC-121 tests that publish a non-workflow file do not satisfy this test by
  themselves.
- Expected result: the source publisher retains its own lineage, lease, and
  bundle checks without inheriting the caller publisher's workflow-file
  rejection, and its token mint includes the permission required to prevent
  GitHub's #1013 App-token workflow-file push rejection.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-02 — Caller publisher still refuses workflow files without workflow-write

- Covers: `VOC-124-AC-01`
- Preconditions: caller `publish` job in `implement.yml`; existing
  `tests/test_implementer_bundle.py` workflow-file rejection coverage
- Procedure: Confirm the caller `publish` job still contains
  `cannot publish workflow-file changes` and still omits
  `permission-workflows: write` on its own mint. Keep the existing
  real-repository test that a caller bundle changing `.github/workflows/`
  fails closed. Update `tests/test_live_evidence_reconcile.py` so its
  `publish_job` assertions inspect only the caller `publish` job, not
  `publish-source`.
- Expected result: caller same-repository workflow publication remains
  refused and still lacks workflows permission.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-03 — Missing App credentials fail closed with no caller-token fallback

- Covers: `VOC-124-AC-03`, `VOC-124-AC-04`
- Preconditions: `publish-source` job text
- Procedure: Keep and, if needed, extend
  `test_source_publisher_requires_app_token_without_caller_token_fallback`:
  missing `KARSIFT_BOT_APP_ID` or `KARSIFT_BOT_PRIVATE_KEY` fails closed;
  `PUBLISH_TOKEN` remains `${{ steps.app-token.outputs.token }}` with no
  `|| github.token`. Do not print secret values.
- Expected result: cross-repository fallback remains impossible.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-04 — Invalid bundles, stale bases, and stale leases still fail closed

- Covers: `VOC-124-AC-03`
- Preconditions: existing `tests/test_voc121_source_carrier_publisher.py`
  publisher fixture
- Procedure: Keep existing cases for missing bundle, malformed metadata,
  unverifiable lineage, advanced integration history, and racing/stale
  `EXPECTED_SOURCE_HEAD_SHA` lease. Add a workflow-file bundle to that matrix
  only if needed to prove those checks still apply when the commit touches
  `.github/workflows/**`.
- Expected result: token-permission repair does not loosen bundle, base, or
  lease fail-closed behavior.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-05 — Isolation, named-ref bundle, lease, and retry limits remain

- Covers: `VOC-124-AC-04`
- Preconditions: final implementation branch
- Procedure: Inspect `implement.yml` and existing VOC-121/VOC-123 policy
  tests. Confirm: helpers still copied before nested removal; named-ref
  `create-bundle` still used; `publish-source` still requires App credentials
  with no `|| github.token` fallback; force-with-lease still uses
  `EXPECTED_SOURCE_HEAD_SHA`; publisher still fetches
  `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"` after `bundle verify`;
  two-attempt bound unchanged; source PR body still `Relates to` without a
  closing keyword; caller still `Closes #N`; no gitlink; no App token on the
  model-controlled runner; no secrets in bundles.
- Expected result: workflow-write lands without weakening VOC-121/VOC-123
  safety gates.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-06 — Carrier current-state text matches active A-004

- Covers: `VOC-124-AC-05`
- Preconditions: final `implement.yml` and `karsift-ai-infra/README.md`
- Procedure: Assert the caller `publish` PR body no longer contains
  `required human approval are still pending`. Assert current-state
  source-carrier comments/README describe the infrastructure token's
  `workflows: write` request and still describe caller workflow-file refusal.
  Assert `karsift-ai-infra/CHANGELOG.md` historical isolated-publication
  entry that says the caller token has no workflows permission is not
  rewritten.
- Expected result: live carrier text matches A-004; historical audit text
  stays.
- Evidence: `VOC-124-EV-00`

## VOC-124-TEST-07 — Source self-CI, caller suites, docs, pin, bootstrap, and VOC-122 retry record

- Covers: `VOC-124-AC-06`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. If the fixture consumed the infra change,
  assert the pin equals the exact reviewed infra merge and that matching
  foundation pin literals match. If not, assert the pin is unchanged and
  evidence records why. Confirm `t00-evidence.md` records the bootstrap PR,
  exact reviewed head, merge SHA, separate merger, exhaustion, explicit
  non-publication of `f90eb630743c8c523e2e6e8dff017acbb31a7f43`, and that
  #1003 / #1012 remain the existing VOC-122 carrier to retry against that
  revision.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable; bootstrap cannot be reused; VOC-122 is
  not reimplemented here.
- Evidence: `VOC-124-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
