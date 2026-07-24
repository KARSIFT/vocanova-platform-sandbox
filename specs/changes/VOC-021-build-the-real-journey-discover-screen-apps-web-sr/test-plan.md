# VOC-021 — Test Plan

Note on tooling: `apps/web` has no test runner yet (no Vitest/RTL/Playwright
— the root `test` script runs only foundation and Go tests), the same gap
VOC-017, VOC-019, and VOC-020 recorded. These tests therefore combine the
available deterministic build/type/lint/format checks (including
`scripts/foundation/check-tailwind-token-usage.mjs`, run by `lint:web`) with
a structured code-inspection checklist. No secrets or production data are
used by any test.

## VOC-021-TEST-00 — Seven situation cards render from local mock data

- Covers: `VOC-021-AC-00`
- Preconditions: `VOC-021-T00` complete.
- Procedure: inspect `discover/page.tsx`; confirm exactly seven situation
  entries render (Airport, Restaurant, Hotel Check-in, Job Interview, Daily
  Conversation, Work Meeting, University Class, in that order), sourced from
  a single local mock data structure; confirm no `fetch`, API client import,
  or `TanStack Query` usage appears in the file.
- Expected result: seven situation entries visible, in the specified order,
  sourced from local mock data only.
- Evidence: `VOC-021-EV-00`

## VOC-021-TEST-01 — Each card shows a label and a short description only

- Covers: `VOC-021-AC-01`
- Preconditions: `VOC-021-T00` complete.
- Procedure: inspect `discover/page.tsx`; confirm each situation entry
  renders a title and one short descriptive sentence; confirm no category
  badge, `level_band` badge, word-count figure, saved-word indicator, search
  box, or filter control appears anywhere on the page.
- Expected result: each entry shows exactly a label and a short description;
  no additional field or control is present.
- Evidence: `VOC-021-EV-01`

## VOC-021-TEST-02 — Cards are non-interactive; no drill-down route exists

- Covers: `VOC-021-AC-02`
- Preconditions: `VOC-021-T00` complete.
- Procedure:
  1. Inspect `discover/page.tsx`; confirm no situation entry renders an
     `href`, `onClick`, `<Link>`, or `<button>`.
  2. Run `git diff --stat` (or list the change set) and confirm no
     `apps/web/src/app/(app)/discover/[situation]` (or similarly named
     dynamic-segment) route file was added.
- Expected result: no interactive affordance on any situation entry; no
  drill-down route exists.
- Evidence: `VOC-021-EV-02`

## VOC-021-TEST-03 — Static UI only: no backend, no protected-area touch

- Covers: `VOC-021-AC-03`
- Preconditions: `VOC-021-T00` complete.
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
- Evidence: `VOC-021-EV-03`

## VOC-021-TEST-04 — Only wired design tokens are used, no bare duration/easing utility

- Covers: `VOC-021-AC-04`
- Preconditions: `VOC-021-T00` complete.
- Procedure:
  1. List every Tailwind color/spacing/radius/shadow/easing utility class
     and every `var(--…)` reference in the new/changed files.
  2. Confirm each resolves to one of the 64 custom properties
     `apps/web/src/app/tokens.generated.css` emits (`spacing`, `neutral`,
     `primary`, `secondary`, `fontSize`/`text-*`, `radius`,
     `elevation`/`shadow-*`, `easing`/`ease-*`, `duration`).
  3. Confirm no raw hex color literal and no `feedback`/`success`/`warning`/
     `error` utility class appears.
  4. Run `pnpm run lint:web` and confirm
     `scripts/foundation/check-tailwind-token-usage.mjs` reports no bare
     `duration-<name>`/`ease-<name>` utility finding.
- Expected result: every color/spacing/radius/shadow/easing reference
  resolves to a VOC-018-wired custom property; no raw hex, no unwired
  `feedback` scale usage, no bare duration/easing utility class.
- Evidence: `VOC-021-EV-04`

## VOC-021-TEST-05 — Semantic list structure, single `<h1>`, no second `<main>` landmark

- Covers: `VOC-021-AC-05`
- Preconditions: `VOC-021-T00` complete.
- Procedure:
  1. Inspect `discover/page.tsx`; confirm the seven situation entries are
     wrapped in a `<ul>` with one `<li>` per entry.
  2. Confirm exactly one `<h1>` ("Journey") and seven `<h2>` elements (one
     per situation title).
  3. Run `pnpm run lint:web` and confirm
     `scripts/foundation/check-tailwind-token-usage.mjs` reports no
     nested-`<main>` finding; separately confirm the file's root element is
     a `<div>`.
- Expected result: valid `<ul>`/`<li>` list structure, one `<h1>` plus seven
  `<h2>` elements, no `<main>` element in `discover/page.tsx`.
- Evidence: `VOC-021-EV-05`

## VOC-021-TEST-06 — Deterministic check suite passes

- Covers: `VOC-021-AC-06`
- Preconditions: `VOC-021-T00` complete.
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
- Evidence: `VOC-021-EV-06`

No migration or rollback test is applicable (a static front-end file
replacement with no persisted state; rollback is a plain `git revert`). No
authorization/security test is applicable (no auth, input, network, or
storage surface — see `impact-analysis.md`). No automated a11y/component
test is applicable (`apps/web` has no test runner yet — VOC-017's
already-flagged follow-up, `VOC-017-D08`/`DEP-05`, is not reopened or
resolved by this task).
