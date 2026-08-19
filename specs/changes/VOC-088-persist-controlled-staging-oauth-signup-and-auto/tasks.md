# VOC-088 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → T03**.

## VOC-088-T00 — Persist staging allowlist secret and constrain ephemeral dispatch input

- Requirement source: `VOC-088-D00`, `VOC-088-D01`, `VOC-088-D02`, `VOC-088-D07`
- Acceptance criteria: `VOC-088-AC-00`, `VOC-088-AC-01`, `VOC-088-AC-02`,
  `VOC-088-AC-03`
- Tests: `VOC-088-TEST-00`, `VOC-088-TEST-01`, `VOC-088-TEST-02`,
  `VOC-088-TEST-03`, `VOC-088-TEST-04`
- Evidence: `VOC-088-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Modify `deploy-staging.yml`'s "Write staging application configuration" step
   to read `NEW_USER_SIGNUP_ALLOWLIST` from a persistent GitHub secret (e.g.
   `STAGING_NEW_USER_SIGNUP_ALLOWLIST`) instead of or in addition to the
   `workflow_dispatch` input. The secret value is the single source of truth.
2. Either remove the `new_user_signup_allowlist` dispatch input entirely, or
   make it additive-only (union with the secret). It must never silently replace
   or erase the persistent secret's value.
3. Add a fail-closed guard: when `GOOGLE_OAUTH_ENABLED` evaluates to true and the
   resolved allowlist is empty, the step exits non-zero with a non-sensitive
   diagnostic ("controlled signup cohort is empty; refusing to deploy with OAuth
   enabled and no allowlisted users"). Never print email addresses.
4. Log only a non-sensitive count ("allowlist: N entries synced") or boolean,
   never email values.
5. Add deterministic tests (`scripts/foundation/voc088-*.test.mjs` or equivalent)
   covering:
   - Secret precedence: secret present → secret value written to api.env.
   - Dispatch input removal or additive behavior: push deploy with no input →
     secret-backed value preserved.
   - Fail-closed: OAuth enabled + empty allowlist → non-zero exit.
   - Redaction: no email address in stdout/stderr of the config step.
6. Update the deploy-staging.yml header comment to document the new secret and
   the removal/constraint of the dispatch input.

### Explicitly out of scope for this task

- Readiness endpoint or synthetic extension (T01).
- Failure-to-issue agent (T02).
- Operator docs or deploy-and-verify evidence (T03).
- Production signup-policy change.
- GitHub secret creation itself (that is an operator/founder action outside the
  repository, documented in T03).

## VOC-088-T01 — Extend staging OAuth/readiness synthetic for controlled signup readiness

- Requirement source: `VOC-088-D03`, `VOC-088-D05`
- Acceptance criteria: `VOC-088-AC-04`, `VOC-088-AC-05`, `VOC-088-AC-03`
- Tests: `VOC-088-TEST-05`, `VOC-088-TEST-06`, `VOC-088-TEST-07`
- Evidence: `VOC-088-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-088-T00`

### Required work

1. Extend the API's `/healthz` response (or add a dedicated non-sensitive
   readiness field) to include `signup_allowlist_count` (integer) when
   `NEW_USER_SIGNUP_ENABLED=false` and `SignupAllowlist` is non-nil. The field
   must never expose email values.
2. Extend `synthetic.staging.oauth-expected-state` (or add a sibling synthetic
   in `infra/monitoring/synthetics.yaml`) to assert that when
   `GOOGLE_OAUTH_ENABLED=true` and `NEW_USER_SIGNUP_ENABLED=false`, the
   `/healthz` response contains `signup_allowlist_count > 0`. The synthetic fails
   when the controlled cohort is empty.
3. Update `scheduled-synthetics.yml`'s `staging-oauth-expected-state` job (or add
   a new job) to execute this extended check.
4. Add deterministic tests covering:
   - Readiness endpoint with non-empty allowlist → count > 0.
   - Readiness endpoint with empty allowlist → count = 0.
   - Synthetic failure path when count = 0.
   - Synthetic success path when count > 0.
   - No email values in readiness response.
5. Update `infra/monitoring/synthetics.yaml` coverage metadata if the synthetic
   gains new coverage.

### Explicitly out of scope for this task

- Allowlist persistence in deploy-staging.yml (T00).
- Failure-to-issue agent (T02).
- Live deploy-and-verify (T03).

## VOC-088-T02 — Add governed failure-to-issue agent for scheduled synthetics and deploys

- Requirement source: `VOC-088-D04`, `VOC-088-D05`, `VOC-088-D06`, `VOC-088-D07`
- Acceptance criteria: `VOC-088-AC-06`, `VOC-088-AC-07`, `VOC-088-AC-03`
- Tests: `VOC-088-TEST-08`, `VOC-088-TEST-09`, `VOC-088-TEST-10`,
  `VOC-088-TEST-11`
- Evidence: `VOC-088-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-088-T01`

### Required work

1. Add a reusable mechanism (standalone workflow, composite action, or shell
   script callable from workflow `if: failure()` steps) that:
   - Opens a plain, unlabeled GitHub issue using `GITHUB_TOKEN`.
   - Deduplicates: if an open issue with the same fingerprint already exists, no
     new issue is created.
   - Sanitizes the issue body: workflow name, run URL, failure step, and a
     non-sensitive summary. Never includes secrets, session values, OAuth state,
     raw user identifiers, or unfiltered container logs.
2. Wire this mechanism into `scheduled-synthetics.yml`,
   `deploy-staging.yml`, and `deploy-production.yml` as a failure handler.
3. Add deterministic tests covering:
   - First failure → exactly one issue created.
   - Repeat failure with same fingerprint → no duplicate.
   - Issue body contains no secrets or personal data.
   - Issue is plain and unlabeled (triggers `plan-from-issue`).
   - `GITHUB_TOKEN` App identity is used (not a personal token).
4. Verify that `error-monitoring.yml` is not modified and its Sentry-only scope
   is preserved.

### Explicitly out of scope for this task

- Allowlist persistence (T00).
- Readiness endpoint or synthetic extension (T01).
- Live deploy-and-verify evidence (T03).
- Changing `error-monitoring.yml`.
- Changing `sync-monitoring.yml`.

## VOC-088-T03 — Document operator procedure and deploy-and-verify evidence

- Requirement source: issue #746 requirements 9, 10
- Acceptance criteria: `VOC-088-AC-08`, `VOC-088-AC-09`, `VOC-088-AC-03`
- Tests: `VOC-088-TEST-12`, `VOC-088-TEST-13`
- Evidence: `VOC-088-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-088-T02`

### Required work

1. Add or update `docs/operations/` documentation with a secure operator
   procedure for adding/removing a staging cohort member:
   - How to update the GitHub secret.
   - Verification that the next deploy picks up the change.
   - Proof that a later normal push deploy preserves the cohort.
2. Record deploy-and-verify evidence:
   - A permitted first-time staging Google user can sign in.
   - An unlisted user remains blocked (503).
   - Scheduled checks pass (staging OAuth/readiness synthetic green).
   - A safe controlled failure fixture opens exactly one sanitized issue.
   - A repeat controlled failure creates no duplicate issue.
3. All evidence must be scrubbed of email addresses, credentials, and personal
   data.

### Explicitly out of scope for this task

- Code changes to workflows or application code (T00–T02).
- Production signup-policy change.

## Task ordering notes

- T00 blocks the allowlist persistence required by T01's readiness check.
- T01 blocks the synthetic readiness assertion that T02's failure agent monitors.
- T02 blocks the end-to-end deploy-and-verify evidence in T03.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
