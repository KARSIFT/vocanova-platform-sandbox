# VOC-031 — P5 Integrated Core Loop Mock Disposition Inventory

## Scope and authority

This document is drafted before adoption/implementation, per this repository's
own package template. It is updated at `T11` implementation time to record
the actual post-`T00`–`T10` state.

## Draft-time confirmation (2026-07-27, by direct repository inspection)

Unlike every prior milestone (VOC-025 through VOC-030), which each retired
one or more `MOCK_*` frontend constants introduced by an earlier milestone,
**VOC-031 inherits zero legacy mocks.** A repository-wide search of
`apps/web/src` (`grep -rn "MOCK_" apps/web/src` and a broader case-insensitive
`grep -rni "mock"`) returns no matches at all — P1–P4's mock retirement is
already complete: `MOCK_HOME_STATE` and `MOCK_PROGRESS_STATE` were both fully
retired by `VOC-030-T05` (see
`specs/changes/VOC-030-begin-milestone-p4-daily-habit-and-progress-per/mock-inventory.md`),
and no earlier milestone left anything else behind.

This means VOC-031's own mock-inventory obligation is different in kind from
every prior milestone's: there is nothing inherited to retire. This
document's job at `T11` is instead to:

1. Confirm the zero-legacy-mock state still holds immediately before this
   package's own work begins (re-run the same search at the adopted base
   SHA, since time may have passed since this draft).
2. Record any mock this package's *own* tasks temporarily introduce (expected
   to be none — every task in `tasks.md` pairs its frontend wiring with a
   real backend contract in the same or an immediately preceding task,
   specifically to avoid needing one) and its disposition before this
   package is considered complete.
