# VOC-050 — Implementation Plan

## Preconditions and protected areas

Do not begin any task until this package is adopted (`change.yaml`'s
`status: adopted` and `implementation_authorized: true`), and until each task
is individually authorized/dispatched, per this repository's own convention
for prior packages (VOC-039 through VOC-049).

Protected areas touched:
- `apps/api/migrations` - a new migration for the synthetic account; must
  follow `.karsift/lessons.md`'s Atlas-directive and duplicate-unique-index
  lessons.
- `apps/api/business/auth` - the session-minting mechanism, an
  authentication protected area.
- `.github/workflows/deploy-staging.yml` and
  `.github/workflows/deploy-production.yml` - workflow-path-class protected
  area.
- `infra/scripts/smoke-test-production.sh` - must preserve its existing
  no-side-effects design.

No other file may be touched by any task without a package amendment. In
particular, no task may modify anything in the `karsift-ai-infra` repository
- that repository is entirely out of this package's reach (see
`specification.md`'s open question 1).

## File reconciliation and implementation sequence

Existing state, confirmed by reading the files during drafting:
- `apps/api/cmd/seed/main.go` is a precedent for an idempotent, rerun-safe
  seed command loading fixed-primary-key content; it seeds product content,
  not a user account, but its "safe to rerun" pattern is the one `T00`
  should follow for the synthetic account.
- `apps/api/app/api/production.go`'s `RegisterMonitoringSentryTest` is a
  precedent for a secret-gated, fail-closed-when-unset mechanism
  (`if strings.TrimSpace(expectedToken) == "" { return }`); `T01` should
  mirror this pattern rather than inventing a new one.
- `infra/scripts/smoke-test-production.sh`'s core-loop section (`# 5.`) and
  `deploy-production.yml`'s "Run production smoke-test suite" step already
  exist and already anticipate `SMOKE_TEST_SESSION_COOKIE`; `T04` only needs
  to make that variable non-empty and real, not build new script logic
  (though `T04`'s implementer must re-verify the script's exact expectations
  against the synthetic account's actual seeded data - e.g. that
  `GET /api/v1/journey-situations` returns 200 for it - before assuming no
  script change is needed).
- `deploy-staging.yml` has no equivalent core-loop section at all today;
  `T02` and `T03` add one, mirroring the existing `/healthz`/`/`-poll steps'
  fail-closed style (no `continue-on-error`, explicit error messages on
  timeout).

Ordered, reversible steps:

1. `VOC-050-T00` (seed the synthetic account): add a migration creating the
   synthetic test user (plus any supporting flag/column), and an idempotent
   seed mechanism (new `apps/api/cmd/seed-*` command, or an extension of the
   existing seed command - implementer's choice, documented in the PR) that
   ensures the account exists on every staging/production deploy. Resolve
   `change.yaml`'s `VOC-050-DEP-02` (the reachability-blocking mechanism)
   concretely in this task. Reversible by reverting the migration/seed
   commit; the seeded row itself can be deleted manually if ever needed.
2. `VOC-050-T01` (mint a session server-side): add a narrowly-scoped,
   secret-gated mechanism in `apps/api/business/auth` (or a small CLI/admin
   endpoint calling into it) that produces a valid session for the synthetic
   account only, without a real magic-link/OAuth round-trip, mirroring
   `RegisterMonitoringSentryTest`'s fail-closed-when-unset pattern. Depends
   on `T00`'s seeded account existing. Reversible by reverting the commit;
   removes the mechanism entirely with no data-compatibility concern.
3. `VOC-050-T02` (staging core-loop step): add a staging-targeted
   core-loop check (either a modified `core-loop.spec.ts` or a new sibling
   spec file - implementer's choice, per `specification.md`'s open question
   3) that uses `T01`'s minted session to run the real journey against
   `https://staging.vocanova.site`, and wire it into `deploy-staging.yml`
   immediately after the existing `/healthz` poll. Depends on `T00`+`T01`.
   Reversible by reverting the workflow/spec commit; does not affect
   already-deployed staging containers.
4. `VOC-050-T03` (fail-closed signal + cross-repo documentation): confirm
   `deploy-staging.yml`'s new step has no `continue-on-error` and that any
   journey-step failure genuinely fails the job (verified live against a
   deliberately-broken staging deploy or a disposable test branch, not just
   read from the YAML). Document, in this task's own pull request, the
   confirmed-open cross-repo dependency on `karsift-ai-infra`'s `release.yml`
   actually gating on this job's conclusion (see `specification.md`'s open
   question 1) as an explicit item for the reviewing human, since this task
   cannot resolve it itself. Depends on `T02`.
