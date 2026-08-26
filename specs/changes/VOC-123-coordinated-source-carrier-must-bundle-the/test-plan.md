# VOC-123 — Test Plan

## VOC-123-TEST-00 — Raw-SHA positive tip reproduces empty-bundle

- Covers: `VOC-123-AC-01`, `VOC-123-AC-05`
- Preconditions: a real temporary Git repository with a base commit and one
  child commit; no secrets or production data
- Procedure: Set `base_sha=$(git rev-parse HEAD~1)` and
  `head_sha=$(git rev-parse HEAD)`. Run
  `git bundle create raw.bundle "$base_sha..$head_sha"`. Assert exit 128 and
  stderr contains `Refusing to create empty bundle`. Keep this as a
  regression of pipeline run `32915078678` / job `98017696468`, not only as a
  comment. Also assert that the production `implement.yml` commit step no
  longer uses `"${{ steps.infra-checkout.outputs.base_sha }}..$SOURCE_HEAD_SHA"`
  (or an equivalent raw-SHA positive tip) as the `bundle create` revision.
- Expected result: the #1003 failure class is deterministically reproduced
  and the production workflow no longer contains that revision form.
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-01 — Named-ref source-carrier path advertises the exact committed head

- Covers: `VOC-123-AC-00`, `VOC-123-AC-05`
- Preconditions: same class of temporary nested repository as
  `test_committed_nested_edits_bundle_before_caller_staging_without_gitlink`
  in `tests/test_voc121_implement_policy.py`; isolated nested commit present
- Procedure: Drive the fixed source-bundle creation (extracted helper or the
  exact `implement.yml` commit-step fragment, substituting fixture paths).
  Assert:
  - bundle file is non-empty;
  - `git bundle list-heads` advertises exactly the recorded
    `SOURCE_HEAD_SHA` (and the intended temporary named ref);
  - `git bundle verify` succeeds against a repository that has the
    prerequisite base;
  - the temporary named ref is absent after cleanup;
  - caller index still has no `karsift-ai-infra` gitlink.
  Existing VOC-121 tests that create `source_base..HEAD` or
  `base_sha..$branch` do not satisfy this test by themselves.
- Expected result: the production source-carrier path produces a publishable
  bundle of the exact nested commit, matching issue #1005's named-ref
  reproduction (`base_sha..carrier` succeeds and list-heads advertises that
  ref).
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-02 — Wrong, missing, or multiple advertised heads fail closed

- Covers: `VOC-123-AC-02`, `VOC-123-AC-05`
- Preconditions: fixture that can inject a bundle or list-heads result
- Procedure: Separate cases for (1) bundle that advertises no heads,
  (2) bundle that advertises a SHA other than `SOURCE_HEAD_SHA`,
  (3) bundle that advertises more than the expected head/ref (unrelated
  branch, tag, or extra commit). Assert non-zero exit, no
  `has_source_changes=true` continuation that would upload that bundle, and
  no publisher fetch of the unexpected ref.
- Expected result: only a bundle whose advertised head is exactly the
  bound committed SHA can proceed.
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-03 — Wrong base, malformed SHA, cleanup mismatch, and unrelated refs fail closed

- Covers: `VOC-123-AC-02`, `VOC-123-AC-05`
- Preconditions: real temporary repositories plus the existing
  `tests/test_voc121_source_carrier_publisher.py` publisher fixture
- Procedure:
  - malformed `SOURCE_HEAD_SHA` or `base_sha` (not 40 hex) fails before
    bundle create;
  - prerequisite/base that is not an ancestor of the committed head fails
    closed at create-verify or at publisher `bundle verify` / lineage check;
  - after successful create, if advertised object ≠ recorded
    `SOURCE_HEAD_SHA`, fail closed and do not upload;
  - a local unrelated branch, tag, or extra commit in the nested worktree
    is not advertised and is not fetchable as a publishable head;
  - keep existing publisher cases: missing bundle, stale live head,
    unverifiable lineage, integration-history race.
- Expected result: create-path and publish-path remain fail-closed; unrelated
  objects never become the published branch head.
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-04 — Caller and planner `..HEAD` bundles are proven safe or fixed

- Covers: `VOC-123-AC-03`, `VOC-123-AC-05`
- Preconditions: real temporary repositories; `implement.yml` recovery
  fragment `integration_sha..HEAD` and `plan.yml` fragment `base_sha..HEAD`
- Procedure: After a commit, create each bundle with the production revision
  form. Assert `git bundle list-heads` advertises a named head (`HEAD` and/or
  the current branch) whose object equals `git rev-parse HEAD`, that the
  bundle is non-empty, and that the existing publisher fetch
  `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"` still resolves that
  object. Include a detached-HEAD case if the production checkout can be
  detached. If a path reproduces empty-bundle or advertises an unexpected
  extra head that would be published, apply the same named-ref contract as
  the source carrier and add equivalent regression coverage.
- Expected result: working `..HEAD` paths stay; defective paths are repaired
  and tested. Evidence records which outcome applied to each path.
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-05 — Isolation, App-token split, lease, and retry limits remain

- Covers: `VOC-123-AC-04`
- Preconditions: final implementation branch
- Procedure: Inspect `implement.yml` / `plan.yml` (if changed) and existing
  VOC-121 policy tests. Confirm: helpers still copied before nested removal;
  `publish-source` still requires App credentials with no `|| github.token`
  fallback; force-with-lease still uses `EXPECTED_SOURCE_HEAD_SHA`; publisher
  still fetches `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"` after
  `bundle verify`; two-attempt bound unchanged; source PR body still
  `Relates to` without a closing keyword; caller still `Closes #N`; no
  gitlink; no App token on the model-controlled runner; no secrets in
  bundles.
- Expected result: named-ref repair lands without weakening VOC-121 safety
  gates.
- Evidence: `VOC-123-EV-00`

## VOC-123-TEST-06 — Source self-CI, caller suites, docs, and pin match consumption

- Covers: `VOC-123-AC-06`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. Confirm current-state comments no longer
  present a raw-SHA positive tip as a working source-bundle contract. If the
  fixture consumed the infra change, assert the pin equals the exact reviewed
  infra merge and that `test_voc121_implement_policy.py` plus any
  `scripts/foundation/voc097`, `voc104`, and `voc108` pin literals match. If
  not, assert the pin is unchanged and evidence records why. Confirm
  `t00-evidence.md` records #1003 as a distinct VOC-122 re-dispatch, not as
  work implemented here.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable.
- Evidence: `VOC-123-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
