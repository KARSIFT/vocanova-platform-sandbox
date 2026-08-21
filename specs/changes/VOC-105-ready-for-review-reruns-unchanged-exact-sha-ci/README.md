# VOC-105 — Ready-for-review reruns unchanged exact-SHA CI and model review

| Field                     | Value                                                                                      |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| Package                   | `VOC-105`                                                                                  |
| Title                     | Ready-for-review reruns unchanged exact-SHA CI and model review                            |
| Path                      | `specs/changes/VOC-105-ready-for-review-reruns-unchanged-exact-sha-ci`                     |
| Status                    | `draft` (not adopted)                                                                      |
| Risk                      | `R4` (draft proposal; path floor measured R3; human + classifier govern each task)         |
| Authority model           | A-004 active                                                                               |
| Requirement source        | GitHub issue [#872](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/872)       |
| Target branch             | `develop`                                                                                  |
| Approval                  | `not-approved`                                                                             |
| Implementation authorized | `false`                                                                                    |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule)                                                 |

## Problem

Changing a governed task PR from draft to ready triggers a second full CI and
independent model review even when the base SHA and head SHA are unchanged and
the same exact SHA already has green required checks plus a trusted App-signed
PASS verdict.

Sanitized observed evidence from issue #872 (2026-08-21):

- PR #868, head `744e5845f2e8a9ec717fe0c38144dbd3b594559d`: initial pipeline
  run `32492018782` passed; ready-for-review run `32492586621` repeated full CI
  and review before merge.
- PR #869, head `fe3cf7d60157317e0564d2dd4de751bd6109a5df`: initial pipeline
  run `32493543324` passed with live-evidence attestation true; ready-for-review
  run `32494037984` repeated full CI and review before merge.

The duplicate path adds several minutes and repeats model/tool compute without
changing the reviewed code or evidence.

Drafting-time read of calling-repo `.github/workflows/pipeline.yml` shows
`ready_for_review` is intentionally subscribed so draft-aware merge-gate can
re-evaluate an unchanged SHA, but `ci`, `review`, and `plan-review` currently
run for every non-closed `pull_request` action — including `ready_for_review`.

## Required outcome (summary)

1. Preserve the rule that draft PRs never auto-merge.
2. On `ready_for_review`, reuse prior evidence only when the current base/head
   pair is unchanged, all required checks for that exact head are successful,
   and a trusted App-authored independent PASS verdict is bound to that exact
   base/head and task/package authority.
3. In that safe unchanged case, skip full CI and model review and run only the
   deterministic merge-gate re-evaluation needed to merge.
4. Fail closed and run the normal path when the head or base changed, required
   checks are missing/non-successful, the verdict is missing/waiting/failing/
   malformed/untrusted, live-evidence attestation is required but absent, or
   identity/scope metadata does not match.
5. Never treat a human/implementer comment as reusable verification authority.
6. Keep exact-SHA stale-run protections and independent-role separation.
7. Add deterministic shared-infra policy tests and calling-repository
   fixture/foundation coverage.
8. Record latency/compute evidence using run IDs and job conclusions only; do
   not copy logs or secrets.
9. Independently review, merge, and prove the optimized path on a controlled
   draft-to-ready transition before closing this issue.

## Tasks

| Task | Summary                                                                              | Depends on |
| ---- | ------------------------------------------------------------------------------------ | ---------- |
| T00  | Unchanged-SHA ready-for-review reuse gate, fail-closed semantics, docs, tests        | —          |
| T01  | Controlled draft-to-ready proof (operator-owned live evidence)                       | T00        |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Weaken draft-never-auto-merge, exact-SHA stale-run guards, or App-only verdict
  trust.
- Treat human or implementer comments as reusable verification authority.
- Address deprecated action inputs, Node runtime warnings, dependency alerts, or
  deterministic remediation preflight (separate focused roots per issue #872).
- Change production application behavior, signup policy, secrets, databases, or
  Kuma/synthetic inventory IDs.
- Self-adopt or self-authorize this package.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
