# VOC-031 — Test Plan

No test, fixture, seed file, or evidence may contain a real secret, production URL/data, another
learner's personal content, or a raw session/CSRF token. Discover installed commands at the
adopted base; a missing integration, staging environment (`VOC-031-DEP-02`), open decision
(`VOC-031-DEP-03`, `DEP-04`), or browser/automation tool is never reported as a pass — it is a
recorded limitation or blocker.

## VOC-031-TEST-00 — Discover routes gain loading state
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: render `discover`, `discover/[situation]`, `discover/[situation]/[word]` in a pending
  data-fetch state and assert a calm, `aria-busy`/`aria-label`-carrying skeleton renders.
- Expected result: consistent loading skeleton on all three discover routes. Evidence:
  `VOC-031-EV-00`.

## VOC-031-TEST-01 — Reviews route gains loading state
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: render `reviews` in a pending data-fetch state and assert the same skeleton pattern
  renders.
- Expected result: consistent with discover/home/progress. Evidence: `VOC-031-EV-01`.

## VOC-031-TEST-02 — Discover routes gain error/retry state
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: force a thrown error on each discover route and assert the on-brand error boundary
  renders with a working `reset()` "Try again" control and a navigate-away fallback.
- Expected result: no unstyled Next.js default error page reachable. Evidence: `VOC-031-EV-02`.

## VOC-031-TEST-03 — Reviews route gains error/retry state
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: force a thrown error on `reviews` and assert the same error-boundary pattern renders.
- Expected result: consistent with discover/home/progress. Evidence: `VOC-031-EV-03`.

## VOC-031-TEST-04 — First-time/no-data empty states preserved or added
- Covers: `VOC-031-AC-00`; Preconditions: T00.
- Procedure: render each of the four routes with no data (new learner) and assert an explanatory
  empty state, not a blank area.
- Expected result: matches DOC-03 §9. Evidence: `VOC-031-EV-04`.

## VOC-031-TEST-05 — Root fallback boundary renders on-brand
- Covers: `VOC-031-AC-01`; Preconditions: T00.
- Procedure: force an error inside `apps/web/src/app/layout.tsx` (a location no route-level
  boundary can catch) and assert `global-error.tsx` renders an on-brand, keyboard-accessible
  fallback.
- Expected result: no default Next.js error page reachable anywhere in the app. Evidence:
  `VOC-031-EV-05`.

## VOC-031-TEST-06 — Route-level boundaries remain first line of defense
- Covers: `VOC-031-AC-01`; Preconditions: T00.
- Procedure: force an error inside a route segment covered by `T00` and assert the route-level
  `error.tsx`, not `global-error.tsx`, handles it.
- Expected result: `global-error.tsx` is a last resort, not the primary handler. Evidence:
  `VOC-031-EV-06`.

## VOC-031-TEST-07 — Genuine 401 still redirects to sign-in
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: mock `GET /api/v1/me` returning `401` and assert the middleware still redirects to
  `/signin?returnTo=...`, unchanged from today.
- Expected result: no regression to the existing A1 auth-gate behavior. Evidence: `VOC-031-EV-07`.

## VOC-031-TEST-08 — Backend 5xx does not force a sign-out
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: mock `GET /api/v1/me` returning `500`/`503` and assert the middleware allows the
  request through rather than redirecting to `/signin`.
- Expected result: a backend hiccup never discards an active session. Evidence: `VOC-031-EV-08`.

## VOC-031-TEST-09 — Network exception does not force a sign-out
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: mock a thrown network exception from the `/api/v1/me` fetch and assert the same
  allow-through behavior as `TEST-08`.
- Expected result: consistent with `TEST-08`. Evidence: `VOC-031-EV-09`.

## VOC-031-TEST-10 — Ambiguous auth check never grants data access
- Covers: `VOC-031-AC-02`; Preconditions: T01.
- Procedure: for the `5xx`/network-exception cases above, assert the destination route's own
  server-side auth check (`requireAuthRedirect`) still independently rejects an actually-invalid
  session — the middleware relaxation never becomes a second, weaker authorization path.
