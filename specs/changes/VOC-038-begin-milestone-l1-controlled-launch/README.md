# VOC-038 — Begin Milestone L1 (Controlled Launch)

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It exists so the founder has a ready-to-review scoping proposal for L1 on return, prepared
under the 2026-08-02 standing delegation ("implement next plans on develop, manage it
yourself") once R2 (VOC-037) evidence merged to `main` and no other open work existed.

## Why this exists now

DOC-12 (`docs/product/12-mvp-implementation-plan.md`) §3 names L1 as the milestone directly
after R2, and §5 gives L1's objective and gate verbatim (quoted in `specification.md`). With
R2's evidence merged and issue #265 (the R2 go/no-go record) the only open item — explicitly
founder-reserved — there was no more R0–R3 work I could responsibly generate on my own
initiative. Drafting L1's scope is legitimate prep work per DOC-12 §3 ("work may be prepared
early, but a milestone can't be *accepted* before its dependency passes").

## What this package deliberately does NOT do

- It does not declare R2 closed. Only the founder's own go/no-go record on issue #265 can do
  that.
- It does not adopt itself. `change.yaml` sets `status: proposed`, `approval_status: proposed`,
  `implementation_authorized: false` everywhere. No task in `tasks.md` may be dispatched until
  a real adoption decision is recorded, exactly like every prior milestone package.
- It does not answer the two genuine founder judgment calls L1 needs (initial allowlist
  composition; expansion thresholds) — see `specification.md` "Open questions." I could have
  guessed defaults, but DOC-12 §4 assigns exactly this kind of decision to the founder, and a
  guessed threshold that's wrong in either direction (too loose → real users see a bad AI
  feedback experience or a cost overrun; too strict → launch stalls on an arbitrary number) is
  worse than an honest gap.

## Operational note: `develop` was recreated during this drafting pass

While preparing this package I found `origin/develop` no longer existed (deleted, apparently by
the orchestrator's promotion tooling immediately after PR #285 merged — repo-level
`delete_branch_on_merge` is `false`, so this is pipeline behavior, not a GitHub default). Since
`deploy-staging.yml` triggers on push to `develop`, a permanently-missing `develop` would silently
stop staging deploys. I recreated `develop` from `main`'s current tip (`dcae793`, which already
contains everything through PR #285 — verified the one commit I'd pushed to the old `develop`
before it was deleted, `90eb19e`, cherry-picks as empty against `main`, i.e. no work was lost) and
based this proposal branch on the recreated `develop`. Worth the founder confirming this matches
the orchestrator's intended `develop` lifecycle rather than assuming it's fixed.

## Structure

Mirrors the VOC-037 (R2) package's convention: `specification.md`, `acceptance-criteria.md`,
`impact-analysis.md`, `implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the founder

1. Record `VOC-037-AC-05` on issue #265 (R2 go/no-go). If R2 is accepted, L1 can proceed to
   review; if not, this package should be revised to reflect whatever follow-up work R2's
   closure requires first.
2. Read `specification.md`'s two open questions and either answer them directly or authorize
   `T00`/`T05` as decision-record-only tasks (same pattern VOC-037 used for `T00`/`T02`).
3. Adopt (or request changes to) this package the same way prior milestone packages were
   adopted — an adoption PR with founder approval recorded.
