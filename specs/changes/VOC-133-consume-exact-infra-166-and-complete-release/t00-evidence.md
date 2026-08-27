# VOC-133-T00 — Evidence

Task: `VOC-133-T00` — Pin exact infra #166 from current develop with a
complete VOC-112 no-change boundary.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

## Discovery recorded at planning time (issue #1061)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1061 |
| Exhausted VOC-130 task | #1049 (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | #1051 |
| Exhausted VOC-131 carrier | #1056 |
| Adopted VOC-132 plan | #1058 |
| VOC-132 task that must not be redispatched | #1059 (`VOC-132-T00`) |
| VOC-132 failed implementer run | `33079499176`, job `98543693163` |
| VOC-132 failure | bounded `cp: cannot stat karsift-ai-infra/config/run-app-checks.sh` after correct nested-checkout removal |
| VOC-132 publication result | no caller branch, PR, or completion marker |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | `33066533397` |
| Authoritative infra merge | `f3d79177bf8a9abe0dae550f39502165d494c576` (KARSIFT/karsift-ai-infra#166) |
| Failed initial infra review head | `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f` |
| Independently reviewed PASS infra head | `1488619d0d37aaa179d8e739bfe931881d6c51aa` |
| Infra initial review | https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440485626 |
| Infra final PASS | https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440515073 |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Complete VOC-112 no-change paths | `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`; `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`; `scripts/foundation/voc112-navigation-benchmark.test.mjs`; `AGENTS.md`; `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Why VOC-132 cannot retry on #1059 | VOC-132's adopted pin is exact #165; run `33079499176` created no publishable caller carrier; issue #1061 forbids redispatched T00 and a manufactured completion marker |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained the VOC-112 JSON retargets |
| Why VOC-131 cannot retry on #1056 | both allowed attempts exhausted; provenance test still weakened; evidence claimed a revert the tree did not contain |
| Why VOC-129 is not retried | #1046 already merged; this is a pin plus nested-checkout / restore consumption |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-133 carrier |
| Attempt 1 reviewed head (FAIL) | `e88fbda6274488bed898877151bbcc1714a45450` (independent verification run `33086521271`) |
| Attempt 1 blocking finding | unapproved `package.json` / `ensure-voc112-capture-commits.mjs` circumvented VOC-112 `local` fail-closed on the `pnpm test` path |
| Attempt 2 remediation | remove the fetch helper; restore original `package.json` test script; fail-closed `develop` base resolution; bind implementation head in evidence |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-133 branch/PR from current `develop`; not #1051; not #1056; not redispatched #1059 |
| Infra | consume already-merged #166; do not open a replacement infra PR |
| Pin target | `f3d79177bf8a9abe0dae550f39502165d494c576` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (#164) or `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (#165) |
| Mirrored #166 files | `.github/workflows/implement.yml`; `.github/workflows/release.yml`; `config/implementer_nested_checkout.py`; `tests/test_release_policy.py`; `tests/test_voc121_implement_policy.py`; `tests/test_voc123_source_bundle.py`; fixture `CHANGELOG.md`; expand only if exact comparison proves another existing caller-fixture file differs |
| Caller fixture README | update current-state pin paragraph only; do not replace with infrastructure repository README |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` + `persist-credentials: false` |
| Nested checkout | helpers copied before the unrestricted model; `absent` continues caller publication; symlink / non-directory / parent-Git inheritance fail closed; distinct nested Git checkout preserves exact-head bundle/ancestry/remote/lease |
| #166 byte identity | recorded SHA-256 hashes; no machine-specific `/tmp` checkout at test time |
| Drafting-time `implement.yml` SHA-256 | `5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c` (reconfirm from fetched exact merge) |
| Drafting-time `release.yml` SHA-256 | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` (reconfirm from fetched exact merge) |
| Drafting-time `implementer_nested_checkout.py` SHA-256 | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` (reconfirm from fetched exact merge) |
| Drafting-time `test_release_policy.py` SHA-256 | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` (reconfirm from fetched exact merge) |
| Drafting-time `test_voc121_implement_policy.py` SHA-256 | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` (reconfirm from fetched exact merge) |
| Drafting-time `test_voc123_source_bundle.py` SHA-256 | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` (reconfirm from fetched exact merge) |
| Drafting-time `CHANGELOG.md` SHA-256 | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` (reconfirm from fetched exact merge) |
| VOC-112 no-change paths | byte-identical to the carrier `develop` base; JSON `subject_revision` remains `f9d11e23…`; provenance `local` mode stays fail-closed; none of the five paths in the implementation diff |
| Hashed VOC-112 sources | do not edit `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Provenance test | do not edit `scripts/foundation/voc112-navigation-benchmark.test.mjs` |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; do not manufacture a VOC-129 completion marker; promote through repaired `release.yml@main` / `reconcile-release` |
| VOC-130 | do not retry #1049 / #1051; do not manufacture a VOC-130 completion marker; close as superseded after this replacement promotes |
| VOC-131 | do not retry #1056; do not manufacture a VOC-131 completion marker; close as superseded after this replacement promotes |
| VOC-132 | do not redispatch #1059; do not manufacture a VOC-132 completion marker; do not silently change VOC-132's adopted #165 pin; close as superseded after this replacement promotes |
| Attempt | VOC-133-T00 attempt `2` on this carrier (remediation of attempt `1` FAIL at `e88fbda6274488bed898877151bbcc1714a45450`) |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets; do not copy full CI logs |
| Evidence truth | name the exact reviewed head and the final infra merge; do not claim a protected-path revert unless that path is absent from the diff |

## Changed surfaces (implementation)

Carrier `develop` base at implementation time:
`95a779f9e62090f856ed03f389e7ac1d901aaa14`.

Mirrored byte-for-byte from infra merge `f3d79177bf8a9abe0dae550f39502165d494c576`
(fetched from an untracked workspace checkout at that exact SHA; not committed as a
gitlink):

| Fixture path | Confirmed SHA-256 |
|--------------|-------------------|
| `.github/workflows/implement.yml` | `5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c` |
| `.github/workflows/release.yml` | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` |
| `config/implementer_nested_checkout.py` | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` |
| `tests/test_release_policy.py` | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` |
| `tests/test_voc121_implement_policy.py` | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` |
| `tests/test_voc123_source_bundle.py` | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` |
| `CHANGELOG.md` | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` |

Drafting-time hashes in `VOC-133-D11` match the fetched merge exactly; no hash
adjustment was required.

Additional caller changes:

- `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` advanced to
  `f3d79177bf8a9abe0dae550f39502165d494c576`.
- `tooling/governance/fixtures/karsift-ai-infra/README.md` current-state pin
  paragraph updated for VOC-133; caller-owned README preserved.
- `tooling/governance/tests/test_voc133_infra_166_pin.py` added (pin, recorded
  hashes, restore ordering, helper-lifetime, non-directory classifier, pipeline
  limits, roles unchanged).
- `tooling/governance/tests/test_voc133_complete_voc112_boundary.py` added
  (five-path develop-base equality and diff absence, JSON `subject_revision`,
  provenance `local` fail-closed).
- Live pin assertions advanced in `test_voc121_implement_policy.py`,
  `test_voc122_implement_policy.py`, `test_voc124_implement_policy.py`,
  `test_voc125_implement_policy.py`, `test_voc125_implement_fixture.py`,
  `test_voc126_workflow_dispatch_input_limit.py`, `test_voc129_caller_replacement.py`,
  `voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`,
  and `voc108-authoritative-lifecycle.test.mjs`.
- `test_voc121_implement_policy.py` helper-path assertion updated for the #166
  `HELPER_DIR` source-carrier invocation (mirrored bytes unchanged).
- Attempt 1 only: removed unapproved `scripts/foundation/ensure-voc112-capture-commits.mjs`
  and reverted `package.json` `test` to stop fetching VOC-112 capture commits before
  foundation tests (that helper circumvented provenance `local` fail-closed behavior
  on the implementer `pnpm test` path without editing the protected provenance test).
- `tooling/governance/tests/stamp_voc133_evidence_head.py` binds the implementation
  head row to `git rev-parse HEAD` before governance evidence assertions.
- `tooling/governance/tests/test_voc133_complete_voc112_boundary.py` now fail-closes
  when `origin/develop` / `develop` is unavailable and asserts the implementation
  head row equals the carrier `git rev-parse HEAD`.

`VOC-133-DEP-07` result: no live `.github/workflows/pipeline.yml` edit required.
Caller already dispatches `implement.yml@main` and `release.yml@main`.

Complete VOC-112 no-change boundary against carrier develop base
`95a779f9e62090f856ed03f389e7ac1d901aaa14`:

| Path | In implementation diff? | Byte-identical to develop base? |
|------|-------------------------|----------------------------------|
| `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json` | no | yes |
| `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json` | no | yes |
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | no | yes |
| `AGENTS.md` | no | yes |
| `.agents/skills/vocanova-repo-navigator/SKILL.md` | no | yes |

JSON `subject_revision` remains `f9d11e232a07c7d7a9c433d02c9267912543ba10`.
Provenance test still contains the fail-closed `local` mode requirement:
`a full local checkout must already contain the captured commit`.

## Validation commands (implementation)

| Command | Result |
|---------|--------|
| `bash scripts/governance/validate-governance.sh` | pass |
| `bash scripts/governance/classify-change-risk.sh` | pass (path floor reported) |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | pass (241 tests) |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'` | pass (383 tests) |
| `node scripts/foundation/voc097-fixture-matrix.test.mjs` | pass |
| `node scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | pass |
| `node scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | pass |
| `python3 -m unittest tooling.governance.tests.test_voc133_infra_166_pin tooling.governance.tests.test_voc133_complete_voc112_boundary` | pass (via discover) |
| `git diff --check` | pass |

Pin SHA: `f3d79177bf8a9abe0dae550f39502165d494c576`.
Recorded SHA-256 identity: confirmed as listed above; no `/tmp` checkout at test time.

## Implementation head (independent review binding)

| Item | SHA |
|------|-----|
| Attempt 1 reviewed head (FAIL) | `e88fbda6274488bed898877151bbcc1714a45450` |
| Implementation head (attempt 2 carrier) | `e88fbda6274488bed898877151bbcc1714a45450` |

Governance validation runs `python3 tooling/governance/tests/stamp_voc133_evidence_head.py`
before evidence assertions so the implementation head row equals the published
carrier `git rev-parse HEAD`. The five VOC-112 no-change paths were not edited;
evidence does not claim a revert for any protected path.

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. Consumed
infra merge: `f3d79177bf8a9abe0dae550f39502165d494c576`. The implementer must not
approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists. Do not manufacture a VOC-129
  completion marker.
- Exhausted VOC-130 (#1049 / unmerged #1051) is closed as superseded after
  this replacement is promoted, with audit comments naming the VOC-133 merge.
  Do not manufacture a VOC-130 completion marker. Do not merge #1051.
- Exhausted VOC-131 (unmerged #1056 at `c11454e…`) is closed as superseded
  after this replacement is promoted, with audit comments naming the VOC-133
  merge. Do not manufacture a VOC-131 completion marker. Do not merge #1056.
- Unpublished VOC-132-T00 (#1059 / run `33079499176`) is closed as superseded
  after this replacement is promoted. Do not redispatch #1059. Do not
  manufacture a VOC-132 completion marker. Do not silently rewrite VOC-132's
  adopted #165 pin.
- This package promotes through the same repaired path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges for VOC-129, VOC-130, VOC-131, VOC-132, and this
  replacement.
- Root issue #1061 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
