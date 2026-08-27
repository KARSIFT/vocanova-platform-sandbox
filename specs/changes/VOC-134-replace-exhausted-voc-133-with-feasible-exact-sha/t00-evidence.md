# VOC-134-T00 — Evidence

Task: `VOC-134-T00` — Pin exact infra #166 from current develop with feasible
exact-SHA evidence and a complete VOC-112 no-change boundary.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the immutable carrier base, the exact
infra merge, hashes, and validation. It does **not** require this file, in
the same Git commit, to contain that commit's own SHA. The live
implementation head is bound by the App-authored independent-review
comment/check on the implementation PR. Merge-gate must reject any mismatch.
Post-merge / root-issue audit may record the reviewed head and merge SHA.

## Discovery recorded at planning time (issue #1066)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1066 |
| Exhausted VOC-130 task | #1049 (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | #1051 |
| Exhausted VOC-131 carrier | #1056 |
| Adopted VOC-132 plan | #1058 |
| VOC-132 task that must not be redispatched | #1059 (`VOC-132-T00`) |
| VOC-132 failed implementer run | `33079499176`, job `98543693163` |
| VOC-132 failure | bounded `cp: cannot stat karsift-ai-infra/config/run-app-checks.sh` after correct nested-checkout removal |
| VOC-132 publication result | no caller branch, PR, or completion marker |
| Adopted VOC-133 plan | #1062 |
| Exhausted VOC-133 task | #1063 (`VOC-133-T00`) |
| Exhausted VOC-133 carrier | #1065 |
| VOC-133 attempt-1 FAIL head | `e88fbda6274488bed898877151bbcc1714a45450` |
| VOC-133 attempt-1 FAIL | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441321518 |
| VOC-133 attempt-1 supervisor finding | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441214837 |
| VOC-133 attempt-2 FAIL head | `70930cf0572f73f254429a240c5328640846e0a9` |
| VOC-133 attempt-2 FAIL | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1065#issuecomment-5441659390 |
| VOC-133 attempt-1 defect | added `package.json` plus `scripts/foundation/ensure-voc112-capture-commits.mjs`; fetched a missing VOC-112 capture commit before `pnpm test`; required full-checkout `local` fail-closed behavior became unreachable |
| VOC-133 attempt-2 defect | removed the helper but changed `package.json` to force `VOC112_CAPTURE_PROVENANCE_MODE=pr-validation` with both PR SHAs set to `HEAD`; added a mutable evidence-stamping helper and in-memory rewrite so tests could pass while committed evidence named the prior failed head |
| VOC-133 specification defect | VOC-133-D12 / TEST-10 as interpreted by those attempts required a tracked file to contain the SHA of the same commit; that is Git-impossible |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | `33066533397` |
| Authoritative infra merge | `f3d79177bf8a9abe0dae550f39502165d494c576` (KARSIFT/karsift-ai-infra#166) |
| Failed initial infra review head | `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f` |
| Independently reviewed PASS infra head | `1488619d0d37aaa179d8e739bfe931881d6c51aa` |
| Infra initial review | https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440485626 |
| Infra final PASS | https://github.com/KARSIFT/karsift-ai-infra/pull/166#issuecomment-5440515073 |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| Expected immutable carrier base at drafting | `95a779f9e62090f856ed03f389e7ac1d901aaa14` |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Complete VOC-112 no-change paths | `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`; `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`; `scripts/foundation/voc112-navigation-benchmark.test.mjs`; `AGENTS.md`; `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Additional no-change path | `package.json` |
| Why VOC-133 cannot retry on #1063 / #1065 | both allowed attempts exhausted; provenance bypass plus Git-impossible self-SHA evidence |
| Why VOC-132 cannot retry on #1059 | VOC-132's adopted pin is exact #165; run `33079499176` created no publishable caller carrier; issue #1066 forbids redispatched T00 and a manufactured completion marker |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained the VOC-112 JSON retargets |
| Why VOC-131 cannot retry on #1056 | both allowed attempts exhausted; provenance test still weakened; evidence claimed a revert the tree did not contain |
| Why VOC-129 is not retried | #1046 already merged; this is a pin plus nested-checkout / restore consumption plus feasible evidence |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-134 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-134 branch/PR from current `develop`; not #1051; not #1056; not #1065; not redispatched #1059; not redispatched #1063 |
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
| Immutable carrier-base SHA | expected `95a779f9e62090f856ed03f389e7ac1d901aaa14`; revalidate at implementation dispatch before any in-scope edit; fail closed if `develop` moved unexpectedly until the newly observed exact SHA is recorded as the pre-edit constant; do not use a moving branch ref after selection |
| VOC-112 no-change paths | byte-identical to the immutable carrier-base SHA; JSON `subject_revision` remains `f9d11e23…`; provenance `local` mode stays fail-closed; none of the five paths in the implementation diff; fail—not skip—if the commit or a path cannot resolve |
| `package.json` | byte-identical to the immutable carrier-base SHA; no capture-commit fetch helper; no provenance-mode wrapper; no evidence-stamping helper; no test-time evidence mutation |
| Hashed VOC-112 sources | do not edit `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Provenance test | do not edit `scripts/foundation/voc112-navigation-benchmark.test.mjs` |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head; merge-gate rejects mismatch; post-merge audit may name reviewed head and merge SHA; a commit must not be required to contain its own SHA |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; do not manufacture a VOC-129 completion marker; promote through repaired `release.yml@main` / `reconcile-release` |
| VOC-130 | do not retry #1049 / #1051; do not manufacture a VOC-130 completion marker; close as superseded after this replacement promotes |
| VOC-131 | do not retry #1056; do not manufacture a VOC-131 completion marker; close as superseded after this replacement promotes |
| VOC-132 | do not redispatch #1059; do not manufacture a VOC-132 completion marker; do not silently change VOC-132's adopted #165 pin; close as superseded after this replacement promotes |
| VOC-133 | do not redispatch #1063; do not reuse #1065; do not manufacture a VOC-133 completion marker; do not silently rewrite VOC-133 in place; close as superseded after this replacement promotes |
| Attempt | VOC-134-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets; do not copy full CI logs |
| Evidence truth | name the immutable carrier base and the final infra merge; bind the live head via App-authored review/check; do not claim a protected-path revert unless that path is absent from the diff; do not mutate this file at test/runtime |

## Changed surfaces (implementation)

Implementation resolved current `develop` to
`b9e74fc2db4691c48c637639b265d527de9f4505` before any in-scope edit. That SHA
is the immutable carrier-base constant for VOC-112 no-change and `package.json`
identity checks (drafting expected `95a779f9e62090f856ed03f389e7ac1d901aaa14`;
`develop` moved after adoption roster merge #1069).

Mirrored byte-for-byte from infrastructure merge
`f3d79177bf8a9abe0dae550f39502165d494c576` (#166):

- `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml`
- `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml`
- `tooling/governance/fixtures/karsift-ai-infra/config/implementer_nested_checkout.py`
  (new)
- `tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py`
- `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc121_implement_policy.py`
- `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc123_source_bundle.py`
- `tooling/governance/fixtures/karsift-ai-infra/CHANGELOG.md`
- `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`

Reconfirmed SHA-256 hashes at implementation (match `VOC-134-D11` drafting-time
values):

| Fixture path | SHA-256 |
|--------------|---------|
| `.github/workflows/implement.yml` | `5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c` |
| `.github/workflows/release.yml` | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` |
| `config/implementer_nested_checkout.py` | `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` |
| `tests/test_release_policy.py` | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` |
| `tests/test_voc121_implement_policy.py` | `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` |
| `tests/test_voc123_source_bundle.py` | `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` |
| `CHANGELOG.md` | `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` |

Advanced live pin assertions from `863fc1f…` (#164) to `f3d791…` (#166) in
`scripts/foundation/voc097-fixture-matrix.test.mjs`,
`voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs`,
and matching `tooling/governance/tests/test_voc12*.py` suites. Added
`tooling/governance/tests/test_voc134_caller_replacement.py` for restore
ordering, #166 helper-lifetime / nested-checkout coverage (including
`nested_checkout_not_directory`), recorded-hash byte identity, immutable
carrier-base VOC-112 boundary, `package.json` identity, and feasible
exact-revision evidence contract.

Updated current-state pin paragraph in fixture `README.md`. No live
`.github/workflows/pipeline.yml` edit (`VOC-134-DEP-07`: caller already uses
`implement.yml@main` and `release.yml@main`).

Protected-path equality against immutable carrier base
`b9e74fc2db4691c48c637639b265d527de9f4505`:

- all five VOC-112 no-change paths byte-identical; absent from
  `git diff b9e74fc2db4691c48c637639b265d527de9f4505`
- `package.json` byte-identical; absent from that diff
- no capture-commit fetch helper, provenance-mode wrapper, or evidence-stamping
  helper added

## Validation commands (implementation)

Recorded after implementation (exact commands and exit codes):

```text
python3 -m unittest discover -s tooling/governance/tests -p 'test_voc134*.py'  → exit 0 (15 tests)
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'  → exit 0 (236 tests)
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'  → exit 0
node scripts/foundation/voc097-fixture-matrix.test.mjs  → exit 0
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs  → exit 0
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs  → exit 0
bash scripts/governance/validate-governance.sh  → exit 0
bash scripts/governance/classify-change-risk.sh  → exit 0 (path floor R4)
git diff --check  → exit 0
```

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
App-authored independent-review comment/check must bind the live PR head
exactly. Merge-gate must reject any mismatch. Record a pointer to that
comment/check here after it exists; do not write the live head SHA into this
file as a value that the same commit is required to contain. Confirm the
consumed infra merge is
`f3d79177bf8a9abe0dae550f39502165d494c576`. The implementer must not approve
or merge its own work. Do not claim that
`scripts/foundation/voc112-navigation-benchmark.test.mjs`, `package.json`, or
any other protected path was reverted unless `git diff` against the immutable
carrier-base SHA does not name that path.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists. Do not manufacture a VOC-129
  completion marker.
- Exhausted VOC-130 (#1049 / unmerged #1051) is closed as superseded after
  this replacement is promoted, with audit comments naming the VOC-134 merge.
  Do not manufacture a VOC-130 completion marker. Do not merge #1051.
- Exhausted VOC-131 (unmerged #1056 at `c11454e…`) is closed as superseded
  after this replacement is promoted, with audit comments naming the VOC-134
  merge. Do not manufacture a VOC-131 completion marker. Do not merge #1056.
- Unpublished VOC-132-T00 (#1059 / run `33079499176`) is closed as superseded
  after this replacement is promoted. Do not redispatch #1059. Do not
  manufacture a VOC-132 completion marker. Do not silently rewrite VOC-132's
  adopted #165 pin.
- Exhausted VOC-133-T00 (#1063 / unmerged #1065 at `e88fbda…` / `70930cf…`)
  is closed as superseded after this replacement is promoted. Do not
  redispatch #1063. Do not reuse or merge #1065. Do not manufacture a VOC-133
  completion marker. Do not silently rewrite VOC-133 in place.
- This package promotes through the same repaired path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Tree-equivalent sync must not keep staging scheduled. Production deployment
  proceeds where existing path selection requires it.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges for VOC-129 through VOC-133 and this replacement. Post-merge
  audit may record the independently reviewed head and the promotion merge SHA.
- Root issue #1066 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
