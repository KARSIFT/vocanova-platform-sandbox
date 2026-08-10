# VOC-053 — Staging Core-Loop E2E: "Words Reviewed Today" Reads 1, Then 0, Moments Later

## Identity and lifecycle

- Package ID: VOC-053
- Title: Investigate and Fix the Same-Run, Same-Day Decrease of the "Words
  Reviewed Today" Counter on Staging's Real Core-Loop E2E Check
- Canonical path: `specs/changes/VOC-053-staging-core-loop-e2e-words-reviewed-today`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`, not a determination)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issue: [#450](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/450)

## Objective and requirement source

Make `tests/staging-e2e/core-loop.staging.spec.ts`'s step 7 assertion
(`reviewedAfter >= reviewedBefore + reviewedCards`) pass reliably against real
staging, by finding and fixing the real cause of a same-run, same-day decrease
in the "words reviewed today" count the 2026-08-09 run observed (`1` then `0`,
about ten seconds apart, no midnight boundary crossed) — not by weakening,
retrying past, or otherwise routing around the assertion. Full requirement
grounding is in `specification.md` and `change.yaml`'s `requirement_source`.

## Scope, non-goals, risk, and protected areas

See `specification.md` for full scope, non-goals, and the three candidate root
causes issue #450 lists as non-prescriptive starting points (Next.js fetch/
route caching; a real backend day-boundary/timezone bug in how
`reviews_completed` is computed; a synthetic-account test-data interaction).
See `impact-analysis.md` for risk, protected areas, and dependencies. This
package's own drafting pass independently re-read every file issue #450
points at (`apps/web/src/app/(app)/home/page.tsx`,
`apps/web/src/lib/api-server.ts`, `apps/api/business/missions/repository.go`,
`apps/api/business/gamification/timezone.go`) and could not, from static
reading alone, find a code-level bug in the backend's `reviews_completed`
increment or local-date computation, and confirmed `HomePage` already calls
Next.js's `headers()` (via `createServerApiClient`) on every render, which
Next.js's App Router treats as a dynamic API that opts the route out of the
Full Route Cache — narrowing, but not resolving, the caching candidate. This
package does not resolve which candidate is the real cause; that requires the
live HTTP response header / cache-status inspection issue #450 itself says
this environment cannot do blind. See `specification.md`'s open questions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries no standing approval; adoption, implementation authorization,
independent verification, and any required human approval remain to be
recorded against the exact implemented revision, per AGENTS.md and CLAUDE.md.

## Supersession (VOC-063, 2026-08-10)

`VOC-053-T00` completed its investigation objective: three independent passes
(two `VOC-053-T00` implementer attempts plus issue #473's third pass with live
staging/production access) ruled out all named root-cause candidates from issue
#450 with direct evidence
([third-pass comment](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/450#issuecomment-5238054774)).
`VOC-053-T01` and `VOC-053-T02` are **cancelled** and superseded by
[VOC-063](../VOC-063-voc-053-investigation-exhausted-3-independent/) (issue
[#473](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/473)), which
hardens staging core-loop E2E step 7 with bounded retry-and-reverify instead of
pursuing a production fix no investigation pass located. Issue
[#450](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/450) remains
open for the original symptom. See `tasks.md` for per-task status.
