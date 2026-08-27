# VOC-128-T01 — Live scheduled-synthetics verification

Evidence for `VOC-128-T01` (`VOC-128-AC-04`, `VOC-128-AC-05`,
`VOC-128-TEST-06`, `VOC-128-TEST-07`). This task uses the governed
operator-owned live-evidence path from VOC-097. The implementer does not
dispatch or inspect Actions runs.

## Result summary

| Criterion | Status |
|-----------|--------|
| VOC-128-AC-04 (develop schedule success) | **Pending** — waiting for T00 merge and operator-owned live evidence |
| VOC-128-AC-05 (main production jobs success, no duplicate issue) | **Pending** |
| T00 dependency gate | **Pending** |

## Live-evidence contract

Machine-readable contract:

`specs/changes/VOC-128-operational-failure-scheduled-synthetics-failure/.karsift/live-evidence/VOC-128-T01.yaml`

Plan-time lineage is `integration_contains_pr_head` only as a schema-valid
placeholder. Per `VOC-128-D07` / VOC-110, T01's first commit must retarget
`sha_lineage.mode` to `exact_sha` of the T00 merge (or the develop SHA
actually dispatched that contains T00) before entering waiting.

Required develop-run jobs:

- `synthetic.staging.oauth-expected-state`
- `synthetic.staging.authenticated-core-journey`
- `dispatch-production-synthetics-on-main`

The three production jobs must be skipped on that develop run, not failed.
Their success is proven on the follow-on `main` run (AC-05).

## Allowlisted metadata to record (no logs)

| Field | Value |
|-------|-------|
| Develop run ID | pending |
| Develop event | pending (`schedule` or `workflow_dispatch`) |
| Develop head SHA | pending |
| Develop conclusion | pending |
| Main production run ID | pending |
| Main production jobs conclusion | pending |

## Secrets and redaction

No logs, credentials, session values, OAuth state, cookies, tokens, email
addresses, or other personal data are recorded in this evidence.
