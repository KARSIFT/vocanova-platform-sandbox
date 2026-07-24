# VOC-024 — Implementation Plan

## Preconditions and protected areas

Do not begin until a human adopts this package, authorizes implementation, records the current `develop` base SHA, and resolves `VOC-024-D05`. No protected paths are planned. Preserve existing VOC-022 list content and do not edit `apps/api`, auth, home/progress, bottom navigation, layouts, manifests, or token files.

## File reconciliation and implementation sequence

Existing target: `[situation]/page.tsx` owns the VOC-022 mock data and list. Reconcile it by moving—not duplicating—the data into private `[situation]/_lib/mock-word-data.ts`, then enrich it while keeping the existing list output intact. New target: `[situation]/[word]/page.tsx`. In the second task, add the detail lookup/render and the minimum word-row links required to reach it. Do not change `discover/page.tsx`; its situation navigation is already established by VOC-022.

## Validation and independent verification

Run from repository root:

```bash
pnpm run lint:web
pnpm run typecheck:web
pnpm run build:web
pnpm run format:check
```

The independent verifier reviews the exact final SHA, path classification, route lookup/not-found behavior, all required fields, mock-only save state, exclusions, token and landmark constraints, accessibility checklist, and confirmation that the implementer did not approve or merge its own work.

## Deployment and rollback

No deployment is authorized. If merged under the applicable future approval flow, rollback is a plain revert of the implementation merge commit. The last-known-good reference is the adoption-time `develop` base SHA.
