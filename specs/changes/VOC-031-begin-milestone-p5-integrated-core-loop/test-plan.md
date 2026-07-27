# VOC-031 — Test Plan

No test, fixture, seed file, OpenAPI example, CI config, or evidence may
contain a real secret, production URL/data, another learner's personal
content, or a raw session/CSRF token. Discover installed commands at the
adopted base; a missing integration, staging credential, open decision
(`VOC-031-DEP-02` through `DEP-04`), or browser tool is never reported as a
pass — it is a recorded limitation or blocker.

## VOC-031-TEST-00 — `user_onboarding_profiles` migration invariants
- Covers: `VOC-031-AC-00`; Preconditions: T00, disposable PostgreSQL.
- Procedure: apply the T00 migration and assert the FK, uniqueness (one row
  per user), and the `daily_review_target` 5–100 check constraint.
- Expected result: migration and constraints pass; production startup does
  not migrate. Evidence: `VOC-031-EV-00`.

## VOC-031-TEST-01 — Migration compatibility and no A1/P1/P2/P3/P4 regression
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: run the adopted migration validation and disposable
  forward/recovery rehearsal against an existing VOC-030 DB; assert every
  prior-milestone table/column/constraint is byte-for-byte unchanged.
- Expected result: recoverable migration; no prior-milestone schema
  regression. Evidence: `VOC-031-EV-01`.

## VOC-031-TEST-02 — Onboarding-completion seeds `daily_review_target` when unset
- Covers: `VOC-031-AC-01`; Preconditions: T00, `D07` resolved.
- Procedure: complete onboarding for a user with no prior `user_settings`
  row (schema-default `daily_review_target=20`) and assert the row is
  updated to the onboarding answer.
- Expected result: seeding applies when no customization exists. Evidence:
  `VOC-031-EV-02`.

## VOC-031-TEST-03 — Seeding never overwrites an existing customization
- Covers: `VOC-031-AC-01`, `VOC-031-R05`; Preconditions: T00, `D07`
  resolved.
- Procedure: set `user_settings.daily_review_target` to a non-default value,
  then complete onboarding with a different answer; assert the stored value
  is unchanged.
- Expected result: no overwrite of a customized value. Evidence:
  `VOC-031-EV-03`.

## VOC-031-TEST-04 — Timezone resolution unchanged by onboarding
- Covers: `VOC-031-AC-01`; Preconditions: T00.
- Procedure: complete onboarding (no timezone question exists per DOC-03
  §3) and assert `user_settings.timezone` resolution still follows the
  unchanged `VOC-030-D01` chain (stored value → client-supplied IANA
  timezone → UTC).
- Expected result: no timezone regression. Evidence: `VOC-031-EV-04`.

## VOC-031-TEST-05 — `onboarding` module is transaction-scoped
- Covers: `VOC-031-AC-01`, DOC-06 §3; Preconditions: T00.
- Procedure: assert the exported `onboarding` functions accept an existing
  `*sql.Tx` parameter and never call `Begin()`/open a new connection
  transaction themselves.
- Expected result: no cross-module transaction violation. Evidence:
  `VOC-031-EV-05`.

## VOC-031-TEST-06 — Onboarding submission upserts and completes
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: submit all five DOC-03 §3 answers and assert
  `user_onboarding_profiles` is upserted, `users.onboarding_status` becomes
  `completed`, and the `AC-01` seeding runs, all in one transaction.
- Expected result: single, correct completion transaction. Evidence:
  `VOC-031-EV-06`.

## VOC-031-TEST-07 — Re-submission (revisiting onboarding) does not duplicate
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: submit onboarding twice with different answers; assert the
  second call updates the same row (no duplicate row, unique constraint
  holds) and `AC-01`'s "never overwrite" rule is respected for
  `user_settings`.
- Expected result: idempotent upsert semantics. Evidence: `VOC-031-EV-07`.

## VOC-031-TEST-08 — Invalid enum value rejected
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: submit an out-of-enum `english_level`/`learning_goal`/
  `main_use_case` value and assert a 400 with no partial write.
- Expected result: validation rejects before any write. Evidence:
  `VOC-031-EV-08`.

