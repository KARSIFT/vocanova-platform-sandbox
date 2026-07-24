# VOC-021 — Tasks

## VOC-021-T00 — Build the real Journey/Discover situation list against local mock data

- Requirement source: `VOC-021-D00`, `VOC-021-D01`, `VOC-021-D02`,
  `VOC-021-D03`, `VOC-021-D04`, `VOC-021-D05`, `VOC-021-D06`, `VOC-021-D07`,
  `VOC-021-D08`
- Acceptance criteria: `VOC-021-AC-00`, `VOC-021-AC-01`, `VOC-021-AC-02`,
  `VOC-021-AC-03`, `VOC-021-AC-04`, `VOC-021-AC-05`, `VOC-021-AC-06`
- Tests: `VOC-021-TEST-00`..`TEST-06`
- Evidence: `VOC-021-EV-00`..`EV-06`
- Status: pending

Single task: one cohesive screen replacing one placeholder file (plus, only
if needed for readability, colocated presentational subcomponents) — the
same small, single-PR shape VOC-019's Home screen and VOC-020's Progress
screen used. Confined to `apps/web/src/app/(app)/discover/`.

Steps:

1. Define a single, clearly-named local mock data structure (e.g.
   `MOCK_DISCOVER_SITUATIONS`), an array of exactly seven entries, each with
   a `slug`, `title`, and `shortDescription` (naming inspired by, but not
   dependent on, DOC-05 §8's `journey_situations` columns), with a source
   comment noting it is placeholder data pending real API wiring
   (`VOC-021-D00`, `VOC-021-D01`). The following text is this draft's
   proposed content — a human may accept it as-is or direct different
   wording at adoption (`VOC-021-D08`, OPEN):
   - Airport — "Check in, get through security, and find your gate with
     confidence."
   - Restaurant — "Order food, ask about the menu, and handle the bill with
     ease."
   - Hotel Check-in — "Check in and out smoothly and ask about hotel
     amenities."
   - Job Interview — "Talk about your experience and answer common interview
     questions."
   - Daily Conversation — "Handle everyday small talk and casual exchanges
     naturally."
   - Work Meeting — "Follow along, share updates, and contribute to
     workplace discussions."
   - University Class — "Understand lectures, ask questions, and join class
     discussion."
2. Replace the contents of `apps/web/src/app/(app)/discover/page.tsx`
   (currently the VOC-017 placeholder) with a server component whose root
   element is a `<div>` — never `<main>`, since `(app)/layout.tsx` already
   renders the group's one `<main>` (`VOC-021-AC-05`) — rendering, in order:
   - one `<h1>Journey</h1>` (matching the existing placeholder and the
     `BottomNav` tab label, `VOC-021-D07`);
   - a short intro line (e.g. "Choose a situation to explore practical
     vocabulary.");
   - a `<ul>` with one `<li>` per situation (`VOC-021-D04`), each rendering
     its title as an `<h2>` and its short description as a plain `<p>`
     (`VOC-021-AC-00`, `VOC-021-AC-01`, `VOC-021-AC-05`) — no `href`,
     `onClick`, `<Link>`, or `<button>` anywhere in the list
     (`VOC-021-AC-02`).
3. Use only Tailwind utilities/`var(--…)` references that resolve to the 64
   custom properties VOC-018 wired (`spacing`, `neutral`, `primary`,
   `secondary`, `fontSize`, `radius`, `elevation`, `easing`, `duration`); no
   raw hex color, no `feedback`/`success`/`warning`/`error` utility, and no
   bare `duration-<name>`/`ease-<name>` utility class — arbitrary-value
   `var(--…)` form only (`VOC-021-AC-04`). Ordinary Tailwind responsive
   breakpoint variants (e.g. `sm:grid-cols-2`) for the single-column-to-grid
   layout are unaffected by this restriction (`VOC-021-D04`).
4. Verify `pnpm run lint:web`, `pnpm run typecheck:web`,
   `pnpm run build:web`, and `pnpm run format:check` all pass
   (`VOC-021-AC-06`), and confirm no other file changed: `bottom-nav.tsx`,
   `(app)/layout.tsx`, `home/page.tsx`, `progress/page.tsx`,
   `apps/web/package.json`, and the lockfile are all untouched
   (`VOC-021-AC-03`).

Scope guards: introduce no runtime dependency and no change to
`apps/web/package.json` or the lockfile; introduce no API client import, no
`fetch`, no `TanStack Query`, no `/discover/[situation]` (or similarly
named) route, no `href`/`onClick`/`<Link>`/`<button>` on a situation entry,
no category/level/word-count/saved-word/search/filter element, and no
change outside `apps/web/src/app/(app)/discover/`. Rollback is a plain
`git revert`.
