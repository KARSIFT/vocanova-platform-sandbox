# VOC-111 — Test Plan

## VOC-111-TEST-00 — Evidence references issue #920 runs and missing path filter

- Covers: `VOC-111-AC-00`
- Preconditions: T00 evidence file drafted at implementation time
- Procedure: Read `t00-evidence.md`; assert it names runs `32568473144`, `32568622178`,
  and `32572863842`, example merge SHAs `86df6779` and `60822aa5`, evidence-only PR #917,
  and that pre-change `deploy-staging` had no push path filter.
- Expected result: Evidence bounds remediation to unnecessary full deploys on non-runtime
  merges
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-01 — Docs-only changed files do not select push deploy

- Covers: `VOC-111-AC-01`, `VOC-111-D01`
- Preconditions: T00 task branch with selector implementation
- Procedure: Run `voc111-deploy-staging-paths.test.mjs` negative fixtures for
  representative `docs/**`-only diffs.
- Expected result: Fixture reports non-selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-02 — Specs/evidence-only changed files do not select push deploy

- Covers: `VOC-111-AC-01`, `VOC-111-D01`
- Preconditions: T00 task branch
- Procedure: Run negative fixtures for `specs/changes/**` plan/roster/evidence-only
  diffs and `.karsift/**` carriers that do not also change allowlisted runtime paths.
- Expected result: Non-selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-03 — Application and shared-package changes remain selected

- Covers: `VOC-111-AC-02`, `VOC-111-D03`, `VOC-111-D05`
- Preconditions: T00 task branch
- Procedure: Run positive fixtures for `apps/api/**`, `apps/web/**`, and
  `packages/**` changes, including at least one shared-package manifest edit.
- Expected result: Selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-04 — Infra/deploy assets and staging e2e changes remain selected

- Covers: `VOC-111-AC-02`
- Preconditions: T00 task branch
- Procedure: Run positive fixtures for `infra/**` and `tests/staging-e2e/**` edits.
- Expected result: Selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-05 — Every repository-root file remains selected

- Covers: `VOC-111-AC-02`
- Preconditions: T00 task branch
- Procedure: Run positive fixtures for `package.json`, `pnpm-lock.yaml`,
  `pnpm-workspace.yaml`, hidden root build inputs, and an otherwise-unlisted future
  root filename. Assert the root-only pattern does not accidentally match nested
  docs/specs files.
- Expected result: Selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-06 — Deploy workflow and selector test edits remain selected

- Covers: `VOC-111-AC-02`
- Preconditions: T00 task branch
- Procedure: Assert `.github/workflows/deploy-staging.yml` and
  `scripts/foundation/voc111-deploy-staging-paths.test.mjs` are allowlisted so selector
  changes cannot silently bypass themselves.
- Expected result: Selection for push trigger
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-07 — Live docs/evidence-only push produces no deploy-staging run

- Covers: `VOC-111-AC-01`
- Preconditions: T00 and the separate T01 docs-only fixture merged to `develop`
- Procedure: Operator resolves T01's completed integration SHA and changed-file set,
  then queries Actions metadata showing zero matching `deploy-staging` `push` runs on
  `develop` for that SHA. Read `t02-evidence.md`.
- Expected result: Absence proof with allowlisted metadata only
- Evidence: `VOC-111-EV-02`

## VOC-111-TEST-08 — workflow_dispatch and selected-push deploy semantics preserved

- Covers: `VOC-111-AC-03`
- Preconditions: T00 task branch
- Procedure: Inspect `deploy-staging.yml` for unchanged `workflow_dispatch` inputs and
  job steps when selected; run applicable `voc084`, `voc088`, and `voc094` deploy-staging
  wiring tests.
- Expected result: Manual dispatch path intact; selected pushes retain full deploy steps,
  concurrency block, and fail-closed guards
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-09 — Stale near-no-op documentation removed

- Covers: `VOC-111-AC-04`
- Preconditions: T00 task branch
- Procedure: Read deploy-staging header comments and any updated DevOps doc sections;
  assert they describe skipped non-runtime pushes rather than cached near-no-op deploys.
- Expected result: Documentation matches implemented behavior
- Evidence: `VOC-111-EV-00`

## VOC-111-TEST-10 — Governed fixture is docs/specs-only and non-circular

- Covers: `VOC-111-AC-01`
- Preconditions: T00 merged; T01 task PR open
- Procedure: Inspect the exact T01 task diff and assert every changed path is under
  this package in `specs/changes/**`; assert T01 records only pre-merge fixture
  metadata and leaves integration-SHA/run-absence claims to T02.
- Expected result: T01 can merge to create a bounded non-runtime push without claiming
  evidence about its own future merge
- Evidence: `VOC-111-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
