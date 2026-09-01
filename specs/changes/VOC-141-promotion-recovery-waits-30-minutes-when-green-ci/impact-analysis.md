# VOC-141 — Impact Analysis

## Security and privacy

This package repairs promotion recovery dispatch so a GitHub-required
SUCCESS `ci / ci` row whose composed parent is not uniquely attestable
immediately plans dedicated `promotion-pr-validation` instead of polling
until timeout. It does not introduce new secret values, OAuth/session
material, production-data access, or user-facing data flows. It does not
rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`. It does not change
the VOC-140 two-token contract.

Security controls that must remain:

- An in-progress or failed release carrier is never trusted recovered
  `ci / ci`. Dedicated `promotion-pr-validation PR #<n>` must be
  completed/successful, or a completed non-carrier `pull_request` `ci / ci`
  run may attest only when it is not confused with a newer carrier check.
- GitHub-required SUCCESS is not sufficient to skip dedicated dispatch when
  composed CI is unattestable.
- Exact PR base/head SHAs, repository, refs, and `pr-validation` remain
  required. Promotion PRs are not switched to `--squash-safe-push`. Doomed
  `pull_request` `ci / ci` jobs are not rerun as a strategy.
- The production merge guard still requires an effective active
  repository-owned ruleset, a pull-request rule, strict non-empty required
  checks, and `bypass_actors: []`. The mutation token remains exactly
  Contents/Issues/Pull requests write. The isolated guard token remains
  Administration-write-only, current-repository scoped, and unused for merge
  or mutations.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Recovery dispatch is CI orchestration
only.

## Risks, dependencies, and evidence

- `VOC-141-R00`: **High release-safety risk** if recovery still polls 1,800
  seconds with no dispatch because GitHub-required `ci / ci` is SUCCESS while
  composed evidence is unattestable, repeating job `99334840338`. Mitigation:
  `VOC-141-D01`, `VOC-141-D02`, `VOC-141-AC-00`, `VOC-141-TEST-00`.
- `VOC-141-R01`: **High recovery-safety risk** if
  `apply_promotion_pr_recovery_plan` continues to derive `dispatch_contexts`
  only from `plan_required_check_recovery`. Mitigation: `VOC-141-D02`,
  `VOC-141-TEST-00`, `VOC-141-TEST-01`.
- `VOC-141-R02`: **High recovery-safety risk** if
  `suppress_active_or_successful_dispatches` SUCCESS/PENDING filtering is
  reused unchanged and drops `pipeline.yml` for unattestable CI. Mitigation:
  `VOC-141-D04`, `VOC-141-TEST-01`, `VOC-141-TEST-03`.
- `VOC-141-R03`: **High recovery-safety risk** if a release carrier or
  duplicate native carrier suppresses dedicated dispatch. Mitigation:
  `VOC-141-D04`, `VOC-141-TEST-03`.
- `VOC-141-R04`: **Medium operability risk** if timeout diagnostics still
  report `missing_checks: none` with no unattestable-CI reason. Mitigation:
  `VOC-141-D05`, `VOC-141-TEST-04`.
- `VOC-141-R05`: **High attestation-safety risk** if the repair starts
  trusting in-progress/failed release carriers or rerunning doomed
  `pull_request` `ci / ci` jobs. Mitigation: `VOC-141-D06`, `VOC-141-TEST-05`.
- `VOC-141-R06`: **High authorization risk** if the production merge guard or
  two-token contract is weakened while repairing dispatch. Mitigation:
  `VOC-141-D06`, `VOC-141-DEP-04`, `VOC-141-TEST-05`.
- `VOC-141-R07`: **High coverage risk** if tests only call
  `suppress_active_or_successful_dispatches` without GitHub-required SUCCESS
  rows. Mitigation: `VOC-141-D07`, `VOC-141-TEST-01`.
- `VOC-141-R08`: **High correctness risk** if the caller pin stays at
  `67bdfd13…` after infra has changed, or is retargeted without mirroring
  the new merge. Mitigation: `VOC-141-D08`, `VOC-141-TEST-06`.
- `VOC-141-R09`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-141-D13`, `VOC-141-TEST-08`.
- `VOC-141-R10`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), creates a duplicate
  promotion PR or audit, or adds a VOC-097 evidence-carrier task.
  Mitigation: `VOC-141-D09`, `VOC-141-TEST-09`.
- `VOC-141-R11`: **Medium documentation risk** if current-state docs still
  claim recovery dispatches dedicated validation whenever no completed
  non-carrier run exists, without the SUCCESS-plus-unattestable-parent
  planner rule. Mitigation: `VOC-141-D10`, `VOC-141-TEST-07`.
- `VOC-141-R12`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow reversion
  plus infra revert of the coordinated PR.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, this
  package directory, named current-state docs, and reusable
  recovery/attestation sources.
- `VOC-141-DEP-00` through `VOC-141-DEP-09`: see `change.yaml`.
- `VOC-141-EV-00`: T00 evidence — implementation PR base, new infra merge,
  unattestable-SUCCESS dispatch repair, timeout-diagnostic token, pin
  advance, validation after commit, exact-head binding contract, and release
  handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI recovery dispatch, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/` and mutates required-check recovery used before
production merge, but the path classifier and independent verifier remain
authoritative. This is a draft proposal, not a determination.
