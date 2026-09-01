# VOC-146 — Impact Analysis

## Security and privacy

This package repairs governance validation so malformed `--base`/`--head`
range metadata cannot be treated as an empty valid diff. It does not
introduce new secret values, OAuth/session material, production-data access,
or user-facing data flows. It does not rotate `KARSIFT_BOT_APP_ID` /
`KARSIFT_BOT_PRIVATE_KEY`.

Security controls that must remain:

- An unresolved commit or invalid three-dot diff fails closed.
- A successful empty diff remains a valid empty change set, not a skip of
  monitoring-impact or risk classification policy.
- VOC-086 `pull_request` missing-range fail-closed remains.
- `--files-from` remains an explicit changed-file source.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Range loading is CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-146-R00`: **High governance-safety risk** if `validate-governance.sh`
  still prints success after Git reports an invalid symmetric difference,
  repeating issue #1127. Mitigation: `VOC-146-D03`, `VOC-146-AC-00`,
  `VOC-146-TEST-00`.
- `VOC-146-R01`: **High governance-safety risk** if `mapfile < <(git diff)`
  remains the load path so nonzero Git status is still swallowed.
  Mitigation: `VOC-146-D02`, `VOC-146-AC-03`, `VOC-146-TEST-03`.
- `VOC-146-R02`: **High governance-safety risk** if revisions are resolved
  but a no-merge-base three-dot diff is still accepted as empty.
  Mitigation: `VOC-146-D05`, `VOC-146-AC-02`, `VOC-146-TEST-02`.
- `VOC-146-R03`: **High governance-safety risk** if a partial `--base` or
  `--head` falls through to working-tree discovery. Mitigation:
  `VOC-146-D06`, `VOC-146-AC-04`, `VOC-146-TEST-04`.
- `VOC-146-R04`: **High coverage risk** if tests only grep wrappers and do
  not invoke the live scripts on nonexistent SHAs. Mitigation:
  `VOC-146-D09`, `VOC-146-TEST-00`.
- `VOC-146-R05`: **High consistency risk** if `classify-change-risk.sh`
  keeps the swallowed-status loader. Mitigation: `VOC-146-D08`,
  `VOC-146-AC-06`, `VOC-146-TEST-06`.
- `VOC-146-R06`: **Medium compatibility risk** if valid PR ranges,
  `--files-from`, or VOC-086 missing-range fail-closed regress.
  Mitigation: `VOC-146-D07`, `VOC-146-AC-05`, `VOC-146-TEST-05`.
- `VOC-146-R07`: **Medium documentation risk** if `AGENTS.md` still describes
  fail-closed only for a missing range. Mitigation: `VOC-146-D10`,
  `VOC-146-AC-07`, `VOC-146-TEST-07`.
- `VOC-146-R08`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-146-D13`, `VOC-146-TEST-08`.
- `VOC-146-R09`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), recaptures VOC-112
  fixtures, or opens an unrelated infra pin. Mitigation: `VOC-146-D11`,
  `VOC-146-TEST-08`.
- `VOC-146-R10`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is script/test/doc reversion.
- Protected surfaces: `scripts/governance/`, named current-state docs
  including `AGENTS.md`, foundation tests, and this package directory.
- `VOC-146-DEP-00` through `VOC-146-DEP-09`: see `change.yaml`.
- `VOC-146-EV-00`: T00 evidence — implementation PR base, range-loading
  repair, negative-case results, validation after commit, exact-head
  binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI fail-closed repair, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected governance-validation
scripts under `scripts/governance/` and mutates fail-closed range handling
used by CI, but the path classifier and independent verifier remain
authoritative. This is a draft proposal, not a determination.
