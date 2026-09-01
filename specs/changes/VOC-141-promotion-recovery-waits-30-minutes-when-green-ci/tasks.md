# VOC-141 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #1109 is one
recovery-dispatch repair outcome: when GitHub-required `ci / ci` is SUCCESS
and composed evidence is unattestable, immediately dispatch dedicated
`promotion-pr-validation`, name that reason on timeout, pin the new
infrastructure merge, update current-state docs, and let ordinary
`reconcile-release` proceed without a 1,800-second hang. Repository count,
dispatch-versus-diagnostics, workflow-versus-tests-versus-docs, and
pin-versus-release are not split reasons.

Cross-repo note: coordinated PRs in `KARSIFT/karsift-ai-infra` and this
caller remain one task. Infrastructure merge
`67bdfd13ef875dead23ce4be01d7d0e8b976e289` is the defective live contract
for this failure class; T00 must open a new infra PR and pin its exact
merge. Do not treat the untracked local `karsift-ai-infra/` checkout (if
present) as this repo's tracked tree. No bootstrap exception. T00's first
run is attempt `1` on a new VOC-141 carrier from current `develop`. Do not
reuse PR #1090 as this package's implementation PR.

This is not a "promote current develop to main" package and not a snapshot of
the develop/main gap (`karsift-ai-infra#15`). Workaround runs `33341923799`
and `33342062118` are incident evidence that dedicated dispatch unblocks this
class; they are not a substitute for the code repair. A VOC-097 live-evidence
second task is not used: an evidence-carrier PR cannot satisfy sha_lineage
against a live promotion HEAD.

## VOC-141-T00 — Dispatch dedicated promotion-pr-validation immediately when green CI is unattestable

- Requirement source: issue #1109; `VOC-141-D00` through `VOC-141-D15`
- Acceptance criteria: `VOC-141-AC-00` through `VOC-141-AC-08`
- Tests: `VOC-141-TEST-00` through `VOC-141-TEST-08`
- Post-T00 release/closure verification: `VOC-141-TEST-09` (not an
  implementation-task close or retry prerequisite)
