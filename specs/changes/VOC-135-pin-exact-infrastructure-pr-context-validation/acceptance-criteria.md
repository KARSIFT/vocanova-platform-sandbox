# VOC-135 — Acceptance Criteria

## VOC-135-AC-00 — Caller fixture and pin equal exact infrastructure #167 merge

- Requirement source: `VOC-135-D02`, `VOC-135-D11`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-00`, `VOC-135-TEST-01`
- Evidence: `VOC-135-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals
`b263c0c110591cc798b89277dfc35542abb1597b`. Every live caller pin assertion
equals that same SHA. The pin does not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`, does not equal
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, and does not equal
`f3d79177bf8a9abe0dae550f39502165d494c576`. Fixture
`.github/workflows/ci.yml`, `.github/workflows/implement.yml`,
`.github/workflows/release.yml`, `config/run-app-checks.sh`,
`config/implementer_nested_checkout.py`, `tests/test_app_check_context.py`,
`tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`,
`tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md` match the
recorded SHA-256 hashes of those paths at infra merge `b263c0c…`. Test-time
comparison does not depend on a machine-specific `/tmp` checkout.

## VOC-135-AC-01 — Identify restores shared policy after caller checkout before validate-task

- Requirement source: `VOC-135-D03`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-02`, `VOC-135-TEST-04`
- Evidence: `VOC-135-EV-00`
- Result: pending

Fixture `identify` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-task`. The restore uses
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`. The #164 checkout-ref resolver still runs after
the first shared-policy checkout and before the caller checkout.

## VOC-135-AC-02 — Converge restores shared policy after caller checkout before validate-roster

- Requirement source: `VOC-135-D03`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-03`, `VOC-135-TEST-04`
- Evidence: `VOC-135-EV-00`
- Result: pending

Fixture `converge` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-roster`. The restore uses the same
immutable reusable-workflow revision as `identify`, with credentials not
persisted.

## VOC-135-AC-03 — #166 post-implementer nested-checkout contract is present

- Requirement source: `VOC-135-D05`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-06`
- Evidence: `VOC-135-EV-00`
- Result: pending

Fixture `implement.yml` copies lifecycle helpers to immutable `/tmp` paths
before the unrestricted model. After the model, nested-checkout classification
uses the preserved helper. An absent nested checkout means no infrastructure
source carrier while caller changes continue. A plain subdirectory inheriting
caller Git, a non-directory, or a symlink fails closed. A distinct nested Git
checkout preserves exact-head bundle, ancestry, remote, and lease
publication.

## VOC-135-AC-04 — #164 contracts remain after the #167 pin

- Requirement source: `VOC-135-D04`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-05`
- Evidence: `VOC-135-EV-00`
- Result: pending

The pinned fixture still selects existing `develop` without reading `main`,
falls back to live `main` when `develop` is absent, fails closed on
ambiguous or malformed refs, binds and advances `develop` to
`mergeCommit.oid` before audit close, and does not restore a missing
integration ref with `CHECKED_HEAD_SHA`. Unique develop commits remain
fail-closed. Live `reconcile-production-change` remains the exceptional
main-only identity under the 25-input limit.

## VOC-135-AC-05 — #167 immutable PR-context validation is present and does not fetch evidence

- Requirement source: `VOC-135-D06`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-07`
- Evidence: `VOC-135-EV-00`
- Result: pending

Fixture `run-app-checks.sh` accepts and validates an immutable PR base/head
pair, selects `pr-validation` when the capture fixture is unchanged,
selects `pr-ancestry` when that fixture is added, modified, or deleted, and
fails closed on comparison errors. It does not `git fetch` evidence commits.
Fixture `ci.yml` checks out full reachable history (`fetch-depth: 0`) and
passes the event base/head SHAs. Implementation pre-push and
post-self-correction use the integration anchor plus current committed HEAD.
The VOC-134 class — full PR-style checks under default `local` provenance on
a checkout missing `f9d11e23…` — cannot recur unnoticed in the pinned
fixture.

## VOC-135-AC-06 — Existing controls, roles, and docs remain

- Requirement source: `VOC-135-D08`, `VOC-135-D09`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-08`, `VOC-135-TEST-09`
- Evidence: `VOC-135-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, sanitized raw-error controls,
and rollback controls remain. `config/roles.yml` is unchanged:
implementer and escalation `cursor/composer-2.5`; planner, reviewer,
reviewer_fast_retry, and plan_reviewer
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route is added.
Current-state fixture README/comments name pin `b263c0c…`, the
post-caller-checkout restore, the post-implementer helper-lifetime contract,
and the immutable PR-context contract. Historical A-003 / VOC-075 / VOC-127 /
VOC-129 / VOC-130 / VOC-131 / VOC-132 / VOC-133 / VOC-134 records are not
rewritten. `AGENTS.md` and the navigator skill are not edited for this pin.
The caller-owned fixture README is not replaced by the infrastructure
repository README.

## VOC-135-AC-07 — Eight-path no-change boundary remains the immutable carrier-base bytes

- Requirement source: `VOC-135-D14`, `VOC-135-D15`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-10`
- Evidence: `VOC-135-EV-00`
- Result: pending