- Expected result: no authorization bypass introduced. Evidence: `VOC-031-EV-10`.

## VOC-031-TEST-11 — Sentence-practice input preserved across a failed submission
- Covers: `VOC-031-AC-03`; Preconditions: T01.
- Procedure: submit a sentence, force the request to fail, and assert the typed sentence text is
  still present in the input afterward.
- Expected result: no learner input lost on failure. Evidence: `VOC-031-EV-11`.

## VOC-031-TEST-12 — No duplicate submission on retry
- Covers: `VOC-031-AC-03`; Preconditions: T01.
- Procedure: fail then retry a sentence submission and a review-rating submission; assert exactly
  one persisted attempt/rating each, relying on the existing P2/P3 idempotency keys.
- Expected result: no duplicate reward or duplicate persisted attempt. Evidence: `VOC-031-EV-12`.

## VOC-031-TEST-13 — No fabricated completion state on failure
- Covers: `VOC-031-AC-03`; Preconditions: T01.
- Procedure: force a review-submission or sentence-submission failure and assert the UI never
  renders a completed/rewarded/mission-advanced state the backend did not confirm.
- Expected result: backend remains authoritative even under failure. Evidence: `VOC-031-EV-13`.

## VOC-031-TEST-14 — Pre-existing P2/P3 happy-path and idempotency behavior unchanged
- Covers: `VOC-031-AC-03`; Preconditions: T01.
- Procedure: run the existing VOC-027/VOC-028 review-submission and sentence-submission test
  suites against the `T01` code and assert no regression.
- Expected result: byte-for-byte unchanged happy-path behavior. Evidence: `VOC-031-EV-14`.

## VOC-031-TEST-15 — Design-token compliance across real screens
- Covers: `VOC-031-AC-04`; Preconditions: T02.
- Procedure: run the (possibly extended) `check-tailwind-token-usage.mjs` across every real
  screen's source files and assert no literal, non-token color/spacing value is present.
- Expected result: full token compliance. Evidence: `VOC-031-EV-15`.

## VOC-031-TEST-16 — Loading/error/retry structural parity
- Covers: `VOC-031-AC-04`; Preconditions: T00, T02.
- Procedure: compare the DOM structure/class pattern of each `T00`-added loading/error state
  against the home/progress baseline.
- Expected result: one consistent pattern across all six screens. Evidence: `VOC-031-EV-16`.

## VOC-031-TEST-17 — Sentence-practice parity across its three entry points
- Covers: `VOC-031-AC-04`; Preconditions: T02.
- Procedure: exercise `SentenceFeedback` as invoked from Home, Word Detail, and Review Completion
  and assert identical pending/disabled/disclaimer/report-feedback behavior at each site.
- Expected result: no entry-point-specific divergence. Evidence: `VOC-031-EV-17`.

## VOC-031-TEST-18 — Mobile-first viewport pass (360px, 430px)
- Covers: `VOC-031-AC-04`; Preconditions: T02.
- Procedure: render every real screen at 360px and 430px viewport widths and assert no horizontal
  overflow, clipped content, or overlapping interactive element.
- Expected result: usable at both ends of the target range. Evidence: `VOC-031-EV-18`.

## VOC-031-TEST-19 — 44px-minimum touch targets
- Covers: `VOC-031-AC-04`; Preconditions: T02.
- Procedure: measure every interactive control's rendered hit area on every real screen at 360px.
- Expected result: no control below 44×44px. Evidence: `VOC-031-EV-19`.

## VOC-031-TEST-20 — Desktop-width usability
- Covers: `VOC-031-AC-04`, `VOC-031-D04`; Preconditions: T02.
- Procedure: render every real screen at the `D04`-proposed desktop width (≥1024px) and assert the
  layout remains a coherent, usable wider version of the same experience.