- Evidence: `VOC-141-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #1109 evidence in `t00-evidence.md` (PR #1090, head
   `c3a53bab3035b7f08c0fb959bdf1b56bf330d291`, pin
   `67bdfd13ef875dead23ce4be01d7d0e8b976e289`, release carrier
   `33340381776`, job `99334840338`, timeout diagnostics
   `missing_checks: none` / `successful: 6`, duplicate carrier
   `33340516672`, workaround dedicated run `33341923799`, subsequent
   reconcile-release `33342062118`). Do not copy raw provider responses,
   credentials, or the full log.
2. Resolve current `develop` to a 40-character SHA **before any in-scope
   edit**. Record that SHA as the implementation PR base at PR creation. Fail
   closed on unrelated/material movement of `develop`. This package's own
   plan/adoption/roster commits after issue-creation head `c3a53bab…` do
   not count as protected-file drift.
3. Open a new `KARSIFT/karsift-ai-infra` PR. Implement D01–D07 there:
   `apply_promotion_pr_recovery_plan` must consult composed attestable CI,
   not only `plan_required_check_recovery`; SUCCESS-plus-unattestable-parent
   immediately dispatches exactly one `recover-promotion-pr-checks`; completed
   dedicated parents complete without redispatch; active/successful exact
   dedicated recovery suppresses duplicates; release carriers do not; timeout
   diagnostics name unattestable CI evidence. Do not rerun doomed
   `pull_request` `ci / ci` jobs. Do not weaken production/release-carrier
   fail-closed boundaries or the VOC-140 two-token guard.
4. Add deterministic tests that call the live planner path
   (`apply_promotion_pr_recovery_plan` or the recovery loop) with GitHub-required
   SUCCESS rows. Cover the five issue classes. Do not treat helper-only
   `suppress_active_or_successful_dispatches` tests that omit SUCCESS required
   rows as covering #1109. If that helper is reused, prove its SUCCESS/PENDING
   filter no longer drops dedicated dispatch for unattestable CI.
5. Obtain independent exact-revision review of the infra PR and merge it.
   Record the exact merge SHA. From current caller `develop`, create a new
   VOC-141 implementation branch. Pin `PINNED_SHA.txt` to that infra merge
   and mirror every changed authoritative fixture file. Update fixture
   README and all current-state docs found by exhaustive tracked-source search,
   including `docs/operations/11-devops-and-ci-cd.md`. Do not rewrite
   VOC-140, VOC-139, or VOC-138 package records. Reconcile every live caller
   pin-lock/current-pin assertion and mirrored-file hash table with the new
   merge. Preserve issue-era `AUTHORITATIVE_PIN` values and historical
   package evidence; split current versus authoritative constants where a
   test currently conflates them.
6. Confirm fixture `roles.yml` is unchanged and no OpenAI route is added.
   Confirm no fabricated-status helper, bypass-actor addition, App-token mint
   change, or test-time evidence mutation was added.
7. Record in `t00-evidence.md` the implementation PR base, new infra merge,
   dispatch-when-unattestable change, timeout-diagnostic token, negative-case
   results, source-search disposition, pin advance, validation commands after
   commit, and the feasible exact-head binding contract. Evidence must not
   require a commit to contain its own SHA.
8. Run applicable validation **after the repair is tracked and committed**
   and record results in `t00-evidence.md`:
   - `bash scripts/governance/validate-governance.sh` with exact base/head;
   - `bash scripts/governance/classify-change-risk.sh` with exact base/head
     (expect R4);
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'`;
   - targeted VOC-141 unattestable-SUCCESS dispatch, duplicate-suppression,
     timeout-diagnostic, and unchanged carrier fail-closed cases;
   - `git diff --check`;
   - hosted required checks and independent exact-revision review that
     explicitly evaluates SUCCESS-plus-unattestable-parent dispatch,
     completed dedicated parent without redispatch, active/successful exact
     dedicated suppression, timeout unattestable-CI diagnostics, unchanged
     production/release-carrier fail-closed boundaries, no doomed `ci / ci`
     rerun, and pin advance.
9. This package's caller PR `Closes` only its own VOC-141 task issue. Do
   not snapshot the current develop/main gap. After the exact reviewed
   caller merge, ordinary `reconcile-release` may merge the live promotion at
   the then-current `develop` head and converge `develop` to that exact
   merge SHA. Root issue #1109 closes only after allowlisted metadata from
   a successful recovery/release run exists. Do not create a duplicate
   promotion PR or release audit.

10. Preserve independent exact-SHA review, risk classification, protected
    checks, and App-token isolation. The independent review comment must bind
    the live PR head exactly. Merge-gate must reject any mismatch.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider,
  `roles.yml`, or monitor-inventory changes.
- Weakening the production merge guard, adding bypass actors, fabricating
  statuses, or manually merging a promotion PR.
- Changing the VOC-140 two-token contract or App-token mints.
- Creating a duplicate promotion PR or release audit issue.
- Switching the promotion PR application check to `--squash-safe-push`.
- Rerunning doomed `pull_request` `ci / ci` jobs as a strategy.
- Repairing GitHub reusable-job cancellation of sleeping recovery runners.
- Requiring a commit to contain its own SHA.
- Snapshotting the current develop/main gap, or treating "promote current
  develop" as this package's work rather than as post-repair release
  evidence.
- A VOC-097 operator-owned live-evidence second task.
- OpenAI credentials or execution routes.
- A supervised bootstrap exception.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting dispatch, diagnostics, pin, tests, docs, or release
  reconciliation into separate tasks.
- Rewriting VOC-140, VOC-139, or VOC-138 package records.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the infra contract, caller pin, unattestable-SUCCESS dispatch,
  timeout diagnostics, tests, docs, evidence, and release handoff are one
  repair outcome.
- Infrastructure merge `67bdfd13…` is already merged and is the defective
  live contract; consume a new independently reviewed infra merge inside this
  same task.
- Workaround dedicated run `33341923799` and reconcile-release `33342062118`
  are incident proof, not a second VOC-141 task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
