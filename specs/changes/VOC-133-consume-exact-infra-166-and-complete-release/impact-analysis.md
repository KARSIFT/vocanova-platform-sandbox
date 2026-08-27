# VOC-133 — Impact Analysis

## Security and privacy

This package pins the caller fixture to already-merged infrastructure #166 so
both release jobs restore nested shared policy after the caller owns the
workspace root, and so implementer publication copies lifecycle helpers before
the unrestricted model and classifies a surviving nested path as a distinct
Git checkout or fails closed. It keeps a complete VOC-112 no-change boundary.
It does not introduce new secret values, OAuth/session material,
production-data access, or user-facing data flows. It does not rotate
`KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not change the
`karsift-ai-infra-bot` installation.

Security controls that must remain:

- Restore checkout uses `job.workflow_repository` and `job.workflow_sha` with
  `persist-credentials: false`, not a mutable caller ref and not an operator
  SHA.
- Nested-checkout classification uses a preserved helper copy, not a
  model-controlled path that may have been deleted or replaced with a
  symlink / non-directory / parent-Git directory.
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
  already declared by #164/#165/#166; recovery metadata reads stay on the job
  `GITHUB_TOKEN`.
- Raw helper errors remain sanitized. No credential values are printed in
  logs, tests, or evidence. Evidence must not copy full CI logs or raw
  provider responses from run `33079499176`.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- VOC-112 capture fixtures, provenance test, and hashed sources remain the
  published `develop` bytes and are not recaptured or weakened onto an
  implementation head.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. The #165 restore and #166
nested-checkout classifier are CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-133-R00`: **High correctness risk** if the caller pin remains
  `863fc1f…` or is set to `8ce2b77…` while live `implement.yml@main` /
  `release.yml@main` are already #166. Local suites would then assert a stale
  fixture. Mitigation: `VOC-133-D02`, `VOC-133-AC-00`.
- `VOC-133-R01`: **High release-safety risk** if `identify` restores shared
  policy but `converge` does not, repeating the latent defect that skipped
  promotion in run `33066533397`. Mitigation: `VOC-133-D03`, `VOC-133-AC-02`.
- `VOC-133-R02`: **High correctness risk** if restore uses a mutable ref
  instead of `job.workflow_sha`, or persists credentials, so later helpers
  are not the same immutable reusable-workflow revision used to resolve the
  caller ref. Mitigation: `VOC-133-D03`, `VOC-133-TEST-04`.
- `VOC-133-R03`: **High publication-safety risk** if helpers are still copied
  after the unrestricted model, repeating run `33079499176`'s
  `cp: cannot stat karsift-ai-infra/config/run-app-checks.sh` class.
  Mitigation: `VOC-133-D05`, `VOC-133-AC-03`, `VOC-133-TEST-06`.
- `VOC-133-R04`: **High correctness risk** if a surviving nested path that
  inherits caller Git, is a symlink, or is a non-directory is treated as
  `absent` and discarded. Mitigation: `VOC-133-D05`, `VOC-133-TEST-06`.
- `VOC-133-R05`: **High data-loss risk** if pinning #166 accidentally drops
  #164 unique-develop fail-closed sync or missing-`develop` recovery.
  Mitigation: `VOC-133-D04`, `VOC-133-AC-04`.
- `VOC-133-R06`: **High delivery risk** if this package snapshots the current
  develop/main gap (`karsift-ai-infra#15`), reuses PR #1051 or PR #1056,
  redispatches VOC-132-T00 (#1059), or re-implements VOC-129 / VOC-130 /
  VOC-131 / VOC-132. Mitigation: `VOC-133-D01`, `VOC-133-D09`,
  `VOC-133-AC-07`.
- `VOC-133-R07`: **High correctness / review-safety risk** if the
  implementation weakens
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` missing-subject
  `local` mode, repeating the VOC-131 exhaustion class on #1056 head
  `c11454e717a6d778143de1f2023acc4480305845`. Mitigation: `VOC-133-D10`,
  `VOC-133-AC-06`, `VOC-133-TEST-09`.
- `VOC-133-R08`: **High correctness / review-safety risk** if the
  implementation retargets VOC-112 JSON evidence, repeating the VOC-130
  exhaustion class on #1051 heads `a04a41a…` and `e846cc2…`. Changing
  `voc112-navigation-benchmark-traces.json` also switches
  `repository-governance.yml` into `pr-ancestry` mode. Mitigation:
  `VOC-133-D10`, `VOC-133-AC-06`.
- `VOC-133-R09`: **High coverage risk** if the identity regression again
  guards only the two JSON paths, leaving the provenance test, `AGENTS.md`,
  and the navigator skill editable. That is the VOC-131 root cause.
  Mitigation: `VOC-133-D10`, `VOC-133-TEST-09`.
- `VOC-133-R10`: **High correctness risk** if #166 byte identity depends on a
  machine-specific `/tmp` checkout that CI or another machine does not have.
  Mitigation: `VOC-133-D11`, `VOC-133-TEST-01`.
- `VOC-133-R11`: **High review-safety risk** if `t00-evidence.md` claims a
  protected-path revert that the exact reviewed tree does not contain, as on
  #1056. Mitigation: `VOC-133-D12`, `VOC-133-TEST-10`.
- `VOC-133-R12`: **Medium coverage risk** if caller tests advance the pin but
  do not assert restore-before-helper ordering or helper-copy-before-model
  ordering. Mitigation: `VOC-133-D06`, `VOC-133-TEST-02`, `VOC-133-TEST-03`,
  `VOC-133-TEST-06`.
- `VOC-133-R13`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is workflow/config/test/doc reversion.
- Protected surfaces: caller `tooling/governance/` fixtures and tests,
  possibly live `.github/workflows/*` if `VOC-133-DEP-07` proves a caller
  dispatch file must change, current-state fixture README, and this package
  directory. The five VOC-112 no-change paths are protected against change.
- `VOC-133-DEP-00` through `VOC-133-DEP-09`: see `change.yaml`.
- `VOC-133-EV-00`: T00 evidence — pin SHA, restore coverage, nested-checkout
  coverage, complete VOC-112 no-change boundary, hash-based #166 identity,
  preserved #164 contracts, validation commands, exact reviewed head, final
  infra merge, and VOC-129 / VOC-130 / VOC-131 / VOC-132 / VOC-133 promotion
  handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
fixtures under `tooling/governance/`, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
