# VOC-104 — Skip redundant ready-for-review CI and model review on unchanged exact SHA

| Field                     | Value                                                                                |
| ------------------------- | ------------------------------------------------------------------------------------ |
| Package                   | `VOC-104`                                                                            |
| Title                     | Skip redundant ready-for-review CI and model review on unchanged exact SHA           |
| Path                      | `specs/changes/VOC-104-ready-for-review-reruns-unchanged-exact-sha-ci`               |
| Status                    | `adopted`                                                                            |
| Risk                      | `R4` (DOC-15 §17.3 path floor; each task diff is reclassified independently)         |
| Authority model           | A-004 active                                                                         |
| Requirement source        | GitHub issue [#872](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/872) |
| Target branch             | `develop`                                                                            |
| Approval                  | `autonomously-adopted-after-independent-verification`                                |
| Implementation authorized | `true`                                                                               |
| `automatic_merge_allowed` | `true`                                                                               |

## Problem

Changing a governed task PR from draft to ready triggers a second full CI and
independent model review even when the base SHA and head SHA are unchanged and
the same exact SHA already has green required checks plus a trusted App-signed
PASS verdict.

Sanitized observed evidence from issue #872 (2026-08-21):

- PR #868, head `744e5845f2e8a9ec717fe0c38144dbd3b594559d`: initial pipeline run
  `32492018782` passed; ready-for-review run `32492586621` repeated full CI and
  review before merge.
- PR #869, head `fe3cf7d60157317e0564d2dd4de751bd6109a5df`: initial pipeline run
  `32493543324` passed with live-evidence attestation true; ready-for-review run
  `32494037984` repeated full CI and review before merge.

The duplicate path adds several minutes and repeats model/tool compute without
changing the reviewed code or evidence.

## Required outcome (summary)

1. Preserve the rule that draft PRs never auto-merge.
2. On `ready_for_review`, reuse prior evidence only when the current base/head
   pair is unchanged, all required checks for that exact head are successful, and
   a trusted App-authored independent PASS verdict is bound to that exact
   base/head and task/package authority.
3. In that safe unchanged case, run the eligibility decision, emit the required
   `ci / ci` context through a successful reuse marker, skip checkout/full
   application validation and model review, then run deterministic merge-gate
   re-evaluation.
4. Fail closed and run the normal path when the head or base changed, required
   checks are missing/non-successful, the verdict is missing/waiting/failing/
   malformed/untrusted, live-evidence attestation is required but absent, or
   identity/scope metadata does not match.
5. Never treat a human/implementer comment as reusable verification authority.
6. Keep exact-SHA stale-run protections and independent-role separation.
7. Add deterministic shared-infra policy tests and calling-repository
   fixture/foundation coverage.
8. Record latency/compute evidence using run IDs and job conclusions only; do not
   copy logs or secrets.
9. Independently review, merge, and prove the optimized path on a controlled
   draft-to-ready transition before closing this issue.

## Tasks

| Task | Summary                                                                       | Depends on |
| ---- | ----------------------------------------------------------------------------- | ---------- |
| T00  | Ready-for-review reuse policy, fail-closed path, docs, deterministic tests    | —          |
| T01  | Controlled draft-to-ready optimized-path proof (operator-owned live evidence) | T00        |

See `tasks.md` for full task definitions.

## What this package deliberately does NOT do

- Allow draft PRs to auto-merge.
- Reuse human, implementer, or non-App review comments as verification authority.
- Weaken exact-SHA stale-run guards, independent-role separation, or merge-gate
  fail-closed behavior.
- Address deprecated action inputs, Node runtime warnings, dependency alerts, or
  deterministic remediation preflight (explicit separate focused roots per #872).
- Change application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID behavior.
- Change the package authority recorded by the governed adoption workflow.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. The governed plan-review/adoption workflow recorded this
package's implementation authority before T00 began.
