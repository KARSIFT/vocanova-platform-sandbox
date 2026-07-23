# VOC-019 — Tasks

## VOC-019-T00 — Build the real Home screen against local mock data

- Requirement source: `VOC-019-D00`, `VOC-019-D01`, `VOC-019-D02`, `VOC-019-D03`,
  `VOC-019-D04`, `VOC-019-D05`
- Acceptance criteria: `VOC-019-AC-00`, `VOC-019-AC-01`, `VOC-019-AC-02`,
  `VOC-019-AC-03`, `VOC-019-AC-04`, `VOC-019-AC-05`, `VOC-019-AC-06`
- Tests: `VOC-019-TEST-00`..`TEST-06`
- Evidence: `VOC-019-EV-00`..`EV-06`
- Status: pending

Single task: one cohesive screen replacing one placeholder file (plus, only if
needed for readability, colocated presentational subcomponents) — no
sub-splittable pieces the way VOC-017's shell (nav + layout + three routes) had
to be reasoned about as a whole. It is a small, single-PR change confined to
`apps/web/src/app/(app)/home/`.

Steps:

1. Define a single, clearly-named local mock data structure (e.g.
   `MOCK_HOME_STATE`) representing the day's mission target, words reviewed so
   far, current streak, and due-review count, with a source comment noting it
   is placeholder data pending real API wiring (`VOC-019-D00`). Colocate it
   either inline in `home/page.tsx` or in a colocated
   `(app)/home/_lib/mock-data.ts` — either is acceptable; do not add a shared
   or cross-feature location for it.
2. Replace the contents of `apps/web/src/app/(app)/home/page.tsx` (currently
   the VOC-017 placeholder) with a server component rendering, in order:
   - a "Today's Mission" section: the review target and progress toward it, as
     both a visual indicator and explicit text (`VOC-019-AC-00`);
   - the current streak as a text stat (`VOC-019-AC-01`);
   - the due-review count as a text stat, with no link (`VOC-019-AC-02`);
   - exactly one `<Link>` into `/discover`, with a visible text label, a
     ≥44×44px touch target, and a visible focus indicator, matching the
     conventions `BottomNav` (`apps/web/src/app/(app)/_components/bottom-nav.tsx`)
     already established (`VOC-019-AC-03`).
3. Use only Tailwind utilities/`var(--…)` references that resolve to the 64
   custom properties VOC-018 wired (`spacing`, `neutral`, `primary`,
   `secondary`, `fontSize`, `radius`, `elevation`, `easing`, `duration`); no raw
   hex color, and no `feedback`/`success`/`warning`/`error` utility
   (`VOC-019-AC-05`).
4. Verify `pnpm run lint:web`, `pnpm run typecheck:web`, `pnpm run build:web`,
   and `pnpm run format:check` all pass (`VOC-019-AC-06`), and confirm no other
   file changed: `bottom-nav.tsx`, `(app)/layout.tsx`, `discover/page.tsx`,
   `progress/page.tsx`, `apps/web/package.json`, and the lockfile are all
   untouched (`VOC-019-AC-04`).

Scope guards: introduce no runtime dependency and no change to
`apps/web/package.json` or the lockfile; introduce no API client import, no
`fetch`, no `TanStack Query`, no `/review` route, no sentence-practice
component, no `feedback`-scale utility, and no change outside
`apps/web/src/app/(app)/home/`. Rollback is a plain `git revert`.
