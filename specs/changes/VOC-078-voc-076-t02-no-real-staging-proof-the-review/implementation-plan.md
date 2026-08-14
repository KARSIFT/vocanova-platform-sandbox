# VOC-078 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is
authorized. Adoption should record `VOC-078-DEP-00` and `VOC-078-DEP-01`
(and note any known post-#598 staging run for `VOC-078-DEP-02`) so
implementers are not guessing about VOC-076-T02 disposition or process
follow-up scope.

In-scope paths:

- This package directory (evidence files created by implementers)
- `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button/t02-evidence.md`
- `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button/acceptance-criteria.md`
  (`VOC-076-AC-03` Result only)
- Conditional (T01 FAIL path only):
  - `apps/web/src/app/(app)/reviews/_components/review-session.tsx`
  - `apps/web/src/app/(app)/reviews/_components/review-session-prompt.ts`
  - `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`
  - `apps/web/tests/lib/review-session-prompt-readiness.test.ts`

Explicitly out of default scope: `.github/workflows/deploy-staging.yml`
(R3), `apps/api/`, migrations, production secrets, merge-gate /
`karsift-ai-infra` FAIL-merge hardening (unless adoption expands
`VOC-078-DEP-01` into a separate package).

Measured drafting-time path floor for evidence-only and for
evidence+review/E2E remediation: **R1**. Proposed package risk: **R2**
(draft proposal — see `change.yaml`).

## File reconciliation and implementation sequence

1. Confirm adoption decisions for DEP-00 / DEP-01 / risk class.
2. **T00:** Confirm tip contains VOC-076 T01 + #598 gap fix; locate or wait
   for a real `deploy-staging` run; write `t00-evidence.md` with honest
   PASS or FAIL.
3. On PASS: update VOC-076 `t02-evidence.md` and AC-03 Result; close #575;
   mark T01 N/A.
4. On FAIL: leave #575 open and AC-03 unmet; proceed to T01.
5. **T01 (FAIL only):** narrow fix + local validation + new green staging
   run; update VOC-076 evidence/AC-03; close #575; write `t01-evidence.md`.
6. Open/land task PR(s); wait for independent review PASS; merge via normal
   merge-gate (no FAIL override — the #598 failure mode must not recur).

## Validation and independent verification

Deterministic commands before claiming complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

When T01 touches `apps/web`, also run the applicable commands from
`docs/development.md` (for example `pnpm --filter @vocanova/web` test /
typecheck / lint targets for touched packages). Do not invent or report an
unavailable check as passing.

Independent verification (per `CLAUDE.md`) must confirm against the exact
implemented revision:

- Green staging run URL present iff AC-00/AC-01 claim PASS.
- VOC-076 evidence and AC-03 Result aligned; #575 disposition correct.
- No unauthorized workflow expansion; VOC-074/VOC-050 boundaries intact.
- Declared risk meets or exceeds path floor.
- Implementer-role occupant did not approve or merge its own
  implementation.
- Active authority model remains `a003-active`.

## Deployment and rollback

No separate deployment authorization is granted by this package. Staging
verification uses the existing `deploy-staging.yml` path after merges to
`develop`. If T01 changes product code, production impact follows the
repository's automatic develop→main→deploy path once the package's task
roster closes (per AGENTS.md "Release and deployment authority") — that is
not additional authority granted here beyond the normal governed loop.

Rollback trigger: false PASS evidence; secret leakage; unauthorized
workflow edits; or T01 remediation that regresses review-queue safety.

Rollback mechanism: revert the implementation commit(s). Last-known-good:
repository state immediately preceding this package's implementation merge
(VOC-076 AC-03 still unmet / #575 open is the known pre-fix evidence
state).

Owner: implementer of the active task (`VOC-078-T00` / `VOC-078-T01`).
