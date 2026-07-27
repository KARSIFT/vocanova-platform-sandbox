# VOC-031 — Impact Analysis

## Security and privacy

`VOC-031-R00`: **auth-check relaxation weakening the security boundary.** `T01`'s change to
`apps/web/src/middleware.ts` narrows what counts as "not signed in" so a backend hiccup no longer
forces a false sign-out. Implemented incorrectly, this could instead treat an *actually* invalid
or expired session as "still authenticated" on ambiguous input — a real authorization regression,
not a UX fix. Mitigate with: fail-closed design (an ambiguous/errored check only changes what the
learner is *shown*, never what is *granted*; the destination route's own server-side
`requireAuthRedirect` check remains the actual authorization gate and is independently verified to
still reject an invalid session — `VOC-031-TEST-10`), and exact-SHA security/authorization review
of `T01` specifically, consistent with `apps/web/src/middleware.ts`'s listed protected-area status
(`docs/governance/protected-areas.md`, "Authentication and authorization").

`VOC-031-R01`: **mid-flow reliability changes regressing existing idempotency guarantees.** `T01`
also touches the review-session and sentence-practice retry paths that P2 (`VOC-027`) and P3
(`VOC-028`) already made idempotent. A naive retry-safety fix (e.g. a client-side "resubmit on
reconnect" that bypasses the existing idempotency key) could introduce a *new* duplicate-reward
vector on top of the one P4 (`VOC-030-AC-08`) already closed. Mitigate with: reusing the existing
P2 `client_attempt_id` / P3 request-hash idempotency mechanisms rather than inventing a new one,
and a regression test (`VOC-031-TEST-14`) asserting the pre-existing P2/P3 happy-path and
idempotency behavior is unchanged.

`VOC-031-R02`: **third-party data exposure via new tooling.** If `VOC-031-D02`/`D03` activate
axe-core, Playwright, or Lighthouse CI, these must run against local/CI-only fixtures, never
against real learner data, sessions, or production endpoints. Mitigate with a review confirming no
production credential, session, or learner content reaches any newly added tool.

## Data and migrations

None. This package introduces no Ent schema, no migration, and no new table (`VOC-031-AC-07`,
verified by `VOC-031-TEST-33`). No existing A1/P1/P2/P3/P4 write path is modified; `T01` only
changes how the frontend/middleware *reacts* to an already-existing response, never what the
backend writes.

## Analytics and accessibility

Analytics is unaffected — no new event, identifier, or field is introduced.

Accessibility is the direct subject of `T03` and is named explicitly in the DOC-12 §5 P5 gate
wording itself, unlike any prior milestone. `VOC-031-R03`: repeating the A1–P4 precedent of "absent
test automation recorded as a limitation" for a fifth consecutive milestone, on the one milestone
whose own gate names accessibility directly, risks the P5/R1 gate accepting an unverified
accessibility claim on trust alone. Mitigate with `VOC-031-D02` surfacing this explicitly as an
open decision rather than silently repeating the precedent, and with `T03`'s manual pass being
specific and screen-by-screen (`VOC-031-TEST-27`) rather than a bare assertion, whichever way `D02`
resolves.

## Risks, dependencies, and evidence

- `VOC-031-R04`: **false gate declaration.** The DOC-12 §5 P5 gate requires the loop to work
  "coherently in staging." `VOC-031-DEP-02` (F3 does not exist) makes live staging evidence
  impossible for this package, exactly as it was for `VOC-025`/`VOC-026`/`VOC-027`/`VOC-028`/
  `VOC-030`, now compounding across five consecutive milestones. Mitigate with `T05` explicitly
  refusing to declare the gate complete and naming every reason it cannot be (per the user request's
  own instruction to mirror `VOC-025-DEP-01`'s explicit-block pattern, not fabricate staging
  evidence).
- `VOC-031-R05`: **scope creep into R1/R2 territory.** DOC-11 §3 names four production kill
  switches (`AI_FEATURES_ENABLED`, `EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`,
  `NEW_USER_SIGNUP_ENABLED`) that do not exist anywhere in `apps/api` (confirmed at `VOC-031-D00`).
  These are release-operations controls for R2 ("production infrastructure... release operations"),
  not integration/reliability/accessibility/performance work for P5; building them here would blur
  milestone boundaries DOC-12 §6 treats as a strict dependency chain. Mitigate by recording them as
  explicitly out of scope with a forward reference for R2, not silently building or silently
  ignoring them.
- `VOC-031-R06`: **onboarding/Settings scope ambiguity (`D01`).** Building onboarding/Settings
  without a recorded decision would invent unapproved product scope this milestone was never given
  authority for; permanently skipping them without a recorded decision would leave DOC-08's own
  "MVP completion criteria" bullet unsatisfied indefinitely with no visible tracking. Mitigate via
  `VOC-031-D01` surfacing the DOC-01 §3 / DOC-03 §3 / DOC-08 contradiction explicitly, per DOC-12
  §11's change-control rule, rather than resolving it silently in either direction.
- `VOC-031-R07`: **protected-area scope of `.github/workflows/` if `D02`/`D03` activate CI
  additions.** `docs/governance/protected-areas.md` lists `/.github/workflows/` under "Deployment
  and rollback" (supply chain, permissions, environment gates) independent of the authentication
  protected-area concern `T01` already carries. Mitigate by treating any CI-wiring sub-task that
  results from `D02`/`D03` as its own reviewed unit with that specialist concern in view, not folded
  invisibly into `T03`/`T04`'s UI-facing work.
- `VOC-031-DEP-01`..`DEP-04`: dependencies recorded in `change.yaml`.
- `VOC-031-EV-00`..`EV-39`: UI-state, reliability, consistency, accessibility, performance,
  no-new-scope, evaluation, mock-inventory, staging, rollback, and exact-SHA review evidence
  referenced by the acceptance criteria.
