# VOC-122 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs promotion-check recovery so a required pull-request row that
appears after the invocation's first metadata snapshot can be recovered in
that same bounded wait.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 coordinated source + caller merge | implementer + independent verifier | Adoption + task authorization; reviewed shared-infra PR precedes caller fixture pin when consumed | `VOC-122-EV-00` — replan during polling; once-per-run-ID / once-per-absent-context; exact SHAs |
| Post-merge effect | repository maintainers + future release recovery jobs | T00 merged to integration branch | A later-appearing cancelled/failed required row is recovered in the same invocation instead of waiting for a subsequent recovery job or 1800-second timeout |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring intended promotion recovery may allow an already-reviewed promotion
PR to merge through existing gates after genuine exact-head checks. That is
not a new deploy policy. Promotion PR #1000 already merged via a later
convergence job; this package prevents the next occurrence of that first-job
timeout class.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Replan duplicates dispatches or reruns; fails closed on legitimate pending rows; or treats alternate success as satisfying a selected cancelled row |
| Mechanism | Revert the T00 infra and caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored VOC-121 recovery behavior |
| Last-known-good | Promotion recovery after VOC-121 infra merge `99476c2a1018e42d4bd442657b5257885ac9f1c9`, before VOC-122 |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that an absent-then-cancelled required row is rerun in
   the same recovery invocation.
3. Deterministic proof that run IDs and absent-context dispatches are
   once-per-invocation, and that alternate success cannot override GitHub's
   selected row.
4. Deterministic proof that ambiguous, foreign, and mismatched later snapshots
   fail closed.
5. Confirmation that the 1800-second timeout, App-token split, exact-head
   merge decision, first-attempt identity, and `integration_push` semantics
   remain.
6. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed, or an explicit non-consumption note.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, and the recorded evidence shows the #1000
first-job timeout class is fail-closed. Do not conflate package merge with
runtime release; this package has no direct product deployment effect. An
issue-close event may wake evaluation, but closed state alone is not
completion proof.
