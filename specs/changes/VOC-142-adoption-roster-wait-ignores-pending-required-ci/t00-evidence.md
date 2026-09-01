# VOC-142-T00 — Evidence

Task: `VOC-142-T00` — Wait for the complete required roster-check set and
reuse the exact open roster PR on reconcile.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

## Implementation PR base

Recorded before the first in-scope edit:

`1fc3473576bef96cffd861f4304168ad147296ef`

Issue-creation plan PR #1110 merge was
`bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`8993e867640dfb604dec0466c4e0787e68d8e258`

Issue-creation / VOC-140 pin (defective for this failure class; historical
audit): `67bdfd13ef875dead23ce4be01d7d0e8b976e289`.

## Issue #1113 incident record

| Item | Value |
|------|-------|
| Source issue | #1109 |
| Merged plan PR | #1110 |
| Reviewed plan head | `178ae81ed4f0224fbb12359c90ddb67c6687ca9a` |
| Plan merge | `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2` |
| Adoption run | `33343125733`, job `99342230038` |
| Generated task | #1111 |
| Generated roster PR | #1112 |
| Exact roster head | `98dd0936a73b64a6b548da6cf2000a6d000917ac` |
| Roster pipeline run | `33343147453` |
| Required CI job | `99342299218` |
| Wait started | `2026-08-30T23:57:08Z` |
| Wait reported SUCCESS | `2026-08-30T23:58:46Z` |
| Merge failed | `2026-08-30T23:58:47Z` |
| `ci / ci` started | `2026-08-30T23:57:21Z` |
| `ci / ci` SUCCESS | `2026-08-30T23:59:16Z` |
| Reconcile run | `33343250178`, job `99342577393` |
| Reconcile failure step | `Open roster PR` |

No roster PR was manually merged, closed, recreated, or bypassed. #1112
remains the single exact VOC-141 carrier.

## Wait and reuse contract

| Case | Required result |
|------|-----------------|
| Required `ci / ci` IN_PROGRESS; `governance-policy` and `validate` SUCCESS | wait does not complete; merge is not attempted |
| Required `ci / ci` not yet registered; two stable SUCCESS-only snapshots | wait does not complete |
| Complete required set SUCCESS on exact head | wait may complete; merge may proceed |
| Exact matching OPEN roster PR already exists | reuse that PR; do not call `gh pr create` |
| Exact matching already-merged roster PR exists | do not create another PR; follow merged-roster cleanup/dispatch |
| Mismatched or ambiguous carriers | fail closed |
| Existing OPEN task issue; no implementation PR | reuse the task; dispatch root once after checked merge |
| Production merge guard and in-progress-parent attestation | unchanged |

Live predicates after implementation:

- `config/roster_pr_wait.py`: `missing_required_roster_contexts`,
  `roster_required_set_complete`, `roster_wait_snapshot`
- `config/roster-pr-wait-runner.py`: wraps `authoritative-checks-runner.py`
  with `--roster-pr-gate` and augments output with `required_complete` /
  `missing_required`
- `config/roster_carrier.py`: `resolve_roster_carrier`
- `config/roster-carrier-runner.py`: gh-backed carrier lookup for adopt
- `authoritative-checks-runner.py`: `_roster_pr_pipeline_parent` admits
  in-progress `karsift/roster-*` pipeline checks as pending without making
  non-terminal parents attestable for merge-gate or release

## Exhaustive source-search disposition

| Pattern family | Path | Disposition |
|----------------|------|-------------|
| `stable_green_count` without required-set guard | prior `adopt.yml` wait loop | updated in infra `adopt.yml` |
| `authoritative-checks-runner.py` without roster gate | prior wait path | updated via `roster-pr-wait-runner.py` |
| `gh pr create` unconditional on `changed == true` | prior `Open roster PR` | updated via `roster-carrier-runner.py` |
| `67bdfd13…` as `CURRENT_PIN` / live pin | governance tests, foundation mjs, fixture README historical tail | updated current pin to `8993e867…`; preserved `67bdfd13…` as `AUTHORITATIVE_PIN` / historical |
| `AGENTS.md` reconcile procedure | caller root | updated for complete required set and carrier reuse |
| fixture `README.md` roster wait paragraph | infra fixture README | updated |
| `adopt.yml` header comment | infra fixture | updated |
| VOC-141/VOC-140/VOC-139 package directories | `specs/changes/VOC-141-…` etc. | historical; not modified |
| VOC-112 JSON fixtures | `scripts/foundation/fixtures/voc112-*.json` | historical; not recaptured (package non-goal); attempt-2 reverted erroneous `agents_sha256` edits |
| `voc112-navigation-benchmark.test.mjs` AGENTS binding | caller foundation tests | narrowed for exact `pr-validation`: its supplied immutable base/head remain the hash anchor; all other modes retain captured-revision/working-tree binding without retargeting fixtures |

## Validation results (working tree; pre-commit)

Implementation head for the commands below is the caller workflow commit that
contains this evidence file.

| Command | Result |
|---------|--------|
| `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'` | **291 tests OK** |
| `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'` | **478 tests OK** |
| `PR_BASE_SHA=1fc34735… PR_HEAD_SHA=<working-tree> VOC112_CAPTURE_PROVENANCE_MODE=pr-validation node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | **17 tests OK** |
| `bash scripts/governance/validate-governance.sh --base 1fc34735… --head <working-tree>` | **passed** |
| `bash scripts/governance/classify-change-risk.sh --base 1fc34735… --head <working-tree>` | **R4 floor detected** |
| `git diff --check` | **passed** |

Attempt-2 repair: reverted out-of-scope `voc112-*.json` fixture hash edits;
`voc112-navigation-benchmark.test.mjs` now binds `AGENTS.md` provenance via
merge-base, PR-head, canonical-capture-subject, and fixture-agents-base
resolution without recapturing VOC-112 fixtures.

## Validation commands (after the repair is tracked and committed)

```bash
bash scripts/governance/validate-governance.sh --base 1fc3473576bef96cffd861f4304168ad147296ef --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base 1fc3473576bef96cffd861f4304168ad147296ef --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
python3 -m unittest tooling.governance.fixtures.karsift-ai-infra.tests.test_voc142_roster_wait_and_carrier
git diff --check
```

Targeted VOC-142 cases are covered by
`tests/test_voc142_roster_wait_and_carrier.py` and caller
`test_voc142_adoption_roster_wait.py`.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

## Release handoff

After the exact reviewed caller merge, ordinary `reconcile` for plan PR
#1110 may reuse #1112 when that PR still matches. Do not snapshot the
develop/main gap. Do not manually merge, close, recreate, or bypass #1112.
Do not create a duplicate VOC-141 task, roster PR, promotion PR, or release
audit. Root issue #1113 closes only after allowlisted metadata from a
successful adopt/reconcile run exists. Incident runs `33343125733`,
`33343147453`, and `33343250178` remain incident evidence only.

## Bootstrap sequencing note

This package's own first native adoption may still execute the pre-pin
`adopt.yml` until the independently reviewed infra merge above is what
callers pin. That timing tension is not a bootstrap exception and does not
authorize manual roster merge or gate weakening.
