---
evidence_id: VOC-098-EV-01
task_id: VOC-098-T01
acceptance_criteria:
  - VOC-098-AC-03
  - VOC-098-AC-04
tests:
  - VOC-098-TEST-05
  - VOC-098-TEST-06
date: 2026-08-21
related_change: VOC-098
gate_status: awaiting-contract-qualified-run
live_observer_claimed: false
deduplication_claimed: true
supporting_pre_lineage_fixture_claimed: true
---

# VOC-098-T01 evidence — controlled observer proof pending exact lineage

## Current gate

The observer repair has demonstrated the intended live behavior, including
deduplication, but the runs below predate this task PR head and therefore do
**not** satisfy the adjacent `integration_contains_pr_head` contract. The gate
remains open until the repository-controlled reconciler records a new
qualifying observer run after this exact pre-result PR head is contained by
`main`.

`live_observer_claimed: false` is deliberate. The earlier runs are retained as
supporting operational evidence and as the completed create/dedupe inspection;
they are not presented as the contract-qualified run.

## Supporting live outcome (not contract-qualified)

The T00 repair was promoted to the default branch through PR #848 at main SHA
`b0f11c22b71a2f41fb18489641892f96227bea1f`. Normal production deployment run
`32462537764` completed successfully before the fixture began, including public
health, controlled-signup configuration validation, content seed, session mint,
and the non-mutating production smoke suite.

The operator then used the documented, reserved staging synthetic cancellation
fixture twice. No application or production failure was manufactured.

| Observation | Watched run | Watched conclusion | Observer run | Observer conclusion | Result |
| --- | ---: | --- | ---: | --- | --- |
| First | `32462772340` | `cancelled` | `32462796407` | `success` | Created one sanitized issue as the GitHub App |
| Repeat | `32462883655` | `cancelled` | `32462907004` | `success` | Deduplicated to the existing marker; no second issue |

## Issue and containment evidence

- Fixture issue: #849, titled `Operational failure: scheduled-synthetics
  (cancelled)`.
- Marker: `<!-- operational-failure:scheduled-synthetics:cancelled -->`.
- During proof, exactly one open issue owned the marker.
- The issue was GitHub-App-authored and had no labels.
- Its body retained the fixed sanitized schema; no logs, account data,
  credentials, OAuth material, sessions, cookies, tokens, or user identifiers
  were attached.
- The created issue started pipeline run `32462812488`; its
  `plan-from-issue / plan` job was observed in progress and the run was cancelled
  before a fixture package or PR could be published.
- A scrubbed metadata-only proof comment was added, then fixture issue #849 was
  closed on `2026-08-21`.

## Acceptance mapping before reconcile

| Acceptance criterion | Status | Evidence |
| --- | --- | --- |
| VOC-098-AC-03 | pending | The supporting cancellations invoked the repaired observer successfully, but a new run must satisfy the exact PR-head lineage contract. |
| VOC-098-AC-04 | pass | The first observation created one sanitized, unlabeled App-authored issue; the repeat retained exactly one marker owner and created no duplicate. |
| VOC-098-TEST-05 | pending | Awaiting the repository-controlled result record for a new contract-qualified observer run. |
| VOC-098-TEST-06 | pass | Create-or-dedupe, App identity class, marker ownership, sanitization, and duplicate count recorded above. |

This file records only allowlisted operational metadata. It does not contain
logs, secrets, personal data, OAuth state or codes, callback query parameters,
cookies, sessions, or tokens.
