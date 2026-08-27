# VOC-127 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It repairs `release.yml` so a successful governed `--merge` promotion advances
`develop` to that exact merge SHA before audit close, makes that sync
idempotently recoverable through `reconcile-release`, and adds an exceptional
adopted exact-SHA path for main-only hotfix/evidence reconciliation.

Adoption and task implementation authorization remain separate gates under
active A-004.

This package is not a request to promote the current integration tip to
production. Both refs were already aligned by the 2026-08-27 settings repair.
Do not treat merge of this package as a substitute for that repair, and do
not add a snapshot-then-drift task (`karsift-ai-infra#15`).

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 source merge | implementer + independent verifier + separate merger | Adoption + task authorization; clean branch from current infra `main`; no bootstrap exception | Exact reviewed infra head and merge SHA; source self-CI; `release.yml` binds and advances `develop` to `mergeCommit.oid` before audit close |
| T00 caller merge | implementer + independent verifier | Infra merge live; exceptional dispatch present; staging skip proven; fixture pinned when consumed | `VOC-127-EV-00` — bind-and-sync; fail-closed races; pin SHA |
| Post-merge promotion | existing `release.yml` | T00 infra and caller changes live on `develop` and promoted | The promotion of *this* package must itself leave `develop` at that promotion's merge SHA once the new `release.yml@main` is what later promotions call. The first caller merge still uses pre-change `release.yml@main` until this package is promoted; that limitation is expected and must be recorded in evidence rather than worked around with a settings rewrite |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

## Rollback

| Item | Value |
|------|-------|
| Trigger | `develop` left at checked head after `--merge`; unique develop commits erased; second promotion PR or `gh pr merge`; operator SHA paste; unadopted main-only sync; tree-equivalent sync fully deploys staging; `roles.yml` changed |
| Mechanism | Revert the T00 infra and caller workflow/helper/fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored `release.yml` that preserves `CHECKED_HEAD_SHA` |
| Last-known-good | Infra `release.yml` before this package's infra merge, and caller pin `60afda3a44fd06b8c00b219771de7112f1aded6e` at drafting. That last-known-good still leaves `develop` behind merge-commit promotions; rollback is to reviewed state, not to a fully converged post-merge develop. Do not repeat the 2026-08-27 settings recreation as rollback |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR and for
   any coordinated infra PR. The implementer must not approve or merge its own
   work on either carrier.
2. Deterministic proof that after `--merge`, `develop` is advanced to
   `mergeCommit.oid` before audit close, and that the promotion PR/merge
   commit remain.
3. Deterministic proof that unique develop commits, moved `main`, malformed
   merges, missing refs, and unmerged closed PRs fail closed without erasing
   work.
4. Deterministic proof that `reconcile-release` on an already-merged PR with
   an open audit syncs `develop` without a new PR or second merge, and that
   equal tips do not create a release loop.
5. Deterministic proof that a missing integration ref is recreated at the
   merge SHA.
6. Deterministic proof that exceptional main-only reconciliation requires an
   adopted package and merged main-targeting PR number, rejects operator SHAs,
   and is not the ordinary workflow.
7. Deterministic proof that tree-equivalent develop sync does not keep
   staging scheduled, and that VOC-111 allowlisted paths still do.
8. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, and non-closing source PR remain.
9. Confirmation that current-state docs describe exact-merge-SHA sync and
   exceptional main-only reconciliation, and that historical CHANGELOG /
   A-003 / VOC-075 records were not rewritten.
10. Recorded pin applicability: exact infra merge SHA in `PINNED_SHA.txt`
    when consumed (not `60afda3a…`), or an explicit non-consumption note.
11. Confirmation that no snapshot-gap task and no bootstrap exception were
    used.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification on each carrier, and the recorded evidence shows the #1033
checked-head restore class is fail-closed. Do not conflate package merge with
runtime release; this package has no direct product deployment effect. An
issue-close event may wake evaluation, but closed state alone is not
completion proof.
