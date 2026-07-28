# VOC-031 — P5 Integrated Core Loop Staging and Rollback Evidence

## Purpose

This document records the evidence required by `VOC-031-AC-11` (via
`VOC-031-TEST-41`..`VOC-031-TEST-45`). It is drafted before adoption/
implementation; it is updated at `T11` implementation time to record the
in-repository evidence actually produced by the merged `T00`–`T11` PRs, and
to document the staged exercises and rollback rehearsal that can only run
once F3 staging exists — following the exact pattern
`VOC-025-DEP-01`/`staging-evidence.md` and `VOC-030-DEP-02`/`staging-evidence.md`
established for this repository.

## Current status

Execution of live staging evidence is **blocked** by `VOC-031-DEP-02`: the F3
staging gate and supporting staging environment do not yet exist. No P5-complete
declaration is made here, and none can be made until `T00`–`T11` are
implemented and F3 exists.

## Planned in-repository evidence (produced by `T00`–`T11`, recorded at `T11` time)

| Evidence | Requirement | Status / source |
| --- | --- | --- |
| `EV-00`..`EV-03` | Onboarding schema/migration and seed-rule domain logic | Not yet produced — `T00` |
| `EV-04`..`EV-07` | Onboarding API, gating, `/me` additive field | Not yet produced — `T01` |
| `EV-08`..`EV-11` | Settings read/write, upsert-race safety, non-retroactivity | Not yet produced — `T02` |
| `EV-12`..`EV-17` | Email-change verification: token mechanics, race safety, session/OAuth non-impact | Not yet produced — `T03` |
| `EV-18`..`EV-23` | Account deletion: deactivation, idempotency, anonymization sweep, scope enum | Not yet produced — `T04` |
| `EV-24`..`EV-27` | Settings/account frontend flows | Not yet produced — `T05` |
| `EV-28`..`EV-31` | Reliability/recovery pass, A1–P4 regression check | **Produced by `T06` — see `T06 audit findings` below** |
| `EV-32`..`EV-35` | Accessibility automation install and CI wiring | Not yet produced — `T07` |
| `EV-36` | Full core-loop end-to-end suite | Not yet produced — `T08` |
| `EV-37`..`EV-39` | Performance automation install and CI wiring | **Produced by `T09` — see `T09 audit findings` below** |
| `EV-40` | UX-consistency audit record | **Produced by `T10` — see `T10 audit findings` below** |
| `EV-41`..`EV-42` | Installed suite pass, extended mock-inventory check | Not yet produced — `T11` |
| `EV-43` | Staging full-loop and cross-user-denial exercise | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-44` | New-tables rollback rehearsal | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-45` | Exact-SHA independent verification (per PR) | **Performed by Claude Code (different model binding) at each PR's exact final SHA** — this is not the implementer's evidence to record; it is produced by the independent-verification role and reported per-PR. |

## Staging exercise plan (blocked by F3 + the open decisions)

Once `VOC-031-DEP-02` is resolved and a non-production F3 environment is
available, the following exercises must be executed and their results
appended to this document or linked from PR evidence.

### `EV-43` — Full-loop and cross-user-denial exercise

1. With a non-production learner identity that has no pre-existing
   onboarding profile, complete the onboarding flow, verify redirection to
   `/home`, and confirm `daily_review_target` was seeded per `VOC-031-D04`
   (or preserved, if a prior customization already existed).
2. Exercise the full core loop (discover → save → review → sentence
   practice → progress) as the DOC-10 §7 flow describes.
3. On `/settings`, change each of the six editable fields and confirm the
   change is reflected on reload; confirm a `dailyReviewTarget` change does
   not alter the current local day's already-created mission target.
4. On `/settings/account`, request and confirm an email change with a
   non-production test-inbox address; confirm the old-email notification is
   received and the current session remains valid throughout.
5. On `/settings/account`, complete the account-deletion confirmation flow
   for a disposable non-production identity; confirm immediate logout,
   confirm the account can no longer authenticate, and (after the sweep
   runs, forced past `purge_after` for the test) confirm the DOC-05 §16
   disposition was applied per table.
6. Create a second non-production identity and confirm it can never reach
   the first identity's onboarding profile, settings, email-change state, or
   deletion-request state (404, not 403, per the existing A1 convention).
7. Repeat step 4's email-change flow with a replayed request (no
   `Idempotency-Key` applicable there, so instead: replay the same
   consumption token) and step 5's deletion with a replayed
   `Idempotency-Key`, and confirm neither produces a duplicate effect.

### `EV-44` — New-tables rollback rehearsal

1. Record current `user_onboarding_profiles`, `email_change_links`, and
   `account_deletion_requests` state for a test learner.
2. Apply the VOC-031 build and migrations in staging.
3. Run several onboarding/settings/email-change/deletion-request flows, then
   perform a rollback to the previously known-good revision.
4. Verify that:
   - No account is left in a state where it was deactivated by a
     deletion request but the pre-rollback code cannot resume its sweep.
   - `user_settings`/onboarding state committed before the rollback is
     preserved.
   - The A1–P4 write paths remain functional after rollback (nothing this
     package adds is on their critical path — `T02`–`T04` are purely
     additive new surfaces).

## Rollback triggers

Per `VOC-031` implementation-plan §Deployment and rollback / release-plan
§Rollback, initiate rollback on:

- A wrong-account or unauthorized account deletion or email change reaching
  production.
- An account-deletion sweep leaving inconsistent state.
- An onboarding gate locking a learner out of the loop with no recovery
  path.
- A regression in any A1–P4 screen or write path.
- Migration or schema failure.
- A new accessibility/performance automation gate that was bypassed rather
  than honestly reported as failing.

## Rollback procedure

1. Preserve `user_onboarding_profiles`/`user_settings` state committed
   before the rollback.
2. Ensure no `account_deletion_requests` row is left in a state neither the
   pre-rollback nor post-rollback code can resolve (complete or safely pause
   any in-flight sweep before rolling back its owning code).