- Expected result: no desktop-only defect. Evidence: `VOC-031-EV-20`.

## VOC-031-TEST-21 — Labelled controls, including icon-only ones
- Covers: `VOC-031-AC-05`; Preconditions: T03.
- Procedure: inspect every interactive control on every real screen for an accessible name
  (visible label, `aria-label`, or equivalent).
- Expected result: no unlabeled control. Evidence: `VOC-031-EV-21`.

## VOC-031-TEST-22 — Keyboard focus order and no trap
- Covers: `VOC-031-AC-05`; Preconditions: T03.
- Procedure: tab through every real screen's interactive elements and assert a logical order, a
  visible focus indicator at each stop, and no trap.
- Expected result: fully keyboard-operable. Evidence: `VOC-031-EV-22`.

## VOC-031-TEST-23 — Non-color-only state indication
- Covers: `VOC-031-AC-05`; Preconditions: T03.
- Procedure: inspect review-rating correctness, mission/streak status, and AI-feedback result
  presentation for a text or icon label in addition to any color cue.
- Expected result: no state conveyed by color alone. Evidence: `VOC-031-EV-23`.

## VOC-031-TEST-24 — Screen-reader-announced errors
- Covers: `VOC-031-AC-05`; Preconditions: T00, T03.
- Procedure: assert each `T00`-added error state uses an appropriate live region or role so
  assistive technology announces it.
- Expected result: errors are perceivable non-visually. Evidence: `VOC-031-EV-24`.

## VOC-031-TEST-25 — Contrast sufficiency
- Covers: `VOC-031-AC-05`; Preconditions: T03.
- Procedure: check text/background and control/background contrast ratios against WCAG 2.2 AA
  thresholds across every real screen's token-driven color pairs.
- Expected result: no failing pair. Evidence: `VOC-031-EV-25`.

## VOC-031-TEST-26 — Automated accessibility scan (if `D02` activates tooling)
- Covers: `VOC-031-AC-05`, `VOC-031-D02`; Preconditions: T03, `D02` resolved to activate.
- Procedure: run an axe-core (or equivalent) scan against every real screen.
- Expected result: no critical/serious violation. Evidence: `VOC-031-EV-26`. If `D02` does not
  activate tooling, this test is recorded as not-run with the gap noted as a limitation, never a
  pass.

## VOC-031-TEST-27 — Manual accessibility pass documented (if `D02` does not activate tooling)
- Covers: `VOC-031-AC-05`, `VOC-031-D02`; Preconditions: T03.
- Procedure: document the manual/visual review performed for `TEST-21`..`TEST-25` per screen.
- Expected result: an honest, specific record — not a bare "passed" assertion. Evidence:
  `VOC-031-EV-27`.

## VOC-031-TEST-28 — Bundle-size spot-check
- Covers: `VOC-031-AC-06`; Preconditions: T04.
- Procedure: inspect the production build output size per route and compare against the
  pre-`VOC-031` baseline.
- Expected result: no unexplained regression from this package's own changes. Evidence:
  `VOC-031-EV-28`.

## VOC-031-TEST-29 — Route-level code splitting intact
- Covers: `VOC-031-AC-06`; Preconditions: T04.
- Procedure: confirm each route still produces its own chunk and no `T00`–`T03` change forced an
  unnecessary shared/monolithic bundle.
- Expected result: splitting preserved. Evidence: `VOC-031-EV-29`.

## VOC-031-TEST-30 — No unnecessary new dependency
- Covers: `VOC-031-AC-06`; Preconditions: T00..T04.
- Procedure: diff `apps/web/package.json` and root `package.json` against the adopted base and
  justify every addition against `D02`/`D03`.
- Expected result: every new dependency traces to an activated decision. Evidence:
  `VOC-031-EV-30`.

