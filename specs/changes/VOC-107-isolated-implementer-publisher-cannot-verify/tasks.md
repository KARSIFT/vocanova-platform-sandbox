# VOC-107 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

Cross-repo note: T00 primarily changes `KARSIFT/karsift-ai-infra`. The
implementer opens PRs there for that behavior; this package is the authorizing
change package for the required outcome. Do not treat the untracked local
`karsift-ai-infra/` checkout (if present) as this repo's tracked tree.
Calling-repo foundation-test or doc changes land in this repository under the
same package.

## VOC-107-T00 — Integration-anchored implementer bundle lineage, publisher guards, docs, deterministic Git fixture

- Requirement source: issue #891; `VOC-107-D00`–`D07`
- Acceptance criteria: `VOC-107-AC-00` through `VOC-107-AC-06`
- Tests: `VOC-107-TEST-00` through `VOC-107-TEST-08`
- Evidence: `VOC-107-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml` so recovery/
   publish bundle creation uses an explicit integration-anchor SHA through
   `HEAD` (complete task-branch-only lineage), not only
   `pre_model_tip..HEAD`, per `VOC-107-D00`.
2. Record the integration-anchor SHA when the implementation branch is prepared
   (fresh checkout from integration, or after successful rebase / fresh-after-
   conflict restart onto integration). Keep today’s pre-model tip as the
   soft-reset commit boundary (`VOC-107-D01`).
3. Wire the isolated publish job so `PUBLISH_BASE_SHA` / ancestry / workflow-path
   deny scanning use that same integration anchor while preserving exact
   published SHA check and SHA-valued force-with-lease (`VOC-107-D02`).
4. Do not change the two-attempt cap. Do not add a model rerun path whose only
   justification is a missing publisher prerequisite on an otherwise valid
   committed bundle (`VOC-107-D03`).
5. Cover rebase-derived remediation lineage so locally recreated bases cannot
   strand publication (`VOC-107-D04`).
6. Add a deterministic end-to-end Git fixture (infra self-ci and/or
   `scripts/foundation/voc107-*.test.mjs`) per `VOC-107-D05`: integration,
   attempt-1, attempt-2; clean bare repo with only integration verifies/imports
   the attempt-2 artifact; negative malformed/stale incomplete lineage fails
   closed; attempt-1 regression retained.
7. Update karsift-ai-infra README so operators understand integration-anchored
   bundles. Update calling-repo docs only if current wording would become false.
8. Record that the caller already consumes reusable workflows at `@main`; no pin
   bump is expected. If implementation discovers a different current reference,
   reconcile it explicitly and record the actual consumption mechanism.
9. Run applicable tests and governance validation for changed calling-repo paths;
   record commands and results in `t00-evidence.md` using allowlisted metadata
   only (`VOC-107-D06`).
10. Do not address Node runtime deprecations, Go cache warnings, dependency
    updates, OAuth/application behavior, deployment changes, or `plan.yml`
    planner-bundle behavior (out of scope per `VOC-107-D07`). If an identical
    planner failure class is discovered, open a separate unlabeled issue rather
    than expanding this task.

### Explicitly out of scope for this task

- Operator-owned live Actions proof beyond the deterministic Git fixture.
- Granting implementer Actions credentials or App tokens on the model runner.
- Application or monitoring-inventory ID changes.
- Changing the two-attempt remediation/implementer policy.

## Task ordering notes

- Single implementable task. Deterministic Git fixtures are the acceptance proof
  for this root; no operator-owned live-evidence contract is required.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
