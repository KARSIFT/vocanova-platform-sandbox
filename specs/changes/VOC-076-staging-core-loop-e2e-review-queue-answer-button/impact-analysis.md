# VOC-076 — Impact Analysis

## Security and privacy

No new secret, credential, or personal-data handling is introduced. Staging
verification continues to use only the existing non-personal synthetic
smoke-test account (VOC-050). Any product fix in `review-session.tsx` must
preserve existing CSRF (`X-CSRF-Token`), session-expiry handling via
`handleApiError`, and authenticated review submission boundaries —
independent verification must confirm the fix does not bypass those checks or
allow answer submission while a prior submit is still in flight if that was
intentional loading protection.

## Data and migrations

None expected. No schema or migration changes. No historical data backfill.

Integrity / operational risk while unfixed: `develop` staging deploys can fail
the core-loop gate at step 5, blocking confidence in staging; if the cause is
product-side, learners may be unable to select multiple-choice meanings while
the card appears ready.

## Analytics and accessibility

None expected for analytics pipelines. Accessibility: preserve fieldset/legend
naming ("Choose the meaning for …"), `aria-pressed`, and keyboard focus styles.
A fix that clears a stuck `disabled` state improves accessibility; a fix that
only lengthens E2E timeouts without enabling controls for users does not.
Keep the Tailwind `max-w-*` workaround in `review-session.tsx` (see
`.karsift/lessons.md`).

## Risks, dependencies, and evidence

- `VOC-076-R00`: **Staging gate remains red (current defect).** Develop deploys
  fail core-loop E2E at the MC answer click. Mitigation: T00 confirm, T01 fix,
  T02 verify on real staging.
- `VOC-076-R01`: **Real learners stuck on disabled MC options.** If
  `isLoading` hangs in production, review progress stalls. Mitigation: T00 must
  distinguish product vs E2E; T01 fixes product when confirmed.
- `VOC-076-R02`: **E2E-only band-aid over a product hang.** Raising timeouts or
  skipping MC without fixing stuck `isSubmitting`/`isRefetching` leaves users
  broken. Mitigation: T00 evidence; AC-01; independent review rejects timeout-only
  "fixes" when product hang is confirmed.
- `VOC-076-R03`: **Over-broad review-session refactor.** Changing prompt
  scheduling, rating UX, or unrelated helpers. Mitigation: narrow T01; reject
  drive-bys.
- `VOC-076-R04`: **Scope creep into `deploy-staging.yml`.** Unneeded workflow
  edits raise path floor to R3. Mitigation: default exclude (`VOC-076-DEP-01`).
- `VOC-076-R05`: **Evidence confusion with VOC-074.** Mixing increment-bug
  evidence with this timeout. Mitigation: `VOC-076-DEP-02`; AC-04.
- `VOC-076-DEP-00`: Unresolved at drafting — root cause (T00).
- `VOC-076-DEP-01`: Unresolved at drafting — deploy-staging.yml scope.
- `VOC-076-DEP-02`: Unresolved at drafting — VOC-074 coordination.
- `VOC-076-EV-00`: T00 evidence with confirmed cause.
- `VOC-076-EV-01`: T01 diff, regression/readiness proof, local validation.
- `VOC-076-EV-02`: Real staging deploy run proving step 5 past the MC click.
