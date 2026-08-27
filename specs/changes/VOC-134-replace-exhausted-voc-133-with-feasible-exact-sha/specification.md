# VOC-134 — Replace exhausted VOC-133 with feasible exact-SHA evidence and complete infra #166 release: Specification

## Objective and requirement source

Deliver the caller-side pin, fixture, tests, docs, and feasible exact-revision
evidence that consume the exact authoritative infrastructure merge
`f3d79177bf8a9abe0dae550f39502165d494c576` (KARSIFT/karsift-ai-infra#166)
from a fresh branch based on current `develop`, with a complete VOC-112
no-change boundary and an unchanged `package.json`. The final caller must
restore shared lifecycle policy after the root caller checkout in both
`identify` and `converge` before task-completion helpers run, preserve #166
post-implementer nested-checkout classification and helper-lifetime behavior,
keep the five named VOC-112 no-change paths and `package.json` byte-identical
to an immutable carrier-base SHA selected before implementation, pass
independent exact-revision review and all protected checks without
self-referential commit SHAs or provenance-mode bypasses, and then complete
ordinary develop-to-main promotion for this package and the still-unfinished
VOC-129 promotion that run `33066533397` skipped.

**Requirement source:** [GitHub issue #1066](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1066).
VOC-129 (`specs/changes/VOC-129-replace-exhausted-voc-127-caller-carrier-with-the`)
already merged caller PR [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046)
at `429d8c6d49303148ca1cc14dba5f6768a7863346`. VOC-130, VOC-131, VOC-132, and
VOC-133 adopted overlapping #165/#166 pin outcomes; VOC-130-T00 (#1049 /
PR #1051) and VOC-131-T00 (PR #1056) exhausted on VOC-112 edits,
VOC-132-T00 (#1059) failed in run `33079499176` before any caller branch, PR,
or completion marker existed, and VOC-133-T00 (#1063 / PR #1065) exhausted
both attempts on provenance-bypass and an impossible self-referential
exact-head evidence contract. This package is the governed replacement
carrier. It is not a second VOC-129 implementation, not a retry of VOC-130
through VOC-133, and not a silent retarget of VOC-132's exact #165 pin or
VOC-133's exhausted carrier.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

### Confirmed problem evidence (issue #1066)

| Item | Value |
|------|-------|
| Exhausted VOC-130 task | [#1049](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049) (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | [#1051](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051) |
| Exhausted VOC-131 carrier | [#1056](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056) |
| Adopted VOC-132 plan | [#1058](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1058) |
| VOC-132 task that must not be redispatched | [#1059](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1059) (`VOC-132-T00`) |
| VOC-132 failed implementer run | [`33079499176`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33079499176), job `98543693163` |
| Adopted VOC-133 plan | [#1062](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1062) |
| Exhausted VOC-133 task | [#1063](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1063) (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | [#1065](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065) |
| VOC-133 attempt-1 FAIL head | `e88fbda6274488bed898877151bbcc1714a45450` |
| VOC-133 attempt-1 FAIL | [comment 5441321518](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441321518) |
| VOC-133 attempt-1 supervisor finding | [comment 5441214837](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441214837) |
| VOC-133 attempt-2 FAIL head | `70930cf0572f73f254429a240c5328640846e0a9` |
| VOC-133 attempt-2 FAIL | [comment 5441659390](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441659390) |
| VOC-129 caller PR | [#1046](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1046) at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release run | [`33066533397`](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33066533397) |
| Authoritative infra merge | `f3d79177bf8a9abe0dae550f39502165d494c576` (#166) |
| Failed initial infra review head | `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f` |
| Independently reviewed PASS infra head | `1488619d0d37aaa179d8e739bfe931881d6c51aa` |
| Current `develop` pin | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) |
| Expected immutable carrier base | `95a779f9e62090f856ed03f389e7ac1d901aaa14` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Why VOC-133 cannot retry on #1063 / #1065 | both allowed attempts exhausted; attempt 1 fetched a missing capture commit before `pnpm test`; attempt 2 forced `pr-validation` via `package.json` and mutated evidence at test time; VOC-133-D12/TEST-10 self-referential SHA is Git-impossible |
| Why VOC-132 cannot be retargeted or redispatched | VOC-132's adopted pin is exact #165; run `33079499176` created no publishable caller carrier |
| Why VOC-130 / VOC-131 cannot retry | both allowed attempts exhausted on VOC-112 JSON retargets / provenance-test weakening |
| Why VOC-129 is not retried | #1046 already merged; this is a pin plus nested-checkout / restore consumption plus feasible evidence |

## Scope and non-goals

### In scope

1. Set `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` and every
   live pin assertion (caller governance tests and matching
   `scripts/foundation/*` pin literals) to
   `f3d79177bf8a9abe0dae550f39502165d494c576`.
2. Mirror from exact infra merge #166 every necessary changed authoritative
   #164→#166 caller fixture file, byte-for-byte. Required pair plus helper and
   tests, matching issue #1066:
   - `.github/workflows/implement.yml`
   - `.github/workflows/release.yml`
   - `config/implementer_nested_checkout.py` (new in the caller fixture)
   - `tests/test_release_policy.py`
   - `tests/test_voc121_implement_policy.py`
   - `tests/test_voc123_source_bundle.py`
   - documentation: fixture `CHANGELOG.md` from the same merge; current-state
     pin paragraph in the caller fixture `README.md` (do not replace that
     caller-owned README with the infrastructure repository README).
3. Add deterministic caller regressions for (a) exact pin consistency with
   #166 and inequality with both `863fc1f…` and `8ce2b77…`, (b) restore
   ordering in both release jobs using `job.workflow_repository` /
   `job.workflow_sha` / `path: karsift-ai-infra` with
   `persist-credentials: false`, (c) #166 helper-copy-before-model and
   nested-checkout classifier behavior, (d) SHA-256 identity of the mirrored
   #166 files without a machine-specific `/tmp` checkout, (e) byte-identity
   of all five VOC-112 no-change paths and `package.json` against the
   immutable carrier-base SHA, and (f) feasible exact-revision evidence that
   does not require a commit to contain its own SHA.
4. Preserve VOC-113 through VOC-133 fail-closed contracts already on the
   caller: #164 missing-`develop` recovery, branch exact-SHA
   synchronization, unique-develop fail-closed behavior, promotion checks,
   release serialization, roster markers, required-check recovery, independent
   review, retry caps, App-token isolation, Cursor Composer implementer,
   Cursor Grok review, unchanged `config/roles.yml`, no OpenAI route.
5. Update current-state fixture README / comments that would otherwise still
   name the #164 pin or omit the #165 restore / #166 nested-checkout contract.
   Do not edit `AGENTS.md`, `.agents/skills/vocanova-repo-navigator/SKILL.md`,
   `scripts/foundation/voc112-navigation-benchmark.test.mjs`, or
   `package.json` for this pin.
6. After the exact reviewed caller pin is merged, complete ordinary
   develop-to-main promotion, exact develop synchronization, production
   deployment where selected, and audit reconciliation for VOC-129, exhausted
   VOC-130, exhausted VOC-131, superseded VOC-132, exhausted VOC-133, and this
   package. This package's implementation PR `Closes` only its own VOC-134
   task issue. Root issue #1066 closes only after that promotion/audit
   evidence exists.

### Non-goals / explicitly excluded

- Re-implementing or replacing already-merged infrastructure #165 or #166.
- Opening a new `KARSIFT/karsift-ai-infra` PR unless a later exact-revision
  review proves the caller cannot consume `f3d791…` as merged. That exception
  is not expected and is not this package's design.
- Silently changing VOC-132's adopted exact pin from #165 to #166.
- Redispatching VOC-132-T00 (#1059) or VOC-133-T00 (#1063), resetting either
  attempt counter, or treating this package as another VOC-132 or VOC-133
  implementation attempt.
- Reusing, merging, cherry-picking, modifying, or composing PR #1051,
  PR #1056, PR #1065, or their branches.
- Re-implementing VOC-129, re-merging PR #1046, or manufacturing a VOC-129,
  VOC-130, VOC-131, VOC-132, or VOC-133 completion marker.
- Changing any of the five VOC-112 no-change paths, including weakening
  `local` mode fail-closed behavior on a missing capture commit in a full
  checkout, retargeting JSON `subject_revision` values, or editing hashed
  sources to keep hashes current.
- Changing `package.json`, adding
  `scripts/foundation/ensure-voc112-capture-commits.mjs` or any other
  capture-commit fetch helper, adding a provenance-mode wrapper, adding an
  evidence-stamping helper, mutating evidence at test or runtime, or any
  invocation that masks default `local` fail-closed behavior.
- Requiring a tracked file to contain the SHA of the same commit that
  contains it.
- Using a moving branch ref (`develop`, `origin/develop`) as the VOC-112
  no-change authority after the immutable carrier-base SHA is selected.
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
  VOC-131, VOC-132, or VOC-133 audit records except for a new current-state
  note where a live contract would otherwise stay false. The fixture copy of
  infrastructure `CHANGELOG.md` may be replaced byte-for-byte from merge
  `f3d791…`; that is a fixture mirror, not a rewrite of caller audit
  packages.
- Replacing the caller-owned fixture `README.md` with the infrastructure
  repository README.
- Copying nested infrastructure tests that are not already in the caller
  fixture subset merely to force a pin (`tests/test_implementer_bundle.py`
  and similar unmirrored infra-only tests).
- Changing application runtime behavior, deployment topology, product
  permissions, or monitor inventory.
- Splitting workflow logic, tests, docs, fixture/pin, VOC-112 no-change
  boundary, package.json identity, feasible evidence, or release
  reconciliation into separate tasks.
- Self-adoption or self-authorization of this package.
- Operator-owned live-evidence contracts: acceptance is deterministic tests
  plus exact-SHA review. VOC-129 through VOC-134 promotion/closure are
  ordinary release-path evidence, not a VOC-097 live-evidence gate.
- Depending on a machine-specific `/tmp` checkout of karsift-ai-infra as the
  test-time source of #166 bytes.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #166; current-state fixture README / pin comments.
  The five VOC-112 no-change paths and `package.json` are protected *against*
  change: they must not appear in the implementation diff.
- Protected technical effect: whether the caller fixture and pin equal the
  authoritative #166 merge; whether both release jobs restore shared policy
  after caller checkout before task-completion helpers; whether post-model
  helpers survive nested-checkout removal; whether nested-checkout
  classification fails closed on ambiguous paths; whether VOC-112 capture
  identity and provenance behavior remain the published `develop` bytes;
  whether exact-revision binding remains fail-closed without a
  self-referential Git-tree SHA. No application runtime effect is intended
  for ordinary promotions.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but workflow and governance-fixture changes still
  require exact-SHA independent verification and fail-closed controls.

This risk class is a **draft proposal** for adoption review, not a
determination. The path-based classifier and independent verifier remain
authoritative.

## Decisions

`VOC-134-D00`: This is one outcome-sized caller replacement. Use one
end-to-end implementation task covering fixture/pin, the necessary #164→#166
files, #165 restore coverage, #166 nested-checkout coverage, the complete
VOC-112 no-change boundary, `package.json` identity, hash-based #166 byte
identity, feasible exact-revision evidence, tests, docs, and release
reconciliation. Coordinated infrastructure work is already merged as #166;
this task consumes that exact merge rather than re-implementing it.
Repository count, file count, and workflow-versus-tests-versus-docs are not
split reasons. VOC-129 and superseded VOC-130 through VOC-133
promotion/closure after the repaired path is live are evidence of this
outcome, not additional VOC-134 tasks and not a further attempt on those
exhausted or unpublished carriers.

`VOC-134-D01`: Consume already-merged infrastructure #166. Do not compose a
replacement infra PR. Do not silently retarget VOC-132 from pin #165 to pin
#166. Do not redispatch VOC-132-T00 (#1059) or VOC-133-T00 (#1063). Do not
reuse, merge, cherry-pick, modify, or compose PR #1051, PR #1056, or PR #1065
or their branches. Do not treat the untracked local `karsift-ai-infra/`
checkout (if present) as this repository's tracked tree.

`VOC-134-D02`: Pin and fixture identity is exact merge
`f3d79177bf8a9abe0dae550f39502165d494c576`. `PINNED_SHA.txt` and every live
pin assertion must equal that SHA. They must not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681` and must not equal
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Mirror from that merge every
necessary changed authoritative file, byte-for-byte. Required set:
`tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml`,
`.github/workflows/release.yml`,
`config/implementer_nested_checkout.py`,
`tests/test_release_policy.py`,
`tests/test_voc121_implement_policy.py`,
`tests/test_voc123_source_bundle.py`,
and fixture `CHANGELOG.md`. If exact comparison proves another #166 file in
the existing caller fixture subset differs from the current caller fixture,
record that finding in `t00-evidence.md` and mirror it; do not copy unrelated
infra-only files merely to force a pin. Consumption of the #165 restore and
#166 nested-checkout contracts in those files is required.

`VOC-134-D03`: Both `identify` and `converge` must restore shared lifecycle
policy after `Checkout caller release state` and before
`task-completion-runner.py` (`validate-task` in identify, `validate-roster`
in converge). The restore checkout must use
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`.

`VOC-134-D04`: Preserve the already-approved #164 behavioral contract in the
mirrored fixture and live caller docs: absent-`develop` fallback /
checkout-ref ordering; after `--merge`, `develop` advances to
`mergeCommit.oid` before audit close; `reconcile-release` retries that sync;
unique develop commits, moved `main`, malformed merges, and missing/unbound
refs fail closed; auto-deleted `develop` is recreated at the merge SHA; equal
tips do not open a second promotion PR; tree-equivalent sync must not keep
staging scheduled; live `reconcile-production-change` remains the exceptional
main-only identity.

`VOC-134-D05`: Preserve #166 post-implementer nested-checkout behavior:

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

`VOC-134-D06`: Deterministic tests must prove:

1. `PINNED_SHA.txt` equals `f3d79177bf8a9abe0dae550f39502165d494c576` and does
   not equal `863fc1f35b1d35e4981a59166b0e939be1a2b681` or
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`;
2. matching foundation pin literals equal the same SHA;
3. the mirrored authoritative files match recorded SHA-256 hashes of those
   paths at infra merge `f3d791…`, without requiring a `/tmp`
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
11. live `pipeline.yml` still exposes `reconcile-production-change`, stays at
    most 25 `workflow_dispatch` inputs, and does not expose operator SHA
    inputs;
12. `config/roles.yml` is unchanged and no OpenAI route is added;
13. all five VOC-112 no-change paths and `package.json` are byte-identical to
    the immutable carrier-base SHA and do not appear in `git diff` against
    that SHA;
14. the provenance test still fail-closes `local` mode when a full checkout
    is missing the captured commit (the #1056 weakening class);
15. no capture-commit fetch helper, provenance-mode wrapper, evidence-stamping
    helper, or test-time evidence mutation is present;
16. `t00-evidence.md` records the immutable carrier base, the final infra
    merge, confirmed hashes, validation, and the contract that the live
    implementation head is bound by the App-authored independent-review
    comment/check, and does not require a commit to contain its own SHA;
17. this package's caller PR `Closes` only its own VOC-134 task issue.

Tests must not mint real App tokens, use secrets, or use production data. The
mirrored fixture copies of `tests/test_voc121_implement_policy.py` and related
files must remain byte-identical to merge `f3d791…`. Any additional
non-directory coverage that those mirrored tests omit must live in caller
`tooling/governance/tests/`, not as an edit to the mirrored bytes.

`VOC-134-D07`: Preserve VOC-113 through VOC-133 fail-closed contracts: roster
markers over issue-closed state, exact-head promotion checks, ruleset
attestation, required-check recovery, App-token mutation isolation, job-token
recovery reads, force-with-lease / match-head-commit, two-attempt implementer
bound, nested isolation, credential-free bundles, Cursor Composer
implementer, Cursor Grok exact-revision review, no OpenAI route, `roles.yml`
unchanged, no secrets in bundles/logs/fixtures, no credential values printed,
sanitized raw-error controls. This package's caller PR `Closes` only its own
VOC-134 task issue.

`VOC-134-D08`: Current-state comments in fixture `implement.yml` /
`release.yml` / README that would otherwise still name pin `863fc1f…` or omit
the post-caller-checkout restore or post-implementer helper-lifetime contract
must be updated in the same task. Historical amendment records stay
unchanged. Do not edit `AGENTS.md`, the navigator skill,
`voc112-navigation-benchmark.test.mjs`, or `package.json` to satisfy VOC-112
hash, provenance, or evidence checks; those hashed sources, the two VOC-112
fixtures, the provenance test, and `package.json` are out of scope. Do not
replace the caller-owned fixture `README.md` with the infrastructure
repository README.

`VOC-134-D09`: VOC-129 caller PR #1046 remains the VOC-129 carrier. Exhausted
VOC-130-T00 (#1049 / PR #1051), exhausted VOC-131-T00 (PR #1056), unpublished
VOC-132-T00 (#1059), and exhausted VOC-133-T00 (#1063 / PR #1065) are not
retried. Do not manufacture a VOC-129, VOC-130, VOC-131, VOC-132, or VOC-133
completion marker. After the repaired `implement.yml@main` / `release.yml@main`
path is live and this package's exact reviewed merge is promoted:

1. VOC-129 promotion that run `33066533397` skipped proceeds through ordinary
   release evaluation or `reconcile-release` when a valid App-authored
   completion marker exists;
2. this package promotes through the same repaired path;
3. audit comments name the exact promotion merges and close VOC-129,
   VOC-130, VOC-131, VOC-132, VOC-133, and this replacement, then root issue
   #1066, with precise links. Do not snapshot the current develop/main gap.
   Do not treat closed state alone as completion proof.

`VOC-134-D10`: Complete VOC-112 no-change boundary. All of the following must
remain byte-for-byte identical to the immutable carrier-base SHA selected
before implementation and must be absent from the implementation PR diff:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`

At current `develop` the JSON `subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. Do not retarget or recapture the
JSON fixtures, including as a "provenance fix" for `pr-ancestry` mode.
Changing `voc112-navigation-benchmark-traces.json` also switches
`repository-governance.yml` into `pr-ancestry` mode; this package must not
take that path. Do not weaken `local` mode: a full checkout missing the
captured commit must still fail closed, not fall through to implicit
`squash-safe-push`. A deterministic regression must fail if any named path
differs from that immutable SHA or appears in the implementation diff.
VOC-130/VOC-131 JSON-only guards and VOC-133 moving-`develop`-ref guards are
insufficient.

`VOC-134-D11`: Exact mirrored fixture bytes are proven by recorded SHA-256
hashes of the #166 files, committed in the caller regression so CI does not
depend on a machine-specific `/tmp` checkout of karsift-ai-infra. Drafting-time
hashes reconfirmed from an untracked nested checkout whose `main` resolved to
`f3d79177bf8a9abe0dae550f39502165d494c576`:

| Fixture path | SHA-256 |
|--------------|---------|
| `.github/workflows/implement.yml` | `5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c` |
| `.github/workflows/release.yml` | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` |
| `config/implementer_nested_checkout.py` | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` |
| `tests/test_release_policy.py` | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` |
| `tests/test_voc121_implement_policy.py` | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` |
| `tests/test_voc123_source_bundle.py` | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` |
| `CHANGELOG.md` | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` |

Implementation must reconfirm those hashes from a fetched exact merge before
committing them. The untracked nested checkout is not this repository's
tracked tree and is not a test-time dependency. Note that `release.yml` and
`tests/test_release_policy.py` at #166 are byte-identical to the #165 restore
contract; they still must be mirrored because the live caller fixture remains
at #164.

`VOC-134-D12`: Feasible exact-revision evidence. Exact reviewed-revision
binding remains fail-closed. It is represented by the immutable App-authored
independent-review comment/check and PR metadata, and after merge by an
audit comment that may record the reviewed head and merge SHA. A tracked file
in a commit must not be required to contain the SHA of that same commit:
changing the file would change the commit SHA, which is the VOC-133-D12 /
TEST-10 defect that made PR #1065 unsatisfiable.

Committed `t00-evidence.md` must record:

1. the immutable carrier-base SHA selected before implementation;
2. the exact infra merge `f3d79177bf8a9abe0dae550f39502165d494c576`;
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
The VOC-131 false-revert class on #1056 and the VOC-133 evidence-stamp class
on #1065 are blocking defects if they recur. Evidence must not copy raw
provider responses, credentials, or the full log from run `33079499176` or
PR #1065.

`VOC-134-D13`: No bootstrap exception. VOC-124 already requested
`permission-workflows: write` on `publish-source`. T00's first run is attempt
`1` on a new VOC-134 carrier from current `develop`. Do not treat an
untracked local `karsift-ai-infra/` checkout as this repository's tracked
tree. Do not treat infrastructure merge `863fc1f…` or `8ce2b77…` as the pin
target. Do not treat PR #1051, PR #1056, PR #1065, VOC-132-T00 (#1059), or
VOC-133-T00 (#1063) as a publishable or composable source.

`VOC-134-D14`: `package.json` identity and provenance-mode prohibition. Live
`package.json` at the carrier base must remain byte-identical and absent from
the implementation diff. Do not add `scripts/foundation/ensure-voc112-capture-commits.mjs`
or any other capture-commit fetch helper. Do not wrap `pnpm test` or
foundation tests with `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation`, with
PR SHAs set to `HEAD`, or with any other invocation that masks default
`local` fail-closed behavior. Do not add an evidence-stamping helper. Do not
modify evidence at test or runtime. A deterministic assertion must compare
`package.json` to the immutable carrier-base SHA and fail if it differs or
cannot be resolved.

`VOC-134-D15`: Immutable carrier-base SHA. Expected current `develop` at
drafting is `95a779f9e62090f856ed03f389e7ac1d901aaa14`. Implementation must
resolve current `develop` to a 40-character SHA **before any in-scope edit**.
If that SHA equals `95a779f9e62090f856ed03f389e7ac1d901aaa14`, that SHA is
the immutable carrier base. If it differs, fail closed: do not implement
against a moved tip without first recording the newly observed exact SHA as
the pre-edit immutable constant in the regression and in `t00-evidence.md`.
A moved tip is unexpected at drafting time. After the constant is selected,
tests must not use a moving branch ref (`develop`, `origin/develop`) as the
comparison authority. The regression must require that exact commit object
and all five VOC-112 paths plus `package.json` to resolve (`git cat-file`
/`git show SHA:path` or equivalent) and must fail—not skip—if the commit or
any named path cannot be read.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is governed-automation reliability
work only. No database, schema, seed, analytics instrumentation, or
user-interface accessibility effect.

## Security, privacy, and authorization

No new secret values are written into the repository. The #165 restore uses
the existing reusable-workflow checkout (`job.workflow_repository` /
`job.workflow_sha`) with `persist-credentials: false`. The #166 nested-checkout
classifier and helper copies use existing `/tmp` runner paths and do not
broaden installation permissions, must not grant the model-controlled runner
the App token, and must not accept free-form SHAs as authority.

Abuse/process risks:

1. Pinning #164 (`863fc1f…`) or #165 (`8ce2b77…`) after #166 merged —
   forbidden by `VOC-134-D02`.
2. Silently retargeting VOC-132 or redispatched VOC-132-T00 (#1059) or
   VOC-133-T00 (#1063) — forbidden by `VOC-134-D01` and `VOC-134-D09`.
3. Covering only `identify` and leaving `converge` unrestored — forbidden by
   `VOC-134-D03`.
4. Copying helpers after the unrestricted model, or treating a surviving
   non-distinct nested path as `absent` — forbidden by `VOC-134-D05`.
5. Reusing PR #1051, PR #1056, or PR #1065 — forbidden by `VOC-134-D01`.
6. Re-implementing VOC-129, manufacturing a VOC-129 through VOC-133
   completion marker, or snapshotting the develop/main gap — forbidden by
   `VOC-134-D09` and `karsift-ai-infra#15`.
7. Retargeting VOC-112 JSON evidence — the VOC-130 exhaustion class,
   forbidden by `VOC-134-D10`.
8. Weakening VOC-112 provenance `local` mode fail-closed behavior, or
   claiming that edit was reverted while the tree still contains it — the
   VOC-131 exhaustion class, forbidden by `VOC-134-D10` and `VOC-134-D12`.
9. Fetching a missing capture commit before `pnpm test`, wrapping tests in
   `pr-validation` via `package.json`, or mutating evidence at test time —
   the VOC-133 exhaustion class, forbidden by `VOC-134-D14`.
10. Requiring a commit to contain its own SHA — the VOC-133-D12 defect,
    forbidden by `VOC-134-D12`.
11. Using a moving `develop` ref as the no-change authority after merge —
    forbidden by `VOC-134-D15`.
12. Validating #166 bytes only against a machine-specific `/tmp` checkout —
    forbidden by `VOC-134-D11`.
13. Erasing unique develop commits — still forbidden by VOC-127-D02, preserved
    in the #164/#165/#166 fixture.
14. Treating a missing task-completion helper as a successful no-op after this
    repair — the restore must make the helper present; tests must fail if it
    can still disappear before use.
15. Printing App tokens, private keys, or secret values, or copying full CI
    logs / raw provider responses into evidence — forbidden.

## Contradictions and open questions

1. **Live caller workflow edit (`VOC-134-DEP-07`):** required behavior is
   settled (pin/fixture equal #166; both jobs restore shared policy; #166
   nested-checkout contract is present). Whether any live
   `.github/workflows/*.yml` file besides the fixture must change is proven at
   implementation time. Expected: no live `pipeline.yml` edit because the
   caller already calls `implement.yml@main` and `release.yml@main`.
2. **VOC-129 promotion timing:** hosted `release.yml@main` may already contain
   the #165 restore, and hosted `implement.yml@main` may already contain the
   #166 helper-lifetime repair, before this caller pin merges.
   `reconcile-release` for VOC-129 may therefore succeed independently of this
   package's merge. That does not remove this package's pin/test/doc
   obligation. Do not block VOC-129 promotion on this fixture sync if the live
   reusable workflow already has the repair, and do not treat VOC-129
   promotion as this package's implementation diff.
3. **Fixture pin applicability:** consumption of the restore and
   nested-checkout contracts is required. Pin to
   `f3d79177bf8a9abe0dae550f39502165d494c576`. If exact comparison proves
   another existing caller-fixture file differs from #166, record that as
   implementation evidence and mirror it; do not copy unrelated infra-only
   files merely to force a pin.
4. **Caller fixture README vs infrastructure README:** they are different
   documents. Update the caller fixture README's current-state pin paragraph.
   Do not overwrite it with `KARSIFT/karsift-ai-infra/README.md`.
5. **AGENTS.md / DOC-15 / DOC-16:** do not update those hashed or historical
   documents to name pin `f3d791…` if doing so would require VOC-112
   recapture. Put the current-state pin sentence in fixture README and this
   package's evidence. Historical correction notes stay unchanged.
6. **Infrastructure already merged:** this package must not snapshot, re-open,
   or rewrite #166. If later exact-revision review finds the caller cannot
   consume `f3d791…`, record that as a blocking finding rather than silently
   forking infrastructure inside the caller.
7. **VOC-132 exact pin:** VOC-132 remains an adopted #165 package whose
   implementation never published. This replacement consumes #166 under new
   authority rather than editing VOC-132's pin in place.
8. **VOC-133 exhausted carrier:** VOC-133 remains an adopted #166 package
   whose implementation exhausted. This replacement consumes #166 under new
   authority with a feasible evidence contract rather than editing VOC-133's
   D12/TEST-10 in place or redispatched #1063.
9. **PR #1051, PR #1056, and PR #1065:** the pin/restore portion of those
   unmerged carriers is not authority to cherry-pick the whole heads. A fresh
   carrier from current `develop` is required so VOC-112 JSON retargets, the
   provenance-test weakening, `package.json` provenance wrappers, capture-commit
   fetch helpers, and evidence-stamping helpers cannot ride along.
10. **Drafting-time SHA-256 values (`VOC-134-D11`):** hashes above were
    reconfirmed from an untracked nested checkout at `f3d791…`. Implementation
    must reconfirm them from a fetched exact merge. If they differ, record the
    fetched hashes in `t00-evidence.md` and use those as the committed
    expected values; do not silently keep a stale drafting-time hash.
11. **Non-directory classifier coverage:** the #166 helper rejects a
    non-directory nested path, but the mirrored
    `tests/test_voc121_implement_policy.py` at drafting time has no dedicated
    `nested_checkout_not_directory` test. Caller `tooling/governance/tests/`
    must cover that class without editing the mirrored fixture test bytes.
12. **Immutable carrier-base SHA (`VOC-134-D15`):** expected drafting-time
    develop is `95a779f9e62090f856ed03f389e7ac1d901aaa14`. If live `develop`
    has moved at implementation dispatch, that is unexpected; implementation
    fails closed unless the newly observed exact SHA is recorded as the
    pre-edit immutable constant before any in-scope edit. Do not guess a new
    base after files have already changed.
