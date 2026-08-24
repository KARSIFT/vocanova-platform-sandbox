# VOC-115 — Reduce plan and task fragmentation in governed engineering workflow: Specification

## Objective and requirement source

Reduce orchestration overhead in the governed engineering lifecycle without reducing
review rigor. A coherent user objective, feature, or bug should normally move through
the lifecycle as one change package and one end-to-end implementation task unless a
real boundary requires otherwise.

**Requirement source:** [GitHub issue #962](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/962).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004-governed decision.

## Scope and non-goals

### In scope

1. Update canonical operating-model/governance documentation so it states the new
   default: one coherent objective maps to one change package, and one new package
   maps to one end-to-end implementation task unless a concrete split boundary exists.
2. Update planner prompt text and mirrored fixture copies so the drafting default is
   one package / one task rather than "small, ordered" decomposition.
3. Define when a causally related defect found during implementation, verification,
   merge, promotion, or reconciliation may remain under the active package/carrier:
   only when it stays within the original objective, acceptance criteria, risk
   ceiling, and protected-area scope.
4. Define when a new issue/plan remains required: unrelated scope, changed product
   intent, expanded authority/risk boundary, or work that cannot honestly satisfy the
   original acceptance criteria.
5. Require every task after the first to record a concrete split reason tied to a
   real merge-order dependency, independently releasable/rollbackable unit, distinct
   owner/authority/risk boundary, mutually exclusive execution environment,
   post-merge evidence that cannot be produced by the implementation carrier, or
   review-size boundary too large for a reliable single PR.
6. Add deterministic validation/tests that reject or fail closed on unexplained task
   splitting and keep justified multi-task packages compatible with existing
   sequential auto-advance behavior.
7. Update change-package templates and mirrored consumer fixtures/tests so they stay
   consistent with the new policy.
8. Reconcile DOC-15's seven-task component example and the development workflow's
   `L`/800-line split directives with DOC-15's anti-artificial-splitting and
   coherent-PR requirements.
9. Define task IDs as minimum-sufficient traceability groupings and make one
   outcome-sized implementation PR the default. Size remains a reviewability signal,
   never a sufficient split reason by itself.
10. Update the primary `KARSIFT/karsift-ai-infra` prompt/validation contract and its
    caller fixture in the same T00; the fixture alone is not runtime truth.

### Non-goals / explicitly excluded

- Removing exact-SHA independent verification.
- Lowering deterministic path-based risk floors.
- Weakening protected-branch rules, merge-gate checks, or fail-closed behavior.
- Allowing unrelated bug fixes to piggyback on an active package.
- Allowing in-scope remediation to exceed the original package's risk ceiling or
  protected-area scope without a new plan.
- Converting every historical multi-task package to one task retroactively.
- Changing product runtime behavior, credentials, deploy topology, or monitor
  inventory.
- Self-adoption or self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R4**.
- Protected areas: `AGENTS.md`, operating-model/governance documents, package
  templates, lifecycle validation, mirrored karsift-ai-infra prompts/workflows, and
  task-sequencing logic/tests.
- Protected technical effect: planning/adoption/task-release policy only; no product
  runtime effect is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance/workflow changes still require exact-SHA
  independent verification and fail-closed controls.

## Decisions

`VOC-115-D00`: The default planner unit is one **coherent objective per package**.
A causally related defect discovered while implementing, verifying, merging,
promoting, or reconciling an active package should remain under that package when,
and only when, it stays within the original objective, acceptance criteria, risk
ceiling, and protected-area scope.

`VOC-115-D01`: A new package is still required for unrelated scope, changed product
intent, an authority/risk boundary not already covered by the active package, or
work that cannot honestly satisfy the original acceptance criteria.

`VOC-115-D02`: The default task unit for every new package is **one end-to-end
implementation task** containing code, tests, documentation, migration/config
updates, and acceptance evidence that can be produced by the same carrier PR or its
governed workflow.

`VOC-115-D03`: Splitting into multiple tasks is allowed only when at least one
concrete boundary exists:

1. hard merge-order dependency;
2. independently releasable or rollbackable unit;
3. distinct owner, authority, or risk boundary;
4. mutually exclusive execution environment;
5. post-merge evidence that cannot be produced in the implementation carrier; or
6. change size too large for reliable single-PR review.

The sixth boundary requires a concrete reviewability explanation tied to cognitive
load, risk isolation, or rollback. An `L` label, file count, component count, skill
count, or changed-line threshold alone does not satisfy it.

`VOC-115-D04`: Every task after the first must record an explicit **split reason**
from the allowlist in `VOC-115-D03`. Reasons such as "small", file type, tests vs
code, docs vs code, or general review convenience are insufficient by themselves.

`VOC-115-D05`: Packages proposing more than three tasks require explicit package-level
justification for why consolidation is unsafe and plan-review scrutiny should treat
that as exceptional rather than normal.

`VOC-115-D06`: Exact-SHA independent verification, deterministic risk floors,
protected-branch rules, and fail-closed validation remain unchanged. The efficiency
change is consolidation of carriers and scope handling, not relaxation of gates.

`VOC-115-D07`: Evidence should stay on the same carrier whenever possible. The
lifecycle must not create a separate issue/PR whose only purpose is to carry
evidence that the same implementation carrier or its governed workflow could have
produced.

`VOC-115-D08`: Automation must remain compatible with justified multi-task packages.
When multiple tasks are still warranted, sequential roster advancement, exact task
dependency order, and operator/live-evidence handling remain intact.

`VOC-115-D09`: Canonical docs, planner/plan-review prompts, templates, validation,
and mirrored consumer fixtures/tests must be updated together in the same package so
no source claims the old fragmentation rule after the new one lands.

`VOC-115-D10`: One coherent outcome defaults to one outcome-sized implementation PR
per repository. Fixed size labels and changed-line thresholds are review signals,
not automatic plan, task, or PR split rules. When a cross-repository contract needs
one PR in each repository, those coordinated carriers may remain one task.

`VOC-115-D11`: Task IDs are minimum-sufficient traceability groupings around
independently meaningful outcomes. Components, files, skills, schema/service/API/UI
layers, tests, docs, and same-carrier evidence remain in one task when they jointly
deliver the same outcome. In particular, adding several related skills plus their
configuration, adapters, documentation, and tests is one plan and one task.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is a governance/workflow change only.

## Security, privacy, and authorization

No new secrets, tokens, or privileged automation are introduced by the package
definition itself. The package must preserve least-privilege workflow behavior and
must not let a broader scope ride under an active package merely for convenience.

The principal abuse risk is policy drift that accidentally permits unrelated work or
expanded protected-area changes to stay inside an active package. Mitigation:
explicit boundary rules, deterministic regression tests, and fail-closed validation.

## Contradictions and open questions

1. **Canonical wording reconciliation:** `DOC-15` still says tasks are "small,
   ordered, verifiable, and stable." T00 must update current-state docs so that
   wording no longer contradicts the one-task default while preserving the ability
   to justify real multi-task packages.
2. **Enforcement layer split:** The implementation may enforce split reasons in more
   than one place (planner prompt, plan review, adoption, deterministic validation,
   or fixture tests). This draft requires the outcome, not one particular internal
   split of enforcement responsibility.
3. **Carrier-remediation boundary wording:** The package should define policy in a
   way that clearly distinguishes in-scope causal remediation from an unrelated bug
   discovered incidentally during the same work, so automation and reviewers remain
   fail-closed on scope creep.
4. **Cross-repository delivery:** the shared-infra source and caller fixture require
   separate repository PRs, but both are one T00. Evidence must bind the caller pin
   to the exact reviewed infra merge so the extra carrier does not become another
   plan/task.
