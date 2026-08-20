---
evidence_id: VOC-097-EV-02
task_id: VOC-097-T02
acceptance_criteria:
  - VOC-097-AC-03
  - VOC-097-AC-04
  - VOC-097-AC-05
  - VOC-097-AC-06
  - VOC-097-AC-07
tests:
  - VOC-097-TEST-06
  - VOC-097-TEST-07
  - VOC-097-TEST-08
  - VOC-097-TEST-09
  - VOC-097-TEST-10
  - VOC-097-TEST-11
  - VOC-097-TEST-12
  - VOC-097-TEST-13
  - VOC-097-TEST-14
date: 2026-08-20
related_change: VOC-097
gate_status: complete
live_fixture_claimed: false
post_merge_source_run_claimed: false
---

# VOC-097-T02 — Allowlisted observe/dispatch reconciler

## Outcome

Repository-controlled live-evidence reconciliation now lives in
`KARSIFT/karsift-ai-infra`:

| Component | Purpose |
| --- | --- |
| `config/live_evidence_reconcile.py` | Contract parse/validate, allowlisted sanitization, timeout/dedup helpers |
| `config/live-evidence-reconcile-runner.py` | Actions runner using GitHub metadata APIs only |
| `.github/workflows/live-evidence-reconcile.yml` | Reusable observe / dispatch / timeout workflow |
| `tests/test_live_evidence_reconcile.py` | Deterministic negative/positive matrix for TEST-06–14 |

The implementer workflow still has **no** general `actions` permission. Only the
reconcile workflow holds narrowly scoped Actions access (`actions: write` solely
so a declared `dispatch` block may call `gh workflow run` with contract-bound
inputs).

## Mechanism

1. **Observe:** caller `pipeline.yml` listens for any successful
   `workflow_run` completion and forwards the run id to
   `live-evidence-reconcile.yml` in `mode: observe`.
2. **Validate:** the runner loads the waiting task PR's
   `<package>/.karsift/live-evidence/<task_id>.yaml` contract from the PR head,
   fetches run/job metadata via the Actions API, and fail-closes on wrong
   workflow identity, job set, event, branch, SHA lineage, conclusion, or
   `max_age`.
3. **Sanitize:** qualifying wakes post `**Live-evidence reconcile — qualified**`
   plus `LIVE_EVIDENCE: READY FOR RE-REVIEW` and a JSON block containing only
   allowlisted keys (`workflow_*`, `event`, `branch`, `head_sha`, `run_id`,
   `job_ids`, timestamps, bounded `duration_seconds`). Rejects post
   `**Live-evidence reconcile — rejected**` with a sanitized reason only.
4. **Wake:** one idempotent wake per `run_id` chains fresh exact-SHA
   `review.yml` and `merge-gate.yml` (`pr_number` input added for workflow_call
   wake path).
5. **Timeout:** hourly `schedule` runs `mode: timeout`; after the contract or
   default 72-hour bound, a single `**Live-evidence reconcile — timeout
   escalation**` comment is posted and automatic loops stop.
6. **Dispatch (optional):** `workflow_dispatch action=reconcile-live-evidence`
   dispatches only when the contract includes a `dispatch` block mirroring an
   allowlisted workflow and inputs.

Reconcile outputs never call log-download or artifact-download APIs.

## Calling-repository integration

`/.github/workflows/pipeline.yml` now:

- adds `workflow_run: types: [completed]` and hourly `schedule` triggers;
- adds manual `reconcile-live-evidence` dispatch;
- calls `live-evidence-reconcile.yml@main` with `integration_branch: develop` and
  `auto_merge_enabled: "true"`;
- widens top-level `permissions.actions` to `write` so declared dispatch is
  possible while the implementer path remains unchanged.

Deterministic caller locks live in
`scripts/foundation/voc097-reconcile.test.mjs`.

## Consumption note

This task lands the mechanism in the local `karsift-ai-infra/` checkout bundled
with the task PR for review. Production effect for all callers requires merging
the corresponding infra PR to `KARSIFT/karsift-ai-infra@main` before stranded
tasks (T04) can rely on hosted `@main` behavior.

## Validation

```bash
cd karsift-ai-infra && python3 -m unittest discover -s tests -p 'test_*.py'
node --test scripts/foundation/voc097-reconcile.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Local results: infra policy suite (56 tests) passed; focused T02 caller tests
passed; governance validation and `git diff --check` passed.

No secrets, logs, OAuth material, personal identifiers, or token values are
included. T03 expands the full cross-repo fixture matrix; T05 owns controlled
live proof (`live_fixture_claimed: false`).
