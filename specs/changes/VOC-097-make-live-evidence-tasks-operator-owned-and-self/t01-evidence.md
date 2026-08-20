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
gate_status: complete
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
| Shared repository issue | `KARSIFT/karsift-ai-infra#45` closed after merge |
| Shared policy PR | `KARSIFT/karsift-ai-infra#46` merged |
| Reviewed head | `a0c769b9f30fa42bfa7a65a67e40920c18827181` |
| Merge commit | `36c2f54fe0e4100b7b281b4d3be766e098b2ca16` |
| Hosted self-CI | Run `32422649940`: actionlint, shellcheck, YAML parse, and 43 policy tests passed |
| Independent exact-SHA verification | PASS with no blocking or non-blocking findings |

The shared policy provides:

- an explicit waiting verdict in the reviewer contract;
- a fail-dominant shared verdict classifier;
- waiting suppression in remediation while preserving genuine failure retry;
- mandatory triggering-head inputs for review, remediation, and merge-gate;
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

1. cancels superseded pull-request runs for the same PR to avoid duplicate CI
   and reviewer/model cost;
2. passes the triggering PR head SHA to reusable review, remediation, and
   merge-gate jobs; and
3. leaves the implementer's permissions and secrets unchanged.

`scripts/foundation/voc097-waiting-lifecycle.test.mjs` locks these caller
bindings and the privacy/least-privilege evidence contract.

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
