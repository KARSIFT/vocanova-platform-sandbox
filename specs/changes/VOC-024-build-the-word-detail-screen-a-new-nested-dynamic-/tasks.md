# VOC-024 — Tasks

## VOC-024-T00 — Consolidate and enrich the existing VOC-022 word mock data

- Requirement source: `VOC-024-D00`, `VOC-024-D01`, `VOC-024-D02`
- Acceptance criteria: `VOC-024-AC-00`, `VOC-024-AC-04`, `VOC-024-AC-05`
- Tests: `VOC-024-TEST-00`, `VOC-024-TEST-04`, `VOC-024-TEST-06`
- Evidence: `VOC-024-EV-00`, `VOC-024-EV-04`, `VOC-024-EV-06`
- Status: pending

Extract the existing `MOCK_SITUATION_WORD_LISTS` from `[situation]/page.tsx` to a private colocated module such as `[situation]/_lib/mock-word-data.ts`; import it back into the situation page. Preserve all current situation/term/isSaved entries and add only route/detail fields: word slug, meanings with part of speech, examples, and the three usage-note categories. Mark it as placeholder data pending API wiring. This task is independently reviewable because it preserves current rendered behavior while establishing one source of truth.

## VOC-024-T01 — Add detail routing and static Word Detail UI

- Requirement source: `VOC-024-D01`..`VOC-024-D05`
- Acceptance criteria: `VOC-024-AC-01`..`VOC-024-AC-05`
- Tests: `VOC-024-TEST-01`..`VOC-024-TEST-06`
- Evidence: `VOC-024-EV-01`..`VOC-024-EV-06`
- Status: pending

Create `[situation]/[word]/page.tsx` as an async server component and await both dynamic params. Use the shared data, scope word lookup to the situation, and call `notFound()` for invalid routes. Update only `[situation]/page.tsx` as needed to turn its existing word rows into token-styled, focus-visible links. Render the required detail sections and a disabled/presentational save-state indicator; add no client component, action handler, persistence, sentence practice, review labels, or backend code. Run the required checks and inspect the final path set.
