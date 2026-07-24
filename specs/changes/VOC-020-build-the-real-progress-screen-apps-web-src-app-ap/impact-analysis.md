# VOC-020 — Impact Analysis

## Security and privacy

None. The screen renders static mock markup only — no links, buttons, input
fields, network/API calls, data fetching, storage, cookies/tokens,
authentication or authorization surface, and no personal data — the mock
points/streak/history values are fixed placeholder numbers, not derived from
or associated with any real learner. DOC-03 §1's "backend is authoritative"
rule is not engaged at runtime because nothing here fetches or persists
progress/streak/completion state. No secrets are introduced. `apps/api` is
untouched.

## Data and migrations

None. No database, schema, migration, or persisted state of any kind. Purely
a front-end file replacement under `apps/web/src/app/(app)/progress/`.
Rollback is a plain `git revert` of the merge commit with no data
implications.

## Analytics and accessibility

- **Analytics:** none — a static mock Progress screen emits no analytics
  events.
- **Accessibility:** in scope. Each completion-history day marker expresses
  completed/not-completed state in both an icon/text label and a fill
  difference — never color alone (DOC-03 §10); Confidence Points, current
  streak, and longest streak are plain text, inherently screen-reader
  accessible. Honest limitation, same as VOC-017/VOC-019
  (`VOC-017-D08`/`DEP-05`): `apps/web` has no component/a11y test runner yet,
  so this is verified by construction and structured inspection, not
  automated axe/Playwright assertions — that automated coverage remains the
  named follow-up VOC-017 already flagged, not a gap this package reopens or
  resolves.

## Risks, dependencies, and evidence

- `VOC-020-R00` (Low, scoping): "static UI only" against DOC-01 §2's/DOC-03
  §8's fuller Progress description is easy to overrun into building a real
  ledger/streak calculation, a charting component, or a celebration
  animation. Mitigated by the hard non-goals in `specification.md` and the
  explicit `VOC-020-D03`/`D07`/`D08` decisions; the verifier should reject
  any such creep.
- `VOC-020-R01` (Medium, semantic): like VOC-019's Home screen, this renders
  token colors across a large content surface. WCAG 2.2 AA contrast (1.4.3
  text, 1.4.11 non-text) for every `primary`/`secondary`/`neutral`
  text-on-background pairing used must be verified by the implementer and
  re-checked by the independent reviewer.
- `VOC-020-R02` (Medium, accessibility-specific): the completion-history
  strip is the first UI in this codebase to convey a binary state
  (completed/not-completed) per-item across multiple repeated markers — a
  pattern prone to "color alone" mistakes (e.g. green dot vs. gray dot with
  no text). The verifier must specifically check each marker's accessible
  name/label, not just its visual fill (`VOC-020-AC-03`).
- `VOC-020-R03` (Low, forward-compatibility): the mock data shape (e.g.
  `MOCK_PROGRESS_STATE`) is not itself a contract — when the real
  API-wiring package lands, it replaces the mock with a fetched/typed value
  and is expected to reshape or discard this structure freely. No downstream
  code should be written against the mock's exact shape as if it were a DTO.
- `VOC-020-R04` (Low): using the `feedback` token scale (unwired,
  `VOC-018-DEP-05`) would be an easy way to add "success green" to a
  completed-day marker; the spec forbids it (`VOC-020-D05`) precisely
  because it is not wired into `apps/web`'s Tailwind `@theme` layer yet —
  using it would either fail to resolve or require a raw hex value, either
  of which this package disallows.
- `VOC-020-R05` (Low): a bare `duration-<name>`/`ease-<name>` utility class
  is an easy mistake to reintroduce (it was found live on VOC-019-T00's first
  implementation attempt, per `scripts/foundation/check-tailwind-token-usage.mjs`'s
  own header comment) — now caught deterministically by `lint:web`, not just
  by inspection.
- `VOC-020-DEP-01`: Requirement must reach a founder-approved,
  implementation-ready state at adoption. DOC-01 and DOC-03 are approved
  documents, but an approved document is not itself implementation authority
  (`AGENTS.md`); this draft is not authority on its own.
- `VOC-020-DEP-02`: Base state (`base_sha`) is pinned to the drafting-time
  `develop` head (`66731642103b57604c7693e45b279838b9f36fcf`) and must be
  re-pinned to the then-current `develop` head at adoption.
- `VOC-020-DEP-03`: Depends on the VOC-017 shell (`/progress` route,
  `BottomNav`, `(app)/layout.tsx`) and VOC-018's wired Tailwind v4 `@theme`
  tokens (`apps/web/src/app/tokens.generated.css`, 64 custom properties),
  both present on `develop` at drafting time (confirmed by direct
  inspection).
- `VOC-020-DEP-04`: Open scope decision — completion-history granularity,
  day-by-day (this draft's proposal) vs. week-by-week (`VOC-020-D01`).
  Pending founder confirmation at adoption.
- `VOC-020-DEP-05`: Open scope decision — whether the DOC-03 §9 Progress-named
  true empty/first-mission state must be added now or may be deferred to the
  real API-wiring follow-up package (`VOC-020-D06`). Pending founder
  confirmation at adoption.
- `VOC-020-DEP-06`: Required follow-up, not part of this task — DOC-03 §9
  loading/error states (once Progress fetches live data) and wiring the
  `feedback` token scale (`VOC-020-D05`/`D07`).
- `VOC-020-EV-00`..`VOC-020-EV-07`: CI run output (lint/typecheck/build/format)
  plus the independent reviewer's exact-SHA verdict, produced at
  implementation time, not now.
