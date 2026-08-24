# VOC-115 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `AGENTS.md`, `docs/operations/10-development-workflow.md`,
  `docs/operations/15-ai-native-product-and-engineering-operating-model.md`,
  `docs/governance/16-autonomous-development-operating-model.md`,
  `docs/templates/change-specification.md`, `specs/templates/change-package/`,
  `tooling/governance/`,
  `tooling/governance/tests/`, and mirrored
  `tooling/governance/fixtures/karsift-ai-infra/` prompts/workflows/config/tests.
- Prerequisites: confirm current planner prompt, plan-review prompt, adopt roster
  behavior, and any deterministic package-validation rules that currently assume or
  encourage multi-task decomposition.
- Preserve existing compatible justified multi-task behavior; this package changes
  defaults and enforcement, not the existence of multi-task packages.

## File reconciliation and implementation sequence

### T00 — Consolidate package/task defaults, causal-remediation policy, validation, and regression coverage

| Target | Action | Notes |
|--------|--------|-------|
| `specs/changes/VOC-115-.../t00-evidence.md` | create/update | Record final policy wording, changed surfaces, commands, and results |
| `docs/operations/10-development-workflow.md` | modify | Remove automatic `L`/800-line splitting; make size a reviewability signal subordinate to outcome/risk/rollback boundaries |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | modify | Replace fragmentation-encouraging task language with one-task default + explicit split boundaries |
| `docs/governance/16-autonomous-development-operating-model.md` | modify | Reconcile live governed lifecycle wording with one-package/one-task default and bounded in-scope remediation |
| `AGENTS.md` | modify | Update bug-handling / governed-loop language so in-scope causal remediation can stay under the active package; preserve unrelated-bug new-plan rule |
| `docs/templates/change-specification.md` | modify | Make task IDs minimum-sufficient outcome traceability groupings and remove component-oriented split cues |
| `specs/templates/change-package/` | modify | Update template wording so tasks and task-specific evidence no longer imply fragmentation by default |
| `tooling/governance/fixtures/karsift-ai-infra/prompts/plan.md` | modify | Replace "small, ordered" default with one-task default + split-reason rules |
| `tooling/governance/fixtures/karsift-ai-infra/prompts/plan-review.md` | modify | Make multi-task review check require explicit split reasons and bounded scope handling |
| `KARSIFT/karsift-ai-infra` primary repository | modify in the same T00 before pinning the caller fixture | Land the planner/plan-review/validation/test contract through one reviewed infra PR; do not treat the fixture as runtime truth |
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/adopt.yml` | modify if needed | Preserve sequential dependency edges while remaining compatible with one-task default and split-reason enforcement |
| `tooling/governance/fixtures/karsift-ai-infra/config/*` | modify if needed | Add deterministic parsing/validation helpers only where required by the accepted enforcement design |
| `tooling/governance/tests/*` and mirrored fixture tests | modify/extend | Add one-task default, split-reason, justified multi-task, and causal-remediation policy coverage |

Ordered steps:

1. Reconcile the canonical policy language across the development workflow,
   `DOC-15`, `DOC-16`, and `AGENTS.md`
   so the one-package/one-task default and active-package causal-remediation boundary
   are stated consistently.
2. Update change-package template wording to make one end-to-end task the default and
   to require explicit split reasoning after the first task.
3. Update the primary shared-infra planner and plan-review prompts plus deterministic
   validation/tests through one reviewed infra PR, then sync and pin the caller's
   mirrored fixture to that exact infra merge. Both repository PRs remain one T00;
   the repository boundary is not a reason to create another plan or task.
4. Decide and implement the enforcement split across prose, deterministic validation,
   and fixture tests so packages without split reasons fail closed.
5. Add regression coverage for:
   - ordinary coherent request -> one package / one task;
   - code + tests + docs staying together by default;
   - several related skills + configuration + docs + tests staying in one task;
   - changed-line count and `L` sizing remaining review signals rather than
     automatic task/PR split commands;
   - invalid split without reason;
   - valid justified multi-task case with preserved sequential order;
   - in-scope causal remediation vs unrelated or authority-expanding follow-up.
6. Preserve exact-SHA review, risk-floor, protected-branch, and fail-closed
   invariants throughout.
7. Run applicable validation and record results in `t00-evidence.md`.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'
git diff --check
```

If implementation changes mirrored fixture tests that use another targeted command,
record the exact command in `t00-evidence.md` and run it in addition to the suite
above.

Independent verifier (exact reviewed task PR SHA) should confirm:

- one-task default is stated consistently across docs, template, planner prompt, and
  plan-review prompt;
- every extra task requires an explicit concrete split reason;
- justified multi-task sequencing still works;
- active-package causal remediation remains bounded by objective, acceptance criteria,
  risk ceiling, and protected-area scope;
- exact-SHA verification, risk floors, protected-branch controls, and fail-closed
  behavior are unchanged.
- the caller fixture pin equals the exact reviewed shared-infra merge that provides
  the runtime planner and plan-review contract.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime.
- **Operational effect:** Future planning/adoption/implementation cycles should use
  fewer carriers for coherent work while keeping the same gates.
- **Rollback trigger:** Scope boundaries become ambiguous; justified multi-task
  packages stop validating; auto-advance sequencing regresses; or docs/prompts drift.
- **Rollback mechanism:** Revert the governance/template/fixture/test changes in the
  implementation PR(s), restoring the prior behavior.
- **Last-known-good reference:** Current `main`/`develop` governance lifecycle before
  VOC-115 implementation lands.
