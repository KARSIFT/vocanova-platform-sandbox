# VOC-131-T00 — Evidence

Task: `VOC-131-T00` — Pin exact infra #165 from current develop without
rewriting VOC-112 evidence.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1052)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1052 |
| Exhausted VOC-130 task | #1049 (`VOC-130-T00`) |
| Unmerged VOC-130 carrier | #1051 |
| Attempt-1 reviewed head | `a04a41a19176d1322faee1a3365d82153d61af1b` |
| Attempt-1 review | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5438961796 |
| Attempt-2 reviewed head | `e846cc2f62556c2d6616282f2e9c4929c00655e7` |
| Attempt-2 review | https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1051#issuecomment-5439141140 |
| Durable exhaustion marker | https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1049#issuecomment-5439143793 |
| VOC-130 exhaustion cause | both reviewed heads retained unrelated VOC-112 evidence retargets |
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
| Out-of-scope VOC-112 files | `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`; `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json` |
| Why VOC-130 cannot retry on #1051 | both allowed attempts exhausted; both reviewed heads retained the VOC-112 retargets |
| Why VOC-129 is not retried | #1046 already merged; this is a checkout-lifetime pin plus a VOC-112 identity constraint |
| Why bootstrap is not required | VOC-124 already requested `permission-workflows: write` on `publish-source`; T00's first run is attempt `1` on a new VOC-131 carrier |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Carrier | new VOC-131 branch/PR from current `develop`; not #1051 |
| Infra | consume already-merged #165; do not open a replacement infra PR |
| Pin target | `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| Pin must not equal | `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Mirrored #165 files | `.github/workflows/release.yml` and `tests/test_release_policy.py` only, byte-for-byte |
| Restore | both `identify` and `converge` restore shared policy after caller checkout and before task-completion helpers |
| Restore identity | `job.workflow_repository` + `job.workflow_sha` + `path: karsift-ai-infra` + `persist-credentials: false` |
| VOC-112 fixtures | byte-identical to the carrier `develop` base; `subject_revision` remains `f9d11e23…`; neither path in the implementation diff |
| Hashed VOC-112 sources | do not edit `AGENTS.md` or `.agents/skills/vocanova-repo-navigator/SKILL.md` |
| Preserve | #164 missing-`develop` recovery, exact-SHA develop sync, unique-develop fail-closed, promotion checks, serialization, review/implementer split, retry bounds, secret/raw-error controls |
| Exceptional action | live identity remains `reconcile-production-change` |
| Operator identity | adopted authority issue number; no free-form SHA inputs on caller `workflow_dispatch` |
| `existing_pr_number` | remains implement-only |
| VOC-129 | do not re-implement #1046; do not manufacture a VOC-129 completion marker; promote through repaired `release.yml@main` / `reconcile-release` |
| VOC-130 | do not retry #1049 / #1051; do not manufacture a VOC-130 completion marker; close as superseded after this replacement promotes |
| Attempt | VOC-131-T00 attempt `1` on this carrier |
| `roles.yml` | unchanged |
| OpenAI | not authorized |
| Snapshot-the-gap task? | **no** — forbidden (`karsift-ai-infra#15`) |
| Bootstrap | none |
| Secrets | do not print credential values; do not rotate App secrets |

## Changed surfaces (implementation)

| Surface | Action |
|---------|--------|
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | set to `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` |
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml` | byte-identical copy from infra #165 merge |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py` | byte-identical copy from infra #165 merge |
| `tooling/governance/fixtures/karsift-ai-infra/README.md` | current-state pin paragraph for VOC-131-T00 and post-caller-checkout restore |
| `tooling/governance/tests/test_voc131_caller_replacement.py` | new deterministic regressions for pin, restore ordering, VOC-112 identity, carrier constraints |
| `tooling/governance/tests/test_voc121_implement_policy.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc122_implement_policy.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc124_implement_policy.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc125_implement_policy.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc125_implement_fixture.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc126_workflow_dispatch_input_limit.py` | live pin literal advanced to `8ce2b77…` |
| `tooling/governance/tests/test_voc129_caller_replacement.py` | authoritative pin advanced to `8ce2b77…`; `#164` pin retained as stale class |
| `scripts/foundation/voc097-fixture-matrix.test.mjs` | pin literal advanced to `8ce2b77…` |
| `scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | pin literal advanced to `8ce2b77…` |
| `scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | pin literal advanced to `8ce2b77…` |
| `.github/workflows/pipeline.yml` | no edit — caller already dispatches `release.yml@main` |
| `scripts/foundation/fixtures/voc112-*.json` | not modified |
| `AGENTS.md`, navigator skill | not modified |

Exact comparison of every other fixture path shared with infra merge
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398` found no additional byte drift
beyond `release.yml`, `test_release_policy.py`, and caller-local README /
CHANGELOG commentary.

## Validation commands (implementation)

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
python3 -m unittest tooling.governance.tests.test_voc131_caller_replacement
```

Record exact pass/fail output below after the workflow runs validation.

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | OK (231 tests) |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'` | OK (379 tests) |
| `node scripts/foundation/voc097-fixture-matrix.test.mjs` | pass 5 |
| `node scripts/foundation/voc104-ready-for-review-reuse.test.mjs` | pass 6 |
| `node scripts/foundation/voc108-authoritative-lifecycle.test.mjs` | pass 5 |
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_voc131*.py'` | OK (10 tests) |
| `bash scripts/governance/validate-governance.sh` | passed |
| `bash scripts/governance/classify-change-risk.sh` | passed (R4 path floor) |
| `git diff --check` | clean |

## Independent verification (implementation)

Pending exact-SHA independent review of the caller implementation PR. The
implementer must not approve or merge its own work.

## Promotion and closure (post-merge)

Pending. After the exact reviewed caller merge:

- VOC-129 skipped promotion from run `33066533397` completes through the
  repaired `release.yml@main` path or `reconcile-release` when a valid
  App-authored completion marker exists. Do not manufacture a VOC-129
  completion marker.
- Exhausted VOC-130 (#1049 / unmerged #1051) is closed as superseded after
  this replacement is promoted, with audit comments naming the VOC-131 merge.
  Do not manufacture a VOC-130 completion marker. Do not merge #1051.
- This package promotes through the same repaired path.
- `develop` is advanced to each successful promotion merge SHA before audit
  close.
- Release/task/requirement records close with audit comments naming the exact
  promotion merges for VOC-129, VOC-130, and this replacement.
- Closed state alone is not completion proof.
