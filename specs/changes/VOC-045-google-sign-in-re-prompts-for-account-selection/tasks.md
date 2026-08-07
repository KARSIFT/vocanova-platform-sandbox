# VOC-045 — Tasks

## VOC-045-T00 — Confirm whether the app-level session persists as intended, independent of Google's prompt behavior

- Requirement source: `specification.md`'s "What remains unconfirmed" section
- Acceptance criteria: `VOC-045-AC-00`
- Tests: `VOC-045-TEST-00`
- Evidence: `VOC-045-EV-00`
- Status: pending

Investigate (do not assume) whether the app's own session
(`SESSION_LIFETIME=720h`, `SESSION_COOKIE_SECURE=true`,
`SESSION_COOKIE_DOMAIN=.vocanova.site`) is correctly issued, persisted, and
honored across page loads and browser restarts, separate from Google's own
OAuth prompt behavior. Document the finding, with reproducible evidence, in
this task's pull request - even if the finding is "the app-level session is
working correctly and the entire symptom is Google's own prompt behavior."
This task must complete, and its finding must be reviewed, before
`VOC-045-T01` begins, since `T01`'s scope may need to change depending on
the outcome. No production code change is required by this task unless the
investigation itself surfaces a session-handling defect distinct from the
`AuthURL` `prompt` behavior - if so, record that as a new open item for the
reviewing human rather than silently expanding this task's own scope.

## VOC-045-T01 — Stop requesting `prompt=consent` unconditionally on every Google sign-in

- Requirement source: `specification.md`'s objective and "Confirmed finding"
  section; gated on `VOC-045-T00`'s completed finding
- Acceptance criteria: `VOC-045-AC-01`, `VOC-045-AC-02`, `VOC-045-AC-03`,
  `VOC-045-AC-04`
- Tests: `VOC-045-TEST-01`, `VOC-045-TEST-02`, `VOC-045-TEST-03`,
  `VOC-045-TEST-04`
- Evidence: `VOC-045-EV-01`
- Status: pending

Modify `apps/api/business/auth/google_oauth.go`'s `AuthURL` and/or
`apps/api/business/auth/service.go`'s `OAuthStart` so that Google's
account-selection/consent screen is not forced on every sign-in attempt,
resolving `specification.md`'s open question 2 (the exact replacement
condition) explicitly in the implementing pull request's description, not
left implicit in the diff. Preserve the existing CSRF state-token check in
`OAuthCallback` unchanged (`VOC-045-AC-03`). Preserve a genuine OAuth
round-trip requirement after explicit sign-out (`VOC-045-AC-02`). Update
`google_oauth_test.go` and `service_test.go` to cover the new behavior,
including a regression test asserting `prompt=consent` is NOT present in
the generated URL for the case(s) the chosen fix addresses, and any test
needed to confirm the `email_verified`-freshness trade-off is intentionally
handled (`VOC-045-AC-04`), not silently dropped. Depends on `VOC-045-T00`'s
finding; if `T00` finds an app-level session defect, this task's scope must
be revisited with the reviewing human before proceeding, per
`change.yaml`'s `VOC-045-DEP-00`.

Tasks preserve scope, separation of duties, and rollback safety. Neither
task may be dispatched before this package is adopted.
