# VOC-130 — Impact Analysis

## Security and privacy

This package pins the caller fixture to already-merged infrastructure #165 so
both release jobs restore nested shared policy after the caller owns the
workspace root. It does not introduce new secret values, OAuth/session
material, production-data access, or user-facing data flows. It does not
rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not change
the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- Restore checkout uses `job.workflow_repository` and `job.workflow_sha` with
  `persist-credentials: false`, not a mutable caller ref and not an operator
  SHA.
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
  already declared by #164/#165; recovery metadata reads stay on the job
  `GITHUB_TOKEN`.
- Raw helper errors remain sanitized. No credential values are printed in
  logs, tests, or evidence.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. The #165 restore is CI orchestration
only.

## Risks, dependencies, and evidence

- `VOC-130-R00`: **High correctness risk** if the caller pin remains
  `863fc1f…` while live `release.yml@main` is already #165. Local suites would
  then assert a stale restore-less fixture. Mitigation: `VOC-130-D02`,
  `VOC-130-AC-00`.
- `VOC-130-R01`: **High release-safety risk** if `identify` restores shared
  policy but `converge` does not, repeating the latent defect that skipped
  promotion in run `33066533397`. Mitigation: `VOC-130-D03`, `VOC-130-AC-02`.
- `VOC-130-R02`: **High correctness risk** if restore uses a mutable ref
  instead of `job.workflow_sha`, so later helpers are not the same immutable
  reusable-workflow revision used to resolve the caller ref. Mitigation:
  `VOC-130-D03`, `VOC-130-TEST-04`.
- `VOC-130-R03`: **High data-loss risk** if pinning #165 accidentally drops
  #164 unique-develop fail-closed sync or missing-`develop` recovery.
  Mitigation: `VOC-130-D04`, `VOC-130-AC-03`.
- `VOC-130-R04`: **High delivery risk** if this package snapshots the current
  develop/main gap (`karsift-ai-infra#15`) or re-implements VOC-129. Mitigation:
  `VOC-130-D08`, `VOC-130-AC-05`.
- `VOC-130-R05`: **Medium coverage risk** if caller tests advance the pin but
  do not assert restore-before-helper ordering. Mitigation: `VOC-130-D05`,
  `VOC-130-TEST-02`, `VOC-130-TEST-03`.
- `VOC-130-R06`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is workflow/config/test/doc reversion.
- Protected surfaces: caller `tooling/governance/` fixtures and tests,
  possibly live `.github/workflows/*` if `VOC-130-DEP-07` proves a caller
  dispatch file must change, current-state fixture README, and this package
  directory.
- `VOC-130-DEP-00` through `VOC-130-DEP-07`: see `change.yaml`.
- `VOC-130-EV-00`: T00 evidence — pin SHA, restore coverage, preserved #164
  contracts, validation commands, and VOC-129 / VOC-130 promotion handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
fixtures under `tooling/governance/`, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
