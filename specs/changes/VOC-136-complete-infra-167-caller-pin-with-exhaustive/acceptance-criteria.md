# VOC-136 — Acceptance Criteria

## VOC-136-AC-00 — Caller fixture and pin equal exact infrastructure #167 merge

- Requirement source: `VOC-136-D02`, `VOC-136-D11`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-00`, `VOC-136-TEST-01`
- Evidence: `VOC-136-EV-00`
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

## VOC-136-AC-01 — Identify restores shared policy after caller checkout before validate-task

- Requirement source: `VOC-136-D03`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-02`, `VOC-136-TEST-04`
- Evidence: `VOC-136-EV-00`
- Result: pending

Fixture `identify` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-task`. The restore uses
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`. The #164 checkout-ref resolver still runs after
the first shared-policy checkout and before the caller checkout.

## VOC-136-AC-02 — Converge restores shared policy after caller checkout before validate-roster

- Requirement source: `VOC-136-D03`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-03`, `VOC-136-TEST-04`
- Evidence: `VOC-136-EV-00`
- Result: pending

Fixture `converge` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-roster`. The restore uses the same
immutable reusable-workflow revision as `identify`, with credentials not
persisted.

## VOC-136-AC-03 — #166 post-implementer nested-checkout contract is present

- Requirement source: `VOC-136-D05`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-06`
- Evidence: `VOC-136-EV-00`
- Result: pending

Fixture `implement.yml` copies lifecycle helpers to immutable `/tmp` paths
before the unrestricted model. After the model, nested-checkout classification
uses the preserved helper. An absent nested checkout means no infrastructure
source carrier while caller changes continue. A plain subdirectory inheriting
caller Git, a non-directory, or a symlink fails closed. A distinct nested Git
checkout preserves exact-head bundle, ancestry, remote, and lease
publication.

## VOC-136-AC-04 — #164 contracts remain after the #167 pin

- Requirement source: `VOC-136-D04`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-05`
- Evidence: `VOC-136-EV-00`
- Result: pending

The pinned fixture still selects existing `develop` without reading `main`,
falls back to live `main` when `develop` is absent, fails closed on
ambiguous or malformed refs, binds and advances `develop` to
`mergeCommit.oid` before audit close, and does not restore a missing
integration ref with `CHECKED_HEAD_SHA`. Unique develop commits remain
fail-closed. Live `reconcile-production-change` remains the exceptional
main-only identity under the 25-input limit.

## VOC-136-AC-05 — #167 immutable PR-context validation is present and does not fetch evidence

- Requirement source: `VOC-136-D06`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-07`
- Evidence: `VOC-136-EV-00`
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

## VOC-136-AC-06 — Existing controls, roles, and docs remain

- Requirement source: `VOC-136-D08`, `VOC-136-D09`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-08`, `VOC-136-TEST-09`
- Evidence: `VOC-136-EV-00`
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
VOC-129 / VOC-130 / VOC-131 / VOC-132 / VOC-133 / VOC-134 / VOC-135 records
are not rewritten. `AGENTS.md` and the navigator skill are not edited for
this pin. The caller-owned fixture README is not replaced by the
infrastructure repository README.

## VOC-136-AC-07 — Eight-path no-change boundary remains the protected comparison-anchor bytes

- Requirement source: `VOC-136-D14`, `VOC-136-D15`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-10`, `VOC-136-TEST-17`
- Evidence: `VOC-136-EV-00`
- Result: pending

All eight named paths are byte-for-byte identical to protected comparison
anchor `b9e74fc2db4691c48c637639b265d527de9f4505` and are absent from the
implementation PR diff against that SHA:

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
The regression requires the exact protected comparison-anchor commit object
and all eight paths to resolve and fails—not skips—if proof cannot run. After
the implementation PR base is recorded, a moving branch ref is not the
eight-path comparison authority. Unrelated/material movement of `develop`
fails closed. This package's own plan/adoption/roster commits after
`b9e74fc2…` do not count as protected-file drift.

## VOC-136-AC-08 — Exhaustive scan rejects hydration, fetch, provenance override, or local fail-closed bypass anywhere including caller tests

- Requirement source: `VOC-136-D14`, `VOC-136-D16`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-11`, `VOC-136-TEST-12`, `VOC-136-TEST-15`, `VOC-136-TEST-16`
- Evidence: `VOC-136-EV-00`
- Result: pending

