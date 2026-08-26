# VOC-123 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs coordinated source-carrier Git-bundle creation so an authorized
nested `karsift-ai-infra` commit can be advertised through a named bundle
tip and published by the existing clean `publish-source` job.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 bootstrap source merge | implementer + independent verifier + separate merger | Adoption + task authorization; clean branch from current infra `main`; bounded `VOC-123-D08` scope | Exact reviewed infra head and merge SHA; source self-CI |
| T00 normal caller merge | implementer + independent verifier | Bootstrap infra merge live on `implement.yml@main`; bootstrap authority exhausted; caller fixture pinned to exact merge | `VOC-123-EV-00` — named-ref bundle tip; advertised-head proof; exact SHAs |
| Post-merge effect | repository maintainers + future implement jobs | T00 merged to integration branch | A nested source commit produces a non-empty source bundle and can open the infrastructure carrier PR through existing gates |
| Dependent #1003 | existing VOC-122 roster, not this package | Exact reviewed VOC-123 infra merge available | Re-dispatch or reconcile `VOC-122-T00` against that revision; do not treat this package as VOC-122 completion |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

Restoring intended source-carrier publication may allow an already-authorized
coordinated infrastructure PR to open and merge through existing gates after
genuine exact-SHA review. That is not a new deploy policy. Task #1003 already
failed before any carrier PR; this package prevents the next occurrence of
that empty-bundle class and unblocks a later governed re-dispatch of #1003.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Source bundle is empty; advertises the wrong head; publishes unrelated refs; or previously working caller/planner `..HEAD` bundles become empty or unsafe |
| Mechanism | Revert the T00 infra and caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored VOC-121 source-carrier workflows |
| Last-known-good | Source-carrier publication after VOC-121 infra merge `99476c2a1018e42d4bd442657b5257885ac9f1c9`, before VOC-123. That revision still has the raw-SHA empty-bundle defect; rollback is to reviewed state, not to a working nested publisher |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that a raw-SHA positive tip still reproduces
   empty-bundle, and that the fixed source-carrier path advertises exactly
   the expected committed head.
3. Deterministic proof that wrong/missing/multiple advertised heads, wrong
   base, malformed SHA, cleanup/publish mismatch, and unrelated refs fail
   closed.
4. Deterministic proof that caller and planner `..HEAD` paths are safe or
   were repaired with equivalent coverage.
5. Confirmation that nested isolation, App-token split, force-with-lease,
   two-attempt bound, and non-closing source PR remain.
6. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt` when
   consumed, or an explicit non-consumption note.
7. Explicit record that #1003 / VOC-122 is not implemented here and remains
   a distinct re-dispatch against the reviewed infra SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, and the recorded evidence shows the #1003
empty-bundle class is fail-closed. Do not conflate package merge with runtime
release; this package has no direct product deployment effect. An issue-close
event may wake evaluation, but closed state alone is not completion proof.
Do not treat VOC-123 closure as VOC-122 / #1003 completion.
