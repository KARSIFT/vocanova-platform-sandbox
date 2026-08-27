# VOC-135 — Pin exact infrastructure PR-context validation and remove all VOC-112 hydration bypasses: Specification

## Objective and requirement source

Deliver the caller-side pin, fixture, tests, docs, complete no-hydration
boundary, and feasible exact-revision evidence that consume the exact
authoritative infrastructure merge
`b263c0c110591cc798b89277dfc35542abb1597b` (KARSIFT/karsift-ai-infra#167)
from a fresh branch based on current `develop`. The final caller must apply
an immutable PR base/head pair to application checks, restore shared
lifecycle policy after the root caller checkout in both `identify` and
`converge` before task-completion helpers run, preserve #166 post-implementer
nested-checkout classification and helper-lifetime behavior, keep the eight
named no-change paths byte-identical to an immutable carrier-base SHA
selected before implementation, pass independent exact-revision review and
all protected checks without self-referential commit SHAs or provenance
hydration bypasses, and then complete ordinary develop-to-main promotion for
this package and cleanup of superseded VOC-127 / VOC-130 through VOC-134
carriers.

**Requirement source:** [GitHub issue #1071](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1071).
VOC-129 (`specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the`)
already merged caller PR [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
at `429d8c6d49303148ca1cc14dba5f6768a7863346`. VOC-130 through VOC-134
adopted overlapping #165/#166 pin outcomes; VOC-134-T00 (#1068 / PR #1070)
exhausted both attempts on import-time capture fetch and a relocated
hydrate helper. This package is the governed replacement carrier. It is not
a second VOC-129 implementation, not a retry of VOC-130 through VOC-134, and
not a silent retarget of VOC-134's exhausted #166 pin.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1071)

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| VOC-132 task that must not be redispatched | [#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059) (`VOC-132-T00`) |
| Exhausted VOC-133 task | [#1063](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1063) (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | [#1065](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065) |
| Adopted VOC-134 plan | [#1067](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1067) |
| Exhausted VOC-134 task | [#1068](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1068) (`VOC-134-T00`) |
| Exhausted VOC-134 carrier | [#1070](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1070) |
| VOC-134 attempt-1 FAIL head | `f373d9a61f90659e982173395a96ff54776fc945` |
| VOC-134 attempt-1 defect | import-time capture fetch in `scripts/foundation/voc112-navigation-benchmark-run.mjs` |
| VOC-134 attempt-2 FAIL head | `718bfdcdaec65e5df6e918fbc25f0fe64f6a2cad` |
| VOC-134 attempt-2 defect | `scripts/foundation/hydrate-voc112-git-objects.mjs` invoked from `scripts/foundation/validate-workspace.mjs` |
| Truthful pre-push failure | full checkout missing subject `f9d11e232a07c7d7a9c433d02c9267912543ba10` → `a full local checkout must already contain the captured commit` |
| VOC-129 caller PR | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Authoritative infra merge | `b263c0c110591cc798b89277dfc35542abb1597b` (#167) |
| Independently reviewed PASS infra head | `eb11c4fc6841ec73816e2e064dcd449d98c1e933` |
| Infra suite at that head | 442/442; hosted actionlint, policy-tests, shellcheck, YAML parse passed |
| Current `develop` pin | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| VOC-134 unmerged pin target | `f3d79177bf8a9abe0dae550f39502165d494c576` (#166) |
| Expected immutable carrier base | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Why VOC-134 cannot retry on #1068 / #1070 | both allowed attempts exhausted; hydration bypass relocated from the runner to `validate-workspace.mjs`; old pre-push had no PR context |
| Why VOC-133 cannot retry on #1063 / #1065 | both allowed attempts exhausted; provenance bypass plus Git-impossible self-SHA evidence |
| Why VOC-132 cannot be retargeted or redispatched | VOC-132's adopted pin is exact #165; run `33079499176` created no publishable caller carrier |
| Why VOC-130 / VOC-131 cannot retry | both allowed attempts exhausted on VOC-112 JSON retargets / provenance-test weakening |
| Why VOC-129 is not retried | #1046 already merged; this is a pin plus PR-context / nested-checkout / restore consumption plus feasible evidence |

## Scope and non-goals

### In scope

1. Set `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` and every
   live pin assertion (caller governance tests and matching
   `scripts/foundation/*` pin literals) to
   `b263c0c110591cc798b89277dfc35542abb1597b`.
2. Mirror from exact infra merge #167 every necessary changed authoritative
   #164→#167 caller fixture file, byte-for-byte. Required #167 files plus
   every #166 file still required by the release synchronization outcome,
   matching issue #1071:
   - `.github/workflows/ci.yml`
   - `.github/workflows/implement.yml`
   - `config/run-app-checks.sh`
   - `tests/test_app_check_context.py` (new in the caller fixture)
   - `.github/workflows/release.yml`
   - `config/implementer_nested_checkout.py` (new in the caller fixture)
   - `tests/test_release_policy.py`
   - `tests/test_voc121_implement_policy.py`
   - `tests/test_voc123_source_bundle.py`
   - documentation: fixture `CHANGELOG.md` from the same merge; current-state
     pin paragraph in the caller fixture `README.md` reflecting the infra
     README PR-context contract (do not replace that caller-owned README with
     the infrastructure repository README).
3. Add deterministic caller regressions for (a) exact pin consistency with
   #167 and inequality with `863fc1f…`, `8ce2b77…`, and `f3d791…`, (b) restore
   ordering in both release jobs, (c) #166 helper-copy-before-model and
   nested-checkout classifier behavior, (d) #167 immutable PR-context
   validation (`--pr-base-sha` / `--pr-head-sha`, `pr-validation` vs
   `pr-ancestry`, fail-closed comparison errors, no evidence fetch, CI
   `fetch-depth: 0`, implementer integration-anchor plus live HEAD including
   after self-correction), (e) SHA-256 identity of the mirrored #167 files
   without a machine-specific `/tmp` checkout, (f) byte-identity of all eight
   no-change paths against the immutable carrier-base SHA, (g) a complete
   caller-diff scan of changed executable paths for hydration/fetch/provenance
   bypasses, and (h) feasible exact-revision evidence that does not require a
   commit to contain its own SHA.
4. Preserve VOC-113 through VOC-134 fail-closed contracts already on the
   caller: #164 missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, roster markers, required-check recovery, independent
   review, retry caps, App-token isolation, Cursor Composer implementer,
   Cursor Grok review, unchanged `config/roles.yml`, no OpenAI route.
5. Update current-state fixture README / comments that would otherwise still
   name the #164 pin or omit the #165 restore / #166 nested-checkout / #167
   PR-context contract. Do not edit `AGENTS.md`,
   `.agents/skills/vocanova-repo-navigator/SKILL.md`,
   `scripts/foundation/voc112-navigation-benchmark.test.mjs`,
   `scripts/foundation/voc112-navigation-benchmark-run.mjs`,
   `scripts/foundation/validate-workspace.mjs`, or `package.json` for this pin.
6. After the exact reviewed caller pin is merged, complete ordinary
   develop-to-main promotion, exact develop synchronization without a staging
   redeploy for tree-equivalent sync, production deployment where selected,
   and audit reconciliation for VOC-129, exhausted / superseded VOC-127 and
   VOC-130 through VOC-134, and this package. This package's implementation
   PR `Closes` only its own VOC-135 task issue. Root issue #1071 closes only
   after that promotion/audit evidence exists.

### Non-goals / explicitly excluded

- Re-implementing or replacing already-merged infrastructure #165, #166, or
  #167.
- Opening a new `KARSIFT/karsift-ai-infra` PR unless a later exact-revision
  review proves the caller cannot consume `b263c0c…` as merged. That exception
  is not expected and is not this package's design.
- Silently changing VOC-132's adopted exact pin from #165, or VOC-133 /
  VOC-134's adopted pin from #166, in place.
- Redispatching VOC-132-T00 (#1059), VOC-133-T00 (#1063), or VOC-134-T00
  (#1068), resetting any attempt counter, or treating this package as another
  VOC-132 / VOC-133 / VOC-134 implementation attempt.
- Reusing, merging, cherry-picking, modifying, or composing PR #1051,
  PR #1056, PR #1065, PR #1070, or their branches.
- Re-implementing VOC-129, re-merging PR #1046, or manufacturing a VOC-127,
  VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, or VOC-134 completion marker.
- Changing any of the eight no-change paths, including weakening `local` mode
  fail-closed behavior on a missing capture commit in a full checkout,
  retargeting JSON `subject_revision` values, editing hashed sources to keep
  hashes current, or adding import-time fetch / hydrate side effects to the
  runner or `validate-workspace.mjs`.
- Adding any caller-side capture/evidence Git-object fetch,
  hydrate/materialize helper, package/test wrapper, import side effect,
  provenance-mode override, environment override, evidence mutation/stamping
  helper, skip, or equivalent under any filename or location.
- Requiring a tracked file to contain the SHA of the same commit that
  contains it.
- Using a moving branch ref (`develop`, `origin/develop`) as the no-change
  authority after the immutable carrier-base SHA is selected.
- Implementing against a moved `develop` tip. Issue #1071 requires fail-closed
  behavior if `develop` moves.
- Snapshotting the current develop/main gap as a later drift gate
  (`karsift-ai-infra#15`).
- Fast-forwarding `main` instead of preserving `--merge` promotion commits.
- Normalizing direct-to-main as the ordinary workflow.
- Operator-typed SHA inputs; overloading `existing_pr_number`.
- Changing `config/roles.yml`, adding OpenAI execution, or requesting
  `OPENAI_API_KEY`.
- Weakening exact-SHA review, risk floors, protected checks, App-token
  isolation, force-with-lease, retry caps, unique-develop fail-closed
  behavior, or fail-closed missing-binding behavior.
- Implementing VOC-122 promotion-recovery replan, merging unrelated open
  carriers, or rotating `KARSIFT_BOT_*` secrets / App installation
  permissions.
- A supervised bootstrap exception.
- Rewriting historical CHANGELOG, A-003, VOC-075, VOC-127, VOC-129, VOC-130,
  VOC-131, VOC-132, VOC-133, or VOC-134 audit records except for a new
  current-state note where a live contract would otherwise stay false. The
  fixture copy of infrastructure `CHANGELOG.md` may be replaced byte-for-byte
  from merge `b263c0c…`; that is a fixture mirror, not a rewrite of caller
  audit packages.
- Replacing the caller-owned fixture `README.md` with the infrastructure
  repository README.
- Copying nested infrastructure tests that are not already in the caller
  fixture subset merely to force a pin (`tests/test_implementer_bundle.py`
  and similar unmirrored infra-only tests), except the required new
  `tests/test_app_check_context.py`.
- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 no-change
  boundary, hydration-bypass regression, package.json identity, feasible
  evidence, or release reconciliation into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. VOC-127 through VOC-135 promotion/closure are
  ordinary release-path evidence, not a VOC-097 live-evidence gate.
- Depending on a machine-specific `/tmp` checkout of karsift-ai-infra as the
  test-time source of #167 bytes.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #167; current-state fixture README / pin comments.
  The eight no-change paths are protected *against* change: they must not
  appear in the implementation diff.
- Protected technical effect: whether the caller fixture and pin equal the
  authoritative #167 merge; whether application checks receive an immutable
  PR base/head pair without fetching evidence; whether both release jobs
  restore shared policy after caller checkout before task-completion helpers;
  whether post-model helpers survive nested-checkout removal; whether
  nested-checkout classification fails closed on ambiguous paths; whether
  VOC-112 capture identity and provenance behavior remain the published
  `develop` bytes; whether exact-revision binding remains fail-closed without
  a self-referential Git-tree SHA. No application runtime effect is intended
  for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-135-D00`: This is one outcome-sized caller replacement. Use one
end-to-end implementation task covering fixture/pin, the necessary #164→#167
files, #165 restore coverage, #166 nested-checkout coverage, #167 immutable
PR-context validation, the complete eight-path no-change boundary, complete
caller-diff hydration/bypass scan, `package.json` identity, hash-based #167
byte identity, feasible exact-revision evidence, tests, docs, and release
reconciliation. Coordinated infrastructure work is already merged as #167;
this task consumes that exact merge rather than re-implementing it.
Repository count, file count, and workflow-versus-tests-versus-docs are not
split reasons. VOC-129 and superseded VOC-127 / VOC-130 through VOC-134
promotion/closure after the repaired path is live are evidence of this
outcome, not additional VOC-135 tasks and not a further attempt on those
exhausted or unpublished carriers.

`VOC-135-D01`: Consume already-merged infrastructure #167. Do not compose a
replacement infra PR. Do not silently retarget VOC-132 from pin #165 or
VOC-133 / VOC-134 from pin #166. Do not redispatch VOC-132-T00 (#1059),
VOC-133-T00 (#1063), or VOC-134-T00 (#1068). Do not reuse, merge,
cherry-pick, modify, or compose PR #1051, PR #1056, PR #1065, or PR #1070 or
their branches. Do not treat the untracked local `karsift-ai-infra/` checkout
(if present) as this repository's tracked tree.

`VOC-135-D02`: Pin and fixture identity is exact merge
`b263c0c110591cc798b89277dfc35542abb1597b`. `PINNED_SHA.txt` and every live
pin assertion must equal that SHA. They must not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`, must not equal
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, and must not equal
`f3d79177bf8a9abe0dae550f39502165d494c576`. Mirror from that merge every
necessary changed authoritative file, byte-for-byte. Required set:
`tooling/governance/fixtures/karsift-ai-infra/.github/workflows/ci.yml`,
`.github/workflows/implement.yml`,
`.github/workflows/release.yml`,
`config/run-app-checks.sh`,
`config/implementer_nested_checkout.py`,
`tests/test_app_check_context.py`,
`tests/test_release_policy.py`,
`tests/test_voc121_implement_policy.py`,
`tests/test_voc123_source_bundle.py`,
and fixture `CHANGELOG.md`. If exact comparison proves another #167 file in
the existing caller fixture subset differs from the current caller fixture,
record that finding in `t00-evidence.md` and mirror it; do not copy unrelated
infra-only files merely to force a pin. Consumption of the #165 restore,
#166 nested-checkout, and #167 PR-context contracts in those files is
required.

`VOC-135-D03`: Both `identify` and `converge` must restore shared lifecycle
policy after `Checkout caller release state` and before
`task-completion-runner.py` (`validate-task` in identify, `validate-roster`
in converge). The restore checkout must use
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`.

`VOC-135-D04`: Preserve the already-approved #164 behavioral contract in the
mirrored fixture and live caller docs: absent-`develop` fallback /
checkout-ref ordering; after `--merge`, `develop` advances to
`mergeCommit.oid` before audit close; `reconcile-release` retries that sync;
unique develop commits, moved `main`, malformed merges, and missing/unbound
refs fail closed; auto-deleted `develop` is recreated at the merge SHA; equal
tips do not open a second promotion PR; tree-equivalent sync must not keep
staging scheduled; live `reconcile-production-change` remains the exceptional
main-only identity.

`VOC-135-D05`: Preserve #166 post-implementer nested-checkout behavior:

1. `Preserve post-implementer lifecycle helpers` copies
   `run-app-checks.sh`, `prepare_cursor_model.py`,
   `implementer_source_carrier.py`, `cross_repo_reference.py`, and
   `implementer_nested_checkout.py` to immutable `/tmp` paths **before** the
   unrestricted model runs.
2. After the model, `Commit implementer's work` classifies
   `karsift-ai-infra` with the preserved helper, not by reading the possibly
   deleted nested tree.
3. Classifier result `absent` means no infrastructure source carrier, while
   caller changes continue.
4. A surviving path that is a symlink, a non-directory, or a directory that
   inherits the caller Git root fails closed
   (`nested_checkout_symlink`, `nested_checkout_not_directory`,
   `nested_checkout_inherits_parent_git`).
5. A distinct nested Git checkout preserves exact-head bundle creation,
   ancestry, remote binding, and force-with-lease publication.

`VOC-135-D06`: Consume the #167 immutable PR-context validation contract
without fetching evidence:

1. `config/run-app-checks.sh` accepts `--pr-base-sha SHA --pr-head-sha SHA`
   or `--squash-safe-push`. Conflicting or malformed SHA pairs fail closed.
2. Both commits must exist locally (`git cat-file`) and share history
   (`git merge-base`) before any provenance export. Unavailable or unrelated
   commits fail closed. The script must not `git fetch` evidence commits.
3. When an exact PR pair is supplied, the script diffs
   `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
   between those trees. Unchanged selects `pr-validation`. Add, modify, or
   delete selects `pr-ancestry`. A comparison error (`git diff` status other
   than 0 or 1) fails closed.
4. Reusable `ci.yml` checks out the caller with `fetch-depth: 0` and, on
   `pull_request`, passes `github.event.pull_request.base.sha` /
   `head.sha` into `run-app-checks.sh`. Non-PR recovery uses
   `--squash-safe-push`.
5. Implementation pre-push and post-self-correction re-run pass
   `steps.branch.outputs.integration_sha` as base and `git rev-parse HEAD` as
   head, including after self-correction recommit.
6. The workflow never fetches missing VOC-112 evidence commits. Caller
   provenance assertions remain fail-closed.

`VOC-135-D07`: Deterministic tests must prove:

1. `PINNED_SHA.txt` equals `b263c0c110591cc798b89277dfc35542abb1597b` and does
   not equal `863fc1f…`, `8ce2b77…`, or `f3d791…`;
2. matching foundation pin literals equal the same SHA;
3. the mirrored authoritative files match recorded SHA-256 hashes of those
   paths at infra merge `b263c0c…`, without requiring a `/tmp`
   karsift-ai-infra checkout at test time;
4. fixture `release.yml` contains exactly two `Restore shared lifecycle policy
   after caller checkout` steps, one in `identify` and one in `converge`;
5. in each release job, caller checkout precedes restore, and restore precedes
   the task-completion helper;
6. restore uses `job.workflow_repository`, `job.workflow_sha`,
   `path: karsift-ai-infra`, and `persist-credentials: false`;
7. the #164 checkout-ref ordering / missing-`develop` fallback remains;
8. fixture `release.yml` still binds develop to `mergeCommit.oid` and does not
   restore `CHECKED_HEAD_SHA` as success;
9. fixture `implement.yml` copies helpers before the unrestricted model and
   classifies nested checkout from the preserved helper;
10. absent nested checkout continues caller publication; symlink /
    non-directory / parent-Git inheritance fail closed; a distinct nested
    checkout keeps exact-head bundle/lease publication;
11. fixture `run-app-checks.sh` binds exact PR context without fetching
    evidence; unchanged capture fixture selects `pr-validation`;
    add/modify/delete select `pr-ancestry`; comparison errors fail closed;
12. fixture `ci.yml` uses `fetch-depth: 0` and passes event base/head SHAs;
    implementer pre-push and post-self-correction use the integration anchor
    plus live committed HEAD;
13. live `pipeline.yml` still exposes `reconcile-production-change`, stays at
    most 25 `workflow_dispatch` inputs, and does not expose operator SHA
    inputs;
14. `config/roles.yml` is unchanged and no OpenAI route is added;
15. all eight no-change paths are byte-identical to the immutable
    carrier-base SHA and do not appear in `git diff` against that SHA;
16. the provenance test still fail-closes `local` mode when a full checkout
    is missing the captured commit (the #1056 weakening class);
17. no capture-commit fetch helper, hydrate/materialize helper,
    provenance-mode wrapper, import side effect, environment override,
    evidence-stamping helper, or test-time evidence mutation is present
    under any filename, proven by a complete caller-diff scan of changed
    executable paths rather than a single prohibited filename;
18. `t00-evidence.md` records the immutable carrier base, the final infra
    merge, confirmed hashes, validation, and the contract that the live
    implementation head is bound by the App-authored independent-review
    comment/check, and does not require a commit to contain its own SHA;
19. this package's caller PR `Closes` only its own VOC-135 task issue.

Tests must not mint real App tokens, use secrets, or use production data. The
mirrored fixture copies of `tests/test_app_check_context.py`,
`tests/test_voc121_implement_policy.py`, and related files must remain
byte-identical to merge `b263c0c…`. Any additional non-directory coverage
that those mirrored tests omit must live in caller `tooling/governance/tests/`,
not as an edit to the mirrored bytes.

`VOC-135-D08`: Preserve VOC-113 through VOC-134 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, `roles.yml`
unchanged (`implementer` / `implementer_escalation` `cursor/composer-2.5`;
`planner` / `reviewer` / `reviewer_fast_retry` / `plan_reviewer`
`cursor/grok-4.6[effort=high,fast=false]`), no secrets in
bundles/logs/fixtures, no credential values printed, sanitized raw-error
controls. This package's caller PR `Closes` only its own VOC-135 task issue.

`VOC-135-D09`: Current-state comments in fixture `implement.yml` /
`release.yml` / `ci.yml` / `run-app-checks.sh` / README that would otherwise
still name pin `863fc1f…` or omit the post-caller-checkout restore,
post-implementer helper-lifetime, or immutable PR-context contract must be
updated in the same task. Historical amendment records stay unchanged. Do
not edit `AGENTS.md`, the navigator skill, the provenance test, the
navigation-benchmark runner, `validate-workspace.mjs`, or `package.json` to
satisfy VOC-112 hash, provenance, or evidence checks. Do not replace the
caller-owned fixture `README.md` with the infrastructure repository README.
Update that README's current-state pin paragraph so it names pin `b263c0c…`
and the #167 PR-context contract.

`VOC-135-D10`: VOC-129 caller PR #1046 remains the VOC-129 carrier. Exhausted
VOC-130-T00 (#1049 / PR #1051), exhausted VOC-131-T00 (PR #1056), unpublished
VOC-132-T00 (#1059), exhausted VOC-133-T00 (#1063 / PR #1065), and exhausted
VOC-134-T00 (#1068 / PR #1070) are not retried. Do not manufacture a
VOC-127, VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, or VOC-134 completion
marker. After the repaired `implement.yml@main` / `release.yml@main` /
`ci.yml@main` path is live and this package's exact reviewed merge is
promoted:

1. VOC-129 promotion proceeds through ordinary release evaluation or
   `reconcile-release` when a valid App-authored completion marker exists;
2. this package promotes through the same repaired path;
3. audit comments name the exact promotion merges and close VOC-127,
   VOC-130, VOC-131, VOC-132, VOC-133, VOC-134, and this replacement, then
   root issue #1071, with precise links. Do not snapshot the current
   develop/main gap. Do not treat closed state alone as completion proof.

`VOC-135-D11`: Exact mirrored fixture bytes are proven by recorded SHA-256
hashes of the #167 files, committed in the caller regression so CI does not
depend on a machine-specific `/tmp` checkout of karsift-ai-infra.
Drafting-time hashes reconfirmed from an untracked nested checkout whose
`main` resolved to `b263c0c110591cc798b89277dfc35542abb1597b`:

| Fixture path | SHA-256 |
|--------------|---------|
| `.github/workflows/ci.yml` | `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb` |
| `.github/workflows/implement.yml` | `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a` |
| `.github/workflows/release.yml` | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` |
| `config/run-app-checks.sh` | `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2` |
| `config/implementer_nested_checkout.py` | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` |
| `tests/test_app_check_context.py` | `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0` |
| `tests/test_release_policy.py` | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` |
| `tests/test_voc121_implement_policy.py` | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` |
| `tests/test_voc123_source_bundle.py` | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` |
| `CHANGELOG.md` | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` |
| infrastructure `README.md` documentation source (do not overwrite caller-owned fixture README) | `06bcea7696ff80a59c9bab846061ce7a4c6d8bc175a1d02a2d7eefab0bcf46a3` |

Implementation must reconfirm those hashes from a fetched exact merge before
committing them. The untracked nested checkout is not this repository's
tracked tree and is not a test-time dependency. Note that `release.yml`,
`implementer_nested_checkout.py`, `tests/test_release_policy.py`,
`tests/test_voc121_implement_policy.py`, `tests/test_voc123_source_bundle.py`,
and `CHANGELOG.md` at #167 are byte-identical to the #166 pin recorded by
VOC-134-D11; they still must be mirrored because the live caller fixture
remains at #164. `ci.yml`, `implement.yml`, `run-app-checks.sh`, and
`tests/test_app_check_context.py` are the #167 delta.

`VOC-135-D12`: Feasible exact-revision evidence. Exact reviewed-revision
binding remains fail-closed. It is represented by the immutable App-authored
independent-review comment/check and PR metadata, and after merge by an
audit comment that may record the reviewed head and merge SHA. A tracked file
in a commit must not be required to contain the SHA of that same commit.

Committed `t00-evidence.md` must record:

1. the immutable carrier-base SHA selected before implementation;
2. the exact infra merge `b263c0c110591cc798b89277dfc35542abb1597b`;
3. the reconfirmed SHA-256 hashes;
4. the validation commands actually run;
5. the contract that the live implementation head is bound by the
   App-authored independent-review comment/check on the implementation PR,
   not by a self-referential value in the same Git tree;
6. where that later exact-head binding will be published (the independent
   review comment/check; post-merge/root-issue audit may additionally name
   the reviewed head and merge SHA).

The final independent-review comment must bind the live PR head exactly.
Merge-gate must reject any mismatch between that bound SHA and the head it
would merge. Tests must not rewrite `t00-evidence.md` in memory or on disk
at test/runtime to insert `HEAD`. Evidence must not claim a protected-path
revert unless `git diff` against the carrier base does not name that path.
The VOC-131 false-revert class, the VOC-133 evidence-stamp class, and the
VOC-134 hydrate-helper class are blocking defects if they recur. Evidence
must not copy raw provider responses, credentials, or the full log from
PR #1070.

`VOC-135-D13`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-135 carrier from current `develop`. Do not treat an
untracked local `karsift-ai-infra/` checkout as this repository's tracked
tree. Do not treat infrastructure merge `863fc1f…`, `8ce2b77…`, or `f3d791…`
as the pin target. Do not treat PR #1051, PR #1056, PR #1065, PR #1070,
VOC-132-T00 (#1059), VOC-133-T00 (#1063), or VOC-134-T00 (#1068) as a
publishable or composable source.

`VOC-135-D14`: Eight-path no-change boundary and provenance-bypass
prohibition. All of the following must remain byte-for-byte identical to the
immutable carrier-base SHA selected before implementation and must be absent
from the implementation PR diff:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `scripts/foundation/voc112-navigation-benchmark-run.mjs`
- `scripts/foundation/validate-workspace.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`
- `package.json`

At current `develop` the JSON `subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. Do not retarget or recapture the
JSON fixtures. Do not weaken `local` mode. Do not add
`scripts/foundation/ensure-voc112-capture-commits.mjs`,
`scripts/foundation/hydrate-voc112-git-objects.mjs`, or any equivalent under
any filename. Do not wrap `pnpm test` or foundation tests with
`VOC112_CAPTURE_PROVENANCE_MODE`, with PR SHAs set to `HEAD`, or with any
other invocation that masks default `local` fail-closed behavior. Do not add
an import-time Git fetch of capture subjects. Do not add an evidence-stamping
helper. Do not modify evidence at test or runtime.

`VOC-135-D15`: Immutable carrier-base SHA. Expected current `develop` at
drafting is `b9e74fc2db4691c48c637639b265d527de9f4505`. Implementation must
resolve current `develop` to a 40-character SHA **before any in-scope edit**
and record that SHA as the immutable carrier base at PR creation. If that
SHA equals `b9e74fc2db4691c48c637639b265d527de9f4505`, that SHA is the
immutable carrier base. If it differs, fail closed: do not implement against
a moved tip. Issue #1071 requires this fail-closed rule; do not retarget
after files have already changed. After the constant is selected, tests must
not use a moving branch ref (`develop`, `origin/develop`) as the comparison
authority. The regression must require that exact commit object and all eight
no-change paths to resolve (`git cat-file` / `git show SHA:path` or
equivalent) and must fail—not skip—if the commit or any named path cannot be
read.

`VOC-135-D16`: Complete caller-diff hydration/bypass scan. Regression coverage
must inspect the complete caller diff / changed executable paths against the
immutable carrier-base SHA, not only one prohibited filename. It must fail
if any newly added caller script or any changed existing caller script:

1. invokes Git fetch for VOC-112 capture subjects or evidence objects;
2. sets `VOC112_CAPTURE_PROVENANCE_MODE`, `PR_BASE_SHA`, or `PR_HEAD_SHA`
   around tests, `pnpm test`, or `validate-workspace`;
3. hydrates, materializes, or otherwise plants evidence Git objects;
4. adds an import-time or pre-test side effect that fetches or stamps
   evidence;
5. skips or otherwise makes default `local` fail-closed unreachable.

The scan must include `scripts/**`, `package.json`, and any other added or
modified executable path in the implementation diff. A filename allowlist
that only bans `hydrate-voc112-git-objects.mjs` is insufficient: VOC-134
attempt 2 relocated the same bypass. Fixture copies of infra
`run-app-checks.sh` that legitimately export provenance for an immutable PR
pair are in-scope mirrors of #167, not caller-side overrides, and are not
this scan's failure class.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The #165 restore uses
the existing reusable-workflow checkout (`job.workflow_repository` /
`job.workflow_sha`) with `persist-credentials: false`. The #166 nested-checkout
classifier and helper copies use existing `/tmp` runner paths. The #167
PR-context pair uses already-present Git objects (CI `fetch-depth: 0` and
the implementation integration checkout) and must not fetch missing evidence
commits, must not grant the model-controlled runner the App token, and must
not accept free-form SHAs as authority.

Abuse/process risks:

1. Pinning #164 (`863fc1f…`), #165 (`8ce2b77…`), or #166 (`f3d791…`) after
   #167 merged — forbidden by `VOC-135-D02`.
2. Silently retargeting VOC-132 / VOC-133 / VOC-134 or redispatched T00
   issues — forbidden by `VOC-135-D01` and `VOC-135-D10`.
3. Covering only `identify` and leaving `converge` unrestored — forbidden by
   `VOC-135-D03`.
4. Copying helpers after the unrestricted model, or treating a surviving
   non-distinct nested path as `absent` — forbidden by `VOC-135-D05`.
5. Fetching missing VOC-112 evidence commits from `run-app-checks.sh`, CI, or
   implementer pre-push — forbidden by `VOC-135-D06`.
6. Reusing PR #1051, PR #1056, PR #1065, or PR #1070 — forbidden by
   `VOC-135-D01`.
7. Re-implementing VOC-129, manufacturing a VOC-127 through VOC-134
   completion marker, or snapshotting the develop/main gap — forbidden by
   `VOC-135-D10` and `karsift-ai-infra#15`.
8. Retargeting VOC-112 JSON evidence — the VOC-130 exhaustion class,
   forbidden by `VOC-135-D14`.
9. Weakening VOC-112 provenance `local` mode fail-closed behavior, or
   claiming that edit was reverted while the tree still contains it — the
   VOC-131 exhaustion class, forbidden by `VOC-135-D14` and `VOC-135-D12`.
10. Fetching a missing capture commit before `pnpm test`, wrapping tests in
    `pr-validation` via `package.json`, adding an import-time fetch in the
    runner, or hydrating evidence objects from `validate-workspace.mjs` —
    the VOC-133 / VOC-134 exhaustion classes, forbidden by `VOC-135-D14` and
    `VOC-135-D16`.
11. Requiring a commit to contain its own SHA — the VOC-133-D12 defect,
    forbidden by `VOC-135-D12`.
12. Using a moving `develop` ref as the no-change authority after merge, or
    implementing against a moved tip — forbidden by `VOC-135-D15`.
13. Validating #167 bytes only against a machine-specific `/tmp` checkout —
    forbidden by `VOC-135-D11`.
14. Scanning only one prohibited filename for hydration bypasses — forbidden
    by `VOC-135-D16`.
15. Erasing unique develop commits — still forbidden by VOC-127-D02, preserved
    in the #164/#165/#166/#167 fixture.
16. Treating a missing task-completion helper as a successful no-op after this
    repair — the restore must make the helper present; tests must fail if it
    can still disappear before use.
17. Printing App tokens, private keys, or secret values, or copying full CI
    logs / raw provider responses into evidence — forbidden.

## Contradictions and open questions

1. **Live caller workflow edit (`VOC-135-DEP-07`):** required behavior is
   settled (pin/fixture equal #167; both jobs restore shared policy; #166
   nested-checkout contract is present; #167 PR-context contract is present).
   Whether any live `.github/workflows/*.yml` file besides the fixture must
   change is proven at implementation time. Expected: no live `pipeline.yml`
   edit because the caller already calls `implement.yml@main` and
   `release.yml@main`.
2. **Live pin vs issue wording:** issue #1071 says the caller fixture is still
   pinned to #166. Live tracked `PINNED_SHA.txt` at drafting equals #164
   (`863fc1f…`) because VOC-134 never merged. This is not an ambiguity about
   the pin target: this package pins to #167 (`b263c0c…`) either way, and
   that merge already contains the #166 files. Implementation must not leave
   the pin at #164 or #166.
3. **Caller fixture README vs infrastructure README (`VOC-135-DEP-12`):** they
   are different documents. Update the caller fixture README's current-state
   pin paragraph to name `b263c0c…` and the #167 PR-context contract. Do not
   overwrite it with `KARSIFT/karsift-ai-infra/README.md`. The infra README
   SHA-256 is recorded in `VOC-135-D11` as the documentation-source hash.
4. **AGENTS.md / DOC-15 / DOC-16:** do not update those hashed or historical
   documents to name pin `b263c0c…` if doing so would require VOC-112
   recapture. Put the current-state pin sentence in fixture README and this
   package's evidence. Historical correction notes stay unchanged.
5. **Infrastructure already merged:** this package must not snapshot, re-open,
   or rewrite #167. If later exact-revision review finds the caller cannot
   consume `b263c0c…`, record that as a blocking finding rather than silently
   forking infrastructure inside the caller.
6. **VOC-134 exhausted carrier:** VOC-134 remains an adopted #166 package
   whose implementation exhausted. This replacement consumes #167 under new
   authority rather than editing VOC-134 in place or redispatched #1068.
7. **PR #1070 and earlier unmerged carriers:** the pin/restore portion of
   those unmerged carriers is not authority to cherry-pick the whole heads. A
   fresh carrier from current `develop` is required so VOC-112 JSON
   retargets, provenance-test weakening, `package.json` provenance wrappers,
   capture-commit fetch helpers, import-time fetch, and hydrate helpers
   cannot ride along.
8. **Drafting-time SHA-256 values (`VOC-135-D11`):** hashes above were
   reconfirmed from an untracked nested checkout at `b263c0c…`. Implementation
   must reconfirm them from a fetched exact merge. If they differ, record the
   fetched hashes in `t00-evidence.md` and use those as the committed
   expected values; do not silently keep a stale drafting-time hash.
9. **Non-directory classifier coverage:** the #166/#167 helper rejects a
   non-directory nested path, but the mirrored
   `tests/test_voc121_implement_policy.py` at drafting time has no dedicated
   `nested_checkout_not_directory` test. Caller `tooling/governance/tests/`
   must cover that class without editing the mirrored fixture test bytes.
10. **Immutable carrier-base SHA (`VOC-135-D15`):** expected drafting-time
    develop is `b9e74fc2db4691c48c637639b265d527de9f4505`. If live `develop`
    has moved at implementation dispatch, fail closed. Do not guess a new
    base after files have already changed.
11. **VOC-129 promotion timing:** hosted `release.yml@main` / `implement.yml@main`
    / `ci.yml@main` may already contain the #165 / #166 / #167 repairs before
    this caller pin merges. That does not remove this package's pin/test/doc
    obligation. Do not treat VOC-129 promotion as this package's
    implementation diff.
