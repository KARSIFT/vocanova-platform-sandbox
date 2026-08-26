# VOC-124 — Acceptance Criteria

## VOC-124-AC-00 — Infrastructure publisher token requests workflow-write

- Requirement source: `VOC-124-D01`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-00`, `VOC-124-TEST-01`
- Evidence: `VOC-124-EV-00`
- Result: pending

On the exact reviewed infrastructure revision, the `publish-source` job's
`actions/create-github-app-token` step requests `permission-workflows: write`
in addition to contents, issues, and pull-requests writes, still scoped to
`karsift-ai-infra`. A mint that omits `permission-workflows` is a failing
result even if GitHub later happens to accept a non-workflow commit.

## VOC-124-AC-01 — Caller publisher still omits workflow-write and refuses workflow files

- Requirement source: `VOC-124-D01`, `VOC-124-D02`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-02`
- Evidence: `VOC-124-EV-00`
- Result: pending

The caller `publish` job token mint does not request
`permission-workflows: write`. That job still rejects
`.github/workflows/**` changes before push. Adding workflow-write to the
caller token, or removing the caller workflow-file refusal, is a failing
result.

## VOC-124-AC-02 — An authorized workflow-file source bundle is covered by the required permission

- Requirement source: `VOC-124-D01`, `VOC-124-D03`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-01`
- Evidence: `VOC-124-EV-00`
- Result: pending

Deterministic tests prove that an authorized nested source bundle whose
commit changes `.github/workflows/**` is not rejected by the
`publish-source` publication script's own checks and that the token mint
covering that push requests `permission-workflows: write`. Tests do not mint
a real App token or print credentials. The #1013 / job `98147443377` class
is kept as a regression of a mint that omitted that permission.

## VOC-124-AC-03 — Missing credentials, invalid bundles, stale bases, and stale leases still fail closed

- Requirement source: `VOC-124-D02`, `VOC-124-D03`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-03`, `VOC-124-TEST-04`
- Evidence: `VOC-124-EV-00`
- Result: pending

Missing App credentials still fail closed with no `github.token` fallback on
`publish-source`. Missing or unverifiable bundles, malformed SHA/branch
metadata, stale integration bases, and stale or racing branch leases still
fail closed. Exact bundle/head/base verification, isolated bare-repository
publication, force-with-lease, and protected-branch controls remain.

## VOC-124-AC-04 — VOC-121/VOC-123 isolation, tokens, lease, and retry limits remain

- Requirement source: `VOC-124-D02`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-05`
- Evidence: `VOC-124-EV-00`
- Result: pending

The change does not weaken nested-repository isolation, gitlink refusal,
VOC-123 named-ref bundle tips, credential-free source-bundle handoff, clean
`publish-source` App-token separation (no caller-token fallback),
`bundle verify` plus authorized-SHA fetch, force-with-lease against
`EXPECTED_SOURCE_HEAD_SHA`, two-attempt implementer bound, non-closing source
PR reference, or caller `Closes #N`. It does not mix the GitHub App token
onto the model-controlled runner or bundle secrets.

## VOC-124-AC-05 — Current-state A-004 text in this carrier is corrected; historical records stay

- Requirement source: `VOC-124-D05`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-06`
- Evidence: `VOC-124-EV-00`
- Result: pending

Current-state `implement.yml` carrier comments and the caller `publish` PR
body no longer say that required human approval is still pending as an
engineering-workflow merge gate under active A-004. They still require
independent exact-revision review and still say the PR is not authorized to
merge on its own. Historical CHANGELOG and other audit records that describe
past policy are not rewritten. Source-carrier current-state text describes
the infrastructure token's `workflows: write` request without implying that
the caller publisher gained that permission.

## VOC-124-AC-06 — Bootstrap is exhausted; existing VOC-122 carrier is retried; pin follows the reviewed infra merge

- Requirement source: `VOC-124-D04`, `VOC-124-D06`, `VOC-124-D07`
- Tasks: `VOC-124-T00`
- Tests: `VOC-124-TEST-07`
- Evidence: `VOC-124-EV-00`
- Result: pending

Evidence records the bounded supervised bootstrap infra PR, its exact
reviewed head and merge SHA, separate merger, and exhaustion before the
normal caller carrier resumed. The bootstrap was not used to publish VOC-122
nested head `f90eb630743c8c523e2e6e8dff017acbb31a7f43`. Both repositories are
validated on their exact reviewed revisions. If the caller fixture consumes
the infrastructure change, `PINNED_SHA.txt` equals that exact infra merge
SHA and matching caller pin assertions are advanced. After that merge is live
on `implement.yml@main`, the existing `VOC-122-T00` carrier is re-dispatched
or reconciled (not replaced); its authoritative infrastructure PR is
published from a newly verified bundle, independently reviewed, and merged;
caller PR #1012 is then updated to that exact infrastructure merge SHA with
truthful evidence and is not merged by this package.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
