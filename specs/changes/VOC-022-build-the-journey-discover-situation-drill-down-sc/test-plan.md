# VOC-022 — Test Plan

Note on tooling: `apps/web` has no test runner yet (no Vitest/RTL/Playwright
— the root `test` script runs only foundation and Go tests), the same gap
VOC-017, VOC-019, VOC-020, and VOC-021 recorded. These tests therefore
combine the available deterministic build/type/lint/format checks
(including `scripts/foundation/check-tailwind-token-usage.mjs`, run by
`lint:web`) with a structured code-inspection checklist. No secrets or
production data are used by any test.

## VOC-022-TEST-00 — New dynamic route renders a per-situation word list from local mock data

- Covers: `VOC-022-AC-00`
- Preconditions: `VOC-022-T00` complete.
- Procedure: inspect `[situation]/page.tsx`; confirm it is declared `async`
  and `await`s its `params` prop; confirm a single local mock data
  structure keyed by all seven known slugs
  (`airport`/`restaurant`/`hotel-check-in`/`job-interview`/
  `daily-conversation`/`work-meeting`/`university-class`) exists, each with
  at least three word entries; confirm no `fetch`, API client import, or
  `TanStack Query` usage appears in the file.
- Expected result: for every known slug, at least three word entries render
  from local mock data only.
- Evidence: `VOC-022-EV-00`

## VOC-022-TEST-01 — Each word entry shows word and meaning; saved marker uses icon and text

- Covers: `VOC-022-AC-01`
- Preconditions: `VOC-022-T00` complete.
- Procedure: inspect the mock data and rendered markup for each situation;
  confirm at least one entry has `isSaved: true` and at least one has
  `isSaved: false` (or is absent from a `isSaved`-filtered subset); confirm
  a saved entry renders both an `aria-hidden="true"` icon element and a
  visible text label (e.g. "Saved"); confirm a not-saved entry renders
  neither.
- Expected result: every situation's list demonstrates both the saved and
  not-saved visual states, with the saved marker conveyed by icon-plus-text,
  never color alone.
- Evidence: `VOC-022-EV-01`

## VOC-022-TEST-02 — No word detail, save/unsave, or sentence-practice affordance

- Covers: `VOC-022-AC-02`
- Preconditions: `VOC-022-T00` complete.
- Procedure: inspect `[situation]/page.tsx`; confirm no word entry renders
  an `href`, `onClick`, `<Link>`, or `<button>`; confirm no control anywhere
  on the page can mutate the mock `isSaved` value; grep the file for any
  sentence-practice-related import or component reference and confirm none
  is present.
- Expected result: no interactive affordance on any word entry; no
  save/unsave control; no sentence-practice entry point.
- Evidence: `VOC-022-EV-02`

## VOC-022-TEST-03 — Unknown/invalid situation slug triggers Next's built-in not-found handling

- Covers: `VOC-022-AC-03`
- Preconditions: `VOC-022-T00` complete; `pnpm run build:web` succeeds (a
  built app is required to exercise routing behavior, since `apps/web` has
  no dev-server-based test runner).
- Procedure: inspect `[situation]/page.tsx`; confirm the branch handling an
  unrecognized `params.situation` value calls `notFound()` imported from
  `next/navigation`, with no alternative fallback (e.g. a default situation)
  and no thrown/uncaught error.
- Expected result: an unrecognized slug resolves to `notFound()`, not a
  custom lookup, thrown error, or blank page.
- Evidence: `VOC-022-EV-03`

## VOC-022-TEST-04 — `discover/page.tsx`'s only change is the navigation addition

- Covers: `VOC-022-AC-04`
- Preconditions: `VOC-022-T00` complete.
- Procedure: run `git diff` on `discover/page.tsx` against its pre-VOC-022
  state (the VOC-021 merge commit); confirm the diff contains only the new
  `next/link` import and the `<Link>` wrapper (plus, if used, moving
  existing className styling from `<li>` onto the `<Link>` and adding
  hover/focus-visible utility classes); confirm `MOCK_DISCOVER_SITUATIONS`'s
  content/order, the `<h1>Journey</h1>` heading, and the intro paragraph are
  unchanged.
- Expected result: the diff is confined to the navigation edit; no other
  content in the file changed.
- Evidence: `VOC-022-EV-04`

