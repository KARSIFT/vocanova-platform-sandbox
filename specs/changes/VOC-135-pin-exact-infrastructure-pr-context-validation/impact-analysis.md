# VOC-135 — Impact Analysis

## Security and privacy

This package pins the caller fixture to already-merged infrastructure #167 so
application checks receive an immutable PR base/head pair without fetching
evidence, both release jobs restore nested shared policy after the caller
owns the workspace root, and implementer publication copies lifecycle helpers
before the unrestricted model and classifies a surviving nested path as a
distinct Git checkout or fails closed. It keeps a complete eight-path
no-change boundary, forbids relocated hydration/fetch/provenance bypasses
via a complete caller-diff scan, and represents exact-head binding as
App-authored review/check metadata. It does not introduce new secret values,
OAuth/session material, production-data access, or user-facing data flows.
It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does
not change the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- Restore checkout uses `job.workflow_repository` and `job.workflow_sha` with
  `persist-credentials: false`, not a mutable caller ref and not an operator
  SHA.
- Nested-checkout classification uses a preserved helper copy, not a
  model-controlled path that may have been deleted or replaced with a
  symlink / non-directory / parent-Git directory.
- Application-check provenance uses already-present Git objects (CI
  `fetch-depth: 0` and the implementation integration checkout). Missing
  evidence commits are never fetched.
- Exceptional production-change identity remains an adopted authority issue
  number plus `reconcile-production-change`, not free-form SHAs on caller
  `workflow_dispatch`.
- `existing_pr_number` remains implement-only.
- Develop-ref updates in the mirrored helpers use compare-and-swap /
  non-force Git ref mutation. Unique develop commits fail closed and are
  never erased.
- The model-controlled implementer runner never receives the GitHub App
  token and still has no general `actions` permission.
- Release / production-change App token remains limited to the permissions
  already declared by #164/#165/#166/#167; recovery metadata reads stay on
  the job `GITHUB_TOKEN`.
- Raw helper errors remain sanitized. No credential values are printed in
  logs, tests, or evidence. Evidence must not copy full CI logs or raw
  provider responses from PR #1070.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- VOC-112 capture fixtures, provenance test, runner, `validate-workspace.mjs`,
  hashed sources, and `package.json` remain the published `develop` bytes and
  are not recaptured, wrapped, hydrated, or weakened onto an implementation
  head.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection, not
  through a self-referential SHA in the same Git tree.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. The #165 restore, #166 nested-checkout
