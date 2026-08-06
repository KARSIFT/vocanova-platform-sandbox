# VOC-040 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and `VOC-040-T00` are approved and implementation
is authorized, per this repository's own `AGENTS.md` ("a chat prompt or issue alone
is not implementation authority"). `AGENTS.md` itself is not listed as a protected
area by `scripts/governance/classify-change-risk.sh` at drafting time beyond being
a governance document read into every planner/implementer/reviewer prompt — edit
it carefully and additively, not by restructuring existing sections.

## File reconciliation and implementation sequence

Existing target: `AGENTS.md` (read in full at drafting time; reproduced above in
this package's own governing rules). No conflicting in-flight work against this
file is known at drafting time. This package's change is purely additive — a new
section — and does not alter any existing sentence in `AGENTS.md`.

Ordered steps:

1. `VOC-040-T00`: add the new section (see `tasks.md` for its required content).
   Placement suggestion: adjacent to the existing "Reporting a bug found outside
   the normal loop" section, since both address process gaps in the governed
   automation loop, but the implementer may choose a different location if a
   clearer fit exists once the surrounding document is read fresh at
   implementation time.
2. Run `git diff --check` (per `AGENTS.md`'s own "Current validation" section for
   governance/documentation changes) to confirm no trailing-whitespace or
   whitespace-only diff issues.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s own "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

This package does not touch `apps/web`, `apps/api`, or shared `packages/`, so
`pnpm validate` is not required by `AGENTS.md`'s own rule, but running it causes no
harm if the implementer prefers a broader check.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision, confirm `VOC-040-AC-00` is
satisfied (the six required elements listed there are all present and accurate),
and confirm no self-approval occurred.

## Deployment and rollback

No deployment applies. This is a documentation-only change to a file read by
future automation prompts, not a runtime artifact deployed anywhere.
`change.yaml` sets `rollback_required: false` accordingly; if the reviewing human
disagrees (e.g. because `AGENTS.md`'s prompt-injection footprint warrants treating
any edit as reversible-by-default regardless of category), that judgment should be
recorded at adoption time rather than assumed here. A plain `git revert` of the
implementation commit is available regardless, per this repository's normal
version-control practice.
