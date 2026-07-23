# VOC-019 — Impact Analysis

## Security and privacy

None. The screen renders static mock markup and one navigational link. It has
no input fields, no network/API calls, no data fetching, no storage, no
cookies/tokens, no authentication or authorization surface, and no personal
data — the mock streak/due-count/progress numbers are fixed placeholder
values, not derived from or associated with any real learner. DOC-03 §1's
"backend is authoritative" rule is not engaged at runtime because nothing here
fetches or persists progress/mission/scheduling state. No secrets are
introduced. `apps/api` is untouched.

## Data and migrations

None. No database, schema, migration, or persisted state of any kind. Purely a
front-end file replacement under `apps/web/src/app/(app)/home/`. Rollback is a
plain `git revert` of the merge commit with no data implications.

## Analytics and accessibility

- **Analytics:** none — a static mock Home screen emits no analytics events.
- **Accessibility:** in scope. Mission progress is shown both visually and as
  explicit text (never color alone, DOC-03 §10); the streak and due-review
  count are plain text, inherently screen-reader accessible; the Journey/
  Discover link is a real `<Link>` with a visible label, ≥44×44px touch target,
  and visible focus indicator, matching the conventions VOC-017's `BottomNav`
  established (`VOC-017-D03`). Honest limitation, same as VOC-017
  (`VOC-017-D08`/`DEP-05`): `apps/web` has no component/a11y test runner yet, so
  this is verified by construction and structured inspection, not automated
  axe/Playwright assertions — that automated coverage remains the named
  follow-up VOC-017 already flagged, not a gap this package reopens or resolves.

## Risks, dependencies, and evidence

- `VOC-019-R00` (Low, scoping): "static UI only" against DOC-01 §2's broader
  Home description ("… quick entry to discovery and sentence practice") is easy
  to overrun into building the sentence-practice component or a `/review`
  route. Mitigated by the hard non-goals in `specification.md` and the open
  scope decision `VOC-019-D01`; the verifier should reject any such creep.
- `VOC-019-R01` (Medium, semantic): this is the first screen to render token
  colors across a large content surface rather than a thin persistent bar
  (VOC-017's `BottomNav`). WCAG 2.2 AA contrast (1.4.3 text, 1.4.11 non-text)
  for every `primary`/`secondary`/`neutral` text-on-background pairing used
  must be verified by the implementer and re-checked by the independent
  reviewer — this is exactly the "consuming change" obligation VOC-018's impact
  analysis flagged and deferred to whichever package first renders real
  content with these tokens.
- `VOC-019-R02` (Low, forward-compatibility): the mock data shape
  (`MOCK_HOME_STATE` or equivalent) is not itself a contract — when the real
  API-wiring package lands, it replaces the mock with a fetched/typed value and
  is expected to reshape or discard this structure freely. No downstream code
  should be written against the mock's exact shape as if it were a DTO.
- `VOC-019-R03` (Low): using the `feedback` token scale (unwired,
  `VOC-018-DEP-05`) would be an easy way to add "success green" to a progress
  indicator; the spec forbids it (`VOC-019-D05`) precisely because it is not
  wired into `apps/web`'s Tailwind `@theme` layer yet — using it would either
  fail to resolve or require a raw hex value, either of which this package
  disallows.
- `VOC-019-DEP-01`: Requirement must reach a founder-approved,
  implementation-ready state at adoption. DOC-01 and DOC-03 are approved
  documents, but an approved document is not itself implementation authority
  (`AGENTS.md`); this draft is not authority on its own.
- `VOC-019-DEP-02`: Base state (`base_sha`) is pinned to the drafting-time
  `develop` head (`a6da9bf3244048773cc08cdf20d536aba42b517b`) and must be
  re-pinned to the then-current `develop` head at adoption.
- `VOC-019-DEP-03`: Depends on the VOC-017 shell (`/home` route, `BottomNav`,
  `(app)/layout.tsx`) and VOC-018's wired Tailwind v4 `@theme` tokens
  (`apps/web/src/app/tokens.generated.css`, 64 custom properties), both present
  on `develop` at drafting time (confirmed by direct inspection).
- `VOC-019-DEP-04`: Open scope decision — whether the sentence-practice
  quick-entry point and a review-start action belong in this task or a
  follow-up (`VOC-019-D01`). Pending founder confirmation at adoption.
- `VOC-019-DEP-05`: Required follow-up, not part of this task — DOC-03 §9
  empty/loading/error states (once Home fetches live data) and wiring the
  `feedback` token scale (`VOC-019-D05`/`D06`).
- `VOC-019-EV-00`..`VOC-019-EV-06`: CI run output (lint/typecheck/build/format)
  plus the independent reviewer's exact-SHA verdict, produced at
  implementation time, not now.
