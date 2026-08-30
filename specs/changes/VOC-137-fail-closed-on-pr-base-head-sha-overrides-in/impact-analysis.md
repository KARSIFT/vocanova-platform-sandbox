# VOC-137 — Impact Analysis

## Security and privacy

This package tightens the existing exhaustive caller-diff bypass scanner so a
caller executable cannot assign `PR_BASE_SHA` or `PR_HEAD_SHA` around
validation or tests merely by choosing a filename that is not
`validate-workspace*` and does not end in `.test.mjs`. That rename bypass
could otherwise mask default `local` fail-closed provenance. The package does
not introduce new secret values, OAuth/session material, production-data
access, or user-facing data flows. It does not rotate `KARSIFT_BOT_APP_ID` /
`KARSIFT_BOT_PRIVATE_KEY` and does not change the `karsift-ai-infra-bot`
installation.

Security controls that must remain:

- The only executable-tree scan exclusion is
  `tooling/governance/fixtures/karsift-ai-infra/**`, so legitimate #167
  `export PR_BASE_SHA=` in mirrored `run-app-checks.sh` is not treated as a
  caller-side override.
- Caller test modules under `tooling/governance/tests/` remain in the
  complete-diff scan because they execute at import/discovery time.
- VOC-112 capture fixtures, provenance test, runner, `validate-workspace.mjs`,
  hashed sources, and `package.json` remain the published comparison-anchor
  bytes.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection, not
  through a self-referential SHA in the same Git tree.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Filename-independent PR SHA override
detection is CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-137-R00`: **High provenance-safety risk** if `PR_SHA_SET_PATTERN`
  remains gated on `validate-workspace` / `.test.mjs`, so
  `scripts/arbitrary-wrapper.sh` continues to pass. Mitigation:
  `VOC-137-D01`, `VOC-137-AC-00`, `VOC-137-TEST-02`.
- `VOC-137-R01`: **High coverage risk** if the existing
  `validate-workspace-wrapper.mjs` case is treated as sufficient.
  Mitigation: `VOC-137-D03`, `VOC-137-TEST-02` through `VOC-137-TEST-05`.
- `VOC-137-R02`: **High coverage / review-safety risk** if
  `tooling/governance/tests/` is added to `SCAN_EXCLUDE_PREFIXES` to avoid
  self-scan after the filename gate is removed (VOC-135 attempt-2 class).
  Mitigation: `VOC-137-D06`, `VOC-137-AC-03`.
- `VOC-137-R03`: **High coverage risk** if widening the scan false-positives
  on the scanner's own pattern or on
  `test_voc136_caller_replacement.py`'s contiguous `export PR_BASE_SHA=`
  literal (VOC-135 attempt-1 class), or if those tests are skipped/xfailed
  instead of made source-safe. Mitigation: `VOC-137-D02`, `VOC-137-D14`,
  `VOC-137-TEST-01`.
- `VOC-137-R04`: **High correctness risk** if this correction retargets
  `PINNED_SHA.txt` or edits the mirrored #167 fixture. Mitigation:
  `VOC-137-D05`, `VOC-137-TEST-07`.
- `VOC-137-R05`: **High review-safety risk** if VOC-136 package records are
  rewritten or PR #1080's missed-gap PASS is reused as this package's
  review. Mitigation: `VOC-137-D11`, `VOC-137-TEST-11`.
- `VOC-137-R06`: **High delivery risk** if this package snapshots the current
  develop/main gap (`karsift-ai-infra#15`) or treats canceled release runs /
  draft PR #1082 as implementation work. Mitigation: `VOC-137-D12`,
  `VOC-137-TEST-12`.
- `VOC-137-R07`: **High review-safety risk** if committed evidence is required
  to contain the SHA of the same commit that contains it. Mitigation:
  `VOC-137-D10`, `VOC-137-TEST-10`.
- `VOC-137-R08`: **High correctness risk** if any of the eight VOC-112
  no-change paths is edited, including the provenance test that already
  assigns `process.env.PR_BASE_SHA` for its own contract tests.
  Mitigation: `VOC-137-D07`, `VOC-137-TEST-08`.
- `VOC-137-R09`: **Medium coverage risk** if `os.putenv` for PR SHA names is
  omitted; the required issue cases still pass, but a trivial Python rename
  of `os.environ[...] =` to `os.putenv` could remain. Compatible
  strengthening is allowed (`specification.md` open question 3) and should
  be included if it does not create a self-scan false positive.
- `VOC-137-R10`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is scanner/test/doc reversion.
- Protected surfaces: caller `tooling/governance/tests/` scanner and
  regressions, and this package directory. Fixture tree and eight VOC-112
  paths are protected against change.
- `VOC-137-DEP-00` through `VOC-137-DEP-09`: see `change.yaml`.
- `VOC-137-EV-00`: T00 evidence — implementation PR base, filename-gate
  removal, source-safe construction, arbitrary-filename negative cases,
  benign/fixture positive controls, pin freeze, validation after commit,
  exact-head binding contract, preserved VOC-136 records, and ordinary
  release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance scanner behavior, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
tests under `tooling/governance/`, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