## VOC-031-TEST-09 — Redirect rules: incomplete-onboarding and already-complete
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: render an `(app)` route for an onboarding-incomplete user
  (assert redirect to `/onboarding`) and render `/onboarding` for an
  onboarding-complete user (assert redirect to `/home`).
- Expected result: both redirect rules hold. Evidence: `VOC-031-EV-09`.

## VOC-031-TEST-10 — Authentication and self-scoping on onboarding endpoints
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: call both endpoints unauthenticated (401); confirm neither
  accepts a caller-supplied user/ID parameter.
- Expected result: 401 unauthenticated; no cross-user parameter exists.
  Evidence: `VOC-031-EV-10`.

## VOC-031-TEST-11 — Contract and OpenAPI drift (onboarding)
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: regenerate OpenAPI, run drift/golden checks, verify the
  matched client compiles, and that no Ent/internal type leaks through
  either DTO.
- Expected result: OpenAPI/client agree; no internal exposure. Evidence:
  `VOC-031-EV-11`.

## VOC-031-TEST-12 — `GET`/`PATCH /api/v1/settings` contract and boundary (`D06`)
- Covers: `VOC-031-AC-03`; Preconditions: T02, `D06` resolved.
- Procedure: read and update `dailyReviewTarget`/`reviewIntervalPreset`/
  `appLanguage`/`notificationsEnabled`/`marketingEmailsEnabled`; assert
  `timezone` is not present as a writable field.
- Expected result: boundary matches the adopted `D06` resolution. Evidence:
  `VOC-031-EV-12`.

## VOC-031-TEST-13 — `GET`/`PATCH /api/v1/account` contract and boundary (`D06`)
- Covers: `VOC-031-AC-03`; Preconditions: T02, `D06` resolved.
- Procedure: read and update `displayName`; assert no email-address or
  account-deletion field/endpoint exists.
- Expected result: boundary matches the adopted `D06` resolution. Evidence:
  `VOC-031-EV-13`.

## VOC-031-TEST-14 — CSRF required on both write endpoints
- Covers: `VOC-031-AC-03`, `VOC-031-R01`; Preconditions: T02.
- Procedure: call `PATCH /api/v1/settings` and `PATCH /api/v1/account`
  without a valid CSRF token and assert 403.
- Expected result: CSRF enforced identically to existing P1/P2 writes.
  Evidence: `VOC-031-EV-14`.

## VOC-031-TEST-15 — Idempotency-Key required and duplicate-safe
- Covers: `VOC-031-AC-03`, `VOC-031-R01`; Preconditions: T02.
- Procedure: call either write endpoint without an `Idempotency-Key` (assert
  a required-header error); replay the same key twice and assert no
  duplicate side effect.
- Expected result: idempotency enforced identically to existing P1/P2
  writes. Evidence: `VOC-031-EV-15`.

## VOC-031-TEST-16 — Requester scoping and 401/400
- Covers: `VOC-031-AC-03`; Preconditions: T02.
- Procedure: call both endpoints unauthenticated (401); submit an
  out-of-range `dailyReviewTarget` (400); confirm neither endpoint accepts a
  caller-supplied user/ID parameter.
- Expected result: 401/400 as appropriate; no cross-user parameter exists.
  Evidence: `VOC-031-EV-16`.

## VOC-031-TEST-17 — Pre-existing internal settings-resolution chain unchanged
- Covers: `VOC-031-AC-03`, `VOC-031-R01`; Preconditions: T02.
- Procedure: re-run the VOC-030 `gamification` timezone/target-resolution
  test suite against the T02 code and assert no regression.
- Expected result: byte-for-byte unchanged internal behavior. Evidence:
  `VOC-031-EV-17`.

## VOC-031-TEST-18 — Contract and OpenAPI drift (settings, account)
- Covers: `VOC-031-AC-03`; Preconditions: T02.
- Procedure: regenerate OpenAPI, run drift/golden checks, verify the
  matched client compiles, and that no Ent/internal type leaks through
  either DTO.
- Expected result: OpenAPI/client agree; no internal exposure. Evidence:
  `VOC-031-EV-18`.

## VOC-031-TEST-19 — Cross-user access denied on both endpoints
- Covers: `VOC-031-AC-03`; Preconditions: T02.
- Procedure: attempt to read/update another learner's settings/account
  state (no ID parameter exists, so this asserts the implicit self-scoping
  cannot be bypassed via header/body manipulation).
