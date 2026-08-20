---
evidence_id: VOC-090-EV-01
task_id: VOC-090-T01
acceptance_criteria:
  - VOC-090-AC-05
  - VOC-090-AC-06
tests:
  - VOC-090-TEST-06
date: 2026-08-21
related_change: VOC-090
cites: VOC-090-EV-00
accountable_owner: unassigned
gate_status: live-verification-pass
live_verification_claimed: true
t00_merge_sha: c996c01
verification_run_id: 32278474057
---

# VOC-090-T01 — Live scheduled-synthetics verification

## Scope and outcome

Post-T00 `workflow_dispatch` of `scheduled-synthetics.yml` on `develop` completed
with conclusion `success`. Job `synthetic.staging.authenticated-core-journey`
succeeded within the declared 40-minute job wall clock. No duplicate open
operational-failure issue exists for the `scheduled-synthetics:cancelled`
fingerprint.

## T00 dependency

| Item | Value |
| --- | --- |
| Task | `VOC-090-T00` merged to `develop` |
| Merge commit | `c996c01` (PR #764) |
| T00 evidence | `t00-evidence.md` in this package directory |

T00 raised staging core-journey job `timeout-minutes` to **40**, added pnpm and
Playwright browser caching, and aligned registry
`synthetic.staging.authenticated-core-journey.timeout_seconds` to **2400**.

## Live workflow evidence

| Field | Value |
| --- | --- |
| Dispatch mode | `workflow_dispatch` staging-only (`synthetic.staging.authenticated-core-journey`) |
| Run URL | [32278474057](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32278474057) |
| Run ID | `32278474057` |
| Head branch | `develop` |
| Head SHA | `c996c017441f48a3a70bfbb8f985e99530d87146` |
| Workflow conclusion | `success` |
| Workflow wall clock | 1193s (~19m) |
| Job `synthetic.staging.authenticated-core-journey` conclusion | `success` |
| Job duration | 419s (~6m) |
| Declared job budget | 40 minutes (`timeout-minutes: 40` / `timeout_seconds: 2400`) |

### Not valid as T01 proof

| Run | Why excluded |
| --- | --- |
| [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931) | Pre-remediation failure; cancelled at 30m |
| [#23](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32276804129) | Hourly `schedule` on `main` — not `develop` `workflow_dispatch` with T00 |

## VOC-090-AC-06 — Operational-failure fingerprint

GitHub search for open issues containing
`operational-failure:scheduled-synthetics:cancelled` returned **1**
open result(s) at verification time.

| Issue | State | Notes |
| --- | --- | --- |
| [#759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759) | open | Origin issue for run #22; sole open fingerprint owner |

No second open duplicate was created by the green verification run.

## Hourly schedule confirmation

A later full hourly run on promoted `main` also completed successfully:

| Field | Value |
| --- | --- |
| Run | [32414684349](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32414684349) |
| Event | `schedule` |
| Head SHA | `1007bd6e68a3afdcbd440e09fe5c4b54762619f1` |
| Workflow conclusion | `success` |
| Job `synthetic.staging.authenticated-core-journey` conclusion | `success` |
| Job wall clock | 67s |

That revision contains T00 and later promoted remediation. All five canonical
synthetic jobs passed in the same hourly run.

## Secrets and redaction

No secret, session cookie, CSRF token, mint token, OAuth state, or personal
data appears in this evidence.

| Artifact | Path |
| --- | --- |
| T00 root-cause evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t00-evidence.md` |
| This evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md` |
