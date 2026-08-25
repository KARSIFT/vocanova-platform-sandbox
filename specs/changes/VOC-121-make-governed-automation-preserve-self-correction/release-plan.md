# VOC-121 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs governed implementation and promotion recovery so future adopted
tasks can finish coordinated carriers, retain self-correction helpers, and
recover required checks from GitHub's actual satisfaction state.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 coordinated source + caller merge | implementer + independent verifier | Adoption + task authorization; reviewed shared-infra PR precedes caller fixture pin when consumed | `VOC-121-EV-00` — no silent nested-edit loss; helpers preserved; cancelled-check recovery; exact SHAs |
| Post-merge effect | repository maintainers + future implement/release runs | T00 merged to integration branch | Coordinated tasks no longer require manual infra recovery or helper restoration; promotion recovery redispatches cancelled required check-runs |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring intended promotion recovery may allow an already-reviewed promotion
PR to merge through existing gates after genuine exact-head checks. That is
not a new deploy policy.

## Rollback

| Item | Value |
|------|-------|
| Trigger | False-positive fail-loud blocks legitimate caller-only work; second-carrier publication is unsafe; self-correction cannot find helpers; or recovery redispatches/attests required checks incorrectly |
| Mechanism | Revert the T00 infra and caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored workflows |
| Last-known-good | Implementer staging/self-correction and promotion recovery behavior before VOC-121 implementation |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that authorized nested edits are published in isolation
   or fail loudly, never silently discarded.
3. Deterministic proof that self-correction helpers survive nested-checkout
   removal.
4. Deterministic proof that a cancelled required check-run is recovered on the
   unchanged exact head despite an alternate successful run or same-named
   status.
5. Confirmation that exact-SHA review, risk floors, protected checks,
   App-token isolation, two-attempt limits, and Cursor fail-closed auth remain,
   and that no OpenAI/Codex path was introduced.
6. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed, or an explicit non-consumption note.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, and the recorded evidence shows the three live
failure classes are fail-closed. Do not conflate package merge with runtime
release; this package has no direct product deployment effect. An issue-close
event may wake evaluation, but closed state alone is not completion proof.
