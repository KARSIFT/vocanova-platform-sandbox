# VOC-122 — Promotion recovery must replan required checks that appear after its initial snapshot: Specification

## Objective and requirement source

Make live promotion recovery recover a required pull-request check that appears
after the invocation's first metadata snapshot, without weakening VOC-121's
fail-closed selected-run, dispatch, timeout, or App-token contracts.

**Requirement source:** [GitHub issue #1001](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1001).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1001)

| Item | Value |
|------|-------|
| Promotion PR | [#1000](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1000) |
| `develop` head that opened it | `bb531fc211e91e40cec1847e1e28d98beaedacff` (VOC-121 caller merge) |
| First recovery job | release run `32912818046`, job `98010275057`, `2026-08-25T23:54:50Z` |
| Snapshot-1 action | absent-context `governance-policy` dispatch run `32912851134` (passed) |
| Snapshot-2 selected row | pull-request `governance-policy` run `32912850066` at `2026-08-25T23:54:52Z`, `CANCELLED` by concurrency |
| Required view | `gh pr checks 1000 --required` continued to report that row `CANCELLED` |
| Defect locus | `KARSIFT/karsift-ai-infra/config/actions-check-recovery-runner.py` `main()`: planning and mutation before the `while` loop; loop only calls `run_metadata_phase(...)` and `recovery_complete(...)` |
| Working later invocation | convergence run `32912851044`, job `98010820950`, reran `32912850066` as attempt 2 |
| Merge | App merged #1000 at `70db9d7dbc4789264732e4fe4f347914dd6f764e` with no ruleset bypass |

## Scope and non-goals

### In scope

1. Re-evaluate GitHub's required pull-request view during promotion-recovery
   polling, not only from the initial snapshot.
2. When a required context was absent at snapshot 1 and later appears as a
   failed or cancelled exact pull-request Actions row, rerun that selected run
   in the same recovery invocation after the existing identity checks.
3. Keep planning and mutation fail-closed:
   - validate the open promotion PR, exact head SHA, branch, workflow path,
     event, first-attempt identity, and required context before every rerun;
   - rerun a selected failed/cancelled exact pull-request run at most once per
     recovery invocation (deduplicate run IDs);
   - dispatch a genuinely absent context at most once;
   - do not let an alternate run or same-named status override GitHub's
     selected required row;
   - do not duplicate active or successful work (pending rows wait; successful
     rows are left alone).
4. Preserve the 1800-second timeout, 30-second poll interval, App-token
   separation (job token for recovery reads/reruns/dispatches; App identity for
   merge), exact-head merge decision after recovery, and existing retry limits.
5. Add deterministic time-evolving tests for the #1000 class and the named
   fail-closed cases.
6. Update current-state recovery comments/docs so they no longer describe
   one-shot pre-loop planning as the live contract.
7. Land the infrastructure change through one reviewed infra PR, then pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA when
   the fixture consumes the change, and update caller fixture/pin tests in the
   same task.

### Non-goals / explicitly excluded

- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Reopening VOC-121 selected-run semantics: alternate successful
  workflow-dispatch runs and same-named commit statuses still must not
  override GitHub's required PR view.
- Fabricating statuses, bypassing rulesets, or treating closed issue state as
  completion proof.
- Changing `integration_push` recovery behavior except where a shared helper
  is extracted and proven equivalent.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, the 1800-second bound, or first-attempt rerun identity checks.
- Splitting runner logic, tests, docs, infrastructure, caller pin, or evidence
  into separate tasks.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: reusable CI/CD recovery and release workflows, shared
  recovery policy modules that mutate required Actions runs, and caller
  `tooling/governance/` fixtures and tests.
- Protected technical effect: when and whether an open develop→main promotion
  PR's required checks are rerun or dispatched. No application runtime effect
  is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-122-D00`: This is one outcome-sized reliability change. Use one
end-to-end implementation task covering infrastructure source, tests,
current-state docs/comments, caller fixture/pin, and evidence. Coordinated
pull requests in `KARSIFT/karsift-ai-infra` and this caller remain one task.
Repository count, file count, and runner-versus-tests-versus-docs are not
split reasons.

`VOC-122-D01`: Promotion recovery must apply required-check planning to the
**current** GitHub required PR view at start and again during polling, until
`recovery_complete` or the existing 1800-second deadline. A context that is
absent in snapshot 1 and appears as a failed/cancelled selected pull-request
row in a later snapshot must become a validated exact-run rerun in that same
invocation. Control-flow shape (helper versus inline; plan-before-loop plus
plan-in-loop versus plan-only-in-loop with an immediate first iteration) is an
implementation choice; the observable contract is not.

`VOC-122-D02`: Preserve VOC-121 fail-closed mutation contracts on every plan
application, including replans:

- `validate_selected_workflow_run` (or a strict equivalent) still binds open
  promotion PR number, exact head SHA, branch, workflow path, `pull_request`
  event, first attempt (`run_attempt == 1`), and required context before every
  rerun.
- Deduplicate reruns by run ID for the whole invocation; a selected
  failed/cancelled exact pull-request run is rerun at most once.
- Deduplicate absent-context workflow dispatches by required context for the
  whole invocation; a genuinely absent context is dispatched at most once.
- After an absent-context dispatch, a later-appearing cancelled/failed
  selected pull-request row for that same context is still rerun (different
  run ID). The earlier dispatch success must not suppress that selected row.
- Pending rows wait. Successful selected rows are not rerun or redispatched.
- Alternate runs and same-named statuses cannot override GitHub's selected
  required row.

`VOC-122-D03`: Ambiguous selected-run identity (more than one candidate run
ID for a failed required context), foreign workflow or non-`pull_request`
event, unreadable required-view payload, and SHA/branch/PR/path/attempt
mismatch continue to fail closed immediately. Replan must not swallow those
errors and keep polling.

`VOC-122-D04`: Keep `DEFAULT_TIMEOUT_SECONDS = 1800`,
`POLL_INTERVAL_SECONDS = 30`, job-token recovery reads/reruns/dispatches, App
identity for promotion merge, exact-head merge-decision identity checks in
`release.yml`, and existing implementer two-attempt bounds. Do not reset the
deadline when a new row appears.

`VOC-122-D05`: Deterministic tests must reproduce a time-evolving required
view:

1. snapshot 1: required context absent → at most one allowlisted dispatch and
   no rerun of a not-yet-selected run;
2. snapshot 2: that context appears as cancelled/failed exact pull-request
   run → exactly one validated rerun of that run ID;
3. later snapshots: no second rerun of that run ID and no second dispatch of
   that context; after the rerun succeeds, recovery completes;
4. repeated identical unsatisfied snapshots stay idempotent;
5. ambiguous, foreign, and mismatched rows fail closed on the replan path,
   not only on the initial snapshot.

Positive cases must prove the corrected behavior. Tests must not use secrets
or production data.

`VOC-122-D06`: Current-state comments in `actions-check-recovery-runner.py`,
`recover-actions-checks.yml`, `release.yml` where they describe recovery
planning, and `karsift-ai-infra/README.md` must stop implying that promotion
recovery plans mutations only once before polling. After the authoritative
infrastructure merge SHA is known, pin
`tooling/governance/fixtures/karsift-ai-infra/` when the fixture consumes the
change, or record explicit non-consumption. Advance matching caller pin
assertions, including `tooling/governance/tests/test_voc121_implement_policy.py`
and the `scripts/foundation/voc097-fixture-matrix.test.mjs`,
`voc104-ready-for-review-reuse.test.mjs`, and
`voc108-authoritative-lifecycle.test.mjs` pin literals when those tests still
assert the previous infra merge.

`VOC-122-D07`: Do not change `integration_push` recovery semantics. If
planning is extracted into a shared helper, prove `integration_push` remains
behavior-equivalent (still plans allowlisted missing push workflows from the
initial snapshot and does not newly rerun promotion-PR check-runs).

`VOC-122-D08`: VOC-121's isolated source publisher is already live. This task
must use that coordinated-carrier path. Silent discard of nested
infrastructure edits remains forbidden. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and must not use a
GitHub closing keyword. The caller implementation PR keeps local `Closes #N`.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. Recovery continues to
use the workflow job's short-lived `GITHUB_TOKEN` for metadata reads, exact
reruns, and allowlisted dispatches. App tokens remain limited to the release
and merge mutations that require the App identity. The model-controlled
implementer runner still never receives the GitHub App token.

Abuse/process risks:

1. Polling until timeout while a recoverable cancelled row is visible —
   mitigated by `VOC-122-D01` and time-evolving tests.
2. Duplicate dispatch or rerun of attempt 2 / in-flight work — mitigated by
   invocation-scoped run-ID and absent-context dedupe plus first-attempt
   validation (`VOC-122-D02`).
3. Treating a successful dispatch or same-named status as the required row —
   mitigated by preserving VOC-121 selection (`VOC-122-D02`) and
   `VOC-122-D03`.
4. Broadening recovery credentials — out of scope and forbidden.

## Contradictions and open questions

1. **Helper versus inline replan (`VOC-122-DEP-05`):** the required behavior is
   settled; the exact function extraction is not. T00 may keep planning in
   `main()` or extract a helper, as long as tests pin the live #1000 class.
2. **Whether `plan_required_check_recovery` itself must change:** the current
   pure function already classifies absent versus cancelled for a single
   snapshot. The live defect is that `main()` does not call it again. T00 may
   leave that function unchanged if runner-level replan plus tests are
   sufficient, or tighten it only if needed for invocation-scoped dedupe.
3. **Fixture pin applicability:** pin
   `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` to the exact
   reviewed infra merge when the mirrored fixture consumes the changed
   runner, tests, or recovery comments. Workflow and recovery-module changes
   in this package are expected to be consumed. If some files are not in the
   policy fixture subset, do not copy them merely to force a pin; record
   non-consumption.
