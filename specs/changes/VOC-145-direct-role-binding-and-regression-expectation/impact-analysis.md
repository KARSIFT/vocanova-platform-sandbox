# VOC-145 — Impact Analysis

## Security and privacy

This package reconciles unauthorized live agent-authority bindings so
reviewer, reviewer-retry, and plan-reviewer model/effort/speed changes
cannot remain undeclared after a direct infra `main` mutation and a
rewritten green test. It does not introduce new secret values,
OAuth/session material, production-data access, or user-facing data flows.
It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`. It
does not change the VOC-140 two-token contract.

Security controls that must remain:

- Live `config/roles.yml` is the sole current role-to-model mapping.
  Workflow comments do not override it.
- Missing `CURSOR_API_KEY`, unsupported prefixes, and effort-omitted Grok
  4.6 identifiers fail closed with no silent vendor/model/speed/effort
  fallback.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- Implementer retry remains two attempts. `reviewer_fast_retry` remains a
  bounded review retry, not an extra implementer attempt and not a skip of
  exact-SHA review.
- No OpenAI execution route is added. No credential values are printed in
  logs, tests, or evidence.
- Raw errors remain sanitized.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Role binding is CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-145-R00`: **High authority risk** if unauthorized head `d8720829…`
  remains the live undeclared lineup, or is pinned because self-CI run
  `33443684483` was green. Mitigation: `VOC-145-D04`, `VOC-145-AC-00`,
  `VOC-145-AC-04`, `VOC-145-TEST-00`, `VOC-145-TEST-05`.
- `VOC-145-R01`: **High authority risk** if VOC-117 tests are again
  rewritten to bless a later lineup and that green result is treated as
  adoption evidence. Mitigation: `VOC-145-D03`, `VOC-145-AC-01`,
  `VOC-145-TEST-01`.
- `VOC-145-R02`: **High correctness risk** if Path B is implemented without
  adoption recording `VOC-145-DEP-07`. Mitigation: `VOC-145-D02`;
  implementer must follow Path A unless adoption names Path B.
- `VOC-145-R03`: **High review-safety risk** if retry caps, exact-SHA
  review, provider isolation, or fail-closed model resolution are weakened
  while changing bindings. Mitigation: `VOC-145-D06`, `VOC-145-AC-05`,
  `VOC-145-TEST-07`.
- `VOC-145-R04`: **High correctness risk** if the caller pin stays at
  `8993e867…` after infra has been reconciled, or is retargeted to
  `d8720829…` without a new reviewed merge. Mitigation: `VOC-145-D04`,
  `VOC-145-TEST-05`.
- `VOC-145-R05`: **High coverage risk** if tests only assert that some
  `VOC117_BINDINGS` map exists without checking the authorized six exact
  strings and the historical-versus-current split. Mitigation:
  `VOC-145-D08`, `VOC-145-TEST-00`, `VOC-145-TEST-01`.
- `VOC-145-R06`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-145-D09`, `VOC-145-TEST-08`.
- `VOC-145-R07`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), recaptures VOC-112
  fixtures, or is used to bypass issue #1120. Mitigation: `VOC-145-D07`,
  `VOC-145-D12`, `VOC-145-TEST-09`.
- `VOC-145-R08`: **Medium documentation risk** if README/CHANGELOG/fixture
  README still describe `effort=high,fast=false` for all review roles after
  Path B, or still omit the ungoverned `xhigh` drift after Path A restore.
  Mitigation: `VOC-145-D05`, `VOC-145-TEST-04`, `VOC-145-TEST-06`.
- `VOC-145-R09`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow reversion
  plus infra revert of the coordinated PR.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, this
  package directory, named current-state docs, and reusable `config/roles.yml`
  plus VOC-117 regression tests.
- `VOC-145-DEP-00` through `VOC-145-DEP-10`: see `change.yaml`.
- `VOC-145-EV-00`: T00 evidence — implementation PR base, new infra merge,
  authorized path, binding table, historical-versus-current split,
  validation after commit, exact-head binding contract, and release
  handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI agent-authority reconciliation, not a product route or
critical API endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered by
this package. Issue #1120 remains a separate stopped VOC-112 provenance EHR
operation and is not this package's trigger or carrier.

This draft proposes **R4** because it changes protected CI/CD
agent-authority routing under `tooling/governance/` and live
`config/roles.yml`, but the path classifier and independent verifier remain
authoritative. This is a draft proposal, not a determination.
