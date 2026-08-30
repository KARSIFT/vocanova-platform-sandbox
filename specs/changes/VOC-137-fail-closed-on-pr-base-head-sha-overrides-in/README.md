# VOC-137 — Fail closed on PR base/head SHA overrides in every scannable caller executable

| Field | Value |
|-------|-------|
| Package | `VOC-137` |
| Title | Fail closed on PR base/head SHA overrides in every scannable caller executable |
| Path | `specs/changes/VOC-137-fail-closed-on-pr-base-head-sha-overrides-in` |
| Status | `draft` |
| Risk | `R4` (draft proposal; `tooling/governance/` scanner and tests) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1083](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1083) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

VOC-136 implementation PR [#1080](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080)
merged exact reviewed head `5d0c2350ab9a20ace586eaadd1169203140ffad0` as
`develop` merge `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`, but its exhaustive
bypass scanner does not satisfy a required negative case.

In `tooling/governance/tests/voc136_bypass_scan.py`, `PR_SHA_SET_PATTERN` is
evaluated only when the relative filename contains `validate-workspace` or
ends in `.test.mjs`. The enclosing scan otherwise accepts arbitrary executable
names. Consequently an added file such as `scripts/arbitrary-wrapper.sh`
containing:

```sh
export PR_BASE_SHA=deadbeef
pnpm test
```

is accepted instead of failing closed. The same filename dependency can miss
Node or Python wrappers that set `PR_BASE_SHA` or `PR_HEAD_SHA` around
validation/test execution.

| Item | Value |
|------|-------|
| Pre-merge supervisor evidence | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444745877) |
| Reviewer verdict that missed the gap | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444825638) |
| Post-merge audit | [comment](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1080#issuecomment-5444861382) |
| Canceled release runs | `33113425829`, `33113547909` |
| Drafted-then-stopped release PR | [#1082](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1082) |
| Issue-creation `develop` | `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` |
| Issue-creation `main` | `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Live pin (already #167) | `b263c0c110591cc798b89277dfc35542abb1597b` |

The automatic release was stopped before `main` changed. This package does not
re-implement VOC-136 and does not rewrite or manufacture its merge, completion,
or review records.

## Root cause

The scanner correctly defines a content-semantic `PR_SHA_SET_PATTERN`, but
applies it through a filename heuristic. That contradicts VOC-136's exhaustive
changed-executable requirement and leaves a trivial rename bypass.

## Required outcome (summary)

Use one largest-safe coherent task and one implementation PR:

1. Remove the filename dependency from PR base/head SHA override detection.
   Every added or modified scannable caller executable outside the exact infra
   fixture mirror must fail closed when executable behavior assigns
   `PR_BASE_SHA` or `PR_HEAD_SHA` in shell, Node, or Python forms relevant to
   validation/test invocation.
2. Preserve source-safe scanner construction so the scanner and its tests do
   not falsely reject their own pattern definitions or inert fixtures.
3. Add deterministic negative regression coverage using arbitrary filenames,
   including the required shell, Node, and Python wrappers and an added or
   modified `*.py` path outside the fixture mirror.
4. Add positive controls proving benign textual discussion, pattern
   construction, and the exact excluded infra fixture mirror do not trigger
   false positives. Do not restore a wholesale exclusion for caller tooling,
   governance, or tests.
5. Keep `PINNED_SHA.txt` and the complete VOC-136 mirrored fixture set
   byte-for-byte identical to infrastructure merge
   `b263c0c110591cc798b89277dfc35542abb1597b`. Do not change model bindings or
   request OpenAI credentials.
6. Run the complete caller governance suite, mirrored infra suite, risk
   classification, governance validation, diff checks, and a direct
   reproduction that the arbitrary-wrapper example now fails closed.
7. Obtain independent exact-revision review that explicitly evaluates the
   arbitrary-filename shell/Node/Python negative cases and benign controls
   before merge.
8. Preserve existing VOC-136 merge/completion/review records as audit evidence.
9. After this correction merges into `develop`, allow the normal governed
   release to reconcile outstanding completed packages, promote `develop` to
   `main`, deploy where applicable, and converge `develop` to the exact
   resulting main merge SHA without a tree-equivalent staging loop.

This is a KARSIFT automation reliability fix, not product behavior. Preserve
A-004 risk classification, protected checks, review independence, and release
gates. Treat this as one outcome-sized task.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Fail closed on PR base/head SHA overrides in every scannable caller executable | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` scanner and
test updates belong under the R4 path floor, and because the change protects
the exhaustive executable-bypass contract that gates application-check
provenance. The path-based classifier and independent verifier remain
authoritative; this draft proposal is not a determination.
