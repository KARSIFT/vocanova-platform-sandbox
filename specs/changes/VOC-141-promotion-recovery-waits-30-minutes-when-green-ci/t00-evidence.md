# VOC-141-T00 — Evidence

Task: `VOC-141-T00` — Dispatch dedicated promotion-pr-validation immediately
when green CI is unattestable.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

Issue-creation promotion PR #1090 head was
`c3a53bab3035b7f08c0fb959bdf1b56bf330d291`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure merge

Independently reviewed `KARSIFT/karsift-ai-infra` merge consumed by this pin:

`pending — record the exact new infra merge SHA after independent review.`

Issue-creation / VOC-140 pin (defective for this failure class; historical
audit): `67bdfd13ef875dead23ce4be01d7d0e8b976e289`.

## Issue #1109 incident record

| Item | Value |
|------|-------|
| Promotion PR | #1090 |
| Exact promotion head | `c3a53bab3035b7f08c0fb959bdf1b56bf330d291` |
| Infrastructure policy | `KARSIFT/karsift-ai-infra@67bdfd13ef875dead23ce4be01d7d0e8b976e289` |
| Failing release carrier | run `33340381776`, job `99334840338` |
| Step | `Recover missing exact-head promotion checks` |
| Started | `2026-08-30T22:54:51Z` |
| Ended | `2026-08-30T23:25:23Z` |
| Timeout diagnostics | `mode: promotion_pr`; `missing_checks: none`; `pending: 0`; `failed: 0`; `successful: 6`; then exit 1 |
| Duplicate no-progress carrier | run `33340516672` |
| Workaround dispatch | `gh workflow run pipeline.yml --ref develop -f action=recover-promotion-pr-checks -f promotion_pr_number=1090` |
| Workaround dedicated SUCCESS | run `33341923799`, titled `promotion-pr-validation PR #1090` |
| Subsequent reconcile-release | run `33342062118` succeeded, including the production merge guard |

## Recovery dispatch contract

| Case | Required result |
|------|-----------------|
| Required `ci / ci` SUCCESS plus filtered/unattestable parent | immediately dispatch exactly one `recover-promotion-pr-checks`; do not rerun doomed PR `ci / ci`; do not wait 1,800 seconds |
| Valid completed dedicated parent (`promotion-pr-validation PR #<n>`) | `recovery_complete` true; no redispatch |
| Active or successful exact dedicated recovery | suppress duplicate dedicated dispatch |
| In-progress/successful release carrier or duplicate native carrier | must not suppress dedicated dispatch |
| Timeout with unattestable composed CI | diagnostics name `unattestable_ci_evidence` (or live equivalent), not `missing_checks: none` alone |
| Production/release-carrier fail-closed and VOC-140 two-token guard | unchanged |

Record the live diagnostic token and planner-function names after
implementation.

## Exhaustive source-search disposition

Record searched patterns and path disposition after implementation. Pattern
families must include:

| Pattern family | Examples searched |
|----------------|-------------------|
| Pin and hash assertions | old/current full pin, `CURRENT_PIN`, `AUTHORITATIVE_PIN`, `PINNED_SHA`, SHA-256 tables |
| Recovery completeness | required-check SUCCESS completes recovery; `missing_checks: none`; dedicated dispatch when no completed non-carrier run exists |
| Authority state | active A-004; founder `approved` as a merge gate |
| Historical packages | `specs/changes/VOC-140-…`, `VOC-139-…`, `VOC-138-…` must remain unmodified |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Plus targeted VOC-141 cases for unattestable-SUCCESS dispatch, duplicate
suppression, timeout diagnostics, and unchanged carrier fail-closed
boundaries.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

## Release handoff

After the exact reviewed caller merge, ordinary `reconcile-release` may
recover the live same-repository promotion without waiting 1,800 seconds on
unattestable SUCCESS `ci / ci`. Do not snapshot the develop/main gap. Do not
create a duplicate promotion PR or release audit. Root issue #1109 closes
only after allowlisted metadata from a successful recovery/release run
exists. Workaround runs `33341923799` and `33342062118` remain incident
evidence only.