## VOC-031-TEST-31 — Automated Lighthouse run (if `D03` activates tooling)
- Covers: `VOC-031-AC-06`, `VOC-031-D03`; Preconditions: T04, `D03` resolved to activate.
- Procedure: run Lighthouse CI against every real screen.
- Expected result: scores meet or a documented gap is recorded against DOC-08's thresholds.
  Evidence: `VOC-031-EV-31`. If `D03` does not activate tooling, recorded as not-run with the gap
  noted as a limitation.

## VOC-031-TEST-32 — Manual performance spot-check documented (if `D03` does not activate tooling)
- Covers: `VOC-031-AC-06`, `VOC-031-D03`; Preconditions: T04.
- Procedure: document the manual bundle/Core-Web-Vitals spot-check performed for `TEST-28`/`29`.
- Expected result: an honest, specific record. Evidence: `VOC-031-EV-32`.

## VOC-031-TEST-33 — No new backend business module, schema, or migration
- Covers: `VOC-031-AC-07`; Preconditions: T05.
- Procedure: diff `apps/api/business`, `apps/api/ent/schema`, and `apps/api/migrations` against
  the adopted base after `T00`–`T04` and assert no addition/removal/rename.
- Expected result: backend domain surface is byte-for-byte unchanged. Evidence: `VOC-031-EV-33`.

## VOC-031-TEST-34 — No new API route
- Covers: `VOC-031-AC-07`; Preconditions: T05.
- Procedure: diff the committed OpenAPI document against the adopted base after `T00`–`T04`.
- Expected result: no new operation ID or path. Evidence: `VOC-031-EV-34`.

## VOC-031-TEST-35 — Extended mock-inventory check passes
- Covers: `VOC-031-AC-07`, `VOC-031-AC-08`; Preconditions: T05.
- Procedure: run the extended `scripts/foundation/mock-inventory.mjs` (with its P1–P4 backend
  allow-list sweep) against the final `T00`–`T05` state.
- Expected result: passes, printing an explicit confirmation. Evidence: `VOC-031-EV-35`.

## VOC-031-TEST-36 — Installed deterministic suite passes end to end
- Covers: `VOC-031-AC-08`; Preconditions: T00..T05.
- Procedure: run `pnpm validate` (format/lint/typecheck/test/build) and `go test ./...` /
  `go vet ./...` for `apps/api` (unchanged by this package, run to confirm no regression) at the
  final SHA.
- Expected result: full suite passes. Evidence: `VOC-031-EV-36`.

## VOC-031-TEST-37 — Staged full-loop exercise documented (blocked by F3)
- Covers: `VOC-031-AC-08`; Preconditions: T05, `VOC-031-DEP-02` unresolved.
- Procedure: document the sign-in → discover → save → review → sentence-practice →
  mission-completes → progress-reads exercise across the `D04`-proposed supported layouts, ready
  to execute once F3 exists.
- Expected result: procedure recorded; live execution explicitly recorded as blocked. Evidence:
  `VOC-031-EV-37`.

## VOC-031-TEST-38 — Rollback rehearsal documented (blocked by F3)
- Covers: `VOC-031-AC-08`; Preconditions: T05, `VOC-031-DEP-02` unresolved.
- Procedure: document a rehearsal reverting this package's frontend/middleware changes to the
  last-known-good revision and confirming the pre-P5 A1–P4 loop still functions.
- Expected result: procedure recorded; live execution explicitly recorded as blocked. Evidence:
  `VOC-031-EV-38`.

## VOC-031-TEST-39 — Exact-SHA independent verification (per PR)
- Covers: `VOC-031-AC-08`; Preconditions: T00..T05.
- Procedure: Claude Code (or the bound independent-verification role) reviews each PR's exact final
  SHA.
- Expected result: `PASS`, `PASS WITH NON-BLOCKING FINDINGS`, or `FAIL` recorded per PR, never
  self-approved by the implementer. Evidence: `VOC-031-EV-39`.