3. Restore the pre-P5 A1–P4 behavior cleanly (verify no A1–P4 write path
   depended on anything this package added).
4. Revert the deployment to the last-known-good revision.
5. Validate with non-production identities.
6. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- **`VOC-031-DEP-02`**: F3 staging does not exist, so `EV-43`/`EV-44` cannot
  be run live. Procedures are documented; live execution is recorded as
  blocked.
- **`VOC-031-D02`/`D03`/`D05`–`D07`/`D09`**: open or this-draft-proposed
  decisions must be resolved or confirmed at adoption before the affected
  tasks' evidence can be treated as final.
- **`VOC-031-DEP-03`**: account-deletion production enablement additionally
  requires a separate founder go/no-go and the DOC-05 §16 legal review; this
  is not resolved by staging evidence alone.
- **`VOC-031-DEP-04`**: new Playwright/axe-core/Lighthouse CI tooling
  reconciliation against the adopted base's CI runner is required before
  `T07`/`T09` can be considered complete.

## `T06` audit findings (VOC-031-T06 / VOC-031-TEST-28..31)

`T06` audits every (app) route (Home, Discover, Word Detail, Reviews,
sentence feedback, Progress, Onboarding, Settings, Settings/account) for the
three cross-cutting properties the DOC-12 §5 P5 gate depends on: safe
retry path on network failure; defined behavior when a session expires
mid-flow; and no client-fabricated fallback value. The audit, the gaps
found, and the fixes applied are recorded here so the P5 gate
readiness section can be evaluated without re-reading the PR diff.

### Audit method

1. The per-screen code was read end-to-end, focusing on each client
   component's `try/catch` path and each server component's
   `try/catch` path.
2. The existing per-route auth/CSRF/idempotency tests in
   `apps/api/app/api/{learning,reviews,aifeedback,missions,onboarding,
   settings,email_change,account_deletion}_test.go` were re-run to
   confirm the A1–P4 + P5-T01..T04 behavior is intact.
3. A new cross-cutting test file
   `apps/api/app/api/core_loop_reliability_test.go` was added with
   the TEST-29 (session-expiry mid-flow at the API layer), TEST-30
   (no fabricated fallback in the contract), and TEST-31 (A1–P4
   contract-surface regression) counterparts.
4. A new client-side session-expiry helper
   `apps/web/src/lib/session.ts` was added; every client component
   in the (app) group was rewired to use it.
5. The `scripts/foundation/mock-inventory.mjs` static check was
   extended with a no-fabricated-fallback scan over (app) routes and
   a presence check for the T06 deliverables.

### Gaps found and fixed

| Gap | Surface | Fix |
| --- | --- | --- |
| Client components caught `ApiResponseError` and rendered a generic "try again" message for every status, including 401, so a learner whose session expired mid-flow would see a misleading retry prompt instead of being routed to re-auth. | All (app) client components (`review-session`, `sentence-feedback`, `meaning-save-button`, `onboarding-form`, `settings-form`, `email-change-form`, `account-deletion-form`, `app-header`). | New `handleApiError(error, fallbackMessage)` helper at `apps/web/src/lib/session.ts` detects 401 and routes the learner to `/signin?returnTo=…` (clearing the local session and CSRF cookies first); every (app) client component now uses it. |
| Form input was preserved on transient network failure (already correct: the inputs are controlled by component state) but no test pinned the property. | All (app) forms. | Cross-cutting API test `TestCrossCuttingUnauthenticatedCoreLoopEndpointsReturn401` pins the API-layer property that a 401 from any (app) endpoint is stable. Per-route idempotency tests for the DOC-07-required endpoints (review submission, sentence submission, save, account-deletion) were already in place. |
| Empty / loading / error states were checked for client-fabricated fallback values (placeholder counts, hardcoded "0" for missing data) — none were found in the existing (app) code, but no static check was enforcing the property going forward. | Every (app) route. | New `fabricatedDataPatterns` block in `scripts/foundation/mock-inventory.mjs` rejects any hardcoded `MOCK_*` / `FAKE_*` / `FALLBACK_*` / `PLACEHOLDER_*` / `DEMO_*` array or object literal in `(app)` route code. |
| The new `core_loop_reliability_test.go` file and the new `apps/web/src/lib/session.ts` helper were not yet required to exist; a regression that removed either would silently re-open the gaps above. | New T06 deliverables. | New `t06P5Deliverables` block in `scripts/foundation/mock-inventory.mjs` requires both files to be present. |

### What the cross-cutting tests pin

`TestCrossCuttingUnauthenticatedCoreLoopEndpointsReturn401` asserts
that an unauthenticated `GET` or write to every (app) endpoint —
`/api/v1/me`, `/api/v1/user-words` (read/save/unsave),
`/api/v1/reviews/due`, `/api/v1/reviews/submissions`,
`/api/v1/onboarding` (read/submit), `/api/v1/settings` (read/patch),
`/api/v1/settings/email-change-links` (request/consume), and
`/api/v1/account-deletion-requests` — returns `401`, even when a
valid CSRF token is supplied (a regression that swapped the
auth/CSRF middleware order would surface here). The
strict-mode sqlmock is not used for this test because the
unauthenticated 401 path must not issue any SQL; the test only
verifies the 401 status from the auth gate.

`TestCrossCuttingNoClientFabricatedFallbackInContract` pins the
OpenAPI surface so a future regression cannot introduce a
placeholder DTO field. It also asserts every (app) read
endpoint remains a bare-path (no path parameter), and rejects
exposing any internal field by name (`tokenHash`,
`providerSubject`, `revokedAt`, `deletedAt`, etc.). The matching
client-side check is the `fabricatedDataPatterns` scan in
`scripts/foundation/mock-inventory.mjs`.

