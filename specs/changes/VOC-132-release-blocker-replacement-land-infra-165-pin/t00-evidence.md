# VOC-132-T00 — Evidence

Task: `VOC-132-T00` — Pin exact infra #165 from current develop with a
complete VOC-112 no-change boundary.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1057)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1057 |
| Exhausted VOC-130 task | #1049 (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | #1051 |
| Exhausted VOC-131 carrier | #1056 |
| VOC-131 exhausted head | `c11454e717a6d778143de1f2023acc4480305845` |
| VOC-131 exact review | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439802080 |
| VOC-131 supervisor evidence | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1056#issuecomment-5439768013 |
| VOC-131 reproduction merge-base | `790ea1b66b096414fe6709e6cd4b8342ff5ac587` |
| VOC-131 exhaustion cause | `scripts/foundation/voc112-navigation-benchmark.test.mjs` weakened missing-subject `local` mode in a full checkout; evidence falsely claimed that edit was reverted |
| VOC-130 exhaustion cause | both reviewed #1051 heads retained unrelated VOC-112 JSON evidence retargets |
| Shared root cause of exhausted replacements | deterministic guards protected only the two VOC-112 JSON paths, so the adjacent provenance test could be edited |
| VOC-129 caller PR | #1046 at `429d8c6d49303148ca1cc14dba5f6768a7863346` |
| Original failing/no-op release wake-up | `33066533397` |
| Selected checkout ref | `develop` |
| Observed failure | `karsift-ai-infra/config/task-completion-runner.py` missing after caller root checkout |
| Observed release result | missing validator treated as safe no-op; no audit/promotion; converge skipped |
| Root cause of the release no-op | root caller checkout deletes nested shared policy; restore must follow caller checkout before lifecycle helpers |
| Authoritative infra merge | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (KARSIFT/karsift-ai-infra#165) |
| Independently reviewed infra head | `e33931d02f7bdbb094ae8177fd88324cd19ac5ce` |
| Current `develop` pin at drafting | `863fc1f35b1d35e4981a59166b0e939be1a2b681` (infrastructure #164) |
| VOC-112 `subject_revision` at current `develop` | `f9d11e232a07c7d7a9c433d02c9267912543ba10` |
| Complete VOC-112 no-change paths | `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`; `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`; `scripts/foundation/voc112-navigation-benchmark.test.mjs`; `AGENTS.md`; `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained the VOC-112 JSON retargets |
| Why VOC-131 cannot retry on #1056 | both allowed attempts exhausted; provenance test still weakened; evidence claimed a revert the tree did not contain |
| Why VOC-129 is not retried | #1046 already merged; this is a checkout-lifetime pin plus a complete VOC-112 no-change boundary |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-132 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-132 branch/PR from current `develop`; not #1051; not #1056 |
| In-scope reuse of #1056 | pin, mirrored #165 files, restore coverage, pin-literal updates only; recreate on the new carrier |
| Infra | consume already-merged #165; do not open a replacement infra PR |
| Pin target | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Mirrored #165 files | expected `.github/workflows/release.yml` and `tests/test_release_policy.py` only, byte-for-byte; expand only if exact comparison proves another required #165 delta |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` + `persist-credentials: false` |
| #165 byte identity | recorded SHA-256 hashes; no machine-specific `/tmp` checkout at test time |
| Drafting-time `release.yml` SHA-256 | `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` (reconfirm from fetched exact merge) |
| Drafting-time `test_release_policy.py` SHA-256 | `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` (reconfirm from fetched exact merge) |
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
| Attempt | VOC-132-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |
| Evidence truth | name the exact reviewed head; do not claim a protected-path revert unless that path is absent from the diff |

## Changed surfaces (implementation)

Pending implementation. Record the exact fixture files mirrored from
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (expected:
`.github/workflows/release.yml` and `tests/test_release_policy.py` only),
the reconfirmed SHA-256 hashes, every live pin assertion advanced from
`863fc1f…`, confirmation that all five VOC-112 no-change paths are absent
from the diff, any live workflow edit required by `VOC-132-DEP-07`, and the
commands actually run.

## Validation commands (implementation)

Pending implementation. Expected commands are listed in
`implementation-plan.md`. Record exact commands and results here; do not
treat a missing suite as a pass. Do not treat a `/tmp` karsift-ai-infra
checkout as a required comparison source.

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. Record
the exact reviewed head SHA here. The implementer must not approve or merge
its own work. Do not claim that
`scripts/foundation/voc112-navigation-benchmark.test.mjs` or any other
protected path was reverted unless `git diff` against the carrier `develop`
base does not name that path.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists. Do not manufacture a VOC-129
  completion marker.
- Exhausted VOC-130 (#1049 / unmerged #1051) is closed as superseded after
  this replacement is promoted, with audit comments naming the VOC-132 merge.
  Do not manufacture a VOC-130 completion marker. Do not merge #1051.
- Exhausted VOC-131 (unmerged #1056 at `c11454e…`) is closed as superseded
  after this replacement is promoted, with audit comments naming the VOC-132
  merge. Do not manufacture a VOC-131 completion marker. Do not merge #1056.
- This package promotes through the same repaired path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges for VOC-129, VOC-130, VOC-131, and this replacement.
- Root issue #1057 closes only after that audit evidence exists.
- Closed state alone is not completion proof.