classifier, #167 immutable PR-context contract, and feasible exact-revision
evidence contract are CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-135-R00`: **High correctness risk** if the caller pin remains
  `863fc1f…` or is set to `8ce2b77…` or `f3d791…` while live
  `implement.yml@main` / `release.yml@main` / `ci.yml@main` are already #167.
  Local suites would then assert a stale fixture. Mitigation: `VOC-135-D02`,
  `VOC-135-AC-00`.
- `VOC-135-R01`: **High release-safety risk** if `identify` restores shared
  policy but `converge` does not. Mitigation: `VOC-135-D03`, `VOC-135-AC-02`.
- `VOC-135-R02`: **High correctness risk** if restore uses a mutable ref
  instead of `job.workflow_sha`, or persists credentials. Mitigation:
  `VOC-135-D03`, `VOC-135-TEST-04`.
- `VOC-135-R03`: **High publication-safety risk** if helpers are still copied
  after the unrestricted model. Mitigation: `VOC-135-D05`, `VOC-135-AC-03`,
  `VOC-135-TEST-06`.
- `VOC-135-R04`: **High correctness risk** if a surviving nested path that
  inherits caller Git, is a symlink, or is a non-directory is treated as
  `absent` and discarded. Mitigation: `VOC-135-D05`, `VOC-135-TEST-06`.
- `VOC-135-R05`: **High data-loss risk** if pinning #167 accidentally drops
  #164 unique-develop fail-closed sync or missing-`develop` recovery.
  Mitigation: `VOC-135-D04`, `VOC-135-AC-04`.
- `VOC-135-R06`: **High provenance-safety risk** if `run-app-checks.sh`, CI,
  or implementer pre-push still run default `local` provenance on a full
  checkout missing `f9d11e23…`, repeating the VOC-134 exhaustion class.
  Mitigation: `VOC-135-D06`, `VOC-135-AC-05`, `VOC-135-TEST-07`.
- `VOC-135-R07`: **High provenance-safety risk** if application checks fetch
  missing VOC-112 evidence commits. Mitigation: `VOC-135-D06`,
  `VOC-135-TEST-07`.
- `VOC-135-R08`: **High delivery risk** if this package snapshots the current
  develop/main gap (`karsift-ai-infra#15`), reuses PR #1051, PR #1056,
  PR #1065, or PR #1070, redispatches VOC-132-T00 (#1059), VOC-133-T00
  (#1063), or VOC-134-T00 (#1068), or re-implements VOC-127 through VOC-134.
  Mitigation: `VOC-135-D01`, `VOC-135-D10`, `VOC-135-AC-10`.
- `VOC-135-R09`: **High correctness / review-safety risk** if the
  implementation weakens
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` missing-subject
  `local` mode, repeating the VOC-131 exhaustion class. Mitigation:
  `VOC-135-D14`, `VOC-135-AC-07`, `VOC-135-TEST-10`.
- `VOC-135-R10`: **High correctness / review-safety risk** if the
  implementation retargets VOC-112 JSON evidence, repeating the VOC-130
  exhaustion class. Changing `voc112-navigation-benchmark-traces.json` also
  switches `repository-governance.yml` into `pr-ancestry` mode. Mitigation:
  `VOC-135-D14`, `VOC-135-AC-07`.
- `VOC-135-R11`: **High coverage risk** if the identity regression again
  guards only a subset of the eight no-change paths, leaving the runner,
  `validate-workspace.mjs`, the provenance test, `AGENTS.md`, or the
  navigator skill editable. Mitigation: `VOC-135-D14`, `VOC-135-TEST-10`.
- `VOC-135-R12`: **High correctness risk** if #167 byte identity depends on a
  machine-specific `/tmp` checkout that CI or another machine does not have.
  Mitigation: `VOC-135-D11`, `VOC-135-TEST-01`.
- `VOC-135-R13`: **High review-safety risk** if `t00-evidence.md` claims a
  protected-path revert that the exact reviewed tree does not contain, as on
  #1056, or if tests rewrite evidence so a prior failed head appears to pass,
  as on #1065. Mitigation: `VOC-135-D12`, `VOC-135-TEST-13`.
- `VOC-135-R14`: **High review-safety risk** if committed evidence is required
  to contain the SHA of the same commit that contains it. Mitigation:
  `VOC-135-D12`, `VOC-135-AC-09`.
- `VOC-135-R15`: **High correctness / review-safety risk** if `package.json`
  is edited to fetch a missing capture commit or to force
  `VOC112_CAPTURE_PROVENANCE_MODE`, repeating VOC-133 attempts 1 and 2.
  Mitigation: `VOC-135-D14`, `VOC-135-AC-08`, `VOC-135-TEST-11`.
- `VOC-135-R16`: **High correctness / review-safety risk** if an import-time
  capture fetch is added to `voc112-navigation-benchmark-run.mjs` or a
  hydrate helper is invoked from `validate-workspace.mjs`, repeating VOC-134
  attempts 1 and 2, or if the same bypass is relocated under a new filename.
  Mitigation: `VOC-135-D14`, `VOC-135-D16`, `VOC-135-AC-08`,
  `VOC-135-TEST-12`.
- `VOC-135-R17`: **High correctness risk** if the no-change regression uses a
  moving `develop` / `origin/develop` ref after merge, or if implementation
  proceeds against a moved tip. Mitigation: `VOC-135-D15`, `VOC-135-TEST-10`.
- `VOC-135-R18`: **Medium coverage risk** if caller tests advance the pin but
  do not assert restore-before-helper ordering, helper-copy-before-model
  ordering, or PR-context pair passing. Mitigation: `VOC-135-D07`,
  `VOC-135-TEST-02`, `VOC-135-TEST-03`, `VOC-135-TEST-06`, `VOC-135-TEST-07`.
- `VOC-135-R19`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is workflow/config/test/doc reversion.
- Protected surfaces: caller `tooling/governance/` fixtures and tests,
  possibly live `.github/workflows/*` if `VOC-135-DEP-07` proves a caller
  dispatch file must change, current-state fixture README, and this package
  directory. The eight no-change paths are protected against change.
- `VOC-135-DEP-00` through `VOC-135-DEP-12`: see `change.yaml`.
- `VOC-135-EV-00`: T00 evidence — pin SHA, restore coverage, nested-checkout
  coverage, PR-context coverage, complete eight-path no-change boundary
  against the immutable carrier-base SHA, complete-diff hydration/bypass
  scan, `package.json` identity, hash-based #167 identity, feasible
  exact-revision evidence, preserved #164 contracts, validation commands, and
  VOC-127 through VOC-135 promotion handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance/workflow fixture behavior, not a product route or critical
API endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
fixtures under `tooling/governance/`, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