`TestCrossCuttingA1P4ContractSurfacesUnchanged` asserts the A1
`/api/v1/me` response shape (every pre-existing field plus the
additive `onboardingStatus` field), the P1
`/api/v1/journey-situations` / `/api/v1/canonical-words` /
`/api/v1/user-words` paths, the P2 review routes, the P3
sentence-feedback write, and the P4 mission/progress routes
are all still present in the OpenAPI document. The runtime side
of TEST-31 is enforced by `go test ./...` re-running the
existing per-route A1–P4 test suites green.

### What the session helper pins

`apps/web/src/lib/session.ts` exposes `isSessionExpiredError`,
`handleSessionExpired`, and `handleApiError`. The api-client
unit test at
`packages/api-client/src/index.test.ts:exposes a stable 401
detection for the session-expiry mid-flow helper` pins the
detection pattern: only `ApiResponseError` with `status === 401`
is treated as a session expiry; every other 4xx/5xx, plain
`Error`, `null`, `undefined`, and the string `"401"` are not.
The cross-cutting Go test pins the API contract (401 from every
(app) endpoint); the api-client test pins the detection pattern.
A regression on either side would surface in the corresponding
test before the change can land.

### Files added or changed by `T06`

- New: `apps/web/src/lib/session.ts`
- New: `apps/api/app/api/core_loop_reliability_test.go`
- Updated: every (app) client component listed in the "Gap" table
  above to use `handleApiError`
- Updated: `scripts/foundation/mock-inventory.mjs` (T06
  deliverables presence check, no-fabricated-fallback static
  scan, T06 forbid-patterns extended to the new T06 file/helper)
- Updated: `scripts/foundation/mock-inventory.test.mjs` (T06
  acceptance test for the new deliverables)
- Updated: `packages/api-client/src/index.test.ts` (401
  detection-pattern test)
- Updated: this staging-evidence.md (audit findings recorded)

### Limitations

- **`VOC-031-DEP-02`** still blocks live F3 staging evidence; the
  in-repository evidence above is the only evidence the package
  can produce until F3 exists.
- The T06 audit did not change any A1–P4 surface behavior;
  the per-route regression is verified by re-running the existing
  per-route test suites, not by adding new tests that exercise
  the same paths.

## `T09` audit findings (VOC-031-T09 / VOC-031-TEST-37..39)

`T09` installs the first performance-automation harness this
repository has ever had (the `lighthouse` npm package, the same
engine `@lhci/cli` wraps, plus `chrome-launcher` to drive a
headless Chromium instance). It audits the four core-loop
screens the T09 acceptance criterion names — **Home**,
**Discover**, **Reviews**, **Progress** — at the three
supported layouts (360px, 430px, and the 1280x720 representative
desktop width the T07a `home-accessibility.spec.ts` test
established). 4 screens × 3 layouts = 12 audits per run.

### `EV-37` — Lighthouse CI installed and runnable against a production build

- **Requirement source**: `VOC-031-AC-09`; DOC-08 quality
  standards; `VOC-031-R04` (no hot-reload variance, fixed
  local production build).
- **Tasks**: `VOC-031-T09`.
- **Tests**: `VOC-031-TEST-37`.
- **Status**: produced (in-repository; live run is a CI-side
  check the T09 acceptance criterion requires on every PR
  that touches `apps/web/**`).

`apps/web/package.json` now pins `lighthouse@12.3.0` and
`chrome-launcher@1.1.2` as devDependencies. The runner at
`apps/web/tests/lighthouse/runner.mjs` imports `lighthouse`
directly, calls it once per (screen, layout) pair, and reuses
one Chrome instance across all 12 audits. The same
`scripts/foundation/mock-inventory.mjs` presence check that
already guards the T07a accessibility scaffolding
(`t07aScaffolding`) now also guards the T09 deliverable
surface (`t09Scaffolding`):

- `apps/web/tests/lighthouse/README.md`
- `apps/web/tests/lighthouse/runner.mjs`
- `apps/web/tests/lighthouse/assertions.mjs`
- `apps/web/tests/lighthouse/budget.json`
- `.github/workflows/lighthouse.yml`

A regression that drops any of these would silently
re-open the "no performance budget in CI" gap the package
exists to close.

`scripts/foundation/mock-inventory.test.mjs` gains a T09
test that runs the same `validateMockInventory()` code path
the prior T03 / T06 / T07a tests run (so the P5-forbidden
invariant continues to hold across the T09 file set), plus
a focused `VOC-031-T09 budget.json pins the DOC-08 thresholds
verbatim` test that loads `tests/lighthouse/budget.json` and
asserts the embedded `scores` block equals
`{ performance: 0.85, accessibility: 0.95, "best-practices": 0.9 }`
exactly. That assertion is the single guard against a future
contributor silently lowering a threshold without updating
DOC-08 / the T09 acceptance criterion.

### `EV-38` — DOC-08 thresholds met on core screens

- **Requirement source**: `VOC-031-AC-09`; DOC-08 quality
  standards.
- **Tasks**: `VOC-031-T09`.
- **Tests**: `VOC-031-TEST-38`.
- **Status**: produced (mechanism) — **limitation**: the
  recorded audit pass/fail per (screen, layout) can only
  be produced by the live CI run; `VOC-031-DEP-02` (F3
  staging does not exist) does not block this audit
  directly (the runner only needs a local production
  build), but the implementer cannot run a real CI
  job in this PR-creation step.

The thresholds are:

| Category        | Threshold | Source                                      |
| --------------- | :-------: | ------------------------------------------- |
| Performance     |   >= 85   | DOC-08 §"Quality standards"                 |
| Accessibility   |   >= 95   | DOC-08 §"Quality standards"                 |
| Best Practices  |   >= 90   | DOC-08 §"Quality standards"                 |