- Expected result: no cross-user read or write is possible. Evidence:
  `VOC-031-EV-19`.

## VOC-031-TEST-20 — Settings screen wiring
- Covers: `VOC-031-AC-04`; Preconditions: T03.
- Procedure: render `/settings`, change a value, save, and assert the
  persisted value round-trips through a fresh `GET`.
- Expected result: real read/write wiring, no mock. Evidence:
  `VOC-031-EV-20`.

## VOC-031-TEST-21 — Account screen wiring
- Covers: `VOC-031-AC-04`; Preconditions: T03.
- Procedure: render `/settings/account`, change `displayName`, save, and
  assert the persisted value round-trips through a fresh `GET`.
- Expected result: real read/write wiring, no mock. Evidence:
  `VOC-031-EV-21`.

## VOC-031-TEST-22 — Failed save preserves the attempted value
- Covers: `VOC-031-AC-04`; Preconditions: T03.
- Procedure: simulate a backend error on save and assert the form still
  shows the learner's attempted value with a visible error, not a silently
  reverted or silently accepted value.
- Expected result: honest failure state. Evidence: `VOC-031-EV-22`.

## VOC-031-TEST-23 — Settings/account screens use the shared component set
- Covers: `VOC-031-AC-04`, `VOC-031-AC-05`; Preconditions: T03, T04.
- Procedure: assert both new screens' loading/empty/error states are
  rendered by the same shared components `T04` establishes, not a
  screen-local variant.
- Expected result: no component duplication. Evidence: `VOC-031-EV-23`.

## VOC-031-TEST-24 — Shared loading/empty/error components adopted everywhere
- Covers: `VOC-031-AC-05`; Preconditions: T04.
- Procedure: inspect every `(app)`/`(onboarding)` route's loading/empty/
  error rendering path and assert each uses the shared component set, not
  an ad hoc per-route implementation.
- Expected result: one component set, no duplication. Evidence:
  `VOC-031-EV-24`.

## VOC-031-TEST-25 — DOC-03 §4 core-loop transitions are reachable
- Covers: `VOC-031-AC-05`; Preconditions: T04.
- Procedure: walk Home → Discover → Review → Review completion → Sentence
  practice → Progress → Settings and back via client-side navigation only;
  assert no dead end and no unexpected full-page reload.
- Expected result: every documented transition works. Evidence:
  `VOC-031-EV-25`.

## VOC-031-TEST-26 — Home/Review due-count consistency
- Covers: `VOC-031-AC-05`; Preconditions: T04.
- Procedure: load Home and the Review screen for the same requester in the
  same session and assert the due-review count agrees between both reads.
- Expected result: no Home-vs-Review disagreement. Evidence:
  `VOC-031-EV-26`.

## VOC-031-TEST-27 — No P1–P4 backend behavior regression from T04
- Covers: `VOC-031-AC-05`, `VOC-031-R02`; Preconditions: T04.
- Procedure: re-run the VOC-026/027/028/030 frontend integration test
  suites against the T04 code and assert no regression in the underlying
  data/behavior each screen presents.
- Expected result: byte-for-byte unchanged underlying behavior. Evidence:
  `VOC-031-EV-27`.

## VOC-031-TEST-28 — Navigation works at all three supported layouts
- Covers: `VOC-031-AC-05`, `VOC-031-D05`; Preconditions: T04.
- Procedure: repeat `TEST-25` at 360px, 430px, and ≥1024px.
- Expected result: navigation is coherent at every supported layout.
  Evidence: `VOC-031-EV-28`.

## VOC-031-TEST-29 — Retried write after a lost response causes no duplicate
- Covers: `VOC-031-AC-06`, `VOC-031-R03`; Preconditions: T05.
- Procedure: simulate a write that succeeds server-side but whose response
  is lost client-side, then retry; assert no duplicate word-save, review,
  sentence submission, or settings/account update.
- Expected result: exactly-once effect despite a lost response. Evidence:
  `VOC-031-EV-29`.

## VOC-031-TEST-30 — Genuinely failed write remains retryable
- Covers: `VOC-031-AC-06`; Preconditions: T05.
- Procedure: simulate a genuine failure (network error, 5xx) on each of
  P1/P2/P3/`T02`'s writes and assert the UI offers a working retry path
  with no stuck state.
