# VOC-142-T00 — Evidence

Task: `VOC-142-T00` — Wait for the complete required roster-check set and
reuse the exact open roster PR on reconcile.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

Issue-creation plan PR #1110 merge was
`bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`pending — record the exact new infra merge SHA after independent review.`

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

Record the live wait-completeness predicate and carrier-resolution function
names after implementation.

## Exhaustive source-search disposition

Record searched patterns and path disposition after implementation. Pattern
families must include:

| Pattern family | Examples searched |
|----------------|-------------------|
| Pin and hash assertions | old/current full pin, `CURRENT_PIN`, `AUTHORITATIVE_PIN`, `PINNED_SHA`, SHA-256 tables |
| Roster wait completeness | `stable_green_count`; two stable zero-pending snapshots; complete green logical set |
| Roster PR reuse | `gh pr create` in `Open roster PR`; reconcile always opens a new roster PR; existing artifacts reused |
| Authority state | active A-004; founder `approved` as a merge gate |
| Historical packages | `specs/changes/VOC-141-…`, `VOC-140-…`, `VOC-139-…` must remain unmodified |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Plus targeted VOC-142 cases for IN_PROGRESS/late `ci / ci` wait
completeness, exact open-PR reuse, already-merged reuse,
mismatch/ambiguous rejection, and duplicate task/dispatch suppression.

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
