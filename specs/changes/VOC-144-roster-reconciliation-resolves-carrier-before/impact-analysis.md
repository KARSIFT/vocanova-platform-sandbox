# VOC-144 — Impact Analysis

## Security and privacy

This package repairs adoption so documented reconcile does not treat
immediate post-push GitHub REST PR-head lag as a permanent
`MISMATCHED_OPEN_CARRIER`. It does not introduce new secret values,
OAuth/session material, production-data access, or user-facing data flows.
It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.
It does not change the VOC-140 two-token contract.

Security controls that must remain:

- Exact PR base/head SHAs, repository, and refs remain required for roster
  carrier reuse. A still-stale listed head is not reused.
- The single-snapshot identity predicate stays fail-closed. The adapter
  wait only covers unique same-repo/same-ref/same-base SHA lag.
- Ambiguous carriers, repository/base mismatch, and API failure fail closed
  without waiting.
- After the named bound, a still-different SHA remains
  `MISMATCHED_OPEN_CARRIER`. No second roster PR is created.
- VOC-142 complete-required-set roster wait remains: `ci / ci`,
  `governance-policy`, and `validate` must be registered SUCCESS on the
  exact head. IN_PROGRESS required rows are not-ready.
- Merge-gate and release still do not treat an in-progress parent as
  attestable completed `ci / ci`. `statusCheckRollup` and `gh pr checks`
  remain forbidden in this path.
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
user-interface accessibility effect. Roster-carrier SHA-lag wait is CI
orchestration only.

## Risks, dependencies, and evidence

- `VOC-144-R00`: **High recovery-safety risk** if `Open roster PR` still
  resolves from a single post-push REST snapshot, repeating runs
  `33437239322` and `33437514152`. Mitigation: `VOC-144-D01`,
  `VOC-144-TEST-00`.
- `VOC-144-R01`: **High identity-safety risk** if the wait reuses a
  still-stale listed head or drops SHA equality from the predicate.
  Mitigation: `VOC-144-D02`, `VOC-144-D05`, `VOC-144-TEST-00`,
  `VOC-144-TEST-01`.
- `VOC-144-R02`: **High recovery-safety risk** if timeout still-stale is
  treated as create, producing a duplicate roster PR. Mitigation:
  `VOC-144-D06`, `VOC-144-TEST-01`.
- `VOC-144-R03`: **High recovery-safety risk** if ambiguous, wrong-repo, or
  wrong-base carriers enter the SHA-lag wait. Mitigation: `VOC-144-D03`,
  `VOC-144-TEST-02`.
- `VOC-144-R04`: **High availability risk** if the wait is unbounded or
  reuses VOC-141's 1,800-second recovery timeout. Mitigation: `VOC-144-D04`,
  `VOC-144-TEST-01`.
- `VOC-144-R05`: **High coverage risk** if tests only assert
  `MISMATCHED_OPEN_CARRIER` on one stale snapshot and never exercise
  stale-then-converge. Mitigation: `VOC-144-D07`, `VOC-144-TEST-00`.
- `VOC-144-R06`: **High attestation-safety risk** if this repair weakens
  VOC-142 complete-required-set wait, switches to `statusCheckRollup` /
  `gh pr checks`, or treats in-progress parents as attestable SUCCESS.
  Mitigation: `VOC-144-D17`, `VOC-144-TEST-03`, `VOC-144-TEST-07`.
- `VOC-144-R07`: **High correctness risk** if the caller pin stays at
  `8993e867…` after infra has changed, or is retargeted without mirroring
  the new merge. Mitigation: `VOC-144-D08`, `VOC-144-TEST-05`.
- `VOC-144-R08`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-144-D13`, `VOC-144-TEST-07`.
- `VOC-144-R09`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), manually merges #1112,
  creates a duplicate VOC-141 carrier, or adds a VOC-097 evidence-carrier
  task. Mitigation: `VOC-144-D09`, `VOC-144-TEST-08`.
- `VOC-144-R10`: **Medium documentation risk** if `AGENTS.md` still claims
  reconcile reuses an exact head while live resolution fails on the first
  lagged snapshot. Mitigation: `VOC-144-D10`, `VOC-144-TEST-06`.
- `VOC-144-R11`: **Medium sequencing risk** if this package's own first
  native adoption hits the same post-push lag before T00 is live.
  Mitigation: `VOC-144-D16`; do not add a bootstrap exception or
  manual-merge path.
- `VOC-144-R12`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is fixture/test/doc/workflow reversion
  plus infra revert of the coordinated PR.
- Protected surfaces: caller `tooling/governance/` fixtures and tests, this
  package directory, named current-state docs including `AGENTS.md`, and
  reusable `adopt.yml` carrier-resolution sources.
- `VOC-144-DEP-00` through `VOC-144-DEP-10`: see `change.yaml`.
- `VOC-144-EV-00`: T00 evidence — implementation PR base, new infra merge,
  named timeout/poll constants, convergence-wait repair, pin advance,
  validation after commit, exact-head binding contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI adoption recovery, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration
under `tooling/governance/` and mutates exact-SHA carrier identity used
before protected roster merge, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.

## Open questions

Exact timeout and poll-interval seconds within the 60-second ceiling are an
implementer-named choice for adoption-time review if a different bound is
preferred (`VOC-144-D04`). The wait location (adapter versus a thin helper
the adapter owns) is settled as adapter-owned (`VOC-144-D02`); do not move
it into the single-snapshot predicate.
