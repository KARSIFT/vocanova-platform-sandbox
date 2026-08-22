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
date: pending
related_change: VOC-111
gate_status: pending-implementation
live_deployment_claimed: false
---

# VOC-111-T00 — Evidence (pending)

Implementer completes this file during `VOC-111-T00`. Do not copy secrets, SSH
transcripts, session cookies, OAuth state, tokens, or personal data.

## Issue #920 baseline (to record at implementation time)

| Merge class | Example ref | deploy-staging run |
|-------------|-------------|-------------------|
| Plan-only | `86df6779` | `32568473144` |
| Roster-only | `60822aa5` | `32568622178` |
| Evidence-only | PR #917 | `32572863842` |

Record confirmation that pre-change `deploy-staging.yml` had **no** push path filter
and that the header comment described docs-only deploys as a near-no-op.

## Allowlist decision

Document the final push allowlist implemented in `.github/workflows/deploy-staging.yml`
and any root-file handling approach. Must match `VOC-111-D03` or document justified
extensions with matching tests.

## Validation results (to fill at implementation time)

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc111-deploy-staging-paths.test.mjs` | pending |
| Applicable voc084/voc088/voc094 deploy-staging tests | pending |
| `bash scripts/governance/validate-governance.sh` (if required) | pending |
| `git diff --check` | pending |

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-111-AC-00 | pending |
| VOC-111-AC-01 | partial — T01 fixture and T02 live absence proof remain |
| VOC-111-AC-02 | pending |
| VOC-111-AC-03 | pending |
| VOC-111-AC-04 | pending |
