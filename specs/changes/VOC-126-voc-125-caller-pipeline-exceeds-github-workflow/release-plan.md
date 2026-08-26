# VOC-126 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs the caller and project-repo `workflow_dispatch` contracts so GitHub
will accept VOC-125's `existing_pr_number` operator interface without deleting
an active recovery or verifier capability, by relocating the coherent
read-only verifier dispatch surface into a dedicated caller workflow.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 source merge | implementer + independent verifier + separate merger | Adoption + task authorization; clean branch from current infra `main`; no bootstrap exception | Exact reviewed infra head and merge SHA; source self-CI; project-repo template GitHub-valid under 25 inputs with `existing_pr_number` still present |
| T00 caller merge | implementer + independent verifier | Infra merge live; live caller `pipeline.yml` forwards `existing_pr_number` and has ≤25 inputs; dedicated verifier workflow present; fixture pinned to exact merge when consumed | `VOC-126-EV-00` — input count; relocated verifiers; exact SHAs |
| Post-merge promotion | repository maintainers | T00 infra and caller dispatch live on `develop` and promoted | Live caller route is GitHub-valid |
| Supersede #1024 and close VOC-125 | implementer evidence + maintainers | Live caller route valid, reviewed, merged, and promoted | #1024 closed as superseded; #1022 / #1020 closed without a VOC-125 completion marker bound to #1024 |
| Dependent #1003 / #1012 | existing VOC-122 roster, not a replacement task | Exact reviewed VOC-126 infra merge and caller dispatch available | Existing carrier resumes through the repaired `pipeline.yml` route; #1012 remains this package's out-of-scope draft until VOC-122's own work completes |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring a GitHub-valid caller dispatch may allow an already-authorized
implementation PR to continue through existing gates after genuine
exact-SHA review. That is not a new deploy policy. PR #1024 never entered
pipeline review; this package prevents the next occurrence of that
definition-invalid class and unblocks a later governed resume of #1003.

No OpenAI credential or execution route is needed or authorized.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Any live `workflow_dispatch` exceeds 25 inputs; `existing_pr_number` missing; a verifier or recover capability missing; verifier workflow gains secrets or `actions: write`; operator SHA paste becomes the caller interface; attempt `3` accepted |
| Mechanism | Revert the T00 infra and caller workflow/fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored GitHub-valid 25-input caller `pipeline.yml` (without `existing_pr_number`) plus the last reviewed `implement.yml` bind contract |
| Last-known-good | Live caller `pipeline.yml` on `develop` before this package lands, and `implement.yml` after VOC-125 infra merge `1f1705d…`. Do not restore PR #1024 / `8621f12…`; that revision is GitHub-invalid. That last-known-good still lacks a GitHub-valid `existing_pr_number` caller interface; rollback is to reviewed state, not to a working existing-carrier resume |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that every live and template `workflow_dispatch` block
   has at most 25 inputs.
3. Deterministic proof that caller/template `pipeline.yml` expose and forward
   `existing_pr_number` and do not expose operator SHA inputs.
4. Deterministic proof that the five read-only verifier jobs exist on the
   dedicated workflow with the same reusable calls and named inputs, and that
   mutating recover/reconcile/implement/plan actions remain on `pipeline.yml`.
5. Deterministic proof that the dedicated verifier workflow is read-only.
6. Confirmation that nested isolation, named-ref bundle tips, App-token split,
   force-with-lease, two-attempt bound, Cursor Composer/Grok roles, and
   non-closing source PR remain.
7. Confirmation that carrier current-state text describes attempt `2` plus
   `existing_pr_number` on `pipeline.yml` and verifier dispatch on the
   dedicated workflow, and that historical CHANGELOG records were not
   rewritten.
8. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed (not `1f1705d…`), or an explicit non-consumption note.
9. Explicit record that #1024 is not merged; that #1022 / #1020 close only
   after the live route is valid, without a VOC-125 completion marker bound to
   #1024; and that #1003 / VOC-122 / #1012 is not implemented or merged here
   and remains the existing carrier to resume at attempt `2` with
   `existing_pr_number=1012`.
10. Confirmation that no bootstrap exception was used and VOC-125 was not
    dispatched as attempt `3`.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, the recorded evidence shows the #1025
26-input class is fail-closed, the live caller route is GitHub-valid, #1024
is superseded rather than merged, and the existing VOC-122 carrier can be
resumed through the repaired path. Do not conflate package merge with runtime
release; this package has no direct product deployment effect. An issue-close
event may wake evaluation, but closed state alone is not completion proof.
Do not treat VOC-126 closure as VOC-122 / #1003 / #1012 completion until
that existing carrier has resumed, published, reviewed, and merged its own
authoritative revision.
