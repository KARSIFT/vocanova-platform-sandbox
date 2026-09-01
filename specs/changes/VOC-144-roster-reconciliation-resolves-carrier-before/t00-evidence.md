# VOC-144-T00 — Evidence

Task: `VOC-144-T00` — Boundedly wait for existing roster-PR head metadata to
expose the pushed SHA, then reuse that carrier.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

This package's own plan/adoption/roster commits after that SHA are
governance-only and do not count as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`pending — record the exact new infra merge SHA after independent review.`

Issue-creation / VOC-142 pin (defective for this failure class; historical
audit): `8993e867640dfb604dec0466c4e0787e68d8e258`.

## Named wait constants

Record after implementation:

| Constant | Value |
|----------|-------|
| Timeout ceiling required by this package | 60 seconds |
| Implemented timeout | pending — named constant, ≤ 60 seconds |
| Implemented poll interval | pending — named constant |

## Issue #1122 incident record

| Item | Value |
|------|-------|
| Merged plan PR | #1110 |
| Generated roster PR | #1112 |
| Deterministic branch | `karsift/roster-voc-141` |
| Issue-creation pin | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Reconcile run 1 | `33437239322` |
| Pushed head 1 | `958e0fedf742173320bb89cfe690ec7070b49e93` |
| Run 1 failure | `Open roster PR` / `MISMATCHED_OPEN_CARRIER` |
| After run 1 | PR #1112 exposed `958e0fed…`; local resolution returned `reuse_open` |
| Reconcile run 2 | `33437514152` |
| Pushed head 2 | `0206cb70437cb751a19a2e715d8202b672060b50` |
| Run 2 failure | same `MISMATCHED_OPEN_CARRIER` class |
| After run 2 | PR #1112 exposed `0206cb70…` |

No roster PR was manually merged, closed, recreated, or bypassed. #1112
remains the single exact VOC-141 carrier.

## Convergence-wait contract

| Case | Required result |
|------|-----------------|
| Unique same-repo/same-ref/same-base OPEN PR; first snapshot stale SHA; later snapshot equals local HEAD | `reuse_open`; do not call `gh pr create` |
| Unique otherwise-matching OPEN PR; listed SHA stays different through the bound | `MISMATCHED_OPEN_CARRIER`; do not create |
| Two OPEN same-ref PRs, wrong base, or wrong repository | fail closed without SHA-lag wait |
| GitHub API or parse failure | fail closed without create |
| Exact first-snapshot OPEN match | `reuse_open` immediately |
| Exact already-merged match | `reuse_merged`; do not create |
| Zero matching carriers | may create exactly one PR |
| VOC-142 complete-required-set wait and in-progress-parent attestation | unchanged |

Record the live helper/function names after implementation.

## Exhaustive source-search disposition

Record searched patterns and path disposition after implementation. Pattern
families must include:

| Pattern family | Examples searched |
|----------------|-------------------|
| Pin and hash assertions | old/current full pin, `CURRENT_PIN`, `AUTHORITATIVE_PIN`, `PINNED_SHA`, SHA-256 tables |
| Immediate carrier resolution | `roster-carrier-runner.py` invoked once after push; `MISMATCHED_OPEN_CARRIER` with no re-fetch; reconcile reuses exact head without waiting for post-push metadata |
| Roster wait completeness | complete required set including `ci / ci`; two stable subset snapshots are not complete (must remain) |
| Authority state | active A-004; founder `approved` as a merge gate |
| Historical packages | `specs/changes/VOC-142-…`, `VOC-141-…`, `VOC-140-…`, `VOC-139-…` must remain unmodified |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Plus targeted VOC-144 cases for stale-then-converge, timeout-still-stale,
durable-mismatch-does-not-wait, API failure, and preserved VOC-142 carrier
paths.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

## Release handoff

After the exact reviewed caller merge, ordinary `reconcile` for plan PR
#1110 may reuse #1112 when that PR still matches, including across
post-push REST lag. Do not snapshot the develop/main gap. Do not manually
merge, close, recreate, or bypass #1112. Do not create a duplicate VOC-141
task, roster PR, promotion PR, or release audit. Root issue #1122 closes
only after allowlisted metadata from a successful adopt/reconcile run
exists. Incident runs `33437239322` and `33437514152` remain incident
evidence only.

## Bootstrap sequencing note

This package's own first native adoption may still execute the pre-pin
`adopt.yml` until the independently reviewed infra merge above is what
callers pin. That timing tension is not a bootstrap exception and does not
authorize manual roster merge or gate weakening.
