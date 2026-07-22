# VOC-014 — Design Tokens: Elevation / Shadow Scale

**Draft package — not adopted, not approved, not implementation authority.**
Prepared by the planner role from a free-text request. A human must review and
adopt it (and record an approved requirement source) before any implementation.

## Identity and lifecycle

- Package ID: `VOC-014`
- Canonical path:
  `specs/changes/VOC-014-add-an-elevation-shadow-token-scale-to-packages-de/`
- Lifecycle state: `draft` (unadopted; see `change.yaml`)
- Proposed risk: R1 (draft proposal only — the authoritative floor is whatever
  `scripts/governance/classify-change-risk.sh` computes at implementation time;
  `packages/design-tokens/src/*` matches no R2/R3/R4 pattern today, so R1 is the
  expected path-detected floor, but a human's judgment governs the final class)
- Owner (decision): founder (m-e-h-r-d-a-a-d)
- Authority issue: not yet assigned — a founder-approved issue must be recorded
  at adoption (the VOC-010→VOC-013 line used issues #4–#7; #8 is the likely next
  number but is **not** assumed approved here)
- Target branch: `develop`

## Objective and requirement source

Continues the design-tokens foundation from VOC-010 (spacing, neutral color),
VOC-011 (typography), VOC-012 (border radius), and VOC-013 (motion — duration
and easing). This package adds a typed **elevation** scale — a small set of
`box-shadow` value tokens — as the next primitive, exported and re-exported the
same way as its sibling token files.

Requirement source: the free-text planning request that produced this draft. It
is **not** an approved requirement on its own; a founder-approved issue must be
recorded before implementation (see `AGENTS.md`: "a chat prompt or issue alone is
not implementation authority").

## Scope, non-goals, risk, and protected areas

In scope: a single typed `elevation` scale object (five `box-shadow` string
values keyed `none`/`sm`/`md`/`lg`/`xl`), in a new
`packages/design-tokens/src/elevation.ts`, re-exported from
`packages/design-tokens/src/index.ts` alongside the existing `spacing`,
`neutral`, `fontSize`, `radius`, `duration`, and `easing` exports.

Non-goals: no component-level usage, no CSS-variable / Tailwind generation, no
dark-mode or per-surface shadow variants, no consumption changes in `apps/web`.

No protected areas touched. No production impact.

## Verification, approvals, release, and closure

Deterministic evidence (at implementation time): `pnpm run lint:packages`,
`pnpm run typecheck:packages`, `pnpm run build:packages`. Independent
verification is exact-SHA per `CLAUDE.md`. Because this is a draft, no approval,
merge, or release authority is claimed here; those gates remain closed in
`change.yaml` and are a human's decision at adoption.
