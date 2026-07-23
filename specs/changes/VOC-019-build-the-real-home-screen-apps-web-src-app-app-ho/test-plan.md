# VOC-019 — Test Plan

Note on tooling: `apps/web` has no test runner yet (no Vitest/RTL/Playwright —
the root `test` script runs only foundation and Go tests), the same gap
VOC-017 recorded (`VOC-017-D08`/`DEP-05`). These tests therefore combine the
available deterministic build/type/lint/format checks with a structured
code-inspection checklist. No secrets or production data are used by any test.

## VOC-019-TEST-00 — Today's Mission shows review target and progress

- Covers: `VOC-019-AC-00`
- Preconditions: `VOC-019-T00` complete.
- Procedure:
  1. Inspect `home/page.tsx`: confirm a "Today's Mission" section renders the
     mock review target and a progress value derived from local mock data.
  2. Confirm progress is expressed both visually (e.g. a bar/indicator) and as
     explicit text (e.g. "3 of 10 words reviewed today") — disabling color
     must still leave the progress amount legible from the text.
  3. Confirm no `fetch`, API client import, or `TanStack Query` usage appears
     in the file.
- Expected result: mission target and progress are both visible and
  text-legible, sourced from local mock data only.
- Evidence: `VOC-019-EV-00`

## VOC-019-TEST-01 — Current streak is shown

- Covers: `VOC-019-AC-01`
- Preconditions: `VOC-019-T00` complete.
- Procedure: inspect `home/page.tsx`; confirm the mock streak count renders as
  plain text.
- Expected result: streak stat visible as text.
- Evidence: `VOC-019-EV-01`

## VOC-019-TEST-02 — Due-review count is shown, display-only

- Covers: `VOC-019-AC-02`
- Preconditions: `VOC-019-T00` complete.
- Procedure: inspect `home/page.tsx`; confirm the mock due-review count renders
  as plain text and is **not** wrapped in a `<Link>`, `<button>`, `<a>`, or any
  other interactive element.
- Expected result: due-review count visible as non-interactive text.
- Evidence: `VOC-019-EV-02`

## VOC-019-TEST-03 — Accessible route into Journey/Discover

- Covers: `VOC-019-AC-03`
- Preconditions: `VOC-019-T00` complete.
- Procedure: inspect `home/page.tsx`; confirm exactly one `<Link>` (or
  equivalent Next.js navigational element) targets `/discover`, with:
  1. a visible text label (not icon-only);
  2. a ≥44×44px rendered touch target (e.g. `min-h-11 min-w-11` plus padding);
  3. a visible `focus-visible` outline/ring, never suppressed without
     replacement.
- Expected result: exactly one correctly-targeted, accessible link into
  `/discover`.
- Evidence: `VOC-019-EV-03`

## VOC-019-TEST-04 — Static UI only: no backend, no protected-area touch

- Covers: `VOC-019-AC-04`
- Preconditions: `VOC-019-T00` complete.
- Procedure:
  1. Run `git diff --stat` against the base SHA; confirm every changed path is
     under `apps/web/src/app/(app)/home/`.
  2. Confirm `bottom-nav.tsx`, `(app)/layout.tsx`, `discover/page.tsx`, and
     `progress/page.tsx` are byte-for-byte unchanged.
  3. Confirm no file under `apps/api/**` changed, and no auth or migration file
     changed.
  4. Grep the new/changed files for `fetch(`, an `@vocanova/api-client`-style
     import, or `TanStack Query` usage; confirm none are present.
- Expected result: change set confined to `apps/web/src/app/(app)/home/`; no
  backend, auth, or migration file touched; no data-fetching code introduced.
- Evidence: `VOC-019-EV-04`

## VOC-019-TEST-05 — Only wired design tokens are used

- Covers: `VOC-019-AC-05`
- Preconditions: `VOC-019-T00` complete.
- Procedure:
  1. List every Tailwind color/spacing/radius/shadow/easing utility class and
     every `var(--…)` reference in the new/changed files.
  2. Confirm each resolves to one of the 64 custom properties
     `apps/web/src/app/tokens.generated.css` emits (`spacing`, `neutral`,
     `primary`, `secondary`, `fontSize`/`text-*`, `radius`,
     `elevation`/`shadow-*`, `easing`/`ease-*`, `duration`).
  3. Confirm no raw hex color literal and no `feedback`/`success`/`warning`/
     `error` utility class appears.
- Expected result: every color/spacing/radius/shadow/easing reference resolves
  to a VOC-018-wired custom property; no raw hex, no unwired `feedback` scale
  usage.
- Evidence: `VOC-019-EV-05`

## VOC-019-TEST-06 — Deterministic check suite passes

- Covers: `VOC-019-AC-06`
- Preconditions: `VOC-019-T00` complete.
- Procedure: run, from repo root:
  ```bash
  pnpm run lint:web
  pnpm run typecheck:web
  pnpm run build:web
  pnpm run format:check
  ```
  Then confirm `git diff` shows no change to `apps/web/package.json` or the
  lockfile and no new runtime dependency.
- Expected result: all four commands exit zero with no new findings elsewhere
  in the workspace; no dependency/lockfile change.
- Evidence: `VOC-019-EV-06`

No migration or rollback test is applicable (a static front-end file
replacement with no persisted state; rollback is a plain `git revert`). No
authorization/security test is applicable (no auth, input, network, or storage
surface — see `impact-analysis.md`). No automated a11y/component test is
applicable (`apps/web` has no test runner yet — VOC-017's already-flagged
follow-up, `VOC-017-D08`/`DEP-05`, is not reopened or resolved by this task).
