# VOC-038 — Release Plan

## Release and deployment authorization

Not authorized by this package. A merged package does not itself authorize
production deployment. This package's tasks provision the production
infrastructure and its deploy workflow, but no task triggers a real production
deploy or points real DNS/traffic at the new target — that remains a separate,
later, explicitly founder-authorized action, consistent with `VOC-037-T05`'s own
role as "the R4 founder decision this whole package leads to" within `VOC-037`.
This package's own closure (all tasks `T00`-`T04` merged and independently
verified) is a precondition for `VOC-037-T03`/`T04` to proceed, not itself a
production release.

## Preconditions, monitoring, and outcome

Preconditions before any task's PR merges: this package adopted;
`scripts/governance/classify-change-risk.sh` run against the task's actual diff
with no unresolved higher-floor finding; independent verification (Claude Code)
passed at the exact reviewed commit SHA; the `production` GitHub Actions
environment's required-reviewer configuration inspected and confirmed correct
before `VOC-038-T02`'s workflow is allowed to merge (since a merged workflow with
a misconfigured environment gate is a live risk even before it is triggered).

Monitoring: not applicable to this package directly (this package does not
configure Sentry/uptime monitoring — that is `VOC-037-T04`'s scope, which depends
on this package). This package's own outcome evidence is the acceptance-criteria
results in `acceptance-criteria.md`, recorded per task as each merges.

Outcome owner: the package's implementer records each task's evidence artifact
(`VOC-038-EV-00` through `VOC-038-EV-04`) in that task's PR; the founder-gate
delegate or founder records this package's overall closure once all five tasks'
evidence is independently verified.

## Rollback

Trigger: any task found to fail its own acceptance criterion during or after
implementation, or a later discovery (e.g. by `VOC-037-T03`'s kill-switch
verification, which depends on this package) that the provisioned target does not
behave as this package's acceptance criteria describe.

Mechanism: git revert of the specific task's PR, plus (for `VOC-038-T00` only) a
recorded manual host-side step to delete the created directory tree and OS user —
see implementation-plan.md's "Deployment and rollback" section for the full
per-task breakdown. No task in this package requires a database migration
rollback, since no task creates or migrates data.

Accountable owner: the task's implementer, with independent verification
confirming rollback success.

Validation: re-run this package's own `VOC-038-TEST-00` through `VOC-038-TEST-04`
against the post-rollback state and confirm the checks that previously passed for
staging (none of which this package should have touched) still pass, proving the
rollback did not regress staging.

Last-known-good reference: this package's `base_sha`, recorded in `change.yaml` at
adoption time.

## Independent verification, human approvals, and closure

Independent verification: Claude Code, per CLAUDE.md, at the exact merged SHA of
each task's PR. Verification must explicitly confirm Codex (or whichever
implementer) did not approve or merge its own implementation, identify the active
authority model (A-003, effective since 2026-07-17T16:44:34Z), and report every
still-required gate.

Required human approvals: under active A-003, this package's tasks are routine R3
production-infrastructure/CI-CD work and do not require standing technical-steward
or founder approval solely because they are R3 — strengthened applicable controls
and independent verification (above) are required instead. This package's two
open questions (specification.md: production hostname placeholder, shared-host
resource-limit values) are implementer/reviewer-time decisions within R3's scope,
not new R4 matters, and do not themselves require founder sign-off beyond the
independent verification already required for R3. If any task's actual diff or a
discovered fact raises the class to R4 (e.g. if resolving the resource-limit
question requires a spend decision, or if the hostname confirmation surfaces a
material product-direction question), founder approval becomes required for that
specific task before it may merge, per the classifier's own re-run requirement in
implementation-plan.md.

Remaining hosted controls after this package closes: `VOC-037-T03` (kill-switch
and rollback verification against this package's now-real target) and
`VOC-037-T04` (monitoring/alerting readiness against the same target) remain
required before `VOC-037-T05`'s R2 release PR and founder go/no-go decision can
proceed — this package's closure does not itself constitute R2 readiness, only one
of its preconditions.

Closure evidence: all five tasks' acceptance criteria recorded `pass` in
`acceptance-criteria.md`, each with its evidence artifact linked, and independent
verification's `PASS` or `PASS WITH NON-BLOCKING FINDINGS` result recorded against
the exact final merged SHA of this package's last task.
