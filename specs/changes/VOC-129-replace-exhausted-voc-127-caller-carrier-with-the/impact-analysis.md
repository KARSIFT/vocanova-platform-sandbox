# VOC-129 — Impact Analysis

## Security and privacy

This package replaces the exhausted VOC-127 caller carrier by pinning the
caller fixture to already-merged infrastructure #164 and wiring live
`reconcile-production-change` dispatch. It does not introduce new secret
values, OAuth/session material, production-data access, or user-facing data
flows. It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`
and does not change the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- Exceptional production-change identity is an adopted authority issue number
  plus the #164 reusable workflow, not free-form SHAs on caller
  `workflow_dispatch`.
- `existing_pr_number` remains implement-only.
- Develop-ref updates in the mirrored #164 helpers use compare-and-swap /
  non-force Git ref mutation. Unique develop commits fail closed and are
  never erased.
- The model-controlled implementer runner never receives the GitHub App
  token and still has no general `actions` permission.
- Release / production-change App token remains limited to the permissions
  already declared by #164; recovery metadata reads stay on the job
  `GITHUB_TOKEN`.
- No credential values are printed in logs, tests, or evidence.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-129-R00`: **High correctness risk** if the caller pin remains
  `a9df74a6…` or `60afda3a…` while live `release.yml@main` is already #164.
  Mitigation: `VOC-129-D02`, `VOC-129-AC-00`.
- `VOC-129-R01`: **High delivery risk** if PR #1041 is published or VOC-127
  is dispatched as attempt `3`. Mitigation: `VOC-129-D01`, `VOC-129-D08`,
  `VOC-129-AC-05`.
- `VOC-129-R02`: **High data-loss risk** if mirrored sync force-updates
  `develop` over unique commits. Mitigation: preserved VOC-127-D02 in the
  #164 fixture, `VOC-129-AC-03`.
- `VOC-129-R03`: **High authority risk** if exceptional reconciliation
  accepts operator SHAs or uses a second action dialect
  (`reconcile-main-to-develop`). Mitigation: `VOC-129-D03`, `VOC-129-AC-02`.
- `VOC-129-R04`: **High coverage risk** if caller tests still omit exact pin
  equality and the #164 checkout-ref / missing-`develop` path, repeating the
  attempt-2 miss. Mitigation: `VOC-129-D05`, `VOC-129-TEST-00`,
  `VOC-129-TEST-02`.
- `VOC-129-R05`: **Medium operational risk** if tree-equivalent develop sync
  schedules a full staging deploy. Mitigation: `VOC-129-DEP-07`,
  `VOC-129-AC-04`.
- `VOC-129-R06`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is workflow/config/test/doc reversion.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, live
  `pipeline.yml` and possibly `deploy-staging.yml`, current-state
  release/branch docs, and this package directory.
- `VOC-129-DEP-00` through `VOC-129-DEP-07`: see `change.yaml`.
- `VOC-129-EV-00`: T00 evidence — pin SHA, fixture #164 identity,
  checkout-ref coverage, live dispatch, validation commands, and
  #1041 / #1039 / #1035 handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
