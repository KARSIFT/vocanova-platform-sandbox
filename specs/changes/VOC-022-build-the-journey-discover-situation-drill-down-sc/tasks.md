# VOC-022 — Tasks

## VOC-022-T00 — Build the situation drill-down word list and link it from the Discover situation list

- Requirement source: `VOC-022-D00`, `VOC-022-D01`, `VOC-022-D02`,
  `VOC-022-D03`, `VOC-022-D04`, `VOC-022-D05`, `VOC-022-D06`, `VOC-022-D07`,
  `VOC-022-D08`, `VOC-022-D09`
- Acceptance criteria: `VOC-022-AC-00`..`AC-08`
- Tests: `VOC-022-TEST-00`..`TEST-08`
- Evidence: `VOC-022-EV-00`..`EV-08`
- Status: pending

Single task: one new route plus one minimal, disclosed edit to an existing
file — the same small, single-PR shape VOC-019 through VOC-021 used.
Confined to `apps/web/src/app/(app)/discover/`.

Steps:

1. Create `apps/web/src/app/(app)/discover/[situation]/page.tsx` as an
   `async` server component that `await`s its `params` prop
   (`{ params }: { params: Promise<{ situation: string }> }`, per
   `VOC-022-D01` — Next.js 16's App Router resolves `params` as a
   `Promise`).
2. Define a single, clearly-named local mock data structure (e.g.
   `MOCK_SITUATION_WORD_LISTS`, keyed by the same seven slugs VOC-021
   already uses), with a source comment noting it is placeholder data
   pending real API wiring (`VOC-022-D00`). Each situation entry has a
   `title` and a `words` array; each word has a word/phrase, a short
   meaning, and a mock `isSaved` boolean. The following content is this
   draft's proposed six-word-per-situation content — a human may accept it
   as-is or direct different content/count at adoption (`VOC-022-D02`,
   OPEN):
   - **Airport** — boarding pass (saved), gate, security check, luggage
     (saved), customs, layover.
   - **Restaurant** — menu (saved), reservation, appetizer, bill (saved),
     tip, take-out.
   - **Hotel Check-in** — reservation (saved), front desk, key card,
     amenities (saved), wake-up call, check-out.
   - **Job Interview** — resume (saved), qualifications, references, salary
     expectations (saved), cover letter, follow-up.
   - **Daily Conversation** — small talk (saved), catch up, weekend plans,
     greeting (saved), farewell, casual.
   - **Work Meeting** — agenda (saved), action items, deadline, follow-up
     (saved), stakeholder, brainstorm.
   - **University Class** — lecture (saved), assignment, syllabus, office
     hours (saved), group project, deadline.

   Each word needs a one-sentence learner-facing meaning gloss (naming
   inspired by, but not dependent on, DOC-05 §7's
   `word_meanings.short_definition`); the implementer may draft this text
   directly following the pattern VOC-021's `tasks.md` used for its own
   situation descriptions.
3. If `params.situation` does not match one of the seven known keys, call
   `notFound()` from `next/navigation` (`VOC-022-D04`) — no custom lookup,
   no thrown error.
4. Render, in order (root element a `<div>`, never `<main>` —
   `VOC-022-D08`):
   - a "Back to Journey" `<Link href="/discover">`;
   - one `<h1>` with the situation's title;
   - a short intro line (e.g. "Words already saved are marked below.");
   - a `<ul>` with one `<li>` per word, each rendering the word/phrase, its
     short meaning, and — only when `isSaved` is `true` — a marker
     combining an `aria-hidden="true"` icon glyph with a visible "Saved"
     text label (`VOC-022-D03`, `VOC-022-AC-01`). No `href`, `onClick`,
     `<Link>`, or `<button>` on any word entry (`VOC-022-AC-02`).
5. Edit `apps/web/src/app/(app)/discover/page.tsx` with **only** this
   change (`VOC-022-D05`, OPEN on exact diff shape,
   `VOC-022-DEP-05`): add `import Link from "next/link";`, then wrap each
   card's existing `<h2>`/`<p>` pair in
   `<Link href={\`/discover/${situation.slug}\`}>`, moving the card's
   existing className (border/background/padding/shadow) from the `<li>`
   onto the new `<Link>` and adding a hover and `focus-visible` treatment
   built only from wired tokens (e.g.
   `hover:border-primary-300 focus-visible:outline
   focus-visible:outline-2 focus-visible:outline-offset-[-2px]
   focus-visible:outline-primary-600`). The mock
   `MOCK_DISCOVER_SITUATIONS` array, the `<h1>Journey</h1>` heading, and the
   intro paragraph remain byte-identical.
6. Use only Tailwind utilities/`var(--…)` references that resolve to the 64
   custom properties VOC-018 wired (`spacing`, `neutral`, `primary`,
   `secondary`, `fontSize`, `radius`, `elevation`, `easing`, `duration`); no
   raw hex color, no `feedback`/`success`/`warning`/`error` utility, and no
   bare `duration-<name>`/`ease-<name>` utility class — arbitrary-value
   `var(--…)` form or an auto-generated color/spacing utility only
   (`VOC-022-AC-06`).
7. Verify `pnpm run lint:web`, `pnpm run typecheck:web`,
   `pnpm run build:web`, and `pnpm run format:check` all pass
   (`VOC-022-AC-08`), and confirm no other file changed: `bottom-nav.tsx`,
   `(app)/layout.tsx`, `home/page.tsx`, `progress/page.tsx`,
   `apps/web/package.json`, and the lockfile are all untouched
   (`VOC-022-AC-05`).

Scope guards: introduce no runtime dependency and no change to
`apps/web/package.json` or the lockfile; introduce no API client import, no
`fetch`, no `TanStack Query`, no word-detail route/modal, no save/unsave
control, no sentence-practice entry point, and no change outside
`apps/web/src/app/(app)/discover/` (and, within that folder, no change to
`discover/page.tsx` beyond the single navigation edit in step 5). Rollback
is a plain `git revert`.