- Expected result: every failure mode is recoverable. Evidence:
  `VOC-031-EV-30`.

## VOC-031-TEST-31 — Session expiry mid-flow redirects and resumes
- Covers: `VOC-031-AC-06`; Preconditions: T05.
- Procedure: expire a session mid-review and mid-sentence-practice; assert
  redirect to sign-in and, on successful re-authentication, return to the
  prior route where reasonably possible.
- Expected result: no lost place in the loop beyond what re-authentication
  itself requires. Evidence: `VOC-031-EV-31`.

## VOC-031-TEST-32 — Slow/failed AI feedback does not block the loop
- Covers: `VOC-031-AC-06`, DOC-12 §5 P3; Preconditions: T05.
- Procedure: simulate a slow or failed AI-feedback call and assert Home/
  Review/mission completion remain fully usable.
- Expected result: AI-optionality gate language holds under P5's
  reliability work too. Evidence: `VOC-031-EV-32`.

## VOC-031-TEST-33 — Idempotency-Key regeneration/reuse policy is deterministic
- Covers: `VOC-031-AC-06`, `VOC-031-R03`; Preconditions: T05.
- Procedure: assert the client's retry logic derives its `Idempotency-Key`
  deterministically per logical action (never generating a fresh key for
  the same logical retry, never reusing a key across two different logical
  actions).
- Expected result: no key-reuse or key-churn bug. Evidence: `VOC-031-EV-33`.

## VOC-031-TEST-34 — No P1–P4 backend behavior regression from T05
- Covers: `VOC-031-AC-06`, `VOC-031-R02`; Preconditions: T05.
- Procedure: re-run the VOC-026/027/028/030 backend test suites against the
  T05 code and assert no regression.
- Expected result: byte-for-byte unchanged backend behavior. Evidence:
  `VOC-031-EV-34`.

## VOC-031-TEST-35 — Accessibility tooling installed and CI-wired
- Covers: `VOC-031-AC-07`; Preconditions: T06.
- Procedure: confirm `playwright` and `@axe-core/playwright` are installed
  dependencies and a CI job runs the sweep on every pull request.
- Expected result: tooling present and wired, not aspirational. Evidence:
  `VOC-031-EV-35`.

## VOC-031-TEST-36 — Zero unresolved axe violations across all routes
- Covers: `VOC-031-AC-07`; Preconditions: T06.
- Procedure: run the sweep over every `(public)`/`(onboarding)`/`(app)`
  route at 360px, 430px, and ≥1024px and assert zero unresolved violations.
- Expected result: WCAG 2.2 AA target met per the automated coverage.
  Evidence: `VOC-031-EV-36`.

## VOC-031-TEST-37 — New P5 screens included in the sweep
- Covers: `VOC-031-AC-07`; Preconditions: T06, T01, T03.
- Procedure: confirm `/onboarding`, `/settings`, and `/settings/account`
  are in the sweep's route list and pass with zero violations.
- Expected result: no new screen ships without accessibility coverage.
  Evidence: `VOC-031-EV-37`.

## VOC-031-TEST-38 — CI job is blocking, not advisory
- Covers: `VOC-031-AC-07`, `VOC-031-R06`; Preconditions: T06.
- Procedure: introduce a deliberate violation on a disposable branch and
  confirm the CI job fails the build (not merely reports a warning).
- Expected result: the accessibility gate actually blocks. Evidence:
  `VOC-031-EV-38`.

## VOC-031-TEST-39 — Lighthouse CI installed and CI-wired against a production build
- Covers: `VOC-031-AC-08`; Preconditions: T07.
- Procedure: confirm `@lhci/cli` is installed, runs against a production
  build (not `next dev`), and a CI job executes it on every pull request.
- Expected result: tooling present and correctly configured. Evidence:
  `VOC-031-EV-39`.

## VOC-031-TEST-40 — All six routes meet the DOC-08 thresholds
- Covers: `VOC-031-AC-08`; Preconditions: T07.
- Procedure: run Lighthouse CI against `/home`, `/discover`, `/reviews`,
  `/progress`, `/settings`, `/onboarding` and assert Performance ≥85,
  Accessibility ≥95, Best Practices ≥90 on each.
