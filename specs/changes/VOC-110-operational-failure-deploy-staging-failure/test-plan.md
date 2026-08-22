# VOC-110 — Test Plan

## VOC-110-TEST-00 — Root-cause evidence references run 32566405628 and failing step

- Covers: `VOC-110-AC-00`
- Preconditions: T00 evidence file drafted at implementation time
- Procedure: Read `t00-evidence.md`; assert it names run 32566405628, head SHA
  `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46`, Dependabot PR #859 context, job
  `deploy to staging`, conclusion `failure`, failing step
  `Poll staging.vocanova.site/`, HTTP 502 impact, and the Next.js 16.3.1 standalone
  `@swc/helpers` omission.
- Expected result: Evidence bounds remediation to the real deploy failure, not
  VOC-094 cancellation or VOC-095 install-timeout classes
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-01 — Next.js 16.3.2 repair matches recorded failure mode

- Covers: `VOC-110-AC-01`
- Preconditions: T00 task branch with fix
- Procedure: Assert `next` and `@next/eslint-plugin-next` are both stable 16.3.2,
  the lockfile agrees, and the other seven PR #859 updates remain present.
- Expected result: Targeted upstream-fixed dependency repair; no bulk revert
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-02 — Real production web image boots and serves HTTP

- Covers: `VOC-110-AC-01`
- Preconditions: T00 changes merged or in task branch
- Procedure: Build `apps/web/Dockerfile`, start the resulting image without secrets,
  assert the container remains running, and poll its root route for HTTP 2xx; always
  remove the test container/image.
- Expected result: The exact shipped standalone boundary is runnable on Node 24
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-03 — Container runtime check is merge-gating and path-aware

- Covers: `VOC-110-AC-02`, `VOC-110-D06`
- Preconditions: T00 evidence and task PR
- Procedure: Inspect `pipeline.yml` and deterministic fixture. Assert root manifest,
  lockfile, `apps/web/**`, and relevant shared-package changes run the Docker job;
  irrelevant plan/docs-only changes take the cheap no-op path; merge-gate depends on
  the job; runtime failure cannot be converted to success; cleanup always executes.
- Expected result: Relevant deployable-artifact regressions block merge without
  imposing a Docker build on irrelevant changes
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-04 — Deterministic workflow fixture passes

- Covers: `VOC-110-AC-02`
- Preconditions: T00 adds `voc110-web-container-runtime.test.mjs`
- Procedure: Run `node --test scripts/foundation/voc110-*.test.mjs` (and any extended
  suites named in `t00-evidence.md`).
- Expected result: All VOC-110 foundation tests exit 0 and lock the real gate contract
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-05 — Deploy-staging regression foundation suites remain green

- Covers: `VOC-110-AC-02`
- Preconditions: T00 task branch
- Procedure: Run applicable regression suites:
  `node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs`,
  `node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs`,
  `node --test scripts/foundation/voc095-playwright-install.test.mjs`, and any
  other deploy-staging wiring tests referenced in the task PR.
- Expected result: Existing deploy-staging deterministic tests exit 0
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-06 — Live deploy-staging success on develop after fix

- Covers: `VOC-110-AC-03`, `VOC-110-AC-04`
- Preconditions: T00 merged to `develop`; T01 live-evidence contract satisfied
- Procedure: Read `t01-evidence.md` and reconciled `.karsift/live-evidence/VOC-110-T01.result.json`
  (if present); assert qualifying run has conclusion `success`, branch `develop`, and
  SHA lineage per contract.
- Expected result: Post-fix staging deploy gate green; no duplicate open
  `deploy-staging:failure` issue beyond #911
- Evidence: `VOC-110-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
