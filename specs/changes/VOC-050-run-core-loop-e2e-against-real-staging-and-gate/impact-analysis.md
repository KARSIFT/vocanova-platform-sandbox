# VOC-050 — Impact Analysis

## Security and privacy

- Authentication surface: `T01`'s session-minting mechanism is new code
  capable of producing a valid session outside the normal magic-link/OAuth
  round-trip. It must be narrowly scoped (one synthetic account only),
  secret-gated, and fail closed when the gating token is unset - see
  `acceptance-criteria.md`'s `VOC-050-AC-02`. This is the single highest-risk
  surface this package introduces; the reviewing human should scrutinize it
  at least as closely as any other authentication-path change, per
  `docs/governance/change-risk-classification.md`'s R3 floor for that area.
- Secret handling: the minted session value (and any gating token) are
  credential-shaped and must follow this repository's existing `env:`-block
  convention for GitHub Actions secrets (never inlined into a script body
  where workflow logs would echo it), matching every other secret in
  `deploy-staging.yml`/`deploy-production.yml`.
- Synthetic-account isolation: the account must be unambiguously
  distinguishable from a real user (reserved domain and/or explicit flag)
  and must not be reachable via any real signup path - see
  `acceptance-criteria.md`'s `VOC-050-AC-00`/`AC-01`. `VOC-050-DEP-02` in
  `change.yaml` records that the exact guard mechanism is unconfirmed at
  drafting time and must be resolved by `T00`'s implementer against the
  actual current signup-path code, not assumed.
- No real user data, production secrets, or credentials beyond the
  synthetic account's own scoped session are introduced.

## Data and migrations

In scope. `T00` requires a new migration creating the synthetic test account
(and any supporting flag/column). The implementer must follow this
repository's documented Atlas lessons in `.karsift/lessons.md`:
`-- atlas:txmode file` (not the previously-invalid `transaction` value), and
must avoid the documented duplicate-unique-index pitfall class if the new
migration adds any unique constraint alongside an inline `UNIQUE` column
definition. The migration must be idempotent/rerun-safe (fixed identity,
`ON CONFLICT`-style upsert or equivalent), mirroring `apps/api/cmd/seed`'s
existing rerun-safe convention, since the issue requires automatic seeding
on every staging and production deploy, not a one-time manual step.

## Analytics and accessibility

- Analytics: None applicable, evidence-backed. A drafting-time repository
  search found no product-analytics pipeline (PostHog, Segment, or similar)
  anywhere in `apps/api` or `apps/web`. The issue's "must never appear in
  real product analytics" constraint is therefore currently satisfied by
  there being no analytics pipeline to leak into. This is a point-in-time
  finding, not a permanent guarantee - if an analytics system is added in a
  future package, that package must independently confirm this constraint
  still holds and add an explicit exclusion if needed.
- Accessibility: Not applicable to this package's own new surface (CI
  workflow steps, a seed script, a session-minting mechanism, and the
  staging-targeted core-loop variant's setup code). `core-loop.spec.ts`'s
  underlying journey already has separate, existing accessibility coverage
  via the T07a/T07b scans; this package does not change or duplicate that.

## Risks, dependencies, and evidence

- `VOC-050-R00`: A synthetic account that is imperfectly isolated could be
  discoverable or reachable by a real user (e.g. via a predictable email
  guessed through the magic-link request path), which would be a real
  security/privacy exposure, not just a test-hygiene issue. Mitigated by
  `acceptance-criteria.md`'s `VOC-050-AC-01` requiring the account to be
  unreachable via real signup paths, verified by `VOC-050-TEST-01`.
- `VOC-050-R01`: An imperfectly-scoped session-minting mechanism could become
  a general-purpose authentication bypass if its gating token were ever
  compromised or its scope check were buggy. Mitigated by
  `acceptance-criteria.md`'s `VOC-050-AC-02` and by mirroring the existing,
  already-reviewed `RegisterMonitoringSentryTest` fail-closed pattern rather
  than inventing a new one from scratch.
- `VOC-050-R02`: This package cannot guarantee end-to-end that a failed
  staging core-loop run actually blocks production auto-promotion, because
  the promoting workflow (`release.yml`) lives in a different repository
  (`karsift-ai-infra`) this planner run has no authority to modify. Mitigated
  by `acceptance-criteria.md`'s `VOC-050-AC-04` scoping the criterion to what
  this repository can actually guarantee (an accurate failure signal) and
  requiring `T03`'s pull request to explicitly flag the remaining cross-repo
  dependency for the reviewing human, rather than silently overstating what
  was delivered.
- `VOC-050-R03`: If `T02`'s staging core-loop variant reuses
  `core-loop.spec.ts`'s existing mock-only cookie-injection shortcuts
  unmodified, the test will not actually exercise real authentication and
  will give a false sense of coverage. Mitigated by `specification.md`'s
  "Confirmed findings" section calling this out explicitly and open question
  2/3 requiring `T02`'s implementer to resolve it with a real mechanism, not
  silently keep the mock-only shortcut against a real backend.
- `VOC-050-DEP-00`, `VOC-050-DEP-01`, `VOC-050-DEP-02`: see `change.yaml`.
- `VOC-050-EV-00`: `VOC-050-T00`'s implementing pull request, including the
  migration/seed diff and a demonstrated idempotent rerun.
- `VOC-050-EV-01`: `VOC-050-T01`'s implementing pull request, including the
  session-minting mechanism's diff and its fail-closed-when-unset test
  coverage.
- `VOC-050-EV-02`: `VOC-050-T02`'s implementing pull request, including the
  staging-targeted core-loop variant and `deploy-staging.yml`'s new step.
- `VOC-050-EV-03`: `VOC-050-T03`'s implementing pull request, including its
  explicit documentation of the cross-repo `release.yml` dependency.
- `VOC-050-EV-04`: `VOC-050-T04`'s implementing pull request, including
  `deploy-production.yml`'s updated smoke-test step and a demonstrated
  non-SKIP run.
