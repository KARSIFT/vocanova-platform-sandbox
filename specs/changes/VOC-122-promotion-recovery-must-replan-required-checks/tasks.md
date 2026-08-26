# VOC-122 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #1001 is one
promotion-recovery reliability outcome. Coordinated caller and infrastructure
pull requests remain one task; repository count, runner-versus-tests-versus-docs,
and fixture/pin work are not split reasons.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for promotion-recovery
replanning during polling. The implementer opens the infra PR for that
behavior; this package is the authorizing change package for the required
outcome. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. Caller fixture/pin, tests, and evidence
land in this repository under the same task. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a
closing keyword.

## VOC-122-T00 — Replan newly appearing required-check rows during promotion-recovery polling

- Requirement source: issue #1001; `VOC-122-D00` through `VOC-122-D08`
- Acceptance criteria: `VOC-122-AC-00` through `VOC-122-AC-07`
- Tests: `VOC-122-TEST-00` through `VOC-122-TEST-07`
- Evidence: `VOC-122-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1001 live failure in `t00-evidence.md` (promotion PR
   #1000, release run `32912818046` / job `98010275057`, absent-context
   dispatch `32912851134`, cancelled pull-request run `32912850066`, later
   successful convergence run `32912851044` / job `98010820950`).
2. In `KARSIFT/karsift-ai-infra/config/actions-check-recovery-runner.py`, make
   promotion recovery re-evaluate GitHub's required pull-request view during
   polling, not only from `initial_gate_summary` / `pr_required_checks` before
   the `while` loop. A context that is absent in snapshot 1 and later appears
   as a failed/cancelled selected pull-request row must be rerun in the same
   invocation after identity checks.
3. Preserve VOC-121 fail-closed contracts on every plan application:
   validate open promotion PR, exact head SHA, branch, workflow path,
   `pull_request` event, first attempt, and required context before every
   rerun; rerun a selected failed/cancelled exact pull-request run at most
   once per invocation (deduplicate run IDs); dispatch a genuinely absent
   context at most once; do not let an alternate run or same-named status
   override GitHub's selected required row; do not duplicate active or
   successful work; fail closed immediately on ambiguous, foreign, or
   mismatched rows.
4. Keep `DEFAULT_TIMEOUT_SECONDS = 1800`, `POLL_INTERVAL_SECONDS = 30`,
   job-token recovery reads/reruns/dispatches, App identity for merge,
   exact-head merge decision, and existing retry limits. Do not reset the
   deadline when a new row appears. Do not change `integration_push`
   semantics except to prove equivalence if a shared helper is extracted.
5. Add deterministic time-evolving tests that reproduce:
   - absent snapshot 1 → cancelled selected row in snapshot 2 → exactly one
     validated rerun → success;
   - repeated snapshots that do not rerun or redispatch;
   - earlier absent-context dispatch plus later cancelled selected pull-request
     run ID (the #1000 class);
   - ambiguous, foreign, and mismatched rows failing closed on the replan path.
6. Update current-state comments/docs (`actions-check-recovery-runner.py`,
   `recover-actions-checks.yml`, `release.yml` recovery comments as needed,
   `karsift-ai-infra/README.md`) so they no longer describe one-shot pre-loop
   planning as the live contract.
7. Land the infra change through one reviewed infra PR using the live
   VOC-121 coordinated-carrier publisher. Pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA when
   the fixture consumes the change. Update caller fixture regressions and any
   `scripts/foundation/*` pin literals that still assert the previous SHA, in
   the same task.
8. Run applicable validation and record results in `t00-evidence.md`:
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - targeted foundation pin tests if those files change;
   - `git diff --check`;
   - exact reviewed infra SHA and pin applicability;
   - any narrower targeted commands added by the implementation.
9. Preserve independent exact-SHA review for each carrier, risk
   classification, protected checks, and App-token isolation. Do not
   fabricate statuses or bypass rulesets.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Reopening VOC-121 selected-run semantics except as they interact with replan.
- Weakening exact-SHA review, risk floors, protected checks, timeout, or retry
  caps.
- Splitting runner logic, tests, docs, infrastructure, caller pin, or evidence
  into separate tasks.
- Operator-owned live evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. The #1000 incident is already recorded; the next live
  promotion is not an implementer-dispatched VOC-097 evidence obligation.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the live defect, the fail-closed contracts, the tests, the
  docs, and the caller pin are one reliability outcome.
- Infra should merge first when the caller fixture/pin consumes that change;
  otherwise the two reviewed PRs may complete under the same task without a
  pin bump.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
