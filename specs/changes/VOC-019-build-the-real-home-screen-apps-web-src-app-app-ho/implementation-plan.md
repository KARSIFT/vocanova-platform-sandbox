# VOC-019 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is authorized: a
human sets `status`, `approval_status`, and `implementation.authorized` in
`change.yaml` (all at unadopted defaults in this draft), records a
founder-approved implementation-ready state (`VOC-019-DEP-01`), and resolves
the open scope decision on the sentence-practice/review-start affordance
(`VOC-019-D01`, `VOC-019-DEP-04`). Depends on the VOC-017 shell and VOC-018's
wired tokens already on `develop` (`VOC-019-DEP-03`, present at drafting). No
protected areas are touched — every target is under
`apps/web/src/app/(app)/home/**`.

## File reconciliation and implementation sequence

Existing state (do not disturb):
- `apps/web/src/app/(app)/home/page.tsx` — the VOC-017 placeholder
  (`<h1>Home</h1>` + one filler line). This is the file being **replaced** by
  this task; it is the only existing target.
- `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`,
  `apps/web/src/app/(app)/discover/page.tsx`,
  `apps/web/src/app/(app)/progress/page.tsx` — untouched; the Home screen reads
  the existing `/discover` route as its link target but does not modify any of
  these files.
- `apps/web/src/app/tokens.generated.css` — the VOC-018 token layer; reused
  as-is via Tailwind utility classes, not edited.

New targets:
- `apps/web/src/app/(app)/home/page.tsx` (rewritten in place, not a new path).
- Optionally, `apps/web/src/app/(app)/home/_lib/mock-data.ts` (or the mock data
  inline in `page.tsx`) and `apps/web/src/app/(app)/home/_components/*.tsx` if
  the implementer chooses to extract presentational pieces — private,
  non-routable folders, matching the `_components` convention `bottom-nav.tsx`
  already established at the `(app)` level.

Ordered steps (single task, `VOC-019-T00`):

1. Add the local mock data structure (`VOC-019-D00`), commented as placeholder
   pending real API wiring.
2. Rewrite `home/page.tsx` as a server component rendering, in order: the
   Today's Mission section (visual + textual progress), the streak stat, the
   due-review stat (no link, `VOC-019-D02`), and the single `/discover` link
   (`VOC-019-D03`), using only the wired token utilities (`VOC-019-D05`).
3. Run the deterministic checks and confirm no other file changed.

## Validation and independent verification

Deterministic commands (run from repo root):

```bash
pnpm run lint:web
pnpm run typecheck:web
pnpm run build:web
pnpm run format:check
```

(`build:web` proves `/home` still compiles and resolves; `typecheck:web` covers
`next typegen && tsc --noEmit`; `format:check` includes `apps/web` in its
prettier target set.)

Independent verification (exact-SHA, per `CLAUDE.md`): the reviewer confirms
the four required elements render from local mock data only (`VOC-019-AC-00`..
`AC-02`); confirms the single `/discover` link meets the accessibility
checklist (`VOC-019-AC-03`); confirms no API client import, `fetch`,
`TanStack Query`, `/review` route, or sentence-practice component was
introduced, and that `apps/api`, auth, and any migration file are untouched
(`VOC-019-AC-04`); confirms every color/spacing/radius/shadow/easing reference
resolves to one of VOC-018's 64 wired custom properties with no raw hex and no
`feedback`-scale utility (`VOC-019-AC-05`); confirms
`bottom-nav.tsx`/`(app)/layout.tsx`/`discover/page.tsx`/`progress/page.tsx`/
`package.json`/lockfile are byte-for-byte unchanged; and confirms WCAG 2.2 AA
contrast for every token color pairing actually used (`VOC-019-R01`). The
verifier binds its verdict to the exact reviewed commit SHA and confirms the
implementer did not self-approve or self-merge.

## Deployment and rollback

`release.deployment: prohibited`. Merging the implementation to `develop` is
the entire scope; nothing here authorizes a production deployment. Rollback is
a plain `git revert` of the merge commit — the change replaces one static
placeholder file with another static file, and nothing depends on this screen
at runtime beyond ordinary navigation to `/home`. Last-known-good reference is
`develop` at this package's (adoption-time re-pinned) `base_sha`.
