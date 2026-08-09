# VOC-052 — Seed the Canonical P1 Content on Staging (and, Conditionally, Production): Specification

## Objective and requirement source

Close the gap reported in [GitHub issue #437](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/437):
staging's real database has never had the canonical P1 content
(`journey_situations`, `canonical_words`, `word_meanings`, `word_examples`,
`usage_notes`, `journey_words`) seeded into it, so the real `/discover` page
renders zero journey-situation links, and VOC-050's first real run of
`tests/staging-e2e/core-loop.staging.spec.ts` (in `deploy-staging.yml`, after every
merge to `develop`) failed at its "discover a situation and open a word" step:

```
Error: expect(received).toBeGreaterThan(expected)
Expected: > 0
Received:   0
  > 239 |       expect(await situationLinks.count()).toBeGreaterThan(0);
```

The objective is: after this package's implementation, the real staging core-loop
E2E check passes at the discover step because the real database actually holds the
canonical P1 content, via an idempotent, deploy-time seed step — not by weakening
the test, mocking the check, or otherwise routing around the real gap.

## Confirmed findings

Independently re-confirmed during drafting, not merely assumed from the issue text:

- `apps/api/migrations/20260725100000_voc026_p1_content_tables.sql` creates the
  content schema (`journey_situations`, `canonical_words`, `word_meanings`,
  `word_examples`, `usage_notes`, `journey_words`) and contains zero `INSERT`
  statements — confirmed by reading the file directly.
- `apps/api/cmd/seed/main.go` exists, embeds `apps/api/cmd/seed/voc026-p1.json` via
  `//go:embed`, reads `DATABASE_URL` from the environment, and applies every table's
  rows inside a single transaction using `INSERT ... ON CONFLICT (id) DO UPDATE SET
  ...` for each table — genuinely idempotent by fixed primary key, and containing no
  `DELETE` statement anywhere in `applySeed`.
- `.github/workflows/deploy-staging.yml` bundles and runs
  `apps/api/scripts/migrate.sh` (T06's Atlas wrapper) and
  `apps/api/scripts/seed-synthetic-smoke-user.sh` (VOC-050-T00's synthetic
  smoke-test account seed, run via `docker compose exec postgres psql`) after
  migrations apply, but has no step anywhere that builds, copies, or runs
  `apps/api/cmd/seed` — confirmed by searching the full workflow file for `cmd/seed`
  and any reference to the P1 content seed, and finding none.
- `.github/workflows/deploy-production.yml` has the same gap: it bundles and runs
  `seed-synthetic-smoke-user.sh` (mirroring staging) but not `apps/api/cmd/seed`,
  and its own real-backend core-loop smoke check already has session-minting
  wiring from VOC-050-T04, so production would hit the identical empty-`/discover`
  failure the moment that check is activated for real (its
  `PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN`-gated path already exists in the
  workflow file).
- Every prior PR-time E2E/accessibility check (e.g. `apps/web/tests/e2e/core-loop.spec.ts`)
  runs against `mock-api-server.mjs`, which serves fixed fixture content regardless
  of the real database — so this gap could not have been caught by any check that
  existed before VOC-050-T02 added the first real-backend staging check. This is
  not a regression VOC-050 introduced.

## Scope and non-goals

In scope:

- Add a step to `deploy-staging.yml` that builds and runs `apps/api/cmd/seed`
  against the real staging database, after migrations apply and before (or as part
  of) bringing the stack up, mirroring `seed-synthetic-smoke-user.sh`'s placement
  and idempotency posture (T00).
- Re-run the real staging core-loop E2E check as verification evidence that the
  discover step now passes against real content (T01).
- Propose (not silently apply) the same addition to `deploy-production.yml`,
  explicitly conditional on a human decision this package does not have the
  authority to make on its own (T02; see `VOC-052-DEP-02`).

Non-goals / explicitly excluded:

- No change to `apps/api/cmd/seed`'s own logic, its embedded `voc026-p1.json`
  content, or the migration files that define the schema. The tool already exists,
  already works, and is already documented as safe to rerun — the gap is purely
  that nothing invokes it during deploy.
- No change to the seed content itself (which situations/words/meanings exist) —
  that is a product-content decision outside this package's scope.
- No weakening, mocking, or skipping of
  `tests/staging-e2e/core-loop.staging.spec.ts`'s existing assertion. The fix is to
  make the real assertion pass for a real reason, not to make it pass trivially.
- No change to VOC-050's release-gating mechanism (whether a failed staging
  core-loop run blocks auto-promotion) — that mechanism already exists per
  VOC-050 and is out of scope here; this package only makes the underlying content
  gap it depends on actually closeable.
- This package does not draft a snapshot-then-recheck-drift task pattern for
  landing anything on `develop` — there is no integration-branch-to-production
  promotion concern here; every task here is new deploy-pipeline content, not a
  sync of already-merged work to another branch.

## Risk and protected areas

Both `deploy-staging.yml` and (conditionally) `deploy-production.yml` are inside
`.github/workflows/`, which `docs/governance/change-risk-classification.md`'s
path-based floor classifies at R3 (CI/CD and deployment pipeline). This package
proposes R3 for the change as a whole (see `change.yaml`'s
`planned_implementation_risk_floor`) — a draft proposal for the reviewing human at
adoption time, not a determination; the actual floor must be confirmed by running
`scripts/governance/classify-change-risk.sh` against each task's real file list at
implementation time, and may float higher if the implementer's chosen mechanism
(see Open Question 1 below) touches something this drafting pass did not
anticipate.

Protected/production effect: the new step runs directly against the real staging
database (and, if T02 is adopted, the real production database) during a live
deploy. The tool itself is additive/idempotent by design, which bounds — but does
not eliminate — the risk; a bug in a *new* invocation mechanism this package adds
(e.g. a build step, an SCP step, a new SSH command) is still new deploy-pipeline
surface regardless of how safe the existing tool's own SQL is.

## Decisions, contradictions, security, and privacy

No `VOC-052-D00`-numbered decision is recorded here because this package is not
yet adopted; per the template, decisions are only defined after approval. The
following are recorded as open questions for the reviewing human and/or the
eventual implementer, not resolved by this drafting pass:

1. **Build/run mechanism for `apps/api/cmd/seed` in CI (`VOC-052-DEP-01`).** The
   tool is a plain `go build` target with data embedded via `//go:embed`, unlike
   `migrate.sh` (wraps a prebuilt Atlas binary) or
   `seed-synthetic-smoke-user.sh` (raw SQL through `docker compose exec postgres
   psql`). Two viable mechanisms exist: (a) `go build` the binary on the GitHub
   Actions runner (which already has the pinned Go toolchain for `pnpm test:api`)
   and SCP it to the staging host as part of the existing deploy bundle, then run
   it there with `DATABASE_URL` pointed at the private Postgres bridge IP the same
   way `migrate.sh` already does; or (b) run it directly from the GitHub Actions
   runner against the staging database, which would require exposing the staging
   Postgres port beyond its current non-published state (a new, narrower exposure
   decision this package does not make). This package does not choose between
   them; the implementer must record which one is used and why.
2. **Whether `deploy-production.yml` parity (T02) is in scope now
   (`VOC-052-DEP-02`).** Production's real-backend core-loop smoke check already
   has session-minting wiring from VOC-050-T04, but full activation of
   production's real-backend smoke gating is a separate founder decision that has
   not yet been made (the issue itself notes `SMOKE_TEST_SESSION_COOKIE` is
   "intentionally left unset" on production today). Adding the seed step to
   `deploy-production.yml` ahead of that activation decision is new, unrequested
   scope on top of what issue #437 strictly asked for (which focuses on staging);
   deferring it leaves production with the same latent gap staging had. This
   package proposes T02 as an explicitly conditional task rather than silently
   including or silently omitting it.
3. **Failure mode if the seed step fails.** Should a seed failure abort the
   deploy under `set -e` before `docker compose up -d` (mirroring
   `seed-synthetic-smoke-user.sh`'s fail-closed behavior, leaving previously-running
   containers untouched), or should it be treated as a softer warning? This
   package proposes mirroring the existing fail-closed pattern (see
   `implementation-plan.md`) as the default, but flags it for explicit reviewer
   confirmation since a seed failure blocking every staging deploy is a stronger
   consequence than the synthetic-user seed's failure mode had when it was
   introduced.

No new secret, credential, or personal-data handling is introduced: the seed data
(`voc026-p1.json`) is static, non-personal, non-secret educational content already
committed to the repository, and `apps/api/cmd/seed` reads only the same
`DATABASE_URL` the migration step already uses.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** No new migration is added or changed by this package.
  The existing schema (VOC-026) is unchanged; this package only adds a step that
  populates already-existing, already-empty tables. The seed is idempotent
  (upsert by fixed primary key), so rerunning it on every deploy is safe and
  matches the existing `seed-synthetic-smoke-user.sh` posture.
- **Analytics:** None applicable — no new user-facing behavior, event, or metric
  is introduced. The real content simply becomes visible where the schema and
  frontend already expected it.
- **Accessibility:** None directly applicable to this package's own diff (it adds
  no new UI). Indirectly, once real content renders on `/discover`, the existing
  accessibility suite (which currently runs against `mock-api-server.mjs`'s fixed
  fixtures per this specification's "Confirmed findings" section) will, for the
  first time on staging, be exercisable against real rendered content in a
  follow-up sense — but re-pointing the accessibility suite at real staging is
  out of this package's scope (it targets the E2E core-loop check specifically,
  per the issue).