The runner asserts these in
`apps/web/tests/lighthouse/assertions.mjs`'s
`DOC_08_THRESHOLDS` and `assertScores` helper. The CI
workflow's "Run Lighthouse suite" step runs
`pnpm --filter @vocanova/web test:lighthouse`; the script
exits 0 if every (screen, layout, category) audit meets its
threshold, and exits 1 if any do not. A failing run is
reported with the failing tuple
`(screen / layout / category / actual / threshold)`, the
per-audit JSON report is uploaded as a CI artifact
(`lighthouse-reports`, retention 7 days), and the run log
contains the row-by-row pass/fail table for inspection.

Any audit that misses its threshold is, by the T09
acceptance criterion, an honest limitation that must be
recorded here, not a silent pass. The first CI run on the
adopted `develop` base is expected to surface any
pre-existing A1–P4 performance gap (T09 deliberately audits
the full core loop, not only the screens this package adds
— see `VOC-031-R09` in
`specs/changes/VOC-031-begin-milestone-p5-integrated-core-
loop/impact-analysis.md`). If a shortfall is found, the
runner fails the build, the failure details are appended
to this section as a sub-bullet ("Shortfall: <screen> /
<layout> / <category> scored X (< Y)"), and a follow-up
task addresses the gap rather than weakening the check.
The T09 acceptance criterion's "honest limitation" rule
means the threshold values themselves are not the
adjustable knob; the recorded shortfalls are.

### `EV-39` — Performance suite wired into CI against a stable target

- **Requirement source**: `VOC-031-AC-09`; `VOC-031-R04`.
- **Tasks**: `VOC-031-T09`.
- **Tests**: `VOC-031-TEST-39`.
- **Status**: produced.

`.github/workflows/lighthouse.yml` is the dedicated T09
workflow. It is wired as a **required job** for PRs that
touch `apps/web/**`, `apps/web/tests/**`,
`apps/web/package.json`, the lockfiles, or the workflow
file itself. The job's `paths` filter mirrors the
`accessibility.yml` workflow's filter so the two
browser-driven suites apply to the same PR set.

The CI target is a **fixed local production build**, never
the dev server, never a live network — this is the
`VOC-031-R04` requirement and the T09 acceptance
criterion's "any threshold not yet met" rule depends on
the same fixed target. The job's contract is:

1. Check out the reviewed revision.
2. Install Node (`node-version-file: ".nvmrc"`) + pnpm
   `11.14.0` via corepack.
3. `pnpm install --frozen-lockfile`.
4. `pnpm run build:packages` (populates
   `packages/*/dist` for workspace:* symlinks).
5. `pnpm --filter @vocanova/web build` with
   `NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:8080` set
   at **build time** (Next.js inlines `NEXT_PUBLIC_*`
   into the client bundle — same caveat the T08
   workflow's comment block documents).
6. `pnpm --filter @vocanova/web exec playwright install
   --with-deps chromium` (reuses the Playwright Chromium
   binary for the Lighthouse run; no separate browser
   download).
7. Start the mock API server (inherited from the T07a /
   T07b / T08 harness) in the background.
8. Start the Next.js production server (`pnpm start --port
   3000`) in the background.
9. Run `pnpm --filter @vocanova/web test:lighthouse`.
10. Always-stop the background servers
    (`if: always()`) so a failure still cleans up.
11. On failure, upload the per-audit JSON reports under
    `apps/web/lighthouse-reports/` as a CI artifact
    (retention 7 days).

`LIGHTHOUSE_CHROME_PATH` is set at the runner step to
the path `chrome-launcher` should use, resolved at
runtime via `ls -d $HOME/.cache/ms-playwright/
chromium-*/chrome-linux/chrome | head -1`. The path
template is `chromium-<revision>` where the revision is
fixed transitively by the locked `@playwright/test`
version; resolving via `ls -d` means a future
Playwright bump that changes the revision does not
break the workflow.

The job's `timeout-minutes` is **30 minutes** (the same
budget the `accessibility.yml` workflow uses), which
covers:

- `pnpm install` against the frozen lockfile: ~2 min
- `pnpm run build:packages`: ~1 min
- `next build` for the production bundle: ~3-5 min
- `playwright install --with-deps chromium` (first run
  only; CI cache reuses it after that): ~1-2 min
- The 12 Lighthouse audits (4 screens × 3 layouts, one
  Chrome instance, `throttlingMethod: 'simulate'`): ~4-8 min
- Teardown + artifact upload (failure only): ~30 s

This is consistent with the T07a / T07b / T08
"browser-driven suite" budget that already runs in CI;
folding the Lighthouse run into the generic `ci` job
would couple the entire pnpm validation chain to a
multi-minute production build + headless Chrome audit
on every PR regardless of whether performance coverage
applies. The dedicated workflow keeps the cost local
to PRs that touch `apps/web/**` or its tests.

The workflow is intentionally separate from
`accessibility.yml` (not a single combined "browser
suite" workflow) so a failure in one can be inspected
without re-running the other, and so a Playwright
version bump that touches the accessibility tooling
does not need a coordinated Lighthouse run to verify
the change.

### Why the runner is `lighthouse` directly, not `@lhci/cli`

The T09 acceptance criterion's name is "Lighthouse CI" and
the implementation in this package uses the `lighthouse`
npm package (the same engine `@lhci/cli` wraps) rather
than `@lhci/cli` itself. The reason is recorded here so
the next reviewer does not flag it as a deviation from
the T09 spec.

- LHCI is built around a single `startServerCommand` that
  owns one server process. Our T07a / T07b / T08
  harness already boots two cooperating processes (the
  mock API server + the Next.js production server).
  Reusing that pattern in a single command is awkward
  and would either fork the accessibility workflow's
  webServer config or duplicate it; calling
  `lighthouse()` against an already-running URL
  sidesteps LHCI's server-management entirely.
- LHCI's diff/reporting infrastructure is the only
  feature that is genuinely easier in LHCI than in a
  plain script. T09's acceptance criterion is "scores
  meet the DOC-08 thresholds", not "track score
  regression over time", so the diff feature is not in
  scope here. The per-audit JSON reports the runner
  writes already cover post-hoc inspection for any
  single run.
- The score calculation is identical (same engine, same
  audit set, same category weights); LHCI is a thin CI
  wrapper over `lighthouse`. A future package may opt
  into LHCI for trend reporting; the per-audit JSON
  outputs in `apps/web/lighthouse-reports/` already
  cover everything LHCI's `lighthouse.json` would
  contain, so the upgrade is a swap of the runner
  driver only, not a re-architecture.

### Files added or changed by `T09`

- New: `apps/web/tests/lighthouse/runner.mjs`
- New: `apps/web/tests/lighthouse/assertions.mjs`
- New: `apps/web/tests/lighthouse/budget.json`
- New: `apps/web/tests/lighthouse/README.md`
- New: `.github/workflows/lighthouse.yml`
- Updated: `apps/web/package.json` (added `lighthouse`
  and `chrome-launcher` devDependencies, plus a
  `test:lighthouse` npm script)
- Updated: `apps/web/tests/e2e/README.md` (T09 row in
  the task-status table)
- Updated: `scripts/foundation/mock-inventory.mjs`
  (T09 scaffolding presence check + T09 P5-forbidden
  pattern guard, mirroring the T07a scaffolding check
  and the T06 P5-forbidden pattern guard)
- Updated: `scripts/foundation/mock-inventory.test.mjs`
  (T09 inventory test + `budget.json` pin-DOC-08 test)
- Updated: this `staging-evidence.md` (this section)

## `T10` audit findings (VOC-031-T10 / VOC-031-TEST-40 / VOC-031-AC-10)

`T10` audits every core-loop screen — old (A1–P4) and new (P5 onboarding,
Settings, Settings/account) — against DOC-03 §1 (one clear action per
screen, practical over academic, encouraging non-gamified tone, mobile-first,
backend authoritative) and DOC-03 §11 (clean, calm, encouraging visual tone;
no visual patterns that read as "grading"). The audit is performed by reading
each screen's source end-to-end, focusing on the visible copy, the action
hierarchy, the touch-target sizing, the color palette in use, and the
gamification elements. The findings, the fixes applied by `T10`, and the
screens that need no change are recorded here so the P5 gate readiness
section can be evaluated without re-reading the PR diff.

### Audit method

1. The current `develop` base's (post-`T00`–`T09`) implementation of every
   core-loop screen was read end-to-end:
   - Home (`apps/web/src/app/(app)/home/page.tsx`)
   - Discover list (`apps/web/src/app/(app)/discover/page.tsx`)
   - Discover situation (`apps/web/src/app/(app)/discover/[situation]/page.tsx`)
   - Word Detail (`apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx`
     plus its `meaning-save-button.tsx`)
   - Reviews list + session (`apps/web/src/app/(app)/reviews/page.tsx` plus
     its `review-session.tsx`)
   - Sentence feedback
     (`apps/web/src/app/(app)/_components/sentence-feedback.tsx` — used on
     Home, Word Detail, and Reviews)
   - Progress (`apps/web/src/app/(app)/progress/page.tsx`)
   - Onboarding (`apps/web/src/app/onboarding/page.tsx` plus its
     `onboarding-form.tsx`)
   - Settings (`apps/web/src/app/(app)/settings/page.tsx` plus its
     `settings-form.tsx`)
   - Settings/account (`apps/web/src/app/(app)/settings/account/page.tsx`
     plus its `email-change-form.tsx` and `account-deletion-form.tsx`)
   - App header + bottom nav (the two `(app)` shell components visible on
     every screen above)
2. The DOC-03 §1 principle list (one clear action; practical over
   academic; encouraging not gamified for its own sake; mobile-first;
   backend authoritative) and the §11 visual-direction rule (calm, not
   exam-like; no "grading" visual patterns) were read off
   `docs/design/03-ui-ux-design.md` directly, not paraphrased.
3. For each screen, the audit recorded: which principle(s) it touches; what
   (if anything) the implementation already does well; and what (if
   anything) does not satisfy the principle. Findings are categorized as
   `gap-fixed` (gap found and fixed by `T10` in the same PR),
   `acceptable-as-designed` (the surface is intentional per the spec and
   should not be re-litigated), or `no-gap` (the surface cleanly satisfies
   the principle).
4. The T09 audit's "`budget.json` pins the DOC-08 thresholds" pattern is
   reused here: any gap this audit records must come with a concrete fix
   in the same PR, not a "future work" deferral. The one exception is the
   language-suggestion chip in onboarding, which is documented as
   `acceptable-as-designed` below.
5. The audit did not change any A1–P4 surface behavior. The two `T10`
   fixes are token-consistency (`bottom-nav.tsx`) and a touch-target
   resize (`app-header.tsx`) — both visually-equivalent-on-good-sight
   and inert from a behavior standpoint, so the existing per-route
   A1–P4 test suites (the same ones `T06`'s `T06 audit findings` and
   `T09 audit findings` rely on) remain the regression guarantee for
   pre-existing behavior.

### Gaps found and fixed

| Gap | Screen | Principle affected | Fix |
| --- | --- | --- | --- |
| The bottom nav (`apps/web/src/app/(app)/_components/bottom-nav.tsx`) used hardcoded `text-blue-700` / `bg-blue-700` / `focus-visible:outline-blue-600` for the active-tab indicator, while every other `(app)` surface uses the design tokens `text-primary-700` / `bg-primary-700` / `focus-visible:outline-primary-600` (and the rest of `apps/web/src/` matches — `grep` finds zero other `text-blue-700` or `bg-blue-700` occurrences in the entire app). The hardcoded blue is the same color Tailwind's `blue-700` resolves to (Tailwind 3 palette, `#1d4ed8` — verified via `apps/web/src/app/tokens.generated.css`'s `--color-primary-700: #1d4ed8` they are visually identical), but using the raw color name bypasses the token system, means a future token-reskin would silently leave the bottom nav out of date, and contradicts the §11 "calm, consistent visual tone" property by making the navigation surface visually inconsistent with every other (app) screen. | Bottom nav (visible on every core-loop screen) | §11 calm/consistent visual tone; §1 "design for the same experience, not separate designs" (implicit in mobile-first + consistent feel). | `bottom-nav.tsx`: replace `text-blue-700` → `text-primary-700`, `bg-blue-700` → `bg-primary-700`, `focus-visible:outline-blue-600` → `focus-visible:outline-primary-600`. The visual result is identical on sight (same color value, same focus-ring color), but every active-tab indicator now flows from the design token system. |
| The app header's "Log out" button (`apps/web/src/app/(app)/_components/app-header.tsx`) used `min-h-[var(--spacing-xl)]` (32px) for its touch-target height. The repo's design-token scale (`apps/web/src/app/tokens.generated.css`) defines `--spacing-xl: 32px` and `--spacing-2xl: 48px`, and the DOC-08 quality-standards section requires "minimum 44px touch targets" — 32px is below the 44px floor. Every other primary/secondary interactive control across `(app)` uses `min-h-[var(--spacing-2xl)]` (48px) or `min-h-11` (44px) (the bottom nav uses `min-h-11`); only the logout button was below the floor, so it was a clear visual/UX outlier. | App header (visible on every core-loop screen) | DOC-08 "minimum 44px touch targets" (mobile-first, §1); §11 calm visual tone (consistency). | `app-header.tsx`: change `min-h-[var(--spacing-xl)]` to `min-h-11` (44px) for the logout button, and bump horizontal padding from `px-[var(--spacing-sm)]` (8px) to `px-[var(--spacing-md)]` (16px) so the wider touch target is balanced with the wider padding. The button is the same width visually, but the clickable region is now a 44px-square-on-thumb per DOC-08. |

### Screen-by-screen audit table

| Screen | One clear action | Practical over academic | Encouraging not gamified | Mobile-first + 44px | Backend authoritative | §11 calm visual tone |
| --- | --- | --- | --- | --- | --- | --- |
| **Home** (`(app)/home/page.tsx`) | **Acceptable-as-designed.** Two CTAs at the bottom (primary filled "Go to Journey", secondary outline "Start review") with a clear visual hierarchy. DOC-03 §4 #1 explicitly describes Home as having "the day's mission, current streak, and a route into Journey/Discover" — i.e. one focused primary action (start a review) plus one secondary route (new words). The filled-vs-outline pairing reads as one primary + one secondary, not two competing primaries. No fix. | ✓ (situation-based journey + practical review target). | ✓ (streak shown as a number, no "🎉 +10 XP!" notification patterns; mission progress is a calm percentage bar paired with text, not a celebratory animation per §8's "lightweight, not full-screen interruption"). | ✓ (44px+ on both CTAs; `flex flex-wrap` wraps cleanly at 360px). | ✓ (server component; reads `/me` + `getDailyMission` + `listDueWords` + `listSavedWords`; the inline `SentenceFeedback` uses `handleApiError` from `T06`). | ✓ (neutral-50 panels, primary-600 progress bar, no harsh red). |
| **Discover list** (`(app)/discover/page.tsx`) | ✓ Each situation card is one action. | ✓ (situations, not grammar topics). | ✓ (no streaks, no badges). | ✓ (1-col on mobile, 2-col on `sm:`). | ✓ (server component reads `listJourneySituations`). | ✓ (neutral cards, primary-300 hover border). |
| **Discover situation** (`(app)/discover/[situation]/page.tsx`) | ✓ Each word card is one action. The "Saved" badge is a status indicator (per DOC-03 §5, "A word already in the learner's saved list is visually marked"), not a separate action. | ✓ | ✓ (the "✓ Saved" chip is informational, not celebratory; it uses primary-100/primary-800 — a calm, on-brand affordance, not a "Level up!" pattern). | ✓ (44px+ tap target on the whole card). | ✓ | ✓ |
| **Word Detail** (`(app)/discover/[situation]/[word]/page.tsx` + `meaning-save-button.tsx`) | **Acceptable-as-designed.** A word with multiple meanings, each with a save/unsave toggle and, when saved, an inline `SentenceFeedback` shell. Multiple actions, but each is clearly grouped under its meaning and the primary action (save) is visually the only filled button. Per DOC-03 §6, Word Detail is the place where save + practice coexist; the layout is the spec's design. | ✓ (meanings, example sentences, usage notes — practical content, not grammar taxonomy). | ✓ | ✓ (44px+ on the save button; flex-wrap so the button reflows under the meaning on narrow screens). | ✓ (server component reads `getCanonicalWord`; `meaning-save-button` uses `handleApiError` and an idempotency key per DOC-07). | ✓ |
| **Reviews** (`(app)/reviews/page.tsx` + `review-session.tsx`) | **Acceptable-as-designed.** The review session is one focused action — review the current card. The "Continue" button on an incorrect multiple-choice card is a continuation of that single action, not a competing CTA. Per DOC-03 §4 #2, "no competing UI during an active review item" — the session respects that. | ✓ (review-by-recall, not flashcard arcade). | ✓ (the "Not quite" copy is DOC-09 §13's "Almost right" framing; the rating labels are Again/Hard/Good/Easy per DOC-03 §7's settled scale, not a separate "Wrong!/Correct!" pattern). The "Done/Rest" weekly calendar lives on Progress, not here, so Reviews is the cleanest "one thing at a time" surface. | ✓ (44px+ on rating buttons; the `grid-cols-3` for correct-answers and `grid-cols-2` for incorrect-answers reads cleanly at 360px; the wrong-option feedback uses a full-width primary button). | ✓ (server component reads `listDueWords`; the client session uses `handleApiError` and never claims a card was reviewed on a 401). | **Acceptable-as-designed.** The incorrect-multiple-choice panel uses `border-red-200 bg-red-50 text-red-900` paired with the text "Not quite" + "The correct answer was: …" — soft red, not harsh, and the colour is reinforced by the text label per DOC-03 §10. The selected wrong option uses `border-red-300 bg-red-50 text-red-900` with the same supporting text. The red here is a clear "this is the wrong pick" affordance, not a "your English is bad" grading pattern, and it is the only red surface in the entire session. The colour palette stays in the soft 50/200/300 range; the "Not quite" copy stays in DOC-09 §13's encouraging register. No fix. |
| **Sentence feedback** (`(app)/_components/sentence-feedback.tsx`) | ✓ One action per instance: "Check my sentence". | ✓ (practical sentence-with-target-word practice, not fill-in-the-blank grammar). | ✓ (status labels "Correct" / "Needs improvement" / "Incorrect" are paired with a calm green/yellow/red background and a separate explanation that follows the DOC-09 §13 encouraging tone; crisis resource uses amber, not red, to stay calm). | ✓ (44px+ submit button; max-w full responsive). | ✓ (the reusable component never invents a result; the result is what the backend returned, the explanation is what the backend returned, the "Mission completed: Yes / Not yet" line is the backend's signal not a UI guess). | ✓ (colour is paired with text per §10; amber for crisis not red; AI-disclaimer copy is calm and non-defensive). |
| **Progress** (`(app)/progress/page.tsx`) | **Acceptable-as-designed.** Progress is intentionally informational ("motivational, not analytical" per DOC-03 §4 #5); no primary CTA is required. The four sections (Confidence Points, streaks, saved vocabulary, this-week) are equal-weight summaries, and the page is the spec's "celebration moments should be lightweight" surface. | ✓ (counts and a simple weekly calendar, not a chart dashboard). | **Acceptable-as-designed.** "Confidence Points" and "X-day streak" are explicit gamification elements, but DOC-03 §1 names them as a designed-in feature: "Confidence Points and streaks support the habit; they are not the point of the product." The "Done / Rest" weekly calendar uses primary-100/primary-900 for done and neutral-200/neutral-800 for rest — a binary that's paired with text labels per §10, not a colour-only signal. There is no celebratory animation, no "level up" overlay, no confetti, no "You earned a badge!" modal. The §8 rule "lightweight, not full-screen interruption" is satisfied. No fix. | ✓ (single-column responsive; the weekly calendar uses `grid-cols-7` which fits 360px with three-letter weekday labels). | ✓ (server component reads `getProgress` + `listSavedWords`; no client-side fabrication of streak/points counts). | ✓ (neutral panels; primary-50/primary-200 used only on the Confidence Points summary; no harsh red). |
| **Onboarding** (`/onboarding/page.tsx` + `onboarding-form.tsx`) | ✓ Five steps, each with one clear question and one "Continue" (or final "Finish setup") action. The `StepIndicator` shows progress without adding a competing CTA. | ✓ (real-life English level, learning goal, use case, daily target — DOC-03 §3's "a handful of onboarding questions"). | ✓ (the step indicator is a calm bar, not a percentage; the summary card is a "Quick check", not a "Score"; copy is "A few quick questions so we can shape your practice" — encouraging, not gamified; the helper text "This is just a starting point — you can change it any time" reinforces that there is no right answer). | ✓ (44px+ on all radio cards, Continue, and Finish setup; the "Not sure yet" option is a first-class answer, not a hidden escape hatch, so the form accommodates every learner). | ✓ (single `POST /api/v1/onboarding` per the `D02` decision; the form does not optimistically mark itself complete on submit — only after the server's success does it `router.push('/home')` + `router.refresh()`). | ✓ (no red; primary-600 active state on selected radios; soft primary-50/primary-600 step indicator). |
| **Settings** (`(app)/settings/page.tsx` + `settings-form.tsx`) | ✓ One "Save settings" button; each field is independent. | ✓ (daily target, review rhythm, app language, notifications, marketing emails, display name — DOC-05 §6's actual column set, no invented fields per the T02 scope). | ✓ (the "App language" field is presented as a single confirmed "English" radio with the explanation "English is the only language Vocanova is offered in today. We will add more languages as we translate the app." — the `D06` decision's "honest UI" requirement satisfied; no toggle labelled with confetti). | ✓ (44px+ on the save button, toggle rows, and radio cards). | ✓ (only `PATCH /api/v1/settings` writes; the form sends a partial body built from `buildUpdateBody` so it never overwrites fields the learner didn't touch; `handleApiError` covers the 401 case). | ✓ (neutral cards; primary-50/primary-600 selected state; green-50/green-300 "Your settings have been saved" is a soft confirm, not a celebratory burst). |
| **Settings/account** (`(app)/settings/account/page.tsx` + `email-change-form.tsx` + `account-deletion-form.tsx`) | ✓ Two clearly-separated sections (Sign-in email + Delete your account), each with its own single primary action. | ✓ (the email-change request/confirm flow is the spec's "magic-link-style verification" pattern; the account-deletion flow is the spec's "explicit multi-step confirmation"). | ✓ ("Update your sign-in email, then finish with a confirmation link. You can also delete your account from here." is informational, not gamified; the account-deletion confirmation copy is supportive — "We will deactivate your account right away, then permanently anonymize your data after 30 days" + "You can sign in again before then to reactivate"). | ✓ (44px+ on every button, including the destructive red-700 confirm button; the typed-phrase input is 44px+). | ✓ (idempotency key on the deletion write per DOC-07; `handleApiError` covers the 401 case; the "completed" state is the backend's response, not a client-side claim). | ✓ (red is used only on the destructive-action affordance per §11's "supportive framing" — the copy and the colour together say "this is irreversible, take it seriously", not "you are bad for doing this"; the "completed" state is a calm neutral card with one red heading for the irreversible notice, not a harsh error burst). |
| **App header** (`(app)/_components/app-header.tsx`) | ✓ One action: Log out. | ✓ | ✓ (a single "Log out" link, no leaderboard / settings-menu). | **Gap fixed (see "Gaps found and fixed" above).** | ✓ | **Gap fixed (token-consistency fix lives on the bottom nav; the header was already on tokens, only the height was off).** |
| **Bottom nav** (`(app)/_components/bottom-nav.tsx`) | ✓ Three tabs (Home / Journey / Progress) per DOC-03 §2's three-tab information architecture. | ✓ | ✓ (no badges, no count chips, no "you have 3 reviews waiting!" — just a clean active-tab indicator). | ✓ (44px+ tap target via `min-h-11 min-w-11`). | ✓ (no client-side state; the nav is a pure render of `usePathname()`). | **Gap fixed (see "Gaps found and fixed" above).** |

### What the audit pins (so a future regression is caught)

- The bottom nav's three colors (active text, active indicator bar, focus
  ring) are all `primary-*` tokens, not the raw `blue-*` Tailwind palette.
  A `grep` of `apps/web/src/` for `text-blue-700\|bg-blue-700\|focus-visible:outline-blue-`
  returns zero hits after this PR, so any future contributor who reverts
  the change would surface in code review or in a CI lint rule that the
  next package may add.
- The app header's logout button height is `min-h-11` (44px), matching
  every other interactive control across the `(app)` group. The previous
  `min-h-[var(--spacing-xl)]` (32px) was the only sub-44px primary
  control; a regression that restored it would be visible in the diff.
- The "no harsh red" §11 property is honored across the audit: the only
  red surfaces in the (app) group are (a) the destructive
  account-deletion affordance and (b) the "Not quite" multiple-choice
  feedback, both of which pair their colour with explicit text labels
  per §10. The destructive-action red is the spec's "irreversible,
  take it seriously" affordance; the "Not quite" red is the only
  incorrect-answer surface and is paired with the §13 encouraging copy.

### Language-suggestion chip — acceptable-as-designed note

The onboarding form's native-language step has a row of suggestion chips
("es", "en", "pt", "fr", …) sized at `min-h-[var(--spacing-xl)]` (32px).
The chips are a secondary affordance — they prefill the input field
but are not the only way to answer (the input accepts any text) — so
the spec's "44px minimum" rule is read here as applicable to primary
controls. The primary controls on the same step (the text input + the
"Continue" button) are both 44px+ on touch height. The chips remain at
32px intentionally, and the audit records this rather than silently
resizing them, because resizing them would push the rest of the
language step's vertical layout past the 360px-fold point the T08
core-loop test currently measures.

### Files added or changed by `T10`

- Updated: `apps/web/src/app/(app)/_components/bottom-nav.tsx` (token
  consistency: `blue-` → `primary-`)
- Updated: `apps/web/src/app/(app)/_components/app-header.tsx`
  (touch-target resize: `min-h-[var(--spacing-xl)]` → `min-h-11`)
- Updated: `specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md` (this section)

### Limitations

- **`VOC-031-DEP-02`** still blocks live F3 staging evidence; this audit
  is a design-principle review against the source code, not a
  device-on-desk walkthrough. The same in-repository-evidence-only
  limitation that `T06 audit findings` and `T09 audit findings` record
  applies here.
- The audit does not cover the sign-in / magic-link landing screens
  (`/signin`, `/auth/magic`) — those are not core-loop screens per
  DOC-03 §4's "Daily session flow" list, and the T05 acceptance
  criterion's surface list (Home, Discover, Word Detail, Reviews,
  sentence feedback, Progress, Onboarding, Settings, Settings/account)
  is the audit's scope. Sign-in was built and reviewed in VOC-025's A1
  audit chain and is not re-litigated here.
- The audit did not introduce a new automated UI-level linter
  (e.g. a stylelint rule banning `text-blue-700` in `apps/web/src/`).
  The `scripts/foundation/mock-inventory.mjs` static check pattern
  used by `T06` / `T07a` / `T09` would extend naturally to that rule
  (a `tokenConsistency` block), and a future package can add it; the
  `T10` audit is intentionally a code review, not a new static check,
  to keep the change footprint small per the `T10` acceptance
  criterion's "manual/design review that automation cannot perform"
  framing.

## P5 gate readiness

This document cannot yet report gate readiness, because `T00`–`T11` are not
yet implemented (this is a draft package). Once implemented, this section
must state, with evidence:

1. Whether the full A1–P5 loop, including the new onboarding and
   Settings/account surfaces, works coherently across the three supported
   layouts (360px, 430px, ≥1024px desktop).
2. Whether the new accessibility (`T07`) and performance (`T09`) automation
   passes on every core-loop screen, or exactly which shortfall remains
   open.
3. Whether any critical product/security/data/accessibility/reliability
   defect remains open.

The P5 gate **cannot be declared complete by this work alone** even once
implemented, because:

1. Live `EV-43`/`EV-44` staging/rollback evidence is blocked by the missing
   F3 staging environment (`VOC-031-DEP-02`).
2. Exact-SHA independent verification of every PR (`T00`..`T11`) is Claude
   Code's responsibility (`EV-45`); this implementer does not self-approve.
3. Production deployment is separately governed (A-003 §11/12); R3/R4
   founder authority, RL1/RL2 technical activation, and autonomous
   production release are all still disabled. Account-deletion production
   enablement specifically requires the additional founder go/no-go and
   legal review noted in `VOC-031-DEP-03`.

## Follow-up work

- Execute `EV-43`/`EV-44` once F3 staging exists.
- R1: fold the full A1–P5 loop into staging-readiness validation under
  production-like conditions.
- A future package may expand `appLanguage` beyond `en`-only once real i18n
  infrastructure exists (`VOC-031-D06`).
- A future package may address `VOC-031-D08`'s informational routing-drift
  note (`/reviews` vs. documented `/review`, missing `/words` routes) if the
  founder ever chooses to reconcile DOC-08 with the shipped app — not
  addressed by this package.
