# VOC-093 — Test Plan

## VOC-093-TEST-00 — Root-cause evidence references run 32288703894 and failing check

- Covers: `VOC-093-AC-00`
- Preconditions: T00 evidence file drafted
- Procedure: Read `t00-evidence.md`; assert it names run 32288703894, job
  `synthetic.production.authenticated-route-content-sweep`, and the specific
  smoke `FAIL:` line(s) or route label(s); assert it explains journey-content
  sibling success as scope bounding.
- Expected result: Evidence bounds remediation to route-sweep section or harness
  logic specific to that section
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-01 — Route-sweep fix addresses recorded failing check

- Covers: `VOC-093-AC-01`
- Preconditions: T00 changes in task branch
- Procedure: Compare `t00-evidence.md` failing check to the task diff; assert the
  diff targets that check (harness assertion, route handler, or documented workflow
  env fix).
- Expected result: Fix is traceable to the recorded failure, not unrelated cleanup
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-02 — Non-mutating route inventory preserved

- Covers: `VOC-093-AC-01`
- Preconditions: T00 changes
- Procedure: Inspect `smoke-test-production.sh` section **# 6**; assert ten fixed
  routes and API-derived discover paths remain; assert no new mutating requests.
- Expected result: VOC-085 route inventory unchanged or intentionally extended with
  reviewer approval recorded in task PR
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-03 — Protected routes still fail closed on sign-in redirect

- Covers: `VOC-093-AC-01`
- Preconditions: T00 changes
- Procedure: Run selftest case covering sign-in redirect rejection; inspect
  `assert_web_route_reachable` for protected routes.
- Expected result: Sign-in redirects on protected routes still fail the suite
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-04 — Smoke selftests pass including new regression fixture

- Covers: `VOC-093-AC-02`
- Preconditions: T00 task branch
- Procedure: Run `bash infra/scripts/smoke-test-production.selftest.sh`.
- Expected result: All cases pass, including any new case modeling the run 32288703894
  failure mode
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-05 — Foundation route-sweep and scheduled-synthetics tests pass

- Covers: `VOC-093-AC-02`, `VOC-093-AC-03`
- Preconditions: T00 task branch
- Procedure: Run `node --test scripts/foundation/voc085-production-route-sweep.test.mjs`
  and `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`.
- Expected result: All tests pass
- Evidence: `VOC-093-EV-00`

## VOC-093-TEST-06 — Live workflow_dispatch route sweep completes success

- Covers: `VOC-093-AC-04`, `VOC-093-AC-05`
- Preconditions: T00 merged to `develop`
- Procedure: Operator dispatches `scheduled-synthetics.yml` targeting
  `synthetic.production.authenticated-route-content-sweep`; inspect run conclusion
  and job status and duration.
- Expected result: Conclusion `success`; route-sweep job success; no duplicate open
  failure fingerprint issue
- Evidence: `VOC-093-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.

## Regression tests (existing, must remain green)

- `VOC-085-TEST-06`, `VOC-085-TEST-07` — route inventory and fail-closed auth handling
- `VOC-086-TEST-11`, `VOC-086-TEST-12` — scheduled synthetics registry/workflow wiring
- `VOC-088-TEST-08`–`VOC-088-TEST-11` — failure-to-issue agent (unchanged files)
