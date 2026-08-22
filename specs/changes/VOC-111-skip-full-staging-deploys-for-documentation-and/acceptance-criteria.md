# VOC-111 — Acceptance Criteria

## VOC-111-AC-00 — Problem evidence and scope recorded

- Requirement source: `VOC-111-D00`, issue #920
- Tasks: `VOC-111-T00`
- Tests: `VOC-111-TEST-00`
- Evidence: `VOC-111-EV-00`
- Result: pending

Task evidence records the issue #920 run IDs and merge classes (plan-only `86df6779`,
roster-only `60822aa5`, evidence-only PR #917) and states that today's push trigger
has no path filter. Evidence contains no secrets, full logs, or personal data.

## VOC-111-AC-01 — Push selection skips non-runtime merges

- Requirement source: `VOC-111-D01`, `VOC-111-D04`
- Tasks: `VOC-111-T00`, `VOC-111-T01`
- Tests: `VOC-111-TEST-01`, `VOC-111-TEST-02`, `VOC-111-TEST-07`
- Evidence: `VOC-111-EV-00`, `VOC-111-EV-01`
- Result: pending

After T00 merges to `develop`, a push whose changed-file set includes only
documentation, governed change-package material under `specs/**`, and/or package
evidence carriers does **not** schedule `deploy-staging`. T01 records operator-owned
metadata proving zero matching workflow runs for the verified push SHA.

## VOC-111-AC-02 — Runtime/deployment-relevant pushes remain fail-closed selected

- Requirement source: `VOC-111-D00`, `VOC-111-D03`, `VOC-111-D05`
- Tasks: `VOC-111-T00`
- Tests: `VOC-111-TEST-03`, `VOC-111-TEST-04`, `VOC-111-TEST-05`, `VOC-111-TEST-06`
- Evidence: `VOC-111-EV-00`
- Result: pending

Deterministic selector tests prove that changes under `apps/**`, `packages/**`,
`infra/**`, root workspace manifests/lockfiles, `tests/staging-e2e/**`, and edits to
`.github/workflows/deploy-staging.yml` or the selector test file itself remain
selected for push-triggered deploy. No runtime-relevant fixture silently bypasses
deployment.

## VOC-111-AC-03 — workflow_dispatch and staging deploy semantics preserved

- Requirement source: issue #920 required outcome items 3–4; `VOC-111-D02`
- Tasks: `VOC-111-T00`
- Tests: `VOC-111-TEST-08`, regression on applicable `VOC-084-TEST-*`,
  `VOC-088-TEST-*`, `VOC-094-TEST-*` deploy-staging wiring
- Evidence: `VOC-111-EV-00`
- Result: pending

`workflow_dispatch` retains its existing inputs and full deploy behavior. When a
push **is** selected, the workflow still performs image build/push, SSH deploy,
health/OAuth checks, and the staging core-loop gate with the same fail-closed
semantics, secrets handling, concurrency group, and operational-failure markers as
before this change.

## VOC-111-AC-04 — Stale near-no-op documentation corrected

- Requirement source: issue #920 required outcome item 6
- Tasks: `VOC-111-T00`
- Tests: `VOC-111-TEST-09`
- Evidence: `VOC-111-EV-00`
- Result: pending

The deploy-staging header comment and any affected DevOps documentation no longer
claim that docs-only pushes perform a cheap cached near-no-op deploy. They accurately
describe push selection and manual `workflow_dispatch` retry behavior.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
