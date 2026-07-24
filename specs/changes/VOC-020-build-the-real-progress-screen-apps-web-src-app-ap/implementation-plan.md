# VOC-020 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is authorized:
a human sets `status`, `approval_status`, and `implementation.authorized` in
`change.yaml` (all at unadopted defaults in this draft), records a
founder-approved implementation-ready state (`VOC-020-DEP-01`), and resolves
the two open scope decisions — completion-history granularity
(`VOC-020-D01`, `VOC-020-DEP-04`) and whether the true empty/first-mission
state is in scope now (`VOC-020-D06`, `VOC-020-DEP-05`). Depends on the
VOC-017 shell and VOC-018's wired tokens already on `develop`
(`VOC-020-DEP-03`, present at drafting). No protected areas are touched —
every target is under `apps/web/src/app/(app)/progress/**`.

## File reconciliation and implementation sequence

Existing state (do not disturb):
- `apps/web/src/app/(app)/progress/page.tsx` — the VOC-017 placeholder
  (`<h1>Progress</h1>` + one filler line). This is the file being
  **replaced** by this task; it is the only existing target.
- `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`,
  `apps/web/src/app/(app)/home/page.tsx`,
  `apps/web/src/app/(app)/discover/page.tsx` — untouched; the Progress
  screen does not link to or modify any of these files.
- `apps/web/src/app/tokens.generated.css` — the VOC-018 token layer; reused
  as-is via Tailwind utility classes, not edited.

New targets:
- `apps/web/src/app/(app)/progress/page.tsx` (rewritten in place, not a new
  path).
- Optionally, `apps/web/src/app/(app)/progress/_lib/mock-data.ts` (or the
  mock data inline in `page.tsx`) and
  `apps/web/src/app/(app)/progress/_components/*.tsx` if the implementer
  chooses to extract presentational pieces — private, non-routable folders,
  matching the `_components` convention `bottom-nav.tsx` and VOC-019's Home
  screen already established at the `(app)` level.

Ordered steps (single task, `VOC-020-T00`):

1. Add the local mock data structure (`VOC-020-D00`), commented as
   placeholder pending real API wiring, with distinct fields for the
   Confidence Points total, current-streak count, longest-streak count, and
   a 7-entry day-by-day completion array (`VOC-020-D01`, `VOC-020-D02`,
   `VOC-020-D07`).
2. Rewrite `progress/page.tsx` as a server component whose root element is
   a `<div>` or `<section>` (never `<main>` — `VOC-020-D04`), rendering, in
   order: the Confidence Points stat, the current-streak stat, the
   longest-streak stat (visibly distinct from current streak,
   `VOC-020-AC-02`), and the completion-history strip (each day marker
   carrying both an icon/text label and a visual fill difference,
   `VOC-020-AC-03`), using only the wired token utilities (`VOC-020-D05`)
   and no charting library.
3. Run the deterministic checks and confirm no other file changed.

## Validation and independent verification

Deterministic commands (run from repo root):

```bash
pnpm run lint:web
pnpm run typecheck:web
pnpm run build:web
pnpm run format:check
```

(`lint:web` runs `scripts/foundation/check-tailwind-token-usage.mjs`, which
deterministically rejects a bare `duration-<name>`/`ease-<name>` utility and
a second `<main>` landmark under `apps/web/src/app/(app)/**/*.tsx`.
`build:web` proves `/progress` still compiles and resolves; `typecheck:web`
covers `next typegen && tsc --noEmit`; `format:check` includes `apps/web` in
its prettier target set.)

Independent verification (exact-SHA, per `CLAUDE.md`): the reviewer confirms
the four required elements render from local mock data only (`VOC-020-AC-00`..
`AC-02`); confirms the completion-history strip is motivational (not a
chart/graph) and every day marker's state is legible without color
(`VOC-020-AC-03`); confirms no API client import, `fetch`, `TanStack Query`,
charting library, or celebration-animation code was introduced, and that
`apps/api`, auth, and any migration file are untouched (`VOC-020-AC-04`);
confirms every color/spacing/radius/shadow/easing reference resolves to one
of VOC-018's 64 wired custom properties with no raw hex and no
`feedback`-scale utility (`VOC-020-AC-05`); confirms `progress/page.tsx`
renders no second `<main>` (`VOC-020-AC-06`); confirms
`bottom-nav.tsx`/`(app)/layout.tsx`/`home/page.tsx`/`discover/page.tsx`/
`package.json`/lockfile are byte-for-byte unchanged; and confirms WCAG 2.2 AA
contrast for every token color pairing actually used (`VOC-020-R01`). The
verifier binds its verdict to the exact reviewed commit SHA and confirms the
implementer did not self-approve or self-merge.

## Deployment and rollback

`release.deployment: prohibited`. Merging the implementation to `develop` is
the entire scope; nothing here authorizes a production deployment. Rollback
is a plain `git revert` of the merge commit — the change replaces one static
placeholder file with another static file, and nothing depends on this
screen at runtime beyond ordinary navigation to `/progress`. Last-known-good
reference is `develop` at this package's (adoption-time re-pinned)
`base_sha`.
