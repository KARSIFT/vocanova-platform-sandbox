# VOC-129 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the remaining VOC-127 caller-side release-synchronization contract
by consuming already-merged infrastructure #164
(`863fc1f35b1d35e4981a59166b0e939be1a2b681`) on a new VOC-129 carrier, and by
superseding unpublishable PR #1041 without a third VOC-127 attempt.

Adoption and task implementation authorization remain separate gates under
active A-004.

This package is not a request to promote the current integration tip to
production as its work. The replacement proceeds through normal `develop`
merge, `develop` → `main` promotion, exact post-promotion `develop`
convergence, and applicable deployment. Tree-equivalent convergence must not
trigger an unnecessary staging deployment. Closed state alone is not
completion proof.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; infra #164 already merged as `863fc1f…`; new VOC-129 branch from current `develop`; #1041 not used | `VOC-129-EV-00` — pin SHA; checkout-ref coverage; live `reconcile-production-change`; exact reviewed caller SHA |
| Post-merge promotion | existing `release.yml@main` (#164) | T00 caller changes live on `develop` | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end at the same SHA |
| Staging | VOC-111 path selection | Tree-equivalent develop sync or specs-only paths | No unnecessary staging deployment; allowlisted runtime/deploy paths still deploy |
| Supersede #1041 and close VOC-127 | implementer evidence + maintainers | VOC-129 caller merge reviewed, merged, and promoted | #1041 closed as superseded (never merged); #1039 and #1035 closed with audit comments naming the VOC-129 merge; no VOC-127 completion marker bound to #1041 |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Pin not equal to `863fc1f…`; #1041 published; VOC-127 attempt `3`; unique develop commits erased; operator SHA paste; tree-equivalent sync fully deploys staging; `roles.yml` changed |
| Mechanism | Revert the T00 caller workflow/fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` pin `60afda3a44fd06b8c00b219771de7112f1aded6e` |
| Last-known-good | Caller `develop` before this package's merge (pin `60afda3a…`). That last-known-good still lacks the #164 caller contract; rollback is to reviewed state, not to a working #164 pin. Do not restore PR #1041. Do not revert infrastructure #164 |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. Infrastructure #164 is
   already independently reviewed at merge
   `863fc1f35b1d35e4981a59166b0e939be1a2b681`; this package does not reopen
   that review except to confirm the caller consumed that exact SHA.
2. Deterministic proof that `PINNED_SHA.txt` and live pin assertions equal
   `863fc1f…` and do not equal `a9df74a6…` or `60afda3a…`.
3. Deterministic proof that the #164 checkout-ref ordering / missing-`develop`
   path is present in the pinned fixture and covered by tests.
4. Deterministic proof that live `pipeline.yml` exposes
   `reconcile-production-change`, stays at most 25 inputs, and does not
   expose operator SHAs.
5. Deterministic proof that fixture `release.yml` no longer treats
   `CHECKED_HEAD_SHA` restoration as successful post-merge sync.
6. Deterministic proof that tree-equivalent develop sync does not keep
   staging scheduled, and that VOC-111 allowlisted paths still do.
7. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, and non-closing source PR remain.
8. Confirmation that current-state docs describe exact-merge-SHA sync and
   `reconcile-production-change`, and that historical CHANGELOG / A-003 /
   VOC-075 / VOC-127 records were not rewritten.
9. Explicit record that #1041 is not merged; that VOC-127 was not dispatched
   as attempt `3`; and that #1039 / #1035 close only after this package is
   promoted, without a VOC-127 completion marker bound to #1041.
10. Confirmation that no snapshot-gap task and no bootstrap exception were
    used.
11. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge; tree-equivalent convergence did not trigger an
    unnecessary staging deployment; release/task/requirement records close
    in the order above.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, then promotes through the existing #164 `release.yml@main`
path. Do not conflate package merge with runtime release; this package has
no direct product deployment effect. An issue-close event may wake
evaluation, but closed state alone is not completion proof.
