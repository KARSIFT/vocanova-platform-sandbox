---
evidence_id: VOC-097-EV-01
task_id: VOC-097-T01
acceptance_criteria:
  - VOC-097-AC-01
  - VOC-097-AC-02
  - VOC-097-AC-03
tests:
  - VOC-097-TEST-02
  - VOC-097-TEST-03
  - VOC-097-TEST-04
  - VOC-097-TEST-05
date: 2026-08-21
related_change: VOC-097
gate_status: corrective-validation-pending
live_fixture_claimed: false
---

# VOC-097-T01 — Waiting lifecycle and remediation suppression

## Outcome

The reusable automation now distinguishes pending operator-owned evidence from
an implementation defect with this machine-readable state:

```text
VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE
```

For a current, exact-SHA waiting verdict with green CI, remediation emits
`should_retry=false`; it does not spend attempt 2 and does not call the
implementer. A genuine implementation `FAIL`, CI failure, or reviewer-job error
still enters the existing bounded retry path. If WAITING and FAIL both appear,
FAIL is dominant.

This task records deterministic lifecycle proof. It does not claim the live
fixture reserved for T03/T05 (`live_fixture_claimed: false`).

## Shared policy delivery

| Evidence | Result |
| --- | --- |
| Shared repository issue | `KARSIFT/karsift-ai-infra#45` closed after corrective merge |
| Initial shared policy PR | `KARSIFT/karsift-ai-infra#46` merged as `36c2f54fe0e4100b7b281b4d3be766e098b2ca16` |
| Corrective shared policy PR | `KARSIFT/karsift-ai-infra#47` merged as `4e65103bdf1ea6f3b49c7e423cb5ef681ebcde7c` |
| Corrective reviewed head | `51b5317abf0ab979722e0b00fdc604610ec1f323` |
| Corrective hosted self-CI | Run `32425572722`: actionlint, shellcheck, YAML parse, and 43 policy tests passed |
| Corrective independent exact-SHA verification | PASS with no findings |

The shared policy provides:

- an explicit waiting verdict in the reviewer contract;
- a fail-dominant shared verdict classifier;
- waiting suppression in remediation while preserving genuine failure retry;
- caller-required triggering-head inputs for review, remediation, and merge-gate;
- rollout-compatible reusable input schemas whose missing or invalid values
  still skip review, refuse remediation, and leave merge pending at runtime;
- stale-run checks before reviewer work and remediation decisions;
- a retry head check before model work and a second check immediately before
  publish, enforced by an explicit SHA-valued force-with-lease;
- atomic merge with the reviewed head; and
- deterministic fixtures for waiting, real failure, mixed verdicts, missing or
  invalid SHA, stale SHA, and current SHA.

The implementer workflow still has no general `actions` permission and receives
no workflow inspection or dispatch credential. Operator reconciliation remains
a separate T02 responsibility.

## Calling-repository integration

VocaNova's `pipeline.yml` now:

1. cancels superseded active pull-request runs for the same PR to avoid
   duplicate CI and reviewer/model cost, while a `closed` event cannot cancel
   the source run that just merged the PR;
2. passes the triggering PR head SHA to reusable review, remediation, and
   merge-gate jobs; and
3. leaves the implementer's permissions and secrets unchanged.

`scripts/foundation/voc097-waiting-lifecycle.test.mjs` locks these caller
bindings and the privacy/least-privilege evidence contract.

## Live rollout correction

The initial caller PR `#834` merged at
`435b61bb16074210e3ef015e873000a807ea459d`. Its source pipeline run
`32423431859` completed CI, independent review, remediation decision,
merge-gate status, and automatic merge, but the subsequent `pull_request:
closed` run `32424360422` entered the same concurrency group and canceled the
source run. The source run is therefore recorded as canceled, not successful.

The first shared rollout also made `expected_head_sha` schema-required before
the older default-branch caller had adopted the new inputs. Issue-close runs
`32424363595` and `32424363212` consequently failed at workflow startup with no
jobs. Shared PR `#47` corrected the rollout contract without adding a live-head
fallback: old callers can start, but omission remains a runtime fail-closed
state for review, remediation, and merge.

The initial merge still deployed safely: staging deploy run `32424359429` and
controlled-signup OAuth E2E run `32424359449` both passed at merge commit
`435b61bb16074210e3ef015e873000a807ea459d`. This does not substitute for the
pending corrective caller pipeline proof that the source run itself now ends
successfully after its PR closes.

## Validation

```bash
node --test scripts/foundation/voc097-waiting-lifecycle.test.mjs
pnpm validate
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Local results: the focused T01 tests passed. `pnpm validate` passed workspace
validation, formatting, lint, type checks, all 213 foundation tests, API-client
tests, and web middleware tests. It then reached the API suite and stopped only
because the local WSL environment has no Docker binary for the disposable
Postgres OAuth harness. No application failure was observed, and this is not
recorded as a full local pass. The hosted PR CI environment must complete the
entire suite before merge. Governance structure validation and `git diff
--check` passed. Package, web, and API production builds passed locally.

No secrets, logs, personal identifiers, OAuth material, or account data are
included. T02 owns observe/dispatch reconciliation; T03/T05 own controlled live
proof.
