# VOC-053 — Implementation Plan

## Preconditions and protected areas

Do not begin `VOC-053-T01`'s fix until `VOC-053-T00` has recorded a
confirmed root cause with direct evidence (per `VOC-053-DEP-00`) — this
package's whole premise is that guessing the cause and implementing a fix
anyway is explicitly out of scope (see `specification.md`'s open questions
and issue #450's own "not prescriptive" framing). Protected areas touched
depend on which candidate is confirmed:

- Candidate (a) caching: `apps/web/src/lib/api-server.ts` and/or
  `apps/web/src/app/(app)/home/page.tsx`, under `apps/web/src/`.
- Candidate (b) backend bug: `apps/api/business/missions/`,
  `apps/api/business/gamification/`, and/or `apps/api/app/api/missions.go`
  (or wherever the live request handler is actually found during T00's
  trace), and — conditionally, only if `VOC-053-R02` resolves to "a migration
  is needed" — a new file under `apps/api/migrations/`, the most sensitive
  protected path either candidate could touch.
- Candidate (c) test-data interaction: most likely
  `tests/staging-e2e/core-loop.staging.spec.ts`'s helper functions or the
  synthetic-account seeding/reset scripts it depends on
  (`apps/api/scripts/seed-synthetic-smoke-user.sh` or equivalent) — not the
  spec's own step 7 assertion.

## File reconciliation and implementation sequence

1. **`VOC-053-T00`** — No file change. Investigate per `tasks.md`'s T00
   description: inspect real staging HTTP response headers/cache-status for
   `/home` and its underlying data fetch; trace the backend request-handling
   path for the daily-mission read from the HTTP handler down through
   `gamification.LocalDate` and `missions.Repository`; if inconclusive, test
   the fresh-account reproduction issue #450 itself suggests before
   concluding candidate (c) alone explains it. Record the confirmed cause and
   its evidence before proceeding.
2. **`VOC-053-T01`** — Implement the fix scoped to whichever candidate T00
   confirmed (see `tasks.md`'s T01 description for the per-candidate fix
   shape). Preserve every other existing behavior on the affected file(s)
   unchanged — this is a targeted bug fix, not a refactor. Do not touch
   `tests/staging-e2e/core-loop.staging.spec.ts`'s step 7 assertion.
3. **`VOC-053-T02`** — No further file change expected (beyond any narrow gap
   this verification pass itself surfaces). Trigger a real staging deploy
   with T01's fix in place; specifically arrange or wait for a run where the
   synthetic account already has `reviewedBefore >= 1` residue from a prior
   run, so the verification actually exercises the originally-observed
   failure condition rather than only a fresh-state pass.

## Validation and independent verification

Deterministic commands to run before claiming any task complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

For `apps/web`/`apps/api` code changes, additionally run the workspace
validation documented in `docs/development.md`
(`pnpm lint` / `pnpm typecheck` / `pnpm test` / `pnpm test:api` as applicable
to the actually-changed package). Live staging HTTP header inspection and the
real E2E rerun cannot be exercised from a local or CI-only environment
without the same class of staging access prior packages (e.g. VOC-050,
VOC-052) have required — a syntactically valid, logically reviewed diff plus
a real triggered run against the actual staging host (once merged to
`develop`) is the realistic verification path for `VOC-053-T02`. Independent
verification (per `CLAUDE.md`) must re-review the exact implemented
revision's diff against this specification and acceptance criteria, confirm
`VOC-053-T00`'s evidence genuinely supports the claimed root cause (not
merely asserts it), confirm the fix is scoped to that cause and does not
weaken the step 7 assertion, and confirm no unrelated change was introduced.

## Deployment and rollback

No separate deployment authorization is required beyond the existing
`develop`-merge and (if the `karsift-ai-infra` auto-promotion path applies)
`main`-promotion mechanisms already governing this repository. Rollback
trigger: if the fix causes a new staging deploy failure, or if the fixed
behavior is itself observed to be wrong on a subsequent real run, revert the
task's diff; because no candidate fix under consideration involves a
destructive data change (any migration, if needed per `VOC-053-R02`, would be
additive/corrective, not destructive — this must be confirmed explicitly if
that path is taken), reverting the code change is expected to be sufficient
without a separate data rollback. Owner: whoever implements `VOC-053-T01` is
accountable for confirming the revert path works if invoked. Last-known-good
reference: the affected file's revision immediately preceding this package's
merge.
