# VOC-088 — Impact Analysis

## Security and privacy

- **Secrets:** The new repository secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST`
  stores normalized email addresses. These are never committed to git,
  never printed in workflow logs, never included in GitHub issues, and never
  exposed by the readiness endpoint. The deploy step logs only a boolean.
- **Personal data:** Email addresses in the allowlist are personal data. The secret
  storage, deploy sync (mode 0600 api.env), and redaction guards keep them within
  the GitHub secret → staging host api.env path only.
- **Readiness endpoint:** Exposes `controlled_signup_ready` (boolean) only. No
  email values or cohort cardinality. The `/healthz` endpoint is already publicly
  accessible; the boolean adds the minimum operational fact.
- **Failure-to-issue agent:** Issue bodies are sanitized by construction. The
  deduplication fingerprint uses workflow name + failure category, not user data.
- **No production signup-policy change.**

## Application and operational surface

- `apps/api/app/api/production.go` and its tests change only the public health
  response shape by adding a non-sensitive boolean. HTTP status and Docker health
  semantics remain unchanged.
- `.github/workflows/operational-failure-monitoring.yml` is a new observer. The
  three source workflows remain independently executable and do not embed
  duplicated failure handlers.
- `deploy-production.yml` is observed but its signup settings, secrets, data,
  containers, and deployment behavior are unchanged.

## Data and migrations

- No application schema migration.
- No database mutation.
- The only persistent data change is a new GitHub secret (operator-created) and
  workflow configuration changes in git.
- Rollback must not restore the previous ephemeral-empty behavior while OAuth is
  enabled. Preserve the secret-backed `api.env` value during rollback, or disable
  staging OAuth before reverting the sync logic. The GitHub secret remains in
  place unless explicitly rotated or removed after OAuth is disabled.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-088-R00`: If the operator creates the GitHub secret with incorrect email
  formatting (extra spaces, mixed case), the normalized comparison in
  `killswitches.go` may not match. Mitigation: the deploy step should normalize
  (lowercase, trim) the secret value before writing to api.env, matching the
  normalization `killswitches.go` already applies. T00 tests cover this.
- `VOC-088-R01`: The fail-closed guard (`VOC-088-D02`) blocks deploys when OAuth
  is enabled but the allowlist is empty. If the operator intends to have OAuth
  enabled for existing users only (no new signups expected), this would be a false
  positive. Mitigation: the guard should apply only when controlled first-time
  signup is documented as an expected staging capability, which is the current
  state per issue #746. If the expectation changes, the guard can be removed in a
  later package.
- `VOC-088-R02`: A default `GITHUB_TOKEN` issue event does not trigger downstream
  workflows. Mitigation: the standalone `workflow_run` observer mints the existing
  automation GitHub App installation token and tests prove `plan-from-issue`
  compatibility.
- `VOC-088-R03`: Merging T00 before the repository allowlist secret is populated
  would make the automatic staging deploy fail closed. Mitigation: creation of
  `STAGING_NEW_USER_SIGNUP_ALLOWLIST` is a release precondition and T00 may not
  merge until the operator confirms it exists; no email value is recorded.
- `VOC-088-R04`: An inline failure handler misses cancellation, timeout, and some
  early workflow failures. Mitigation: use a standalone `workflow_run` observer
  for completed non-success runs.
- `VOC-088-DEP-00`: Resolved — repository secret
  `STAGING_NEW_USER_SIGNUP_ALLOWLIST`, consumed only by staging deployment.
- `VOC-088-DEP-01`: Resolved — standalone `workflow_run` observer using the
  existing automation GitHub App installation token.
- `VOC-088-DEP-02`: VOC-084 merged; staging Google OAuth live.
- `VOC-088-DEP-03`: VOC-086-T03 merged; synthetics infrastructure exists.
- `VOC-088-EV-00`: T00 evidence — secret precedence, dispatch constraint,
  fail-closed, redaction tests.
- `VOC-088-EV-01`: T01 evidence — readiness endpoint, synthetic readiness tests.
- `VOC-088-EV-02`: T02 evidence — failure-to-issue deduplication, sanitization,
  responsibility separation tests.
- `VOC-088-EV-03`: T03 evidence — operator procedure, deploy-and-verify checklist.