3. Extend `scripts/foundation/mock-inventory.mjs`'s allow lists for the new
   `users`/`accounts` modules, tables, and routes, and add a check that no
   P5-invented route/table/behavior beyond this package's own documented
   scope was introduced (mirrors `VOC-030-T06`'s "no P5 route/table/behavior"
   sweep, adapted to sweep for anything beyond VOC-031's own scope instead).

## Expected new real modules/tables/routes (added by `T00`–`T09`)

| Area | New addition | VOC source |
| --- | --- | --- |
| Business module | `apps/api/business/users/` | DOC-06 §3 (T00, T02) |
| Business module | `apps/api/business/accounts/` | DOC-06 §3 (T03, T04) |
| Ent schema | `useronboardingprofile.go` | DOC-05 §6 (T00) |
| Ent schema | `emailchangelink.go` | DOC-06 §6 (T03) |
| Ent schema | `accountdeletionrequest.go` | DOC-05 §16 (T04) |
| Migration | onboarding-profile migration | DOC-05 §6 (T00) |
| Migration | email-change-link migration | (T03) |
| Migration | account-deletion-request migration + `idempotency_keys.scope` widening | DOC-05 §13/§16, `VOC-031-D09` (T04) |
| API route | `GET`/`POST /api/v1/onboarding` | DOC-03 §3 (T01) |
| API route | `GET`/`PATCH /api/v1/settings` | DOC-08 (T02) |
| API route | `POST /api/v1/settings/email-change-links` (+ `/consume`) | DOC-06 §6 (T03) |
| API route | `POST /api/v1/account-deletion-requests` | DOC-07 (T04) |
| API client methods | `@vocanova/api-client` onboarding/settings/email-change/account-deletion methods | (T01–T04) |
| Web route | `apps/web/src/app/onboarding` | DOC-03 §3 (T01) |
| Web route | `apps/web/src/app/(app)/settings`, `.../settings/account` | DOC-08 (T05) |
| Test tooling | `apps/web/tests/e2e/` (Playwright), axe-core integration | DOC-03 §10, DOC-08 (T07, T08) |
| CI config | Lighthouse CI configuration | DOC-08 quality standards (T09) |

## Verified boundaries (to be enforced by the extended `scripts/foundation/mock-inventory.mjs`)

- No `MOCK_*` object (or a renamed equivalent) may appear in `/onboarding`,
  `/settings`, or `/settings/account`. Because each of these routes' backend
  contract lands before or alongside its frontend wiring in the fixed
  `T00`–`T05` order, no interim mock should ever be needed; if implementation
  finds one was necessary, it must be recorded here with an explicit
  disposition and removal task, not left silently in place.
- The three new tables, two new business modules, and five new API routes
  above are the **only** P5 additions the check should admit; anything else
  found is flagged as scope drift.
- `appLanguage`'s restriction to `en`-only (`VOC-031-D06`) is asserted as a
  real, enforced validation rule, not a UI-only pretense with a permissive
  backend.

## T11 follow-up (to be completed at implementation time)

- Confirm point 1 above still holds at the adopted base SHA.
- Record the actual disposition of any temporary mock introduced during
  `T00`–`T10` (expected: none).
- Link the evidence produced by the extended mock-inventory check
  (`VOC-031-EV-42`).

### T11 post-implementation state (2026-07-28, attempt 1)

The three draft-time follow-up points are resolved as follows, in the order
they appear above.

#### Point 1: zero-legacy-mock confirmation re-run at the adopted base SHA

A repository-wide search at the adopted base SHA (post-`T00`–`T10` PR #171)
confirms the draft-time finding: `grep -rn "MOCK_\|FAKE_\|FALLBACK_\|PLACEHOLDER_\|DEMO_"
apps/web/src/` returns matches only inside `apps/web/tests/e2e/mock-api-server.mjs`
and `apps/web/playwright.config.ts` (the `MOCK_API_PORT`/`MOCK_API_HOST`
environment-variable names) and inside `apps/web/tests/e2e/*-accessibility.spec.ts`
documents that explicitly reference the mock API server — not data
mocks of the kind the prior milestones retired. The `(app)` group
(`apps/web/src/app/(app)/`) — the only surface this inventory's
fabricatedDataPatterns static guard covers — returns zero `MOCK_*` / `FAKE_*`
/ `FALLBACK_*` / `PLACEHOLDER_*` / `DEMO_*` matches. Every previously
shipped A1–P4 mock was already retired by `VOC-030-T05` (`MOCK_HOME_STATE`,
`MOCK_PROGRESS_STATE`) and the P1/P2 mocks retired earlier in the chain
(`MOCK_DISCOVER_SITUATIONS`, `MOCK_SITUATION_WORD_LISTS`); none of these
re-appeared at the post-T10 base. P5 inherits a clean app source tree
and ships with zero data mocks of its own. This is the `VOC-031-EV-42`
zero-legacy-mock confirmation.

#### Point 2: disposition of any temporary mock introduced during T00–T10

**Zero temporary mocks were introduced by `T00`–`T10`.** Every task in the
P5 ordered PR sequence (`T00` → `T11`) paired its frontend wiring with a
real backend contract in the same or an immediately preceding task,
specifically to avoid the need for an interim mock (per the
`implementation-plan.md` file-reconciliation section). The
`fabricatedDataPatterns` static guard the `T06` install added and the new
`T11` extended check (see point 3) confirm this at the code level: no
hardcoded `MOCK_*` / `FAKE_*` / `FALLBACK_*` / `PLACEHOLDER_*` / `DEMO_*`
array or object literal exists anywhere in the `(app)` source tree, and
the T05 Settings + T05 Settings/account + T01 Onboarding routes that
this package introduces contain no such literals. A future regression
that re-introduced one would surface immediately in the
`scripts/foundation/mock-inventory.mjs` run (exit code 1 with a
file/pattern pinpoint).

#### Point 3: link to the evidence produced by the extended mock-inventory check

`VOC-031-EV-42` is the `scripts/foundation/mock-inventory.mjs` run output
itself, recorded in this package's CI workflow (`.github/workflows/pipeline.yml`).
The `T11` extension of the check is the static guard the package's
documentation-level claim — "no P5-invented route/table/behavior beyond
this package's own documented scope" — is enforced against. Specifically,
`T11` adds:

- `expectedRouteDirectories` now includes the `T05` Settings + Settings/
  account `(app)` route directories (presence check, file system).
- A new `expectedTopLevelRouteDirectories` list (with its own presence
  check) covers the `T01` `/onboarding` route, which — per `DOC-03 §3`'s
  "redirect to `/onboarding` before any `(app)` route" flow — lives at
  the top of `apps/web/src/app/` (outside the `(app)` group) and so is
  not covered by the existing `(app)`-only `expectedRouteDirectories`
  check.
- The `apps/web/src/middleware.ts` `matcher` `requiredPatterns` list now
  requires the T05 additions `"/settings"` and `"/settings/:path*"` in
  addition to the previously listed `"/onboarding"` (T01), so the
  auth/CSRF middleware gate is verified to cover every P5 route a
  learner can reach.
- A `D06` source-level check pins
  `apps/api/business/users/settings.go`'s `SupportedAppLanguages` to
  exactly `[]string{"en"}`, so a future contributor who silently widens
  the set (e.g. to admit a second placeholder locale) would surface
  here, in the existing service-layer test, and in the OpenAPI
  `enum:"en"` annotation the T02 PR added. The "no multi-language picker
  before i18n infrastructure exists" guarantee is pinned in three
  places; the static guard here is the one that catches a contributor
  who updates only the data structure without updating the test or the
  OpenAPI annotation.
- A T11 deliverable-presence check pins the `T06`/`T09`/`T10` audit
  findings (recorded in this same `staging-evidence.md`) as the
  in-repository evidence T11 cites, so a future contributor who silently
  truncated any of them would surface here.
- A `T11` P5-forbid-patterns sweep applies the same
  `p5[_-]?leaderboard`/`p5[_-]?badge`/`p5[_-]?reward[_-]?store` pattern
  guard the T06 / T07a / T09 sweeps already use, but to T11's own
  deliverable surface (this file, `staging-evidence.md`, and the
  extended `mock-inventory.mjs` itself), so a future contributor cannot
  introduce out-of-scope P5 behavior through T11 either.

The complete test surface for the T11 extension is in
`scripts/foundation/mock-inventory.test.mjs` (two new tests: the
T11-run-the-check test and the T11-`SupportedAppLanguages`-pinned test,
both added in the same PR as this file). The `node --test
scripts/foundation/mock-inventory.test.mjs` run reports `8 passed / 0
failed` at this PR's working-tree state; the script's stdout is
`VOC-031-T11 mock inventory validation passed (T07a accessibility
scaffolding + T06 cross-cutting reliability deliverables + T09
performance-automation scaffolding + T11 P5 completeness check present).`
No error path of the check fires; no `MOCK_*` / `FAKE_*` / `FALLBACK_*`
/ `PLACEHOLDER_*` / `DEMO_*` literal exists in the `(app)` source tree;
the only API paths the API-route-files scan finds are inside the
allow-list; the only Ent schemas and migrations present are inside the
allow-list; the only business modules present are inside the allow-list;
the middleware matcher covers every required P5 route; and the T11
deliverable presence checks all pass.

#### Relationship to the package's other deliverables

The T11 extension is the last guard rail the P5 ordered PR sequence
adds; T07a, T09, and T11 each install their own presence check
(`t07aScaffolding`, `t09Scaffolding`, and the T11 P5 completeness check
respectively), and the T06 P5-forbid-patterns sweep is mirrored by the
T07a, T09, and T11 sweeps for the same reason — each task's own
deliverable surface needs its own pattern guard, not just an inherited
one from an earlier task. The pattern is the same one the prior
milestones (VOC-027, VOC-030) used to gate the end of a package's
mock-decommission chain.
