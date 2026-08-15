# VOC-082 — Impact Analysis

## Security and privacy

- No authentication, authorization, secrets, or personal-data schema
  changes are in default scope.
- Completing-review writes continue inside the existing review
  submission transaction; the fix removes a false-positive integrity
  error rather than loosening ledger uniqueness.
- Do not weaken idempotency keys for daily-mission completion points or
  streak grace-day ledger entries.
- Staging E2E evidence must not commit production secrets or raw session
  cookies into package files.

## Data and migrations

- No Atlas/Postgres schema migration anticipated.
- Forward-fix only by default: accounts already stuck at
  `reviews_completed = review_target - 1` with `status=open` after a
  rolled-back completing review are not automatically repaired unless
  adoption expands open question 2 in `specification.md`.
- Successful completing reviews after the fix should leave today
  `completed` with completion reward + streak updates in one commit.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability for
  product analytics instrumentation.
- **Accessibility:** None expected for the default API-only fix. If T01
  requires a narrow E2E wait tweak only, preserve existing a11y
  attributes and the `max-w-*` workaround noted in `.karsift/lessons.md`
  if any review UI file is touched (not expected).

## Risks, dependencies, and evidence

- `VOC-082-R00`: **Incomplete fix** that only special-cases one test
  shape but still fails when today is completed in-list with
  `currentCompletion=true` under real `applyP4ReviewWiring` ordering.
  Mitigation: AC-00/TEST-00 mirror the live fetch-after-complete
  sequence.
- `VOC-082-R01`: **Weakened future-date guard** accidentally accepting
  corrupt future snapshots. Mitigation: AC-01/TEST-02.
- `VOC-082-R02`: **Double completion reward / streak double-advance** if
  the today+currentCompletion path bypasses idempotency. Mitigation:
  preserve existing ledger keys and status=`open` completion guards;
  TEST-03.
- `VOC-082-R03`: **Staging still fails** for an unrelated core-loop
  reason after the streak fix, masking verification. Mitigation: T01
  records honest run evidence; do not claim AC-03 on a different
  failure mode without diagnosis.
- `VOC-082-R04`: **Scope bleed into VOC-081** monitor routing.
  Mitigation: AC-04; DEP-01.
- `VOC-082-DEP-00`: Confirmed root cause (resolved at drafting).
- `VOC-082-DEP-01`: Isolation from VOC-081 (resolved at drafting).
- `VOC-082-EV-00`: T00 fix + deterministic test evidence
  (`t00-evidence.md`).
- `VOC-082-EV-01`: T01 staging core-loop evidence (`t01-evidence.md`).
