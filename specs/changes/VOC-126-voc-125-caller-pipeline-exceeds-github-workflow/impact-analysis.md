# VOC-126 — Impact Analysis

## Security and privacy

This package repairs the caller and project-repo `workflow_dispatch` contracts
so GitHub will accept VOC-125's `existing_pr_number` operator interface
without deleting an active recovery or verifier capability. It does not
introduce new secret values, OAuth/session material, production-data access,
or user-facing data flows. It does not rotate `KARSIFT_BOT_APP_ID` /
`KARSIFT_BOT_PRIVATE_KEY` and does not change the `karsift-ai-infra-bot`
installation.

Security controls that must remain:

- Operator resume identity is an existing PR number on `pipeline.yml`, not
  free-form SHAs on any caller `workflow_dispatch`.
- The dedicated verifier workflow is read-only: no `secrets: inherit`, no
  App-token mint, no `actions: write`, no model credentials.
- Mutating recovery (`recover-integration-push`,
  `recover-promotion-pr-checks`) stays on `pipeline.yml`.
- The model-controlled implementer runner never receives the GitHub App
  token and still has no general `actions` permission.
- `publish-source` remains an infrastructure-scoped App token with no
  caller-token fallback.
- Caller `publish` still omits `permission-workflows` and still rejects
  `.github/workflows/**` before push.
- Force-with-lease continues to bind expected heads.
- Nested `karsift-ai-infra/.git` is never staged as a caller gitlink.
- No credential values are printed in logs, tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-126-R00`: **High delivery risk** if `pipeline.yml` still has more than
  25 `workflow_dispatch` inputs after adding `existing_pr_number`. GitHub
  rejects the definition before jobs start, repeating run `32977045898`.
  Mitigation: `VOC-126-D01`, `VOC-126-AC-00`.
- `VOC-126-R01`: **High capability-loss risk** if a recovery or verifier
  input/job is dropped or packed into unvalidated JSON to get below 26.
  Mitigation: `VOC-126-D01`, `VOC-126-D03`, `VOC-126-AC-02`.
- `VOC-126-R02`: **High privilege-expansion risk** if the dedicated verifier
  workflow inherits secrets or gains `actions: write`, or if mutating
  recovery is moved onto that workflow. Mitigation: `VOC-126-D03`,
  `VOC-126-D04`, `VOC-126-AC-03`.
- `VOC-126-R03`: **High identity-spoofing risk** if caller dispatch accepts
  free-form SHAs, or if `existing_pr_number` is omitted while claiming
  VOC-125 resume is restored. Mitigation: `VOC-126-D02`, `VOC-126-AC-01`.
- `VOC-126-R04`: **High retry-policy risk** if unusable PR #1024 is merged,
  or VOC-125-T00 is dispatched as attempt `3`. Mitigation: `VOC-126-D08`,
  `VOC-126-AC-05`, `VOC-126-AC-06`.
- `VOC-126-R05`: **Medium coverage risk** if tests assert `existing_pr_number`
  exists but still omit the 25-input maximum, repeating the VOC-125 source
  miss. Mitigation: `VOC-126-D06`, `VOC-126-TEST-00`.
- `VOC-126-R06`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored operator resume may continue an already-authorized VOC-122
  carrier through existing gates.
- Protected surfaces: `KARSIFT/karsift-ai-infra` project-repo templates and
  tests, caller pipeline and dedicated verifier workflow, caller
  `tooling/governance/` fixtures and tests, and this package directory.
- `VOC-126-DEP-00` through `VOC-126-DEP-07`: see `change.yaml`.
- `VOC-126-EV-00`: T00 evidence — input-count mechanism, relocated verifier
  proof, validation commands, exact infra SHA, pin applicability, and
  #1024 / #1022 / #1003 / #1012 handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
