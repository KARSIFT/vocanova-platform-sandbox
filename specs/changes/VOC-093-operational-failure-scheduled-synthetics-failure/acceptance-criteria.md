# VOC-093 — Acceptance Criteria

## VOC-093-AC-00 — Root cause recorded and bounded to route-sweep section

- Requirement source: `VOC-093-D00`, `VOC-093-D01`
- Tasks: `VOC-093-T00`
- Tests: `VOC-093-TEST-00`
- Evidence: `VOC-093-EV-00`
- Result: pending

Task evidence identifies run 32288703894 and job
`synthetic.production.authenticated-route-content-sweep` as the failure source,
names the failing smoke `FAIL` line(s) or route label(s), and explains why
sibling `journey-content` success bounds scope to section **# 6** or harness logic
specific to route-sweep.

## VOC-093-AC-01 — Route-sweep root cause remediated

- Requirement source: `VOC-093-D02`, `VOC-093-D03`
- Tasks: `VOC-093-T00`
- Tests: `VOC-093-TEST-01`, `VOC-093-TEST-02`, `VOC-093-TEST-03`
- Evidence: `VOC-093-EV-00`
- Result: pending

The implemented fix addresses the recorded failing check. Route-sweep remains
non-mutating GET-only with ten fixed routes and two API-derived discover routes.
Protected routes still fail closed on sign-in redirects. No route is silently
removed from coverage without an explicit, reviewed decision recorded in the
task PR.

## VOC-093-AC-02 — Deterministic tests cover the fix or regression fixture

- Requirement source: `VOC-093-D02`, `VOC-093-D03`
- Tasks: `VOC-093-T00`
- Tests: `VOC-093-TEST-04`, `VOC-093-TEST-05`, regression on `VOC-085-TEST-06`,
  `VOC-085-TEST-07`, `VOC-086-TEST-11`, `VOC-086-TEST-12`
- Evidence: `VOC-093-EV-00`
- Result: pending

`bash infra/scripts/smoke-test-production.selftest.sh`,
`node --test scripts/foundation/voc085-production-route-sweep.test.mjs`, and
`node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` pass.
New or extended tests lock the corrected behavior or a representative regression
fixture derived from the failing run.

## VOC-093-AC-03 — Production journey-content and observer contracts unchanged

- Requirement source: `VOC-093-D02`
- Tasks: `VOC-093-T00`
- Tests: `VOC-086-TEST-11`, `VOC-086-TEST-12`, `VOC-088-TEST-08` (regression)
- Evidence: `VOC-093-EV-00`
- Result: pending

`production-journey-content` job wiring, OAuth kill-switch expectations, and
session mint usage remain correct. `operational-failure-monitoring.yml` and
`error-monitoring.yml` are not modified unless a separate package authorizes it.

## VOC-093-AC-04 — Live production route sweep completes green

- Requirement source: issue #771 remediation outcome
- Tasks: `VOC-093-T01`
- Tests: `VOC-093-TEST-06`
- Evidence: `VOC-093-EV-01`
- Result: pending

After T00 merges to `develop` and a production release containing it reaches
protected `main`, a repository-controlled `workflow_dispatch` of
`scheduled-synthetics.yml` at the exact deployed SHA completes with conclusion
`success` and job `synthetic.production.authenticated-route-content-sweep`
success. Evidence records the run URL and duration without secrets, session
values, or personal data.

## VOC-093-AC-05 — No spurious duplicate operational issue for the same fingerprint

- Requirement source: `VOC-088-D06` (observer contract preserved)
- Tasks: `VOC-093-T01`
- Tests: `VOC-088-TEST-09` (regression)
- Evidence: `VOC-093-EV-01`
- Result: pending

Successful remediation does not leave an open duplicate issue for
`scheduled-synthetics:failure` when issue #771 already owns the fingerprint.
Closure of issue #771 follows normal roster/package closure.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
