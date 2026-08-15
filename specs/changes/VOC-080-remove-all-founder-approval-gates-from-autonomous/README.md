# VOC-080 — Remove All Founder Approval Gates from Autonomous Engineering Workflows

**Status: adopted and implementation-authorized.** This package was approved on
[PR #628](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/628) in response to
[issue #627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627),
with independent verification bound to the approved plan revision. Risk is **R4**.

## Identity and lifecycle

- Package ID: VOC-080
- Title: Remove All Founder Approval Gates from Autonomous Engineering Workflows
- Canonical path:
  `specs/changes/VOC-080-remove-all-founder-approval-gates-from-autonomous`
- Lifecycle state: `adopted` (implementation authorized)
- Risk: `R4` (see `change.yaml`'s
  `planned_implementation_risk_floor`; path floors include R4 amendment /
  DOC-15 / governance tooling paths)
- Owner: autonomous implementer with independent verification
- Approval evidence: PR #628 — `approval_status: approved`,
  `implementation_authorized: true`
- Target branch: `develop`
- Linked GitHub issues:
  - [#627](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/627)
    (this package's requirement source)
  - [#573](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/573) /
    [VOC-075](../VOC-075-governance-still-violates-approve-only-r4-28-non)
    (approve-only-R4 predecessor; superseded for workflow gates)
  - [#624](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/624) /
    [VOC-079](../VOC-079-retire-production-nginx-8443-bridge-and-complete)
    (must resume post-activation without founder gates)
  - [#301](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/301) /
    [VOC-040](../VOC-040-adopt-yml-has-no-recovery-path-when-a-plan-pr)
    (merged-but-unadopted recovery gap this package closes structurally)

## Why this exists

VOC-075 / issue #573 made written policy match “approve only R4.” Issue #627
goes further: the founder directs that the governed engineering system must
**no longer require founder approval under any risk class or workflow
condition**. Agents and workflows must progress when deterministic checks,
independent verification, scope, and other non-founder gates pass.

Live pain this package addresses (from #627):

- R4 plan and implementation PRs still wait on a founder `approved` comment.
- Plan packages stay draft/unauthorized and depend on a human adoption flip.
- `merge-gate.yml` is wired with `founder_username`; unparseable risk and
  residual `automatic_merge_allowed: false` still demand founder attention.
- Bot-mediated merge of plan PR #625 did not emit adoption, leaving VOC-079
  merged as draft and forcing manual recovery (#626) under the same approval
  architecture.

## What this package does

1. **Successor authority** (`VOC-080-T00`): author the post-transition
   amendment (recommended A-004) and transition-state scaffolding without
   self-activating.
2. **Merge-gate autonomy** (`VOC-080-T01`): in `karsift-ai-infra`, remove
   founder-comment merge gates for R0–R4; fail unparseable risk for
   correction; never let an approval comment override failed/missing
   CI/independent verification.
3. **Autonomous adoption + reconcile** (`VOC-080-T02`): independently
   reviewed plan packages that pass validation become adopted /
   implementation-authorized automatically with audit evidence; add
   idempotent `workflow_dispatch` reconciliation for merged-but-unadopted
   packages and missing task rosters.
4. **Release / remediate / deploy path** (`VOC-080-T03`): preserve
   auto-promote and push-to-main deploy; remove residual founder-comment or
   environment-reviewer requirements on repository-controlled paths;
   keep fail-closed remediation.
5. **Canonical doc + caller wiring** (`VOC-080-T04`): reconcile AGENTS.md,
   CLAUDE.md, DOC-15/16, matrices, templates, pipeline comments/inputs, and
   repository-settings docs with implemented behavior.
6. **Regression harness** (`VOC-080-T05`): tests covering R0–R4, plan/task
   PRs, remediation, recovery, release, and deployment behaviors.
7. **Rehearsal** (`VOC-080-T06`): prove the new loops in sandbox/dry-run
   before activation.
8. **Activation** (`VOC-080-T07`): exact-revision independent verification and
   recorded activation evidence enable the post-transition
   model so VOC-079 and later work need no founder `approved` comment.

## What this package deliberately does NOT do

- Not removing independent verification, CI, risk classification,
  protected-path floors, rollback, secrets isolation, or least privilege.
- Not allowing an agent to self-review its own exact revision.
- Not replacing founder gates with another standing human approval role.
- Not rewriting historical VOC-075 / A-003 / audit records as if prior rules
  never existed.
- Not implementing VOC-079 bridge retirement (only unblocks its post-
  activation progression).
- Not claiming the post-transition model authorized its own adoption; adoption
  evidence remains the pre-transition decision recorded on PR #628.

## Adoption decisions

The drafting questions were resolved at adoption as follows:

1. A-004 successor vs in-place A-003 edit (`VOC-080-DEP-00`).
2. Product/legal R4 *decisions* vs engineering-workflow gates
   (`VOC-080-DEP-01`).
3. `automatic_merge_allowed` neutralize vs retire (`VOC-080-DEP-02`).
4. Cross-repo sequencing with `karsift-ai-infra` (`VOC-080-DEP-03`).
5. Rehearsal venue and proof bar (`VOC-080-DEP-04`).
6. Risk is **R4**. Activation requires exact-revision independent verification,
   deterministic checks, rehearsal evidence, and rollback readiness—not another
   founder approval comment.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.

This package was adopted under the pre-transition authority recorded on PR #628.
No later founder approval gate is required. Independent verification of each
exact revision remains mandatory throughout.
