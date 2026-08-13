# VOC-076 — Implementation Plan

## Preconditions and protected areas

Do not begin any task until this package is adopted and implementation is
authorized. `VOC-076-T01` must not start until `VOC-076-T00` names a specific
evidence-backed cause (or adoption explicitly scopes the fix path despite
residual ambiguity). `VOC-076-T02` must not start until `VOC-076-T01` merges.

Protected / sensitive areas:

- `apps/web/src/app/(app)/reviews/_components/review-session.tsx` — conditional
  edit if T00 finds a product disabled-state hang; preserve CSRF, session
  error handling, accessibility, and `max-w-*` workaround.
- `apps/web/tests/staging-e2e/core-loop.staging.spec.ts` — conditional edit for
  enabled/readiness waits; T02 verification target; preserve VOC-074
  `reviewedCards >= 1` hardening.
- `.github/workflows/deploy-staging.yml` — **out of default scope**; only if
  adoption expands `VOC-076-DEP-01` (raises path floor to R3).
- `apps/api/` — read-only for latency investigation unless T00 proves an
  in-scope API hang that must be fixed in this package.
- `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button` —
  this package's evidence files.

Explicitly out of scope by default: governance docs, VOC-074 package fields,
production secrets, broad review UX redesign.

## File reconciliation and implementation sequence

1. **`VOC-076-T00`** — Investigation only. Produce `t00-evidence.md`. Prefer
   reproducing the disabled MC button against staging timing and tracing
   `isSubmitting` / `isRefetching` / `phase` vs the E2E visibility-only wait
   (see `tasks.md`).
2. **`VOC-076-T01`** — Narrow fix + regression/readiness coverage per T00.
   Produce `t01-evidence.md`. Run applicable `apps/web` validation per
   `docs/development.md`.
3. **`VOC-076-T02`** — No file change expected. After T01 merges to `develop`,
   record a real staging pass in `t02-evidence.md` (with MC coverage per
   `VOC-076-AC-03`).

Prefer separate PRs per task for reviewability.

## Validation and independent verification

Deterministic commands before claiming any task complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

For `VOC-076-T01`, additionally run applicable `apps/web` lint / typecheck /
unit or component tests for the touched files, and any staging-e2e local
checks `docs/development.md` documents as available. Do not invent
unavailable staging credentials as a pass — T02 is the real staging proof.

Do not invent unavailable checks as passing.

Independent verification (per `CLAUDE.md`) must confirm against the exact
implemented revision:

- T00 evidence supports T01 scope (product vs E2E vs both).
- Fix is narrow; no unrelated refactor; CSRF/session/a11y/`max-w-*` preserved.
- E2E does not merely raise timeouts without a readiness signal when race was
  confirmed; does not skip VOC-074 hardening.
- T02 staging evidence includes step 5 completion and MC coverage per AC-03 —
  not only a green Playwright line from a self-check-only run if MC was the
  failure mode.
- Path risk floor re-measured; semantic R2 proposal still justified (or raised
  / lowered with evidence).
- No unauthorized `deploy-staging.yml` expansion (`VOC-076-DEP-01`).
- Implementer did not approve or merge their own work.

## Deployment and rollback

No separate deployment authorization beyond the existing governed
`develop` → (automatic) `main` / production path described in AGENTS.md. Closing
this package's task roster participates in the existing auto-release mechanism
once adopted and implemented.

Rollback trigger: after T01, review submissions fail, MC options never disable
when they should (double-submit), E2E becomes flaky for wrong reasons, or
independent review finds the fix incorrect.

Rollback mechanism: revert T01 commit(s).

Last-known-good reference: the revision immediately preceding the affected
task's merge.

Owner: implementer of the affected task.
