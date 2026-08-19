# VOC-088 — Test Plan

## VOC-088-TEST-00 — Secret precedence: persistent secret written to api.env

- Covers: `VOC-088-AC-00`
- Preconditions: Simulated deploy-staging config step with secret env var set
- Procedure: Run the config step logic with `STAGING_NEW_USER_SIGNUP_ALLOWLIST`
  set to a synthetic test value; assert `api.env` contains that value under
  `NEW_USER_SIGNUP_ALLOWLIST`.
- Expected result: api.env has the secret-backed allowlist value
- Evidence: `VOC-088-EV-00`

## VOC-088-TEST-01 — Push deploy preserves secret-backed allowlist

- Covers: `VOC-088-AC-00`
- Preconditions: Simulated automatic push deploy (no dispatch inputs)
- Procedure: Run the config step logic without dispatch input; assert api.env
  contains the persistent secret value, not empty.
- Expected result: Allowlist is preserved from the secret, not erased
- Evidence: `VOC-088-EV-00`

## VOC-088-TEST-02 — Dispatch input cannot erase secret-backed allowlist

- Covers: `VOC-088-AC-01`
- Preconditions: Secret set; dispatch input either absent or set to different value
- Procedure: If input is removed, verify no input path exists. If additive, verify
  the secret value is always included in the resolved allowlist.
- Expected result: Secret value is never lost by a dispatch input
- Evidence: `VOC-088-EV-00`

## VOC-088-TEST-03 — Fail closed on OAuth enabled + empty allowlist

- Covers: `VOC-088-AC-02`
- Preconditions: GOOGLE_OAUTH_ENABLED=true, resolved allowlist empty
- Procedure: Run the config step guard logic; assert non-zero exit.
- Expected result: Step fails; diagnostic logged without email addresses
- Evidence: `VOC-088-EV-00`

## VOC-088-TEST-04 — No email addresses in deploy logs or test output

- Covers: `VOC-088-AC-03`
- Preconditions: Config step run with synthetic allowlist
- Procedure: Capture stdout/stderr; assert no `@` followed by a domain pattern
  (excluding `@synthetic.vocanova.invalid`).
- Expected result: No real-looking email addresses in output
- Evidence: `VOC-088-EV-00`

## VOC-088-TEST-05 — Readiness synthetic fails when controlled signup is not ready

- Covers: `VOC-088-AC-04`
- Preconditions: Mock /healthz returning `controlled_signup_ready: false`
- Procedure: Run the synthetic check logic; assert non-zero exit.
- Expected result: Synthetic fails when controlled cohort is empty
- Evidence: `VOC-088-EV-01`

## VOC-088-TEST-06 — Readiness synthetic passes when controlled signup is ready

- Covers: `VOC-088-AC-04`
- Preconditions: Mock /healthz returning `controlled_signup_ready: true`
- Procedure: Run the synthetic check logic; assert zero exit.
- Expected result: Synthetic passes when controlled cohort is non-empty
- Evidence: `VOC-088-EV-01`

## VOC-088-TEST-07 — Readiness endpoint exposes a boolean but no cohort metadata

- Covers: `VOC-088-AC-05`
- Preconditions: API unit test with a non-empty allowlist
- Procedure: Call /healthz (or readiness endpoint); assert response contains
  `controlled_signup_ready` as a boolean and no email-like strings or count.
- Expected result: Boolean present; no email values or cardinality exposed
- Evidence: `VOC-088-EV-01`

## VOC-088-TEST-08 — First failure opens exactly one issue

- Covers: `VOC-088-AC-06`
- Preconditions: Mock GitHub API; no existing open issue with matching fingerprint
- Procedure: Invoke failure-to-issue logic with a simulated failure; assert
  exactly one `POST /repos/.../issues` call.
- Expected result: One issue created
- Evidence: `VOC-088-EV-02`

## VOC-088-TEST-09 — Repeat failure creates no duplicate issue

- Covers: `VOC-088-AC-06`
- Preconditions: Mock GitHub API; one existing open issue with matching fingerprint
- Procedure: Invoke failure-to-issue logic with same fingerprint; assert no new
  issue created.
- Expected result: No duplicate issue
- Evidence: `VOC-088-EV-02`

## VOC-088-TEST-10 — Issue body is sanitized

- Covers: `VOC-088-AC-06`, `VOC-088-AC-03`
- Preconditions: Simulated failure with synthetic context
- Procedure: Capture the issue body passed to the mock API; assert it contains
  workflow name and run URL but no secrets, session values, OAuth state, raw user
  identifiers, or container logs.
- Expected result: Issue body is safe for public visibility
- Evidence: `VOC-088-EV-02`

## VOC-088-TEST-11 — Observer identity, coverage, and responsibility separation

- Covers: `VOC-088-AC-07`
- Preconditions: Repository file inspection
- Procedure: Assert `error-monitoring.yml` is unchanged; the standalone observer
  covers the three exact workflow names and non-success conclusions (failure,
  cancelled, timed_out); issue creation uses the automation GitHub App token and
  never the default `GITHUB_TOKEN`; observer/self-test runs are not observed.
- Expected result: Early failures and terminal non-success runs are caught, issue
  events can trigger `plan-from-issue`, recursion is prevented, and each monitoring
  concern retains its documented scope
- Evidence: `VOC-088-EV-02`

## VOC-088-TEST-12 — Operator procedure documented

- Covers: `VOC-088-AC-08`
- Preconditions: docs/operations/ updated
- Procedure: Verify the procedure documents: how to update the GitHub secret,
  verification that the next deploy picks it up, proof of cohort preservation.
- Expected result: Complete operator procedure exists
- Evidence: `VOC-088-EV-03`

## VOC-088-TEST-13 — Deploy-and-verify evidence

- Covers: `VOC-088-AC-09`
- Preconditions: Package merged and deployed to staging
- Procedure: Execute the deploy-and-verify checklist: permitted user signs in,
  unlisted user blocked, synthetics pass, and a bounded controlled fixture proves
  one sanitized App-created issue, downstream planning, and deduplication without
  triggering a recursive remediation chain.
- Expected result: All five checklist items pass
- Evidence: `VOC-088-EV-03`

Tests must not use secrets or production data. All test fixtures use
`@synthetic.vocanova.invalid` addresses only.