## VOC-022-TEST-05 — Static UI only: no backend, no protected-area touch

- Covers: `VOC-022-AC-05`
- Preconditions: `VOC-022-T00` complete.
- Procedure:
  1. Run `git diff --stat` against the base SHA; confirm every changed path
     is under `apps/web/src/app/(app)/discover/`.
  2. Confirm `bottom-nav.tsx`, `(app)/layout.tsx`, `home/page.tsx`, and
     `progress/page.tsx` are byte-for-byte unchanged.
  3. Confirm no file under `apps/api/**` changed, and no auth or migration
     file changed.
  4. Grep the new/changed files for `fetch(`, an
     `@vocanova/api-client`-style import, or `TanStack Query` usage; confirm
     none are present.
- Expected result: change set confined to
  `apps/web/src/app/(app)/discover/`; no backend, auth, or migration file
  touched; no data-fetching code introduced.
- Evidence: `VOC-022-EV-05`

## VOC-022-TEST-06 — Only wired design tokens are used, no bare duration/easing utility

- Covers: `VOC-022-AC-06`
- Preconditions: `VOC-022-T00` complete.
- Procedure:
  1. List every Tailwind color/spacing/radius/shadow/easing utility class
     and every `var(--…)` reference in the new/changed files, including the
     new `<Link>` hover/focus-visible styling in `discover/page.tsx`.
  2. Confirm each resolves to one of the 64 custom properties
     `apps/web/src/app/tokens.generated.css` emits (`spacing`, `neutral`,
     `primary`, `secondary`, `fontSize`/`text-*`, `radius`,
     `elevation`/`shadow-*`, `easing`/`ease-*`, `duration`).
  3. Confirm no raw hex color literal and no `feedback`/`success`/`warning`/
     `error` utility class appears, and confirm the new focus-visible style
     does not reuse `bottom-nav.tsx`'s non-token `blue-600`.
  4. Run `pnpm run lint:web` and confirm
     `scripts/foundation/check-tailwind-token-usage.mjs` reports no bare
     `duration-<name>`/`ease-<name>` utility finding.
- Expected result: every color/spacing/radius/shadow/easing reference
  resolves to a VOC-018-wired custom property; no raw hex, no unwired
  `feedback` scale usage, no bare duration/easing utility class.
- Evidence: `VOC-022-EV-06`

## VOC-022-TEST-07 — Semantic structure: one `<h1>`, `<ul>`/`<li>` word list, no second `<main>`

- Covers: `VOC-022-AC-07`
- Preconditions: `VOC-022-T00` complete.
- Procedure:
  1. Inspect `[situation]/page.tsx`; confirm exactly one `<h1>` (the
     situation title) and no heading element for individual word entries.
  2. Confirm the word list is wrapped in a `<ul>` with one `<li>` per word.
  3. Confirm a "Back to Journey" link to `/discover` is present.
  4. Run `pnpm run lint:web` and confirm
     `scripts/foundation/check-tailwind-token-usage.mjs` reports no
     nested-`<main>` finding; separately confirm the file's root element is
     a `<div>`.
- Expected result: valid `<ul>`/`<li>` list structure, exactly one `<h1>`,
  no `<main>` element, and a working back link.
- Evidence: `VOC-022-EV-07`

## VOC-022-TEST-08 — Deterministic check suite passes

- Covers: `VOC-022-AC-08`
- Preconditions: `VOC-022-T00` complete.
- Procedure: run, from repo root:
  ```bash
  pnpm run lint:web
  pnpm run typecheck:web
  pnpm run build:web
  pnpm run format:check
  ```
  Then confirm `git diff` shows no change to `apps/web/package.json` or the
  lockfile and no new runtime dependency.
- Expected result: all four commands exit zero with no new findings
  elsewhere in the workspace; no dependency/lockfile change.
- Evidence: `VOC-022-EV-08`

No migration or rollback test is applicable (a static front-end addition
plus one minimal existing-file edit, with no persisted state; rollback is a
plain `git revert`). No authorization/security test is applicable (no auth,
input, network, or storage surface — see `impact-analysis.md`). No automated
a11y/component test is applicable (`apps/web` has no test runner yet —
VOC-017's already-flagged follow-up, `VOC-017-D08`/`DEP-05`, is not reopened
or resolved by this task).
