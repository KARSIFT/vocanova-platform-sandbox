# VOC-050 — Run core-loop E2E Against Real Staging and Gate Production Auto-Promotion On It

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #391](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/391),
prepared for founder/steward review at adoption time.

## Why this exists

`apps/web/tests/e2e/core-loop.spec.ts` is a full, real, authenticated
10-step user-journey Playwright test, but it has only ever run against
`mock-api-server.mjs`. `deploy-staging.yml` deploys on every merge to
`develop` and only polls `/healthz` — proof the containers started, not that
the app works against the real Go API and Postgres.
`deploy-production.yml`'s smoke-test suite already anticipates a real
core-loop check but it is a non-fatal `SKIP` because
`SMOKE_TEST_SESSION_COOKIE` was never provisioned. Combined with the
founder-authorized `develop` → `main` auto-promotion path (see `AGENTS.md`'s
"Release and deployment authority" section), a change can currently pass
mocked CI, deploy to staging, look "healthy," and reach production without
the real user journey ever being exercised against real infrastructure.

## What was confirmed during drafting

Reading the actual files (not just the issue text) confirmed every premise
the issue states, and surfaced one important complication:

- `core-loop.spec.ts`'s first step sets `vocanova_session`/`vocanova_csrf`
  cookies to **arbitrary values** the mock server accepts unconditionally.
  The real Go API does not have this behavior — running this exact spec
  unmodified against real staging/production will not work. A real session,
  minted through `T01`'s new mechanism, is required. See
  `specification.md`'s "Confirmed findings" for the full list.
- `release.yml` — the workflow that actually performs `develop` → `main`
  auto-promotion — lives in the **`karsift-ai-infra` repository**, not this
  one. This planner run has no authority to touch that repository. This
  package's `T03` can only make `deploy-staging.yml` fail closed on a
  core-loop failure; it explicitly cannot prove, and does not claim to
  prove, that `release.yml` gates its promotion decision on that signal.
  See `specification.md`'s open question 1 and `change.yaml`'s
  `VOC-050-DEP-01`.
- No product-analytics system exists anywhere in the codebase yet, so the
  issue's "must never appear in real product analytics" constraint is
  currently satisfied by there being nothing to leak into — a
  point-in-time, evidence-backed finding, not a permanent guarantee.

## What this package deliberately does NOT do

- It does not modify, or claim to modify, anything in the `karsift-ai-infra`
  repository. `release.yml`'s actual gating behavior is out of this
  package's reach; the reviewing human must separately confirm or arrange
  that cross-repo change.
- It does not change any R0–R4 risk classification rule or founder-approval
  requirement — the issue explicitly says this only adds a new gate ahead of
  the already-authorized auto-promotion path, and this package does not
  reinterpret that.
- It does not build a product-analytics system or retrofit analytics
  exclusion logic where nothing exists yet to exclude from.
- It does not touch the existing PR-time accessibility/E2E suite's use of
  `mock-api-server.mjs` — that suite's fast, deterministic, network-free
  purpose is unrelated and must keep passing unmodified.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) the cross-repo
`release.yml` gating dependency this package cannot resolve alone; (2)
whether the real backend should ever honor mock-only override cookies, even
narrowly scoped, versus seeding the synthetic account already past
onboarding; (3) whether the staging-targeted journey should be a modified
`core-loop.spec.ts` or a separate sibling spec file; and (4) whether the
synthetic account should be re-seeded/re-verified on every deploy or only
when absent (this package defaults to "every deploy, idempotently," mirroring
`apps/api/cmd/seed`'s existing convention, unless the reviewing human
disagrees).

## Structure

Mirrors recent packages' convention (e.g. VOC-048, VOC-047, VOC-049):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R2` package-level risk classification in
   `change.yaml`, and separately consider whether `T00` (migrations) and
   `T01` (a session-minting mechanism, however narrowly scoped) should be
   treated as R3 individually, per this repository's own
   `docs/governance/change-risk-classification.md` floors.
2. Decide whether the cross-repo `release.yml` gating dependency
   (`specification.md`'s open question 1) needs a companion change filed
   against `karsift-ai-infra` before or alongside adopting this package, or
   whether it can be confirmed as already-satisfied by that repository's
   existing behavior.
3. Resolve `change.yaml`'s `VOC-050-DEP-02` direction (the exact
   synthetic-account reachability-blocking mechanism) or leave it to `T00`'s
   implementer, as preferred.
4. Adopt (or request changes to) this package, then dispatch `T00` → `T01`
   → (`T02`, `T04`) → `T03`, consistent with `tasks.md`'s documented
   dependency order.
