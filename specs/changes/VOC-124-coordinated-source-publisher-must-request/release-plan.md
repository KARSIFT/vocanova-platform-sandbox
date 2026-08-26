# VOC-124 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs the coordinated source-carrier App-token mint so an authorized
nested `karsift-ai-infra` commit that changes `.github/workflows/**` can be
pushed by the existing clean `publish-source` job.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 bootstrap source merge | implementer + independent verifier + separate merger | Adoption + task authorization; clean branch from current infra `main`; bounded `VOC-124-D04` scope; no VOC-122 bundle publication | Exact reviewed infra head and merge SHA; source self-CI |
| T00 normal caller merge | implementer + independent verifier | Bootstrap infra merge live on `implement.yml@main`; bootstrap authority exhausted; caller fixture pinned to exact merge | `VOC-124-EV-00` — token permission; caller isolation; exact SHAs |
| Post-merge effect | repository maintainers + existing VOC-122 roster | T00 infra merge live | Re-dispatch or reconcile existing `VOC-122-T00` / #1003 against that revision |
| Dependent #1003 / #1012 | existing VOC-122 roster, not a replacement task | Exact reviewed VOC-124 infra merge available | Newly verified VOC-122 source PR publishes, is independently reviewed, and merges; #1012 is then updated to that exact infrastructure merge SHA and remains this package's out-of-scope draft until that happens |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring intended source-carrier publication may allow an already-authorized
coordinated infrastructure PR to open and merge through existing gates after
genuine exact-SHA review. That is not a new deploy policy. Task #1003 already
failed after a valid bundle; this package prevents the next occurrence of
that missing-workflows-permission class and unblocks a later governed
re-dispatch of #1003.

No OpenAI credential or execution route is needed or authorized.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Authorized workflow-file source commits cannot be pushed; caller publisher gains workflows permission or stops refusing caller workflow files; missing credentials, invalid bundles, or stale leases start succeeding |
| Mechanism | Revert the T00 infra and caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored VOC-123 source-carrier workflows |
| Last-known-good | Source-carrier publication after VOC-123 infra merge `7500a4171d96a8e0d38889a9c92ad5dc092ad8dd`, before VOC-124. That revision still has the omitted `permission-workflows` defect; rollback is to reviewed state, not to a working nested workflow-file publisher |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that `publish-source` requests
   `permission-workflows: write` and that the caller `publish` mint does not.
3. Deterministic proof that an authorized `.github/workflows/**` source
   bundle is covered by that permission and is not rejected by the source
   publisher script.
4. Deterministic proof that missing App credentials, invalid bundles, stale
   bases, and stale leases still fail closed.
5. Confirmation that nested isolation, named-ref bundle tips, App-token split,
   force-with-lease, two-attempt bound, and non-closing source PR remain.
6. Confirmation that carrier current-state text matches active A-004 and that
   historical CHANGELOG records were not rewritten.
7. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed, or an explicit non-consumption note.
8. Recorded bootstrap exhaustion: exact reviewed head, merge SHA, separate
   merger, no direct `main` push, and no publication of VOC-122 nested head
   `f90eb630743c8c523e2e6e8dff017acbb31a7f43`.
9. Explicit record that #1003 / VOC-122 / #1012 is not implemented or merged
   here and remains the existing carrier to retry against the reviewed infra
   SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, the recorded evidence shows the #1013
workflows-permission class is fail-closed, the bootstrap exception is
exhausted, and the existing VOC-122 carrier can be retried through the
repaired path. Do not conflate package merge with runtime release; this
package has no direct product deployment effect. An issue-close event may
wake evaluation, but closed state alone is not completion proof.
Do not treat VOC-124 closure as VOC-122 / #1003 / #1012 completion until
that existing carrier has published, reviewed, and merged its own
authoritative infrastructure revision.
