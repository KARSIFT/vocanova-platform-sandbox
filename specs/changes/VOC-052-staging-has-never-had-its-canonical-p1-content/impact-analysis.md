# VOC-052 — Impact Analysis

## Security and privacy

No new secret is introduced. The seed step reuses the same `DATABASE_URL` value
`apps/api/scripts/migrate.sh` already consumes on the staging host (sourced from
`/opt/vocanova/infra/secrets/api.env`, founder-populated, never written by this
workflow) — no new credential, connection string, or environment variable needs to
be provisioned. If the implementer chooses mechanism (a) from
`specification.md`'s open question 1 (build on the runner, SCP the binary, run over
SSH), the binary itself carries no embedded secret — `voc026-p1.json` is static,
non-personal, non-secret educational content already committed to the repository.
If mechanism (b) is chosen instead (running against staging Postgres directly from
the GitHub Actions runner), that would require exposing the Postgres port beyond
its current non-published state, which is itself a new security-relevant surface
this package's drafting pass explicitly does not authorize or recommend — see
`VOC-052-R00` below.

No personal data is touched: every seeded table (`journey_situations`,
`canonical_words`, `word_meanings`, `word_examples`, `usage_notes`,
`journey_words`) holds static educational content, not user data.

## Data and migrations

No migration is added, changed, or reverted by this package. The existing VOC-026
schema is unchanged. The seed step populates already-existing, already-empty
tables via `INSERT ... ON CONFLICT (id) DO UPDATE SET ...` keyed on each row's
fixed ID — additive and idempotent, with no `DELETE` anywhere in
`apps/api/cmd/seed`'s `applySeed`. Rollback of this package (removing the new
deploy step) does not require reverting any data: the seeded rows are inert,
static content that causes no harm if left in place even if the workflow change
itself is rolled back.

## Analytics and accessibility

No new analytics event, metric, or user-facing behavior is introduced — this
package makes existing, already-built UI (the `/discover` page and its downstream
word views) render real data it was always designed to render. No accessibility
change is made to any component; the existing accessibility suite's coverage is
unaffected by this package (see `specification.md`'s accessibility section for why
re-pointing that suite at real staging content is explicitly out of scope here).

## Risks, dependencies, and evidence

- `VOC-052-R00`: Exposing the staging Postgres port to the GitHub Actions runner
  (specification.md's open-question-1 mechanism (b)) would be a new, unrequested
  security-relevant surface change beyond what issue #437 asked for. This package
  does not choose that mechanism and flags it as a risk to avoid rather than a
  recommended path; mechanism (a) (build-and-SCP, run over the existing SSH
  session against the private bridge IP, exactly as `migrate.sh` already does)
  keeps the same trust boundary the repository already accepted for migrations.
- `VOC-052-R01`: A `go build ./cmd/seed` step added to `deploy-staging.yml` adds
  build time and a new failure mode (a Go compile error) to every staging deploy,
  where previously only `apps/api/Dockerfile`'s own multi-stage build compiled Go
  code. This is a maintainability/reliability risk (a new place a deploy can fail)
  rather than a security risk, and is mitigated by CI already running
  `pnpm test:api` (which itself compiles `apps/api`) on every PR before merge, so
  a compile failure at deploy time would be a novel, not a previously-latent,
  problem.
- `VOC-052-R02`: If T02 (`deploy-production.yml` parity) is adopted ahead of the
  founder's separate decision to fully activate production's real-backend smoke
  gating, the new step becomes live production-database-writing surface before
  that gating decision is made — see `VOC-052-DEP-02`. This package proposes T02
  as explicitly conditional on that decision rather than assuming it.
- `VOC-052-DEP-00`: Resolved at drafting time — the root-cause findings in
  `specification.md`'s "Confirmed findings" section were independently verified
  by reading the actual files, not merely assumed from the issue text.
- `VOC-052-DEP-01`: Unresolved at drafting time — see `specification.md` open
  question 1 (build/run mechanism choice for `apps/api/cmd/seed`).
- `VOC-052-DEP-02`: Unresolved at drafting time — see `specification.md` open
  question 2 (whether `deploy-production.yml` parity is in scope now).
- `VOC-052-EV-00`: Required evidence — the exact `deploy-staging.yml` diff, a
  passing workflow run log showing the new seed step executing and reporting its
  row counts (`apps/api/cmd/seed`'s own `fmt.Printf` summary line), and
  confirmation the step ran after migrations and before (or as part of) bringing
  the stack up.
- `VOC-052-EV-01`: Required evidence — a passing
  `tests/staging-e2e/core-loop.staging.spec.ts` run against real staging,
  post-seed, specifically showing the discover-step assertion
  (`situationLinks.count()`) now returns a positive count, plus the full spec
  passing end-to-end.
- `VOC-052-EV-02`: Required evidence, only if T02 is adopted — the same class of
  evidence as `VOC-052-EV-00`/`VOC-052-EV-01` but for `deploy-production.yml` and
  a production-safe verification path (see `test-plan.md`'s constraint against
  using production data in tests).
