# VOC-146 — Governance validation accepts an invalid base/head range

| Field | Value |
|-------|-------|
| Package | `VOC-146` |
| Title | Governance validation accepts an invalid base/head range |
| Path | `specs/changes/VOC-146-governance-validation-accepts-an-invalid-base` |
| Status | `draft` |
| Risk | `R4` (draft proposal; `scripts/governance/*` is an R4 path floor) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1127](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1127) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

`scripts/governance/validate-governance.sh` reports success when `--base` or
`--head` names a nonexistent commit. The nested monitoring-impact validator
emits Git's fatal range error but continues and returns zero, so malformed
pull-request range metadata is not fail-closed.

From clean commit `79b2b3f1f4224235bdda3f77ee887c3004978deb`:

```bash
bash scripts/governance/validate-governance.sh \
  --base 376e00dd769afb0fe850052b3a5cb48f729e73ad \
  --head 79b2b3f1f4224235bdda3f77ee887c3004978deb
echo $?
```

Observed output includes Git's fatal symmetric-difference error and then
`Governance structure validation passed.` The exit status is `0`.

The same `mapfile < <(git diff "$base...$head")` pattern exists in
`scripts/governance/classify-change-risk.sh`. An invalid range there becomes
an empty file list and `No changed files to classify.` with exit `0`.

## Root cause

`scripts/governance/validate-monitoring-impact.sh` obtains changed files with:

```bash
mapfile -t files < <(git diff --no-renames --name-only --diff-filter=ACDMRTUXB "$base...$head")
```

The failing `git diff` runs inside process substitution; its nonzero status is
not propagated through `mapfile`, even under `set -euo pipefail`. The resulting
empty file list is then accepted, and the parent wrapper prints success.

This was discovered during local exact-range verification of emergency PR
#1126. It is unrelated to that PR's VOC-112 provenance objective.

## Required outcome (summary)

Use one largest-safe coherent task and one implementation PR:

1. An unresolved `--base`/`--head` commit or invalid diff range must make
   governance validation return nonzero before it claims success.
2. Resolve both revisions explicitly and capture `git diff` through a
   status-preserving path before loading its output.
3. Apply the same fail-closed range loading to `classify-change-risk.sh`,
   which uses the identical process-substitution pattern.
4. Add deterministic negative tests for missing base, missing head, and
   unrelated/no-merge-base revisions, while preserving valid PR and
   `--files-from` behavior.
5. Update current-state docs that describe range fail-closed behavior,
   including `AGENTS.md`.

This is a caller governance-script fail-closed repair, not product behavior
and not a karsift-ai-infra pin change. Preserve A-004 risk classification,
protected checks, review independence, and release gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Fail closed on an unresolved or invalid `--base`/`--head` range in governance validation and risk classification | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because `scripts/governance/*` is an R4 path
floor in `classify-change-risk.sh` and
`.github/approved-policy/protected-paths.yaml`, and because the change
mutates fail-closed governance-validation behavior used by CI. The
path-based classifier and independent verifier remain authoritative; this
draft proposal is not a determination.
