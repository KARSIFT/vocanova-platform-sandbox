# VOC-143-T00 — Evidence

Task: `VOC-143-T00` — Bind promotion-path VOC-112 `AGENTS.md` provenance to
an immutable historical ancestor.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

Issue-creation promotion PR #1119 head was
`376e00dd769253d7a255660f5391fb208781e2f3`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure pin

Live caller pin at issue creation (not this defect; do not pin-advance
unless an infra contract change is proven necessary):

`8993e867640dfb604dec0466c4e0787e68d8e258`

## Issue #1120 incident record

| Item | Value |
|------|-------|
| Promotion PR | #1119 |
| Exact promotion head | `376e00dd769253d7a255660f5391fb208781e2f3` |
| Failing required checks | `validate`, `ci / ci` |
| Release audit | #1118 |
| Prior causal package | VOC-142 (`AGENTS.md` documentation update; fixtures not recaptured per DEP-07) |
| Live pin | `8993e867640dfb604dec0466c4e0787e68d8e258` |

No promotion PR was manually merged, closed, recreated, or bypassed. #1119
remains the VOC-142 promotion carrier at issue creation.

## Provenance contract

| Case | Required result |
|------|-----------------|
| `squash-safe-push`; historical fixture `agents_sha256`; current working-tree `AGENTS.md` differs | pass; bind to immutable ancestor of `PR_HEAD_SHA` or `HEAD` |
| `squash-safe-push`; tampered/unfound `agents_sha256` | fail closed |
| `local`; historical fixture vs current working tree | still requires working-tree equality |
| `pr-ancestry`; historical fixture vs current working tree | still requires working-tree / captured-revision binding |
| Ordinary `pr-validation`; HEAD hashes vs merge base | still requires merge-base anchoring |
| Promotion `pr-validation`; historical fixture `agents_sha256`; HEAD `AGENTS.md` differs; navigator hashes at HEAD | pass |
| Promotion `pr-validation`; merge-base-only navigator hashes | fail closed |
| Promotion `pr-validation`; non-ancestor base | fail closed |
| Promotion check identity | `validate` stays `squash-safe-push`; `ci / ci` stays `--promotion-pr` / `pr-validation` |
| VOC-112 JSON fixtures | byte-identical; not recaptured |

Live predicates after implementation:

- pending — record the function/assertion names and the ancestor-walk tip
  rule actually used.

## Exhaustive source-search disposition

| Pattern family | Path | Disposition |
|----------------|------|-------------|
| pending — record searched patterns and each match as updated, historical, or irrelevant | | |

## Validation results (working tree; pre-commit)

Implementation head for the commands below is the caller workflow commit that
contains this evidence file.

| Command | Result |
|---------|--------|
| pending | pending |

## Validation commands (after the repair is tracked and committed)

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
node --test scripts/foundation/voc114-actions-check-recovery.test.mjs
git diff --check
```

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

## Release handoff

After the exact reviewed caller merge, ordinary `reconcile-release` for
release issue #1118 may re-evaluate #1119 when that PR still matches. Do not
snapshot the develop/main gap. Do not manually merge, close, recreate, or
bypass #1119. Do not create a duplicate promotion PR or release audit. Root
issue #1120 closes only after allowlisted metadata from a successful
recovery/release run exists. Incident PR #1119, head `376e00dd…`, and audit
#1118 remain incident evidence only.
