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
deduplication_claimed: true
supporting_pre_lineage_fixture_claimed: true
---

# VOC-098-T01 evidence — controlled live observer proof

## Contract-controlled outcome

The adjacent `.karsift/live-evidence/VOC-098-T01.result.json`, when present
together with the trusted exact-head App attestation, is the only source that
can satisfy the `integration_contains_pr_head` contract. The operator narrative
does not independently claim that result. Before reconciliation, its absence is
a waiting state; after reconciliation, the result contains allowlisted Actions
metadata only and the independent reviewer evaluates it at the new exact head.
The earlier runs below remain supporting operational evidence for the
create/dedupe inspection; they are not substituted for the contract-controlled
result.

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

## Acceptance mapping

| Acceptance criterion | Status | Evidence |
| --- | --- | --- |
| VOC-098-AC-03 | result-controlled | Satisfied only by the adjacent qualified result and its trusted exact-head App attestation; this narrative makes no independent live-result claim. |
| VOC-098-AC-04 | pass | The first observation created one sanitized, unlabeled App-authored issue; the repeat retained exactly one marker owner and created no duplicate. |
| VOC-098-TEST-05 | result-controlled | Evaluated from the adjacent result record and trusted exact-head App attestation when reconciliation adds them. |
| VOC-098-TEST-06 | pass | Create-or-dedupe, App identity class, marker ownership, sanitization, and duplicate count recorded above. |

This file records only allowlisted operational metadata. It does not contain
logs, secrets, personal data, OAuth state or codes, callback query parameters,
cookies, sessions, or tokens.
