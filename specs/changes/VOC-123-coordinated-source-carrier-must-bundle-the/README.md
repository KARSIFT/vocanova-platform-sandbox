# VOC-123 — Coordinated source carrier must bundle the committed head through a named ref

| Field | Value |
|-------|-------|
| Package | `VOC-123` |
| Title | Coordinated source carrier must bundle the committed head through a named ref |
| Path | `specs/changes/VOC-123-coordinated-source-carrier-must-bundle-the` |
| Status | `draft` |
| Risk | `R4` (draft proposal; coordinated source-carrier publication and `tooling/governance/` fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1005](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1005) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-121 made the governed implementer preserve authorized nested
`karsift-ai-infra` edits, commit them in isolated repository state, and publish
them from a clean `publish-source` job. That publisher cannot run if Git
refuses to create the source bundle.

The implement job creates the nested source bundle from an exclusion range
whose positive tip is only a raw object ID. Git bundle creation needs a named
positive revision to advertise as the bundle head. A range ending only in a
raw SHA traverses the intended commit but advertises no reference, so Git
exits 128 with `fatal: Refusing to create empty bundle.`

### Live reproduction (2026-08-26)

| Item | Value |
|------|-------|
| Adopted task that hit the defect | [#1003](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1003) (`VOC-122-T00`) |
| Pipeline run / job | `32915078678` / `98017696468` |
| Nested commit created on the runner | `db31cc9` (`VOC-122: VOC-122-T00 coordinated source carrier (attempt 1)`) |
| Nested diff | 4 files, 506 insertions, 73 deletions, including `tests/test_voc122_actions_check_recovery.py` |
| Failure | `fatal: Refusing to create empty bundle.` before artifact upload |
| Carrier PRs | none — no source bundle and no caller recovery artifact were uploaded |
| Defect locus | `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` `Commit implementer's work`: `bundle create … "${{ steps.infra-checkout.outputs.base_sha }}..$SOURCE_HEAD_SHA"` |
| Caller fixture mirror | `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml` |

Deterministic local reproduction from issue #1005:

```bash
base_sha=$(git rev-parse HEAD~1)
head_sha=$(git rev-parse HEAD)
git bundle create raw.bundle "$base_sha..$head_sha"
# fatal: Refusing to create empty bundle. (exit 128)

git branch carrier "$head_sha"
git bundle create ref.bundle "$base_sha..carrier"
# succeeds; bundle list-heads advertises refs/heads/carrier
```

This is deterministic and will recur on every coordinated source change. A
workflow rerun cannot recover the runner-only nested commit.

Existing VOC-121 tests do not catch it: they create `source_base..HEAD` or
`base_sha..$branch`, which are named positive tips.

## Required outcome (summary)

Use one largest-safe coherent task to repair every active bundle-creation path
that has the same raw-positive-tip defect, with the minimum required caller
fixture/pin/evidence coordination:

1. Before source bundle creation, bind the exact committed `SOURCE_HEAD_SHA` to
   an isolated temporary ref (or an equivalently safe named ref), create the
   bundle from `base_sha..that-ref`, verify the bundle advertises exactly the
   expected head/ref, and remove the temporary ref afterward.
2. Inspect the caller recovery bundle and planner recovery bundle paths
   (`implement.yml` `integration_sha..HEAD` and `plan.yml` `base_sha..HEAD`).
   Prove whether `HEAD` is advertised safely and change only paths that
   reproduce the defect.
3. Preserve exact base/head binding, nested-repository isolation, bundle
   verification, artifact integrity, force-with-lease publishing, App-token
   separation, retry limits, independent exact-SHA review, protected checks,
   and fail-closed behavior. Never bundle unrelated refs or secrets.
4. Add deterministic tests that create real temporary Git repositories and
   prove the empty-bundle failure, the named-ref success, and the named
   fail-closed cases.
5. Update current-state workflow comments/docs and the caller fixture only
   where consumed. Independently review and merge the infra PR first, then pin
   the caller fixture to that exact infra merge SHA and reconcile active #1003
   delivery against that revision — without implementing VOC-122 here.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Bundle the coordinated source-carrier committed head through a named ref, with tests, docs, and caller pin | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because durable caller fixture/test updates belong
under `tooling/governance/` (R4 path floor) and because the change mutates
coordinated second-repository publication of nested infrastructure. The
path-based classifier and independent verifier remain authoritative; this draft
proposal is not a determination.
