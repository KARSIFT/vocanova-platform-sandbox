# VOC-078 — Impact Analysis

## Security and privacy

No new secrets, credentials, or personal-data handling. Staging verification
continues to use the existing VOC-050 synthetic smoke-test account only.

Residual risks:

- **False PASS evidence** (claiming AC-03 / #575 closed without a green
  run URL) — the exact defect issue #608 reports for PR #598. Mitigation:
  `VOC-078-D00`, AC-00/AC-02, and TEST-00/TEST-01 require a concrete run URL;
  independent review must FAIL URL-less PASS claims.
- **Scope creep into `deploy-staging.yml`.** Mitigation: default exclude;
  R3 path floor if expanded.
- **Product remediation regressions** (double-submit, accessibility
  regressions, Tailwind `max-w-*` collision). Mitigation: T01 instructions
  preserve intentional disables, a11y attributes, and the
  `max-w-[28rem]` workaround; readiness unit tests when product changes.

## Data and migrations

None. No schema, seed, or production application-data change.

## Analytics and accessibility

- **Analytics:** None expected.
- **Accessibility:** Evidence-only path — non-applicable. T01 remediation
  path must preserve review-session fieldset/legend, `aria-pressed`,
  focus-visible, and busy/disabled semantics for assistive tech when the
  prompt is ready vs mid-refetch.

## Risks, dependencies, and evidence

- `VOC-078-R00`: **#575 / staging gate remain unproven if this package is
  not adopted or if implementers invent PASS without a run.** Mitigation:
  this package; review FAIL rule.
- `VOC-078-R01`: **Post-#598 tip still fails on staging** — residual product
  or timing bug. Mitigation: T01 remediation path; leave #575 open until
  green.
- `VOC-078-R02`: **Confusion about VOC-076-T02 / PR #598** leaving two
  competing AC-03 tracks. Mitigation: `VOC-078-DEP-00` disposition at
  adoption; task forbids redispatches of VOC-076-T02.
- `VOC-078-R03`: **FAIL-merge process gap recurs on this package.**
  Mitigation: AC-03 / TEST-03 require independent-review PASS; `VOC-078-DEP-01`
  optional follow-up for pipeline hardening (out of default scope).
- `VOC-078-DEP-00`: Unresolved — VOC-076-T02 disposition after PASS.
- `VOC-078-DEP-01`: Unresolved — whether FAIL-merge hardening is a follow-up.
- `VOC-078-DEP-02`: Unresolved — current post-#598 staging run status.
- `VOC-078-EV-00`: T00 tip SHA, deploy-staging run URL, PASS/FAIL,
  MC coverage, VOC-076 evidence/AC updates, #575 disposition.
- `VOC-078-EV-01`: T01 N/A citation or remediation diff + new green run URL.
