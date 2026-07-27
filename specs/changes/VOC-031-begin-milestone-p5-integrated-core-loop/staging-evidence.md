# VOC-031 — P5 Integrated Core Loop Staging Evidence

## Scope and authority

This document is the `T05` in-repository evidence index for the P5 integrated-core-loop gate
(DOC-12 §5 P5). It is a **draft**, written before adoption and before `T00`–`T05` implementation.
It records the planned evidence mapping now and must be updated at `T05` implementation time with
the evidence actually produced by the merged `T00`–`T04` PRs plus `T05` itself, preserving this
draft text in git history, matching the pattern established by `VOC-027`/`VOC-028`/`VOC-030`'s own
staging-evidence documents.

## Planned in-repository evidence (to be produced by `T00`–`T05`)

| Evidence | Requirement | Planned source |
| --- | --- | --- |
| `EV-00`..`EV-04` | Loading/error/empty states on discover* and reviews | `T00` route-segment tests |
| `EV-05`, `EV-06` | Root fallback boundary; route-level boundaries remain primary | `T00` boundary tests |
| `EV-07`..`EV-10` | Auth-check fail-closed reliability | `T01` middleware tests |
| `EV-11`..`EV-14` | Mid-flow interruption safety; no P2/P3 regression | `T01` review-session/sentence-practice tests |
| `EV-15`..`EV-20` | Cross-feature UX consistency (tokens, pattern parity, sentence-practice parity, mobile/desktop viewport, touch targets) | `T02` audit tests |
| `EV-21`..`EV-27` | Accessibility (labels, focus, non-color-only state, screen-reader errors, contrast, automated-or-manual pass per `D02`) | `T03` accessibility tests/audit |
| `EV-28`..`EV-32` | Performance (bundle size, code splitting, dependency diff, automated-or-manual pass per `D03`) | `T04` performance tests/audit |
| `EV-33`, `EV-34` | No new backend module/schema/migration/route | `T05` no-new-scope diff |
| `EV-35` | Extended mock-inventory check passes | `T05` `scripts/foundation/mock-inventory.mjs` |
| `EV-36` | Installed deterministic suite passes end to end | `T05` `pnpm validate` + `go test`/`go vet` for `apps/api` |
| `EV-37` | Staged full-loop exercise | **Blocked by `VOC-031-DEP-02`** — F3 does not exist |
| `EV-38` | Rollback rehearsal | **Blocked by `VOC-031-DEP-02`** — F3 does not exist |
| `EV-39` | Exact-SHA independent verification (per PR) | Performed by Claude Code (different model binding) at each PR's exact final SHA — not the implementer's own evidence to record |

## Staged full-loop exercise plan (blocked by F3)

Once `VOC-031-DEP-02` is resolved and a non-production F3 environment is available, execute and
append results for:

### `EV-37` — Sign-in → discover → save → review → sentence-practice → mission-completes → progress-reads

1. With a non-production learner identity, sign in, discover a word by situation, save it, review
   it to completion, submit a sentence from the Review Completion entry point, and confirm the
   daily mission completes and the streak/points reflect the activity on the Progress screen.
2. Repeat the sentence-practice step from its other two entry points (Home, Word Detail) and
   confirm identical behavior (`VOC-031-AC-04`).
3. Repeat the full loop at 360px, 430px, and the `VOC-031-D04`-proposed desktop width and confirm
   no layout defect at any of the three.
4. Simulate a backend `5xx`/timeout mid-session (per `T01`) and confirm a recoverable error state
   appears rather than a false sign-out, and that no learner input or progress is lost.

### `EV-38` — Rollback rehearsal

1. Record the current frontend/middleware artifact digest before deploying this package's build.
2. Deploy the `VOC-031` build, run the full-loop exercise above, then roll back to the
   previously recorded artifact.
3. Confirm the pre-P5 A1–P4 loop still functions identically after rollback (this package touches
   no backend write path, migration, or schema, so no learner data state is expected to be at
   risk).

## `D01`–`D04` disposition note

Per the adopted resolution of `VOC-031-D01`–`D04` (recorded at adoption, not by this draft): the
onboarding/Settings scope decision (`D01`), the accessibility-automation decision (`D02`), the
performance-automation decision (`D03`), and the supported-layouts definition (`D04`) each
determine the actual shape and completeness of the evidence above. This document must record
whichever disposition was actually adopted, not this draft's proposed defaults presented as
decided.

## Rollback triggers

Per `VOC-031` implementation-plan §Deployment and rollback / release-plan §Rollback, initiate
rollback on:

- Any authorization regression traced to the `T01` middleware change.
- A duplicate reward or lost-learner-input incident traced to the `T01` retry-safety changes.
- A confirmed critical accessibility regression reaching learners.
- A regression in the underlying A1/P1/P2/P3/P4 loop this package integrates but does not itself
  own.

## Rollback procedure

1. Redeploy the previous known-good frontend/middleware artifact by digest.
2. Confirm the pre-P5 loop still functions (no backend data state is at risk from this package).
3. Validate with non-production identities.
4. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- **`VOC-031-DEP-02`**: F3 staging does not exist, so `EV-37`/`EV-38` cannot be run live.
  Procedures are documented; live execution is recorded as blocked.
- **`D01`–`D04`**: open founder decisions are not yet resolved by this draft (see
  `specification.md` §Decisions).
- **Accessibility/performance automation** (`D02`/`D03`) may remain a documented manual-pass
  limitation rather than an automated pass, depending on their resolution — recorded honestly
  either way, never asserted as a pass without evidence.

## P5 gate readiness

Per the DOC-12 §5 P5 gate wording: **"the full loop works coherently in staging across supported
layouts with no critical product/security/data/accessibility/reliability defect."**

This gate **cannot be declared complete by this draft, nor by `T00`–`T05` alone**, because:

1. Live staging evidence (`EV-37`, `EV-38`) is blocked by the missing F3 environment
   (`VOC-031-DEP-02`) — the gate's own word "staging" cannot be satisfied in-repository.
2. `D01`–`D04` remain open; several acceptance criteria's evidence quality depends on their
   resolution.
3. Exact-SHA independent verification of every PR (`T00`..`T05`) is Claude Code's responsibility
   (`EV-39`); this implementer does not self-approve.
4. Production deployment is separately governed (A-003 §11/12); R3/R4 founder authority, RL1/RL2
   technical activation, and autonomous production release all remain disabled.

## Follow-up work

- Execute `EV-37`/`EV-38` once F3 staging exists.
- R1 (Staging Readiness): validate this package's evidence and every carried-forward A1–P4
  limitation under production-like staging conditions once F3 exists.
- R2 (Production Readiness): the DOC-11 §3 kill switches this package explicitly did not build
  (`VOC-031-R05`) remain that milestone's scope.
