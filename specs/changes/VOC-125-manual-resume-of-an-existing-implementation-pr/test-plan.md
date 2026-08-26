# VOC-125 — Test Plan

## VOC-125-TEST-00 — Caller and template dispatch expose existing_pr_number only

- Covers: `VOC-125-AC-00`
- Preconditions: final live `.github/workflows/pipeline.yml` and
  `karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml`; no
  secrets or production data
- Procedure: Assert both files' `workflow_dispatch` inputs include
  `existing_pr_number` described as implement-only existing-carrier recovery.
  Assert the `implement:` job `with:` block forwards
  `existing_pr_number: ${{ inputs.existing_pr_number }}` (or the live
  equivalent) and still forwards `attempt`. Assert neither file declares
  operator-typed `expected_head_sha` or `expected_base_sha` under
  `workflow_dispatch`. Keep this as a regression of pipeline run
  `32966618512` / job `98170418081`, not only as a comment.
- Expected result: operator resume identity is a PR number; free-form SHA
  paste is not the caller interface.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-01 — Valid attempt-2 resume binds exact head/base and reuses the carrier

- Covers: `VOC-125-AC-00`, `VOC-125-AC-01`
- Preconditions: fixture PR/branch/review metadata matching the #1020 class
  (open PR, deterministic task branch, App-signed review bound to exact
  head/base/task/package/issue, remote head equal to that head); no live
  mutation of #1012
- Procedure: Drive the new bind helper / workflow step with:
  `attempt=2`, `existing_pr_number` set, SHA inputs empty, matching live PR
  and remote head. Assert derived `expected_head_sha` /
  `expected_base_sha` equal the live PR head/base, `verify-expected-head.py`
  returns `CURRENT`, and the step does not create a replacement branch or PR.
  Repeat with both PR number and matching SHAs supplied; assert agreement.
- Expected result: the #1020 resume class proceeds past the exact-head guard
  on the same carrier.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-02 — Resume does not open a replacement PR or delete the branch

- Covers: `VOC-125-AC-01`
- Preconditions: existing open task PR in the fixture
- Procedure: Assert the attempt-2 success path checks out the existing remote
  task branch rather than `git checkout -b` from integration, and that
  publisher/PR-open logic still reuses the open PR (existing
  `conflicting_existing_pr` / find-PR-for-branch contracts remain). Assert a
  closed or merged PR fails closed instead of being reopened.
- Expected result: the existing carrier is the only carrier.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-03 — Attempt cap and attempt-1 rewrite are fail-closed

- Covers: `VOC-125-AC-02`
- Preconditions: final `implement.yml`
- Procedure: Assert attempt `3` (and any value other than `1` or `2`) fails
  closed. Assert attempt `1` with an existing open task PR or remote
  deterministic task branch fails closed. Assert a valid existing-carrier
  resume requires attempt `2`. Keep the existing `remediate.yml` two-attempt
  cap (`next_attempt` greater than `2` stops).
- Expected result: one-retry maximum is preserved; attempt `2` is not
  reclassified as attempt `1`.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-04 — Every mismatch class fails closed before model or mutation

- Covers: `VOC-125-AC-03`
- Preconditions: bind helper / workflow step; fixture remotes; no secrets
- Procedure: For each class, assert failure before model resolution and
  before any branch mutation:

  | Class | Fixture |
  |-------|---------|
  | Empty binding (#1020) | attempt `2`, existing branch, empty SHAs, empty PR number |
  | Malformed SHA | non-40-hex head or base |
  | Stale remote head | live branch head differs from bound head |
  | Wrong PR | PR number exists but belongs to another branch or repo |
  | Wrong branch | PR head ref is not the deterministic task branch |
  | Wrong task/package/issue | PR metadata does not match dispatched identity |
  | Closed/completed task | PR or authority issue is closed, or task already completed |
  | Foreign review | review exists but wrong author, head, base, task, package, or issue |
  | SHA/PR disagreement | supplied SHAs do not match live PR head/base |
  | Changed remote head during bind | fetch/ls-remote race class already used by source-carrier tests |

  Absent App-signed review with an otherwise valid live open PR remains the
  CI-failure class and must still bind live PR head/base, not guessed SHAs.
- Expected result: each mismatch class fails closed; CI-failure absence of
  review is not treated as foreign review.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-05 — Automatic remediate retry still forwards trusted SHAs

- Covers: `VOC-125-AC-04`
- Preconditions: final `remediate.yml`
- Procedure: Extend `test_retry_reuses_implementer_with_incremented_attempt`
  (or equivalent) to assert the `retry:` job still passes
  `expected_head_sha: ${{ inputs.expected_head_sha }}` and
  `expected_base_sha: ${{ inputs.expected_base_sha }}`, and now also passes
  `existing_pr_number` from `pr_number` / decide outputs. Assert retry still
  requires `should_retry == 'true'` and still uses `next_attempt`.
- Expected result: automatic remediation is not regressed onto an operator
  SHA-paste path.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-06 — Isolation, leases, roles, and credentials remain

- Covers: `VOC-125-AC-05`
- Preconditions: final implementation branch
- Procedure: Inspect `implement.yml`, `remediate.yml`, and existing
  VOC-121/VOC-123/VOC-124 policy tests. Confirm: named-ref `create-bundle`
  still used; `publish-source` still requires App credentials with
  `permission-workflows: write` and no `|| github.token`; caller `publish`
  still omits workflow-write and still refuses `.github/workflows/**`;
  force-with-lease still uses expected heads; two-attempt bound unchanged;
  source PR body still `Relates to` without a closing keyword; caller still
  `Closes #N`; implementer permissions still omit `actions`; Cursor Composer
  remains the implementer route and Cursor Grok remains the reviewer route;
  no OpenAI execution path is re-enabled; no App token on the
  model-controlled runner; no secrets in bundles.
- Expected result: recovery identity lands without weakening prior safety
  gates.
- Evidence: `VOC-125-EV-00`

## VOC-125-TEST-07 — Source self-CI, caller suites, docs, pin, and VOC-122 handoff

- Covers: `VOC-125-AC-06`
- Preconditions: reviewed infra SHA and current
  `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
- Procedure: Run infra unit suite and caller governance/fixture suites listed
  in `implementation-plan.md`. If the fixture consumed the infra change,
  assert the pin equals the exact reviewed infra merge and that matching
  foundation pin literals match. If not, assert the pin is unchanged and
  evidence records why. Assert current-state comments/README describe
  attempt `2` plus `existing_pr_number` and do not present free-form SHA
  paste as the operator interface. Confirm `t00-evidence.md` records that
  #1003 / #1012 remain the existing VOC-122 carrier to resume against that
  revision at attempt `2` with `existing_pr_number=1012`, and that no
  bootstrap exception was used.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA and only when applicable; VOC-122 is not reimplemented here.
- Evidence: `VOC-125-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
