# VOC-088 — Impact Analysis

## Security and privacy

- **Secrets:** A new GitHub secret (`STAGING_NEW_USER_SIGNUP_ALLOWLIST` or
  equivalent) stores normalized email addresses. These are never committed to git,
  never printed in workflow logs, never included in GitHub issues, and never
  exposed by the readiness endpoint. The deploy step logs only a count.
- **Personal data:** Email addresses in the allowlist are personal data. The secret
  storage, deploy sync (mode 0600 api.env), and redaction guards keep them within
  the GitHub secret → staging host api.env path only.
- **Readiness endpoint:** Exposes `signup_allowlist_count` (integer) only. No email
  values. The `/healthz` endpoint is already publicly accessible; adding a count
  does not expand the data surface.
- **Failure-to-issue agent:** Issue bodies are sanitized by construction. The
  deduplication fingerprint uses workflow name + failure category, not user data.
- **No production signup-policy change.**

## Data and migrations

- No application schema migration.
- No database mutation.
- The only persistent data change is a new GitHub secret (operator-created) and
  workflow configuration changes in git.
- Rollback: revert the workflow changes; the next deploy falls back to the
  previous ephemeral behavior (empty allowlist). The GitHub secret persists
  harmlessly until deleted.

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
- `VOC-088-R02`: The failure-to-issue agent uses `GITHUB_TOKEN`, which has a
  limited lifetime per workflow run. The `if: failure()` step runs in the same
  job; the token is still valid. Cross-job failure handlers would need a separate
  job with `needs:` and `if: failure()`.
- `VOC-088-DEP-00`: Allowlist secret name — implementer choice.
- `VOC-088-DEP-01`: Failure-to-issue mechanism shape — implementer choice.
- `VOC-088-DEP-02`: VOC-084 merged; staging Google OAuth live.
- `VOC-088-DEP-03`: VOC-086-T03 merged; synthetics infrastructure exists.
- `VOC-088-EV-00`: T00 evidence — secret precedence, dispatch constraint,
  fail-closed, redaction tests.
- `VOC-088-EV-01`: T01 evidence — readiness endpoint, synthetic readiness tests.
- `VOC-088-EV-02`: T02 evidence — failure-to-issue deduplication, sanitization,
  responsibility separation tests.
- `VOC-088-EV-03`: T03 evidence — operator procedure, deploy-and-verify checklist.
