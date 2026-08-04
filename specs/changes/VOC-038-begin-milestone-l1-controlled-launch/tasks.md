# VOC-038 — Tasks

Ordering reflects DOC-12 §5's rollout order, not priority. None is
implementation-authorized by this package; adoption and each task's own
implementation-authorization are separate, mirroring VOC-037's convention.

## VOC-038-T00 — Cohort/allowlist mechanism: decision record + design

- Requirement source: `VOC-038-D00` (founder decided initial composition directly on
  PR #286: founder-only, expandable; see specification.md "Open question 1")
- Acceptance criteria: `VOC-038-AC-00`
- Status: implemented — mechanism is an explicit, normalized email allowlist
  (`SignupAllowlist` on `auth.KillSwitches`, sourced from `NEW_USER_SIGNUP_ALLOWLIST`) checked
  only when `NEW_USER_SIGNUP_ENABLED=false`, satisfying the founder's constraint that
  expanding the cohort requires no code change (env var + redeploy only). This specific
  technical mechanism was Claude Code's engineering choice, not a separate founder decision -
  the founder decided composition/constraint, not the implementation shape.
- Summary: Decision-record-only task (like VOC-037-T00/T02). Present allowlist-mechanism
  options (e.g. an explicit email allowlist checked at signup/login time, distinct from the
  blanket `NEW_USER_SIGNUP_ENABLED` switch) with a recommendation, for the founder to choose
  the composition and mechanism. Does not implement the mechanism.

## VOC-038-T01 — Cohort/allowlist implementation

- Requirement source: `VOC-038-D00`'s accepted decision
- Acceptance criteria: `VOC-038-AC-01`
- Status: implemented in code (auth.KillSwitches.SignupAllowlist, production.go env wiring,
  apps/api/business/auth/service.go call sites); unit-tested; NOT yet verified live against
  production - that verification is still outstanding before AC-01 can be marked satisfied.
- Depends on: `VOC-038-T00`
- Summary: Implement the chosen allowlist mechanism so `NEW_USER_SIGNUP_ENABLED=false` plus an
  explicit allowlist can admit a named small cohort while keeping signup closed to everyone
  else. Must not weaken the existing kill-switch behavior for non-allowlisted users.

## VOC-038-T02 — Production smoke-test suite

- Requirement source: DOC-12 §5's "smoke tests" rollout step
- Acceptance criteria: `VOC-038-AC-02`
- Status: implemented (`infra/scripts/smoke-test-production.sh` + selftest harness,
  wired into `deploy-production.yml` as a post-deploy step); NOT yet run against real
  production - pending a real deploy. `.github/workflows/` and this task both fall under
  the repo's R3 protected-paths policy, so this stays a PR for founder review, not a
  self-merge.
- Summary: A repeatable, scripted smoke-test suite runnable against production immediately
  after any deploy (health endpoints, auth flow reachability, kill-switch state assertions,
  core-loop happy path) — replacing the manual `curl`/SSH checks used ad hoc during R2. Should
  be callable from `deploy-production.yml` as a post-deploy step and runnable standalone.

## VOC-038-T03 — Non-AI core loop validation with the allowlisted cohort

- Requirement source: DOC-12 §5's "validate non-AI core loop" rollout step
- Acceptance criteria: `VOC-038-AC-03`
- Status: pending
- Depends on: `VOC-038-T01`, `VOC-038-T02`
- Summary: With `AI_FEATURES_ENABLED=false` and the cohort mechanism live, confirm the
  allowlisted cohort can complete the full non-AI core loop (auth, discover/save, review,
  missions/progress) with no critical defect, before AI is turned on for anyone real.

## VOC-038-T04 — Enable AI feedback for the allowlisted cohort

- Requirement source: DOC-12 §5's "enable AI for the allowlisted cohort" rollout step
- Acceptance criteria: `VOC-038-AC-04`
- Status: pending
- Depends on: `VOC-038-T03`
- Summary: Turn on `AI_FEATURES_ENABLED` for production, scoped to the cohort only if the
  chosen allowlist mechanism supports feature-level scoping (open question carried from `T00`);
  otherwise document why full-cohort AI enablement is the smallest safe increment. Monitor per
  DOC-12's trigger list from the first real cohort usage.

## VOC-038-T05 — Expansion-threshold decision record

- Requirement source: `VOC-038-D01` (not yet defined — founder decision required;
  see specification.md "Open question 2")
- Acceptance criteria: `VOC-038-AC-05`
- Status: pending
- Summary: Decision-record-only task. Present candidate quantitative thresholds
  (error rate, latency, AI cost/day, sustained over what window) for expanding beyond the
  initial cohort, with a recommendation grounded in current real production baselines
  (Sentry/Kuma data collected during `T03`/`T04`), for the founder to accept or adjust.

## VOC-038-T06 — L1-specific rollback rehearsal

- Requirement source: DOC-12 §5's "rollback controls are proven" gate criterion
- Acceptance criteria: `VOC-038-AC-06`
- Status: pending
- Depends on: `VOC-038-T04`
- Summary: Re-run a genuine rollback rehearsal (same technique as `VOC-037-EV-03`) but with the
  cohort mechanism and AI enabled for real users present, confirming rollback doesn't strand or
  corrupt real (not test-only) cohort user data.

## VOC-038-T07 — L1 release PR and founder hold/expand decision record

- Requirement source: DOC-12 §5's gate, final bullet
- Acceptance criteria: `VOC-038-AC-07`
- Status: pending
- Depends on: `VOC-038-T00` through `VOC-038-T06`
- Summary: Mirrors `VOC-037-T05`. Opens the L1 release PR/issue summarizing all evidence and
  explicitly leaves the hold/expand decision text for the founder to author, exactly as
  `VOC-037-T05` did for R2's go/no-go.
