# VOC-146-T00 — Evidence

Task: `VOC-146-T00` — Fail closed on an unresolved or invalid `--base`/`--head`
range in governance validation and risk classification.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`pending — resolve current develop to a 40-character SHA before any in-scope
edit and record it here.`

Issue-creation reproduction commit was
`79b2b3f1f4224235bdda3f77ee887c3004978deb`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Issue #1127 incident record

| Item | Value |
|------|-------|
| Reproduction head | `79b2b3f1f4224235bdda3f77ee887c3004978deb` |
| Nonexistent `--base` | `376e00dd769afb0fe850052b3a5cb48f729e73ad` |
| Command | `bash scripts/governance/validate-governance.sh --base 376e00dd… --head 79b2b3f1…` |
| Observed Git error | `fatal: Invalid symmetric difference expression 376e00dd…...79b2b3f1…` |
| Observed success line | `Governance structure validation passed.` |
| Observed exit status | `0` |
| Nested loader | `scripts/governance/validate-monitoring-impact.sh` `mapfile -t files < <(git diff … "$base...$head")` |
| Twin loader | `scripts/governance/classify-change-risk.sh` line 30, same pattern |
| Discovery context | local exact-range verification of emergency PR #1126; unrelated to VOC-112 |

## Range-loading contract

| Case | Required result |
|------|-----------------|
| Nonexistent `--base`, resolvable `--head` | nonzero; no `Governance structure validation passed.` |
| Resolvable `--base`, nonexistent `--head` | nonzero; no success line |
| Two resolvable commits with no merge base | nonzero; do not load an empty file list |
| Partial `--base` or `--head` only | nonzero; do not fall through to working-tree discovery |
| Valid `--base`/`--head` with a successful empty diff | valid empty change set; classifier may report no changed files |
| Valid `--base`/`--head` with real file changes | existing success/classification contract |
| `--files-from` well-formed list | preserved |
| `pull_request` with no resolved range (VOC-086) | still fail-closed |
| `--declarations-only` without a range | still usable |
| Working-tree fallback when no range was requested | preserved |

Record the live helper names, status-preserving capture mechanism, and
exact negative-case SHAs after implementation.

## Exhaustive source-search disposition

Record searched patterns and path disposition after implementation. Pattern
families must include:

| Pattern family | Examples searched |
|----------------|-------------------|
| Process-substitution range loaders | `mapfile < <(git diff`, `$base...$head` |
| Missing-range-only fail-closed claims | `AGENTS.md` monitoring-impact paragraph; VOC-086 tests |
| Success-after-Git-fatal | `Governance structure validation passed.` after invalid range |
| Classifier empty-list success | `No changed files to classify.` |
| Historical packages | `specs/changes/VOC-086-…`, `VOC-112-…` must remain unmodified |
| Pin / infra | `PINNED_SHA.txt` must remain unmodified |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
node --test scripts/foundation/voc146-*.test.mjs
node --test scripts/foundation/voc086-monitoring-impact.test.mjs
git diff --check
```

Plus the issue #1127 class command (expect nonzero; no success line) and the
targeted nonexistent-head, no-merge-base, partial-range, valid-range, and
`--files-from` cases.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

## Release handoff

After the exact reviewed caller merge, ordinary later promotion uses
`release.yml`. Do not snapshot the develop/main gap. Do not recapture VOC-112
fixtures. Do not advance `PINNED_SHA.txt`. Root issue #1127 closes only after
allowlisted metadata from a successful implementation/promotion path exists.
Emergency PR #1126 remains discovery context only.
