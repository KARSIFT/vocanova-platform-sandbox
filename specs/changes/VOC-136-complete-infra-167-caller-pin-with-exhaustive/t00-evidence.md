# VOC-136-T00 — Evidence

Task: `VOC-136-T00` — Pin exact infra #167 from current develop with
exhaustive executable bypass scanning.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the protected comparison anchor, the
actual implementation PR base, the exact infra merge, hashes, scan scope,
negative cases, and validation. It does **not** require this file, in the
same Git commit, to contain that commit's own SHA. The live implementation
head is bound by the App-authored independent-review comment/check on the
implementation PR. Merge-gate must reject any mismatch. Post-merge /
root-issue audit may record the reviewed head and merge SHA.

## Discovery recorded at planning time (issue #1076)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1076 |
| Exhausted VOC-130 task | #1049 (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | #1051 |
| Exhausted VOC-131 carrier | #1056 |
| VOC-132 task that must not be redispatched | #1059 (`VOC-132-T00`) |
| Exhausted VOC-133 task | #1063 (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | #1065 |
| Exhausted VOC-134 task | #1068 (`VOC-134-T00`) |
| Exhausted VOC-134 carrier | #1070 |
| Adopted VOC-135 plan | #1072 |
| Exhausted VOC-135 task | #1073 (`VOC-135-T00`), closed not-planned with audit; no completion marker |
| Exhausted VOC-135 root | #1071, closed not-planned with audit |
| Exhausted VOC-135 carrier | #1075, closed/unmerged; remote branch deleted |
| VOC-135 attempt-1 FAIL head | `7c3e821bb3228c0fd2c7279b6d577a388148cb18` |
| VOC-135 attempt-1 defect | complete-diff scan matched its own assertion literal; 240 tests / 1 failure; independent FAIL https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444183121 |
| VOC-135 attempt-2 head | `dcd2da9ccbdf471c5294db30f09c9bb99aacdcd1` |
| VOC-135 attempt-2 defect | added `"tooling/governance/tests/"` to `SCAN_EXCLUDE_PREFIXES`; supervisor https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444311947 |
| VOC-135 attempt-2 review | PASS WITH NON-BLOCKING FINDINGS https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1075#issuecomment-5444342914; **not sufficient** for a replacement that retains the exclusion |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Authoritative infra merge | `b263c0c110591cc798b89277dfc35542abb1597b` (KARSIFT/karsift-ai-infra#167) |
| Independently reviewed PASS infra head | `eb11c4fc6841ec73816e2e064dcd449d98c1e933` |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| Protected comparison anchor | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| Issue-creation `develop` | `5044d62f6f35069412e8ffcf80a0682e7847d1c1` |
| Issue-creation `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Eight no-change paths | `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`; `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`; `scripts/foundation/voc112-navigation-benchmark.test.mjs`; `scripts/foundation/voc112-navigation-benchmark-run.mjs`; `scripts/foundation/validate-workspace.mjs`; `AGENTS.md`; `.agents/skills/vocanova-repo-navigator/SKILL.md`; `package.json` |
| Why VOC-135 cannot retry on #1073 / #1075 | both allowed attempts exhausted; attempt 2 weakened the required scan; operator closed unmerged; no completion marker |
| Why VOC-134 cannot retry on #1068 / #1070 | both allowed attempts exhausted; hydration bypass relocated from the runner to `validate-workspace.mjs` |
| Why VOC-133 cannot retry on #1063 / #1065 | both allowed attempts exhausted; provenance bypass plus Git-impossible self-SHA evidence |
| Why VOC-132 cannot retry on #1059 | VOC-132's adopted pin is exact #165; run `33079499176` created no publishable caller carrier |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained the VOC-112 JSON retargets |
| Why VOC-131 cannot retry on #1056 | both allowed attempts exhausted; provenance test still weakened; evidence claimed a revert the tree did not contain |
| Why VOC-129 is not retried | #1046 already merged; this is a pin plus PR-context / nested-checkout / restore consumption plus exhaustive scan plus feasible evidence |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-136 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-136 branch/PR from current `develop`; not #1051; not #1056; not #1065; not #1070; not #1075; not redispatched #1059; not redispatched #1063; not redispatched #1068; not redispatched #1073 |
| Infra | consume already-merged #167; do not open a replacement infra PR |
| Pin target | `b263c0c110591cc798b89277dfc35542abb1597b` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164), `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165), or `f3d79177bf8a9abe0dae550f39502165d494c576` (#166) |
| Mirrored #167 files | `.github/workflows/ci.yml`; `.github/workflows/implement.yml`; `config/run-app-checks.sh`; `tests/test_app_check_context.py` |
| Mirrored #166 files still required | `.github/workflows/release.yml`; `config/implementer_nested_checkout.py`; `tests/test_release_policy.py`; `tests/test_voc121_implement_policy.py`; `tests/test_voc123_source_bundle.py`; fixture `CHANGELOG.md`; expand only if exact comparison proves another existing caller-fixture file differs |
| Caller fixture README | update current-state pin paragraph only; do not replace with infrastructure repository README |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` + `persist-credentials: false` |
| Nested checkout | helpers copied before the unrestricted model; `absent` continues caller publication; symlink / non-directory / parent-Git inheritance fail closed; distinct nested Git checkout preserves exact-head bundle/ancestry/remote/lease |
| PR-context validation | `run-app-checks.sh` validates immutable PR base/head pair without fetching evidence; unchanged capture fixture → `pr-validation`; add/modify/delete → `pr-ancestry`; comparison error fail-closed; CI `fetch-depth: 0`; implementer pre-push uses integration anchor plus live HEAD including after self-correction |
| #167 byte identity | recorded SHA-256 hashes; no machine-specific `/tmp` checkout at test time |
| Drafting-time `ci.yml` SHA-256 | `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb` (reconfirm from fetched exact merge) |
| Drafting-time `implement.yml` SHA-256 | `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a` (reconfirm from fetched exact merge) |
| Drafting-time `release.yml` SHA-256 | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` (reconfirm from fetched exact merge) |
| Drafting-time `run-app-checks.sh` SHA-256 | `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2` (reconfirm from fetched exact merge) |
| Drafting-time `implementer_nested_checkout.py` SHA-256 | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` (reconfirm from fetched exact merge) |
| Drafting-time `test_app_check_context.py` SHA-256 | `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0` (reconfirm from fetched exact merge) |
| Drafting-time `test_release_policy.py` SHA-256 | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` (reconfirm from fetched exact merge) |
| Drafting-time `test_voc121_implement_policy.py` SHA-256 | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` (reconfirm from fetched exact merge) |
| Drafting-time `test_voc123_source_bundle.py` SHA-256 | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` (reconfirm from fetched exact merge) |
| Drafting-time `CHANGELOG.md` SHA-256 | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` (reconfirm from fetched exact merge) |
| Drafting-time infra README documentation-source SHA-256 | `06bcea7696ff80a59c9bab846061ce7a4c6d8bc175a1d02a2d7eefab0bcf46a3` (do not overwrite caller-owned fixture README) |
| Protected comparison anchor | `b9e74fc2db4691c48c637639b265d527de9f4505`; eight no-change paths compared to this SHA; do not use a moving branch ref after selection |
| Implementation PR base | record at PR creation before any in-scope edit; fail closed on unrelated/material movement of `develop`; this package's own plan/adoption/roster commits after the anchor do not count as protected-file drift |
| Eight no-change paths | byte-identical to the protected comparison anchor; JSON `subject_revision` remains `f9d11e23…`; provenance `local` mode stays fail-closed; none of the eight paths in the implementation diff against the anchor; fail—not skip—if the commit or a path cannot resolve |
| Exhaustive scan | enumerate every added/modified path against the protected comparison anchor; scan `scripts/**`, `package.json`, every added/changed `*.mjs`/`*.js`/`*.sh`/`*.py` except mirrored `tooling/governance/fixtures/karsift-ai-infra/**`, and any other newly executable caller file; **do not** exclude `tooling/governance/tests/**`, this regression's own module, or another executable directory wholesale |
| Source-safe literals | test/assertion literals that would match the scanner must be non-contiguous so a tracked committed module does not self-fail |
| Negative cases | import-time/pre-test git fetch; hydrate/materialize helper under an arbitrary filename; shell/Node/Python `VOC112_CAPTURE_PROVENANCE_MODE` setters; `PR_BASE_SHA`/`PR_HEAD_SHA` setters around tests or `validate-workspace`; import-time side effect in an added `tooling/governance/tests/*.py` module; wrapper/skip that makes default local fail-closed unreachable |
| Tracked pass | the regression must pass after it is tracked and committed, not merely while untracked |
| VOC-135 attempt-2 verdict | do not treat PASS WITH NON-BLOCKING FINDINGS as sufficient if the exclusion remains |
| Hashed VOC-112 sources | do not edit `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Provenance test | do not edit `scripts/foundation/voc112-navigation-benchmark.test.mjs` |
| Runner | do not edit `scripts/foundation/voc112-navigation-benchmark-run.mjs` |
| Workspace validator | do not edit `scripts/foundation/validate-workspace.mjs` |
| `package.json` | byte-identical to the protected comparison anchor |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head; merge-gate rejects mismatch; post-merge audit may name reviewed head and merge SHA; a commit must not be required to contain its own SHA |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; do not manufacture a VOC-129 completion marker; promote through repaired `release.yml@main` / `reconcile-release` |
| VOC-127 / VOC-130 / VOC-131 / VOC-132 / VOC-133 / VOC-134 / VOC-135 | do not retry; do not manufacture completion markers; close as superseded after this replacement promotes |
| Cleanup | obsolete remote automation branches/PRs/issues only; preserve unrelated VOC-128 and all user worktrees |
| Attempt | VOC-136-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged: implementer/escalation `cursor/composer-2.5`; planner/reviewer/reviewer_fast_retry/plan_reviewer `cursor/grok-4.6[effort=high,fast=false]` |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets; do not copy full CI logs |
| Evidence truth | name the protected comparison anchor, the implementation PR base, and the final infra merge; bind the live head via App-authored review/check; do not claim a protected-path revert unless that path is absent from the diff against the anchor; do not mutate this file at test/runtime |

## Changed surfaces (implementation)

Implementation PR base recorded before the first in-scope edit:
`a4f24e7de3ce62651ae25ca49adfff90256c1fa3` (current `develop` at dispatch).
Plan/adoption/roster commits after the protected comparison anchor are
governance-only and do not count as protected-file drift for the eight
no-change paths.

Protected comparison anchor for eight-path boundary:
`b9e74fc2db4691c48c637639b265d527de9f4505`.

Authoritative infra merge consumed:
`b263c0c110591cc798b89277dfc35542abb1597b` (KARSIFT/karsift-ai-infra#167).

Fixture files mirrored byte-for-byte from that merge:

- `.github/workflows/ci.yml` — SHA-256 `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb`
- `.github/workflows/implement.yml` — SHA-256 `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a`
- `.github/workflows/release.yml` — SHA-256 `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08`
- `config/run-app-checks.sh` — SHA-256 `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2`
- `config/implementer_nested_checkout.py` — SHA-256 `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9`
- `tests/test_app_check_context.py` — SHA-256 `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0`
- `tests/test_release_policy.py` — SHA-256 `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07`
- `tests/test_voc121_implement_policy.py` — SHA-256 `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975`
- `tests/test_voc123_source_bundle.py` — SHA-256 `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936`
- `CHANGELOG.md` — SHA-256 `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea`

`PINNED_SHA.txt` and live caller pin assertions advanced from `863fc1f…` to
`b263c0c…`. Caller-owned fixture `README.md` current-state paragraph updated
for VOC-136-T00 (#167 pin, post-caller-checkout restore, post-implementer
helper-lifetime, immutable PR-context contract). Infrastructure repository
`README.md` was not copied over the caller-owned fixture README.

New caller regressions:

- `tooling/governance/tests/voc136_bypass_scan.py` — exhaustive caller-diff
  bypass scanner with source-safe literals; only excludes mirrored
  `tooling/governance/fixtures/karsift-ai-infra/**` (does **not** exclude
  `tooling/governance/tests/**` or this regression module).
- `tooling/governance/tests/test_voc136_caller_replacement.py` — pin/hash,
  restore ordering, nested-checkout, PR-context, eight-path boundary against
  the protected comparison anchor, negative scanner cases, benign-mention
  non-false-positive, and feasible exact-revision evidence contract tests.

Eight no-change paths confirmed byte-identical to protected comparison anchor
`b9e74fc2…` and absent from `git diff` against that SHA. JSON
`subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
`package.json`, `validate-workspace.mjs`, and
`voc112-navigation-benchmark-run.mjs` remain anchor-identical. No live
`pipeline.yml` edit was required (`VOC-136-DEP-07`).

Complete-diff scan scope: every added/modified path against
`b9e74fc2db4691c48c637639b265d527de9f4505`, including all `scripts/**`,
`package.json`, every added/changed `*.mjs` / `*.js` / `*.sh` / `*.py`
outside the mirrored infra fixture subtree, and all changed paths under
`tooling/governance/tests/**`. Negative unit cases reject capture fetch,
hydrate/materialize helpers, provenance-mode overrides, PR SHA setters around
`validate-workspace`, import-time fetch side effects in caller tests, and
local fail-closed bypass wrappers. `SCAN_EXCLUDE_PREFIXES` does not list
`tooling/governance/tests/` (VOC-135 attempt-2 class rejected).

## Validation commands (implementation)

Recorded after the regression was prepared in the working tree (workflow
handoff stages/commits before hosted gates):

```bash
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
# 244 tests, OK

python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
# OK

python3 -m unittest discover -s tooling/governance/tests -p 'test_voc136*.py' -v
# 23 tests, OK (includes complete-diff scan, negative cases, benign mentions)

node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
# all pass

bash scripts/governance/validate-governance.sh \
  --base a4f24e7de3ce62651ae25ca49adfff90256c1fa3 --head HEAD
# passed

git diff --check
# clean
```

`classify-change-risk.sh` requires a staged diff range and will run on the
implementation PR with exact base `a4f24e7de3ce62651ae25ca49adfff90256c1fa3`
(expect R4). Hosted required checks and independent exact-revision review
remain pending on the implementation PR.

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
App-authored independent-review comment/check must bind the live PR head
exactly. Merge-gate must reject any mismatch. Record a pointer to that
comment/check here after it exists; do not write the live head SHA into this
file as a value that the same commit is required to contain. Confirm the
consumed infra merge is
`b263c0c110591cc798b89277dfc35542abb1597b`. The implementer must not approve
or merge its own work. Do not claim that any protected no-change path was
reverted unless `git diff` against the protected comparison anchor does not
name that path. Do not treat PR #1075 attempt-2 PASS WITH NON-BLOCKING
FINDINGS as sufficient if `tooling/governance/tests/` remains excluded from
the scan.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- Genuine staging runs only for the real caller tree change. Tree-equivalent
  post-promotion develop synchronization must not keep staging scheduled.
- VOC-129 skipped promotion completes through the repaired `release.yml@main`
  path or `reconcile-release` when a valid App-authored completion marker
  exists. Do not manufacture a VOC-129 completion marker.
- Exhausted / superseded VOC-127, VOC-130 (#1049 / unmerged #1051), VOC-131
  (unmerged #1056), unpublished VOC-132-T00 (#1059), exhausted VOC-133-T00
  (#1063 / unmerged #1065), exhausted VOC-134-T00 (#1068 / unmerged #1070),
  and exhausted VOC-135-T00 (#1073 / unmerged #1075) are closed as
  superseded after this replacement is promoted, with audit comments naming
  the VOC-136 merge. Do not manufacture completion markers. Do not merge
  those unmerged carriers. Do not redispatch those tasks.
- This package promotes through the same repaired path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close and ends 0 ahead / 0 behind `main` with an identical tree.
- Production deployment proceeds where existing path selection requires it
  and is verified if selected.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges for VOC-127 through VOC-135 and this replacement.
  Post-merge audit may record the independently reviewed head and the
  promotion merge SHA.
- Root issue #1076 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
- Cleanup removes only obsolete remote automation branches/PRs/issues.
  Preserve unrelated VOC-128 and all user worktrees.
