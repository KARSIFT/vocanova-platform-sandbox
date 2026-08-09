# VOC-052 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status: adopted`,
`approval_status: approved`, `implementation_authorized: true` in `change.yaml`),
and until `VOC-052-DEP-01` (build/run mechanism) is resolved for `VOC-052-T00`,
and — only if `VOC-052-T02` is dispatched — `VOC-052-DEP-02` is resolved.
Protected areas touched: `.github/workflows/deploy-staging.yml` and,
conditionally, `.github/workflows/deploy-production.yml`, both R3 path classes
per `docs/governance/change-risk-classification.md`. Both files carry extensive
existing header-comment documentation (deploy sequencing, staging host layout,
required secrets) that must be kept accurate — any structural change to what the
deploy bundle contains or what steps run in what order must update the relevant
header comment in the same diff, per this repository's own documented practice
in these files.

## File reconciliation and implementation sequence

1. **`VOC-052-T00`** — In `.github/workflows/deploy-staging.yml`:
   - Extend the "Bundle deployable artifacts" step (or add a sibling step) to
     produce the `apps/api/cmd/seed` artifact per the chosen mechanism (see
     `specification.md` open question 1). If mechanism (a) is chosen: add a
     `go build` invocation (using the repository's pinned Go toolchain, matching
     `apps/api/Dockerfile`'s `CGO_ENABLED=0 GOOS=linux` build flags for
     consistency) producing a static binary, and add it to the existing
     `/tmp/deploy-bundle/apps/api/scripts/` (or a new `bin/` subdirectory) the
     same way `migrate.sh` and `seed-synthetic-smoke-user.sh` are already
     copied in.
   - Extend the SSH step's command sequence to run the seed binary immediately
     after the existing `migrate.sh` invocation and its neighboring
     `seed-synthetic-smoke-user.sh` call, with `DATABASE_URL` resolved via the
     same private Postgres bridge IP substitution the migration step already
     performs (do not duplicate that resolution logic if it can be reused as a
     shell variable already in scope in that step).
   - Update the workflow's own header comment (the "Staging host layout" list
     and the numbered sequence describing steps 1–6) to include the new
     artifact and step, so the documentation stays accurate to the actual
     sequence — per this repository's own convention that any change to
     deploy sequencing must update the header comment describing it.
   - Ensure the new step has no `continue-on-error` and runs under the same
     `set -euo pipefail` discipline the rest of the SSH step block uses, so a
     seed failure aborts before `docker compose up -d`.
2. **`VOC-052-T01`** — No file change expected. Trigger and observe a real
   staging deploy with `VOC-052-T00`'s step in place; capture the workflow run
   log and the `tests/staging-e2e/core-loop.staging.spec.ts` result as evidence.
   If the spec still fails for a reason distinct from the original empty-content
   issue, diagnose and fix only that specific gap within this task's scope.
3. **`VOC-052-T02`** (conditional) — Mirror step 1's change in
   `.github/workflows/deploy-production.yml`'s equivalent section, only if
   dispatched per `VOC-052-DEP-02`'s resolution.

## Validation and independent verification

Deterministic commands to run before claiming any task complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Workflow-file changes cannot be fully exercised without live SSH credentials to
the staging host (the same `VOC-032-DEP-00`-class limitation every prior
`deploy-staging.yml` change has carried); a syntactically valid, logically
reviewed diff plus a real triggered run against the actual staging host (once
merged to `develop`, which is when the workflow actually executes) is the
realistic verification path — matching how `VOC-050`'s own T02/T03 were
verified. Independent verification (per `CLAUDE.md`) must re-review the exact
implemented revision's diff against this specification and acceptance criteria,
confirm the seed step is genuinely fail-closed and idempotent as claimed (not
merely asserted), and confirm the header-comment updates match the actual new
sequence.

## Deployment and rollback

No separate deployment authorization is required beyond the existing
`develop`-merge and (if the `karsift-ai-infra` auto-promotion path applies)
`main`-promotion mechanisms already governing this repository — this package
does not request or need any new deployment authority. Rollback trigger: if the
new seed step causes a staging deploy failure unrelated to a genuine schema/data
problem (e.g. a build environment issue), revert the workflow diff; because the
seed is additive and idempotent, no data rollback is needed, and the
previously-passing `migrate.sh` + `seed-synthetic-smoke-user.sh` sequence is
unaffected by the revert. Owner: whoever implements `VOC-052-T00`/`T02` is
accountable for confirming the revert path works if invoked. Last-known-good
reference: the `deploy-staging.yml` revision immediately preceding this
package's merge.
