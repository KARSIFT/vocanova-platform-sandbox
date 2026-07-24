# VOC-024 — Build the Word Detail Screen: Specification

## Objective and requirement source

Create `apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx`, a static server-rendered detail view for a known word in a known VOC-022 situation. DOC-03 §6 requires canonical word/phrase, meanings, part of speech, example sentences, and usage notes; DOC-05 §7 supplies the canonical model of meanings, examples, and usage-note types. The supplied request authorizes only placeholder data and visual save state.

## Scope and non-goals

In scope:

1. Move VOC-022's `MOCK_SITUATION_WORD_LISTS` into one private colocated module under `[situation]/`, preserving its situation/word lineage and adding each word's stable URL slug, part of speech, one-or-more meanings, one-or-more example sentences, and collocation, register, and common-mistake notes.
2. Update the existing situation route to import that shared data and link each word row to `/discover/{situation}/{wordSlug}`. This is a deliberate progression from VOC-022's prior non-interactive word rows.
3. Add the nested async route. It awaits `params`, finds the requested situation and word in the shared local mock structure, and calls `notFound()` for either unknown segment.
4. Render a back link, exactly one page `<h1>` with the canonical term, meanings and part of speech, example sentences, the three named usage-note categories, and an inert save-state indicator/control. It is visibly disabled (or otherwise strictly presentational), has no handler, mutation, client state, API call, or persistence, and uses `isSaved` only to choose its label.
5. Use only already wired tokens through allowed Tailwind utilities or `var(--...)`; no raw colors, unwired feedback scale, bare duration/easing utility, or second `<main>`.

Non-goals: sentence practice; `due today`, `learning`, `mastered`, or any review state; real save/unsave; loading/error states for a network source that does not exist; backend, API, auth, persistence, analytics, migration, dependency, `bottom-nav.tsx`, home, or progress changes.

## Risk and protected areas

Proposed R1 is not authoritative. Planned paths fall under the classifier's default application-code rule, not its auth, migration, governance, workflow, manifest, or other protected patterns. The verifier may escalate for semantic risk. Accessibility and the transition of VOC-022 word rows from static text to links are material review points, but no protected area is intended.

## Decisions, contradictions, security, and privacy

`VOC-024-D00` — static local mock data only. The shared module is the enriched form of VOC-022's existing structure, not a new unrelated vocabulary source. It contains no learner data.

`VOC-024-D01` — stable `wordSlug` fields identify detail URLs. Word lookup must be constrained to the requested situation; a word slug from another situation is not a valid route and calls `notFound()`.

`VOC-024-D02` — every detail record contains canonical term, meanings with part of speech, example sentences, and nonempty collocation, register, and common-mistake notes. Placeholder copy is clearly local mock content pending real API wiring.

`VOC-024-D03` — `isSaved` remains a read-only mock flag. The detail screen may display a disabled button or presentational status labelled "Saved"/"Not saved", but never implies an action will succeed; no `onClick`, client state, `fetch`, API client, or persistence exists.

`VOC-024-D04` — semantic structure is one `<h1>`, ordered content sections with headings, semantic lists for multiple meanings/examples/notes where applicable, a visible focus style for links, and no `<main>` because the app layout provides it.

`VOC-024-D05` — **OPEN: deferred DOC-03 §6 behavior.** DOC-03 also calls for sentence-practice entry and live saved-word review state. The supplied request expressly prohibits them. This draft treats that as an approved-package adoption decision: confirm the static-screen deferral now and schedule real behavior only after the backend contract exists.

There is no security or privacy surface: no input, secret, cookie, network, storage, account, or real learner state. The mock saved marker must not be represented as live user data.

## Data, migrations, analytics, and accessibility

No database, migration, persisted data, or analytics event is added. Rollback is a plain revert. Accessibility is in scope: keyboard-operable links, visible focus indication, text labels that do not rely on color, semantic heading/list structure, and 44px minimum touch target for the save-state control if rendered as a button. The existing repository has no web component/a11y test runner, so these are structured inspection items in addition to build checks.
