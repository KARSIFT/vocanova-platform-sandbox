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

Record a post-T00 `workflow_dispatch` of `scheduled-synthetics.yml` that
completes with conclusion `success` and job
`synthetic.staging.authenticated-core-journey` success within the declared
40-minute job wall clock. Confirm no duplicate open operational-failure issue
for the `scheduled-synthetics:cancelled` fingerprint.

**This revision does not claim live green verification.** The implementer
environment for attempt 1 had no `GITHUB_TOKEN` / `gh` authentication and
cannot dispatch GitHub Actions workflows from the cursor-agent sandbox.

## T00 dependency

| Item | Value |
| --- | --- |
| Task | `VOC-090-T00` merged to `develop` |
| Merge commit | `c996c01` (PR #764) |
| T00 evidence | `t00-evidence.md` in this package directory |

T00 raised staging core-journey job `timeout-minutes` to **40**, added pnpm and
Playwright browser caching, and aligned registry
`synthetic.staging.authenticated-core-journey.timeout_seconds` to **2400**.

## Required operator dispatch (not yet executed for this evidence)

Dispatch from **`develop`** after T00 merge. Staging-only proof is sufficient
for the first verification pass:

```bash
gh workflow run scheduled-synthetics.yml \
  --ref develop \
  -f synthetic_id=synthetic.staging.authenticated-core-journey
```

Full-suite proof (empty `synthetic_id`) is also acceptable:

```bash
gh workflow run scheduled-synthetics.yml --ref develop
```

After dispatch, record in the table below:

- run URL and ID
- workflow conclusion
- total wall clock
- job `synthetic.staging.authenticated-core-journey` conclusion and duration
- head SHA (must be `c996c01` or a descendant on `develop` carrying T00)

## Live workflow evidence

| Field | Value |
| --- | --- |
| Dispatch mode | **pending** — operator dispatch required |
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
| [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931) (`32271016931`) | Pre-remediation failure; cancelled at 30m |
| [#23](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32276804129) (`32276804129`, in progress at implement time) | Hourly `schedule` on `main` at `26497973` — does not include T00 caching/timeout on `develop` |

Public API at implement time (2026-08-19): no `workflow_dispatch` run exists
on `develop` with head SHA `c996c01` or later.

## VOC-090-AC-06 — Operational-failure fingerprint (partial, public API)

GitHub search for open issues containing
`operational-failure:scheduled-synthetics:cancelled` returned **one** result:

| Issue | State | Notes |
| --- | --- | --- |
| [#759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759) | open | Origin issue for run #22; sole open fingerprint owner |

No second open duplicate was found at implement time. A successful green
verification run should not open another issue with this fingerprint while #759
remains the owner (regression guard for `VOC-088-TEST-09`).

Issue #759 closure follows normal package roster closure — not part of this
task.

## Hourly schedule confirmation

Deferred: the next schedule-triggered run on `main` will not carry T00 until
`develop`→`main` promotion. After the operator `workflow_dispatch` on
`develop` succeeds, note the first post-promotion hourly `success` here or at
package closure if timing does not allow waiting one hour within the task
window.

## Secrets and redaction

No secret, session cookie, CSRF token, mint token, OAuth state, or personal
data appears in this evidence.

| Artifact | Path |
| --- | --- |
| T00 root-cause evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t00-evidence.md` |
| This evidence | `specs/changes/VOC-090-operational-failure-scheduled-synthetics-cancelled/t01-evidence.md` |
