# VOC-146-T00 — Evidence

Task: `VOC-146-T00` — Fail closed on an unresolved or invalid `--base`/`--head`
range in governance validation and risk classification.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

## Implementation PR base

Recorded before the first in-scope edit:

`6f1e72206d04dbf7327a9194661ae1a4a806572e` (`develop` at task start).

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

Shared helper: `scripts/governance/load-changed-files.sh`

| Mechanism | Detail |
|-----------|--------|
| Commit resolution | `git rev-parse --verify --end-of-options "${rev}^{commit}"` |
| Status-preserving diff | `git diff … >"$tmp"` with explicit nonzero exit propagation |
| Partial range guard | Explicit `--base`/`--head` flags fail closed before PR-event fill-in |
| `--declarations-only` | Skips PR-event/working-tree range loading when no range was requested |

| Case | Required result | Post-repair result |
|------|-----------------|-------------------|
| Nonexistent `--base`, resolvable `--head` | nonzero; no success line | exit `1`; `Unable to resolve --base commit` |
| Resolvable `--base`, nonexistent `--head` | nonzero; no success line | exit `1`; `Unable to resolve --head commit` |
| Two resolvable commits with no merge base | nonzero | exit `1`; `Unable to load changed files for range` |
| Partial `--base` or `--head` only | nonzero; no working-tree fallback | exit `1`; `requires both --base and --head` |
| Valid `--base`/`--head` with successful empty diff | valid empty change set | exit `0`; classifier may report no changed files |
| Valid `--base`/`--head` with real file changes | existing success contract | exit `0` |
| `--files-from` well-formed list | preserved | exit `0` under `GITHUB_EVENT_NAME=pull_request` |
| `pull_request` with no resolved range (VOC-086) | still fail-closed | exit `1` (VOC-086-TEST-16) |
| `--declarations-only` without a range | still usable | exit `0` under `GITHUB_EVENT_NAME=pull_request` |
| Working-tree fallback when no range was requested | preserved | unchanged in `classify-change-risk.sh` |

## Exhaustive source-search disposition

| Pattern family | Path | Disposition |
|----------------|------|-------------|
| `mapfile < <(git diff` range loaders | `scripts/governance/validate-monitoring-impact.sh` | **updated** — uses `load-changed-files.sh` |
| `mapfile < <(git diff` range loaders | `scripts/governance/classify-change-risk.sh` | **updated** — uses `load-changed-files.sh` |
| Missing-range-only fail-closed claims | `AGENTS.md` monitoring-impact paragraph | **updated** — unresolved commits and invalid three-dot diffs are fail-closed |
| Missing-range-only fail-closed claims | `.github/workflows/governance-policy.yml` | **unchanged** — passes explicit `--base`/`--head` |
| Missing-range-only fail-closed claims | `.github/workflows/repository-governance.yml` | **unchanged** — passes explicit `--base`/`--head` |
| Missing-range-only fail-closed claims | `scripts/foundation/voc086-monitoring-impact.test.mjs` | **unchanged** — VOC-086 missing-range contract preserved |
| Success-after-Git-fatal | live scripts | **removed** — Git nonzero status propagates |
| Classifier empty-list on invalid range | `classify-change-risk.sh` | **fixed** — invalid range exits nonzero before empty-list message |
| Historical packages | `specs/changes/VOC-086-…/`, `VOC-112-…/` | **unchanged** (audit evidence) |
| Pin / infra | `PINNED_SHA.txt`, `tooling/governance/fixtures/karsift-ai-infra/**` | **unchanged** |
| `config/roles.yml` | repository root | **unchanged** |

## Validation commands (after the repair is tracked and committed)

Issue #1127 class (nonexistent `--base`; expect nonzero; no success line):

```bash
bash scripts/governance/validate-governance.sh \
  --base 376e00dd769afb0fe850052b3a5cb48f729e73ad \
  --head 6f1e72206d04dbf7327a9194661ae1a4a806572e
# exit 1; no "Governance structure validation passed."
```

Valid implementation-PR range (same SHA empty diff at task start):

```bash
bash scripts/governance/validate-governance.sh \
  --base 6f1e72206d04dbf7327a9194661ae1a4a806572e \
  --head 6f1e72206d04dbf7327a9194661ae1a4a806572e
# exit 0; "Governance structure validation passed."
```

Path-based risk classification on the same range:

```bash
bash scripts/governance/classify-change-risk.sh \
  --base 6f1e72206d04dbf7327a9194661ae1a4a806572e \
  --head 6f1e72206d04dbf7327a9194661ae1a4a806572e
# exit 0; valid empty range reports "No changed files to classify."
```

Foundation suites:

```bash
node --test scripts/foundation/voc146-range-loading.test.mjs   # 8/8 pass
node --test scripts/foundation/voc086-monitoring-impact.test.mjs  # 7/7 pass
git diff --check  # pass
```

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
