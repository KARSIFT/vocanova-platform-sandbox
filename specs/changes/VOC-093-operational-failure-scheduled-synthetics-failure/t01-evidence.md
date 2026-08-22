---
evidence_id: VOC-093-EV-01
task_id: VOC-093-T01
acceptance_criteria:
  - VOC-093-AC-04
  - VOC-093-AC-05
tests:
  - VOC-093-TEST-06
date: 2026-08-22
related_change: VOC-093
cites: VOC-093-EV-00
gate_status: waiting-for-operator-live-evidence
live_verification_claimed: false
live_evidence_contract: specs/changes/VOC-093-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-093-T01.yaml
t00_merge_sha: c755bc3e1cadb903a0c8528251c16d1c4421e11d
t00_merge_pr: 783
migration_base_sha: 4fc78ff66ba8e0b681302191921f46107a706d01
---

# VOC-093-T01 — Live production route-sweep verification

Evidence for `VOC-093-T01` (`VOC-093-AC-04`, `VOC-093-AC-05`,
`VOC-093-TEST-06`). This revision uses the governed operator-owned live-evidence
path from VOC-097 (issue #823). The implementer does not dispatch or inspect
Actions runs.

## Result summary

| Criterion                                              | Status                                            |
| ------------------------------------------------------ | ------------------------------------------------- |
| VOC-093-AC-04 (green route-sweep run)                  | **Waiting** — governed reconcile only             |
| VOC-093-AC-05 (no duplicate failure fingerprint issue) | **Satisfied** — issue #771 is the sole open owner |
| T00 dependency gate                                    | **Satisfied** — harness fix merged to `develop`   |

## T00 dependency gate

| Item              | Value                                                                         |
| ----------------- | ----------------------------------------------------------------------------- |
| Task              | `VOC-093-T00` (issue #778, closed `completed`)                                |
| Merge PR          | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/783                 |
| Merge SHA         | `c755bc3e1cadb903a0c8528251c16d1c4421e11d`                                    |
| Fix path          | Harness (`coerce_same_origin_redirect_path` in `smoke-test-production.sh`)    |
| Prior failure run | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32288703894 |

T00 evidence: `t00-evidence.md` (`VOC-093-EV-00`).

## Live-evidence contract

Machine-readable contract:
`.karsift/live-evidence/VOC-093-T01.yaml`

| Field       | Declared value                                                        |
| ----------- | --------------------------------------------------------------------- |
| Workflow    | `scheduled-synthetics.yml`                                            |
| Job         | `synthetic.production.authenticated-route-content-sweep`              |
| Event       | `workflow_dispatch`                                                   |
| Branch ref  | protected integration branch `develop`                                |
| SHA lineage | `exact_sha` pinned to the reviewed protected-branch migration base    |
| Dispatch    | `synthetic_id=synthetic.production.authenticated-route-content-sweep` |

Qualifying proof is recorded only in allowlisted reconcile metadata
(`.karsift/live-evidence/VOC-093-T01.result.json` on a new PR head after
qualification). No logs, secrets, session values, OAuth state, or personal data
belong in this file.

The initial in-place reset made the branch identical to `develop`, and GitHub
closed the empty PR before a waiting review could run. This evidence-only
recovery pins the exact protected `develop` revision that the reset selected.
It keeps workflow dispatch on the protected branch while giving GitHub a
reviewable contract diff; the reconciler still writes the result and triggers a
fresh exact-SHA review before merge.

## Operator action (after waiting review)

1. Confirm PR #789 shows `VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE` and the
   exact protected-branch SHA declared by the reviewed contract.
2. Dispatch or observe only through repository-controlled reconcile on protected
   branch `develop` — for
   example `pipeline.yml` with `action=reconcile-live-evidence` and
   `live_evidence_mode=dispatch` plus `live_evidence_pr_number=789`, or hourly
   polling once a qualifying run exists.
3. After reconcile writes `VOC-093-T01.result.json`, ensure fresh exact-SHA
   independent review on the new PR head before merge.

## VOC-093-AC-05 — Operational-failure fingerprint deduplication

At migration time, issue #771 remains the sole open issue with marker
`operational-failure:scheduled-synthetics:failure`. No duplicate fingerprint
issue is expected when the route sweep succeeds.

## Acceptance mapping

| ID            | Result                                   |
| ------------- | ---------------------------------------- |
| VOC-093-AC-04 | **waiting** — reconcile-bound live proof |
| VOC-093-AC-05 | **satisfied** at migration baseline      |
