# VOC-140 — Impact Analysis

## Security and privacy

This package repairs promotion recovery/selection so a still-running release
carrier cannot be treated as completed `ci / ci`, and it repairs the App
token/API contract so a separate guard-only token can prove the live
non-bypassable production ruleset before the unchanged mutation token
merges to `main`. It does not introduce new secret values,
OAuth/session material, production-data access, or user-facing data flows.
It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.

Security controls that must remain:

- An in-progress or failed release carrier is never trusted recovered
  `ci / ci`. Dedicated `promotion-pr-validation PR #<n>` must be
  completed/successful, or a completed non-carrier `pull_request` `ci / ci`
  run may attest only when it is not confused with a newer carrier check.
- Exact PR base/head SHAs, repository, refs, and `pr-validation` remain
  required. Promotion PRs are not switched to `--squash-safe-push`. Doomed
  `pull_request` `ci / ci` jobs are not rerun as a strategy.
- The production merge guard still requires an effective active
  repository-owned ruleset, a pull-request rule, strict non-empty required
  checks, and `bypass_actors: []`. Omitted or non-array `bypass_actors` fails
  distinctly from a missing guard.
- The existing mutation token remains exactly Contents/Issues/Pull requests
  write and is the sole App token for `gh pr merge` and mutations. A separate
  ephemeral guard token is scoped only to the current caller repository, has
  only Administration write, and is injected only into guard verification
  immediately before merge. It never reaches mutation, status, issue, PR,
  content, or merge commands. The same separation applies in the production
  branch path of `merge-gate.yml`.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

The guard token's `permission-administration: write` is required for GitHub API
visibility and technically carries ruleset-mutation capability. D05/D06,
workflow isolation, and the exhaustive mint allowlist forbid using it to add
bypass actors, change rulesets, or weaken required checks. External activation
requires the `karsift-ai-infra-bot` registration to request Administration: Read and write
and the owner of KARSIFT organization installation `148001476` to approve that
permission. Drafting-time inspection shows `repository_selection: all`; the
guard token's explicit current-repository restriction limits runtime use but
does not narrow the shared App/private-key ceiling across installation-selected
repositories. No App-ID or private-key secret rotation is required. Until hosted
verification returns explicit `bypass_actors: []`, the merge remains fail closed
with the precise approval-and-rerun operator action.

Token separation limits accidental step-level exposure but both tokens are
minted from the same App registration/private key, so compromise retains the
organization installation's combined permission ceiling across its
`repository_selection: all` scope. A dedicated single-repository guard App/key
would provide stronger credential-root isolation, but is optional future
hardening, not required or authorized by this task.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Recovery identity and production-merge
guard visibility are CI orchestration only.

## Risks, dependencies, and evidence

- `VOC-140-R00`: **High release-safety risk** if recovery still reports
  complete without dispatch because required checks show SUCCESS for the
  still-running release carrier, repeating job `98738317266`. Mitigation:
  `VOC-140-D01`, `VOC-140-D03`, `VOC-140-AC-00`, `VOC-140-TEST-00`.
- `VOC-140-R01`: **High attestation-safety risk** if newest-check selection
  continues to choose an in-progress `pipeline.yml` parent run as attestable
  `ci / ci`. Mitigation: `VOC-140-D01`, `VOC-140-AC-00`, `VOC-140-TEST-00`.
- `VOC-140-R02`: **High recovery-safety risk** if any in-progress
  `pipeline.yml` run, including the release carrier, suppresses dedicated
  `recover-promotion-pr-checks`. Mitigation: `VOC-140-D02`, `VOC-140-D03`,
  `VOC-140-TEST-02`.
- `VOC-140-R03`: **High authorization risk** if omitted `bypass_actors` is
  treated as empty, or if the guard is weakened to skip no-bypass proof.
  Mitigation: `VOC-140-D05`, `VOC-140-D07`, `VOC-140-TEST-06`,
  `VOC-140-TEST-07`.
- `VOC-140-R04`: **High authorization risk** if the isolated guard token still
  cannot see full ruleset fields after the mint change, repeating
  `production_merge_guard_missing` for live ruleset `20575146`. Mitigation:
  `VOC-140-D06`, `VOC-140-D07`, `VOC-140-TEST-07`, `VOC-140-TEST-08`.
- `VOC-140-R05`: **High privilege-expansion risk** if Administration is added
  to the mutation token, if the guard token is not current-repository scoped,
  or if it reaches a merge/mutation/status/issue/PR step. Mitigation:
  `VOC-140-D06`, `VOC-140-D08`, `VOC-140-TEST-08`.
- `VOC-140-R06`: **High coverage risk** if tests only duplicate helper
  logic with complete admin fixtures and never exercise the omitted-field
  App-token shape or the circular-CI parent-run fixture. Mitigation:
  `VOC-140-D08`, `VOC-140-AC-05`.
- `VOC-140-R07`: **High correctness risk** if the caller pin stays at
  `59943683…` after infra has changed, or is retargeted without mirroring
  the new merge. Mitigation: `VOC-140-D09`, `VOC-140-TEST-09`.
- `VOC-140-R08`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-140-D14`, `VOC-140-TEST-11`.
- `VOC-140-R09`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), creates a duplicate
  promotion PR or audit, or adds a VOC-097 evidence-carrier task whose
  sha_lineage cannot bind to #1090 recovery. Mitigation: `VOC-140-D10`,
  `VOC-140-TEST-12`.
- `VOC-140-R10`: **Medium documentation risk** if
  current-state docs retain active-A-003/release-disabled claims or describe
  the combined App-token system as mutation-only instead of documenting the
  isolated mutation and guard tokens.
  Mitigation: `VOC-140-D11`, `VOC-140-TEST-10`.
- `VOC-140-R11`: **High shared-credential residual risk** because the same App
  private key can mint tokens up to the installation permission ceiling.
  Mitigation: workflow token-use isolation, short-lived repository-scoped
  guard tokens, explicit evidence, and documenting an optional dedicated guard
  App as future hardening (`VOC-140-D12`).
- `VOC-140-R12`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow reversion
  plus infra revert of the coordinated PR.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, this
  package directory, named current-state docs, reusable recovery/attestation/
  guard sources, and App-token mints on release and production-branch
  merge-gate.
- `VOC-140-DEP-00` through `VOC-140-DEP-10`: see `change.yaml`.
- `VOC-140-EV-00`: T00 evidence — implementation PR base, new infra merge,
  circular-CI identity repair, dedicated promotion-pr-validation requirement,
  token/API contract, omitted-field distinct failure, pin advance, validation
  after commit, exact-head binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI recovery and production-merge-guard visibility, not a
product route or critical API endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/` and mutates required-check recovery and the App
isolated guard token used to prove production protection, but the path classifier and
independent verifier remain authoritative. This is a draft proposal, not a
determination.
