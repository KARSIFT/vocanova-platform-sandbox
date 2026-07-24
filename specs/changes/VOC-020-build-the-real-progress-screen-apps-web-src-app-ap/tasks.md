# VOC-020 — Tasks

## VOC-020-T00 — Build the real Progress screen against local mock data

- Requirement source: `VOC-020-D00`, `VOC-020-D01`, `VOC-020-D02`,
  `VOC-020-D03`, `VOC-020-D04`, `VOC-020-D05`, `VOC-020-D07`, `VOC-020-D08`
- Acceptance criteria: `VOC-020-AC-00`, `VOC-020-AC-01`, `VOC-020-AC-02`,
  `VOC-020-AC-03`, `VOC-020-AC-04`, `VOC-020-AC-05`, `VOC-020-AC-06`,
  `VOC-020-AC-07`
- Tests: `VOC-020-TEST-00`..`TEST-07`
- Evidence: `VOC-020-EV-00`..`EV-07`
- Status: pending

Single task: one cohesive screen replacing one placeholder file (plus, only
if needed for readability, colocated presentational subcomponents) — no
sub-splittable pieces the way VOC-017's shell had to be reasoned about as a
whole. It is a small, single-PR change confined to
`apps/web/src/app/(app)/progress/`.

Steps:

1. Define a single, clearly-named local mock data structure (e.g.
   `MOCK_PROGRESS_STATE`) representing the Confidence Points total, current
   streak count, longest streak count, and a 7-entry array of
   `{ label: string; completed: boolean }` for the current week, with a
   source comment noting it is placeholder data pending real API wiring
   (`VOC-020-D00`). Colocate it either inline in `progress/page.tsx` or in a
   colocated `(app)/progress/_lib/mock-data.ts` — either is acceptable; do
   not add a shared or cross-feature location for it.
2. Replace the contents of `apps/web/src/app/(app)/progress/page.tsx`
   (currently the VOC-017 placeholder) with a server component whose root
   element is a `<div>` or `<section>` — never `<main>`, since
   `(app)/layout.tsx` already renders the group's one `<main>`
   (`VOC-020-AC-06`) — rendering, in order:
   - the Confidence Points total as a text stat (`VOC-020-AC-00`);
   - the current streak as a text stat (`VOC-020-AC-01`);
   - the longest streak as its own, visibly distinct text stat
     (`VOC-020-AC-02`);
   - the completion-history strip: 7 day markers, each showing its
     completed/not-completed state via both an icon/text label and a
     visual fill difference — never color alone (`VOC-020-AC-03`).
3. Use only Tailwind utilities/`var(--…)` references that resolve to the 64
   custom properties VOC-018 wired (`spacing`, `neutral`, `primary`,
   `secondary`, `fontSize`, `radius`, `elevation`, `easing`, `duration`); no
   raw hex color, no `feedback`/`success`/`warning`/`error` utility, and no
   bare `duration-<name>`/`ease-<name>` utility class — arbitrary-value
   `var(--…)` form only (`VOC-020-AC-05`).
4. Verify `pnpm run lint:web`, `pnpm run typecheck:web`,
   `pnpm run build:web`, and `pnpm run format:check` all pass
   (`VOC-020-AC-07`), and confirm no other file changed: `bottom-nav.tsx`,
   `(app)/layout.tsx`, `home/page.tsx`, `discover/page.tsx`,
   `apps/web/package.json`, and the lockfile are all untouched
   (`VOC-020-AC-04`).

Scope guards: introduce no runtime dependency (including no
charting/graphing library) and no change to `apps/web/package.json` or the
lockfile; introduce no API client import, no `fetch`, no `TanStack Query`,
no celebration/milestone animation, no `feedback`-scale utility, and no
change outside `apps/web/src/app/(app)/progress/`. Rollback is a plain
`git revert`.
