---
evidence_id: VOC-094-EV-01
task_id: VOC-094-T01
acceptance_criteria:
  - VOC-094-AC-05
  - VOC-094-AC-06
tests:
  - VOC-094-TEST-07
  - VOC-088-TEST-09
date: 2026-08-21
related_change: VOC-094
cites: VOC-094-EV-00
gate_status: waiting-for-operator-live-evidence
live_deployment_claimed: false
live_evidence_contract: specs/changes/VOC-094-operational-failure-deploy-staging-cancelled/.karsift/live-evidence/VOC-094-T01.yaml
t00_merge_sha: 7dfd3e000a6b3ed791cd09cd9ef174d8a650b25b
t00_merge_pr: 788
---

# VOC-094-T01 — Live verification: queue posture, supersession hygiene, develop deploy

Recorded for the governed operator-owned live-evidence path (VOC-097 / issue
#823). Deterministic classifier and queue posture proof from the prior stranded
revision remain valid; this file does not re-copy job logs, SSH output, session
values, or personal data.

## T00 merge baseline

| Item | Value |
|------|-------|
| Integration branch | `develop` |
| T00 merge commit | `7dfd3e000a6b3ed791cd09cd9ef174d8a650b25b` |
| T00 merge PR | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/788 |
| Merged at | `2026-08-19T19:27:27Z` |

T00 changes (`queue: max`, benign-cancel classifier, observer wiring) are on
`develop` at the commit above.

## Live-evidence contract

Machine-readable contract:
`.karsift/live-evidence/VOC-094-T01.yaml`

| Field | Declared value |
|-------|----------------|
| Workflow | `deploy-staging.yml` |
| Job | `deploy to staging` |
| Events | `push`, `workflow_dispatch` |
| Branch | `develop` |
| SHA lineage | `integration_contains_pr_head` |
| Mode | observe-only (no `dispatch` block) |

Qualifying proof is recorded only in allowlisted reconcile metadata
(`.karsift/live-evidence/VOC-094-T01.result.json`). No secrets or personal data
belong in governed comments or this evidence file.

## Prior deterministic proof retained (AC-05(b) partial)

The stranded revision documented:

- Live `queue: max` zero-job supersession cancels on runs #299, #302, #303.
- Classifier skip fixtures (`voc094-deploy-concurrency.test.mjs`,
  `voc088-failure-to-issue.test.mjs`) pass for `cancelled` + zero jobs on deploy
  workflows.

That satisfies the classifier and queue portions of AC-05(b) under the package
fallback. It does not replace AC-05(a)'s green head deploy requirement.

## VOC-094-AC-05(a) — Latest develop deploy success

**Result: waiting for governed reconcile.**

The prior stranded attempt recorded run #306 (`32293128673`) for T00 merge SHA
`7dfd3e0` as `cancelled` after a 40-minute job timeout — not concurrency
supersession and out of VOC-094 scope. A subsequent successful `deploy-staging`
`push` run on `develop` whose SHA lineage matches the waiting PR must be
observed through reconcile (hourly poll or manual `reconcile-live-evidence`
`observe` with a scrubbed run id).

## VOC-094-AC-06 — Issue #781 fingerprint hygiene

Issue #781 remains the sole open issue with marker
`operational-failure:deploy-staging:cancelled`. No additional open duplicate
fingerprint issue was observed at migration time.

## Operator action (after waiting review)

1. Confirm PR #791 shows `VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE`, the
   contract on the current head SHA, and an empty diff from `develop` (see
   VOC-097-T04 migration record).
2. After the next successful `deploy-staging` `push` on `develop`, allow hourly
   reconcile or dispatch manual `observe` with the successful run id only.
3. After reconcile writes `VOC-094-T01.result.json`, ensure fresh exact-SHA
   independent review on the new PR head before merge.

## Acceptance mapping

| ID | Result |
|----|--------|
| VOC-094-AC-05(a) | **waiting** — green head deploy via reconcile |
| VOC-094-AC-05(b) | **partial** — deterministic + queue proof retained; live head deploy pending |
| VOC-094-AC-06 | **satisfied** at migration baseline |