5. `VOC-050-T04` (production core-loop activation): update
   `deploy-production.yml`'s "Run production smoke-test suite" step to set
   `SMOKE_TEST_SESSION_COOKIE` from `T01`'s minting mechanism (invoked in a
   prior step, following the same SSH-command pattern the file already uses
   for other pre-deploy configuration), and re-verify
   `smoke-test-production.sh`'s core-loop section (`# 5.`) still matches the
   synthetic account's actual seeded data. Depends on `T00`+`T01`. Reversible
   by reverting the workflow commit, returning to the existing non-fatal
   SKIP behavior.

Each step is independently reviewable and depends only on the steps listed
above it; `T02`-`T04` may be sequenced or combined differently by the
reviewing human at adoption time if that proves cleaner, but the dependency
order (`T00` before `T01` before everything else) is load-bearing.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section, for
changes touching `apps/api`/`apps/web`):

```bash
pnpm validate   # or the narrower pnpm lint / typecheck / test / build, scoped as needed
```

Additionally, per task:
- `T00`: `go test ./apps/api/cmd/seed/...` (or the new seed package's tests)
  and a real migration apply against a disposable Postgres instance,
  confirming idempotent rerun.
- `T01`: `go test ./apps/api/business/auth/...` plus a new test confirming
  the mechanism is unreachable/no-op when its gating token is unset.
- `T02`: `pnpm --filter web exec playwright test` (or this repository's
  documented equivalent) for the new/modified spec, run against a real or
  disposable staging-shaped target before merging, not only asserted to
  work.
- `T03`: a live-observed failing run (deliberately broken target or
  disposable test branch) confirming `deploy-staging.yml`'s job conclusion
  is failure, not merely inspecting the YAML.
- `T04`: a live-observed run of `smoke-test-production.sh` against a real or
  disposable target with a real `SMOKE_TEST_SESSION_COOKIE` set, confirming
  `PASS:` lines replace the previous `SKIP:` line.

Independent verification: Claude Code reviews the exact final revision of
each task's pull request against this package's specification, acceptance
criteria, and the applicable risk floor (which may exceed the package-level
R2 for `T00`/`T01`/any workflow-file task per
`docs/governance/change-risk-classification.md`), per `CLAUDE.md`. No task
may be self-approved by its own implementer.

## Deployment and rollback

- No task in this package is deployment-authorized. `release-plan.md`
  records that deployment/release remains a separate, later decision.
- Rollback mechanism: revert the merged commit/PR for the affected task. The
  seeded synthetic account (`T00`) is the only stateful artifact; rollback
  of `T00` should also delete the seeded row if the migration itself is
  reverted, to avoid leaving an orphaned account with no corresponding code
  path expecting it.
- Rollback trigger: any evidence the synthetic account is reachable via a
  real signup path (regression of `VOC-050-AC-01`), any evidence the
  session-minting mechanism is reachable without its gating token
  (regression of `VOC-050-AC-02`), or any evidence the production core-loop
  check performs a state-mutating request against real infrastructure
  (regression of `VOC-050-AC-05`).
- Rollback owner: whoever holds deployment authority at the time, per this
  repository's active authority model (A-003); this package does not itself
  grant that authority to anyone.
- Last-known-good reference: the `develop` branch commit immediately prior
  to the affected task's merge.
