# VOC-088 — Acceptance Criteria

## VOC-088-AC-00 — Persistent staging allowlist survives automatic deploys

- Requirement source: `VOC-088-D00`, `VOC-088-D01`
- Tasks: `VOC-088-T00`
- Tests: `VOC-088-TEST-00`, `VOC-088-TEST-01`
- Evidence: `VOC-088-EV-00`
- Result: pending

A persistent GitHub secret backs the staging controlled-cohort allowlist.
`deploy-staging.yml` reads this secret and syncs it to `api.env` on every deploy
(push-triggered and manual dispatch). A later normal push deploy preserves the
cohort without operator intervention.

## VOC-088-AC-01 — Ephemeral dispatch input cannot silently erase the cohort

- Requirement source: `VOC-088-D00`
- Tasks: `VOC-088-T00`
- Tests: `VOC-088-TEST-02`
- Evidence: `VOC-088-EV-00`
- Result: pending

The `workflow_dispatch` input `new_user_signup_allowlist` is either removed or
otherwise cannot override the persistent secret. An automatic push
deploy (no dispatch inputs) writes the secret-backed allowlist, not an empty
string.

## VOC-088-AC-02 — Fail closed on malformed configuration

- Requirement source: `VOC-088-D02`
- Tasks: `VOC-088-T00`
- Tests: `VOC-088-TEST-03`
- Evidence: `VOC-088-EV-00`
- Result: pending

When `GOOGLE_OAUTH_ENABLED` is true and the persistent allowlist secret is missing
or empty, the deploy step fails closed (non-zero exit) rather than silently
deploying with an empty cohort. Only a non-sensitive boolean is logged.

## VOC-088-AC-03 — Email addresses never in logs, issues, or evidence

- Requirement source: `VOC-088-D07`
- Tasks: `VOC-088-T00`, `VOC-088-T01`, `VOC-088-T02`, `VOC-088-T03`
- Tests: `VOC-088-TEST-04`
- Evidence: `VOC-088-EV-00`, `VOC-088-EV-01`, `VOC-088-EV-02`
- Result: pending

No deploy log step, synthetic output, failure-to-issue body, test fixture, or
evidence file contains a real email address. Only `@synthetic.vocanova.invalid`
addresses appear in test fixtures.

## VOC-088-AC-04 — Staging OAuth/readiness synthetic asserts controlled signup readiness

- Requirement source: `VOC-088-D03`
- Tasks: `VOC-088-T01`
- Tests: `VOC-088-TEST-05`, `VOC-088-TEST-06`
- Evidence: `VOC-088-EV-01`
- Result: pending

When `GOOGLE_OAUTH_ENABLED=true` and `NEW_USER_SIGNUP_ENABLED=false`, the staging
OAuth/readiness synthetic asserts that `controlled_signup_ready` is `true`. The synthetic
fails when the controlled cohort is empty and controlled first-time signup is an
expected staging capability.

## VOC-088-AC-05 — Non-sensitive readiness endpoint exposes a boolean only

- Requirement source: `VOC-088-D03`
- Tasks: `VOC-088-T01`
- Tests: `VOC-088-TEST-07`
- Evidence: `VOC-088-EV-01`
- Result: pending

The API's `/healthz` (or a dedicated readiness endpoint) exposes
`controlled_signup_ready` as a boolean. It never exposes email values or cohort
cardinality. The boolean accurately reflects whether the runtime policy permits
at least one controlled first-time signup while global signup remains disabled.

## VOC-088-AC-06 — Failed workflows open exactly one deduplicated sanitized issue

- Requirement source: `VOC-088-D04`, `VOC-088-D06`
- Tasks: `VOC-088-T02`
- Tests: `VOC-088-TEST-08`, `VOC-088-TEST-09`, `VOC-088-TEST-10`
- Evidence: `VOC-088-EV-02`
- Result: pending

A failed, cancelled, or timed-out `scheduled-synthetics.yml`,
`deploy-staging.yml`, or `deploy-production.yml` run is observed by a standalone
`workflow_run` workflow and opens exactly one plain, unlabeled GitHub issue using
an installation token minted from `KARSIFT_BOT_APP_ID` and
`KARSIFT_BOT_PRIVATE_KEY`. A repeat observation with the same fingerprint creates
no duplicate. The issue body contains only the workflow name, conclusion, run URL,
and a fixed sanitized summary. It never contains secrets, session values, OAuth
state, raw user identifiers, step output, or unfiltered container logs. The issue
event triggers `plan-from-issue`.

## VOC-088-AC-07 — Responsibility separation preserved

- Requirement source: `VOC-088-D05`
- Tasks: `VOC-088-T02`
- Tests: `VOC-088-TEST-11`
- Evidence: `VOC-088-EV-02`
- Result: pending

Expected signup denials are not turned into Sentry exceptions. The failure-to-issue
agent does not duplicate error-monitoring.yml's Sentry coverage. Sentry, Kuma,
synthetics, and the failure-to-issue agent each cover their documented scope only.

## VOC-088-AC-08 — Operator procedure documented

- Requirement source: issue #746 requirement 9
- Tasks: `VOC-088-T03`
- Tests: `VOC-088-TEST-12`
- Evidence: `VOC-088-EV-03`
- Result: pending

A secure operator procedure documents how to add/remove a staging cohort member
(update the GitHub secret) and proves that a later normal push deploy preserves the
cohort.

## VOC-088-AC-09 — Deploy-and-verify evidence

- Requirement source: issue #746 requirement 10
- Tasks: `VOC-088-T03`
- Tests: `VOC-088-TEST-13`
- Evidence: `VOC-088-EV-03`
- Result: pending

Evidence shows: a permitted first-time staging Google user can sign in; an unlisted
user remains blocked; scheduled checks pass; a safe controlled failure fixture
proves deduplication and App-token-triggered planning without creating an
uncontrolled recursive remediation chain.
