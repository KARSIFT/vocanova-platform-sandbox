# VOC-014 — Design Tokens: Elevation / Shadow Scale: Specification

## Objective and requirement source

Adds a typed `elevation` scale — a small set of `box-shadow` value tokens — to
`packages/design-tokens`, continuing the VOC-010/011/012/013 foundation. The
requirement source is the free-text planning request that produced this draft;
it must be backed by a founder-approved issue before implementation (see
`AGENTS.md`). This document is a draft, not an approved specification.

## Scope and non-goals

In scope:
- `packages/design-tokens/src/elevation.ts` — a typed `elevation` scale as a
  readonly object, with the exact keys and values specified in
  `acceptance-criteria.md` (`VOC-014-AC-00`).
- `packages/design-tokens/src/index.ts` — add a re-export of `elevation`
  alongside the existing `spacing`, `neutral`, `fontSize`, `radius`, `duration`,
  and `easing` exports (do not remove, rename, or alter those).

Explicitly excluded: component-level consumption, CSS custom property / Tailwind
config generation, dark-mode or per-surface shadow variants, any change in
`apps/web`.

## Risk and protected areas

Proposed risk: R1, same reasoning as VOC-010/011/012/013
(`packages/design-tokens/src/*` falls to the classifier's R1 default). This is a
draft proposal — `scripts/governance/classify-change-risk.sh` and a human's
judgment govern the actual class at implementation time. No protected areas
touched.

## Decisions, contradictions, security, and privacy

`VOC-014-D00`: The scale is five fixed, literal `box-shadow` string values (not
computed) — `none` maps to the CSS keyword `"none"`, and `sm`/`md`/`lg`/`xl`
are increasing-depth composite shadows. Each value is a string literal taken
verbatim from the table in `acceptance-criteria.md` (`VOC-014-AC-00`); there is
no rounding or computation step, so there is no ambiguity about the expected
values.

`VOC-014-D01`: The scale is named `elevation` (file `elevation.ts`, export
`elevation`) because the tokens express elevation depth; their *values* are
`box-shadow` CSS strings. This is a naming decision for review — a human may
prefer `shadow` at adoption. If changed, `elevation` must be replaced
consistently across `elevation.ts`, `index.ts`, and every reference in this
package. Sibling token files each own one primitive (`spacing`, `neutral`,
`fontSize`, `radius`, `duration`, `easing`); `elevation` follows that shape.

`VOC-014-D02`: Single task, per VOC-010/011/012's precedent — the shadow file
and its `index.ts` wiring are one tightly-coupled change reviewed against the
whole package's acceptance criteria, so they must not be split across separate
per-task PRs. (This deliberately does **not** repeat VOC-013's one-off two-task
retry experiment.)

No security, secrets, or personal-data impact.

## Data, migrations, analytics, and accessibility

None. No data storage, no migration, no analytics event, no rendered UI is
introduced by this package. Elevation values affect visual presentation only
once consumed; nothing consumes them here.