- Expected result: every route passes every threshold. Evidence:
  `VOC-031-EV-40`.

## VOC-031-TEST-41 — CI job is blocking, not advisory
- Covers: `VOC-031-AC-08`, `VOC-031-R06`; Preconditions: T07.
- Procedure: introduce a deliberate performance regression (e.g. an
  oversized synchronous import) on a disposable branch and confirm the CI
  job fails the build.
- Expected result: the performance gate actually blocks. Evidence:
  `VOC-031-EV-41`.

## VOC-031-TEST-42 — Design-token and touch-target consistency
- Covers: `VOC-031-AC-09`; Preconditions: T08.
- Procedure: audit every P1–P5 screen for ad hoc spacing/typography/color
  values and sub-44px interactive targets at all three supported layouts.
- Expected result: no drift found, or all found drift fixed. Evidence:
  `VOC-031-EV-42`.

## VOC-031-TEST-43 — No already-shipped route renamed
- Covers: `VOC-031-AC-09`, `VOC-031-D08`; Preconditions: T08.
- Procedure: diff the full route list before and after `T08`; assert
  `/reviews` and the absence of a dedicated `/words` route are unchanged.
- Expected result: `D08`'s "do not rename" default is honored. Evidence:
  `VOC-031-EV-43`.

## VOC-031-TEST-44 — Focus-visible and empty/loading/error consistency
- Covers: `VOC-031-AC-09`, `VOC-031-AC-05`; Preconditions: T08.
- Procedure: tab through every P1–P5 screen and assert a visible focus
  indicator on every interactive element; assert every screen's empty/
  loading/error state uses the `T04` shared components.
- Expected result: consistent, accessible interaction states everywhere.
  Evidence: `VOC-031-EV-44`.

## VOC-031-TEST-45 — Installed deterministic and security suite
- Covers: `VOC-031-AC-10`; Preconditions: each PR complete.
- Procedure: run relevant `pnpm validate`/`pnpm test`/`pnpm build`, Go
  format/vet/test/build, web lint/typecheck/build/format,
  `scripts/governance/*` checks as applicable, the new accessibility/
  performance CI jobs, and the extended mock-inventory check.
- Expected result: available checks pass; absent checks reported honestly.
  Evidence: `VOC-031-EV-45`.

## VOC-031-TEST-46 — Staging: full onboard-through-settings loop
- Covers: `VOC-031-AC-10`; Preconditions: F3 staging exists
  (`VOC-031-DEP-02`), `D06`–`D08` resolved, seeded content.
- Procedure: with non-production identities, complete onboarding, discover
  and save a word, submit reviews to complete the daily mission, submit a
  sentence for feedback, view progress, and update a setting and the
  display name, at each of the three supported layouts.
- Expected result: the full loop works coherently in staging across
  supported layouts, recorded without production data. Evidence:
  `VOC-031-EV-46`.

## VOC-031-TEST-47 — New-table rollback rehearsal
- Covers: `VOC-031-AC-10`; Preconditions: staged candidate, approved
  procedure.
- Procedure: rehearse non-production migration rollback for
  `user_onboarding_profiles`; validate `user_settings`/`users.display_name`
  state is preserved and the P1–P4 write paths continue to function after
  rollback.
- Expected result: controlled recovery; no progress/settings-history
  corruption; no P1–P4 write-path outage. Evidence: `VOC-031-EV-47`.

## VOC-031-TEST-48 — Exact-SHA independent verification
- Covers: `VOC-031-AC-10`; Preconditions: each PR at its final SHA.
- Procedure: Claude Code binds to the exact final SHA per PR and verifies
  scope, the classifier floor, migration safety, no A1/P1/P2/P3/P4
  regression, the new write-surface's auth/CSRF/idempotency parity
  (`VOC-031-R01`), that the new CI jobs are blocking not advisory
  (`VOC-031-R06`), the `D06`/`D07`/`D08` resolutions as actually
  implemented, contract/OpenAPI/client drift, accessibility/performance
  evidence, staging/rollback evidence, and implementer separation; reports
  remaining R3/R4/adoption/activation gates.
- Expected result: `PASS` / `PASS WITH NON-BLOCKING FINDINGS` / `FAIL` with
  exact evidence; the implementer did not approve or merge its own work.
  Evidence: `VOC-031-EV-48`.
