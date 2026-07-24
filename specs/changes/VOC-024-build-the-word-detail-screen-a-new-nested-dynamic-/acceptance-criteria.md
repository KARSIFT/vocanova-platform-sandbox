# VOC-024 — Acceptance Criteria

## VOC-024-AC-00 — Shared VOC-022 mock data supplies all detail content

- Requirement source: `VOC-024-D00`, `VOC-024-D01`, `VOC-024-D02`
- Tasks: `VOC-024-T00`
- Tests: `VOC-024-TEST-00`
- Evidence: `VOC-024-EV-00`
- Result: pending

One private colocated data module is the source for both situation list and detail route. Each existing word has a unique stable word slug within its situation plus canonical term, at least one meaning/part of speech, example sentence, and collocation, register, and common-mistake note. No unrelated mock vocabulary source is introduced.

## VOC-024-AC-01 — Nested route renders the required word detail

- Requirement source: `VOC-024-D01`, `VOC-024-D02`, `VOC-024-D04`
- Tasks: `VOC-024-T01`
- Tests: `VOC-024-TEST-01`
- Evidence: `VOC-024-EV-01`
- Result: pending

For every known situation/word slug, `/discover/{situation}/{word}` renders the canonical term, meaning(s), part of speech, example sentence(s), and labelled Collocation, Register, and Common mistakes notes from the shared local mock data.

## VOC-024-AC-02 — Navigation and invalid-route behavior are correct

- Requirement source: `VOC-024-D01`, `VOC-024-D04`
- Tasks: `VOC-024-T01`
- Tests: `VOC-024-TEST-02`
- Evidence: `VOC-024-EV-02`
- Result: pending

The VOC-022 situation list links each word to its own nested detail URL; the detail page has a visible link back to its situation list. The route awaits Promise-shaped params and calls `notFound()` for unknown situation, unknown word, or a word not belonging to that situation.

## VOC-024-AC-03 — Save state is visual only; excluded behavior is absent

- Requirement source: `VOC-024-D03`, `VOC-024-D05`
- Tasks: `VOC-024-T01`
- Tests: `VOC-024-TEST-03`
- Evidence: `VOC-024-EV-03`
- Result: pending

The save-state treatment visibly reflects mock `isSaved` but is disabled or purely presentational and has no mutation path. The detail route contains no sentence-practice entry, review-state label, API/network call, client persistence, or fake save/unsave interaction.

## VOC-024-AC-04 — Scope and token/landmark constraints hold

- Requirement source: `VOC-024-D00`, `VOC-024-D04`
- Tasks: `VOC-024-T00`, `VOC-024-T01`
- Tests: `VOC-024-TEST-04`, `VOC-024-TEST-05`
- Evidence: `VOC-024-EV-04`, `VOC-024-EV-05`
- Result: pending

Changes are confined to `apps/web/src/app/(app)/discover/[situation]/`; no protected/excluded file changes or dependency change occurs. New markup uses only wired design tokens, contains no bare `duration-*`/`ease-*` utility or raw hex, and does not render `<main>`.

## VOC-024-AC-05 — Deterministic checks pass

- Requirement source: `VOC-024-D00`..`VOC-024-D05`
- Tasks: `VOC-024-T00`, `VOC-024-T01`
- Tests: `VOC-024-TEST-06`
- Evidence: `VOC-024-EV-06`
- Result: pending

`pnpm run lint:web`, `pnpm run typecheck:web`, `pnpm run build:web`, and `pnpm run format:check` exit successfully after implementation.
