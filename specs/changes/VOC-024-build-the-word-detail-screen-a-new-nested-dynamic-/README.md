# VOC-024 — Build the Word Detail Screen

**Draft package — not adopted, not approved, and not implementation authority.**
Prepared from the supplied request and approved DOC-03. A human must resolve the
open product-scope question and adopt this package before implementation.

## Identity and lifecycle

- Package ID: `VOC-024`
- Canonical path: `specs/changes/VOC-024-build-the-word-detail-screen-a-new-nested-dynamic-/`
- Lifecycle: `draft`; all adoption and authorization gates remain unadopted in `change.yaml`.
- Proposed risk: **R1**, a proposal only. The implementation-time path classifier and human/verifier assessment govern the actual class. The planned targets are ordinary `apps/web` route and colocated mock-data files, with no protected-path match.
- Decision owner: founder; target branch: `develop`; originating GitHub issue: none (free-text request).

## Objective and requirement source

Add static, mocked Word Detail at `/discover/[situation]/[word]`. It must show the canonical word or phrase, meaning(s), part of speech, example sentences, and collocation, register, and common-mistake notes, using the VOC-022 mock words as the sole mock-data lineage. DOC-03 §6 and DOC-05 §7 provide the approved content model; the request bounds this slice to static UI.

## Scope, non-goals, risk, and protected areas

Scope is restricted to `apps/web/src/app/(app)/discover/[situation]/`: enrich and share VOC-022's mock word data, make its word rows navigate to the nested route, and add the detail route. The save state is visual only: a disabled/presentational control reflects the existing `isSaved` mock flag and never changes it.

Excluded: API/data fetching, persistence, real save/unsave, sentence practice, review-state labels, `apps/api`, auth, home/progress, `bottom-nav.tsx`, database/migrations, dependencies, token wiring, bare `duration-*`/`ease-*` utilities, and a route-level `<main>`. No protected area is planned and no EHR trigger is identified. Production impact is none; deployment remains prohibited.

## Open decision

DOC-03 §6 describes a sentence-practice entry and saved-word review state, while the request expressly excludes both because no backend exists. `VOC-024-D05` records this bounded deferral for founder confirmation at adoption; it does not silently reinterpret DOC-03.

## Verification, approvals, release, and closure

Implementation evidence will be `pnpm run lint:web`, `pnpm run typecheck:web`, `pnpm run build:web`, and `pnpm run format:check`, plus exact-SHA independent verification under `CLAUDE.md`. This draft asserts no approval, merge, activation, or release authority.
