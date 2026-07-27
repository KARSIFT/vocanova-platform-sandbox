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
