# VOC-024 — Test Plan

No component/a11y runner exists in `apps/web`; the following combines installed deterministic checks with structured inspection. No test uses secrets or production data.

## VOC-024-TEST-00 — Shared enriched mock source

- Covers: `VOC-024-AC-00`
- Preconditions: `VOC-024-T00` complete.
- Procedure: inspect the private mock module and its import in the situation page; confirm each existing word has a unique situation-local slug and all required detail fields, and that no second vocabulary dataset exists.
- Expected result: one enriched VOC-022-derived mock source feeds both pages.
- Evidence: `VOC-024-EV-00`

## VOC-024-TEST-01 — Required detail content renders

- Covers: `VOC-024-AC-01`
- Preconditions: `VOC-024-T01` complete.
- Procedure: inspect the nested route and representative records from every situation; confirm canonical term, meaning(s), part of speech, examples, and Collocation/Register/Common mistakes are all rendered from the shared data.
- Expected result: each known nested route has complete static detail.
- Evidence: `VOC-024-EV-01`

## VOC-024-TEST-02 — Navigation and not-found branches

- Covers: `VOC-024-AC-02`
- Preconditions: `VOC-024-T01` complete.
- Procedure: confirm word-row links use their situation and word slug, the back link targets the matching situation, params are awaited, and unknown situation, unknown word, and cross-situation word call `notFound()`.
- Expected result: valid routes navigate correctly and invalid routes use Next handling.
- Evidence: `VOC-024-EV-02`

## VOC-024-TEST-03 — No fake save or deferred feature

- Covers: `VOC-024-AC-03`
- Preconditions: `VOC-024-T01` complete.
- Procedure: inspect the save treatment and grep changed files for `onClick`, `useState`, `fetch`, API client imports, sentence-practice references, and review labels `due today`, `learning`, `mastered`.
- Expected result: save state is disabled/presentational only and all excluded features are absent.
- Evidence: `VOC-024-EV-03`

## VOC-024-TEST-04 — Scope and data safety inspection

- Covers: `VOC-024-AC-04`
- Preconditions: both tasks complete.
- Procedure: inspect the implementation diff against its approved base; confirm changed app paths are only under `[situation]/` and excluded files, dependencies, backend, auth, and migrations are untouched.
- Expected result: bounded static front-end change only.
- Evidence: `VOC-024-EV-04`

## VOC-024-TEST-05 — Token, landmark, and accessibility inspection

- Covers: `VOC-024-AC-04`
- Preconditions: `VOC-024-T01` complete.
- Procedure: inspect class names and CSS references for wired tokens only; confirm no raw hex, bare duration/easing, or `<main>`; confirm one h1, semantic lists/sections, visible link focus, text-labelled save state, and 44px disabled control if a button is used.
- Expected result: token and WCAG-oriented structure constraints hold.
- Evidence: `VOC-024-EV-05`

## VOC-024-TEST-06 — Installed deterministic checks

- Covers: `VOC-024-AC-05`
- Preconditions: both tasks complete.
- Procedure: run `pnpm run lint:web`, `pnpm run typecheck:web`, `pnpm run build:web`, and `pnpm run format:check` from the repository root.
- Expected result: every command exits zero.
- Evidence: `VOC-024-EV-06`
