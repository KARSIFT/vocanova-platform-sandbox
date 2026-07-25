# VOC-025 — A1 Learning-Content Mock Reconciliation

## Scope and authority

This document is the T05 inventory required by `VOC-025-D05` (founder decision,
2026-07-24): every existing VOC-010–VOC-024 learning-content mock/placeholder
source is retained as-is for A1. A1 adds only the identity/session/auth layer
underneath the `(app)` route group. It does not build, wire, or invent any
learning-domain endpoint.

## Inventory

| File | VOC source | Mock identifier | Disposition | Follow-up / note |
|------|------------|-----------------|-------------|------------------|
| `apps/web/src/app/(app)/home/page.tsx` | VOC-019 | `MOCK_HOME_STATE` | **Retain** | Replace with real Home API wiring in a later P1–P4 package. Must remain behind the A1 auth shell. |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-020 | `MOCK_PROGRESS_STATE` | **Retain** | Replace with real Progress API wiring in a later P1–P4 package. Must remain behind the A1 auth shell. |
| `apps/web/src/app/(app)/discover/page.tsx` | VOC-021 | `MOCK_DISCOVER_SITUATIONS` | **Retain** | Replace with real Journey/Discover API wiring in a later P1–P4 package. Must remain behind the A1 auth shell. |
| `apps/web/src/app/(app)/discover/[situation]/page.tsx` | VOC-022 | `MOCK_SITUATION_WORD_LISTS` (imported) | **Retain** | Replace with real situation/word API wiring in a later P1–P4 package. Must remain behind the A1 auth shell. |
| `apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx` | VOC-022 | `MOCK_SITUATION_WORD_LISTS` (imported) | **Retain** | Replace with real word-detail API wiring in a later P1–P4 package. Must remain behind the A1 auth shell. |
| `apps/web/src/app/(app)/discover/[situation]/_lib/mock-word-data.ts` | VOC-022 | `MOCK_SITUATION_WORD_LISTS`, `mockWord` | **Retain** | Shared local data source for the two VOC-022 routes above. Replace together with a real P1–P4 contract. |
| `apps/web/src/app/(app)/_components/app-header.tsx` | VOC-017 | n/a (real component) | **Wire to A1** | Already uses the real `@vocanova/api-client` logout operation with CSRF double-submit. No mock data. |
| `apps/web/src/app/(app)/_components/bottom-nav.tsx` | VOC-017 | n/a (real component) | **Wire to A1** | Pure navigation shell, no data. Already protected by the A1 middleware. |

## Disposition definitions

- **Retain**: keep the file and its local mock data exactly as-is. Do not add
  real API calls, real database access, or real learner state. The source
  comments already identify each constant as placeholder data pending a
  follow-up package.
- **Wire to A1**: the component/route is already real code but must depend on
  the A1 auth layer for identity/logout. It is not a mock and is not covered by
  the retain rule.

## Verified boundaries

- No learning-domain API route exists in `apps/api/app/api/` other than the A1
  auth/current-user contract (`/api/v1/me`, `/api/v1/auth/*`).
- No learning-domain table exists in `apps/api/ent/schema/` or
  `apps/api/migrations/`.
- The `(app)` route group is protected by `apps/web/src/middleware.ts`.
- The A1 auth shell is the only real source wired underneath the retained
  mocks.

## Follow-up packages

- VOC-019 follow-up: real Home screen learner state (P1/P4 boundary).
- VOC-020 follow-up: real Progress screen learner state (P1/P4 boundary).
- VOC-021 follow-up: real Journey/Discover situation catalog (P1/P4 boundary).
- VOC-022 follow-up: real situation drill-down and word detail (P1/P4 boundary).

These replacements are explicitly out of scope for VOC-025 and are blocked
until their own P1–P4 milestone contracts exist.
