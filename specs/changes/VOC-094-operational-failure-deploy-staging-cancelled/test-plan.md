# VOC-094 — Test Plan

## VOC-094-TEST-00 — Root-cause evidence references run 32290409156 and supersession

- Covers: `VOC-094-AC-00`
- Preconditions: T00 evidence file drafted
- Procedure: Read `t00-evidence.md`; assert it names run 32290409156, concurrency
  group `staging-deploy`, zero jobs started, and the higher-priority waiting request
  annotation (or equivalent API field).
- Expected result: Evidence bounds remediation to queue depth + observer filter,
  not deploy script defects
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-01 — deploy-staging.yml declares queue max with cancel-in-progress false

- Covers: `VOC-094-AC-01`
- Preconditions: T00 workflow changes in task branch
- Procedure: Parse `.github/workflows/deploy-staging.yml` concurrency block; assert
  `group: staging-deploy`, `cancel-in-progress: false`, and `queue: max` present.
- Expected result: Committed workflow enables multi-pending queue without in-progress cancel
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-02 — deploy-production.yml declares queue max with cancel-in-progress false

- Covers: `VOC-094-AC-02`
- Preconditions: T00 workflow changes in task branch
- Procedure: Parse `.github/workflows/deploy-production.yml` concurrency block; assert
  `group: production-deploy`, `cancel-in-progress: false`, and `queue: max` present.
- Expected result: Production deploy queue matches staging posture
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-03 — Classifier skips concurrency-superseded deploy-staging cancel fixture

- Covers: `VOC-094-AC-03`
- Preconditions: T00 classifier/helper changes
- Procedure: Run fixture simulating `deploy-staging` + `cancelled` + zero jobs (and/or
  annotation substring); assert no issue POST occurs.
- Expected result: Benign supersession does not create an issue
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-04 — Classifier still creates issue for deploy failure fixture

- Covers: `VOC-094-AC-03`
- Preconditions: T00 classifier/helper changes
- Procedure: Run fixture simulating `deploy-staging` + `failure` (and `deploy-staging` +
  `cancelled` with jobs started / ambiguous metadata); assert issue POST occurs or
  deduplication path matches VOC-088 behavior.
- Expected result: Real failures and ambiguous cancels remain fail-closed toward issue creation
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-05 — Observer workflow wires classifier before open-failure-issue

- Covers: `VOC-094-AC-03`
- Preconditions: T00 operational-failure-monitoring.yml changes
- Procedure: Read workflow; assert deploy `cancelled` path invokes classifier; assert
  still uses App installation token; assert no job-log ingestion steps added.
- Expected result: Bounded metadata gate preserves VOC-088 identity and sanitization
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-06 — Deploy workflow regression tests remain green

- Covers: `VOC-094-AC-04`
- Preconditions: T00 task branch
- Procedure: Run existing deploy-related foundation tests (e.g.
  `voc084-deploy-staging-oauth.test.mjs`, `voc088-deploy-staging-allowlist.test.mjs`,
  `voc081-deploy-convergence.test.mjs` as applicable to unchanged semantics).
- Expected result: All regression tests pass; only concurrency block differs
- Evidence: `VOC-094-EV-00`

## VOC-094-TEST-07 — Live latest develop deploy success and supersession hygiene

- Covers: `VOC-094-AC-05`, `VOC-094-AC-06`
- Preconditions: T00 merged to `develop`
- Procedure: Inspect latest green `deploy-staging` run; execute or document controlled
  supersession proof per T01; verify open-issue index for
  `<!-- operational-failure:deploy-staging:cancelled -->`.
- Expected result: Latest commit deploys successfully; no spurious duplicate open issue
- Evidence: `VOC-094-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.

## Regression tests (existing, must remain green)

- `VOC-088-TEST-08`–`VOC-088-TEST-11` — failure-to-issue agent (extended, not weakened)
- Deploy OAuth/allowlist/convergence foundation tests touched only if files change
