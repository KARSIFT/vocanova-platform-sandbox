# VOC-122 — Impact Analysis

## Security and privacy

This package repairs governed promotion-check recovery. It does not introduce
new secret values, OAuth/session material, production-data access, or
user-facing data flows.

Security controls that must remain:

- Recovery metadata reads, exact reruns, and allowlisted dispatches use the
  workflow job's short-lived `GITHUB_TOKEN` with Checks read, Commit statuses
  read, Contents/Pull requests read, and Actions write.
- App tokens remain limited to the release and merge mutations that require
  the App identity.
- The model-controlled implementer runner never receives the GitHub App token.
- No statuses are fabricated and no ruleset is bypassed.
- First-attempt, exact-SHA, branch, PR, event, and workflow-path identity
  checks remain on every rerun, including replans.
- Caller `.github/workflows/**` changes remain refused on the unreviewed
  automated caller publication path. Infrastructure workflow files belong to
  the infrastructure repository carrier and are independently reviewed there.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-122-R00`: **High delivery risk** if promotion recovery continues to
  plan mutations only from the initial snapshot. A cancelled required row that
  appears seconds later stays visible and unrecovered until timeout, repeating
  PR #1000 / job `98010275057`. Mitigation: replan during polling
  (`VOC-122-D01`, `VOC-122-AC-00`).
- `VOC-122-R01`: **High mutation risk** if replan redispatches an already
  dispatched absent context or reruns the same run ID / attempt 2. Mitigation:
  invocation-scoped run-ID and absent-context dedupe plus first-attempt
  validation (`VOC-122-D02`, `VOC-122-AC-02`).
- `VOC-122-R02`: **High merge-safety risk** if replan treats a successful
  dispatch or same-named status as satisfying GitHub's selected required row,
  regressing VOC-121. Mitigation: preserve required-PR-view selection
  (`VOC-122-D02`, `VOC-122-AC-03`).
- `VOC-122-R03`: **High fail-open risk** if unsafe later snapshots are polled
  through instead of failing closed. Mitigation: `VOC-122-D03`,
  `VOC-122-AC-04`.
- `VOC-122-R04`: **Medium documentation risk** if README/workflow comments
  continue to describe one-shot pre-loop planning as current-state.
  Mitigation: same-task current-state updates (`VOC-122-D06`).
- `VOC-122-R05`: **Low release risk** because no application runtime
  deployment change is intended; rollback is workflow/config/test reversion.
  Restored recovery may merge an already-reviewed promotion PR through
  existing gates after genuine checks.
- Protected surfaces: `KARSIFT/karsift-ai-infra` recovery/release workflows
  and policy modules, App-token versus job-token split, caller
  `tooling/governance/` fixtures and tests, and this package directory.
- `VOC-122-DEP-00` through `VOC-122-DEP-05`: see `change.yaml`.
- `VOC-122-EV-00`: T00 evidence — replan mechanism, dedupe, validation
  commands, exact infra SHA, and pin applicability.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative.
