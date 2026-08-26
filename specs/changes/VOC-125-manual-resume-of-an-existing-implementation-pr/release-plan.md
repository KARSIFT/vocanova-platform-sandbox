# VOC-125 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs the documented operator `action=implement` contract so an existing
implementation PR can be resumed after automatic remediation stops, by
deriving immutable recovery identity from that PR and binding verified
`expected_head_sha` / `expected_base_sha` before any model or mutation step.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 source merge | implementer + independent verifier + separate merger | Adoption + task authorization; clean branch from current infra `main`; no bootstrap exception | Exact reviewed infra head and merge SHA; source self-CI; `implement.yml@main` declares `existing_pr_number` |
| T00 caller merge | implementer + independent verifier | Infra merge live on `implement.yml@main`; caller dispatch forwards `existing_pr_number`; fixture pinned to exact merge when consumed | `VOC-125-EV-00` — recovery identity; mismatch fail-closed; exact SHAs |
| Post-merge effect | repository maintainers + existing VOC-122 roster | T00 infra and caller dispatch live | Resume existing `VOC-122-T00` / #1003 / #1012 at attempt `2` with `existing_pr_number=1012` |
| Dependent #1003 / #1012 | existing VOC-122 roster, not a replacement task | Exact reviewed VOC-125 infra merge and caller dispatch available | Existing carrier resumes through the repaired route; #1012 remains this package's out-of-scope draft until VOC-122's own work completes |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring intended operator resume may allow an already-authorized
implementation PR to continue through existing gates after genuine
exact-SHA review. That is not a new deploy policy. Task #1003 already failed
before model resolution; this package prevents the next occurrence of that
empty-binding class and unblocks a later governed resume of #1003.

No OpenAI credential or execution route is needed or authorized.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Operator resume cannot bind exact head/base; caller dispatch accepts free-form SHAs; attempt `1` rewrites an existing carrier; attempt `3` is accepted; mismatch classes succeed; automatic remediate loses event-derived SHAs |
| Mechanism | Revert the T00 infra and caller workflow/fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored VOC-124 implement/dispatch workflows |
| Last-known-good | Implement/dispatch after VOC-124 infra merge `f406cc95a3f853e8aef5bf8bcf22d37a29d64547`, before VOC-125. That revision still has the missing operator-resume identity defect; rollback is to reviewed state, not to a working existing-carrier resume |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that caller/template pipeline expose and forward
   `existing_pr_number` and do not expose operator SHA inputs.
3. Deterministic proof that a valid attempt-2 resume derives and binds exact
   head/base before model resolution and reuses the existing carrier.
4. Deterministic proof that the #1020 empty-binding class and every
   `VOC-125-D03` mismatch class fail closed.
5. Deterministic proof that attempt `1` with an existing open carrier and
   attempt `3` fail closed, and that `remediate.yml` still forwards
   event-derived SHAs.
6. Confirmation that nested isolation, named-ref bundle tips, App-token split,
   force-with-lease, two-attempt bound, Cursor Composer/Grok roles, and
   non-closing source PR remain.
7. Confirmation that carrier current-state text describes attempt `2` plus
   `existing_pr_number` and that historical CHANGELOG records were not
   rewritten.
8. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed, or an explicit non-consumption note.
9. Explicit record that #1003 / VOC-122 / #1012 is not implemented or merged
   here and remains the existing carrier to resume at attempt `2` with
   `existing_pr_number=1012`.
10. Confirmation that no bootstrap exception was used.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, the recorded evidence shows the #1020
empty-binding class is fail-closed, and the existing VOC-122 carrier can be
resumed through the repaired path. Do not conflate package merge with runtime
release; this package has no direct product deployment effect. An issue-close
event may wake evaluation, but closed state alone is not completion proof.
Do not treat VOC-125 closure as VOC-122 / #1003 / #1012 completion until
that existing carrier has resumed, published, reviewed, and merged its own
authoritative revision.
