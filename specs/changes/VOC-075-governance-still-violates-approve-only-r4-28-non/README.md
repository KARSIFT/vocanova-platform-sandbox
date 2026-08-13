# VOC-075 — Governance Still Violates "Approve Only R4"

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #573](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/573),
prepared for founder/steward review at adoption time.

## Identity and lifecycle

- Package ID: VOC-075
- Title: Governance still violates 'approve only R4': remove non-R4
  `automatic_merge` opt-outs
- Canonical path:
  `specs/changes/VOC-075-governance-still-violates-approve-only-r4-28-non`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; may become `R4` if DOC-15 and/or a
  `scripts/governance` / `tooling/governance` lint is in scope — open
  questions 1 and 3)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#573](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/573)
    (this package's requirement source)
  - [#488](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/488)
    (VOC-068 predecessor; closed)

## Why this exists

VOC-068 fixed half the problem: R0–R2 packages now draft with
`automatic_merge_allowed: true`, and the template no longer silently trains
planners to leave an unconditional `false`. VOC-068 also introduced an R3
"case-by-case, with stated justification" carve-out in `AGENTS.md`, naming
auth, secrets, or production infrastructure as valid reasons for an R3
package to still require founder approval on merge into `develop`.

That carve-out contradicts the founder's explicit, confirmed instruction
(2026-08-14): **approve only R4** — not "R4 plus any R3 someone judges
sensitive."

Issue #573's repo-wide scan lists **28** non-R4 packages still carrying
`automatic_merge_allowed: false`. Most are historical, but **VOC-072** is
active and currently blocking merges that should not need founder attention
under the stated rule.

## What this package does

1. **Rewrite the AGENTS.md drafting rule** (`VOC-075-T00`): R0–R3 always
   draft `automatic_merge_allowed: true`; only R4 sets `false`. Remove the
   R3 case-by-case carve-out entirely.
2. **Align the change-package template** (`VOC-075-T01`) so its comment /
   README no longer describe an R3 justified-false path.
3. **Reconcile governance docs** (`VOC-075-T02`) that would otherwise keep
   contradicting the founder rule; settle DOC-15 under `VOC-075-DEP-00`.
4. **Backfill at least VOC-072** (`VOC-075-T03`); expand only if adoption
   settles `VOC-075-DEP-01` beyond the minimum.
5. **Optionally add a lint** (`VOC-075-T04`) that fails when
   `automatic_merge_allowed: false` appears without `risk: R4` — adoption
   settles inclusion via `VOC-075-DEP-02` (R4 path floor if included).

## What this package deliberately does NOT do

- Not changing `karsift-ai-infra` `merge-gate.yml` (R4 hard block stays).
- Not changing `pipeline.yml` `auto_merge_enabled` or release/deploy
  autonomy switches.
- Not weakening independent verification, CI, EHR, or R4 founder authority.
- Not silently rewriting all 28 historical packages unless adoption expands
  `VOC-075-DEP-01`.
- Does not adopt itself. Every adoption/authorization field stays at the
  unadopted default. This package's own `automatic_merge_allowed: true`
  matches the founder rule for a proposed-R3 draft (flip to `false` at
  adoption only if the package is raised to R4).

## Open questions for the reviewing human

See `specification.md`. The most important:

1. **DOC-15 reconciliation / possible R4 raise** (`VOC-075-DEP-00`).
2. **Backfill scope** (`VOC-075-DEP-01`) — VOC-072 only vs active vs all 28.
3. **Lint inclusion / possible R4 raise** (`VOC-075-DEP-02`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries no standing approval; adoption, implementation authorization,
independent verification, and any required human approval remain to be
recorded against the exact implemented revision, per AGENTS.md and CLAUDE.md.
