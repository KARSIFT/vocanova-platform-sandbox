# VOC-137-T00 — Evidence

Task: `VOC-137-T00` — Fail closed on PR base/head SHA overrides in every
scannable caller executable.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

Committed evidence in this file records the implementation PR base, the
scan-scope change, negative and positive scanner results, the pin freeze,
and validation. It does **not** require this file, in the same Git commit, to
contain that commit's own SHA. The live implementation head is bound by the
App-authored independent-review comment/check on the implementation PR.
Merge-gate must reject any mismatch. Post-merge / root-issue audit may record
the reviewed head and merge SHA.

Independent review must explicitly evaluate the arbitrary-filename
shell/Node/Python negative cases and benign controls before merge.

## Discovery recorded at planning time (issue #1083)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1083 |
| VOC-136 carrier (preserve, do not rewrite) | PR #1080 |
| VOC-136 exact reviewed head | `5d0c2350ab9a20ace586eaadd1169203140ffad0` |
| VOC-136 `develop` merge | `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` |
| Defect | `PR_SHA_SET_PATTERN` evaluated only when the relative filename contains `validate-workspace` or ends in `.test.mjs` |
| Counterexample | `scripts/arbitrary-wrapper.sh` with `export PR_BASE_SHA=deadbeef` then `pnpm test` is accepted |
| Pre-merge supervisor evidence | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444745877 |
| Reviewer verdict that missed the gap | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444825638 |
| Post-merge audit | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444861382 |
| Canceled release runs | `33113425829`, `33113547909` |
| Stopped release PR | #1082 converted to draft |
| Issue-creation `develop` | `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` |
| Issue-creation `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Live pin (already #167) | `b263c0c110591cc798b89277dfc35542abb1597b` |
| Protected comparison anchor (VOC-112 eight-path) | `b9e74fc2db4691c48c637639b265d527de9f4505` |
| VOC-112 `subject_revision` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Live contiguous literal that must become source-safe | `tooling/governance/tests/test_voc136_caller_replacement.py` fixture assertion `export PR_BASE_SHA=` |
| Live insufficient PR SHA negative case | `validate-workspace-wrapper.mjs` (filename still matches the defective heuristic) |
| Why VOC-136 is not retried on #1080 | already merged; records stay audit evidence |
| Why bootstrap is not required | T00's first run is attempt `1` on a new VOC-137 carrier from current `develop` |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-137 branch/PR from current `develop`; not a rewrite of PR #1080 |
| Infra | already pinned to #167; do not open a replacement infra PR; do not re-mirror |
| Pin target (frozen) | `b263c0c110591cc798b89277dfc35542abb1597b` |
| Scanner change | remove filename gate around `PR_SHA_SET_PATTERN`; apply to every scannable caller executable outside the fixture mirror |
| Source-safe construction | rebuild pattern from `_PR_BASE_ENV` / `_PR_HEAD_ENV`; source-safe the VOC-136 contiguous fixture assertion |
| Negative cases | `scripts/arbitrary-wrapper.sh` shell export before `pnpm test`; Node `process.env.PR_HEAD_SHA` under an arbitrary name; Python `os.environ["PR_BASE_SHA"]` under an arbitrary name; added/modified `*.py` outside the fixture mirror |
| Positive controls | benign discussion; pattern construction; excluded fixture `run-app-checks.sh` |
| Wholesale exclusion | **forbidden** for `tooling/governance/tests/`, this module, `scripts/`, or another executable directory |
| Fixture freeze | no path under `tooling/governance/fixtures/karsift-ai-infra/` in the implementation diff |
| Eight no-change paths | byte-identical to `b9e74fc2…`; do not edit `AGENTS.md`, navigator skill, VOC-112 fixtures, provenance test, runner, `validate-workspace.mjs`, or `package.json` |
| Exact-head binding | App-authored independent-review comment/check binds the live PR head and explicitly evaluates the arbitrary-filename cases; merge-gate rejects mismatch; a commit must not be required to contain its own SHA |
| VOC-136 records | preserve; do not rewrite `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/` |
| Attempt | VOC-137-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged: implementer/escalation `cursor/composer-2.5`; planner/reviewer/reviewer_fast_retry/plan_reviewer `cursor/grok-4.6[effort=high,fast=false]` |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not copy full CI logs |
| Evidence truth | name the implementation PR base; bind the live head via App-authored review/check; do not mutate this file at test/runtime |

## Changed surfaces (implementation)

Implementation PR base recorded before the first in-scope edit:
`pending-at-implementation-dispatch` (resolve current `develop` to a
40-character SHA before any in-scope edit; issue-creation develop was
`0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`). Plan/adoption/roster commits
after that SHA are governance-only and do not count as protected-file drift.

Protected comparison anchor for eight-path VOC-112 boundary:
`b9e74fc2db4691c48c637639b265d527de9f4505`.

Authoritative infra merge already pinned (do not change):
`b263c0c110591cc798b89277dfc35542abb1597b`.

VOC-136-D11 fixture hashes (must remain; do not re-mirror):

- `.github/workflows/ci.yml` — `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb`
- `.github/workflows/implement.yml` — `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a`
- `.github/workflows/release.yml` — `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08`
- `config/run-app-checks.sh` — `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2`
- `config/implementer_nested_checkout.py` — `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9`
- `tests/test_app_check_context.py` — `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0`
- `tests/test_release_policy.py` — `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07`
- `tests/test_voc121_implement_policy.py` — `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975`
- `tests/test_voc123_source_bundle.py` — `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936`
- `CHANGELOG.md` — `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea`

Expected in-scope paths (implementation fills actual diff names):

- `tooling/governance/tests/voc136_bypass_scan.py`
- `tooling/governance/tests/test_voc136_caller_replacement.py` (source-safe
  literal and/or additional cases)
- optional `tooling/governance/tests/test_voc137_*.py`
- this package directory, including this file

Complete-diff scan scope after this correction: every added/modified
scannable caller executable against the implementation range, including all
`scripts/**`, `package.json`, every added/changed `*.mjs` / `*.js` / `*.sh`
/ `*.py` outside the mirrored infra fixture subtree, and all changed paths
under `tooling/governance/tests/**`. PR base/head SHA assignment fails
closed regardless of filename. `SCAN_EXCLUDE_PREFIXES` does not list
`tooling/governance/tests/`.

## Validation commands (implementation)

Pending. Record after the regression is tracked and committed, including
exact base/head for governance validation and classification, unittest
discovery for caller and mirrored infra suites, targeted VOC-136/VOC-137
scanner cases, `git diff --check`, and the direct reproduction that
`scan_changed_path_for_bypasses("scripts/arbitrary-wrapper.sh", payload)`
raises. Do not treat an untracked-only pass as acceptance. Expect
`classify-change-risk.sh` to report R4.

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
App-authored independent-review comment/check must bind the live PR head
exactly and must explicitly evaluate the arbitrary-filename shell/Node/Python
negative cases and benign controls. Merge-gate must reject any mismatch.
Record a pointer to that comment/check here after it exists; do not write the
live head SHA into this file as a value that the same commit is required to
contain. The implementer must not approve or merge its own work. Do not treat
PR #1080's review as sufficient for this correction.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- Genuine staging runs only for the real caller tree change. Tree-equivalent
  post-promotion develop synchronization must not keep staging scheduled.
- Ordinary release evaluation (or `reconcile-release`) promotes outstanding
  completed packages, including already-merged VOC-136 and this package.
  Do not manufacture a VOC-136 completion marker. Do not restage draft
  release PR #1082 as a substitute.
- `develop` is advanced to each successful promotion merge SHA before audit
  close and ends 0 ahead / 0 behind `main` with an identical tree.
- Production deployment proceeds where existing path selection requires it
  and is verified if selected.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges. Post-merge audit may record the independently reviewed
  head and the promotion merge SHA.
- Root issue #1083 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
- Canceled release runs `33113425829` and `33113547909` remain audit context
  only.