No capture-commit fetch helper, hydrate/materialize helper, package/test
wrapper, import side effect, provenance-mode override, environment override,
evidence mutation/stamping helper, skip, or equivalent is added under any
filename. Default `local` fail-closed behavior remains reachable. Regression
coverage enumerates every added or modified path against the protected
comparison anchor and scans all `scripts/**`, `package.json`, every
added/changed `*.mjs` / `*.js` / `*.sh` / `*.py` anywhere except the
mirrored `tooling/governance/fixtures/karsift-ai-infra/**` subtree, and any
other newly executable caller file. It does **not** exclude
`tooling/governance/tests/**`, this regression's own module, or another
executable directory wholesale. The VOC-134 attempt-1 import-time fetch
class, the VOC-134 attempt-2 `hydrate-voc112-git-objects.mjs` /
`validate-workspace.mjs` class, the VOC-135 attempt-1 self-scan class, and
the VOC-135 attempt-2 `SCAN_EXCLUDE_PREFIXES` class cannot recur unnoticed.

## VOC-136-AC-09 — Exact-revision evidence is fail-closed and Git-feasible

- Requirement source: `VOC-136-D12`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-13`
- Evidence: `VOC-136-EV-00`
- Result: pending

Committed package evidence records the protected comparison anchor, the
actual implementation PR base, the exact infra merge, confirmed hashes,
complete-diff scan scope and negative cases, validation, and the contract
that the final implementation head is bound by the App-authored
independent-review comment/check. The final review comment must bind the live
PR head exactly. Merge-gate must reject any mismatch. Post-merge/root-issue
audit may record the reviewed head and merge SHA. A commit is not required to
contain its own SHA. Evidence does not claim a protected-path revert unless
that path is absent from the diff against the protected comparison anchor.
Evidence is not rewritten at test or runtime.

## VOC-136-AC-10 — Replacement carrier is VOC-136; no snapshot-gap; VOC-127 through VOC-135 are not re-implemented

- Requirement source: `VOC-136-D01`, `VOC-136-D10`, `VOC-136-D13`, `VOC-136-D17`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-14`
- Evidence: `VOC-136-EV-00`
- Result: pending

This package's implementation PR is a new VOC-136 carrier from current
`develop`. It does not reuse, merge, cherry-pick, or modify PR #1051,
PR #1056, PR #1065, PR #1070, or PR #1075. It does not redispatch
VOC-132-T00 (#1059), VOC-133-T00 (#1063), VOC-134-T00 (#1068), or
VOC-135-T00 (#1073). It `Closes` only its own VOC-136 task issue. VOC-129
PR #1046 is not re-implemented. No VOC-127, VOC-129, VOC-130, VOC-131,
VOC-132, VOC-133, VOC-134, or VOC-135 completion marker is manufactured. No
snapshot of the current develop/main gap is committed. After exact-SHA
review and merge, ordinary release evaluation (or `reconcile-release`)
completes develop-to-main promotion, exact develop synchronization without a
staging redeploy for tree-equivalent sync (`develop` 0 ahead / 0 behind
`main`, identical tree), production deployment where selected, and audit
reconciliation for VOC-127, VOC-130, VOC-131, VOC-132, VOC-133, VOC-134,
VOC-135, and this replacement, then closes root issue #1076. Closed state
alone is not completion proof. Cleanup preserves unrelated VOC-128 and all
user worktrees.

## VOC-136-AC-11 — Source-safe literals, required negative cases, and tracked-after-commit pass

- Requirement source: `VOC-136-D16`
- Tasks: `VOC-136-T00`
- Tests: `VOC-136-TEST-15`, `VOC-136-TEST-16`
- Evidence: `VOC-136-EV-00`
- Result: pending

Test and assertion literals that would otherwise match the scanner are
represented as non-contiguous source-safe values so self-scanning the tracked
committed regression module does not false-positive. Actual semantic setters
and commands fail. Deterministic negative unit cases prove the scanner
rejects at least: an import-time or pre-test `git fetch` for VOC-112/
evidence; a hydrate/materialize helper under an arbitrary new filename;
shell, Node, and Python forms that set `VOC112_CAPTURE_PROVENANCE_MODE`;
relevant `PR_BASE_SHA` / `PR_HEAD_SHA` setters around tests or
`validate-workspace`; an import-time side effect inside an added
`tooling/governance/tests/*.py` module; and a wrapper or skip that makes
default local fail-closed unreachable. Benign assertion/test-data mentions
do not false-positive. The regression itself passes after it is tracked and
committed, not merely while untracked. PR #1075 attempt-2 PASS WITH
NON-BLOCKING FINDINGS is not treated as sufficient if the
`tooling/governance/tests/` exclusion remains.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
