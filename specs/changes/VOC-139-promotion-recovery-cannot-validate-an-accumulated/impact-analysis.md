# VOC-139 — Impact Analysis

## Security and privacy

This package repairs promotion-PR source-hash provenance so an accumulated
same-repository `main` ← `develop` promotion whose `AGENTS.md` or navigator
skill legitimately changed on `develop` can pass required `ci / ci`, and it
makes recovery metadata repository-explicit so GitHub CLI does not depend on
a checkout. It does not introduce new secret values, OAuth/session material,
production-data access, or user-facing data flows. It does not rotate
`KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not change the
`karsift-ai-infra-bot` installation.

Security controls that must remain:

- Ordinary non-promotion `pr-validation` stays merge-base hash-anchored.
  Only the authenticated same-repository `main` ← `develop` promotion signal
  binds hashes to `PR_HEAD_SHA`.
- Exact PR base/head SHAs remain required, and base must be an ancestor of head.
  Promotion PRs are not switched to `--squash-safe-push`.
- No `git fetch` / hydrate / recapture path for evidence commits or JSON
  fixtures.
- Recovery metadata addresses `$GITHUB_REPOSITORY` explicitly, validates
  supported owner/repository fields rather than absent
  `.headRepository.nameWithOwner`, and still rejects wrong branch pairs,
  forks, closed PRs, malformed SHAs, or repository identity. Attestations require genuine Actions
  success bound to the PR number, repository, immutable base/head SHAs,
  configured branch pair, expected workflow/path, and `pr-validation` mode.
- Seven VOC-112 capture/runner/source/package.json paths remain the
  published comparison-anchor bytes. The provenance test is the allowed
  exception.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Provenance hash anchoring and recovery
metadata are CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-139-R00`: **High release-safety risk** if promotion `pr-validation`
  keeps requiring stored hashes to equal historical `main`, so #1090 remains
  blocked after VOC-138. Mitigation: `VOC-139-D02`, `VOC-139-AC-00`,
  `VOC-139-TEST-00`.
- `VOC-139-R01`: **High provenance-safety risk** if merge-base hash
  anchoring is dropped for ordinary PRs. Mitigation: `VOC-139-D03`,
  `VOC-139-AC-01`, `VOC-139-TEST-02`.
- `VOC-139-R02`: **High provenance-safety risk** if promotion is inferred
  from hash mismatch alone, without the authenticated `--promotion-pr`
  signal. Mitigation: `VOC-139-D02`.
- `VOC-139-R03`: **High provenance-safety risk** if promotion PRs are
  switched to `--squash-safe-push`, dropping exact SHA/hash negatives.
  Mitigation: `VOC-139-D05`, `VOC-139-AC-04`.
- `VOC-139-R04`: **High recovery-safety risk** if `promotion-pr-metadata`
  still lacks explicit repository context or reads the absent
  `.headRepository.nameWithOwner` projection, repeating job
  `98718739912`. Mitigation: `VOC-139-D07`, `VOC-139-AC-03`,
  `VOC-139-TEST-07`.
- `VOC-139-R05`: **High coverage risk** if the seven remaining VOC-112
  no-change paths are edited, or if VOC-138 package records are rewritten,
  or if live VOC-138 `NO_CHANGE_PATHS` keeps freezing the provenance test.
  Mitigation: `VOC-139-D10`, `VOC-139-TEST-09`.
- `VOC-139-R06`: **High correctness risk** if the caller pin stays at
  `123735c80…` after infra has changed, or is retargeted without mirroring
  the new merge. Mitigation: `VOC-139-D11`, `VOC-139-TEST-10`.
- `VOC-139-R07`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-139-D15`, `VOC-139-TEST-12`.
- `VOC-139-R08`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`) or adds a VOC-097
  evidence-carrier task whose sha_lineage cannot bind to #1090 recovery.
  Mitigation: `VOC-139-D09`, `VOC-139-TEST-13`.
- `VOC-139-R09`: **Medium documentation risk** if
  `docs/operations/11-devops-and-ci-cd.md` keeps claiming promotion PRs are
  merge-base hash-bound. Mitigation: `VOC-139-D12`, `VOC-139-TEST-11`.
- `VOC-139-R10`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow/
  provenance-test reversion plus infra revert of the coordinated PR.
- Protected surfaces: caller `.github/workflows/pipeline.yml`,
  `tooling/governance/` fixtures and tests, the provenance test, this
  package directory, and named current-state docs. Seven VOC-112 paths are
  protected against change.
- `VOC-139-DEP-00` through `VOC-139-DEP-09`: see `change.yaml`.
- `VOC-139-EV-00`: T00 evidence — implementation PR base, new infra merge,
  promotion head-hash rule, ordinary merge-base retention, no-checkout
  metadata, identity negatives, no-fetch/no-recapture proof, pin advance,
  validation after commit, exact-head binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI provenance and recovery behavior, not a product route
or critical API endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/` and mutates application-check provenance, but
the path classifier and independent verifier remain authoritative. This is a
draft proposal, not a determination.
