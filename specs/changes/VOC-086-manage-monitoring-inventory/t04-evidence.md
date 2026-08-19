---
evidence_id: VOC-086-EV-04
task_id: VOC-086-T04
acceptance_criteria:
  - VOC-086-AC-08
  - VOC-086-AC-09
tests:
  - VOC-086-TEST-14
  - VOC-086-TEST-15
  - VOC-086-TEST-16
date: 2026-08-19
related_change: VOC-086
accountable_owner: unassigned
gate_status: repository-complete-live-monitoring-pending
live_sync_claimed: false
remediation_of: bfb72c0194d45116a896e0bffc2a4c4709bf3b76
---

# VOC-086-T04 — monitoring_impact governance and CI validation

## Scope of this evidence

This task delivers `monitoring_impact` declaration validation, grandfathering of
untouched historical packages, route/critical-endpoint fail-closed checks,
CI wiring that supplies pull-request `--base`/`--head`, change-package template
and `AGENTS.md` drafting docs, and deterministic TEST-14/15/16 coverage. It does
**not** claim live Kuma mutation, scheduled-synthetic green proof, or AC-05/AC-10
closure (`VOC-086-T05`). Historical package `change.yaml` files were not rewritten.

## Remediation (post independent FAIL on `bfb72c0…`)

| Finding | Fix |
| --- | --- |
| High — CI did not enforce new/changed-package or route fail-closed on PRs | `validate-governance.sh` now accepts and forwards `--base`/`--head`/`--files-from`. `governance-policy.yml` and `repository-governance.yml` pass `${{ github.event.pull_request.base.sha }}` / `head.sha` on pull_request (same pattern as `classify-change-risk.sh`). Repository-governance checkout uses `fetch-depth: 0`. The wrapper no longer reads non-existent `GITHUB_BASE_SHA`; it uses explicit SHAs, `GITHUB_EVENT_PATH`, or fail-closed on `pull_request` with no resolved range. |
| Medium — AGENTS.md doc-sync omitted | `AGENTS.md` now has "Drafting `monitoring_impact` in `change.yaml`" and plan-review text names the field. |
| Medium — missing `t04-evidence.md` | This file. |
| Low — `*_test.go` classified as API routes | Classifier excludes `*_test.go`. Route gate still accepts a valid `state: none` declaration (literal AC-08). |

## Decision records for this task

| Decision | Recorded choice |
| --- | --- |
| Validator runtime | Node ESM `infra/monitoring/validate-monitoring-impact.mjs` + shell wrapper `scripts/governance/validate-monitoring-impact.sh` |
| CI changed-file range | Explicit `--base`/`--head` from pull-request SHAs; `GITHUB_EVENT_PATH` fallback; fail-closed if a `pull_request` event still has no range |
| Grandfathering | Require `monitoring_impact` only when `change.yaml` is in the changed-file set (or a new package slug is supplied). Untouched historical packages stay exempt. |
| VOC-086 own declaration | `monitoring_impact.state: add` with the five availability IDs and five synthetic IDs remains valid against the canonical inventory |

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Validator | `infra/monitoring/validate-monitoring-impact.mjs` |
| Shell wrapper | `scripts/governance/validate-monitoring-impact.sh` |
| Governance wiring | `scripts/governance/validate-governance.sh` |
| PR CI SHA passing | `.github/workflows/governance-policy.yml`, `.github/workflows/repository-governance.yml` |
| Template | `specs/templates/change-package/change.yaml`, `specs/templates/change-package/README.md` |
| Agent drafting docs | `AGENTS.md` |
| Deterministic tests | `scripts/foundation/voc086-monitoring-impact.test.mjs` |
| This evidence | `specs/changes/VOC-086-manage-monitoring-inventory/t04-evidence.md` |

## Acceptance mapping

| AC | Repository outcome |
| --- | --- |
| AC-08 | New/changed `change.yaml` must declare `none\|existing\|add\|update`; `none` needs rationale; other states need canonical IDs; CI rejects missing/invalid declarations when the PR range is supplied; route/critical-endpoint changes fail without a valid in-diff declaration; historical unmodified packages grandfathered. |
| AC-09 (governance half) | Template + `AGENTS.md` describe drafting; TEST-14/15/16 cover positive, negative, grandfathering, route fail-closed, and CI SHA wiring. |

## Deterministic validation

Commands (repo root):

```bash
node --test scripts/foundation/voc086-monitoring-impact.test.mjs
bash -n scripts/governance/validate-monitoring-impact.sh
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Results recorded this revision:

| Command | Result |
| --- | --- |
| `node --test scripts/foundation/voc086-monitoring-impact.test.mjs` | pass (7 tests) |
| `bash -n scripts/governance/validate-monitoring-impact.sh` | pass |
| `bash scripts/governance/validate-governance.sh` | pass (`Monitoring impact validation passed.` twice, then structure passed) |
| `bash scripts/governance/classify-change-risk.sh` | path floor **R4** (`scripts/governance/*`, `governance-policy.yml`, `repository-governance.yml`) |
| `git diff --check` | pass |

## Follow-up notes (out of scope for T04)

- Local working-tree classification also listed untracked `karsift-ai-infra/` as R1; that tree is not part of this task.
- `karsift-ai-infra` caller workflows that invoke `validate-governance.sh` without `--base`/`--head` still get `GITHUB_EVENT_PATH` fallback on `pull_request`, then fail-closed if no range can be resolved. Deepening those callers is not required for this package's CI path.
- Route fail-closed still treats `state: none` plus rationale as a valid declaration (matches AC-08 wording; TEST-16 does not reject it).
