# VOC-110 — Test Plan

## VOC-110-TEST-00 — Root-cause evidence references run 32566405628 and failing step

- Covers: `VOC-110-AC-00`
- Preconditions: T00 evidence file drafted at implementation time
- Procedure: Read `t00-evidence.md`; assert it names run 32566405628, head SHA
  `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46`, Dependabot PR #859 context, job
  `deploy to staging`, conclusion `failure`, and the log-identified failing step.
- Expected result: Evidence bounds remediation to the real deploy failure, not
  VOC-094 cancellation or VOC-095 install-timeout classes
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-01 — Fix diff matches recorded failure mode

- Covers: `VOC-110-AC-01`
- Preconditions: T00 task branch with fix
- Procedure: Independent reviewer compares task PR diff to `t00-evidence.md` failing
  step and failure class; confirm the change targets that surface.
- Expected result: No unrelated scope expansion; causal link documented when non-obvious
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-02 — deploy-staging.yml retains fail-closed deploy contract

- Covers: `VOC-110-AC-01`
- Preconditions: T00 changes merged or in task branch
- Procedure: Read `.github/workflows/deploy-staging.yml`; assert no new
  `continue-on-error: true` on deploy/health/OAuth/core-loop steps; assert core-loop
  still runs after both health polls; assert VOC-094 concurrency block unchanged
  unless explicitly justified in evidence.
- Expected result: Deploy workflow contract preserved aside from any proven miswiring fix
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-03 — Dependabot context recorded when dependency-related

- Covers: `VOC-110-AC-01`, `VOC-110-D05`
- Preconditions: T00 evidence and task PR
- Procedure: If fix touches lockfile or package versions, assert task PR or evidence
  names the affected package(s) from PR #859 group.
- Expected result: Dependency remediation is traceable, not a silent bulk revert
- Evidence: `VOC-110-EV-00`

## VOC-110-TEST-04 — New or extended deterministic fixture passes

- Covers: `VOC-110-AC-02`
- Preconditions: T00 adds voc110 or extends existing foundation tests
- Procedure: Run `node --test scripts/foundation/voc110-*.test.mjs` (and any extended
  suites named in `t00-evidence.md`).
- Expected result: All VOC-110 foundation tests exit 0
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
