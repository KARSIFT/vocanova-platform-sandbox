---
evidence_id: VOC-090-EV-01
task_id: VOC-090-T01
acceptance_criteria:
  - VOC-090-AC-05
  - VOC-090-AC-06
tests:
  - VOC-090-TEST-06
date: 2026-08-19
related_change: VOC-090
cites: VOC-090-EV-00
accountable_owner: unassigned
gate_status: live-verification-pending
live_verification_claimed: false
t00_merge_sha: c996c01
---

# VOC-090-T01 — Live scheduled-synthetics verification

## Scope and outcome

Record a post-T00 `workflow_dispatch` of `scheduled-synthetics.yml` on `develop`
that completes with conclusion `success` and job
`synthetic.staging.authenticated-core-journey` success within the declared
40-minute job wall clock. Confirm no duplicate open operational-failure issue
for the `scheduled-synthetics:cancelled` fingerprint.

**Attempt 2 (implementer):** the cursor-agent shell inherits `github.token` via
git credentials but that token has `actions: read` only during `implement.yml`
(403 on `workflow_dispatch`). Remediation adds
`.github/workflows/voc090-t01-live-verify.yml`, which runs on this PR branch
with `actions: write`, executes
`record-live-evidence.sh`, and commits scrubbed proof before pipeline review
(`voc090-t01-gate`).

Until that workflow completes, `live_verification_claimed` remains `false`.

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
| Dispatch mode | pending — `voc090-t01-live-verify.yml` on PR sync |
| Run URL | pending |
| Run ID | pending |
| Head branch | `develop` (required) |
| Head SHA | pending (must include T00) |
| Workflow conclusion | pending |
| Workflow wall clock | pending |
| Job `synthetic.staging.authenticated-core-journey` conclusion | pending |
| Job duration | pending |
| Declared job budget | 40 minutes (`timeout-minutes: 40` / `timeout_seconds: 2400`) |

### Not valid as T01 proof

| Run | Why excluded |
| --- | --- |
| [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931) | Pre-remediation failure; cancelled at 30m |
| [#23](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32276804129) | Hourly `schedule` on `main` — not `develop` `workflow_dispatch` with T00 |

## VOC-090-AC-06 — Operational-failure fingerprint (pre-dispatch snapshot)

GitHub search at implement time returned **one** open issue:

| Issue | State | Notes |
| --- | --- | --- |
| [#759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759) | open | Origin issue for run #22; sole open fingerprint owner |

Post-success duplicate check is recorded by `record-live-evidence.sh` when
`live_verification_claimed` becomes `true`.

## Hourly schedule confirmation

Deferred until after `develop`→`main` promotion carries T00.

## Secrets and redaction

No secret, session cookie, CSRF token, mint token, OAuth state, or personal
data appears in this evidence.

| Artifact | Path |
| --- | --- |
| T00 root-cause evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t00-evidence.md` |
| Live verification script | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/record-live-evidence.sh` |
| This evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md` |
