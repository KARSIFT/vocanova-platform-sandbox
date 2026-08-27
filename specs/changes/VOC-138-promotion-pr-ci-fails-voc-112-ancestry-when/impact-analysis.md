# VOC-138 — Impact Analysis

## Security and privacy

This package repairs promotion-PR provenance selection so a squash-era VOC-112
subject that is not in the synthetic `main` <- `develop` checkout no longer
deadlocks required `ci / ci`. It does not introduce new secret values,
OAuth/session material, production-data access, or user-facing data flows. It
does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not
change the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- Ordinary fixture-changing PRs stay on `pr-ancestry` and fail closed when
  the captured commit object is missing. An authenticated same-repository
  `main` <- `develop` promotion deterministically uses `pr-validation`,
  independent of whether the captured object is missing, resolvable, or a
  nonancestor.
- Exact PR base/head SHAs remain required for promotion application checks.
  Promotion PRs are not switched to `--squash-safe-push`.
- No `git fetch` / hydrate path for evidence commits.
- Recovery attestations require genuine Actions success bound to the PR number,
  repository, immutable base/head SHAs, configured branch pair, expected
  workflow/path, and `pr-validation` mode, and remain excluded from
  authoritative evidence selection.
- VOC-112 capture fixtures, provenance test, runner, `validate-workspace.mjs`,
  hashed sources, and `package.json` remain the published comparison-anchor
  bytes.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Provenance mode selection and required-check
recovery are CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-138-R00`: **High release-safety risk** if promotion PRs keep selecting
  `pr-ancestry` whenever the capture fixture differs from `main`, so #1090
  remains blocked. Mitigation: `VOC-138-D02`, `VOC-138-AC-00`,
  `VOC-138-TEST-00`.
- `VOC-138-R01`: **High provenance-safety risk** if missing-subject fallback
  applies to ordinary PRs, letting a new capture skip ancestry. Mitigation:
  `VOC-138-D03`, `VOC-138-AC-02`, `VOC-138-TEST-03`.
- `VOC-138-R02`: **High provenance-safety risk** if promotion PRs are switched
  to `--squash-safe-push`, dropping exact SHA/hash negatives. Mitigation:
  `VOC-138-D05`, `VOC-138-AC-03`, `VOC-138-TEST-05`.
- `VOC-138-R03`: **High provenance-safety risk** if the subject is fetched or
  hydrated. Mitigation: `VOC-138-D06`, `VOC-138-AC-04`, `VOC-138-TEST-07`.
- `VOC-138-R04`: **High recovery-safety risk** if `reconcile-release` keeps
  rerunning the doomed `pull_request` job (`98691441027` / `98692552949`)
  or accepts same-head squash-safe dispatch `33122158425` as equivalent proof.
  Mitigation: reject the weaker dispatch and accept only a genuine PR-bound
  `pr-validation` recovery success under `VOC-138-D07`, `VOC-138-AC-05`.
- `VOC-138-R05`: **High coverage risk** if the eight VOC-112 no-change paths
  are edited to make `pr-ancestry` succeed. Mitigation: `VOC-138-D09`,
  `VOC-138-TEST-02`.
- `VOC-138-R06`: **High correctness risk** if the caller pin stays at #167
  after infra has changed, or is retargeted without mirroring the new merge.
  Mitigation: `VOC-138-D10`, `VOC-138-TEST-10`.
- `VOC-138-R07`: **High review-safety risk** if committed evidence is required
  to contain the SHA of the same commit that contains it. Mitigation:
  `VOC-138-D14`, `VOC-138-TEST-12`.
- `VOC-138-R08`: **High delivery risk** if this package snapshots the current
  develop/main gap (`karsift-ai-infra#15`). Mitigation: `VOC-138-D16`,
  `VOC-138-TEST-13`.
- `VOC-138-R09`: **Medium documentation risk** if
  `docs/operations/11-devops-and-ci-cd.md` keeps claiming promotion PRs use
  `squash-safe-push`. Mitigation: `VOC-138-D11`, `VOC-138-TEST-11`.
- `VOC-138-R10`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc reversion plus infra
  revert of the coordinated PR.
- Protected surfaces: caller `.github/workflows/pipeline.yml` when metadata
  plumbing is required, `tooling/governance/` fixtures and tests, this package
  directory, and named current-state docs. Eight VOC-112 paths are protected
  against change.
- `VOC-138-DEP-00` through `VOC-138-DEP-09`: see `change.yaml`.
- `VOC-138-EV-00`: T00 evidence — implementation PR base, new infra merge,
  promotion mode selection, ordinary `pr-ancestry` retention, hash/SHA
  negatives, no-fetch proof, recovery contract, pin freeze/advance,
  validation after commit, exact-head binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI provenance and recovery behavior, not a product route
or critical API endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/`, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
