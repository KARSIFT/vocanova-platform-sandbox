---
evidence_id: VOC-111-EV-00
task_id: VOC-111-T00
acceptance_criteria:
  - VOC-111-AC-00
  - VOC-111-AC-01
  - VOC-111-AC-02
  - VOC-111-AC-03
  - VOC-111-AC-04
tests:
  - VOC-111-TEST-00
  - VOC-111-TEST-01
  - VOC-111-TEST-02
  - VOC-111-TEST-03
  - VOC-111-TEST-04
  - VOC-111-TEST-05
  - VOC-111-TEST-06
  - VOC-111-TEST-08
  - VOC-111-TEST-09
date: 2026-08-22
related_change: VOC-111
gate_status: task-complete — T00-owned portions are satisfied; shared AC-01 remains
  partial until T01/T02 record live docs-only absence proof
live_deployment_claimed: false
---

# VOC-111-T00 — Evidence

Recorded from issue #920 baseline metadata, selector implementation, and deterministic
validation. No secrets, SSH transcripts, session cookies, OAuth state, tokens, or
personal data.

## Issue #920 baseline

Pre-change `deploy-staging.yml` had no push path filter on push. Its header comment
described docs-only deploys as a near-no-op because Docker layers are cached — that
stale claim is removed in T00.

| Merge class | Example ref | deploy-staging run |
|-------------|-------------|-------------------|
| Plan-only | `86df6779` | `32568473144` |
| Roster-only | `60822aa5` | `32568622178` |
| Evidence-only | PR #917 | `32572863842` |

Each pre-change run scheduled the full workflow (~3m39s for PR #917 with unchanged
runtime inputs) including image build/push, SSH deploy, and post-deploy gates.

## Allowlist decision (VOC-111-D03)

Push-triggered `deploy-staging` on `develop` uses GitHub native `on.push.paths`
fail-closed allowlist in `.github/workflows/deploy-staging.yml`:

| Class | Pattern |
|-------|---------|
| Repository-root inputs | `*` (root files only — no `/` in path) |
| Application | `apps/**` |
| Shared packages | `packages/**` |
| Deploy bundle / host assets | `infra/**` |
| Staging e2e gate | `tests/staging-e2e/**` |
| Workflow + selector tests | `.github/workflows/deploy-staging.yml`, `scripts/foundation/voc111-deploy-staging-paths.test.mjs` |

Merges touching only paths outside this list (for example `docs/**`, `specs/**`,
`.karsift/**`) do **not** schedule the workflow on push. `workflow_dispatch` is
unchanged — always eligible for manual retry/redeploy.

Deterministic selector tests in `scripts/foundation/voc111-deploy-staging-paths.test.mjs`
parse this allowlist from the workflow, require its exact closed contents, and evaluate
positive/negative fixtures through `pathSelectsPushDeploy`.

## Changed files (T00)

- `.github/workflows/deploy-staging.yml` — push-only path allowlist; header comment
- `scripts/foundation/voc111-deploy-staging-paths.test.mjs` — selector tests (new)
- `docs/operations/11-devops-and-ci-cd.md` — accurate push selection description
- `specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/t00-evidence.md`

## Validation results

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc111-deploy-staging-paths.test.mjs` | pass |
| `node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs` | pass |
| `node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs` | pass |
| `node --test scripts/foundation/voc094-deploy-concurrency.test.mjs` | pass |
| `bash scripts/governance/validate-governance.sh` | pass |
| `git diff --check` | pass |

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-111-AC-00 | complete |
| VOC-111-AC-01 | partial — T01 fixture and T02 live absence proof remain |
| VOC-111-AC-02 | complete |
| VOC-111-AC-03 | complete |
| VOC-111-AC-04 | complete |
