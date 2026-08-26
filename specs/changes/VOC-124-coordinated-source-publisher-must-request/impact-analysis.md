# VOC-124 — Impact Analysis

## Security and privacy

This package repairs the governed coordinated source-carrier App-token mint
so an authorized nested infrastructure commit that changes
`.github/workflows/**` can be pushed. It does not introduce new secret
values, OAuth/session material, production-data access, or user-facing data
flows. It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`
and does not change the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- The model-controlled implementer runner never receives the GitHub App
  token. It uploads a credential-free Git bundle only.
- `publish-source` mints an infrastructure-scoped App token
  (`repositories: karsift-ai-infra`) with no caller-token fallback.
- The new `permission-workflows: write` request is limited to that
  infrastructure mint. The caller `publish` mint still omits it.
- Caller `publish` still rejects `.github/workflows/**` before push, so an
  unreviewed same-repository PR cannot become a secret-bearing execution
  path.
- Publisher `bundle verify` then fetches only
  `"$PUBLISH_HEAD_SHA:refs/heads/$PUBLISH_BRANCH"`.
- Force-with-lease continues to bind `EXPECTED_SOURCE_HEAD_SHA`.
- Nested `karsift-ai-infra/.git` is never staged as a caller gitlink.
- No credential values are printed in logs, tests, or evidence.
- Infrastructure workflow files belong to the infrastructure repository
  carrier and are independently reviewed there before merge.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-124-R00`: **High delivery risk** if `publish-source` continues to omit
  `permission-workflows: write`. Every coordinated nested infrastructure
  change that touches `.github/workflows/**` fails after a successful bundle
  verify, repeating #1003 / job `98147443377`, with draft caller PR #1012
  stuck on the prior pin. Mitigation: `VOC-124-D01`, `VOC-124-AC-00`.
- `VOC-124-R01`: **High isolation risk** if `permission-workflows` is also
  added to the caller `publish` token, or if the caller workflow-file
  refusal is removed. That would allow an unreviewed same-repository
  workflow PR. Mitigation: `VOC-124-D01`, `VOC-124-D02`, `VOC-124-AC-01`.
- `VOC-124-R02`: **High bootstrap-abuse risk** if the self-hosting exception
  is used to hand-push VOC-122 nested head
  `f90eb630743c8c523e2e6e8dff017acbb31a7f43` or is reused after the repair
  merges. Mitigation: `VOC-124-D04`, `VOC-124-D07`, `VOC-124-AC-06`.
- `VOC-124-R03`: **Medium coverage risk** if tests keep publishing
  non-workflow files, or if `test_live_evidence_reconcile.py` continues to
  treat everything after `\n  publish:` as the caller job and therefore
  either blocks the fix or loses the caller-side `NotIn` assertion.
  Mitigation: `VOC-124-D03`, `VOC-124-TEST-00` through `VOC-124-TEST-02`.
- `VOC-124-R04`: **Medium documentation risk** if the caller `publish` PR
  body continues to say required human approval is pending under active
  A-004, or if current-state text implies the caller publisher gained
  workflows permission. Mitigation: `VOC-124-D05`, `VOC-124-TEST-06`.
- `VOC-124-R05`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored source-carrier publication may open an already-authorized infra
  PR through existing gates.
- Protected surfaces: `KARSIFT/karsift-ai-infra` implement workflow, App-token
  versus job-token split, caller `tooling/governance/` fixtures and tests,
  and this package directory.
- `VOC-124-DEP-00` through `VOC-124-DEP-08`: see `change.yaml`.
- `VOC-124-EV-00`: T00 evidence — token-permission mechanism, caller-versus-
  source isolation proof, validation commands, exact infra SHA, pin
  applicability, bootstrap exhaustion, and #1003 / #1012 retry note.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative.
