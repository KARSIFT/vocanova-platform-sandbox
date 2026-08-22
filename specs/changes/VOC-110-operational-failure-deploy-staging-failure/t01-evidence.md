---
evidence_id: VOC-110-EV-01
task_id: VOC-110-T01
acceptance_criteria:
  - VOC-110-AC-03
  - VOC-110-AC-04
tests:
  - VOC-110-TEST-06
date: 2026-08-22
related_change: VOC-110
cites: VOC-110-EV-00
accountable_owner: unassigned
gate_status: live-verification-pass
live_verification_claimed: true
t00_merge_sha: 07a1f268b2d1f7fb338874fbfe838dda373c4955
verification_run_id: 32570915371
---

# VOC-110-T01 — Live deploy-staging verification

## Scope and outcome

After VOC-110-T00 merged to `develop`, the push-triggered `deploy-staging.yml`
run completed with conclusion `success`. Job `deploy to staging` succeeded. The
repository-controlled observer qualified the run against the exact deployed T00
merge SHA and recorded its allowlisted metadata in
`.karsift/live-evidence/VOC-110-T01.result.json`.

No duplicate open operational-failure issue exists for the
`deploy-staging:failure` fingerprint beyond root issue #911.

## T00 dependency

| Item | Value |
| --- | --- |
| Task | `VOC-110-T00` merged to `develop` |
| Merge commit | `07a1f268b2d1f7fb338874fbfe838dda373c4955` (PR #916) |
| T00 evidence | `t00-evidence.md` in this package directory |

## Live workflow evidence

| Field | Value |
| --- | --- |
| Run | [32570915371](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32570915371) |
| Event | `push` |
| Head branch | `develop` |
| Head SHA | `07a1f268b2d1f7fb338874fbfe838dda373c4955` |
| Workflow conclusion | `success` |
| Workflow duration | 217 seconds |
| Job `deploy to staging` conclusion | `success` |
| Reconciled state | `qualified` |

The exact SHA is the T00 merge commit, so the successful deployment proves the
runtime fix that was independently reviewed and merged in PR #916.

## Contract-lineage correction

The adopted contract initially used `integration_contains_pr_head`. That mode
would require a `develop` deployment to contain this still-unmerged T01 evidence
carrier commit, so no run could ever qualify. The first observe attempt therefore
left the task waiting without publishing evidence.

Before qualification, the operator corrected the contract to `exact_sha` and
pinned it to the T00 merge commit above. This preserves the intended requirement:
the successful run must deploy the exact revision containing the runtime fix. The
repository-controlled reconciler then independently validated the run against
that corrected contract; no result is accepted without its App-authored,
exact-result-head attestation and a subsequent independent review.

An initial qualified result was followed by this narrative completion, which
changed the PR head and intentionally invalidated that result's exact-head
attestation. The stale result was removed before this revision so the normal
WAITING → reconcile → attested-result → re-review sequence can run again. This
document will not be edited after the fresh reconciler result is appended.

## VOC-110-AC-04 — Operational-failure fingerprint

A read-only search of open issues containing the stable
`operational-failure:deploy-staging:failure` marker returned exactly **1** result
at verification time.

| Issue | State at verification | Notes |
| --- | --- | --- |
| #911 | open | Root incident and sole open fingerprint owner |

The green deployment and evidence reconciliation created no duplicate incident.

## Roster closure state

- Task issue #914 remains open until this T01 evidence PR merges; PR #917 carries
  the normal closing reference.
- Root issue #911 can close after PR #917 merges and the package roster confirms
  both T00 and T01 complete.
- No manual or out-of-order issue closure is claimed by this evidence revision.

## Secrets and redaction

No logs, credentials, session values, OAuth state, cookies, tokens, email
addresses, or other personal data are recorded in this evidence.
