---
evidence_id: VOC-111-EV-02
task_id: VOC-111-T02
acceptance_criteria:
  - VOC-111-AC-01
tests:
  - VOC-111-TEST-07
date: 2026-08-22
related_change: VOC-111
cites:
  - VOC-111-EV-00
  - VOC-111-EV-01
gate_status: complete
live_absence_claimed: true
---

# VOC-111-T02 — Live zero-run proof for the T01 fixture

The operator completed this file only after the governed T01 fixture PR merged to
`develop`. The accepted result is the absence of a push-triggered `deploy-staging`
run for that exact integration SHA, so success-run reconciliation is not applicable.

Do not copy secrets, workflow logs, session values, OAuth data, or personal data.

## Completed fixture push

| Item | Value |
|------|-------|
| T01 task PR | #927 |
| T01 reviewed head SHA | `6ab4bdb18509533d66e69496f6a0627a45df6e33` |
| Integration push SHA | `34ef014f39423be05ac212f9aadd7cf38cb11d6d` |
| Merged at | `2026-08-22T13:47:09Z` |
| Changed paths | only `specs/changes/VOC-111-skip-full-staging-deploys-for-documentation-and/t01-evidence.md` |

## Actions absence query

The recorded allowlisted metadata shows exactly zero workflow runs where workflow is
`deploy-staging`, event is `push`, branch is `develop`, and `head_sha` equals the T01
integration SHA.

| Field | Value |
|-------|-------|
| Query timestamp | `2026-08-22T13:52:27Z` |
| Observation window | 5 minutes 18 seconds after the recorded merge time |
| Matching run count | `0` |
| Scrubbed outcome | pass — no `deploy-staging` push run on `develop` matched the integration SHA |

The read-only query selected the `deploy-staging` workflow, `event=push`,
`branch=develop`, and then matched the exact integration SHA. Manual dispatch runs
were excluded. No workflow logs or request data were read or copied.

## Push-event controls

The same integration SHA produced the following completed push runs, confirming
that GitHub delivered and processed the push event during the observation window:

| Workflow | Run | Event / branch | Result |
|----------|-----|----------------|--------|
| `Repository Governance` | `32576777887` | `push` / `develop` | success |
| `controlled-signup-oauth-e2e` | `32576777783` | `push` / `develop` | success |

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-111-TEST-07 | pass — exact-SHA query returned zero matching staging deploy runs |
| VOC-111-AC-01 | pass — docs/specs-only integration push skipped full staging deploy |

Manual `workflow_dispatch` remains available and is not part of this negative proof.
Positive runtime selection remains covered by T00 deterministic tests.
