# VOC-050 — Run core-loop E2E Against Real Staging and Gate Production Auto-Promotion On It: Specification

## Objective and requirement source

Close the gap [issue #391](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/391)
describes: a change can currently pass mocked CI, deploy to staging, look
"healthy" (an unauthenticated `/healthz` 200), and auto-promote straight to
production (per `AGENTS.md`'s "Release and deployment authority" section)
without the real, authenticated core user journey ever being exercised
against the real Go API and Postgres. The objective is to run the existing
`apps/web/tests/e2e/core-loop.spec.ts` journey for real against staging after
every staging deploy, make a failure of that run block production
auto-promotion, and turn the existing non-fatal production core-loop SKIP
into a real check.

## Confirmed findings (made during drafting, not assumed)

- `apps/web/tests/e2e/core-loop.spec.ts` is a real, complete, 10-step
  Playwright journey (auth via directly-set cookies, onboarding, discover,
  save, review, sentence feedback, progress, settings, logout, auth-gate
  rejection) — confirmed by reading the file directly. It is currently wired
  only into `apps/web/tests/e2e/mock-api-server.mjs`-backed runs (the
  accessibility/E2E suite at PR time), not against any real deployed
  environment.
- The test's step 1 sets `vocanova_session`/`vocanova_csrf` cookies to
  **arbitrary, test-generated values** (`test-session-${randomUUID()}`) and
  relies on the mock server accepting any such value as a valid session. The
  real Go API (`apps/api/business/auth`) does not have this behavior — a
  session cookie must correspond to a real row created by `CreateSession`
  after a real magic-link consumption or OAuth callback. **Running this exact
  spec unmodified against real staging/production will not work** — the test
  (or a variant of it) needs a way to start from a cookie that the real
  backend actually recognizes as authenticated, which is exactly the
  "mint its session server-side in CI" mechanism the issue asks for.
- The test's step 10 and its onboarding-shortcut in step 1 rely on
  `e2e_onboarding_status` / `e2e_unauthenticated` cookies, which
  `mock-api-server.mjs` interprets as unconditional overrides. Whether the
  real Go API honors (or should ever honor) these same test-only cookie
  names outside of test/CI builds was not found anywhere in
  `apps/api/business/auth` during drafting — no matching code was found.
  This package does not assume the real backend already has, or should
  gain, an equivalent override; see open question 2.
- `.github/workflows/deploy-staging.yml`'s only post-deploy verification is
  two polling steps (`Poll api-staging.vocanova.site/healthz` and
  `Poll staging.vocanova.site/`), confirmed by reading the file — no
  authenticated or functional check exists today.
- `infra/scripts/smoke-test-production.sh`'s core-loop section (its own `# 5.`
  comment) is gated entirely on `SMOKE_TEST_SESSION_COOKIE` being non-empty;
  when unset it prints an explicit `SKIP:` line and exits 0 for that section
  (confirmed by reading the script). `.github/workflows/deploy-production.yml`'s
  "Run production smoke-test suite" step never sets that variable, confirmed
  by reading the step's `env:` block — the SKIP is the entire current state,
  matching the issue's description exactly.
- `apps/api/cmd/seed` already exists as a precedent for an idempotent,
  rerun-safe seeding command (loads canonical content by fixed primary keys,
  safe to rerun), but it seeds product content, not a user account. No
  existing mechanism seeds a synthetic user account or mints a session
  outside the real magic-link/OAuth flow was found anywhere in `apps/api`
  during drafting.
- `release.yml` (the workflow that performs `develop` → `main` auto-promotion
  per `AGENTS.md`'s "Release and deployment authority" section) lives in the
  **karsift-ai-infra repository**, not this repository — confirmed by
  `AGENTS.md`'s own text ("karsift-ai-infra's `release.yml`..."). This
  planner run is scoped to `vocanova-platform-sandbox` only and has no
  authority to write into a different repository; see open question 1.
- No product-analytics system (PostHog, Segment, or similar) exists anywhere
  in `apps/api` or `apps/web` as of drafting time — confirmed by a repository
  search finding no matches. The issue's "must never appear in real product
  analytics" constraint is therefore currently satisfied by there being no
  analytics pipeline for the synthetic account to leak into; if an analytics
  system is added later, that future work must independently honor this
  constraint (recorded as a forward-looking note, not assumed to remain true
  forever).

## Scope and non-goals

In scope:
- A dedicated, obviously-synthetic test user account, seeded automatically
  (not manually) on both staging and production, via a migration/seed
  mechanism that is safe to rerun on every deploy.
- A server-side mechanism (no real magic-link email, no real OAuth
  round-trip) to mint a valid session for that account, invoked from CI, so
  `SMOKE_TEST_SESSION_COOKIE` (production) and a staging equivalent are
  available automatically.
- Adding a real, authenticated core-loop check to `deploy-staging.yml`, run
  immediately after the existing `/healthz` poll passes, against
  `https://staging.vocanova.site`.
- Making a failed staging core-loop run fail `deploy-staging.yml`'s job
  itself (fail-closed, no `continue-on-error`) so that whatever downstream
  mechanism gates `develop` → `main` promotion has a real signal to act on -
  see open question 1 for the limit of what this package alone can guarantee
  about the downstream gate.
- Wiring `infra/scripts/smoke-test-production.sh`'s already-anticipated
  core-loop check to actually run in `deploy-production.yml`, by providing
  `SMOKE_TEST_SESSION_COOKIE` from the same minting mechanism, while
  preserving the script's existing "no state-mutating action against a real
  target" design (this package's core-loop check must use read-only/already
  present-safe requests, matching the script's existing pattern, not extend
  it into new state-mutating territory).

Out of scope:
- Any change to `release.yml`, or any other file, in the `karsift-ai-infra`
  repository. This package can only change files inside
  `vocanova-platform-sandbox`; see open question 1.
- Any change to R0–R4 risk classification rules or founder-approval
  requirements — the issue explicitly states this is not being changed, only
  a new required gate ahead of the already-authorized R0–R2 auto-promotion
  path.
- Building a product-analytics system, or retrofitting analytics exclusion
  logic that does not yet have anything to exclude from (see "Confirmed
  findings" above).
- Any change to the existing PR-time accessibility/E2E suite's use of
  `mock-api-server.mjs` — that suite's purpose (deterministic, fast,
  network-free CI checks) is unrelated to this package's real-infrastructure
  goal and must keep passing unmodified.

## Risk and protected areas

Builder assessment: this package's tasks touch several areas with different
individual risk profiles:
- `apps/api/migrations` (new migration for the synthetic account) is a
  protected, R3-floor-or-higher area per
  `docs/governance/change-risk-classification.md` and the
  `.karsift/lessons.md` Atlas-directive lessons already on file for this
  repository — the implementer must follow those lessons (`-- atlas:txmode
  file`, not `transaction`; avoid the documented duplicate-unique-index
  pitfall) rather than repeating a previously-diagnosed mistake.
- `apps/api/business/auth` (any session-minting mechanism) is squarely the
  authentication protected area, an R3-floor category per the same
  classification document, since it is code capable of producing a valid
  session without a real magic-link/OAuth round-trip - even though it is
  intended for CI-only use, it is still new code in the trust boundary that
  decides who gets a valid session.
- `.github/workflows/*.yml` changes are, per this repository's own precedent
  (`deploy-staging.yml`'s own header comment cites the workflow path-class
  floor), at least R3 on their own.
- The package-level risk proposed in `change.yaml` is R2, matching the
  issue's own stated classification and reflecting that no task changes real
  user-facing product behavior - but the reviewing human should treat T00
  (migration/seed) and T01 (session minting) as the tasks most likely to
  float to R3 individually at `scripts/governance/classify-change-risk.sh`
  time, per this package's own `planned_implementation_risk_floor` note.

## Decisions, contradictions, security, and privacy

No `VOC-050-D0x` decisions are defined here; none may be defined before
adoption. Recording the constraints any implementing task must satisfy:

- The synthetic account must be unambiguously distinguishable from a real
  user at every layer it touches (database row, session, logs) - e.g. via a
  reserved, non-deliverable email domain and/or an explicit boolean flag
  (`is_synthetic_test_account` or equivalent) set at creation and never set
  by any real user-facing signup path.
- The synthetic account must not be reachable via any real signup path
  (magic-link request, OAuth) - only the seed mechanism itself may create or
  authenticate it. `VOC-050-DEP-02` in `change.yaml` records that the exact
  guard mechanism is not yet confirmed against the current signup code and
  must be resolved by `T00`'s implementer, not assumed here.
- The session-minting mechanism must not become a general-purpose "mint any
  session for any user" backdoor. It must be narrowly scoped to the one
  synthetic account, gated by a secret token analogous to
  `MONITORING_TEST_TOKEN`'s existing pattern
  (`apps/api/app/api/production.go`'s `RegisterMonitoringSentryTest`), and
  must not be reachable at all when its gating token is unset (mirroring
  that function's `if strings.TrimSpace(expectedToken) == "" { return }`
  fail-closed pattern).
- `infra/scripts/smoke-test-production.sh`'s existing no-side-effects design
  (documented in its own header comment) must not be weakened. The
  production core-loop check this package activates must only perform
  read-only or already-idempotent requests against the synthetic account's
  own data (e.g. `GET /api/v1/me`, `GET /api/v1/journey-situations`, both
  already anticipated in the script), not introduce a new state-mutating
  request.
- No production secrets, credentials, or real user data are introduced. The
  new `SMOKE_TEST_SESSION_COOKIE`-equivalent value is itself a
  credential-shaped secret (it grants access to the synthetic account) and
  must be handled with the same `env:`-block-not-inlined GitHub Actions
  convention every other secret in `deploy-staging.yml`/`deploy-production.yml`
  already uses.

## Data, migrations, analytics, and accessibility

- Data/migrations: In scope. A new migration is required to create the
  synthetic test user (and any supporting flag/column distinguishing it from
  a real account). Must follow this repository's own documented Atlas
  lessons (`.karsift/lessons.md`) for directive syntax and avoiding the known
  duplicate-unique-index pitfall class.
- Analytics: None currently applicable - confirmed by drafting-time search
  finding no analytics pipeline anywhere in the codebase (see "Confirmed
  findings" above). This is an evidence-backed `None`, not an unchecked
  assumption.
- Accessibility: Not applicable to this package's own new surface (CI
  workflow steps, a seed script, a session-minting mechanism). The
  `core-loop.spec.ts` journey itself already has separate, existing
  accessibility coverage via T07a/T07b's scans (per the file's own header
  comment); this package does not change or duplicate that coverage.

## Open questions

1. **Cross-repo release-gating dependency.** This package can only change
   files inside `vocanova-platform-sandbox`. The issue's requirement 3 ("a
   failed staging core-loop run is a hard block on auto-promotion... release.yml's
   auto_release path must not fire... if the post-deploy staging E2E run
   failed") requires `karsift-ai-infra`'s `release.yml` to actually consult
   `deploy-staging.yml`'s conclusion for the relevant commit before
   promoting. This package's `T03` can only guarantee that
   `deploy-staging.yml` itself fails closed (no `continue-on-error`,
   non-zero exit) when the core-loop check fails; it cannot guarantee
   `release.yml` on the other repository actually gates on that signal
   today, and cannot modify that file to make it do so. The reviewing human
   must separately confirm (or file a companion change against
   `karsift-ai-infra`) that `release.yml`'s auto-promote decision genuinely
   depends on `deploy-staging.yml`'s conclusion, not merely on the target
   package's own task-PR checks.
2. **Whether the real backend should honor `e2e_onboarding_status`/
   `e2e_unauthenticated`-style test-only override cookies at all, even
   narrowly scoped to the synthetic account.** The mock server's shortcuts
   exist purely for fast, deterministic CI; extending any equivalent to the
   real backend (even gated to one synthetic account) is a new trust
   decision this specification does not resolve. The alternative - having
   the staging/production core-loop check perform the *real* onboarding flow
   end-to-end against the synthetic account, or seeding the synthetic
   account already past onboarding via `T00`'s seed step, avoiding the need
   for any override cookie at all - is the option this specification leans
   toward without mandating, since it needs no new trust surface on the real
   backend. `T02`'s implementer must pick and document one.
3. **Whether one shared `core-loop.spec.ts` file with environment-aware setup,
   or a separate staging-specific spec file, is the right shape** given the
   real-backend variant cannot reuse the mock-only cookie-injection step
   (see "Confirmed findings" above). Left to `T02`'s implementer to propose,
   consistent with this repository's existing pattern of separate,
   purpose-specific files (e.g. `accessibility.yml` vs. `lighthouse.yml`)
   rather than one file trying to serve every environment.
4. **Whether the synthetic account needs to be re-seeded (or its session
   re-minted) on every single deploy, or only when absent.** The issue asks
   for automatic seeding "on staging and production deploy"; whether that
   means "ensure it exists, idempotently, every time" (recommended, mirrors
   `apps/api/cmd/seed`'s existing rerun-safe convention) or "only on first
   deploy" is left to `T00`'s implementer, defaulting to the idempotent
   every-deploy behavior unless a reason emerges not to.
