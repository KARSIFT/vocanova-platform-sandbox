# VOC-130-T00 — Evidence

Task: `VOC-130-T00` — Pin exact infra #165 and restore shared policy after
caller checkout.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1047)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1047 |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Failing/no-op release wake-up | `33066533397` |
| Selected checkout ref | `develop` |
| Observed failure | `karsift-ai-infra/config/task-completion-runner.py` missing after caller root checkout |
| Observed release result | missing validator treated as safe no-op; no audit/promotion; converge skipped |
| Root cause | root caller checkout deletes nested shared policy; restore must follow caller checkout before lifecycle helpers |
| Authoritative infra merge | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (KARSIFT/karsift-ai-infra#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Infra verification | 429 policy tests plus hosted actionlint, shellcheck, YAML parsing, and policy checks |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| Why VOC-129 is not retried | #1046 already merged; this is a new checkout-lifetime defect |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-130 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-130 branch/PR from current `develop`; not a VOC-129 rewrite |
| Infra | consume already-merged #165; do not open a replacement infra PR |
| Pin target | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; promote through repaired `release.yml@main` / `reconcile-release` |
| Attempt | VOC-130-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces (implementation)

Pending implementation. Record the exact fixture files mirrored from
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, every live pin assertion
advanced from `863fc1f…`, any live workflow edit required by
`VOC-130-DEP-07`, and the commands actually run.

## Validation commands (implementation)

Pending implementation. Expected commands are listed in
`implementation-plan.md`. Record exact commands and results here; do not
treat a missing suite as a pass.

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
implementer must not approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists.
- This package promotes through the same path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Release/task/requirement records close with audit comments naming both
  exact promotion merges.
- Closed state alone is not completion proof.
