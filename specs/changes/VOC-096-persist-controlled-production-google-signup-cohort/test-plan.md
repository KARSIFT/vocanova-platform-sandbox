# VOC-096 — Test Plan

## VOC-096-TEST-00 — Secret precedence: persistent production secret wired in workflow

- Covers: `VOC-096-AC-00`, `VOC-096-AC-08`
- Preconditions: Read deploy-production.yml at implementation revision
- Procedure: Assert workflow references `secrets.PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`
  in validation and config steps; assert `NEW_USER_SIGNUP_ALLOWLIST` is written from that
  env var.
- Expected result: Secret is the sole workflow source for production allowlist
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-01 — Push deploy preserves secret-backed allowlist

- Covers: `VOC-096-AC-00`
- Preconditions: Workflow has no dispatch-only allowlist input path
- Procedure: Assert push-triggered deploy path cannot resolve allowlist from empty input
  default.
- Expected result: Automatic deploy uses secret, not empty string
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-02 — Dispatch input removed; cannot erase cohort

- Covers: `VOC-096-AC-01`
- Preconditions: deploy-production.yml workflow_dispatch block inspected
- Procedure: Assert `new_user_signup_allowlist` input absent; assert no
  `inputs.new_user_signup_allowlist` references remain.
- Expected result: No ephemeral override path exists
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-03 — Fail closed on OAuth enabled + empty cohort

- Covers: `VOC-096-AC-02`
- Preconditions: Run validate-production-signup-allowlist.sh with OAuth enabled and empty allowlist
- Procedure: Assert non-zero exit and fixed diagnostic without email values
- Expected result: Validator refuses empty cohort when OAuth enabled
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-04 — Fail closed on malformed or multiline cohort

- Covers: `VOC-096-AC-02`
- Preconditions: Validator invoked with multiline or syntactically invalid entries using
  `@synthetic.vocanova.invalid` fixtures only
- Procedure: Assert non-zero exit; assert fixture values do not appear in stderr/stdout
- Expected result: Malformed cohort blocked without leaking values
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-05 — Boolean-only logging on successful validation

- Covers: `VOC-096-AC-03`
- Preconditions: Validator invoked with valid synthetic cohort and OAuth enabled/disabled cases
- Procedure: Capture stdout/stderr; assert only readiness boolean lines; no `@` except
  synthetic fixtures never printed
- Expected result: No cohort metadata in logs
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-06 — Production readiness harness fails when controlled_signup_ready is false

- Covers: `VOC-096-AC-04`
- Preconditions: Static analysis or fixture proving harness checks controlled_signup_ready
- Procedure: Assert harness requires `controlled_signup_ready: true` when EXPECT_OAUTH_ENABLED=true
- Expected result: Harness encodes readiness requirement
- Evidence: `VOC-096-EV-01`

## VOC-096-TEST-07 — Synthetic inventory declares production controlled-signup readiness coverage

- Covers: `VOC-096-AC-04`
- Preconditions: infra/monitoring/synthetics.yaml at implementation revision
- Procedure: Assert synthetic.production.oauth-expected-state lists healthz controlled_signup_ready coverage
- Expected result: Inventory documents readiness assertion
- Evidence: `VOC-096-EV-01`

## VOC-096-TEST-08 — /healthz response exposes boolean only (regression)

- Covers: `VOC-096-AC-05`
- Preconditions: Existing apps/api production_test.go controlled_signup_ready tests
- Procedure: Run Go unit tests; confirm no email-like substrings in serialized healthz fixtures
- Expected result: API tests still pass; boolean-only contract preserved
- Evidence: `VOC-096-EV-01`, `VOC-096-EV-02`

## VOC-096-TEST-09 — Canonical OAuth start/callback checks retained in production harness

- Covers: `VOC-096-AC-06`
- Preconditions: verify-production-oauth-start.sh at implementation revision
- Procedure: Static test asserts accounts.google.com, HTTP 200, and canonical callback checks remain
- Expected result: Start/callback contract unchanged aside from additive readiness check
- Evidence: `VOC-096-EV-01`

## VOC-096-TEST-10 — Production deploy never consumes staging secret

- Covers: `VOC-096-AC-07`
- Preconditions: deploy-production.yml contents
- Procedure: Assert no STAGING_NEW_USER_SIGNUP_ALLOWLIST reference
- Expected result: Staging secret isolated
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-11 — Staging deploy never consumes production secret

- Covers: `VOC-096-AC-07`
- Preconditions: deploy-staging.yml contents
- Procedure: Assert no PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST reference
- Expected result: Production secret isolated
- Evidence: `VOC-096-EV-00`

## VOC-096-TEST-12 — Operator procedure documents secure cohort management

- Covers: `VOC-096-AC-09`
- Preconditions: docs/operations/production-controlled-signup.md exists
- Procedure: Foundation doc test asserts secret name, workflow name, redaction rules, and no real email examples
- Expected result: Operator doc meets secure procedure requirements
- Evidence: `VOC-096-EV-02`

## VOC-096-TEST-13 — Live production deploy-and-verify checklist recorded

- Covers: `VOC-096-AC-10`
- Preconditions: T00/T01 merged and promoted; production deploy succeeded
- Procedure: Review t02-evidence.md for deploy URL, healthz jq, harness pass, synthetic green; confirm scrubbing
- Expected result: Evidence complete without secrets or personal data
- Evidence: `VOC-096-EV-02`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