All eight named paths are byte-for-byte identical to the immutable
carrier-base SHA selected before implementation (expected
`b9e74fc2db4691c48c637639b265d527de9f4505`, revalidated at dispatch; fail
closed if `develop` moved) and are absent from the implementation PR diff
against that SHA:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `scripts/foundation/voc112-navigation-benchmark-run.mjs`
- `scripts/foundation/validate-workspace.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`
- `package.json`

JSON `subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. The provenance test still
fail-closes `local` mode when a full checkout is missing the captured commit.
The regression requires the exact commit object and all eight paths to
resolve and fails—not skips—if proof cannot run. After the SHA is selected,
a moving branch ref is not the authority.

## VOC-135-AC-08 — No caller-side hydration, fetch, provenance override, or local fail-closed bypass

- Requirement source: `VOC-135-D14`, `VOC-135-D16`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-11`, `VOC-135-TEST-12`
- Evidence: `VOC-135-EV-00`
- Result: pending

No capture-commit fetch helper, hydrate/materialize helper, package/test
wrapper, import side effect, provenance-mode override, environment override,
evidence mutation/stamping helper, skip, or equivalent is added under any
filename. Default `local` fail-closed behavior remains reachable. Regression
coverage inspects the complete caller diff / changed executable paths, not
only one prohibited filename, and fails if any new or changed caller script
invokes Git fetch for capture subjects, sets VOC-112 provenance variables
around tests, hydrates evidence objects, or makes local fail-closed
unreachable. The VOC-134 attempt-1 import-time fetch class and attempt-2
`hydrate-voc112-git-objects.mjs` / `validate-workspace.mjs` class cannot
recur unnoticed.

## VOC-135-AC-09 — Exact-revision evidence is fail-closed and Git-feasible

- Requirement source: `VOC-135-D12`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-13`
- Evidence: `VOC-135-EV-00`
- Result: pending

Committed package evidence records the immutable carrier base, the exact
infra merge, confirmed hashes, validation, and the contract that the final
implementation head is bound by the App-authored independent-review
comment/check. The final review comment must bind the live PR head exactly.
Merge-gate must reject any mismatch. Post-merge/root-issue audit may record
the reviewed head and merge SHA. A commit is not required to contain its own
SHA. Evidence does not claim a protected-path revert unless that path is
absent from the diff. Evidence is not rewritten at test or runtime.

## VOC-135-AC-10 — Replacement carrier is VOC-135; no snapshot-gap; VOC-127 through VOC-134 are not re-implemented

- Requirement source: `VOC-135-D01`, `VOC-135-D10`, `VOC-135-D13`
- Tasks: `VOC-135-T00`
- Tests: `VOC-135-TEST-14`
- Evidence: `VOC-135-EV-00`
- Result: pending

This package's implementation PR is a new VOC-135 carrier from current
`develop`. It does not reuse, merge, cherry-pick, or modify PR #1051,
PR #1056, PR #1065, or PR #1070. It does not redispatch VOC-132-T00 (#1059),
VOC-133-T00 (#1063), or VOC-134-T00 (#1068). It `Closes` only its own VOC-135
task issue. VOC-129 PR #1046 is not re-implemented. No VOC-127, VOC-129,
VOC-130, VOC-131, VOC-132, VOC-133, or VOC-134 completion marker is
manufactured. No snapshot of the current develop/main gap is committed.
After exact-SHA review and merge, ordinary release evaluation (or
`reconcile-release`) completes develop-to-main promotion, exact develop
synchronization without a staging redeploy for tree-equivalent sync,
production deployment where selected, and audit reconciliation for VOC-127,
VOC-130, VOC-131, VOC-132, VOC-133, VOC-134, and this replacement, then
closes root issue #1071. Closed state alone is not completion proof.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
