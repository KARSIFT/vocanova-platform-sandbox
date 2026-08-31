# VOC-142 — Impact Analysis

## Security and privacy

This package repairs adoption so a generated roster PR cannot be merged while
ruleset-required `ci / ci` is still unregistered or IN_PROGRESS, and so
documented reconcile reuses a matching open or already-merged roster carrier
instead of failing on `gh pr create`. It does not introduce new secret
values, OAuth/session material, production-data access, or user-facing data
flows. It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.
It does not change the VOC-140 two-token contract.

Security controls that must remain:

- Roster wait requires the complete ruleset-required set, including
  `ci / ci`, to be registered and SUCCESS on the exact head. A partial green
  snapshot is not complete.
- IN_PROGRESS required rows are not-ready. Merge-gate and release still do
  not treat an in-progress parent as attestable completed `ci / ci`.
- Wait continues to use paginated check-runs/statuses and newest-logical-
  attempt selection. `statusCheckRollup` and `gh pr checks` remain forbidden
  in this path.
- Exact PR base/head SHAs, repository, and refs remain required for roster
  carrier reuse. Mismatched or ambiguous carriers fail closed.
- Existing task issues are reused. Root implementation dispatches exactly
  once after the checked roster merges.
- The production merge guard still requires an effective active
  repository-owned ruleset, a pull-request rule, strict non-empty required
  checks, and `bypass_actors: []`.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Adoption wait and PR reuse are CI
orchestration only.

## Risks, dependencies, and evidence

- `VOC-142-R00`: **High release-safety risk** if roster wait still completes
  while required `ci / ci` is IN_PROGRESS or unregistered, repeating job
  `99342230038` then failing protected merge. Mitigation: `VOC-142-D01`,
  `VOC-142-D02`, `VOC-142-AC-00`, `VOC-142-TEST-00`, `VOC-142-TEST-01`.
- `VOC-142-R01`: **High recovery-safety risk** if `_workflow_runs` continues
  to omit IN_PROGRESS `ci / ci` from the wait snapshot so `pending` stays 0.
  Mitigation: `VOC-142-D03`, `VOC-142-TEST-00`.
- `VOC-142-R02`: **High recovery-safety risk** if two stable subset counts
  still count as complete. Mitigation: `VOC-142-D02`, `VOC-142-TEST-01`.
- `VOC-142-R03`: **High recovery-safety risk** if `Open roster PR` still
  always calls `gh pr create` when the exact open carrier exists, repeating
  job `99342577393`. Mitigation: `VOC-142-D04`, `VOC-142-TEST-02`.
- `VOC-142-R04`: **High recovery-safety risk** if a mismatched or ambiguous
  carrier is reused or a second PR is created when an already-merged exact
  carrier exists. Mitigation: `VOC-142-D05`, `VOC-142-TEST-03`,
  `VOC-142-TEST-04`.
- `VOC-142-R05`: **High authority risk** if reconcile opens a second task or
  dispatches the root implementation twice. Mitigation: `VOC-142-D06`,
  `VOC-142-TEST-05`.
- `VOC-142-R06`: **High attestation-safety risk** if the wait repair starts
  treating in-progress parents as attestable SUCCESS for merge-gate or
  release, or switches to `statusCheckRollup` / `gh pr checks`. Mitigation:
  `VOC-142-D03`, `VOC-142-TEST-08`.
- `VOC-142-R07`: **High coverage risk** if tests only assert
  `stable_green_count` exists in YAML without exercising IN_PROGRESS
  `ci / ci` or open-PR reuse. Mitigation: `VOC-142-D07`, `VOC-142-TEST-00`,
  `VOC-142-TEST-02`.
- `VOC-142-R08`: **High correctness risk** if the caller pin stays at
  `67bdfd13…` after infra has changed, or is retargeted without mirroring
  the new merge. Mitigation: `VOC-142-D08`, `VOC-142-TEST-06`.
- `VOC-142-R09`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-142-D13`, `VOC-142-TEST-08`.
- `VOC-142-R10`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), manually merges #1112,
  creates a duplicate VOC-141 carrier, or adds a VOC-097 evidence-carrier
  task. Mitigation: `VOC-142-D09`, `VOC-142-TEST-09`.
- `VOC-142-R11`: **Medium documentation risk** if `AGENTS.md` still claims
  reconcile is idempotent while live `Open roster PR` always creates, or if
  fixture README still treats two stable snapshots as complete. Mitigation:
  `VOC-142-D10`, `VOC-142-TEST-07`.
- `VOC-142-R12`: **Medium sequencing risk** if this package's own first
  native adoption hits the same wait-then-create race before T00 is live.
  Mitigation: `VOC-142-D16`; do not add a bootstrap exception or manual-merge
  path.
- `VOC-142-R13`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow reversion
  plus infra revert of the coordinated PR.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, this
  package directory, named current-state docs including `AGENTS.md`, and
  reusable `adopt.yml` wait/reuse sources.
- `VOC-142-DEP-00` through `VOC-142-DEP-10`: see `change.yaml`.
- `VOC-142-EV-00`: T00 evidence — implementation PR base, new infra merge,
  wait-completeness repair, carrier-reuse repair, pin advance, validation
  after commit, exact-head binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI adoption recovery, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/` and mutates required-check completeness used
before protected roster merge, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
